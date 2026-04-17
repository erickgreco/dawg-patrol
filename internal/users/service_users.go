package users

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = errors.New("Error, email already exists")
)

type UsersRepo interface {
	CreateUser(context.Context, *User) error
	EmailExists(context.Context, string) (bool, error)
}

type Service struct {
	store UsersRepo
}

func NewService(store UsersRepo) *Service {
	return &Service{
		store: store,
	}
}

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
		return nil, ErrEmailAlreadyExists
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
		Verified:     false,
	}

	if err := serv.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return &RegisteredUser{
		Username:  user.Username,
		UserRole:  user.UserRole,
		Active:    user.Active,
		Verified:  user.Verified,
		CreatedAt: user.CreatedAt,
	}, nil
}
