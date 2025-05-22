package websocket

import (
	"context"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/service"
	"github.com/rs/zerolog/log"
	"time"
)

type chatServiceAdapter struct {
	service service.ChatService
}

func NewChatServiceAdapter(chatService service.ChatService) *chatServiceAdapter {
	return &chatServiceAdapter{
		service: chatService,
	}
}

func (a *chatServiceAdapter) GetRecentMessages(ctx context.Context, limit int) ([]Message, error) {
	logger := log.With().
		Str("method", "GetRecentMessages").
		Int("limit", limit).
		Logger()

	domainMessages, err := a.service.GetRecentMessages(ctx, limit)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to get recent messages from chat service")
		return nil, err
	}

	var wsMessages []Message
	for _, dm := range domainMessages {
		createdAt, err := time.Parse(time.RFC3339, dm.CreatedAt)
		if err != nil {
			logger.Warn().
				Err(err).
				Str("created_at", dm.CreatedAt).
				Msg("Failed to parse message timestamp")
			continue
		}

		wsMessages = append(wsMessages, Message{
			Type:      MsgTypeChat,
			Content:   dm.Content,
			Sender:    dm.Username,
			Timestamp: createdAt.Unix(),
			UserID:    dm.UserID,
		})
	}

	logger.Debug().
		Int("message_count", len(wsMessages)).
		Msg("Successfully converted domain messages to WebSocket messages")

	return wsMessages, nil
}
