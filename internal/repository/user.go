package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/sharatchandra/taskflow/internal/model"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// postgresUserRepo implements UserRepository backed by PostgreSQL.
type postgresUserRepo struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL-backed user repository.
func NewUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepo{db: db}
}

func (r *postgresUserRepo) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, name, email, password, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Name, user.Email, user.Password, user.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}

	return nil
}

func (r *postgresUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, name, email, password, created_at FROM users WHERE email = $1`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by email: %w", err)
	}

	return &user, nil
}

func (r *postgresUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `SELECT id, name, email, password, created_at FROM users WHERE id = $1`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by id: %w", err)
	}

	return &user, nil
}
