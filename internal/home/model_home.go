package home

import (
	"github.com/erickgreco/dawg-patrol/internal/robots"
	"github.com/erickgreco/dawg-patrol/internal/users"
)

// Home payload
type HomeResponse struct {
	User   users.UserSummary      `json:"user"`
	Robots []*robots.RobotSummary `json:"robots"`
}
