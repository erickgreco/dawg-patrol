package myerrors

import (
	"errors"
	"log"
	"net/http"

	"github.com/erickgreco/dawg-patrol/pkg/json"
)

var (
	ErrEmailAlreadyExists   = errors.New("error, email already exists")
	ErrDataNotFound         = errors.New("error, no data found")
	ErrInvalidCredentials   = errors.New("error, invalid credentials")
	ErrTokenGeneration      = errors.New("error, token generation failed")
	ErrEmptyToken           = errors.New("error, empty token")
	ErrInvalidToken         = errors.New("error, token is invalid")
	ErrInvalidSigningMethod = errors.New("error, invalid signing method")
	ErrParsingClaims        = errors.New("error, unable to parse claims")
	ErrUserNotFound         = errors.New("error, user not found")
	ErrInvalidSerialNumber  = errors.New("error, invalid serial number")
	ErrInvalidRobotName     = errors.New("error, invalid robot name")
	ErrBatteryOutOfRange    = errors.New("error, battery out of valid range")
	ErrInvalidRobotType     = errors.New("error, invalid role in name")
	ErrInvalidUserRole      = errors.New("error, invalid user role")
	ErrPendingRequest       = errors.New("error, request already pending")
	ErrUnavailableRobot     = errors.New("error, unavailable robot")
	ErrInvalidUserID        = errors.New("error, invalid user ID")
	ErrLowBatteryLevel      = errors.New("error, battery level below 10%")
	ErrRobotNotFound        = errors.New("error, robot not found")
)

func InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s\tpath: %s\terror: %v", r.Method, r.URL.Path, err)

	json.WriteJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request error: %s\tpath: %s\terror: %v", r.Method, r.URL.Path, err)

	json.WriteJSONError(w, http.StatusBadRequest, err.Error())
}

func NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found error: %s\tpath: %s\terror: %v", r.Method, r.URL.Path, err)

	json.WriteJSONError(w, http.StatusNotFound, err.Error())
}

func ConflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("conflict error: %s\tpath: %s\terror: %v", r.Method, r.URL.Path, err)

	json.WriteJSONError(w, http.StatusConflict, err.Error())
}

func UnauthorizedResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("unauthorized: %s\tpath: %s\terror: %v", r.Method, r.URL.Path, err)

	json.WriteJSONError(w, http.StatusUnauthorized, err.Error())
}

func UnprocessableEntityResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("unprocessable entity error: %s\tpath: %s\terror: %v", r.Method, r.URL.Path, err)

	json.WriteJSONError(w, http.StatusUnprocessableEntity, err.Error())
}

func ForbiddenResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("forbidden response error: %s\tpath: %s\terror: %v", r.Method, r.URL.Path, err)

	json.WriteJSONError(w, http.StatusForbidden, err.Error())
}
