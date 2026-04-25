package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
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
//? Posible actualizacion para evitar bugs invisibles
func getClaimsFromCtx(r *http.Request) (*Claims, bool) {
claims, ok := r.Context().Value(claimsCtx).(*Claims)
return claims, ok
}
*/
func GetClaimsFromCtx(r *http.Request) *Claims {
	claims, _ := r.Context().Value(claimsCtx).(*Claims)
	return claims
}

func (mw *TokenService) KeyByUserID(r *http.Request) (string, error) {
	claims := GetClaimsFromCtx(r)

	if claims == nil {
		return "unknown", nil
	}

	return claims.ID, nil
}
