package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/service"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	MsgTypeChat     = 1
	MsgTypeSystem   = 2
	PingInterval    = 25 * time.Second
	WriteTimeout    = 10 * time.Second
	ReadTimeout     = PingInterval * 2
	MaxMsgSize      = 1024
	BufferSize      = 256
	MaxMessageQueue = 100
	ReconnectDelay  = 3 * time.Second
)

var (
	ErrConnectionClosed = errors.New("connection closed")
	upgrader            = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type Message struct {
	Type      int    `json:"type"`
	Content   string `json:"content"`
	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
	UserID    int64  `json:"user_id,omitempty"`
}

type Client struct {
	Conn      *websocket.Conn
	Pool      *Pool
	Username  string
	UserID    int64
	ReadOnly  bool
	Send      chan Message
	mu        sync.Mutex
	done      chan struct{}
	closeOnce sync.Once
	logger    zerolog.Logger
}

type Pool struct {
	Register    chan *Client
	Unregister  chan *Client
	Broadcast   chan Message
	Clients     map[*Client]bool
	clientMutex sync.RWMutex
	ChatService service.ChatService
	shutdown    chan struct{}
	wg          sync.WaitGroup
	logger      zerolog.Logger
}

func NewPool(chatService service.ChatService) *Pool {
	return &Pool{
		Register:    make(chan *Client, 10),
		Unregister:  make(chan *Client, 10),
		Broadcast:   make(chan Message, MaxMessageQueue),
		Clients:     make(map[*Client]bool),
		ChatService: chatService,
		shutdown:    make(chan struct{}),
		logger:      log.With().Str("component", "websocket_pool").Logger(),
	}
}

func (pool *Pool) Start() {
	defer func() {
		if r := recover(); r != nil {
			pool.logger.Error().
				Interface("recover", r).
				Msg("WebSocket pool recovered from panic")
		}
	}()

	pool.logger.Info().Msg("Starting WebSocket pool")

	ticker := time.NewTicker(PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pool.shutdown:
			pool.logger.Info().Msg("Shutting down WebSocket pool")
			return

		case client := <-pool.Register:
			pool.wg.Add(1)
			go func(c *Client) {
				defer pool.wg.Done()
				pool.handleNewClient(c)
			}(client)

		case client := <-pool.Unregister:
			pool.wg.Add(1)
			go func(c *Client) {
				defer pool.wg.Done()
				pool.handleDisconnect(c)
			}(client)

		case message := <-pool.Broadcast:
			pool.wg.Add(1)
			go func(msg Message) {
				defer pool.wg.Done()
				pool.broadcastMessage(msg)
			}(message)

		case <-ticker.C:
			pool.wg.Add(1)
			go func() {
				defer pool.wg.Done()
				pool.sendPingToAll()
			}()
		}
	}
}

func (pool *Pool) handleNewClient(client *Client) {
	defer func() {
		if r := recover(); r != nil {
			pool.logger.Error().
				Interface("recover", r).
				Str("username", client.Username).
				Int64("user_id", client.UserID).
				Msg("Recovered in handleNewClient")
		}
	}()

	pool.clientMutex.Lock()
	pool.Clients[client] = true
	pool.clientMutex.Unlock()

	pool.logger.Info().
		Str("username", client.Username).
		Int64("user_id", client.UserID).
		Msg("New client connected")

	// Отправка истории сообщений
	go func(c *Client) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		messages, err := pool.ChatService.GetRecentMessages(ctx, 50)
		if err != nil {
			pool.logger.Error().
				Err(err).
				Str("username", c.Username).
				Int64("user_id", c.UserID).
				Msg("Failed to get message history")
			return
		}

		for _, msg := range messages {
			wsMsg := Message{
				Type:      MsgTypeChat,
				Content:   msg.Content,
				Sender:    msg.Username,
				Timestamp: parseTime(msg.CreatedAt).Unix(),
				UserID:    msg.UserID,
			}

			select {
			case c.Send <- wsMsg:
			case <-c.done:
				return
			case <-time.After(100 * time.Millisecond):
				pool.logger.Warn().
					Str("username", c.Username).
					Int64("user_id", c.UserID).
					Msg("Client send timeout")
				return
			}
		}
	}(client)
}

func parseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Now()
	}
	return t
}

func (pool *Pool) handleDisconnect(client *Client) {
	defer func() {
		if r := recover(); r != nil {
			pool.logger.Error().
				Interface("recover", r).
				Str("username", client.Username).
				Int64("user_id", client.UserID).
				Msg("Recovered in handleDisconnect")
		}
	}()

	client.closeOnce.Do(func() {
		close(client.done)
		client.mu.Lock()
		defer client.mu.Unlock()

		if client.Conn != nil {
			err := client.Conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				time.Now().Add(WriteTimeout),
			)
			if err != nil {
				pool.logger.Warn().
					Err(err).
					Str("username", client.Username).
					Int64("user_id", client.UserID).
					Msg("Failed to send close message")
			}
			client.Conn.Close()
		}

		pool.clientMutex.Lock()
		delete(pool.Clients, client)
		pool.clientMutex.Unlock()

		pool.logger.Info().
			Str("username", client.Username).
			Int64("user_id", client.UserID).
			Msg("Client disconnected")
	})
}

func (pool *Pool) broadcastMessage(msg Message) {
	defer func() {
		if r := recover(); r != nil {
			pool.logger.Error().
				Interface("recover", r).
				Msg("Recovered in broadcastMessage")
		}
	}()

	pool.clientMutex.RLock()
	defer pool.clientMutex.RUnlock()

	for client := range pool.Clients {
		select {
		case client.Send <- msg:
		case <-client.done:
		default:
			go func(c *Client) {
				select {
				case pool.Unregister <- c:
				case <-time.After(100 * time.Millisecond):
				}
			}(client)
		}
	}
}

func (pool *Pool) sendPingToAll() {
	defer func() {
		if r := recover(); r != nil {
			pool.logger.Error().
				Interface("recover", r).
				Msg("Recovered in sendPingToAll")
		}
	}()

	pool.clientMutex.RLock()
	defer pool.clientMutex.RUnlock()

	for client := range pool.Clients {
		go func(c *Client) {
			c.mu.Lock()
			defer c.mu.Unlock()

			if err := c.Conn.WriteControl(
				websocket.PingMessage,
				nil,
				time.Now().Add(WriteTimeout),
			); err != nil {
				pool.logger.Warn().
					Err(err).
					Str("username", c.Username).
					Int64("user_id", c.UserID).
					Msg("Failed to send ping")

				select {
				case pool.Unregister <- c:
				case <-time.After(100 * time.Millisecond):
				}
			}
		}(client)
	}
}

func (c *Client) Read(chatService service.ChatService) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error().
				Interface("recover", r).
				Msg("Recovered in client read")
		}
		c.Pool.Unregister <- c
	}()

	c.Conn.SetReadLimit(MaxMsgSize)
	c.Conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(ReadTimeout))
		return nil
	})

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				c.logger.Warn().
					Err(err).
					Msg("WebSocket read error")
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.logger.Warn().
				Err(err).
				Str("message", string(data)).
				Msg("Failed to unmarshal message")
			continue
		}

		if c.ReadOnly || len(msg.Content) == 0 || len(msg.Content) > 500 {
			c.logger.Debug().
				Bool("read_only", c.ReadOnly).
				Int("content_length", len(msg.Content)).
				Msg("Message ignored due to restrictions")
			continue
		}

		msg.Timestamp = time.Now().Unix()
		msg.Sender = c.Username
		msg.UserID = c.UserID

		select {
		case c.Pool.Broadcast <- msg:
			c.logger.Debug().
				Str("content", msg.Content).
				Msg("Message broadcasted")
		case <-time.After(100 * time.Millisecond):
			c.logger.Warn().Msg("Broadcast queue full")
		}
	}
}

func (c *Client) Write() {
	ticker := time.NewTicker(PingInterval)
	defer func() {
		ticker.Stop()
		c.Pool.Unregister <- c
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.mu.Lock()
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, nil)
				c.mu.Unlock()
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
			if err := c.Conn.WriteJSON(msg); err != nil {
				c.logger.Warn().
					Err(err).
					Msg("Failed to write message to WebSocket")
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()

		case <-ticker.C:
			c.mu.Lock()
			c.Conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.Warn().
					Err(err).
					Msg("Failed to send ping")
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()

		case <-c.done:
			return
		}
	}
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warn().
			Err(err).
			Str("remote_addr", r.RemoteAddr).
			Msg("WebSocket upgrade failed")
		return nil, fmt.Errorf("upgrade failed: %w", err)
	}

	log.Info().
		Str("remote_addr", r.RemoteAddr).
		Msg("WebSocket connection upgraded")

	return conn, nil
}

func NewClient(conn *websocket.Conn, pool *Pool, username string, userID int64, readOnly bool) *Client {
	return &Client{
		Conn:     conn,
		Pool:     pool,
		Username: username,
		UserID:   userID,
		ReadOnly: readOnly,
		Send:     make(chan Message, BufferSize),
		done:     make(chan struct{}),
		logger: log.With().
			Str("component", "websocket_client").
			Str("username", username).
			Int64("user_id", userID).
			Logger(),
	}
}
