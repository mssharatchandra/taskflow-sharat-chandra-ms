package model

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a project that contains tasks.
type Project struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	OwnerID     uuid.UUID `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProjectWithTasks includes the project details along with its tasks.
type ProjectWithTasks struct {
	Project
	Tasks []Task `json:"tasks"`
}

// CreateProjectRequest is the payload to create a new project.
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Validate checks the create project request.
func (r *CreateProjectRequest) Validate() map[string]string {
	errs := make(map[string]string)

	if r.Name == "" {
		errs["name"] = "is required"
	}

	return errs
}

// UpdateProjectRequest is the payload to update an existing project.
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Validate checks the update project request.
func (r *UpdateProjectRequest) Validate() map[string]string {
	errs := make(map[string]string)

	if r.Name != nil && *r.Name == "" {
		errs["name"] = "cannot be empty"
	}

	return errs
}
