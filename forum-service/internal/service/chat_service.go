package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/repository"
)

type ChatService interface {
	ProcessMessage(ctx context.Context, message *domain.Message) error
	GetRecentMessages(ctx context.Context, limit int) ([]*domain.Message, error)
	GetMessageHistory(ctx context.Context, before time.Time, limit int) ([]*domain.Message, error)
}

type ChatServiceImpl struct {
	repo repository.ChatRepository
}

func NewChatService(repo repository.ChatRepository) ChatService {
	return &ChatServiceImpl{repo: repo}
}

func (s *ChatServiceImpl) ProcessMessage(ctx context.Context, message *domain.Message) error {
	if message == nil {
		return errors.New("message cannot be nil")
	}

	// Валидация пользователя
	if message.Username == "" || strings.HasPrefix(message.Username, "Guest") {
		return errors.New("unauthenticated users cannot send messages")
	}

	// Валидация содержания
	message.Content = strings.TrimSpace(message.Content)
	if message.Content == "" {
		return errors.New("message content cannot be empty")
	}
	if len(message.Content) > 500 {
		return errors.New("message too long (max 500 chars)")
	}

	// Установка времени если не задано
	if message.CreatedAt == "" {
		message.CreatedAt = time.Now().Format(time.RFC3339)
	}

	return s.repo.SaveMessage(ctx, message)
}

func (s *ChatServiceImpl) GetRecentMessages(ctx context.Context, limit int) ([]*domain.Message, error) {
	if limit <= 0 {
		limit = 50
	} else if limit > 1000 {
		limit = 1000
	}

	messages, err := s.repo.GetRecentMessages(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Дополнительная обработка
	for i := range messages {
		if messages[i].CreatedAt == "" {
			messages[i].CreatedAt = time.Now().Format(time.RFC3339)
		}
	}

	return messages, nil
}

func (s *ChatServiceImpl) GetMessageHistory(ctx context.Context, before time.Time, limit int) ([]*domain.Message, error) {
	if limit <= 0 {
		limit = 50
	}

	messages, err := s.repo.GetMessageHistory(ctx, before, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	return messages, nil
}
