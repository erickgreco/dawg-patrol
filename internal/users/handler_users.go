package users

import (
	"net/http"

	"github.com/erickgreco/dawg-patrol/pkg/errors"
	"github.com/erickgreco/dawg-patrol/pkg/json"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

/*
	 Retrieves payload from user, validates data, if json has invalid fields
		it will repond with unknown field error, calls h.service.Register to
		apply business logic, returns data confirmation
*/
func (h *Handler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload Registration

	if err := json.ReadJSON(w, r, &payload); err != nil {
		errors.BadRequestResponse(w, r, err)
		return
	}

	if err := json.Validate.Struct(payload); err != nil {
		errors.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	returnPayload, err := h.service.UserRegistration(ctx, &payload)
	if err != nil {
		errors.InternalServerError(w, r, err)
		return
	}

	if err := json.JSONResponse(w, http.StatusCreated, returnPayload); err != nil {
		errors.InternalServerError(w, r, err)
		return
	}
}
