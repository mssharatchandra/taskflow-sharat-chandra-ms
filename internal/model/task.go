package model

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the current state of a task.
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

// Valid returns true if the status is one of the allowed values.
func (s TaskStatus) Valid() bool {
	switch s {
	case TaskStatusTodo, TaskStatusInProgress, TaskStatusDone:
		return true
	}
	return false
}

// TaskPriority represents the urgency of a task.
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

// Valid returns true if the priority is one of the allowed values.
func (p TaskPriority) Valid() bool {
	switch p {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh:
		return true
	}
	return false
}

// Task represents a unit of work within a project.
type Task struct {
	ID          uuid.UUID    `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	ProjectID   uuid.UUID    `json:"project_id"`
	AssigneeID  *uuid.UUID   `json:"assignee_id"` // nullable
	CreatorID   uuid.UUID    `json:"creator_id"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// CreateTaskRequest is the payload to create a new task.
type CreateTaskRequest struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Priority    TaskPriority `json:"priority"`
	AssigneeID  *uuid.UUID   `json:"assignee_id"`
	DueDate     *time.Time   `json:"due_date"`
}

// Validate checks the create task request.
func (r *CreateTaskRequest) Validate() map[string]string {
	errs := make(map[string]string)

	if r.Title == "" {
		errs["title"] = "is required"
	}
	if r.Priority == "" {
		r.Priority = TaskPriorityMedium // default
	} else if !r.Priority.Valid() {
		errs["priority"] = "must be one of: low, medium, high"
	}

	return errs
}

// UpdateTaskRequest is the payload to update an existing task.
type UpdateTaskRequest struct {
	Title       *string       `json:"title,omitempty"`
	Description *string       `json:"description,omitempty"`
	Status      *TaskStatus   `json:"status,omitempty"`
	Priority    *TaskPriority `json:"priority,omitempty"`
	AssigneeID  *uuid.UUID    `json:"assignee_id,omitempty"`
	DueDate     *time.Time    `json:"due_date,omitempty"`
}

// Validate checks the update task request.
func (r *UpdateTaskRequest) Validate() map[string]string {
	errs := make(map[string]string)

	if r.Title != nil && *r.Title == "" {
		errs["title"] = "cannot be empty"
	}
	if r.Status != nil && !r.Status.Valid() {
		errs["status"] = "must be one of: todo, in_progress, done"
	}
	if r.Priority != nil && !r.Priority.Valid() {
		errs["priority"] = "must be one of: low, medium, high"
	}

	return errs
}

// TaskFilter holds optional query parameters for filtering tasks.
type TaskFilter struct {
	Status   *TaskStatus
	Assignee *uuid.UUID
}
