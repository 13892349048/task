package repository

import (
	"context"
	"database/sql"

	"task/internal/model"

	"github.com/jmoiron/sqlx"
)

type TaskRepository struct {
	db *sqlx.DB
}

func NewTaskRepository(db *DB) *TaskRepository {
	return &TaskRepository{db: db.SQL}
}

func (r *TaskRepository) Create(ctx context.Context, t *model.Task) error {
	query := `INSERT INTO tasks (id, user_id, title, payload, priority, status, result, retries) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.UserID, t.Title, sql.NullString{String: string(t.Payload), Valid: len(t.Payload) > 0}, t.Priority, t.Status, sql.NullString{String: string(t.Result), Valid: len(t.Result) > 0}, t.Retries)
	return err
}

func (r *TaskRepository) GetByID(ctx context.Context, id []byte) (*model.Task, error) {
	var t model.Task
	query := `SELECT id, user_id, title, payload, priority, status, result, retries, created_at, updated_at FROM tasks WHERE id = ? LIMIT 1`
	if err := r.db.GetContext(ctx, &t, query, id); err != nil {
		return nil, err
	}
	return &t, nil
}
