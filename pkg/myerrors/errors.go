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
	log.Printf("Unauthorized: %s\tpath: %s\terror: %v", r.Method, r.URL.Path, err)

	json.WriteJSONError(w, http.StatusUnauthorized, err.Error())
}
