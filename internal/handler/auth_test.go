package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sharatchandra/taskflow/internal/handler"
	"github.com/sharatchandra/taskflow/internal/model"
	"github.com/sharatchandra/taskflow/internal/service"
)

// mockUserRepo is a simple in-memory user repository for testing.
type mockUserRepo struct {
	users map[string]*model.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*model.User)}
}

func (r *mockUserRepo) Create(_ context.Context, user *model.User) error {
	r.users[user.Email] = user
	return nil
}

func (r *mockUserRepo) FindByEmail(_ context.Context, email string) (*model.User, error) {
	user, ok := r.users[email]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (r *mockUserRepo) FindByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

func setupAuthRouter(authService *service.AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	authHandler := handler.NewAuthHandler(authService)
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	return router
}

func TestRegister_Success(t *testing.T) {
	repo := newMockUserRepo()
	authService := service.NewAuthService(repo, "test-secret", 4) // low cost for fast tests
	router := setupAuthRouter(authService)

	body := `{"name":"Test User","email":"test@example.com","password":"password123"}`
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", resp.User.Email)
	}
	if resp.User.Name != "Test User" {
		t.Errorf("expected name Test User, got %s", resp.User.Name)
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	repo := newMockUserRepo()
	authService := service.NewAuthService(repo, "test-secret", 4)
	router := setupAuthRouter(authService)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "missing name",
			body:       `{"email":"test@example.com","password":"password123"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing email",
			body:       `{"name":"Test","password":"password123"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "short password",
			body:       `{"name":"Test","email":"test@example.com","password":"short"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	authService := service.NewAuthService(repo, "test-secret", 4)
	router := setupAuthRouter(authService)

	body := `{"name":"Test User","email":"dup@example.com","password":"password123"}`

	// First registration should succeed
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("first register failed: %d", w.Code)
	}

	// Second registration with same email should fail
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepo()
	authService := service.NewAuthService(repo, "test-secret", 4)
	router := setupAuthRouter(authService)

	// Register first
	regBody := `{"name":"Test User","email":"login@example.com","password":"password123"}`
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("register failed: %d", w.Code)
	}

	// Login
	loginBody := `{"email":"login@example.com","password":"password123"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	repo := newMockUserRepo()
	authService := service.NewAuthService(repo, "test-secret", 4)
	router := setupAuthRouter(authService)

	// Register
	regBody := `{"name":"Test User","email":"creds@example.com","password":"password123"}`
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "wrong password",
			body:       `{"email":"creds@example.com","password":"wrongpassword"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "nonexistent email",
			body:       `{"email":"nobody@example.com","password":"password123"}`,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
