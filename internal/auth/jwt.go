package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Variable created for tokens creation
var Issuer = "dawg-patrol-api"

type TokenService struct {
	jwtSecret []byte
	expiry    time.Duration
}

func NewTokenService(jwtSecret string, expiry time.Duration) *TokenService {
	return &TokenService{
		jwtSecret: []byte(jwtSecret),
		expiry:    expiry,
	}
}

// Data used to create token
type Claims struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

/*
This method generates a JWT token using the provided subject(userID)
and role as claims, uses a JWT_SECRET key to sign token, JWT_SECRET key
and expiration time are configured in api settings via environment variables
*/
func (t *TokenService) Generate(sub string, role string) (string, error) {
	now := time.Now().UTC()

	claims := Claims{
		Sub:  sub,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.expiry)),
			Issuer:    Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(t.jwtSecret)
	if err != nil {
		return "", err
	}
	return ss, nil
}
