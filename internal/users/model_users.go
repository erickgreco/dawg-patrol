package users

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Payload retrieved from user
type Registration struct {
	Username string `json:"username" validate:"required,max=30"`
	Email    string `json:"email" validate:"required,max=250"`
	Password string `json:"password" validate:"required,max=30"`
}

// TODO: work with tokens
type Role string

const (
	RoleAdmin    Role = "ADMIN"
	RoleOperator Role = "OPERATOR"
	RoleViewer   Role = "VIEWER"
)

// Payload to be stored in database
type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	UserRole     Role      `json:"role"`
	Active       bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Payload to be returned to user
type RegisteredUser struct {
	Username  string    `json:"username"`
	UserRole  Role      `json:"role"`
	Active    bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Payload retrived from /login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,max=250"`
	Password string `json:"password" validate:"required,max=30"`
}

// Data used to create token
type Claims struct {
	Sub  string `json:"sub"`
	Role Role   `json:"role"`
	jwt.RegisteredClaims
}

// Payload to return
type AuthResponse struct {
	Token string `json:"token"`
}
