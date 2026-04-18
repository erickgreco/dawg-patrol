package users

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Variable created for tokens creation
var Issuer = "dawg-patrol-api"

type UsersRepo interface {
	CreateUser(context.Context, *User) error
	EmailExists(ctx context.Context, email string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type Service struct {
	store UsersRepo
}

func NewService(store UsersRepo) *Service {
	return &Service{
		store: store,
	}
}

// TODO: work with verification, send email for verification
/*
This method verifies if email exists, hashes password with a default cost,
builds user to be stored in DB, calls to CreateUser and
returns registered data to user
*/
func (serv *Service) UserRegistration(ctx context.Context, data *Registration) (*RegisteredUser, error) {
	exists, err := serv.store.EmailExists(ctx, data.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, myerrors.ErrEmailAlreadyExists
	}

	passbytes := []byte(data.Password)

	hashedpw, err := bcrypt.GenerateFromPassword(passbytes, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:     data.Username,
		Email:        data.Email,
		PasswordHash: string(hashedpw),
		UserRole:     RoleViewer,
		Active:       true,
	}

	if err := serv.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return &RegisteredUser{
		Username:  user.Username,
		UserRole:  user.UserRole,
		Active:    user.Active,
		CreatedAt: user.CreatedAt,
	}, nil
}

// llamar a getbyemail -done
// comparar hashed password con user input password -done
// verificar si el usuario esta activo -done
// crear claims
// crear token con metodo de firma
// firmarlo con mi secret
func (serv *Service) UserLogIn(ctx context.Context, data *LoginRequest) (*AuthResponse, error) {
	user, err := serv.store.GetByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, myerrors.ErrDataNotFound) {
			return nil, myerrors.ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(data.Password),
	)
	if err != nil {
		return nil, myerrors.ErrInvalidCredentials
	}

	if user.Active == false {
		return nil, myerrors.ErrInvalidCredentials
	}

	signingKey := []byte(os.Getenv("JWT_SECRET"))

	now := time.Now()

	claims := Claims{
		Sub:  user.ID.String(),
		Role: user.UserRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(signingKey)
	if err != nil {
		return nil, err
	}

	resp := &AuthResponse{
		Token: ss,
	}

	return resp, nil
}
