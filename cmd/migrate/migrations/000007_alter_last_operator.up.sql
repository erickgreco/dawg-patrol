ALTER TABLE robots
ADD COLUMN last_operator_id UUID,
ADD CONSTRAINT fk_last_operator
FOREIGN KEY (last_operator_id)
REFERENCES users(id)
ON DELETE SET NULL;