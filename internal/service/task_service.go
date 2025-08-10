package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"task/internal/cache"
	"task/internal/model"
	"task/internal/repository"
	"task/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TaskService struct {
	tasks *repository.TaskRepository
	cache cache.Store
	cfg   struct {
		cacheTTL int64 // seconds
		nullTTL  int64 // seconds
	}
}

func NewTaskService(tasks *repository.TaskRepository) *TaskService {
	return &TaskService{tasks: tasks}
}

func NewTaskServiceWithCache(tasks *repository.TaskRepository, store cache.Store, cacheTTLSeconds, nullTTLSeconds int64) *TaskService {
	s := &TaskService{tasks: tasks, cache: store}
	s.cfg.cacheTTL = cacheTTLSeconds
	s.cfg.nullTTL = nullTTLSeconds
	return s
}

// Create keeps same behavior as before.
func (s *TaskService) Create(ctx context.Context, userID uint64, title string, payload []byte, priority int) (string, error) {
	uid := uuid.New()
	id := uid[:]
	t := &model.Task{
		ID:       id,
		UserID:   userID,
		Title:    title,
		Payload:  payload,
		Priority: priority,
		Status:   "queued",
		Retries:  0,
	}
	if err := s.tasks.Create(ctx, t); err != nil {
		return "", err
	}
	return uid.String(), nil
}

// Get returns task and a source hint: "local" | "redis" | "db".
func (s *TaskService) Get(ctx context.Context, id []byte) (*model.Task, string, error) {
	if len(id) != 16 {
		return nil, "", fmt.Errorf("invalid id")
	}
	uid, _ := uuid.FromBytes(id)
	key := "task:" + uid.String()

	if s.cache != nil {
		if tv, ok, src, err := s.cache.GetTaskView(key); err == nil && ok {
			if tv.NotFound {
				return nil, src, fmt.Errorf("not found")
			}
			mt := &model.Task{ID: id, Status: tv.Status, Result: tv.Result, CreatedAt: tv.CreatedAt, UpdatedAt: tv.UpdatedAt}
			return mt, src, nil
		} else if err != nil {
			logger.L().Warn("cache get error", zap.String("key", key), zap.Error(err))
		}
	}

	// DB fallback
	t, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		if s.cache != nil {
			_ = s.cache.SetNull(key, seconds(s.cfg.nullTTL))
		}
		return nil, "db", err
	}

	// populate cache
	if s.cache != nil {
		var resultCopy []byte
		if len(t.Result) > 0 {
			if json.Valid(t.Result) {
				resultCopy = append([]byte(nil), t.Result...)
			}
		}
		_ = s.cache.SetTaskView(key, &cache.TaskView{
			TaskID:    uid.String(),
			Status:    t.Status,
			Result:    resultCopy,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		}, seconds(s.cfg.cacheTTL))
	}
	return t, "db", nil
}

// helpers to avoid importing zap types here
func zapString(k, v string) interface{} { return struct{ K, V string }{K: k, V: v} }
func zapError(err error) interface{}    { return struct{ Error error }{Error: err} }
func seconds(n int64) (d time.Duration) { return time.Duration(n) * time.Second }
