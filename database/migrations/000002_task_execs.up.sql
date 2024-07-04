-- Create the "task_execs" table
CREATE TABLE task_execs (
    id CHAR(26) PRIMARY KEY,
    task_id CHAR(26) NOT NULL,
    status VARCHAR(50) NOT NULL,
    run_at BIGINT,
    started_at BIGINT,
    finished_at BIGINT,
    response JSONB,
    created_at BIGINT,
    updated_at BIGINT
);

-- Create unique index on the "task_id" and "run_at" columns
CREATE UNIQUE INDEX ON task_execs (task_id, run_at);
