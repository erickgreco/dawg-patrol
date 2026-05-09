package robots

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

type RobotsStore struct {
	db *pgxpool.Pool
}

func NewRobotsStore(db *pgxpool.Pool) *RobotsStore {
	return &RobotsStore{
		db: db,
	}
}

// ! As a personal choise, status is set to offline,
// ! so status NEEDS to be updated when using a robot
func (r *RobotsStore) RegisterRobot(ctx context.Context, robot *Robot) error {
	query := `
		INSERT INTO robots (id, serial_number, name, type, battery, status, last_seen_at)
		VALUES($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (serial_number)
		DO UPDATE SET
			battery = EXCLUDED.battery,
			status = EXCLUDED.status,
			last_seen_at = EXCLUDED.last_seen_at,
			updated_at = NOW()

		RETURNING id, serial_number, name, type, battery, status, last_seen_at, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	err := r.db.QueryRow(
		ctx,
		query,
		robot.ID,
		robot.SerialNumber,
		robot.Name,
		robot.Category,
		robot.Battery,
		robot.Status,
		robot.LastSeenAt,
	).Scan(
		&robot.ID,
		&robot.SerialNumber,
		&robot.Name,
		&robot.Category,
		&robot.Battery,
		&robot.Status,
		&robot.LastSeenAt,
		&robot.CreatedAt,
		&robot.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

/*
Currently this method is intended to retrieve all idle
robots from DB since robots are mocked
*/
func (r *RobotsStore) GetIdleRobots(ctx context.Context) ([]*RobotSummary, error) {
	query := `
		SELECT id, name, type, status, battery, last_seen_at
		FROM robots
		WHERE status = 'IDLE'
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var robots []*RobotSummary

	for rows.Next() {
		robot := &RobotSummary{}

		err := rows.Scan(
			&robot.ID,
			&robot.Name,
			&robot.Category,
			&robot.Status,
			&robot.Battery,
			&robot.LastSeenAt,
		)
		if err != nil {
			return nil, err
		}

		robots = append(robots, robot)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return robots, nil
}

/*
Retrieves all unavailable robots, maintenance will be added
*/
func (r *RobotsStore) GetUnavailableRobots(ctx context.Context) ([]*RobotSummary, error) {
	query := `
		SELECT id, name, type, status, battery, last_seen_at
		FROM robots
		WHERE status IN ('IN_USE', 'CHARGING', 'OFFLINE')
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var robots []*RobotSummary

	for rows.Next() {
		robot := &RobotSummary{}

		err := rows.Scan(
			&robot.ID,
			&robot.Name,
			&robot.Category,
			&robot.Status,
			&robot.Battery,
			&robot.LastSeenAt,
		)
		if err != nil {
			return nil, err
		}

		robots = append(robots, robot)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return robots, nil
}

/*
Method created to store robotID in middlewareCtx previous to start ws connection
*/
func (r *RobotsStore) GetByID(ctx context.Context, robotID uuid.UUID) (*RobotSummary, error) {
	query := `
		SELECT id, serial_number, name, type, status, battery, last_seen_at
		FROM robots
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	robot := &RobotSummary{}

	err := r.db.QueryRow(
		ctx,
		query,
		robotID,
	).Scan(
		&robot.ID,
		&robot.SerialNumber,
		&robot.Name,
		&robot.Category,
		&robot.Status,
		&robot.Battery,
		&robot.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerrors.ErrDataNotFound
		}
		return nil, err
	}
	return robot, nil
}

/*
Method to reserve robot previous to start ws conn
*/
func (r *RobotsStore) ReserveRobot(ctx context.Context, robotID, userID uuid.UUID) (*RobotSummary, error) {
	query := `
		UPDATE robots
		SET 
			status = 'IN_USE', 
			last_operator_id = $2, 
			last_seen_at = NOW()
		WHERE id = $1
		AND status = 'IDLE'
		RETURNING id, serial_number, name, type, status, battery, last_seen_at
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	robot := &RobotSummary{}

	err := r.db.QueryRow(
		ctx,
		query,
		robotID,
		userID,
	).Scan(
		&robot.ID,
		&robot.SerialNumber,
		&robot.Name,
		&robot.Category,
		&robot.Status,
		&robot.Battery,
		&robot.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerrors.ErrDataNotFound
		}
		return nil, err
	}
	return robot, nil
}
