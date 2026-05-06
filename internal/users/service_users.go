package users

import (
	"context"
	"errors"

	"github.com/erickgreco/dawg-patrol/internal/auth"
	"github.com/erickgreco/dawg-patrol/internal/domain"
	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UsersRepo interface {
	CreateUser(context.Context, *User) error
	EmailExists(ctx context.Context, email string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetSummaryByID(ctx context.Context, id uuid.UUID) (*UserSummary, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type Service struct {
	store        UsersRepo
	tokenService *auth.TokenService
}

func NewUserService(store UsersRepo, tokenService *auth.TokenService) *Service {
	return &Service{
		store:        store,
		tokenService: tokenService,
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

	hashedpw, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           uuid.New(),
		Username:     data.Username,
		Email:        data.Email,
		PasswordHash: string(hashedpw),
		UserRole:     domain.RoleViewer,
		Active:       true,
	}

	if err := serv.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return &RegisteredUser{
		ID:        user.ID,
		Username:  user.Username,
		UserRole:  user.UserRole,
		CreatedAt: user.CreatedAt,
	}, nil
}

/*
	 This method retrieves userLogIn info from DB, compares hashed password
		with password input, verifies if user is active and responds with a token
		generated with serv.TokenService.Generate
*/
func (serv *Service) UserLogIn(ctx context.Context, data *LoginRequest) (*AuthResponse, error) {
	user, err := serv.store.GetByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, myerrors.ErrDataNotFound) {
			return nil, myerrors.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(data.Password),
	); err != nil {
		return nil, myerrors.ErrInvalidCredentials
	}

	if !user.Active {
		return nil, myerrors.ErrInvalidCredentials
	}

	token, err := serv.tokenService.Generate(user.ID.String(), string(user.UserRole))
	if err != nil {
		return nil, myerrors.ErrTokenGeneration
	}

	return &AuthResponse{
		Token: token,
		ID:    user.ID,
		Role:  user.UserRole,
	}, nil
}

/*
This method is intented to validate status on user actions and allows (if applies)
to perform the action
*/
func (serv *Service) UserProfile(ctx context.Context, id uuid.UUID) (*ProfileResponse, error) {
	user, err := serv.store.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, myerrors.ErrDataNotFound) {
			return nil, myerrors.ErrUserNotFound
		}
		return nil, err
	}

	profile := &Profile{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		UserRole:      user.UserRole,
		Active:        user.Active,
		RequestStatus: user.RequestStatus,
		RequestedAt:   user.RequestedAt,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	return &ProfileResponse{
		Profile: profile,
		Actions: UserActions{
			UpdatePassword:    true,
			UpdateUsername:    true,
			RequestRoleUpdate: user.UserRole == domain.RoleViewer,
		},
	}, nil
}
