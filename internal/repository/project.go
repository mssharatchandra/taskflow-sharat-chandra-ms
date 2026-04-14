package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/sharatchandra/taskflow-sharat-chandra-ms/internal/model"
)

// ProjectRepository defines the interface for project data access.
type ProjectRepository interface {
	Create(ctx context.Context, project *model.Project) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Project, error)
	FindWithTasksByID(ctx context.Context, id uuid.UUID) (*model.ProjectWithTasks, error)
	CanAccess(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (bool, error)
	List(ctx context.Context, userID uuid.UUID, pagination model.PaginationParams) ([]model.Project, int, error)
	GetStats(ctx context.Context, projectID uuid.UUID) (*model.ProjectStats, error)
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

func (r *postgresProjectRepo) FindWithTasksByID(ctx context.Context, id uuid.UUID) (*model.ProjectWithTasks, error) {
	project, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding project with tasks: %w", err)
	}
	if project == nil {
		return nil, nil
	}

	taskQuery := `
		SELECT id, title, description, status, priority, project_id, assignee_id, creator_id, due_date, created_at, updated_at
		FROM tasks
		WHERE project_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, taskQuery, id)
	if err != nil {
		return nil, fmt.Errorf("listing project tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)
	for rows.Next() {
		var t model.Task
		var desc sql.NullString
		var assigneeID sql.NullString
		var dueDate sql.NullTime

		if err := rows.Scan(
			&t.ID, &t.Title, &desc, &t.Status, &t.Priority,
			&t.ProjectID, &assigneeID, &t.CreatorID, &dueDate,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning project task: %w", err)
		}

		if desc.Valid {
			t.Description = desc.String
		}
		if assigneeID.Valid {
			parsedAssigneeID, err := uuid.Parse(assigneeID.String)
			if err != nil {
				return nil, fmt.Errorf("parsing assignee id: %w", err)
			}
			t.AssigneeID = &parsedAssigneeID
		}
		if dueDate.Valid {
			due := dueDate.Time.UTC()
			t.DueDate = &due
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating project tasks: %w", err)
	}

	return &model.ProjectWithTasks{
		Project: *project,
		Tasks:   tasks,
	}, nil
}

func (r *postgresProjectRepo) CanAccess(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM projects p
			WHERE p.id = $1
			  AND (
				p.owner_id = $2
				OR EXISTS (
					SELECT 1
					FROM tasks t
					WHERE t.project_id = p.id
					  AND t.assignee_id = $2
				)
			  )
		)`

	var allowed bool
	if err := r.db.QueryRowContext(ctx, query, projectID, userID).Scan(&allowed); err != nil {
		return false, fmt.Errorf("checking project access: %w", err)
	}

	return allowed, nil
}

// List returns paginated projects that the user owns or has tasks assigned in.
func (r *postgresProjectRepo) List(ctx context.Context, userID uuid.UUID, pagination model.PaginationParams) ([]model.Project, int, error) {
	countQuery := `
		SELECT COUNT(DISTINCT p.id)
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1`

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting projects: %w", err)
	}

	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, pagination.Limit, pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("listing projects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		var desc sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &desc, &p.OwnerID, &p.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning project: %w", err)
		}
		if desc.Valid {
			p.Description = desc.String
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterating projects: %w", err)
	}

	return projects, total, nil
}

func (r *postgresProjectRepo) GetStats(ctx context.Context, projectID uuid.UUID) (*model.ProjectStats, error) {
	stats := &model.ProjectStats{
		ByStatus: map[string]int{
			string(model.TaskStatusTodo):       0,
			string(model.TaskStatusInProgress): 0,
			string(model.TaskStatusDone):       0,
		},
		ByAssignee: make([]model.AssigneeTaskCount, 0),
	}

	statusQuery := `
		SELECT status, COUNT(*)
		FROM tasks
		WHERE project_id = $1
		GROUP BY status`

	statusRows, err := r.db.QueryContext(ctx, statusQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("querying status stats: %w", err)
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status string
		var count int
		if err := statusRows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scanning status stat: %w", err)
		}
		stats.ByStatus[status] = count
	}

	if err := statusRows.Err(); err != nil {
		return nil, fmt.Errorf("iterating status stats: %w", err)
	}

	assigneeQuery := `
		SELECT assignee_id, COUNT(*)
		FROM tasks
		WHERE project_id = $1
		GROUP BY assignee_id
		ORDER BY COUNT(*) DESC`

	assigneeRows, err := r.db.QueryContext(ctx, assigneeQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("querying assignee stats: %w", err)
	}
	defer assigneeRows.Close()

	for assigneeRows.Next() {
		var assigneeID sql.NullString
		var count int
		if err := assigneeRows.Scan(&assigneeID, &count); err != nil {
			return nil, fmt.Errorf("scanning assignee stat: %w", err)
		}

		item := model.AssigneeTaskCount{Count: count}
		if assigneeID.Valid {
			parsedAssigneeID, err := uuid.Parse(assigneeID.String)
			if err != nil {
				return nil, fmt.Errorf("parsing assignee stat id: %w", err)
			}
			item.AssigneeID = &parsedAssigneeID
		}

		stats.ByAssignee = append(stats.ByAssignee, item)
	}

	if err := assigneeRows.Err(); err != nil {
		return nil, fmt.Errorf("iterating assignee stats: %w", err)
	}

	return stats, nil
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
