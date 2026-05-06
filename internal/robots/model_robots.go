package robots

import (
	"time"

	"github.com/erickgreco/dawg-patrol/internal/domain"
	"github.com/google/uuid"
)

// Payload retrieved from robot to register
type RobotRegistration struct {
	SerialNumber string `json:"serial_number"`
	Name         string `json:"name"`
	Battery      int64  `json:"battery"`
}

// Payload to be stored in DB
type Robot struct {
	ID           uuid.UUID   `json:"id"`
	SerialNumber string      `json:"serial_number"`
	Name         string      `json:"name"`
	Role         domain.Role `json:"role"`
	Status       string      `json:"status"`
	Battery      int64       `json:"battery"`
	LastSeenAt   time.Time   `json:"last_seen_at"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// Payload to show on handlers
type RobotSummary struct {
	ID           uuid.UUID   `json:"id"`
	SerialNumber string      `json:"serial_number"`
	Name         string      `json:"name"`
	Role         domain.Role `json:"role"`
	Status       string      `json:"status"`
	Battery      int64       `json:"battery"`
	LastSeenAt   time.Time   `json:"last_seen_at"`
}

// Payload to work with commands
type CommandType string

const (
	CommandForward   CommandType = "move_forward"
	CommandBackward  CommandType = "move_backward"
	CommandTurnRight CommandType = "turn_right"
	CommandTurnLeft  CommandType = "turn_left"
	CommandStop      CommandType = "stop"
)

type Command struct {
	Type    CommandType    `json:"type"`
	Payload map[string]any `json:"payload"`
}
