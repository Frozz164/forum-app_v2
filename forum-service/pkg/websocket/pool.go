package websocket

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	_ "github.com/Frozz164/forum-app_v2/forum-service/internal/repository"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/service"

	"github.com/gorilla/websocket"
)

const (
	MsgTypeChat   = 1
	MsgTypeSystem = 2
	PingInterval  = 30 * time.Second
	WriteTimeout  = 10 * time.Second
	ReadTimeout   = PingInterval * 2
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
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message, 100),
		Clients:    make(map[*Client]bool),
		MaxMsgSize: 1024, // 1KB
		RateLimit:  100 * time.Millisecond,
	}
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

	go func(c *Client) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		messages, err := service.GetRecentMessages(ctx, 50)
		if err != nil {
			log.Printf("Failed to get message history: %v", err)
			return
		}

		c.mu.Lock()
		defer c.mu.Unlock()

		for _, msg := range messages {
			select {
			case c.Send <- Message{
				Type:      MsgTypeChat,
				Content:   msg.Content,
				Sender:    msg.Username,
				Timestamp: msg.CreatedAt.Unix(),
				UserID:    msg.UserID,
			}:
			case <-c.done:
				return
			default:
				log.Printf("Client %s send buffer full", c.Username)
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
	defer pool.clientMutex.RUnlock()

	for client := range pool.Clients {
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
	defer pool.clientMutex.RUnlock()

	for client := range pool.Clients {
		client.mu.Lock()
		client.Conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
		err := client.Conn.WriteMessage(websocket.PingMessage, nil)
		client.mu.Unlock()

		if err != nil {
			pool.Unregister <- client
		}
	}
}

func (c *Client) Read(chatService service.ChatService) {
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
				return
			}

		case <-ticker.C:
			c.mu.Lock()
			c.Conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			c.mu.Unlock()
			if err != nil {
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
