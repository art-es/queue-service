CREATE TABLE tasks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    queue_name          TEXT NOT NULL,
    payload             TEXT NOT NULL,
    status              TEXT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    locked_until        TIMESTAMPTZ DEFAULT NULL,
    last_fail_duration  INT DEFAULT NULL
);

CREATE INDEX idx_tasks_pop
    ON tasks (queue_name, status, locked_until, created_at);

ALTER TABLE tasks
    ADD CONSTRAINT status_check
    CHECK (status IN ('pending', 'processing', 'failed'));
