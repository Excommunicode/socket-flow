package server

import (
	"socket-flow/internal/services"
	"socket-flow/internal/websocket"
	"socket-flow/pkg/postgres"
)

type Services struct {
	AuthService    services.AuthService
	MessageService services.MessageService
	UserServices   services.UserService
	Hub            *websocket.Hub
}

func InitServices(transactionManager postgres.Transactor, repositories *Repositories) *Services {
	authService := services.NewAuthService(transactionManager, repositories.UserRepository, repositories.TokenRepository)
	messageService := services.NewMessageService(repositories.MessageRepository)
	userService := services.NewUserService(transactionManager, repositories.UserRepository)
	hub := websocket.NewHub(messageService)
	return &Services{
		AuthService:    authService,
		MessageService: messageService,
		UserServices:   userService,
		Hub:            hub,
	}
}
