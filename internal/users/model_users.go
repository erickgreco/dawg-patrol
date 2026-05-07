package users

import (
	"time"

	"github.com/erickgreco/dawg-patrol/internal/domain"
	"github.com/google/uuid"
)

// Payload retrieved from user
type Registration struct {
	Username string `json:"username" validate:"required,max=30"`
	Email    string `json:"email" validate:"required,max=250"`
	Password string `json:"password" validate:"required,min=12,max=30"`
}

// Payload to be stored in database
type User struct {
	ID            uuid.UUID   `json:"id"`
	Username      string      `json:"username"`
	Email         string      `json:"email"`
	PasswordHash  string      `json:"password_hash"`
	UserRole      domain.Role `json:"role"`
	Active        bool        `json:"is_active"`
	RequestStatus string      `json:"request_status"`
	RequestedAt   *time.Time  `json:"requested_at"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// Payload to be returned to user
type RegisteredUser struct {
	ID        uuid.UUID   `json:"id"`
	Username  string      `json:"username"`
	UserRole  domain.Role `json:"role"`
	CreatedAt time.Time   `json:"created_at"`
}

// Payload retrived from /login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,max=250"`
	Password string `json:"password" validate:"required,max=30"`
}

// Payload to return
type AuthResponse struct {
	Token string      `json:"token"`
	ID    uuid.UUID   `json:"id"`
	Role  domain.Role `json:"role"`
}

// Payload to be used once user is auth
type UserSummary struct {
	ID       uuid.UUID   `json:"id"`
	Username string      `json:"username"`
	UserRole domain.Role `json:"role"`
}

// Actions to be shown in profile handler
type UserActions struct {
	UpdatePassword    bool        `json:"action_update_password"`
	UpdateUsername    bool        `json:"action_update_username"`
	RequestRoleUpdate RoleRequest `json:"action_request_role_update"`
}

// Payload to be shown in profile handler
type Profile struct {
	ID            uuid.UUID   `json:"id"`
	Username      string      `json:"username"`
	Email         string      `json:"email"`
	UserRole      domain.Role `json:"role"`
	Active        bool        `json:"is_active"`
	RequestStatus string      `json:"request_status"`
	RequestedAt   *time.Time  `json:"requested_at"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// Envelope for all profile data
type ProfileResponse struct {
	Profile *Profile    `json:"profile"`
	Actions UserActions `json:"user_actions"`
}

type RoleRequest struct {
	Action      bool      `json:"action"`
	Status      string    `json:"request_status"`
	RequestDate time.Time `json:"requested_at"`
	Response    string    `json:"request_response"`
}
