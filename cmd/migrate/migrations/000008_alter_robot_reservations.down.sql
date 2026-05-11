DROP INDEX IF EXISTS idx_robot_reservations_robot_id;
DROP INDEX IF EXISTS idx_robot_reservations_user_id;
DROP INDEX IF EXISTS idx_robot_reservations_expires_at;
DROP INDEX IF EXISTS one_active_reservation_per_user;
DROP INDEX IF EXISTS one_active_reservation_per_robot;
DROP TABLE IF EXISTS robot_reservations;