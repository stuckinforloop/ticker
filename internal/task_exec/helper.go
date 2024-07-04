package taskexec

func (t *TaskExecDAO) fillDefaults(taskExec *TaskExec) {
	if taskExec.ID == "" {
		id := t.ULIDSource.New(uint64(t.TimeNow()))
		taskExec.ID = id
	}
}

func typePtr[T any](t T) *T {
	return &t
}
