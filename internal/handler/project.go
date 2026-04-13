package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sharatchandra/taskflow/internal/model"
	"github.com/sharatchandra/taskflow/internal/service"
)

// ProjectHandler handles project-related HTTP requests.
type ProjectHandler struct {
	projectService *service.ProjectService
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

// List handles GET /projects.
func (h *ProjectHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	projects, err := h.projectService.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

// Create handles POST /projects.
func (h *ProjectHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req model.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": errs})
		return
	}

	project, err := h.projectService.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// Get handles GET /projects/:id.
func (h *ProjectHandler) Get(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	project, err := h.projectService.GetByID(c.Request.Context(), projectID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, project)
}

// Update handles PATCH /projects/:id.
func (h *ProjectHandler) Update(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var req model.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": errs})
		return
	}

	project, err := h.projectService.Update(c.Request.Context(), userID, projectID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, project)
}

// Delete handles DELETE /projects/:id.
func (h *ProjectHandler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	if err := h.projectService.Delete(c.Request.Context(), userID, projectID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// handleServiceError maps domain errors to HTTP responses.
func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, service.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
