package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	sshpkg "github.com/matras34/UpworkNotifier/internal/ssh"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 1024 // 1MB
)

type MessageType string

const (
	MessageTypeData        MessageType = "data"
	MessageTypeResize      MessageType = "resize"
	MessageTypeHostKey     MessageType = "hostkey"
	MessageTypeError       MessageType = "error"
	MessageTypeClose       MessageType = "close"
	MessageTypeConnected   MessageType = "connected"
)

type Message struct {
	Type MessageType `json:"type"`
	Data interface{} `json:"data"`
}

type ResizeData struct {
	Rows int `json:"rows"`
	Cols int `json:"cols"`
}

type Client struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	conn        *websocket.Conn
	send        chan []byte
	hub         *Hub
	sshSession  *sshpkg.Session
	mu          sync.RWMutex
	ctx         context.Context
	cancelFunc  context.CancelFunc
}

func NewClient(userID uuid.UUID, conn *websocket.Conn, hub *Hub) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		ID:         uuid.New(),
		UserID:     userID,
		conn:       conn,
		send:       make(chan []byte, 256),
		hub:        hub,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func (c *Client) SetSSHSession(session *sshpkg.Session) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sshSession = session
}

func (c *Client) GetSSHSession() *sshpkg.Session {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sshSession
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		c.cancelFunc()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("websocket error: %v\n", err)
			}
			break
		}

		if err := c.handleMessage(message); err != nil {
			c.SendError(err.Error())
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) SSHReadPump() {
	session := c.GetSSHSession()
	if session == nil {
		return
	}

	buf := make([]byte, 8192)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			n, err := session.Read(buf)
			if err != nil {
				if err.Error() != "EOF" {
					c.SendError(fmt.Sprintf("SSH read error: %v", err))
				}
				return
			}

			if n > 0 {
				msg := Message{
					Type: MessageTypeData,
					Data: string(buf[:n]),
				}
				c.SendMessage(msg)
			}
		}
	}
}

func (c *Client) handleMessage(message []byte) error {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	session := c.GetSSHSession()
	if session == nil && msg.Type != MessageTypeHostKey {
		return fmt.Errorf("no active SSH session")
	}

	switch msg.Type {
	case MessageTypeData:
		data, ok := msg.Data.(string)
		if !ok {
			return fmt.Errorf("invalid data type")
		}
		if session != nil {
			_, err := session.Write([]byte(data))
			return err
		}

	case MessageTypeResize:
		dataMap, ok := msg.Data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid resize data")
		}

		rows, _ := dataMap["rows"].(float64)
		cols, _ := dataMap["cols"].(float64)

		if session != nil {
			return session.Resize(int(rows), int(cols))
		}

	case MessageTypeClose:
		if session != nil {
			session.Close("user requested")
		}
		c.cancelFunc()
	}

	return nil
}

func (c *Client) SendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case c.send <- data:
	default:
		close(c.send)
		c.hub.unregister <- c
	}
}

func (c *Client) SendError(errMsg string) {
	c.SendMessage(Message{
		Type: MessageTypeError,
		Data: errMsg,
	})
}

func (c *Client) SendConnected() {
	c.SendMessage(Message{
		Type: MessageTypeConnected,
		Data: "SSH connection established",
	})
}

func (c *Client) Close() {
	if c.sshSession != nil {
		c.sshSession.Close("websocket closed")
	}
	c.cancelFunc()
	close(c.send)
}
