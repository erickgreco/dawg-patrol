package home

import (
	"net/http"

	"github.com/erickgreco/dawg-patrol/internal/apimiddleware"
	"github.com/erickgreco/dawg-patrol/pkg/json"
	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
)

type HomeHandler struct {
	service *HomeService
}

func NewHomeHandler(service *HomeService) *HomeHandler {
	return &HomeHandler{
		service: service,
	}
}

/*
Home handler displays minimal user data and current
idle robots
*/
func (h *HomeHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	userID, err := apimiddleware.GetUserIDFromClaimsCtx(r)
	if err != nil {
		myerrors.UnauthorizedResponse(w, r, err)
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
	}
}

func (h *HomeHandler) ReserveRobot(w http.ResponseWriter, r *http.Request) {
	userID, err := apimiddleware.GetUserIDFromClaimsCtx(r)
	if err != nil {
		myerrors.UnauthorizedResponse(w, r, err)
		return
	}

	robotID, err := apimiddleware.GetRobotIDFromCtx(r)
	if err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	reservedRobot, err := h.service.robotService.RobotReservation(ctx, robotID, userID)
	if err != nil {
		switch err {
		case myerrors.ErrRobotNotFound:
			myerrors.NotFoundResponse(w, r, err)
		case myerrors.ErrUnavailableRobot:
			myerrors.ConflictResponse(w, r, err)
		case myerrors.ErrLowBatteryLevel:
			myerrors.ConflictResponse(w, r, err)
		default:
			myerrors.InternalServerError(w, r, err)
		}
		return
	}

	if err := json.JSONResponse(w, http.StatusOK, reservedRobot); err != nil {
		myerrors.InternalServerError(w, r, err)
	}
}
