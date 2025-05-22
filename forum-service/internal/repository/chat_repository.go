package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ChatRepository interface {
	SaveMessage(ctx context.Context, message *domain.Message) error
	GetRecentMessages(ctx context.Context, limit int) ([]*domain.Message, error)
	GetMessageHistory(ctx context.Context, before time.Time, limit int) ([]*domain.Message, error)
}
type ChatRepositoryImpl struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewChatRepository(db *sql.DB) ChatRepository {
	return &ChatRepositoryImpl{
		db:     db,
		logger: log.With().Str("component", "chat_repository").Logger(),
	}
}

func (r *ChatRepositoryImpl) SaveMessage(ctx context.Context, message *domain.Message) error {
	logger := r.logger.With().
		Str("method", "SaveMessage").
		Str("username", message.Username).
		Int64("user_id", message.UserID).
		Logger()

	query := `
		INSERT INTO messages (content, username, user_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	createdAt := time.Now()
	if message.CreatedAt != "" {
		var err error
		createdAt, err = time.Parse(time.RFC3339, message.CreatedAt)
		if err != nil {
			logger.Warn().Err(err).
				Str("created_at", message.CreatedAt).
				Msg("Failed to parse message timestamp, using current time")
		}
	}

	_, err := r.db.ExecContext(ctx, query,
		message.Content,
		message.Username,
		message.UserID,
		createdAt,
	)

	if err != nil {
		logger.Error().Err(err).
			Str("content_prefix", truncateString(message.Content, 20)).
			Msg("Failed to save message")
		return fmt.Errorf("failed to save message: %w", err)
	}

	logger.Debug().
		Str("content_prefix", truncateString(message.Content, 20)).
		Msg("Message saved successfully")
	return nil
}

func (r *ChatRepositoryImpl) GetRecentMessages(ctx context.Context, limit int) ([]*domain.Message, error) {
	logger := r.logger.With().
		Str("method", "GetRecentMessages").
		Int("limit", limit).
		Logger()

	if limit <= 0 {
		limit = 50
		logger.Debug().Msg("Using default limit value")
	} else if limit > 1000 {
		limit = 1000
		logger.Debug().Msg("Limiting maximum messages to 1000")
	}

	query := `
		SELECT id, content, username, user_id, created_at
		FROM messages
		ORDER BY created_at DESC
		LIMIT $1
	`

	messages, err := r.queryMessages(ctx, query, limit)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get recent messages")
		return nil, err
	}

	logger.Debug().
		Int("message_count", len(messages)).
		Msg("Successfully retrieved recent messages")
	return messages, nil
}

func (r *ChatRepositoryImpl) GetMessageHistory(ctx context.Context, before time.Time, limit int) ([]*domain.Message, error) {
	logger := r.logger.With().
		Str("method", "GetMessageHistory").
		Time("before", before).
		Int("limit", limit).
		Logger()

	if limit <= 0 {
		limit = 50
		logger.Debug().Msg("Using default limit value")
	}

	query := `
		SELECT id, content, username, user_id, created_at
		FROM messages
		WHERE created_at < $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	messages, err := r.queryMessages(ctx, query, before, limit)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get message history")
		return nil, err
	}

	logger.Debug().
		Int("message_count", len(messages)).
		Msg("Successfully retrieved message history")
	return messages, nil
}

func (r *ChatRepositoryImpl) queryMessages(ctx context.Context, query string, args ...interface{}) ([]*domain.Message, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Warn().Err(err).Msg("Failed to close rows")
		}
	}()

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
