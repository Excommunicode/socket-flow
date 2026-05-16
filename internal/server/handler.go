package server

import (
	"socket-flow/internal/auth"
	"socket-flow/internal/handlers"

	"github.com/gorilla/websocket"
)

type Handler struct {
	MessageHandler     *handlers.MessageHandler
	UserHandler        *handlers.UserHandler
	DeviceTokenHandler *handlers.DeviceTokenHandler
	SocketHandler      *handlers.SocketHandler
	Authenticator      *auth.KeycloakAuthenticator
}

func initHandler(
	services *Services,
	upgrader *websocket.Upgrader,
	authenticator *auth.KeycloakAuthenticator,
) *Handler {
	messageHandler := handlers.NewMessageHandler(services.MessageService)
	userHandler := handlers.NewUserHandler()
	deviceTokenHandler := handlers.NewDeviceTokenHandler(services.DeviceTokenService)
	socketHandler := handlers.NewSocketHandler(services.Hub, upgrader, authenticator)

	return &Handler{
		MessageHandler:     messageHandler,
		UserHandler:        userHandler,
		DeviceTokenHandler: deviceTokenHandler,
		SocketHandler:      socketHandler,
		Authenticator:      authenticator,
	}
}
