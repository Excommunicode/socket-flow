package server

import (
	"socket-flow/internal/config"
	"socket-flow/internal/repositories"
	"socket-flow/pkg/postgres"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repositories struct {
	UserRepository    repositories.UserRepository
	MessageRepository repositories.MessageRepository
	TokenRepository   repositories.TokenRepository
}

func InitRepositories(db postgres.Client, mongoClient *mongo.Client, redisClient *redis.Client,
	cfg config.MongoConfig) *Repositories {

	userRepository := repositories.NewUserRepository(db)
	messageRepository := repositories.NewMessageRepository(mongoClient, cfg)
	tokenRepository := repositories.NewTokenRepository(redisClient)

	return &Repositories{
		UserRepository:    userRepository,
		MessageRepository: messageRepository,
		TokenRepository:   tokenRepository,
	}
}
