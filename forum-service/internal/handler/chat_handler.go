package handler

import (
	"github.com/rs/zerolog"
	"net/http"
	"time"

	"github.com/Frozz164/forum-app_v2/auth-service/pkg/helper"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/service"
	"github.com/Frozz164/forum-app_v2/forum-service/pkg/websocket"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

type ChatHandler struct {
	chatService service.ChatService
	pool        *websocket.Pool
	jwtSecret   string
	rateLimiter *rate.Limiter
	logger      zerolog.Logger
}

func NewChatHandler(chatService service.ChatService, pool *websocket.Pool, jwtSecret string) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		pool:        pool,
		jwtSecret:   jwtSecret,
		rateLimiter: rate.NewLimiter(rate.Every(time.Second), 1),
		logger:      log.With().Str("component", "chat_handler").Logger(),
	}
}

func (h *ChatHandler) WebsocketHandler(c *gin.Context) {
	logger := h.logger.With().
		Str("method", "WebsocketHandler").
		Str("remote_addr", c.Request.RemoteAddr).
		Logger()

	if !h.rateLimiter.Allow() {
		logger.Warn().Msg("Rate limit exceeded")
		c.AbortWithStatus(http.StatusTooManyRequests)
		return
	}

	conn, err := websocket.Upgrade(c.Writer, c.Request)
	if err != nil {
		logger.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}

	token := c.Query("token")
	var username string
	var userID int64
	var readOnly = true

	if token != "" {
		claims, err := helper.ValidateTokenWithClaims(token, h.jwtSecret)
		if err != nil {
			logger.Warn().Err(err).Str("token_prefix", token[:min(10, len(token))]).Msg("Token validation failed")
		} else {
			readOnly = false
			username = claims.Username
			userID = claims.UserID
		}
	}

	if readOnly {
		username = "Guest_" + generateRandomID()
		logger.Debug().
			Str("username", username).
			Msg("Assigning guest username")
	}

	client := websocket.NewClient(conn, h.pool, username, userID, readOnly)

	h.pool.Register <- client

	logger.Info().
		Str("username", username).
		Int64("user_id", userID).
		Bool("read_only", readOnly).
		Msg("New WebSocket client registered")

	go client.Read(h.chatService)
	go client.Write()
}

func generateRandomID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(b)
}
