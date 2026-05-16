package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pkg/errors"
)

type (
	ServerConfig struct {
		Port    string `env:"PORT"     env-default:"8080"`
		BaseURL string `env:"BASE_URL" env-default:"http://localhost:8080"`
		Domain  string `env:"DOMAIN"   env-default:"localhost"`
	}

	KeycloakConfig struct {
		Issuer            string `env:"ISSUER"                env-required:"true"`
		JWKSURL           string `env:"JWKS_URL"              env-required:"true"`
		ClientID          string `env:"CLIENT_ID"             env-required:"true"`
		Audience          string `env:"AUDIENCE"              env-default:""`
		AllowedAlgorithms string `env:"ALLOWED_ALGORITHMS"   env-default:"RS256"`
		JWKSCacheTTL      string `env:"JWKS_CACHE_TTL"       env-default:"10m"`
		ClockSkew         string `env:"CLOCK_SKEW"           env-default:"30s"`
	}

	PGConfig struct {
		Host     string `env:"HOST"     env-default:"localhost"`
		Port     string `env:"PORT"     env-default:"5432"`
		User     string `env:"USER"     env-default:"test"`
		Password string `env:"PASSWORD" env-default:"test"`
		Database string `env:"DB"       env-default:"postgres"`
		SSLMode  string `env:"SSLMODE"  env-default:"disable"`
	}

	MongoConfig struct {
		URI        string `env:"MONGO_URI"        env-default:"mongodb://localhost:27017"`
		Database   string `env:"MONGO_DB"         env-default:"socketflow"`
		Collection string `env:"MONGO_COLLECTION" env-default:"messages"`
	}

	WebSocketConfig struct {
		ReadBufferSize    int    `env:"WS_READ_BUFFER_SIZE"   env-default:"1024"`
		WriteBufferSize   int    `env:"WS_WRITE_BUFFER_SIZE"  env-default:"1024"`
		AllowedOrigins    string `env:"WS_ALLOWED_ORIGINS"    env-default:""`
		EnableCompression bool   `env:"WS_ENABLE_COMPRESSION" env-default:"true"`
	}

	MinioConfig struct {
		Endpoint        string `env:"MINIO_ENDPOINT"   env-default:"localhost:9000"`
		AccessKeyID     string `env:"MINIO_ACCESS_KEY" env-default:"minioadmin"`
		SecretAccessKey string `env:"MINIO_SECRET_KEY" env-default:"minioadmin"`
		Bucket          string `env:"MINIO_BUCKET"     env-default:"uploads"`
		Region          string `env:"MINIO_REGION"     env-default:""`
		UseSSL          bool   `env:"MINIO_USE_SSL"    env-default:"false"`
	}

	FCMConfig struct {
		ProjectID   string `env:"FCM_PROJECT_ID"   env-default:""`
		AccessToken string `env:"FCM_ACCESS_TOKEN" env-default:""`
	}

	SchedulerConfig struct {
		CleanupCron string `env:"cleanupCron" env-default:"0 0 * * *"`
		TTL         string `env:"ttl"         env-default:"1m"`
		Timezone    string `env:"timezone"    env-default:"UTC"`
	}

	AppConfig struct {
		Postgres  PGConfig        `env-prefix:"PG_"`
		Mongo     MongoConfig     `env-prefix:"MONGO_"`
		Server    ServerConfig    `env-prefix:"SERVER_"`
		Keycloak  KeycloakConfig  `env-prefix:"KEYCLOAK_"`
		Minio     MinioConfig     `env-prefix:"MINIO_"`
		WebSocket WebSocketConfig `env-prefix:"WS_"`
		FCM       FCMConfig       `env-prefix:"FCM_"`
		Scheduler SchedulerConfig
	}
)

func (c PGConfig) DSN() string {
	hostPort := net.JoinHostPort(c.Host, c.Port)

	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s",
		url.QueryEscape(c.User),
		url.QueryEscape(c.Password),
		hostPort,
		c.Database,
		c.SSLMode,
	)
}

func (c KeycloakConfig) CacheDuration() time.Duration {
	duration, err := time.ParseDuration(c.JWKSCacheTTL)
	if err != nil {
		return 10 * time.Minute
	}

	return duration
}

func (c KeycloakConfig) ClockSkewDuration() time.Duration {
	duration, err := time.ParseDuration(c.ClockSkew)
	if err != nil {
		return 30 * time.Second
	}

	return duration
}

func (c KeycloakConfig) AllowsAlgorithm(alg string) bool {
	for allowed := range strings.SplitSeq(c.AllowedAlgorithms, ",") {
		if strings.TrimSpace(allowed) == alg {
			return true
		}
	}

	return false
}

func LoadConfig(envProfile bool) (*AppConfig, error) {
	config := new(AppConfig)
	var err error

	if envProfile {
		err = cleanenv.ReadEnv(config)
	} else {
		err = cleanenv.ReadConfig("/Users/farukh/sandbox/socket-flow/text", config)
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to read environment config")
	}

	return config, nil
}
