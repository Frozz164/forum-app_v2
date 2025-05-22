package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/repository"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ChatServiceImpl struct {
	repo   repository.ChatRepository
	logger zerolog.Logger
}
type ChatService interface {
	ProcessMessage(ctx context.Context, message *domain.Message) error
	GetRecentMessages(ctx context.Context, limit int) ([]*domain.Message, error)
	GetMessageHistory(ctx context.Context, before time.Time, limit int) ([]*domain.Message, error)
}

func NewChatService(repo repository.ChatRepository) ChatService {
	return &ChatServiceImpl{
		repo:   repo,
		logger: log.With().Str("component", "chat_service").Logger(),
	}
}

func (s *ChatServiceImpl) ProcessMessage(ctx context.Context, message *domain.Message) error {
	logger := s.logger.With().
		Str("method", "ProcessMessage").
		Str("username", message.Username).
		Int64("user_id", message.UserID).
		Logger()

	if message == nil {
		err := errors.New("message cannot be nil")
		logger.Error().Err(err).Msg("Validation failed")
		return err
	}

	// Валидация пользователя
	if message.Username == "" || strings.HasPrefix(message.Username, "Guest") {
		err := errors.New("unauthenticated users cannot send messages")
		logger.Warn().Err(err).Msg("Validation failed")
		return err
	}

	// Валидация содержания
	message.Content = strings.TrimSpace(message.Content)
	if message.Content == "" {
		err := errors.New("message content cannot be empty")
		logger.Warn().Err(err).Msg("Validation failed")
		return err
	}
	if len(message.Content) > 500 {
		err := errors.New("message too long (max 500 chars)")
		logger.Warn().Err(err).
			Int("content_length", len(message.Content)).
			Msg("Validation failed")
		return err
	}

	if message.CreatedAt == "" {
		message.CreatedAt = time.Now().Format(time.RFC3339)
		logger.Debug().Msg("Set default timestamp for message")
	}

	err := s.repo.SaveMessage(ctx, message)
	if err != nil {
		logger.Error().Err(err).
			Str("content_prefix", truncateString(message.Content, 20)).
			Msg("Failed to save message in repository")
		return fmt.Errorf("failed to save message: %w", err)
	}

	logger.Info().
		Str("content_prefix", truncateString(message.Content, 20)).
		Msg("Message processed successfully")
	return nil
}

func (s *ChatServiceImpl) GetRecentMessages(ctx context.Context, limit int) ([]*domain.Message, error) {
	logger := s.logger.With().
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

	messages, err := s.repo.GetRecentMessages(ctx, limit)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get recent messages from repository")
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	for i := range messages {
		if messages[i].CreatedAt == "" {
			messages[i].CreatedAt = time.Now().Format(time.RFC3339)
			logger.Debug().
				Int64("message_id", messages[i].ID).
				Msg("Set default timestamp for message")
		}
	}

	logger.Debug().
		Int("message_count", len(messages)).
		Msg("Retrieved recent messages successfully")
	return messages, nil
}

func (s *ChatServiceImpl) GetMessageHistory(ctx context.Context, before time.Time, limit int) ([]*domain.Message, error) {
	logger := s.logger.With().
		Str("method", "GetMessageHistory").
		Time("before", before).
		Int("limit", limit).
		Logger()

	if limit <= 0 {
		limit = 50
		logger.Debug().Msg("Using default limit value")
	}

	messages, err := s.repo.GetMessageHistory(ctx, before, limit)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get message history from repository")
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	logger.Debug().
		Int("message_count", len(messages)).
		Msg("Retrieved message history successfully")
	return messages, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
