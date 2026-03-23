package server

import (
	stdErrors "errors"
	"fmt"
	"socket-flow/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/pkg/errors"
)

func runMongoMigration(cfg config.MongoConfig) error {
	databaseURL := fmt.Sprintf("%s/%s", cfg.URI, cfg.Database)

	const mongoMigrationsPath = "file://./migrations/mongo"
	m, err := migrate.New(mongoMigrationsPath, databaseURL)

	if err != nil {
		return errors.Wrap(err, "migrate.New failed")
	}

	defer m.Close()

	if err := m.Up(); err != nil && !stdErrors.Is(err, migrate.ErrNoChange) {
		return errors.Wrap(err, "migrate Up failed")
	}

	return nil
}
