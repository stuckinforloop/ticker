package api

import (
	"encoding/json"
	"net/http"

	"github.com/stuckinforloop/ticker/internal/task"
	"go.uber.org/zap"
)

func (a *API) CreateTask(w http.ResponseWriter, r *http.Request) *Response {
	taskDAO := task.NewTaskDAO(a.dao)

	payload := &task.Task{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		a.dao.Logger.Warn("parse req body", zap.Error(err))

		return &Response{
			StatusCode: http.StatusBadRequest,
			Err:        err,
		}
	}

	t, err := taskDAO.CreateTask(r.Context(), payload)
	if err != nil {
		a.dao.Logger.Error("create task", zap.Error(err))

		return &Response{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	return &Response{
		StatusCode: http.StatusCreated,
		Data:       t,
	}
}
