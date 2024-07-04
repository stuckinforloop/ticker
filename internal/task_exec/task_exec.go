package taskexec

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stuckinforloop/ticker/internal/task"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusFailed    Status = "failed"
	StatusCompleted Status = "completed"
)

type TaskExec struct {
	ID         string `json:"id"`
	TaskID     string `json:"task_id"`
	Status     Status `json:"status"`
	RunAt      int64  `json:"run_at"`
	StartedAt  int64  `json:"started_at"`
	FinishedAt int64  `json:"finished_at"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

func (dao *TaskExecDAO) ScheduleTasks(ctx context.Context) error {
	dao.Logger.Info("in ScheduleTasks")

	db := dao.RO()
	query := `
		SELECT
			id, expression, timezone,
			timeout, instances, url, http_method, http_headers,
			post_data, retry_after, failure_threshold, notify,
			notify_every
		FROM tasks
		WHERE status = $1
	`
	rows, err := db.QueryContext(ctx, query, task.StatusActive)
	if err != nil {
		return fmt.Errorf("get tasks: %w", err)
	}
	defer rows.Close()

	tasks := []task.Task{}
	for rows.Next() {
		t := task.Task{}
		headers := []byte{}
		postData := []byte{}
		if err := rows.Scan(
			&t.ID, &t.Expression, &t.Timezone, &t.Timeout,
			&t.Instances, &t.URL, &t.HTTPMethod, &headers, &postData,
			&t.RetryAfter, &t.FailureThreshold, &t.Notify, &t.NotifyEvery,
		); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}

		if err := json.Unmarshal(headers, &t.HTTPHeaders); err != nil {
			return fmt.Errorf("unmarshal http headers: %w", err)
		}

		if err := json.Unmarshal(postData, &t.PostData); err != nil {
			return fmt.Errorf("unmarshal post data: %w", err)
		}

		tasks = append(tasks, t)
	}

	fmt.Println(tasks)

	return nil
}
