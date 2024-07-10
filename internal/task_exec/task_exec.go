package taskexec

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusFailed    Status = "failed"
	StatusCompleted Status = "completed"
)

const (
	ExecutorQueueName = "task-executor-1.fifo"
	NotifierQueueName = "task-status-updates"
)

type TaskExec struct {
	ID         string         `json:"id"`
	TaskID     string         `json:"task_id"`
	Status     Status         `json:"status"`
	RunAt      int64          `json:"run_at"`
	StartedAt  *int64         `json:"started_at"`
	FinishedAt *int64         `json:"finished_at"`
	Response   map[string]any `json:"response"`
	CreatedAt  int64          `json:"created_at"`
	UpdatedAt  int64          `json:"updated_at"`
}

type ExecutorPayload struct {
	TaskID           string         `json:"task_id"`
	TaskExecID       string         `json:"task_exec_id"`
	RunAt            int64          `json:"run_at"`
	Timezone         string         `json:"timezone"`
	Timeout          *int           `json:"timeout"`
	Instances        *int           `json:"instances"`
	URL              string         `json:"url"`
	HTTPMethod       string         `json:"http_method"`
	HTTPHeaders      map[string]any `json:"http_headers"`
	PostData         map[string]any `json:"post_data"`
	RetryAfter       *int           `json:"retry_after"`
	FailureThreshold *int           `json:"failure_threshold"`
	Notify           bool           `json:"notify"`
	NotifyEvery      *int           `json:"notify_every"`
}

func (dao *TaskExecDAO) ListTaskExecs(
	ctx context.Context, taskId string, limit int64, offset int64,
) ([]TaskExec, error) {
	db := dao.RO()
	query := `
		SELECT
			id, task_id, status, run_at, started_at,
			finished_at, response, created_at, updated_at
		FROM task_execs
		WHERE task_id = $1
		ORDER BY run_at DESC
		LIMIT $2
		OFFSET $3
	`
	execs := []TaskExec{}
	rows, err := db.QueryContext(ctx, query, taskId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list task_execs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		t := TaskExec{}
		response := []byte{}
		if err := rows.Scan(
			&t.ID, &t.TaskID, &t.Status, &t.RunAt,
			&t.StartedAt, &t.FinishedAt, &response,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if response != nil {
			if err := json.Unmarshal(response, &t.Response); err != nil {
				return nil, fmt.Errorf("unmarshal response: %w", err)
			}
		}

		execs = append(execs, t)
	}

	return execs, nil
}

func (dao *TaskExecDAO) findTaskExec(ctx context.Context, taskId string, runAt int64) (*TaskExec, error) {
	db := dao.RO()
	query := `
		SELECT id FROM task_execs
		WHERE task_id = $1 AND run_at = $2
	`
	t := TaskExec{}
	if err := db.QueryRowContext(ctx, query, taskId, runAt).Scan(&t.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("find task_exec: %w", err)
	}

	return &t, nil
}

func (dao *TaskExecDAO) createTaskExec(ctx context.Context, t *TaskExec) (*TaskExec, error) {
	dao.fillDefaults(t)

	t.CreatedAt = dao.TimeNow()
	t.UpdatedAt = dao.TimeNow()

	db := dao.RW()
	query := `
		INSERT INTO task_execs (
			id, task_id, status, run_at, created_at, updated_at
		) Values (
			$1, $2, $3, $4, $5, $6
		)
	`

	if _, err := db.ExecContext(ctx, query,
		t.ID, t.TaskID, t.Status, t.RunAt, t.CreatedAt, t.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("create task_exec: %w", err)
	}

	return t, nil
}

func (dao *TaskExecDAO) updateTaskExec(ctx context.Context, t *TaskExec) error {
	updatedAt := dao.TimeNow()

	resp, err := json.Marshal(t.Response)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	db := dao.RW()
	query := `
        UPDATE task_execs
        SET
            status = $1,
            run_at = $2,
            started_at = $3,
            finished_at = $4,
            response = $5,
            updated_at = $6
        WHERE id = $7
    `
	if _, err := db.ExecContext(ctx, query,
		t.Status,
		t.RunAt,
		t.StartedAt,
		t.FinishedAt,
		resp,
		updatedAt,
		t.ID,
	); err != nil {
		return fmt.Errorf("update task exec: %w", err)
	}

	return nil
}

func (p *ExecutorPayload) execute(ctx context.Context, now int64) ([]byte, error) {
	runAt := time.Unix(p.RunAt, 0)
	duration := runAt.Sub(time.Unix(now, 0))
	time.Sleep(duration)

	var req *http.Request
	var err error

	switch p.HTTPMethod {
	case "GET", "HEAD", "DELETE":
		req, err = http.NewRequestWithContext(ctx, p.HTTPMethod, p.URL, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

	case "POST", "PUT", "PATCH":
		reqPayload, err := json.Marshal(p.PostData)
		if err != nil {
			return nil, fmt.Errorf("unmarshal post data: %w", err)
		}

		req, err = http.NewRequestWithContext(ctx, p.HTTPMethod, p.URL, bytes.NewReader(reqPayload))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
	}

	client := http.Client{
		Timeout: time.Duration(time.Second * time.Duration(*p.Timeout)),
		// TODO: add support for sending cookies
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respJSON, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte(`{"response": ""}`), nil
	}

	return respJSON, nil
}
