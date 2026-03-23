package server

import (
	"database/sql"
	stdErrors "errors"
	"log"

	"socket-flow/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/pkg/errors"
)

func runPgMigrations(cfg config.PGConfig) error {
	migrateDB, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return errors.Wrap(err, "open migrate db")
	}
	defer func() {
		if err := migrateDB.Close(); err != nil {
			log.Printf("close migrate db: %v", err)
		}
	}()

	driver, err := postgres.WithInstance(migrateDB, &postgres.Config{})
	if err != nil {
		return errors.Wrap(err, "postgres driver")
	}

	const (
		migrationsPath = "file://./migrations/postgres"
		driverName     = "postgres"
	)

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, driverName, driver)
	if err != nil {
		return errors.Wrap(err, "migrate instance")
	}

	defer m.Close()

	if err := m.Up(); err != nil && !stdErrors.Is(err, migrate.ErrNoChange) {
		return errors.Wrap(err, "migrate up")
	}
	return nil
}
