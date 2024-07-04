package taskexec

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
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

func (dao *TaskExecDAO) ScheduleTasks(ctx context.Context) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				go func() {
					tasks, err := dao.listTasks(ctx)
					if err != nil {
						dao.Logger.Error("list tasks", zap.Error(err))
					}

					dao.enqueueTasks(ctx, tasks)
				}()
			}
		}
	}()

	select {
	case <-ctx.Done():
		done <- true
	}

	return nil
}

func (dao *TaskExecDAO) UpdateTaskStatusNotify(ctx context.Context) error {
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				dao.Logger.Info("Attempting dequeue")
				messageId, message, err := dao.Queue.Dequeue(ctx, NotifierQueueName, 30)
				if err != nil {
					dao.Logger.Error("Dequeue failed", zap.Error(err))
					continue // Continue to retry
				}

				if message == "" && messageId == "" {
					dao.Logger.Info("Dequeue: empty results")
					continue
				}

				dao.Logger.Info("Dequeue success",
					zap.String("message", message),
					zap.String("message_id", messageId))

				// TODO
				// dao.updateTaskStatus(message)
				// dao.notify(message)

				err = dao.Queue.Acknowledge(ctx, messageId, NotifierQueueName)
				if err != nil {
					dao.Logger.Error("Ack failed", zap.Error(err))
				}

				dao.Logger.Info("Ack success",
					zap.String("message", message),
					zap.String("message_id", messageId))
			}
		}
	}()

	select {
	case <-ctx.Done():
		done <- true
	}

	return nil
}

func (dao *TaskExecDAO) listTasks(ctx context.Context) ([]task.Task, error) {
	dao.Logger.Info("Attempting enqueue")

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
		return nil, fmt.Errorf("get tasks: %w", err)
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
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if err := json.Unmarshal(headers, &t.HTTPHeaders); err != nil {
			return nil, fmt.Errorf("unmarshal http headers: %w", err)
		}

		if err := json.Unmarshal(postData, &t.PostData); err != nil {
			return nil, fmt.Errorf("unmarshal post data: %w", err)
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (dao *TaskExecDAO) enqueueTasks(ctx context.Context, tasks []task.Task) {
	maxConcurrent := 5
	semaphore := make(chan bool, maxConcurrent)
	var wg sync.WaitGroup

	for _, t := range tasks {
		wg.Add(1)
		semaphore <- true
		go func(task task.Task) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if err := dao.enqueueTask(ctx, task); err != nil {
				dao.Logger.Error("enqueue task", zap.Error(err))
			}
		}(t)
	}

	wg.Wait()
}

func (dao *TaskExecDAO) enqueueTask(ctx context.Context, t task.Task) error {
	currentTimeUnix := dao.TimeNow()
	currentTime := time.Unix(currentTimeUnix, 0)

	nextTime := cronexpr.MustParse(t.Expression).Next(currentTime)
	if nextTime.Before(currentTime.Add(10 * time.Second)) {
		runAt := nextTime.Unix()

		existingExec, err := dao.findTaskExec(ctx, t.ID, runAt)
		if err != nil {
			return fmt.Errorf("error finding existing task_exec: %w", err)
		}

		if existingExec != nil {
			dao.Logger.Info(
				"Task Execution exists already",
				zap.String("task_id", t.ID),
				zap.Int64("run_at", runAt),
				zap.String("task_exec_id", existingExec.ID))

			return nil
		}

		exec := &TaskExec{
			TaskID: t.ID,
			Status: StatusPending,
			RunAt:  runAt,
		}

		exec, err = dao.createTaskExec(ctx, exec)
		if err != nil {
			return fmt.Errorf("error creating task_exec: %w", err)
		}

		dao.Logger.Info(
			"Create task_exec success",
			zap.String("task_exce_id", exec.ID))

		dedupeID := fmt.Sprintf("%s-%d", t.ID, nextTime.Unix())
		groupID := t.ID
		payload := ExecutorPayload{
			TaskID:           t.ID,
			TaskExecID:       exec.ID,
			RunAt:            runAt,
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

		dao.Logger.Info(
			"Enqueue success",
			zap.Time("current_time", currentTime),
			zap.String("task", t.ID),
			zap.String("task_expression", t.Expression),
			zap.String("message_id", messageID))
	}

	return nil
}

func (dao *TaskExecDAO) findTaskExec(
	ctx context.Context, taskId string, runAt int64,
) (*TaskExec, error) {
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

func (dao *TaskExecDAO) createTaskExec(
	ctx context.Context, t *TaskExec,
) (*TaskExec, error) {
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
