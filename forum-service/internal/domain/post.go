package domain

type Post struct {
	ID        int64  `json:"id"`
	Title     string `json:"title" validate:"required,min=3,max=100"`
	Content   string `json:"content" validate:"required,min=10"`
	AuthorID  int64  `json:"author_id"`
	CreatedAt string `json:"created_at"`
	Author    string `json:"author"` // Добавлено для фронтенда
}
