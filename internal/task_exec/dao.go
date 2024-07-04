package taskexec

import "github.com/stuckinforloop/ticker/internal/dao"

type TaskExecDAO struct {
	*dao.DAO
}

func NewTaskExecDAO(dao *dao.DAO) *TaskExecDAO {
	return &TaskExecDAO{
		dao,
	}
}
