package websockets

import (
	"context"
	"time"

	"github.com/erickgreco/dawg-patrol/internal/robots"
	"github.com/erickgreco/dawg-patrol/internal/users"
	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
	"github.com/google/uuid"
)

const (
	tickerDuration = 5 * time.Minute
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
