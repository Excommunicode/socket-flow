package server

import (
	"socket-flow/internal/config"
	"socket-flow/internal/postgres"
	"socket-flow/internal/services"
	"socket-flow/internal/ws"
)

type Services struct {
	AuthService        services.AuthService
	MessageService     services.MessageService
	UserServices       services.UserService
	DeviceTokenService services.DeviceTokenService
	Hub                *ws.Hub
}

func initServices(transactionManager postgres.Transactor, repositories *Repositories, cfg *config.AppConfig) *Services {
	authService := services.NewAuthService(transactionManager, repositories.UserRepository, repositories.TokenRepository)
	messageService := services.NewMessageService(repositories.MessageRepository)
	userService := services.NewUserService(transactionManager, repositories.UserRepository)
	deviceTokenService := services.NewDeviceTokenService(repositories.DeviceTokenRepository)
	notificationService := services.NewNotificationService(
		repositories.DeviceTokenRepository,
		cfg.FCM.ProjectID,
		cfg.FCM.AccessToken,
	)
	hub := ws.NewHub(messageService, notificationService)

	return &Services{
		AuthService:        authService,
		MessageService:     messageService,
		UserServices:       userService,
		DeviceTokenService: deviceTokenService,
		Hub:                hub,
	}
}
