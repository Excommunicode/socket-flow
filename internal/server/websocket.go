package server

import (
	"socket-flow/internal/config"
	socket "socket-flow/pkg/websocket"
	"strings"

	"github.com/gorilla/websocket"
)

func InitWebSocket(cfg config.WebSocketConfig) *websocket.Upgrader {
	var allowedOrigins []string
	if cfg.AllowedOrigins != "" {
		allowedOrigins = strings.Split(cfg.AllowedOrigins, ",")
	}
	return socket.NewUpgrader(
		cfg.ReadBufferSize,
		cfg.WriteBufferSize,
		allowedOrigins,
		cfg.EnableCompression,
	)
}
