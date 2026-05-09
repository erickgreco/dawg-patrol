package robots

import (
	"errors"
	"net/http"

	"github.com/erickgreco/dawg-patrol/pkg/json"
	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
)

type Handler struct {
	service *Service
}

func NewRobotHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRobotHandler(w http.ResponseWriter, r *http.Request) {
	var payload RobotRegistration

	if err := json.ReadJSON(w, r, &payload); err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	if err := json.Validate.Struct(payload); err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	resp, err := h.service.RobotRegistration(ctx, &payload)
	if err != nil {
		switch {
		case errors.Is(err, myerrors.ErrInvalidSerialNumber):
			myerrors.UnprocessableEntityResponse(w, r, myerrors.ErrInvalidSerialNumber)
		case errors.Is(err, myerrors.ErrInvalidRobotName):
			myerrors.UnprocessableEntityResponse(w, r, myerrors.ErrInvalidRobotName)
		case errors.Is(err, myerrors.ErrBatteryOutOfRange):
			myerrors.UnprocessableEntityResponse(w, r, myerrors.ErrBatteryOutOfRange)
		default:
			myerrors.InternalServerError(w, r, err)
		}
		return
	}

	if err := json.JSONResponse(w, http.StatusAccepted, resp); err != nil {
		myerrors.InternalServerError(w, r, err)
	}
}

func (h *Handler) IdleRobotsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	robots, err := h.service.IdleRobots(ctx)
	if err != nil {
		myerrors.InternalServerError(w, r, err)
		return
	}

	if err := json.JSONResponse(w, http.StatusOK, robots); err != nil {
		myerrors.InternalServerError(w, r, err)
	}
}
