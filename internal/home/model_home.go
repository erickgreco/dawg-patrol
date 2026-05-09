package home

import (
	"github.com/erickgreco/dawg-patrol/internal/robots"
	"github.com/erickgreco/dawg-patrol/internal/users"
)

// Home payload
type HomeResponse struct {
	User      *users.UserSummary     `json:"user"`
	NumRobots int                    `json:"available_robots_count"`
	Robots    []*robots.RobotSummary `json:"robots"`
}
