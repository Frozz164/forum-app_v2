package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
)

type ChatRepository interface {
	SaveMessage(ctx context.Context, message *domain.Message) error
	GetRecentMessages(ctx context.Context, limit int) ([]*domain.Message, error)
	GetMessageHistory(ctx context.Context, before time.Time, limit int) ([]*domain.Message, error)
}

type ChatRepositoryImpl struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) ChatRepository {
	return &ChatRepositoryImpl{db: db}
}

func (r *ChatRepositoryImpl) SaveMessage(ctx context.Context, message *domain.Message) error {
	query := `
		INSERT INTO messages (content, username, user_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	createdAt := time.Now()
	if message.CreatedAt != "" {
		var err error
		createdAt, err = time.Parse(time.RFC3339, message.CreatedAt)
		if err != nil {
			return fmt.Errorf("invalid timestamp format: %w", err)
		}
	}

	_, err := r.db.ExecContext(ctx, query,
		message.Content,
		message.Username,
		message.UserID,
		createdAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

func (r *ChatRepositoryImpl) GetRecentMessages(ctx context.Context, limit int) ([]*domain.Message, error) {
	if limit <= 0 {
		limit = 50
	} else if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, content, username, user_id, created_at
		FROM messages
		ORDER BY created_at DESC
		LIMIT $1
	`

	return r.queryMessages(ctx, query, limit)
}

func (r *ChatRepositoryImpl) GetMessageHistory(ctx context.Context, before time.Time, limit int) ([]*domain.Message, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, content, username, user_id, created_at
		FROM messages
		WHERE created_at < $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	return r.queryMessages(ctx, query, before, limit)
}

func (r *ChatRepositoryImpl) queryMessages(ctx context.Context, query string, args ...interface{}) ([]*domain.Message, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var msg domain.Message
		var createdAt time.Time

		if err := rows.Scan(
			&msg.ID,
			&msg.Content,
			&msg.Username,
			&msg.UserID,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msg.CreatedAt = createdAt.Format(time.RFC3339)
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return messages, nil
}
