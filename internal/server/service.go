package server

import (
	"socket-flow/internal/postgres"
	"socket-flow/internal/services"
	"socket-flow/internal/ws"
)

type Services struct {
	AuthService    services.AuthService
	MessageService services.MessageService
	UserServices   services.UserService
	Hub            *ws.Hub
}

func initServices(transactionManager postgres.Transactor, repositories *Repositories) *Services {
	authService := services.NewAuthService(transactionManager, repositories.UserRepository, repositories.TokenRepository)
	messageService := services.NewMessageService(repositories.MessageRepository)
	userService := services.NewUserService(transactionManager, repositories.UserRepository)
	hub := ws.NewHub(messageService)
	return &Services{
		AuthService:    authService,
		MessageService: messageService,
		UserServices:   userService,
		Hub:            hub,
	}
}
