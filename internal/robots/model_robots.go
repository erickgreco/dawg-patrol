package robots

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAssistant Role = "ASSISTANT"
	RoleSumo      Role = "SUMO"
	RoleRacer     Role = "RACER"
)

// Payload to be stored in DB
type Robot struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Role      Role      `json:"role"`
	Status    string    `json:"status"`
	Battery   int64     `json:"battery"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Payload to show on handlers
type RobotSummary struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Role    Role      `json:"role"`
	Status  string    `json:"status"`
	Battery int64     `json:"battery"`
}
