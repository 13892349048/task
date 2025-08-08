package model

import (
	"encoding/hex"
	"time"
)

// Task represents the tasks table.
type Task struct {
	ID        []byte    `db:"id" json:"-"`
	UserID    uint64    `db:"user_id" json:"user_id"`
	Title     string    `db:"title" json:"title"`
	Payload   []byte    `db:"payload" json:"-"`
	Priority  int       `db:"priority" json:"priority"`
	Status    string    `db:"status" json:"status"`
	Result    []byte    `db:"result" json:"-"`
	Retries   int       `db:"retries" json:"retries"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// UUIDString returns canonical uuid string from 16-byte id.
func (t *Task) UUIDString() string {
	if len(t.ID) != 16 {
		return ""
	}
	hexStr := hex.EncodeToString(t.ID)
	return hexToUUID(hexStr)
}

func hexToUUID(h string) string {
	if len(h) != 32 {
		return ""
	}
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}
