package server

import (
	"errors"
	"fmt"
	"socket-flow/internal/config"

	"github.com/golang-migrate/migrate/v4"
)

const mongoMigrationsPath = "file://./migrations/mongo"

func runMongoMigration(cfg config.MongoConfig) error {

	databaseURL := fmt.Sprintf("%s/%s", cfg.URI, cfg.Database)

	m, err := migrate.New(mongoMigrationsPath, databaseURL)

	if err != nil {
		return fmt.Errorf("migrate.New failed: %w", err)
	}

	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate Up failed: %w", err)
	}

	return nil
}
