package service

import (
	"context"

	"task/internal/model"
	"task/internal/repository"

	"github.com/google/uuid"
)

type TaskService struct {
	tasks *repository.TaskRepository
}

func NewTaskService(tasks *repository.TaskRepository) *TaskService {
	return &TaskService{tasks: tasks}
}

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

func (s *TaskService) Get(ctx context.Context, id []byte) (*model.Task, error) {
	return s.tasks.GetByID(ctx, id)
}
