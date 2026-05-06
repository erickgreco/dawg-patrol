package auth

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/erickgreco/dawg-patrol/internal/domain"
	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
	"github.com/google/uuid"
)

type claimsKey string

const claimsCtx claimsKey = "claims"

func (mw *TokenService) AuthMiddleware(next http.Handler) http.Handler {
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

		claims, err := mw.Validate(tokenString)
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
func GetClaimsFromCtx(r *http.Request) (*Claims, error) {
	claims, ok := r.Context().Value(claimsCtx).(*Claims)
	if !ok || claims == nil {
		return nil, myerrors.ErrInvalidToken
	}
	return claims, nil
}

// Helper created to parse uuid from claimsCtx easily
func (c *Claims) UserID() (uuid.UUID, error) {
	return uuid.Parse(c.Sub)
}

func GetUserIDFromClaimsCtx(r *http.Request) (uuid.UUID, error) {
	claims, err := GetClaimsFromCtx(r)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID()
}

// Helper to get role from claimsCtx
func RoleFromClaimsCtx(r *http.Request) (domain.Role, error) {
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
func (mw *TokenService) RequireRole(allowedRoles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, err := RoleFromClaimsCtx(r)
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
func (mw *TokenService) KeyByUserID(r *http.Request) (string, error) {
	claims, err := GetClaimsFromCtx(r)
	if err != nil || claims == nil {
		return "unknown", nil
	}

	return claims.ID, nil
}
