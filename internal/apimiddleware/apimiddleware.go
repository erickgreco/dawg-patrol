package apimiddleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/erickgreco/dawg-patrol/internal/auth"
	"github.com/erickgreco/dawg-patrol/internal/domain"
	"github.com/erickgreco/dawg-patrol/internal/robots"
	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Middleware struct {
	tokenService *auth.TokenService
	robotService *robots.Service
}

func NewMiddleware(tokenService *auth.TokenService, robotService *robots.Service) *Middleware {
	return &Middleware{
		tokenService: tokenService,
		robotService: robotService,
	}
}

type claimsKey string

const claimsCtx claimsKey = "claims"

type robotKey string

const robotCtx robotKey = "robot"

func (mw *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		claims, err := mw.tokenService.Validate(tokenString)
		if err != nil {
			switch err {
			case myerrors.ErrEmptyToken, myerrors.ErrInvalidToken:
				http.Error(w, "invalid token", http.StatusUnauthorized)
			case myerrors.ErrInvalidSigningMethod:
				http.Error(w, "invalid token signing method", http.StatusUnauthorized)
			default:
				myerrors.InternalServerError(w, r, err)
			}
			return
		}

		ctx := context.WithValue(r.Context(), claimsCtx, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

/*
Func created to retrieve claims from context and
use them without consulting DB
*/
func GetClaimsFromCtx(r *http.Request) (*auth.Claims, error) {
	claims, ok := r.Context().Value(claimsCtx).(*auth.Claims)
	if !ok || claims == nil {
		return nil, myerrors.ErrInvalidToken
	}
	return claims, nil
}

func GetUserIDFromClaimsCtx(r *http.Request) (uuid.UUID, error) {
	claims, err := GetClaimsFromCtx(r)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID()
}

// Helper to get role from claimsCtx
func ValidRoleFromClaimsCtx(r *http.Request) (domain.Role, error) {
	claims, err := GetClaimsFromCtx(r)
	if err != nil {
		return "", err
	}

	role := domain.Role(claims.Role)

	switch role {
	case domain.RoleAdmin, domain.RoleOperator:
		return role, nil
	default:
		return "", myerrors.ErrInvalidUserRole
	}
}

/*
This is a middleware setup to protect routes according to an
specified role
*/
func (mw *Middleware) RequireRole(allowedRoles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, err := ValidRoleFromClaimsCtx(r)
			if err != nil {
				myerrors.UnauthorizedResponse(w, r, err)
				return
			}
			if slices.Contains(allowedRoles, role) {
				next.ServeHTTP(w, r)
				return
			}
			myerrors.ForbiddenResponse(w, r, err)
		})
	}
}

// Key func created for rate limit keys
func (mw *Middleware) KeyByUserID(r *http.Request) (string, error) {
	claims, err := GetClaimsFromCtx(r)
	if err != nil || claims == nil {
		return "unknown", nil
	}

	return claims.Sub, nil
}

func (mw *Middleware) RobotContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		idParam := chi.URLParam(r, "robotID")

		robotID, err := uuid.Parse(idParam)
		if err != nil {
			myerrors.BadRequestResponse(w, r, err)
			return
		}

		ctx := r.Context()

		robot, err := mw.robotService.RobotByID(ctx, robotID)
		if err != nil {
			myerrors.NotFoundResponse(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, robotCtx, robot)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRobotFromCtx(r *http.Request) (*robots.RobotSummary, error) {
	robot, ok := r.Context().Value(robotCtx).(*robots.RobotSummary)
	if !ok || robot == nil {
		return nil, myerrors.ErrUnavailableRobot
	}
	return robot, nil
}

func GetRobotIDFromCtx(r *http.Request) (uuid.UUID, error) {
	robot, err := GetRobotFromCtx(r)
	if err != nil {
		return uuid.Nil, err
	}
	return robot.ID, nil
}
