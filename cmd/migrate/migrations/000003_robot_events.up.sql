CREATE TYPE robot_event_type AS ENUM ('ROBOT_CONNECTED', 'ROBOT_DISCONNECTED');

CREATE TABLE robot_events (
    id UUID PRIMARY KEY,

    robot_id UUID NOT NULL,
    event robot_event_type NOT NULL,

    issued_by UUID NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_robot_events_robot
        FOREIGN KEY (robot_id)
        REFERENCES robots(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_robot_events_user
        FOREIGN KEY (issued_by)
        REFERENCES users(id)
        ON DELETE SET NULL
);