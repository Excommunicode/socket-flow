package config

import "github.com/ilyakaznacheev/cleanenv"

type (
	ServerConfig struct {
		Port    string `env:"PORT" env-default:"8080"`
		Secret  string `env:"SECRET"`
		BaseURL string `env:"BASE_URL" env-default:"http://localhost:8080"`
		Domain  string `env:"DOMAIN" env-default:"localhost"`
	}

	PGConfig struct {
		DSN string `env:"DSN" env-default:"postgres://test:test@localhost:5432/postgres?sslmode=disable"`
	}

	MongoConfig struct {
		URI        string `env:"MONGO_URI" env-default:"mongodb://localhost:27017"`
		Database   string `env:"MONGO_DB" env-default:"socketflow"`
		Collection string `env:"MONGO_COLLECTION" env-default:"messages"`
	}
	RedisConfig struct {
		Addr     string `env:"REDIS_ADDR" env-default:"localhost:6379"`
		Password string `env:"REDIS_PASSWORD" env-default:""`
		DB       int    `env:"REDIS_DB" env-default:"0"`
	}

	WebSocketConfig struct {
		ReadBufferSize    int    `env:"WS_READ_BUFFER_SIZE" env-default:"1024"`
		WriteBufferSize   int    `env:"WS_WRITE_BUFFER_SIZE" env-default:"1024"`
		AllowedOrigins    string `env:"WS_ALLOWED_ORIGINS" env-default:""`
		EnableCompression bool   `env:"WS_ENABLE_COMPRESSION" env-default:"true"`
	}

	AppConfig struct {
		Postgres  PGConfig
		Mongo     MongoConfig
		Redis     RedisConfig
		Server    ServerConfig
		WebSocket WebSocketConfig
	}
)

func LoadConfig() (*AppConfig, error) {
	var config AppConfig
	if err := cleanenv.ReadEnv(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
