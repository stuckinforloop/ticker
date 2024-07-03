-- Create the "groups" table
CREATE TABLE groups (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

-- Create the "tasks" table
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    group_id TEXT,
    expression TEXT NOT NULL,
    timezone TEXT NOT NULL,
    timeout INT NOT NULL,
    instances INT NOT NULL,
    url TEXT NOT NULL,
    http_method TEXT NOT NULL,
    http_headers JSONB,
    payload JSONB,
    retry_after INT,
    failure_threshold INT NOT NULL,
    notify BOOLEAN NOT NULL,
    notify_every INT NOT NULL,
    status TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

-- Create an index on the "name" column
CREATE INDEX ON tasks (name);
