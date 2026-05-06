CREATE TYPE robot_type AS ENUM ('ASSISTANT', 'SUMO', 'RACER');
CREATE TYPE robot_status AS ENUM ('IDLE', 'IN_USE', 'CHARGING', 'OFFLINE');

CREATE TABLE robots (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    type robot_type NOT NULL,
    status robot_status NOT NULL DEFAULT 'OFFLINE',
    battery INTEGER NOT NULL DEFAULT 100 CHECK (battery BETWEEN 0 AND 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_robots_updated_at
BEFORE UPDATE ON robots
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();