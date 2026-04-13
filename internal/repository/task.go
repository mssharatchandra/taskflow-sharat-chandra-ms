package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sharatchandra/taskflow/internal/model"
)

// TaskRepository defines the interface for task data access.
type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Task, error)
	ListByProject(ctx context.Context, projectID uuid.UUID, filter model.TaskFilter) ([]model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresTaskRepo struct {
	db *sql.DB
}

// NewTaskRepository creates a new PostgreSQL-backed task repository.
func NewTaskRepository(db *sql.DB) TaskRepository {
	return &postgresTaskRepo{db: db}
}

func (r *postgresTaskRepo) Create(ctx context.Context, task *model.Task) error {
	query := `
		INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, creator_id, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.Title, task.Description, task.Status, task.Priority,
		task.ProjectID, task.AssigneeID, task.CreatorID, task.DueDate,
		task.CreatedAt, task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting task: %w", err)
	}

	return nil
}

func (r *postgresTaskRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	query := `
		SELECT id, title, description, status, priority, project_id, assignee_id, creator_id, due_date, created_at, updated_at
		FROM tasks WHERE id = $1`

	var t model.Task
	var desc sql.NullString
	var assignee *uuid.UUID
	var dueDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Title, &desc, &t.Status, &t.Priority,
		&t.ProjectID, &assignee, &t.CreatorID, &dueDate,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding task: %w", err)
	}

	if desc.Valid {
		t.Description = desc.String
	}
	t.AssigneeID = assignee
	if dueDate.Valid {
		t.DueDate = &dueDate.Time
	}

	return &t, nil
}

func (r *postgresTaskRepo) ListByProject(ctx context.Context, projectID uuid.UUID, filter model.TaskFilter) ([]model.Task, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("t.project_id = $%d", argIdx))
	args = append(args, projectID)
	argIdx++

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}

	if filter.Assignee != nil {
		conditions = append(conditions, fmt.Sprintf("t.assignee_id = $%d", argIdx))
		args = append(args, *filter.Assignee)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT t.id, t.title, t.description, t.status, t.priority, t.project_id,
		       t.assignee_id, t.creator_id, t.due_date, t.created_at, t.updated_at
		FROM tasks t
		WHERE %s
		ORDER BY t.created_at DESC`, strings.Join(conditions, " AND "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		var desc sql.NullString
		var assignee *uuid.UUID
		var dueDate sql.NullTime

		if err := rows.Scan(
			&t.ID, &t.Title, &desc, &t.Status, &t.Priority,
			&t.ProjectID, &assignee, &t.CreatorID, &dueDate,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}

		if desc.Valid {
			t.Description = desc.String
		}
		t.AssigneeID = assignee
		if dueDate.Valid {
			t.DueDate = &dueDate.Time
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating tasks: %w", err)
	}

	return tasks, nil
}

func (r *postgresTaskRepo) Update(ctx context.Context, task *model.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, priority = $4,
		    assignee_id = $5, due_date = $6
		WHERE id = $7`

	result, err := r.db.ExecContext(ctx, query,
		task.Title, task.Description, task.Status, task.Priority,
		task.AssigneeID, task.DueDate, task.ID,
	)
	if err != nil {
		return fmt.Errorf("updating task: %w", err)
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

func (r *postgresTaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting task: %w", err)
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
