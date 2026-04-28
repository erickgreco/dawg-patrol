package home

import (
	"context"

	"github.com/erickgreco/dawg-patrol/internal/robots"
	"github.com/erickgreco/dawg-patrol/internal/users"
	"github.com/google/uuid"
)

type HomeService struct {
	userStore  *users.UsersStore
	robotStore *robots.RobotsStore
}

func NewHomeService(us *users.UsersStore, rs *robots.RobotsStore) *HomeService {
	return &HomeService{
		userStore:  us,
		robotStore: rs,
	}
}

/*
Since currently only one robot exists this method only
retrives one robot (must need to implement a method
to retrive all active robots)
*/
func (serv *HomeService) GetHomeData(ctx context.Context, userID uuid.UUID) (*HomeResponse, error) {
	user, err := serv.userStore.GetSummaryByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	robots, err := serv.robotStore.GetIdleRobots(ctx)
	if err != nil {
		return nil, err
	}

	return &HomeResponse{
		User:   *user,
		Robots: robots,
	}, nil
}
