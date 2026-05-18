package users

import (
	"context"
	"errors"
	"math/rand"

	"github.com/erickgreco/dawg-patrol/internal/auth"
	"github.com/erickgreco/dawg-patrol/internal/domain"
	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	none                  = "NONE"
	pending               = "PENDING"
	approved              = "APPROVED"
	rejected              = "REJECTED"
	pendingResponse       = "REQUEST ALREADY PENDING"
	succesfullResponse    = "REQUEST CREATED SUCCESSFULLY"
	requestRecommendation = "TO OPERATE ROBOTS PLEASE REQUEST A ROLE UPGRADE"
	requestRejected       = "REQUEST REJECTED, FOR MORE INFO PLEASE CONTACT AN ADMIN"
	requestApproved       = "REQUEST APPROVED, PLEASE CHECK COMMANDS INFO BEFORE USING A ROBOT"
	operator              = "USER READY TO OPERATE A ROBOT"
	pendingRequests       = "APPROVAL PENDING REQUESTS WILL BE DISPLAYED HERE"
)

type UsersRepo interface {
	CreateUser(context.Context, *User) error
	EmailExists(ctx context.Context, email string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetSummaryByID(ctx context.Context, id uuid.UUID) (*UserSummary, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	CreateUserRequest(ctx context.Context, id uuid.UUID) (*RoleRequest, error)
	CreateRefreshToken(ctx context.Context, tokenID uuid.UUID, userID uuid.UUID) error
	GetRefreshToken(ctx context.Context, tokenID uuid.UUID) (*StoredRefreshToken, error)
	DeleteRefreshToken(ctx context.Context, tokenID uuid.UUID) error
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
		UserRole:     randomRole(),
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

	resp := &AuthResponse{
		Token: token,
		ID:    user.ID,
		Role:  user.UserRole,
	}

	if user.UserRole == domain.RoleAdmin || user.UserRole == domain.RoleOperator {
		refreshToken, jti, err := serv.tokenService.GenerateRefresh(user.ID.String(), string(user.UserRole))
		if err != nil {
			return nil, myerrors.ErrTokenGeneration
		}
		if err := serv.store.CreateRefreshToken(ctx, jti, user.ID); err != nil {
			return nil, err
		}
		resp.RefreshToken = refreshToken
	}

	return resp, nil
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

	requestRole := RoleRequest{
		Action: user.UserRole == domain.RoleViewer && user.RequestStatus == none,
		Status: none,
	}

	if user.RequestStatus == pending {
		requestRole.Response = pendingResponse
	}

	return &ProfileResponse{
		Profile: profile,
		Actions: UserActions{
			UpdatePassword:    true,
			UpdateUsername:    true,
			RequestRoleUpdate: requestRole,
		},
	}, nil
}

func (serv *Service) UserRoleRequest(ctx context.Context, id uuid.UUID) (*RoleRequest, error) {
	user, err := serv.store.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, myerrors.ErrDataNotFound) {
			return nil, myerrors.ErrUserNotFound
		}
		return nil, err
	}

	if user.UserRole != domain.RoleViewer {
		return nil, myerrors.ErrInvalidUserRole
	}

	if user.RequestStatus == pending {
		return nil, myerrors.ErrPendingRequest
	}

	request, err := serv.store.CreateUserRequest(ctx, user.ID)
	if err != nil {
		if errors.Is(err, myerrors.ErrDataNotFound) {
			return nil, myerrors.ErrUserNotFound
		}
		return nil, err
	}

	return &RoleRequest{
		Action:      false,
		Status:      request.Status,
		RequestDate: request.RequestDate,
		Response:    succesfullResponse,
	}, nil
}

// This helper was created to be able to random add role while creating user (used for seed)
func randomRole() domain.Role {
	roles := []domain.Role{domain.RoleAdmin, domain.RoleOperator, domain.RoleViewer}

	userRole := roles[rand.Intn(len(roles))]

	return userRole
}

func (serv *Service) UserSummaryByRole(ctx context.Context, id uuid.UUID) (*UserSummary, error) {
	user, err := serv.store.GetSummaryByID(ctx, id)
	if err != nil {
		if errors.Is(err, myerrors.ErrDataNotFound) {
			return nil, myerrors.ErrInvalidUserID
		}
		return nil, err
	}

	summary := &UserSummary{
		ID:          user.ID,
		Username:    user.Username,
		UserRole:    user.UserRole,
		RequestedAt: user.RequestedAt,
	}

	switch user.UserRole {
	case domain.RoleViewer:
		summary.RequestStatus = statusCheck(user.RequestStatus)
	case domain.RoleOperator:
		summary.RequestStatus = operator
	case domain.RoleAdmin:
		summary.RequestStatus = pendingRequests
	}
	return summary, nil
}

/*
This helper function was created to asign a personalized
message in requestStatus
*/
func statusCheck(status string) string {
	switch {
	case status == approved:
		return requestApproved
	case status == pending:
		return pendingResponse
	case status == rejected:
		return requestRejected
	default:
		return requestRecommendation
	}
}

/*
RefreshAccessToken validates the incoming refresh JWT, verifies its JTI exists
in DB, rotates the token (delete old JTI, issue new refresh JWT + JTI), and
returns a new access token alongside the rotated refresh token.
Only ADMIN and OPERATOR tokens can reach this path since VIEWER never receives
a refresh token on login.
*/
func (serv *Service) RefreshAccessToken(ctx context.Context, rawToken string) (*AuthResponse, error) {
	claims, err := serv.tokenService.ValidateRefresh(rawToken)
	if err != nil {
		return nil, myerrors.ErrInvalidRefreshToken
	}

	jti, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, myerrors.ErrInvalidRefreshToken
	}

	userID, err := claims.UserID()
	if err != nil {
		return nil, myerrors.ErrInvalidRefreshToken
	}

	stored, err := serv.store.GetRefreshToken(ctx, jti)
	if err != nil {
		return nil, myerrors.ErrInvalidRefreshToken
	}

	user, err := serv.store.GetByID(ctx, userID)
	if err != nil {
		return nil, myerrors.ErrInvalidRefreshToken
	}

	if err := serv.store.DeleteRefreshToken(ctx, stored.ID); err != nil {
		return nil, err
	}

	accessToken, err := serv.tokenService.Generate(user.ID.String(), string(user.UserRole))
	if err != nil {
		return nil, myerrors.ErrTokenGeneration
	}

	newRefreshToken, newJTI, err := serv.tokenService.GenerateRefresh(user.ID.String(), string(user.UserRole))
	if err != nil {
		return nil, myerrors.ErrTokenGeneration
	}

	if err := serv.store.CreateRefreshToken(ctx, newJTI, user.ID); err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
		ID:           user.ID,
		Role:         user.UserRole,
	}, nil
}
