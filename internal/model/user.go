package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user in the system.
type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // never include in JSON responses
	CreatedAt time.Time `json:"created_at"`
}

// RegisterRequest is the payload for user registration.
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate checks that all required fields are present and returns a map of
// field-level errors. An empty map means the request is valid.
func (r *RegisterRequest) Validate() map[string]string {
	errs := make(map[string]string)

	if r.Name == "" {
		errs["name"] = "is required"
	}
	if r.Email == "" {
		errs["email"] = "is required"
	}
	if r.Password == "" {
		errs["password"] = "is required"
	} else if len(r.Password) < 8 {
		errs["password"] = "must be at least 8 characters"
	}

	return errs
}

// LoginRequest is the payload for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate checks the login request fields.
func (r *LoginRequest) Validate() map[string]string {
	errs := make(map[string]string)

	if r.Email == "" {
		errs["email"] = "is required"
	}
	if r.Password == "" {
		errs["password"] = "is required"
	}

	return errs
}

// AuthResponse is returned after successful login or registration.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
