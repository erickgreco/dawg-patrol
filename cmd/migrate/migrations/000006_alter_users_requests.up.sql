ALTER TABLE users
ADD COLUMN operator_request_status TEXT NOT NULL DEFAULT 'NONE',
ADD COLUMN operator_requested_at TIMESTAMPTZ,
ADD CONSTRAINT operator_request_check
CHECK (operator_request_status IN ('NONE', 'PENDING', 'APPROVED', 'REJECTED'));