// pkg/websocket/pool.go
package websocket

import (
	"golang.org/x/net/context"
	"log"
	"sync"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/service"
	"github.com/gorilla/websocket"
)

const (
	MsgTypeChat = iota + 1
	MsgTypeSystem
)

type Message struct {
	Type    int    `json:"type"`
	Content string `json:"content"`
	Sender  string `json:"sender"`
}

type Client struct {
	Conn     *websocket.Conn
	Pool     *Pool
	Username string
	ReadOnly bool
	Send     chan Message
}

type Pool struct {
	Register    chan *Client
	Unregister  chan *Client
	Broadcast   chan Message
	Clients     map[*Client]bool
	clientMutex sync.Mutex
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
		Clients:    make(map[*Client]bool),
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.clientMutex.Lock()
			pool.Clients[client] = true
			pool.clientMutex.Unlock()
			log.Printf("New client connected: %s", client.Username)

		case client := <-pool.Unregister:
			pool.clientMutex.Lock()
			if _, ok := pool.Clients[client]; ok {
				delete(pool.Clients, client)
				close(client.Send)
				log.Printf("Client disconnected: %s", client.Username)
			}
			pool.clientMutex.Unlock()

		case message := <-pool.Broadcast:
			pool.clientMutex.Lock()
			for client := range pool.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(pool.Clients, client)
				}
			}
			pool.clientMutex.Unlock()
		}
	}
}

func (c *Client) Read(chatService service.ChatService) {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

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

		err = chatService.ProcessMessage(context.Background(), &domain.Message{
			Content:  msg.Content,
			Username: c.Username,
		})
		if err != nil {
			log.Printf("Failed to process message: %v", err)
			continue
		}

		c.Pool.Broadcast <- Message{
			Type:    MsgTypeChat,
			Content: msg.Content,
			Sender:  c.Username,
		}
	}
}

func (c *Client) Write() {
	defer func() {
		c.Conn.Close()
	}()

	for msg := range c.Send {
		err := c.Conn.WriteJSON(msg)
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			return
		}
	}
}
