package taskexec

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// TODO: add retry logic for tasks with non-zero retry value
func (dao *TaskExecDAO) ExecuteTasks(ctx context.Context) error {
	maxConcurrent := 20
	semaphore := make(chan bool, maxConcurrent)
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return nil
		default:
			dao.Logger.Info("Attempting dequeue")
			messageID, message, err := dao.Queue.Dequeue(ctx, ExecutorQueueName, 30)
			if err != nil {
				dao.Logger.Error("Dequeue failed", zap.Error(err))
				continue
			}

			if message == "" && messageID == "" {
				dao.Logger.Info("Dequeue: empty results")
				continue
			}

			select {
			case semaphore <- true:
				wg.Add(1)
				go func(messageID, message string) {
					defer wg.Done()
					defer func() { <-semaphore }()
					if err := dao.executeTask(ctx, messageID, message); err != nil {
						dao.Logger.Error("execute task", zap.Error(err))
					}
				}(messageID, message)
			case <-ctx.Done():
				return nil
			}
		}
	}
}

func (dao *TaskExecDAO) executeTask(ctx context.Context, messageID, message string) error {
	execPayload := ExecutorPayload{}
	if err := json.Unmarshal([]byte(message), &execPayload); err != nil {
		return fmt.Errorf("unmarshal sqs message failed: %w", err)
	}

	now := dao.TimeNow()
	te := TaskExec{
		ID:         execPayload.TaskExecID,
		TaskID:     execPayload.TaskID,
		Status:     StatusRunning,
		RunAt:      execPayload.RunAt,
		StartedAt:  typePtr(now),
		FinishedAt: nil,
		Response:   nil,
	}
	sqsMsg, err := json.Marshal(te)
	if err != nil {
		return fmt.Errorf("marshal sqs message failed: %w", err)
	}

	if _, err := dao.Queue.Enqueue(ctx, NotifierQueueName, string(sqsMsg), nil, nil, int64(0)); err != nil {
		return fmt.Errorf("sqs enqueue before execution failed (%s): %w", te.ID, err)
	}

	resp, err := execPayload.execute(ctx, now)
	if err != nil {
		return fmt.Errorf("execute task: %w", err)
	}

	if err := json.Unmarshal(resp, &te.Response); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	te.Status = StatusCompleted
	te.FinishedAt = typePtr(dao.TimeNow())

	_, err = dao.Queue.Enqueue(ctx, NotifierQueueName, string(sqsMsg), nil, nil, int64(0))
	if err != nil {
		return fmt.Errorf("sqs enqueue after execution failed (%s): %w", te.ID, err)
	}

	if err := dao.Queue.Acknowledge(ctx, messageID, ExecutorQueueName); err != nil {
		return fmt.Errorf("sqs acknowledge failed: %w", err)
	}

	return nil
}
