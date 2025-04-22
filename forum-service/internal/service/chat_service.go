package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/repository"
)

type ChatService interface {
	ProcessMessage(ctx context.Context, message *domain.Message) error
	GetHistory(ctx context.Context, limit int) ([]*domain.Message, error)
}

type ChatServiceImpl struct {
	repo repository.ChatRepository
}

func NewChatService(repo repository.ChatRepository) ChatService {
	return &ChatServiceImpl{repo: repo}
}

func (s *ChatServiceImpl) ProcessMessage(ctx context.Context, message *domain.Message) error {
	// Проверяем, что сообщение от авторизованного пользователя
	if message.Username == "" || strings.HasPrefix(message.Username, "Guest") {
		return errors.New("unauthenticated users cannot send messages")
	}

	// Валидация содержания сообщения
	if strings.TrimSpace(message.Content) == "" {
		return errors.New("message content cannot be empty")
	}

	if len(message.Content) > 500 {
		return errors.New("message too long (max 500 chars)")
	}

	return s.repo.SaveMessage(ctx, message)
}

func (s *ChatServiceImpl) GetHistory(ctx context.Context, limit int) ([]*domain.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50 // default value
	}
	return s.repo.GetRecentMessages(ctx, limit)
}
