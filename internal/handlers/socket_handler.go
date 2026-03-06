package handlers

import (
	"log/slog"
	"net/http"
	socket "socket-flow/internal/ws"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type SocketHandler struct {
	hub      *socket.Hub
	upgrader *websocket.Upgrader
}

func NewSocketHandler(hub *socket.Hub, upgrader *websocket.Upgrader) *SocketHandler {
	return &SocketHandler{
		hub:      hub,
		upgrader: upgrader,
	}
}
func (s *SocketHandler) ServeMs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIdRaw := r.URL.Query().Get("userId")
	userId, err := uuid.Parse(userIdRaw)

	if err != nil {
		slog.ErrorContext(ctx, "cannot parse uuid", "err", err, "input", userIdRaw)
		http.Error(w, "Valid userId is required", http.StatusBadRequest)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)

	if err != nil {
		slog.Error("Upgrade error", "err", err)
		return
	}

	s.hub.RegisterClient(conn, userId)
}
