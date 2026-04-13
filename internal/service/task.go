package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sharatchandra/taskflow/internal/model"
	"github.com/sharatchandra/taskflow/internal/repository"
)

// TaskService handles business logic for tasks.
type TaskService struct {
	taskRepo    repository.TaskRepository
	projectRepo repository.ProjectRepository
}

// NewTaskService creates a new TaskService.
func NewTaskService(taskRepo repository.TaskRepository, projectRepo repository.ProjectRepository) *TaskService {
	return &TaskService{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
	}
}

// Create adds a new task to a project.
func (s *TaskService) Create(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req model.CreateTaskRequest) (*model.Task, error) {
	// Verify the project exists
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("finding project: %w", err)
	}
	if project == nil {
		return nil, ErrNotFound
	}

	now := time.Now().UTC()
	task := &model.Task{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		Status:      model.TaskStatusTodo,
		Priority:    req.Priority,
		ProjectID:   projectID,
		AssigneeID:  req.AssigneeID,
		CreatorID:   userID,
		DueDate:     req.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("creating task: %w", err)
	}

	return task, nil
}

// GetByID returns a task by ID.
func (s *TaskService) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding task: %w", err)
	}
	if task == nil {
		return nil, ErrNotFound
	}

	return task, nil
}

// ListByProject returns tasks for a project, optionally filtered.
func (s *TaskService) ListByProject(ctx context.Context, projectID uuid.UUID, filter model.TaskFilter) ([]model.Task, error) {
	// Verify the project exists
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("finding project: %w", err)
	}
	if project == nil {
		return nil, ErrNotFound
	}

	tasks, err := s.taskRepo.ListByProject(ctx, projectID, filter)
	if err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}

	if tasks == nil {
		tasks = []model.Task{}
	}

	return tasks, nil
}

// Update modifies a task. Any authenticated user can update tasks in a project.
func (s *TaskService) Update(ctx context.Context, taskID uuid.UUID, req model.UpdateTaskRequest) (*model.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("finding task: %w", err)
	}
	if task == nil {
		return nil, ErrNotFound
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.AssigneeID != nil {
		task.AssigneeID = req.AssigneeID
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("updating task: %w", err)
	}

	// Re-fetch to get the updated_at from the database trigger
	updated, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("fetching updated task: %w", err)
	}

	return updated, nil
}

// Delete removes a task. Only the project owner or task creator can delete.
func (s *TaskService) Delete(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("finding task: %w", err)
	}
	if task == nil {
		return ErrNotFound
	}

	// Check authorization: project owner or task creator
	project, err := s.projectRepo.FindByID(ctx, task.ProjectID)
	if err != nil {
		return fmt.Errorf("finding project: %w", err)
	}

	isProjectOwner := project != nil && project.OwnerID == userID
	isTaskCreator := task.CreatorID == userID

	if !isProjectOwner && !isTaskCreator {
		return ErrForbidden
	}

	if err := s.taskRepo.Delete(ctx, taskID); err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}

	return nil
}
