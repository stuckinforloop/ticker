-- Create the "task_execs" table
CREATE TABLE task_execs (
    id CHAR(26) PRIMARY KEY,
    task_id CHAR(26) NOT NULL,
    status VARCHAR(50) NOT NULL,
    run_at BIGINT,
    started_at BIGINT,
    finished_at BIGINT,
    created_at BIGINT,
    updated_at BIGINT
);

-- Create an index on the "task_id" column
CREATE INDEX ON task_execs (task_id);