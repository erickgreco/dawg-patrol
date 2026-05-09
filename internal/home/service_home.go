package home

import (
	"context"
	"errors"

	"github.com/erickgreco/dawg-patrol/internal/domain"
	"github.com/erickgreco/dawg-patrol/internal/robots"
	"github.com/erickgreco/dawg-patrol/internal/users"
	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
	"github.com/google/uuid"
)

type HomeService struct {
	userService  *users.Service
	robotService *robots.Service
}

func NewHomeService(us *users.Service, rs *robots.Service) *HomeService {
	return &HomeService{
		userService:  us,
		robotService: rs,
	}
}

/*
Method to separate robots to show according to userRole
Currently it only displays idle robots
viewer = idle robots (no assistants) without any robot data
operator = idle robots (no assistants) with robot data
admin = idle robots with data
*/
func (serv *HomeService) GetHomeData(ctx context.Context, userID uuid.UUID) (*HomeResponse, error) {
	user, err := serv.userService.UserSummaryByRole(ctx, userID)
	if err != nil {
		if errors.Is(err, myerrors.ErrInvalidUserID) {
			return nil, myerrors.ErrInvalidToken
		}
		return nil, err
	}

	robots, err := serv.robotService.IdleRobots(ctx)
	if err != nil {
		return nil, err
	}

	racerCount := len(robots.RacerRobots)
	sumoCount := len(robots.SumoRobots)
	assistantCount := len(robots.AssistantRobots)

	operatorRobots := mergeRobotSlices(robots.RacerRobots, robots.SumoRobots)
	adminRobots := mergeRobotSlices(robots.AssistantRobots, robots.RacerRobots, robots.SumoRobots)

	response := &HomeResponse{
		User: user,
	}

	switch user.UserRole {
	case domain.RoleViewer:
		response.NumRobots = racerCount + sumoCount
		response.Robots = nil
	case domain.RoleOperator:
		response.NumRobots = racerCount + sumoCount
		response.Robots = operatorRobots
	case domain.RoleAdmin:
		response.NumRobots = racerCount + sumoCount + assistantCount
		response.Robots = adminRobots
	default:
		return nil, myerrors.ErrInvalidToken
	}
	return response, nil
}

//? Will be used once handler shows more than just idle robots
// func countRobotsByRole(robots []*robots.RobotSummary, category domain.Category) int {
// 	count := 0

// 	for _, robot := range robots {
// 		if robot.Category == category {
// 			count++
// 		}
// 	}
// 	return count
// }

func mergeRobotSlices(slices ...[]*robots.RobotSummary) []*robots.RobotSummary {
	var merged []*robots.RobotSummary

	for _, slice := range slices {
		merged = append(merged, slice...)
	}
	return merged
}
