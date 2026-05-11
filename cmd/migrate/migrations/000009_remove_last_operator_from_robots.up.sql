ALTER TABLE robots
DROP CONSTRAINT IF EXISTS fk_last_operator;

ALTER TABLE robots
DROP COLUMN IF EXISTS last_operator_id;