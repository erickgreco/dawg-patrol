package users

import (
	"context"
	"errors"
	"time"

	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	queryTimeDuration = time.Second * 5
)

// Contiene la conexion a la db
type UsersStore struct {
	db *pgxpool.Pool
}

func NewUserStore(db *pgxpool.Pool) *UsersStore {
	return &UsersStore{
		db: db,
	}
}

/*
Method to create user, this method only works with DB,
does not know anything bout business logic nor http methods
*/
func (s *UsersStore) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, role, is_active)
		VALUES($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	err := s.db.QueryRow(
		ctx,
		query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.UserRole,
		user.Active,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

/*
	 This method checks if email is already registered in DB,
		if DB error returns false, err
		if not exists returns false, nil
		if exists returns true, nil
*/
func (s *UsersStore) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	err := s.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

/*
	 This method is intented to work only with /login
		that is why it does not returns timestamps
*/
func (s *UsersStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, role, is_active
		FROM users
		WHERE email = $1
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	user := &User{}

	err := s.db.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.UserRole,
		&user.Active,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerrors.ErrDataNotFound
		}
		return nil, err
	}
	return user, nil
}

/*
Method intended to retrieve minimal data from user
*/
func (s *UsersStore) GetSummaryByID(ctx context.Context, id uuid.UUID) (*UserSummary, error) {
	query := `
		SELECT id, username, role, operator_request_status, operator_requested_at
		FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	user := &UserSummary{}

	err := s.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.UserRole,
		&user.RequestStatus,
		&user.RequestedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerrors.ErrDataNotFound
		}
		return nil, err
	}
	return user, nil
}

/*
This method is intented to retrieve all user data,
excluding password hash
*/
func (s *UsersStore) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, username, email, role, is_active, operator_request_status, operator_requested_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	user := &User{}

	err := s.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.UserRole,
		&user.Active,
		&user.RequestStatus,
		&user.RequestedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerrors.ErrDataNotFound
		}
		return nil, err
	}
	return user, nil
}

/*
Creates a request if no previous request was created
*/
func (s *UsersStore) CreateUserRequest(ctx context.Context, id uuid.UUID) (*RoleRequest, error) {
	query := `
		UPDATE users
		SET
			operator_request_status = 'PENDING',
			operator_requested_at = NOW()
		WHERE id = $1
		AND operator_request_status = 'NONE'
		RETURNING operator_request_status, operator_requested_at
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	request := &RoleRequest{}

	err := s.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&request.Status,
		&request.RequestDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerrors.ErrDataNotFound
		}
		return nil, err
	}
	return request, nil
}

/*
Method created to retrieve the first available user with OPERATOR or ADMIN role,
intended to be used by the telemetry seed to obtain valid credentials
without knowing a specific user in advance
*/
func (s *UsersStore) GetFirstOperatorOrAdmin(ctx context.Context) (*User, error) {
	query := `
		SELECT id, username, email, role, is_active
		FROM users
		WHERE role IN ('OPERATOR', 'ADMIN')
		LIMIT 1
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	user := &User{}

	err := s.db.QueryRow(ctx, query).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.UserRole,
		&user.Active,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerrors.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

/*
Method created to persist a refresh token JTI linked to a user.
The JTI is extracted from the JWT claims and stored to enable rotation
and revocation without keeping full token strings in the database.
*/
func (s *UsersStore) CreateRefreshToken(ctx context.Context, tokenID uuid.UUID, userID uuid.UUID) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '7 days')
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	_, err := s.db.Exec(ctx, query, tokenID, userID)

	return err
}

/*
Method created to retrieve a refresh token JTI from the database, validating
that it exists and has not expired. Used by RefreshAccessToken to confirm
the token is still valid before issuing a new pair.
*/
func (s *UsersStore) GetRefreshToken(ctx context.Context, tokenID uuid.UUID) (*StoredRefreshToken, error) {
	query := `
		SELECT id, user_id, expires_at
		FROM refresh_tokens
		WHERE id = $1
		AND expires_at > NOW()
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	stored := &StoredRefreshToken{}

	err := s.db.QueryRow(ctx, query, tokenID).Scan(
		&stored.ID,
		&stored.UserID,
		&stored.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerrors.ErrInvalidRefreshToken
		}
		return nil, err
	}
	return stored, nil
}

/*
Method created to delete a refresh token JTI after it has been consumed.
Part of the rotation strategy — each refresh issues a new JTI and removes
the previous one, preventing token reuse.
*/
func (s *UsersStore) DeleteRefreshToken(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	_, err := s.db.Exec(ctx, query, tokenID)

	return err
}
