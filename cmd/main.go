package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Vyary/api/internal/database"
	"github.com/Vyary/api/internal/server"
	"github.com/Vyary/api/pkg/telemetry"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func main() {
	if err := run(); err != nil {
		slog.Error("failed to start service", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	service := os.Getenv("SERVICE_NAME")
	logger := otelslog.NewLogger(service)

	otelShutdown, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		return fmt.Errorf("setting Otel SDK: %w", err)
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	slog.SetDefault(logger)

	db := database.Get()
	defer db.Close()

	srv := server.New(db)

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	select {
	case err = <-srvErr:
		return err
	case <-ctx.Done():
		stop()
	}

	err = srv.Shutdown(context.Background())
	return err
}
