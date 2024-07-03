package task

func (t *TaskDAO) fillDefaults(task *Task) {
	if task.ID == "" {
		id := t.ULIDSource.New(uint64(t.TimeNow()))
		task.ID = id
	}

	if task.Name == nil || *task.Name == "" {
		task.Name = &task.ID
	}

	if task.Timezone == "" {
		task.Timezone = UTC
	}

	if task.Timeout == nil {
		task.Timeout = typePtr(3600)
	}

	if task.Instances == nil {
		task.Instances = typePtr(0)
	}

	if task.FailureThreshold == nil {
		task.FailureThreshold = typePtr(20)
	}

	if task.NotifyEvery == nil {
		task.NotifyEvery = typePtr(0)
	}
}

func typePtr[T any](t T) *T {
	return &t
}
