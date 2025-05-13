package domain

import _ "time"

type Message struct {
	ID        int64  `json:"id"`
	Content   string `json:"content"`
	Username  string `json:"username"`
	UserID    int64  `json:"user_id,omitempty"`
	CreatedAt string `json:"created_at"` // или time.Time в зависимости от реализации
}
