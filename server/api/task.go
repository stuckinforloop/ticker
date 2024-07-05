package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gitploy-io/cronexpr"
	"github.com/go-chi/chi/v5"
	"github.com/stuckinforloop/ticker/internal/task"
	"go.uber.org/zap"
)

func validateTask(payload *task.Task) error {
	if payload.Expression == "" {
		return errors.New("expression is required")
	} else {
		if _, err := cronexpr.Parse(payload.Expression); err != nil {
			return fmt.Errorf("invalid expression: %w", err)
		}
	}

	// TODO: validate URL
	if payload.URL == "" {
		return errors.New("url is required")
	}

	if payload.HTTPMethod == "" {
		return errors.New("http method is required")
	}

	return nil
}

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

	if err := validateTask(payload); err != nil {
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

func (a *API) GetTask(w http.ResponseWriter, r *http.Request) *Response {
	taskDAO := task.NewTaskDAO(a.dao)

	id := chi.URLParam(r, "id")
	t, err := taskDAO.GetTask(r.Context(), id)
	if err != nil {
		a.dao.Logger.Error("get task", zap.Error(err))

		return &Response{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	if t == nil {
		return &Response{
			StatusCode: http.StatusNotFound,
			Err:        errors.New("task not found"),
		}
	}

	return &Response{
		StatusCode: http.StatusOK,
		Data:       t,
	}
}

type GetTasksResponse struct {
	Tasks []task.Task `json:"tasks"`
}

func (a *API) GetTasks(w http.ResponseWriter, r *http.Request) *Response {
	taskDAO := task.NewTaskDAO(a.dao)

	tasks, err := taskDAO.GetTasks(r.Context())
	if err != nil {
		a.dao.Logger.Error("get tasks", zap.Error(err))

		return &Response{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	return &Response{
		StatusCode: http.StatusOK,
		Data:       GetTasksResponse{Tasks: tasks},
	}
}

func (a *API) UpdateTask(w http.ResponseWriter, r *http.Request) *Response {
	taskDAO := task.NewTaskDAO(a.dao)

	id := chi.URLParam(r, "id")
	payload := &task.Task{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		a.dao.Logger.Warn("parse req body", zap.Error(err))

		return &Response{
			StatusCode: http.StatusBadRequest,
			Err:        err,
		}
	}
	payload.ID = id

	if err := taskDAO.UpdateTask(r.Context(), payload); err != nil {
		a.dao.Logger.Error("update task", zap.Error(err))

		return &Response{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	return &Response{
		StatusCode: http.StatusOK,
	}
}

func (a *API) DeleteTask(w http.ResponseWriter, r *http.Request) *Response {
	taskDAO := task.NewTaskDAO(a.dao)

	id := chi.URLParam(r, "id")
	if err := taskDAO.DeleteTask(r.Context(), id); err != nil {
		a.dao.Logger.Error("delete task", zap.Error(err))

		return &Response{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	return &Response{
		StatusCode: http.StatusOK,
	}
}
