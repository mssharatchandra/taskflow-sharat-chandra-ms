package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sharatchandra/taskflow/internal/model"
	"github.com/sharatchandra/taskflow/internal/repository"
)

// ProjectService handles business logic for projects.
type ProjectService struct {
	projectRepo repository.ProjectRepository
}

// NewProjectService creates a new ProjectService.
func NewProjectService(projectRepo repository.ProjectRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo}
}

// Create creates a new project owned by the given user.
func (s *ProjectService) Create(ctx context.Context, userID uuid.UUID, req model.CreateProjectRequest) (*model.Project, error) {
	project := &model.Project{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     userID,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}

	return project, nil
}

// GetByID returns a project and its tasks by project ID.
func (s *ProjectService) GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*model.ProjectWithTasks, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding project: %w", err)
	}
	if project == nil {
		return nil, ErrNotFound
	}

	canAccess, err := s.projectRepo.CanAccess(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("checking project access: %w", err)
	}
	if !canAccess {
		return nil, ErrForbidden
	}

	projectWithTasks, err := s.projectRepo.FindWithTasksByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding project with tasks: %w", err)
	}
	if projectWithTasks == nil {
		return nil, ErrNotFound
	}

	return projectWithTasks, nil
}

// List returns all projects accessible to the given user.
func (s *ProjectService) List(ctx context.Context, userID uuid.UUID, pagination model.PaginationParams) ([]model.Project, model.PaginationMeta, error) {
	projects, total, err := s.projectRepo.List(ctx, userID, pagination)
	if err != nil {
		return nil, model.PaginationMeta{}, fmt.Errorf("listing projects: %w", err)
	}

	// Return empty slice instead of nil for consistent JSON output
	if projects == nil {
		projects = []model.Project{}
	}

	meta := model.NewPaginationMeta(pagination, total)
	return projects, meta, nil
}

// Stats returns aggregate task counts by status and assignee for a project.
func (s *ProjectService) Stats(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) (*model.ProjectStats, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("finding project: %w", err)
	}
	if project == nil {
		return nil, ErrNotFound
	}

	canAccess, err := s.projectRepo.CanAccess(ctx, projectID, userID)
	if err != nil {
		return nil, fmt.Errorf("checking project access: %w", err)
	}
	if !canAccess {
		return nil, ErrForbidden
	}

	stats, err := s.projectRepo.GetStats(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("getting project stats: %w", err)
	}

	return stats, nil
}

// Update modifies a project. Only the owner can update.
func (s *ProjectService) Update(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req model.UpdateProjectRequest) (*model.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("finding project: %w", err)
	}
	if project == nil {
		return nil, ErrNotFound
	}

	if project.OwnerID != userID {
		return nil, ErrForbidden
	}

	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("updating project: %w", err)
	}

	return project, nil
}

// Delete removes a project and all its tasks. Only the owner can delete.
func (s *ProjectService) Delete(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) error {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("finding project: %w", err)
	}
	if project == nil {
		return ErrNotFound
	}

	if project.OwnerID != userID {
		return ErrForbidden
	}

	if err := s.projectRepo.Delete(ctx, projectID); err != nil {
		return fmt.Errorf("deleting project: %w", err)
	}

	return nil
}
