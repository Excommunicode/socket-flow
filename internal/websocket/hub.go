package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"socket-flow/internal/models"
	"socket-flow/internal/services"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Hub struct {
	clients map[uuid.UUID]*websocket.Conn
	mu      sync.RWMutex

	msgService services.MessageService

	saveChan chan models.RequestMessage
}

func NewHub(msgService services.MessageService) *Hub {

	h := &Hub{
		clients:    make(map[uuid.UUID]*websocket.Conn),
		msgService: msgService,
		saveChan:   make(chan models.RequestMessage, 10000),
	}
	go h.runStorageWorker(context.Background())
	return h
}
func (h *Hub) RegisterClient(conn *websocket.Conn, userId uuid.UUID) {
	h.mu.Lock()
	if oldConn, ok := h.clients[userId]; ok {
		oldConn.Close()
	}
	h.clients[userId] = conn
	h.mu.Unlock()

	go h.handleMessages(conn, userId)
}

func (h *Hub) RemoveClient(userId uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conn, ok := h.clients[userId]; ok {
		conn.Close()
		delete(h.clients, userId)
	}
}

func (h *Hub) handleMessages(conn *websocket.Conn, userId uuid.UUID) {
	defer h.RemoveClient(userId)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("read error", "err", err)
			}
			break
		}

		var msg models.RequestMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			slog.Error("json error", "err", err)
			continue
		}

		err = h.SendToUser(msg.To, message)

		msg.IsDelivered = err == nil

		if err != nil && err.Error() != "cannot find the connection with user" {
			slog.Warn("failed to send message", "to", msg.To, "err", err)
		}

		select {
		case h.saveChan <- msg:
		default:
			slog.Warn("save queue full, dropping message")
		}
	}
}

func (h *Hub) SendToUser(userId uuid.UUID, data []byte) error {
	h.mu.RLock()
	conn, ok := h.clients[userId]
	h.mu.RUnlock()

	if !ok {
		return fmt.Errorf("cannot find the connection with user")
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		slog.Error("write error", "err", err)
		return err
	}
	return nil
}

func (h *Hub) runStorageWorker(ctx context.Context) {
	for {
		select {
		case msg := <-h.saveChan:
			saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := h.msgService.CreateMessage(saveCtx, msg); err != nil {
				slog.Error("async save message error", "err", err)
			}
			cancel()

		case <-ctx.Done():
			slog.Info("storage worker stopped")
			return
		}
	}
}
