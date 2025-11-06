package models

import "time"

type Message struct {
	ID           int64      `json:"id"`
	AuthorID     *int64     `json:"author_id"`
	Body         string     `json:"body"`          // MVP: 明文；生产可替换为密文
	ReadDuration int        `json:"read_duration"` // seconds
	CreatedAt    time.Time  `json:"created_at"`
	ReadAt       *time.Time `json:"read_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
	Status       int        `json:"status"` // 0 pending,1 claimed
}
