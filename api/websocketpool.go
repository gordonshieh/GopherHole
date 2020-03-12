package api

import (
	"time"

	"github.com/gorilla/websocket"
)

func toPooledClient(ws *websocket.Conn, pool *WebSocketPool) Client {
	return Client{
		ws,
		pool,
		make(chan []byte, 128),
	}
}

type Client struct {
	conn *websocket.Conn
	pool *WebSocketPool
	send chan []byte
}

func (c *Client) SendToClient() {
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.conn.Close()
				c.pool.unregister <- c
				return
			}
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				c.conn.Close()
				c.pool.unregister <- c
			}
		}
	}
}

type WebSocketPool struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func newWebSocketPool(broadcast chan []byte) *WebSocketPool {
	return &WebSocketPool{
		broadcast:  broadcast,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *WebSocketPool) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
