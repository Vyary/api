package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/Vyary/api/internal/models"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Service interface{
	SaveToken(id string, token models.OAuthToken) error
}

type service struct {
	db *sql.DB
}

func New() Service {
	dbName := os.Getenv("DB_NAME")
	token := os.Getenv("DB_TOKEN")

	if dbName == "" {
		slog.Error("DB_NAME env variable is required")
		os.Exit(1)
	}

	if token == "" {
		slog.Error("TOKEN env variable is required")
		os.Exit(1)
	}

	url := fmt.Sprintf("libsql://%s.turso.io?authToken=%s", dbName, token)

	db, err := sql.Open("libsql", url)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	return &service{db: db}
}

func (s *service) SaveToken(id string, token models.OAuthToken) error {
	query := `INSERT INTO tokens (id, access_token, expires_in, token_type, scope, sub, username) VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, id, token.AccessToken, token.ExpiresIn, token.TokenType, token.Scope, token.Sub, token.Username)
	if err != nil {
		return fmt.Errorf("Failed to save token: "+err.Error())
	}

	return nil
}
