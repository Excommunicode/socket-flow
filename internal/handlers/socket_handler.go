package handlers

import (
	"log/slog"
	"net/http"
	"socket-flow/internal/auth"
	socket "socket-flow/internal/ws"

	"github.com/gorilla/websocket"
)

type SocketHandler struct {
	hub           *socket.Hub
	upgrader      *websocket.Upgrader
	authenticator *auth.KeycloakAuthenticator
}

func NewSocketHandler(
	hub *socket.Hub,
	upgrader *websocket.Upgrader,
	authenticator *auth.KeycloakAuthenticator,
) *SocketHandler {
	return &SocketHandler{
		hub:           hub,
		upgrader:      upgrader,
		authenticator: authenticator,
	}
}
func (s *SocketHandler) ServeMs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tokenString, ok := auth.BearerToken(r.Header.Get("Authorization"))
	if !ok {
		tokenString = r.URL.Query().Get("access_token")
	}

	user, err := s.authenticator.Validate(ctx, tokenString)
	if err != nil {
		slog.WarnContext(ctx, "websocket authorization failed", "err", err)
		http.Error(w, "Valid Keycloak access token is required", http.StatusUnauthorized)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)

	if err != nil {
		slog.Error("Upgrade error", "err", err)
		return
	}

	s.hub.RegisterClient(conn, user.Subject)
}
