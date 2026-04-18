package myerrors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/erickgreco/dawg-patrol/pkg/json"
)

var (
	ErrEmailAlreadyExists = errors.New("error, email already exists")
	ErrDataNotFound       = errors.New("error, no data found")
	ErrInvalidCredentials = errors.New("error, invalid credentials")
)

func InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Printf("internal server error: %s\tpath: %s\terror: %s", r.Method, r.URL.Path, err)

	json.WriteJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Printf("bad request error: %s\tpath: %s\terror: %s", r.Method, r.URL.Path, err)

	json.WriteJSONError(w, http.StatusBadRequest, err.Error())
}

func NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Printf("not found error: %s\tpath: %s\terror: %s", r.Method, r.URL.Path, err)

	json.WriteJSONError(w, http.StatusNotFound, err.Error())
}

func ConflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Printf("conflict error: %s\tpath: %s\terror: %s", r.Method, r.URL.Path, err)

	json.WriteJSONError(w, http.StatusConflict, err.Error())
}

func UnauthorizedResponse(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Printf("Unauthorized: %s\tpath: %s\terror: %s", r.Method, r.URL.Path, err)

	json.WriteJSONError(w, http.StatusUnauthorized, err.Error())
}
