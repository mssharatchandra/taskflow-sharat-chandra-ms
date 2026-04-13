package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/sharatchandra/taskflow/internal/model"
)

// ProjectRepository defines the interface for project data access.
type ProjectRepository interface {
	Create(ctx context.Context, project *model.Project) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Project, error)
	List(ctx context.Context, userID uuid.UUID) ([]model.Project, error)
	Update(ctx context.Context, project *model.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresProjectRepo struct {
	db *sql.DB
}

// NewProjectRepository creates a new PostgreSQL-backed project repository.
func NewProjectRepository(db *sql.DB) ProjectRepository {
	return &postgresProjectRepo{db: db}
}

func (r *postgresProjectRepo) Create(ctx context.Context, project *model.Project) error {
	query := `
		INSERT INTO projects (id, name, description, owner_id, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		project.ID, project.Name, project.Description, project.OwnerID, project.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting project: %w", err)
	}

	return nil
}

func (r *postgresProjectRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	query := `SELECT id, name, description, owner_id, created_at FROM projects WHERE id = $1`

	var p model.Project
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &desc, &p.OwnerID, &p.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding project: %w", err)
	}

	if desc.Valid {
		p.Description = desc.String
	}

	return &p, nil
}

// List returns all projects that the user owns or has tasks assigned in.
func (r *postgresProjectRepo) List(ctx context.Context, userID uuid.UUID) ([]model.Project, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1
		ORDER BY p.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		var desc sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &desc, &p.OwnerID, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning project: %w", err)
		}
		if desc.Valid {
			p.Description = desc.String
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating projects: %w", err)
	}

	return projects, nil
}

func (r *postgresProjectRepo) Update(ctx context.Context, project *model.Project) error {
	query := `UPDATE projects SET name = $1, description = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, project.Name, project.Description, project.ID)
	if err != nil {
		return fmt.Errorf("updating project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *postgresProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
