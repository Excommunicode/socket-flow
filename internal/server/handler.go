package server

import (
	"socket-flow/internal/handlers"

	"github.com/gorilla/websocket"
)

type Handler struct {
	AuthHandler    *handlers.AuthHandler
	MessageHandler *handlers.MessageHandler
	UserHandler    *handlers.UserHandler
	SocketHandler  *handlers.SocketHandler
}

func initHandler(services *Services, upgrader *websocket.Upgrader) *Handler {
	authHandler := handlers.NewAuthHandler(services.AuthService)
	messageHandler := handlers.NewMessageHandler(services.MessageService)
	userHandler := handlers.NewUserHandler(services.UserServices)
	socketHandler := handlers.NewSocketHandler(services.Hub, upgrader)

	return &Handler{
		AuthHandler:    authHandler,
		MessageHandler: messageHandler,
		UserHandler:    userHandler,
		SocketHandler:  socketHandler,
	}
}
