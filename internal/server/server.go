package server

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sharatchandra/taskflow/internal/config"
	"github.com/sharatchandra/taskflow/internal/handler"
	"github.com/sharatchandra/taskflow/internal/middleware"
	"github.com/sharatchandra/taskflow/internal/repository"
	"github.com/sharatchandra/taskflow/internal/service"
)

// Server holds the HTTP server and its dependencies.
type Server struct {
	httpServer *http.Server
	router     *gin.Engine
}

// New creates a fully wired Server with all routes registered.
func New(cfg *config.Config, db *sql.DB) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())
	router.Use(corsMiddleware())

	// Repositories
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.BcryptCost)
	projectService := service.NewProjectService(projectRepo)
	taskService := service.NewTaskService(taskRepo, projectRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskService)

	// Public routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Protected routes
	authorized := router.Group("/")
	authorized.Use(middleware.Auth(cfg.JWTSecret))
	{
		authorized.GET("/projects", projectHandler.List)
		authorized.POST("/projects", projectHandler.Create)
		authorized.GET("/projects/:id", projectHandler.Get)
		authorized.PATCH("/projects/:id", projectHandler.Update)
		authorized.DELETE("/projects/:id", projectHandler.Delete)

		authorized.GET("/projects/:id/tasks", taskHandler.List)
		authorized.POST("/projects/:id/tasks", taskHandler.Create)
		authorized.PATCH("/tasks/:id", taskHandler.Update)
		authorized.DELETE("/tasks/:id", taskHandler.Delete)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return &Server{
		httpServer: srv,
		router:     router,
	}
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	slog.Info("server starting", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully stops the server with a timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("server shutting down")
	return s.httpServer.Shutdown(ctx)
}

// requestLogger logs each incoming HTTP request using slog.
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start).String(),
			"ip", c.ClientIP(),
		)
	}
}

// corsMiddleware handles Cross-Origin Resource Sharing headers.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
