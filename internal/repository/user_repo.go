package repository

import (
	"context"

	"task/internal/model"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db.SQL}
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	query := `INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, u.Username, u.Email, u.PasswordHash)
	return err
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	query := `SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = ? LIMIT 1`
	if err := r.db.GetContext(ctx, &u, query, username); err != nil {
		return nil, err
	}
	return &u, nil
}
