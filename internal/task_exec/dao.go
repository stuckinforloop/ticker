package taskexec

import (
	"github.com/stuckinforloop/ticker/internal/dao"
	"github.com/stuckinforloop/ticker/internal/queue"
)

type TaskExecDAO struct {
	*dao.DAO
	Queue queue.Queue
}

func NewTaskExecDAO(dao *dao.DAO) *TaskExecDAO {
	queue := queue.New()
	return &TaskExecDAO{
		dao,
		queue,
	}
}
