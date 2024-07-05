package taskexec

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gitploy-io/cronexpr"
	"github.com/stuckinforloop/ticker/internal/task"
	"go.uber.org/zap"
)

func (dao *TaskExecDAO) ScheduleTasks(ctx context.Context) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
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
	const maxConcurrent = 5
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
			zap.String("task_exec_id", exec.ID))

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

		dedupeID := fmt.Sprintf("%s-%d", t.ID, nextTime.Unix())
		groupID := t.ID
		messageID, err := dao.Queue.Enqueue(ctx, ExecutorQueueName, string(message), typePtr(dedupeID), typePtr(groupID), int64(0))
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
