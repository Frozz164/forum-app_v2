package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
)

type ChatRepository interface {
	SaveMessage(ctx context.Context, message *domain.Message) error
	GetRecentMessages(ctx context.Context, limit int) ([]*domain.Message, error)
}

type ChatRepositoryImpl struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) ChatRepository {
	return &ChatRepositoryImpl{db: db}
}

func (r *ChatRepositoryImpl) SaveMessage(ctx context.Context, message *domain.Message) error {
	query := `
		INSERT INTO messages (content, username, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query,
		message.Content,
		message.Username,
		time.Now(),
	)

	if err != nil {
		log.Printf("Error saving message: %v", err)
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

func (r *ChatRepositoryImpl) GetRecentMessages(ctx context.Context, limit int) ([]*domain.Message, error) {
	query := `
		SELECT id, content, username, created_at
		FROM messages
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var msg domain.Message
		err := rows.Scan(&msg.ID, &msg.Content, &msg.Username, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}
