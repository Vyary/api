package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Vyary/api/internal/models"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Service interface {
	StoreOAuthToken(id string, token models.OAuthToken) error
	RemoveOAuthToken(id string) error

	StoreRefreshToken(userID string, tokenID string, expiration time.Duration) error
	IsRefreshTokenValid(tokenID string) bool
	RevokeRefreshToken(userID string, tokenID string) error
	RevokeAllRefreshTokens(userID string) error
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

func (s *service) StoreOAuthToken(id string, token models.OAuthToken) error {
	query := `INSERT OR REPLACE INTO tokens (user_id, username, access_token, expires_in, token_type, scope, sub) VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, id, token.Username, token.AccessToken, token.ExpiresIn, token.TokenType, token.Scope, token.Sub)
	if err != nil {
		return fmt.Errorf("failed to store OAuth token to db: %w", err)
	}

	return nil
}

func (s *service) RemoveOAuthToken(id string) error {
	query := `DELETE FROM tokens WHERE user_id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to remove OAuth token: %w", err)
	}

	return nil
}

func (s *service) StoreRefreshToken(userID string, tokenID string, expiration time.Duration) error {
	query := `INSERT INTO refresh_tokens (user_id, token_id, expires_at) VALUES (?, ?, ?)`

	expiresAt := time.Now().Add(expiration)

	_, err := s.db.Exec(query, userID, tokenID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

func (s *service) IsRefreshTokenValid(tokenID string) bool {
	query := `SELECT expires_at FROM refresh_tokens WHERE token_id = ?`

	var expiresAt time.Time

	err := s.db.QueryRow(query, tokenID).Scan(&expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("refresh token not found", "token_id", tokenID)
		} else {
			slog.Error("failed to find refresh token", "error", err, "token_id", tokenID)
		}
		return false
	}

	if time.Now().After(expiresAt) {
		return false
	}

	return true
}

func (s *service) RevokeRefreshToken(userID string, tokenID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = ? AND token_id = ?`

	_, err := s.db.Exec(query, userID, tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (s *service) RevokeAllRefreshTokens(userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = ?`

	_, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (s *service) CleanupRefreshTokens() error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < ?`

	_, err := s.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup refresh tokens: %w", err)
	}

	return nil
}
