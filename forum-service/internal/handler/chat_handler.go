package handler

import (
	"context"
	"log"
	"math/rand"
	_ "net/http"
	_ "strconv"
	"time"

	"github.com/Frozz164/forum-app_v2/auth-service/pkg/helper"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/service"
	"github.com/Frozz164/forum-app_v2/forum-service/pkg/websocket"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService service.ChatService
	pool        *websocket.Pool
	jwtSecret   string
}

func NewChatHandler(chatService service.ChatService, pool *websocket.Pool, jwtSecret string) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		pool:        pool,
		jwtSecret:   jwtSecret,
	}
}

func (h *ChatHandler) WebsocketHandler(c *gin.Context) {
	conn, err := websocket.Upgrade(c.Writer, c.Request)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	token := c.Query("token")
	var username string
	var userID int64
	var readOnly = true

	if token != "" {
		if claims, err := helper.ValidateTokenWithClaims(token, h.jwtSecret); err == nil {
			readOnly = false
			username = claims.Username
			userID = claims.UserID
		}
	}

	if readOnly {
		username = "Guest_" + generateRandomID()
	}

	client := &websocket.Client{
		Conn:     conn,
		Pool:     h.pool,
		Username: username,
		UserID:   userID,
		ReadOnly: readOnly,
		Send:     make(chan websocket.Message, 256),
	}

	h.pool.Register <- client

	if messages, err := h.chatService.GetHistory(context.Background(), 50); err == nil {
		for _, msg := range messages {
			client.Send <- websocket.Message{
				Type:    websocket.MsgTypeChat,
				Content: msg.Content,
				Sender:  msg.Username,
			}
		}
	}

	go client.Read(h.chatService)
	go client.Write()
}

func generateRandomID() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
