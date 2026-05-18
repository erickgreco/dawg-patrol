package auth

import (
	"strings"
	"time"

	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Variable created for tokens creation
var Issuer = "dawg-patrol-api"

const refreshTokenExpiry = 7 * 24 * time.Hour

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

// Data used to create token.
// Type is only set for refresh tokens ("refresh"); access tokens leave it empty.
type Claims struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	Type string `json:"type,omitempty"`
	jwt.RegisteredClaims
}

/*
Helper created to parse uuid from claimsCtx easily
Helper used in package middleware
*/
func (c *Claims) UserID() (uuid.UUID, error) {
	return uuid.Parse(c.Sub)
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

/*
This method validates if the input is empty, parses input token with
claims, (internally it verifies jwt secret key, exp and issuer),
validates if token is valid and returns claims(sub + role)
*/
func (t *TokenService) Validate(tokenString string) (*Claims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, myerrors.ErrEmptyToken
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, myerrors.ErrInvalidSigningMethod
		}

		return []byte(t.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, myerrors.ErrInvalidToken
	}

	return claims, nil
}

/*
GenerateRefresh creates a long-lived JWT for use as a refresh token.
Sets type to "refresh" and uses the RegisteredClaims.ID field (JTI) as a unique
identifier stored in DB for rotation and revocation. Returns the signed token
and the JTI so the caller can persist it.
*/
func (t *TokenService) GenerateRefresh(sub string, role string) (string, uuid.UUID, error) {
	jti := uuid.New()
	now := time.Now().UTC()

	claims := Claims{
		Sub:  sub,
		Role: role,
		Type: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTokenExpiry)),
			Issuer:    Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(t.jwtSecret)
	if err != nil {
		return "", uuid.Nil, err
	}
	return ss, jti, nil
}

/*
ValidateRefresh parses and validates a refresh JWT. Rejects tokens that have an
invalid signature, are expired, or do not carry type "refresh" — preventing access
tokens from being used in place of refresh tokens.
*/
func (t *TokenService) ValidateRefresh(tokenString string) (*Claims, error) {
	claims, err := t.Validate(tokenString)
	if err != nil {
		return nil, myerrors.ErrInvalidRefreshToken
	}

	if claims.Type != "refresh" {
		return nil, myerrors.ErrInvalidRefreshToken
	}

	return claims, nil
}
