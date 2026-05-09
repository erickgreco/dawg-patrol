package users

import (
	"errors"
	"net/http"

	"github.com/erickgreco/dawg-patrol/internal/apimiddleware"
	"github.com/erickgreco/dawg-patrol/pkg/json"
	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
)

type Handler struct {
	service *Service
}

func NewUserHandler(service *Service) *Handler {
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
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	if err := json.Validate.Struct(payload); err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	resp, err := h.service.UserRegistration(ctx, &payload)
	if err != nil {
		switch {
		case errors.Is(err, myerrors.ErrEmailAlreadyExists):
			myerrors.ConflictResponse(w, r, myerrors.ErrEmailAlreadyExists)
		default:
			myerrors.InternalServerError(w, r, err)
		}
		return
	}

	if err := json.JSONResponse(w, http.StatusCreated, resp); err != nil {
		myerrors.InternalServerError(w, r, err)
	}
}

/*
Retrieves payload from user, validates data, if json has invalid fields
it will respond with unknown fields error, calls h.service.UserLogIn to
apply business logic and responds with a token
*/
func (h *Handler) LogInHandler(w http.ResponseWriter, r *http.Request) {
	var payload LoginRequest

	if err := json.ReadJSON(w, r, &payload); err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	if err := json.Validate.Struct(payload); err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	resp, err := h.service.UserLogIn(ctx, &payload)
	if err != nil {
		if errors.Is(err, myerrors.ErrInvalidCredentials) {
			myerrors.UnauthorizedResponse(w, r, myerrors.ErrInvalidCredentials)
			return
		}
		myerrors.InternalServerError(w, r, err)
		return
	}

	if err := json.JSONResponse(w, http.StatusOK, resp); err != nil {
		myerrors.InternalServerError(w, r, err)
	}
}

/*
User profile retrieves all user data excluding password hash,
adicionally posible actions user can implement
*/
func (h *Handler) UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := apimiddleware.GetUserIDFromClaimsCtx(r)
	if err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	profileResp, err := h.service.UserProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, myerrors.ErrUserNotFound) {
			myerrors.UnauthorizedResponse(w, r, myerrors.ErrUserNotFound)
			return
		}
		myerrors.InternalServerError(w, r, err)
		return
	}

	if err := json.JSONResponse(w, http.StatusOK, profileResp); err != nil {
		myerrors.InternalServerError(w, r, err)
	}
}

func (h *Handler) RequestRoleHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := apimiddleware.GetUserIDFromClaimsCtx(r)
	if err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	reqResp, err := h.service.UserRoleRequest(ctx, userID)
	if err != nil {
		if errors.Is(err, myerrors.ErrUserNotFound) {
			myerrors.NotFoundResponse(w, r, myerrors.ErrUserNotFound)
			return
		}
		myerrors.InternalServerError(w, r, err)
		return
	}

	if err := json.JSONResponse(w, http.StatusCreated, reqResp); err != nil {
		myerrors.InternalServerError(w, r, err)
	}
}
