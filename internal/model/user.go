package model

import "time"

// User represents the users table.
type User struct {
	ID           uint64    `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	Email        *string   `db:"email" json:"email,omitempty"`
	PasswordHash string    `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
