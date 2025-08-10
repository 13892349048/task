package cache

import (
	"time"
)

// TaskView is a compact cached view of a task for read endpoints.
type TaskView struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Result    []byte    `json:"result"` // raw JSON bytes (may be nil)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	NotFound  bool      `json:"not_found,omitempty"` // sentinel
}

// Store abstracts multi-level cache operations.
type Store interface {
	GetTaskView(key string) (*TaskView, bool, error)
	SetTaskView(key string, view *TaskView, ttl time.Duration) error
	SetNull(key string, ttl time.Duration) error
	Delete(key string) error
}
