package websocket

import (
	"context"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/service"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	MsgTypeChat   = 1
	MsgTypeSystem = 2
	PingInterval  = 30 * time.Second
	WriteTimeout  = 10 * time.Second
	ReadTimeout   = PingInterval * 2
	MaxMsgSize    = 1024 // 1KB
	BufferSize    = 256  // Размер буфера сообщений
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// В production замените на конкретные домены
		return true
	},
}

type Message struct {
	Type      int    `json:"type"`
	Content   string `json:"content"`
	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
	UserID    int64  `json:"user_id,omitempty"`
}

type Client struct {
	Conn     *websocket.Conn
	Pool     *Pool
	Username string
	UserID   int64
	ReadOnly bool
	Send     chan Message
	mu       sync.Mutex
	done     chan struct{}
}

type Pool struct {
	Register      chan *Client
	Unregister    chan *Client
	Broadcast     chan Message
	Clients       map[*Client]bool
	clientMutex   sync.RWMutex
	MaxMsgSize    int64
	RateLimit     time.Duration
	lastBroadcast time.Time
	ChatService   ChatService // Интерфейс для работы с историей сообщений
}

func NewPool(chatService ChatService) *Pool {
	return &Pool{
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Broadcast:   make(chan Message, BufferSize),
		Clients:     make(map[*Client]bool),
		MaxMsgSize:  MaxMsgSize,
		RateLimit:   100 * time.Millisecond,
		ChatService: chatService,
	}
}

type ChatService interface {
	GetRecentMessages(ctx context.Context, limit int) ([]*domain.Message, error)
}

func (pool *Pool) Start() {
	ticker := time.NewTicker(PingInterval)
	defer ticker.Stop()

	for {
		select {
		case client := <-pool.Register:
			pool.handleNewClient(client)

		case client := <-pool.Unregister:
			pool.handleDisconnect(client)

		case message := <-pool.Broadcast:
			pool.broadcastMessage(message)

		case <-ticker.C:
			pool.sendPingToAll()
		}
	}
}

func (pool *Pool) handleNewClient(client *Client) {
	pool.clientMutex.Lock()
	pool.Clients[client] = true
	pool.clientMutex.Unlock()

	// Отправляем историю сообщений новому клиенту
	go func(c *Client) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		domainMessages, err := pool.ChatService.GetRecentMessages(ctx, 50)
		if err != nil {
			log.Printf("Failed to get message history: %v", err)
			return
		}

		for _, dm := range domainMessages {
			// Преобразуем domain.Message в websocket.Message
			createdAt, _ := time.Parse(time.RFC3339, dm.CreatedAt)
			wsMsg := Message{
				Type:      MsgTypeChat,
				Content:   dm.Content,
				Sender:    dm.Username,
				Timestamp: createdAt.Unix(),
				UserID:    dm.UserID,
			}

			select {
			case c.Send <- wsMsg: // Теперь типы совпадают
			case <-c.done:
				return
			default:
				log.Printf("Client %s send buffer full, skipping history", c.Username)
				return
			}
		}
	}(client)
}

func (pool *Pool) handleDisconnect(client *Client) {
	pool.clientMutex.Lock()
	defer pool.clientMutex.Unlock()

	if _, ok := pool.Clients[client]; ok {
		close(client.done)
		close(client.Send)
		delete(pool.Clients, client)
	}
}

func (pool *Pool) broadcastMessage(message Message) {
	now := time.Now()
	if now.Sub(pool.lastBroadcast) < pool.RateLimit {
		return
	}
	pool.lastBroadcast = now

	pool.clientMutex.RLock()
	clients := make([]*Client, 0, len(pool.Clients))
	for client := range pool.Clients {
		clients = append(clients, client)
	}
	pool.clientMutex.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- message:
		case <-client.done:
		default:
			pool.Unregister <- client
		}
	}
}

func (pool *Pool) sendPingToAll() {
	pool.clientMutex.RLock()
	clients := make([]*Client, 0, len(pool.Clients))
	for client := range pool.Clients {
		clients = append(clients, client)
	}
	pool.clientMutex.RUnlock()

	for _, client := range clients {
		client.mu.Lock()
		client.Conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
		err := client.Conn.WriteMessage(websocket.PingMessage, nil)
		client.mu.Unlock()

		if err != nil {
			pool.Unregister <- client
		}
	}
}

func (c *Client) Read(service.ChatService) {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(c.Pool.MaxMsgSize)
	c.Conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(ReadTimeout))
		return nil
	})

	for {
		var msg Message
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		if c.ReadOnly {
			continue
		}

		if len(msg.Content) == 0 || len(msg.Content) > 500 {
			continue
		}

		msg.Timestamp = time.Now().Unix()
		msg.Sender = c.Username
		msg.UserID = c.UserID

		c.Pool.Broadcast <- msg
	}
}

func (c *Client) Write() {
	ticker := time.NewTicker(PingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.mu.Lock()
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.mu.Unlock()
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
			err := c.Conn.WriteJSON(message)
			c.mu.Unlock()
			if err != nil {
				c.Pool.Unregister <- c
				return
			}

		case <-ticker.C:
			c.mu.Lock()
			c.Conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			c.mu.Unlock()
			if err != nil {
				c.Pool.Unregister <- c
				return
			}

		case <-c.done:
			return
		}
	}
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
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
	}
}
