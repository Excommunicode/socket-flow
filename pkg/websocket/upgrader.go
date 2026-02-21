package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

func NewUpgrader(readBuf, writeBuf int, allowedOrigins []string, enableCompression bool) *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  readBuf,
		WriteBufferSize: writeBuf,
		CheckOrigin: func(r *http.Request) bool {
			if len(allowedOrigins) == 0 {
				return true
			}
			origin := r.Header.Get("Origin")
			for _, o := range allowedOrigins {
				if o == origin {
					return true
				}
			}
			return false
		},
		EnableCompression: enableCompression,
	}
}
