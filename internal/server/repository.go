package server

import (
	"socket-flow/internal/config"
	"socket-flow/internal/postgres"
	"socket-flow/internal/repositories"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repositories struct {
	MessageRepository     repositories.MessageRepository
	DeviceTokenRepository repositories.DeviceTokenRepository
}

func InitRepositories(db *postgres.PgClient, mongoClient *mongo.Client, cfg config.MongoConfig) *Repositories {

	messageRepository := repositories.NewMessageRepository(mongoClient, cfg)

	deviceTokenRepository := repositories.NewDeviceTokenRepository(db)

	return &Repositories{
		MessageRepository:     messageRepository,
		DeviceTokenRepository: deviceTokenRepository,
	}
}
