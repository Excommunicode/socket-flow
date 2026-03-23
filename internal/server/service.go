package server

import (
	"socket-flow/internal/config"
	"socket-flow/internal/postgres"
	"socket-flow/internal/services"
	"socket-flow/internal/ws"
	"time"

	"github.com/pkg/errors"
)

type Services struct {
	AuthService        services.AuthService
	MessageService     services.MessageService
	UserServices       services.UserService
	DeviceTokenService services.DeviceTokenService
	MessageScheduler   *services.MessageScheduler
	Hub                *ws.Hub
}

func initServices(transactionManager postgres.Transactor, repositories *Repositories, cfg *config.AppConfig) (*Services, error) {

	authService := services.NewAuthService(transactionManager, repositories.UserRepository, repositories.TokenRepository)
	messageService := services.NewMessageService(repositories.MessageRepository)
	userService := services.NewUserService(transactionManager, repositories.UserRepository)
	deviceTokenService := services.NewDeviceTokenService(repositories.DeviceTokenRepository)
	notificationService := services.NewNotificationService(
		repositories.DeviceTokenRepository,
		cfg.FCM.ProjectID,
		cfg.FCM.AccessToken,
	)
	location, err := time.LoadLocation(cfg.Scheduler.Timezone)
	if err != nil {
		return nil, errors.Wrap(err, "load scheduler timezone")
	}

	schedulerCron := NewCronWithLocation(location)

	messageScheduler, err := services.NewMessageScheduler(
		schedulerCron,
		cfg.Scheduler.TTL,
		cfg.Scheduler.CleanupCron,
		repositories.MessageRepository,
	)
	if err != nil {
		return nil, err
	}

	hub := ws.NewHub(messageService, notificationService)

	return &Services{
		AuthService:        authService,
		MessageService:     messageService,
		UserServices:       userService,
		DeviceTokenService: deviceTokenService,
		MessageScheduler:   messageScheduler,
		Hub:                hub,
	}, nil
}
