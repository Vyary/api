package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vyary/api/internal/server"
	"github.com/Vyary/api/pkg/telemetry"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

var (
	service = os.Getenv("SERVICE_NAME")
	logger  = otelslog.NewLogger(service)
)

func main() {
	if err := run(); err != nil {
		slog.Error("failed to start service", "error", err, "service", service)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	otelShutdown, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup Otel SDK: ", err)
	}
	defer otelShutdown(context.Background())

	// slog.SetDefault(logger)

	srv := server.New()

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	select {
	case err = <-srvErr:
		return err // TODO: more context
	case <-ctx.Done():
		stop()
	}

	ctxTO, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return srv.Shutdown(ctxTO)
}
