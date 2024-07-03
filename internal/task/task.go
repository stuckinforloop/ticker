package task

import (
	"context"
	"encoding/json"
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
	Payload          map[string]any `json:"payload"`
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

	payload, err := json.Marshal(t.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	db := dao.RW()
	query := `
		INSERT INTO tasks (
			id, name, group_id, expression, timezone,
			timeout, instances, url, http_method, http_headers,
			payload, retry_after, failure_threshold, notify,
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
		&payload, t.RetryAfter, t.FailureThreshold, t.Notify,
		t.NotifyEvery, t.Status, t.CreatedAt, t.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	return t, nil
}
