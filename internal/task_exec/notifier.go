package taskexec

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
)

func (dao *TaskExecDAO) UpdateTaskStatusNotify(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			dao.Logger.Info("Attempting dequeue")
			messageID, message, err := dao.Queue.Dequeue(ctx, NotifierQueueName, 30)
			if err != nil {
				dao.Logger.Error("Dequeue failed", zap.Error(err))
				continue // Continue to retry
			}

			if message == "" && messageID == "" {
				dao.Logger.Info("Dequeue: empty results")
				continue
			}

			dao.Logger.Info("Dequeue success",
				zap.String("message", message),
				zap.String("message_id", messageID))

			te := TaskExec{}
			if err := json.Unmarshal([]byte(message), &te); err != nil {
				dao.Logger.Error("unmarshal sqs message", zap.Error(err))
				continue
			}

			if err := dao.updateTaskExec(ctx, &te); err != nil {
				dao.Logger.Error("update task exec: %w",
					zap.Error(err),
					zap.String("task_exec_id", te.ID),
				)
			}

			err = dao.Queue.Acknowledge(ctx, messageID, NotifierQueueName)
			if err != nil {
				dao.Logger.Error("Ack failed", zap.Error(err))
				continue
			}

			dao.Logger.Info("Ack success",
				zap.String("message", message),
				zap.String("message_id", messageID))
		}
	}
}
