package home

import (
	"net/http"

	"github.com/erickgreco/dawg-patrol/internal/auth"
	"github.com/erickgreco/dawg-patrol/pkg/json"
	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
	"github.com/google/uuid"
)

type HomeHandler struct {
	service *HomeService
}

func NewHomeHandler(service *HomeService) *HomeHandler {
	return &HomeHandler{
		service: service,
	}
}

// ! This is only a test method
func (h *HomeHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.GetClaimsFromCtx(r)
	if err != nil {
		myerrors.UnauthorizedResponse(w, r, err)
		return
	}

	userID, err := uuid.Parse(claims.Sub)
	if err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	data, err := h.service.GetHomeData(ctx, userID)
	if err != nil {
		myerrors.InternalServerError(w, r, err)
		return
	}

	if err := json.JSONResponse(w, http.StatusOK, data); err != nil {
		myerrors.InternalServerError(w, r, err)
		return
	}
}
