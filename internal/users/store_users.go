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
		SELECT id, username, role
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
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerrors.ErrDataNotFound
		}
		return nil, err
	}
	return user, nil
}
