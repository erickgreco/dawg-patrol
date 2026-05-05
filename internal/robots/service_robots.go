package robots

import (
	"context"
)

type RobotsRepo interface {
	RegisterRobot(context.Context, *Robot) error
	RegisterEvent(context.Context, *RobotEvents) error
	GetIdleRobots(context.Context) ([]*RobotSummary, error)
}

type RobotService struct {
	store RobotsRepo
}

func NewRobotService(store RobotsRepo) *RobotService {
	return &RobotService{
		store: store,
	}
}
