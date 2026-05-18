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
Updates robot status to 'IN_USE', in case robot is not IDLE
it returns with an error
*/
func (r *RobotsStore) ReserveRobot(ctx context.Context, reservationID, userID, robotID uuid.UUID) (*RobotReservation, error) {
	query := `
		WITH updated_robot AS (
			UPDATE robots
			SET 
				status = 'IN_USE', last_seen_at = NOW()
			WHERE id = $3
			AND status = 'IDLE'
			RETURNING id, status, last_seen_at
		),
		inserted_reservation AS (
			INSERT INTO robot_reservations (
				id, user_id, robot_id, expires_at, active
			)
			SELECT
				$1, $2, id, NOW() + INTERVAL '30 minutes', TRUE
			FROM updated_robot
			RETURNING
				id, user_id, robot_id, expires_at, active, created_at
		)
		SELECT
    		ir.id, ir.user_id, ir.robot_id, ir.expires_at, ir.active, ir.created_at, ur.status, ur.last_seen_at
		FROM inserted_reservation ir
		JOIN updated_robot ur ON ir.robot_id = ur.id;
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	reservation := &RobotReservation{}

	err := r.db.QueryRow(
		ctx,
		query,
		reservationID,
		userID,
		robotID,
	).Scan(
		&reservation.ID,
		&reservation.UserID,
		&reservation.RobotID,
		&reservation.ExpiresAt,
		&reservation.Active,
		&reservation.CreatedAt,
		&reservation.Status,
		&reservation.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerrors.ErrUnavailableRobot
		}
		return nil, err
	}
	return reservation, nil
}

/*
This method will be periodically executed by a goroutine to clean
expired reservations
*/
func (r *RobotsStore) CleanExpiredReservations(ctx context.Context) error {
	query := `
		WITH expired AS (
			UPDATE robot_reservations
			SET active = FALSE
			WHERE active = TRUE
			AND expires_at <= NOW()
			RETURNING robot_id
		)
		UPDATE robots
		SET status = 'IDLE'
		WHERE id IN (
			SELECT robot_id FROM expired
		);
	`
	_, err := r.db.Exec(ctx, query)

	return err
}

func (r *RobotsStore) ValidateReservation(ctx context.Context, reservationID, userID, robotID uuid.UUID) (uuid.UUID, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM robot_reservations
			WHERE id = $1
				AND user_id = $2
				AND robot_id = $3
				AND active
				AND expires_at > NOW()
		);
	`

	var valid bool

	err := r.db.QueryRow(
		ctx,
		query,
		reservationID,
		userID,
		robotID,
	).Scan(
		&valid,
	)
	if err != nil {
		return uuid.Nil, err
	}

	if !valid {
		return uuid.Nil, myerrors.ErrInvalidReservation
	}

	return reservationID, nil
}

func (r *RobotsStore) GetReservationByID(ctx context.Context, reservationID uuid.UUID) (*RobotReservation, error) {
	query := `
		SELECT id, user_id, robot_id, expires_at, active, created_at
		FROM robot_reservations
		WHERE id = $1
	`

	robotReservation := &RobotReservation{}

	err := r.db.QueryRow(
		ctx,
		query,
		reservationID,
	).Scan(
		&robotReservation.ID,
		&robotReservation.UserID,
		&robotReservation.RobotID,
		&robotReservation.ExpiresAt,
		&robotReservation.Active,
		&robotReservation.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerrors.ErrInvalidReservation
		}
		return nil, err
	}
	return robotReservation, nil
}

func (r *RobotsStore) ExtendReservation(ctx context.Context, reservationID uuid.UUID) error {
	query := `
		UPDATE robot_reservations
		SET expires_at = NOW() + INTERVAL '30 minutes'
		WHERE id = $1
		AND active = TRUE
	`
	_, err := r.db.Exec(ctx, query, reservationID)

	return err
}

/*
Method created to mark when a WS connection is established for a reservation.
Used by CleanNeverConnectedReservations to distinguish active sessions from
reservations that were created but never WS-connected.
*/
func (r *RobotsStore) MarkWSStarted(ctx context.Context, reservationID uuid.UUID) error {
	query := `
		UPDATE robot_reservations
		SET ws_started_at = NOW()
		WHERE id = $1
		AND active = TRUE
	`
	_, err := r.db.Exec(ctx, query, reservationID)

	return err
}

/*
Method created to clean reservations where the robot was reserved but a WS
connection was never started within the 5 minute grace period. Identified by
ws_started_at being NULL, freeing the robot back to IDLE.
*/
func (r *RobotsStore) CleanNeverConnectedReservations(ctx context.Context) error {
	query := `
		WITH abandoned AS (
			UPDATE robot_reservations
			SET active = FALSE
			WHERE active = TRUE
			AND ws_started_at IS NULL
			AND created_at <= NOW() - INTERVAL '5 minutes'
			RETURNING robot_id
		)
		UPDATE robots
		SET status = 'IDLE'
		WHERE id IN (
			SELECT robot_id FROM abandoned
		);
	`
	_, err := r.db.Exec(ctx, query)

	return err
}

/*
Method created to immediately deactivate a reservation and free the robot
when the WS connection closes, avoiding waiting for the periodic cleanup worker.
*/
func (r *RobotsStore) DeactivateReservation(ctx context.Context, reservationID uuid.UUID) error {
	query := `
		WITH deactivated AS (
			UPDATE robot_reservations
			SET active = FALSE
			WHERE id = $1
			AND active = TRUE
			RETURNING robot_id
		)
		UPDATE robots
		SET status = 'IDLE'
		WHERE id IN (
			SELECT robot_id FROM deactivated
		);
	`
	_, err := r.db.Exec(ctx, query, reservationID)

	return err
}
