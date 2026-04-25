CREATE TYPE robot_role AS ENUM ('ASSISTANT', 'SUMO', 'RACER')
CREATE TYPE robot_status AS ENUM ('IDLE', 'IN USE', 'CHARGING', 'OFFLINE')

CREATE TABLE robots (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    role robot_role NOT NULL,
    status TEXT NOT NULL DEFAULT 'idle',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()

);