CREATE INDEX idx_robot_events_robot_id
ON robot_events(robot_id);

CREATE INDEX idx_robot_events_robot_id_created_at
ON robot_events(robot_id, created_at DESC);