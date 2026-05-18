package websockets

import (
	"context"
	"encoding/json"
	"math"
	"math/rand"
	"time"

	"github.com/erickgreco/dawg-patrol/internal/robots"
	"github.com/erickgreco/dawg-patrol/internal/users"
	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
	"github.com/google/uuid"
)

const (
	tickerDuration      = 5 * time.Minute
	mockTelemetryTicker = 2 * time.Second
)

type WSService struct {
	userStore  *users.UsersStore
	robotStore *robots.RobotsStore
}

func NewWebSocketService(us *users.UsersStore, rs *robots.RobotsStore) *WSService {
	return &WSService{
		userStore:  us,
		robotStore: rs,
	}
}

func (s *WSService) SessionServiceValidator(ctx context.Context, reservationID, userID, robotID uuid.UUID) error {
	_, err := s.robotStore.ValidateReservation(ctx, reservationID, userID, robotID)
	if err != nil {
		return myerrors.ErrInvalidReservation
	}
	return nil
}

func (s *WSService) KeepReservationAlive(ctx context.Context, reservationID uuid.UUID) {
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.robotStore.ExtendReservation(ctx, reservationID)
		case <-ctx.Done():
			return
		}
	}
}

/*
MockTelemetry simulates robot telemetry by generating random frames and sending
them to the Hub broadcast channel every mockTelemetryTicker interval.
Intended to be run as a goroutine alongside the WS session — stops when ctx is cancelled.
*/
func (s *WSService) MockTelemetry(ctx context.Context, hub *Hub, reservationID uuid.UUID) {
	ticker := time.NewTicker(mockTelemetryTicker)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			frame := RobotTelemetryFrame{
				Speed:     math.Round(rand.Float64()*3.0*100) / 100,
				Direction: math.Round(rand.Float64()*360*10) / 10,
				Battery:   int64(rand.Intn(100)),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}
			data, err := json.Marshal(frame)
			if err != nil {
				continue
			}
			hub.broadcast <- &TelemetryMessage{
				reservationID: reservationID,
				data:          data,
			}
		}
	}
}

/*
MarkWSStarted signals in the DB that a WS connection has been established for
the given reservation, preventing the cleanup worker from freeing the robot
as an abandoned reservation.
*/
func (s *WSService) MarkWSStarted(ctx context.Context, reservationID uuid.UUID) error {
	return s.robotStore.MarkWSStarted(ctx, reservationID)
}

/*
ReleaseReservation immediately deactivates the reservation and sets the robot
back to IDLE when the WS connection closes, without waiting for the cleanup worker.
*/
func (s *WSService) ReleaseReservation(ctx context.Context, reservationID uuid.UUID) error {
	return s.robotStore.DeactivateReservation(ctx, reservationID)
}
