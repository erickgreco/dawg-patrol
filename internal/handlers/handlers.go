package handlers

import (
	"net/http"

	"github.com/erickgreco/dawg-patrol/internal/auth"
	"github.com/erickgreco/dawg-patrol/internal/users"
)

type HomeHandler struct {
	userService *users.Service
}

func NewHomeHandler(us *users.Service) *HomeHandler {
	return &HomeHandler{
		userService: us,
	}
}

// ! This is only a test handler
func (h *HomeHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromCtx(r)

	w.Write([]byte(claims.Role))

}
