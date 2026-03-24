package server

import (
	"net/http"
	"slices"
	"socket-flow/internal/config"
	"strings"

	"github.com/gorilla/websocket"
)

func InitWebSocket(cfg config.WebSocketConfig) *websocket.Upgrader {
	var allowedOrigins []string
	if cfg.AllowedOrigins != "" {
		allowedOrigins = strings.Split(cfg.AllowedOrigins, ",")
	}

	return &websocket.Upgrader{
		ReadBufferSize:  cfg.ReadBufferSize,
		WriteBufferSize: cfg.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			if len(allowedOrigins) == 0 {
				return true
			}
			origin := r.Header.Get("Origin")
			return slices.Contains(allowedOrigins, origin)
		},
		EnableCompression: cfg.EnableCompression,
	}
}
