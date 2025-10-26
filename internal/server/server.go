// Package server provides HTTP server setup, routing, and error handling
package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Vyary/api/internal/database"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port string
	db   database.Service
}

func New(db database.Service) *http.Server {
	port := os.Getenv("PORT")
	if port == "" {
		slog.Error("PORT env variable is required")
		os.Exit(1)
	}

	srv := &Server{
		port: port,
		db:   db,
	}

	return &http.Server{
		Addr:              fmt.Sprintf(":%s", srv.port),
		Handler:           srv.RegisterRoutes(),
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}
