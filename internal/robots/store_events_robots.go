package robots

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	CONNECTED    EventType = "ROBOT_CONNECTED"
	DISCONNECTED EventType = "ROBOT_DISCONNECTED"
)

// Important events to be stored
type RobotEvents struct {
	ID        uuid.UUID     `json:"event_id"`
	RobotID   uuid.UUID     `json:"robot_id"`
	Event     EventType     `json:"event"`
	IssuedBy  uuid.NullUUID `json:"user_id"` // To be used with Stop, START, ETC
	CreatedAt time.Time     `json:"created_at"`
}

func (r *RobotsStore) RegisterEvent(ctx context.Context, robotEvent *RobotEvents) error {
	query := `
		INSERT INTO robot_events (event_id, robot_id, event, user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING event_id, robot_id, event, user_id, created_at
	`
	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	err := r.db.QueryRow(
		ctx,
		query,
		robotEvent.ID,
		robotEvent.RobotID,
		robotEvent.Event,
		robotEvent.IssuedBy,
	).Scan(
		&robotEvent.ID,
		&robotEvent.RobotID,
		&robotEvent.Event,
		&robotEvent.IssuedBy,
		&robotEvent.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}
