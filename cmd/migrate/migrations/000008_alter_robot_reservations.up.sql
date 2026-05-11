CREATE TABLE robot_reservations (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    robot_id UUID NOT NULL REFERENCES robots(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL, 
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX one_active_reservation_per_robot
ON robot_reservations(robot_id)
WHERE active = TRUE;

CREATE UNIQUE INDEX one_active_reservation_per_user
ON robot_reservations(user_id)
WHERE active = TRUE;

CREATE INDEX idx_robot_reservations_expires_at
ON robot_reservations(expires_at);

CREATE INDEX idx_robot_reservations_user_id
ON robot_reservations(user_id);

CREATE INDEX idx_robot_reservations_robot_id
ON robot_reservations(robot_id);