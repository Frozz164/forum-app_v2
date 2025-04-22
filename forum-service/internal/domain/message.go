package domain

type Message struct {
	ID        int64  `json:"id"`
	Content   string `json:"content" validate:"required,min=1,max=500"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	UserID    int64  `json:"user_id,omitempty"`
}
