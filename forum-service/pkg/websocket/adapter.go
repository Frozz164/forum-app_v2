// pkg/websocket/adapter.go
package websocket

import (
	"context"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/service"
	"time"

	_ "github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
)

type chatServiceAdapter struct {
	service service.ChatService
}

func NewChatServiceAdapter(chatService service.ChatService) *chatServiceAdapter {
	return &chatServiceAdapter{service: chatService}
}

func (a *chatServiceAdapter) GetRecentMessages(ctx context.Context, limit int) ([]Message, error) {
	domainMessages, err := a.service.GetRecentMessages(ctx, limit)
	if err != nil {
		return nil, err
	}

	var wsMessages []Message
	for _, dm := range domainMessages {
		createdAt, _ := time.Parse(time.RFC3339, dm.CreatedAt)
		wsMessages = append(wsMessages, Message{
			Type:      MsgTypeChat,
			Content:   dm.Content,
			Sender:    dm.Username,
			Timestamp: createdAt.Unix(),
			UserID:    dm.UserID,
		})
	}

	return wsMessages, nil
}
