// Package server provides HTTP server setup, routing, and error handling
package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Vyary/api/internal/database"
	"github.com/Vyary/api/internal/models"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port string
	db   database.Service
}

func New() *http.Server {
	port := os.Getenv("PORT")
	if port == "" {
		slog.Error("PORT env variable is required")
		os.Exit(1)
	}

	srv := &Server{
		port: port,
		db:   database.New(),
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

func writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := models.ErrorResponse{
		Error: message,
		Code:  code,
	}

	json.NewEncoder(w).Encode(response)
}
