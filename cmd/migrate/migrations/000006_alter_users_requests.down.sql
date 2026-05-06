ALTER TABLE users
DROP CONSTRAINT IF EXISTS operator_request_check,
DROP COLUMN IF EXISTS operator_request_status,
DROP COLUMN IF EXISTS operator_requested_at;