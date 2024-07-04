package taskexec

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gitploy-io/cronexpr"
	"github.com/stuckinforloop/ticker/internal/task"
	"go.uber.org/zap"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusFailed    Status = "failed"
	StatusCompleted Status = "completed"
)

const ExecutorQueueName = "task-executor-1.fifo"

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

type ExecutorPayload struct {
	TaskID           string         `json:"task_id"`
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

func (dao *TaskExecDAO) ScheduleTasks(ctx context.Context) error {
	for {
		err := dao.EnqueueTasks(ctx)
		if err != nil {
			dao.Logger.Error("enqueue task", zap.Error(err))
		}
		time.Sleep(3 * time.Second)
	}
}

func (dao *TaskExecDAO) EnqueueTasks(ctx context.Context) error {
	dao.Logger.Info(fmt.Sprintf("Attempting enqueue at %s", time.Now().String()))

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

	currentTimeUnix := dao.TimeNow()
	currentTime := time.Unix(currentTimeUnix, 0)

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

		nextTime := cronexpr.MustParse(t.Expression).Next(currentTime)
		if nextTime.Before(currentTime.Add(10 * time.Second)) {
			dedupeID := fmt.Sprintf("%s-%d", t.ID, nextTime.Unix())
			groupID := t.ID
			payload := ExecutorPayload{
				TaskID:           t.ID,
				RunAt:            nextTime.Unix(),
				Timezone:         string(t.Timezone),
				Timeout:          t.Timeout,
				Instances:        t.Instances,
				URL:              t.URL,
				HTTPMethod:       t.HTTPMethod,
				HTTPHeaders:      t.HTTPHeaders,
				PostData:         t.PostData,
				RetryAfter:       t.RetryAfter,
				FailureThreshold: t.FailureThreshold,
				Notify:           t.Notify,
				NotifyEvery:      t.NotifyEvery,
			}

			message, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("marshal sqs message failed: %w", err)
			}

			messageID, err := dao.Queue.Enqueue(
				ctx, ExecutorQueueName, string(message), dedupeID, groupID, int64(0))
			if err != nil {
				return fmt.Errorf("sqs enqueue failed: %w", err)
			}

			fmt.Println("messageID: ", messageID)
		}
	}

	return nil
}
