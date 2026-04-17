package users

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	queryTimeDuration = time.Second * 5
)

// Contiene la conexion a la db
type UsersStore struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *UsersStore {
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
		INSERT INTO users (username, email, password_hash, role, is_active, is_verified)
		VALUES($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	err := s.db.QueryRow(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.UserRole,
		user.Active,
		user.Verified,
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
