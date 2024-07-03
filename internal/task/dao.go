package task

import "github.com/stuckinforloop/ticker/internal/dao"

type TaskDAO struct {
	*dao.DAO
}

func NewTaskDAO(dao *dao.DAO) *TaskDAO {
	return &TaskDAO{
		dao,
	}
}
