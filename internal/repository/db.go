package repository

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// DB wraps sqlx.DB to allow injection/mocking if needed.
type DB struct {
	SQL *sqlx.DB
}

func NewDB(dsn string, maxOpen, maxIdle int, maxLife time.Duration) (*DB, error) {
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(maxLife)
	if err := ping(db.DB); err != nil {
		return nil, err
	}
	return &DB{SQL: db}, nil
}

func ping(db *sql.DB) error {
	deadline := time.Now().Add(5 * time.Second)
	for {
		if err := db.Ping(); err != nil {
			if time.Now().After(deadline) {
				return err
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		return nil
	}
}
