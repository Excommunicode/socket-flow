package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"socket-flow/internal/models"
	"socket-flow/internal/services"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type Hub struct {
	clients map[uuid.UUID]*websocket.Conn
	mu      sync.RWMutex

	msgService services.MessageService

	saveChan chan models.RequestMessage

	notificationService services.NotificationService
}

var errUserConnectionNotFound = errors.New("cannot find the connection with user")

func NewHub(msgService services.MessageService, notifService services.NotificationService) *Hub {

	h := &Hub{
		clients:    make(map[uuid.UUID]*websocket.Conn),
		msgService: msgService,
		saveChan:   make(chan models.RequestMessage, 10000),

		notificationService: notifService,
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
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
			) {
				slog.Warn("unexpected websocket close", "user_id", userId, "err", err)
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

		if err != nil {
			if err == errUserConnectionNotFound {
				go func(toUser uuid.UUID, senderMsg models.RequestMessage) {
					notifCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					title := "Новое сообщение"
					body := senderMsg.Msg

					if err := h.notificationService.SendPushNotification(notifCtx, toUser, title, body); err != nil {
						slog.Error("push notification failed", "to", toUser, "err", err)
					}
				}(msg.To, msg)
			} else {
				slog.Warn("failed to send message", "to", msg.To, "err", err)
			}
		}

		if err != nil && err != errUserConnectionNotFound {
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
		return errUserConnectionNotFound
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		slog.Error("write error", "err", err)

		return errors.Wrap(err, "write message to websocket")
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
