ALTER TABLE robots
DROP CONSTRAINT IF EXISTS fk_last_operator,
DROP COLUMN IF EXISTS last_operator_id,