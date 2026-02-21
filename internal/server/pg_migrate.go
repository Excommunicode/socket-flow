package server

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"socket-flow/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

const (
	migrationsPath = "file://./migrations/postgres"
	driverName     = "postgres"
)

func runPgMigrations(cfg config.PGConfig) error {
	migrateDB, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return fmt.Errorf("open migrate db: %w", err)
	}
	defer func() {
		if err := migrateDB.Close(); err != nil {
			log.Printf("close migrate db: %v", err)
		}
	}()

	driver, err := postgres.WithInstance(migrateDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, driverName, driver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}

	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
