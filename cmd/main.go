package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"socket-flow/internal/server"
	"socket-flow/pkg/logger"

	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	newLogger := logger.New()
	slog.SetDefault(newLogger)

	ctx := context.Background()
	srv, err := server.NewServer(ctx)
	if err != nil {
		slog.Error("failed to initialize server", "err", err)
		os.Exit(1)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed to listen", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		slog.Info("signal received, shutting down...")
	case <-done:
		slog.Info("server stopped unexpectedly")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown failed", "err", err)
	}
}
