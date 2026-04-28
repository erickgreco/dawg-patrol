package robots

import (
	"context"
	"time"

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
		INSERT INTO robots (id, name, role, status, battery)
		VALUES($1, $2, $3, $4. $5)
		RETURNING id, name, role, status, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeDuration)
	defer cancel()

	err := r.db.QueryRow(
		ctx,
		query,
		robot.ID,
		robot.Name,
		robot.Role,
		robot.Status,
		robot.Battery,
	).Scan(
		&robot.ID,
		&robot.Name,
		&robot.Role,
		&robot.Status,
		&robot.Battery,
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
		SELECT id, name, role, status, battery
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
			&robot.Role,
			&robot.Status,
			&robot.Battery,
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
