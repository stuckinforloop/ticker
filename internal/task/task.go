package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type Timezone string
type Status string

const (
	UTC Timezone = "utc"

	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
	StatusExpired  Status = "expired" // when account is expired
)

type Task struct {
	ID               string         `json:"id"`
	Name             *string        `json:"name"`
	GroupID          *string        `json:"group_id"`
	Expression       string         `json:"expression"`
	Timezone         Timezone       `json:"timezone"`
	Timeout          *int           `json:"timeout"`
	Instances        *int           `json:"instances"`
	URL              string         `json:"url"`
	HTTPMethod       string         `json:"http_method"`
	HTTPHeaders      map[string]any `json:"http_headers"`
	PostData         map[string]any `json:"post_data"`
	RetryAfter       *int           `json:"retry_after"`
	FailureThreshold *int           `json:"failure_threshold"`
	Notify           bool           `json:"notify"`
	NotifyEvery      *int           `json:"notify_every"`
	Status           Status         `json:"status"`
	CreatedAt        int64          `json:"created_at"`
	UpdatedAt        int64          `json:"updated_at"`
}

type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (dao *TaskDAO) CreateTask(ctx context.Context, t *Task) (*Task, error) {
	// fill default values if not provided
	dao.fillDefaults(t)

	t.Status = StatusActive
	t.CreatedAt = dao.TimeNow()
	t.UpdatedAt = dao.TimeNow()

	headers, err := json.Marshal(t.HTTPHeaders)
	if err != nil {
		return nil, fmt.Errorf("marshal http headers: %w", err)
	}

	postData, err := json.Marshal(t.PostData)
	if err != nil {
		return nil, fmt.Errorf("marshal post data: %w", err)
	}

	db := dao.RW()
	query := `
		INSERT INTO tasks (
			id, name, group_id, expression, timezone,
			timeout, instances, url, http_method, http_headers,
			post_data, retry_after, failure_threshold, notify,
			notify_every, status, created_at, updated_at
		) Values (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14,
			$15, $16, $17, $18
		)
	`

	if _, err := db.ExecContext(ctx, query,
		t.ID, t.Name, t.GroupID, t.Expression, t.Timezone,
		t.Timeout, t.Instances, t.URL, t.HTTPMethod, &headers,
		&postData, t.RetryAfter, t.FailureThreshold, t.Notify,
		t.NotifyEvery, t.Status, t.CreatedAt, t.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	return t, nil
}

func (dao *TaskDAO) GetTasks(ctx context.Context) ([]Task, error) {
	db := dao.RO()
	query := `
		SELECT 
			id, name, group_id, expression, timezone,
			timeout, instances, url, http_method, http_headers,
			post_data, retry_after, failure_threshold, notify,
			notify_every, status, created_at, updated_at
		FROM tasks
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get tasks: %w", err)
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		t := Task{}
		headers := []byte{}
		postData := []byte{}
		if err := rows.Scan(
			&t.ID, &t.Name, &t.GroupID, &t.Expression,
			&t.Timezone, &t.Timeout, &t.Instances, &t.URL,
			&t.HTTPMethod, &headers, &postData,
			&t.RetryAfter, &t.FailureThreshold, &t.Notify,
			&t.NotifyEvery, &t.Status, &t.CreatedAt, &t.UpdatedAt,
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

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return tasks, nil
}

func (dao *TaskDAO) GetTask(ctx context.Context, id string) (*Task, error) {
	db := dao.RO()
	query := `
		SELECT 
			id, name, group_id, expression, timezone,
			timeout, instances, url, http_method, http_headers,
			post_data, retry_after, failure_threshold, notify,
			notify_every, status, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	t := Task{}
	headers := []byte{}
	postData := []byte{}
	if err := db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.GroupID, &t.Expression,
		&t.Timezone, &t.Timeout, &t.Instances, &t.URL,
		&t.HTTPMethod, &headers, &postData,
		&t.RetryAfter, &t.FailureThreshold, &t.Notify,
		&t.NotifyEvery, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
	}

	if err := json.Unmarshal(headers, &t.HTTPHeaders); err != nil {
		return nil, fmt.Errorf("unmarshal http headers: %w", err)
	}

	if err := json.Unmarshal(postData, &t.PostData); err != nil {
		return nil, fmt.Errorf("unmarshal post_data: %w", err)
	}

	return &t, nil
}

func (dao *TaskDAO) DeleteTask(ctx context.Context, id string) error {
	db := dao.RW()
	query := `
		DELETE FROM tasks
		WHERE id = $1
	`
	if _, err := db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

func (dao *TaskDAO) UpdateTask(ctx context.Context, t *Task) error {
	headers, err := json.Marshal(t.HTTPHeaders)
	if err != nil {
		return fmt.Errorf("marshal http headers: %w", err)
	}

	postData, err := json.Marshal(t.PostData)
	if err != nil {
		return fmt.Errorf("marshal post data: %w", err)
	}

	t.UpdatedAt = dao.TimeNow()

	db := dao.RW()
	query := `
		UPDATE tasks 
		SET
			name = $1,
			group_id = $2,
			expression = $3,
			timezone = $4,
			timeout = $5,
			instances = $6,
			url = $7,
			http_method = $8,
			http_headers = $9::jsonb,
			post_data = $10::jsonb,
			retry_after = $11,
			failure_threshold = $12,
			notify = $13,
			notify_every = $14,
			status = $15,
			created_at = $16,
			updated_at = $17
		WHERE id = $18
	`
	if _, err := db.ExecContext(ctx, query,
		t.Name,
		t.GroupID,
		t.Expression,
		t.Timezone,
		t.Timeout,
		t.Instances,
		t.URL,
		t.HTTPMethod,
		headers,
		postData,
		t.RetryAfter,
		t.FailureThreshold,
		t.Notify,
		t.NotifyEvery,
		t.Status,
		t.CreatedAt,
		t.UpdatedAt,
		t.ID,
	); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	return nil
}
