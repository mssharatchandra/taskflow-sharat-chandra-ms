package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sharatchandra/taskflow/internal/model"
	"github.com/sharatchandra/taskflow/internal/service"
)

// TaskHandler handles task-related HTTP requests.
type TaskHandler struct {
	taskService *service.TaskService
}

// NewTaskHandler creates a new TaskHandler.
func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// List handles GET /projects/:id/tasks.
func (h *TaskHandler) List(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	filter := model.TaskFilter{}

	if statusStr := c.Query("status"); statusStr != "" {
		status := model.TaskStatus(statusStr)
		if !status.Valid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status filter, must be: todo, in_progress, or done"})
			return
		}
		filter.Status = &status
	}

	if assigneeStr := c.Query("assignee"); assigneeStr != "" {
		assigneeID, err := uuid.Parse(assigneeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignee id"})
			return
		}
		filter.Assignee = &assigneeID
	}

	tasks, err := h.taskService.ListByProject(c.Request.Context(), projectID, filter)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// Create handles POST /projects/:id/tasks.
func (h *TaskHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var req model.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": errs})
		return
	}

	task, err := h.taskService.Create(c.Request.Context(), userID, projectID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, task)
}

// Update handles PATCH /tasks/:id.
func (h *TaskHandler) Update(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	var req model.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": errs})
		return
	}

	task, err := h.taskService.Update(c.Request.Context(), taskID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Delete handles DELETE /tasks/:id.
func (h *TaskHandler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	if err := h.taskService.Delete(c.Request.Context(), userID, taskID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
