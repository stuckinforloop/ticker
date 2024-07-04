-- Create the "groups" table
CREATE TABLE groups (
    id CHAR(26) PRIMARY KEY,
    name TEXT NOT NULL
);

-- Create the "tasks" table
CREATE TABLE tasks (
    id CHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    group_id CHAR(26),
    expression VARCHAR(50) NOT NULL,
    timezone CHAR(3) NOT NULL,
    timeout INT NOT NULL,
    instances INT NOT NULL,
    url TEXT NOT NULL,
    http_method VARCHAR(20) NOT NULL,
    http_headers JSONB,
    post_data JSONB,
    retry_after INT,
    failure_threshold INT NOT NULL,
    notify BOOLEAN NOT NULL,
    notify_every INT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

-- Create an index on the "name" column
CREATE INDEX ON tasks (name);
