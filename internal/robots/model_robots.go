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
// Inconsistency between Category and type is bc type is a Go restricted word
type Robot struct {
	ID           uuid.UUID       `json:"id"`
	SerialNumber string          `json:"serial_number"`
	Name         string          `json:"name"`
	Category     domain.Category `json:"type"`
	Status       string          `json:"status"`
	Battery      int64           `json:"battery"`
	LastSeenAt   time.Time       `json:"last_seen_at"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// Payload to show on handlers
type RobotSummary struct {
	ID           uuid.UUID       `json:"id"`
	SerialNumber string          `json:"serial_number"`
	Name         string          `json:"name"`
	Category     domain.Category `json:"type"`
	Status       string          `json:"status"`
	Battery      int64           `json:"battery"`
	LastSeenAt   time.Time       `json:"last_seen_at"`
}

type IdleRobots struct {
	AssistantRobots []*RobotSummary
	SumoRobots      []*RobotSummary
	RacerRobots     []*RobotSummary
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

type RobotReservation struct {
	ID         uuid.UUID `json:"reservation_id"`
	UserID     uuid.UUID `json:"user_id"`
	RobotID    uuid.UUID `json:"robot_id"`
	ExpiresAt  time.Time `json:"expires_at"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	Status     string    `json:"status"`
	LastSeenAt time.Time `json:"last_seen_at"`
}
