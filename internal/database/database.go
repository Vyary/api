// Package database provides the data access layer,
// wrapping persistence logic around libSQL using a Service interface.
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

	StoreStrategy(models.User, models.Strategy) (int, error)
	RetrieveStrategy(id string) (*models.Strategy, error)
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
	query := `
	INSERT INTO users
		(id, username, access_token, expires_in, token_type, scope, sub)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		access_token = excluded.access_token,
		expires_in = excluded.expires_in,
		token_type = excluded.token_type,
		scope = excluded.scope,
		sub = excluded.sub`

	_, err := s.db.Exec(query, token.AccessToken, token.ExpiresIn, token.TokenType, token.Scope, token.Sub, id)
	if err != nil {
		return fmt.Errorf("failed to store OAuth token to db: %w", err)
	}

	return nil
}

func (s *service) RemoveOAuthToken(id string) error {
	query := `
	UPDATE users 
	SET 
		access_token = '', 
		expires_in = 0, 
		token_type = '', 
		scope = '', 
		sub = '' 
	WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to remove OAuth token: %w", err)
	}

	return nil
}

func (s *service) StoreRefreshToken(userID string, tokenID string, expiration time.Duration) error {
	query := `
	INSERT INTO refresh_tokens 
	(user_id, token_id, expires_at) 
	VALUES (?, ?, ?)`

	expiresAt := time.Now().Add(expiration)

	_, err := s.db.Exec(query, userID, tokenID, expiresAt.Unix())
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

func (s *service) IsRefreshTokenValid(tokenID string) bool {
	query := `
	SELECT COUNT(*) 
	FROM refresh_tokens 
	WHERE token_id = ? AND expires_at > ?`

	var count int

	if err := s.db.QueryRow(query, tokenID, time.Now().Unix()).Scan(&count); err != nil {
		slog.Error("failed to validate refresh token", "error", err, "token_id", tokenID)
		return false
	}

	return count > 0
}

func (s *service) RevokeRefreshToken(userID string, tokenID string) error {
	query := `
	DELETE FROM refresh_tokens 
	WHERE user_id = ? AND token_id = ?`

	_, err := s.db.Exec(query, userID, tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (s *service) RevokeAllRefreshTokens(userID string) error {
	query := `
	DELETE FROM refresh_tokens 
	WHERE user_id = ?`

	_, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (s *service) CleanupRefreshTokens() error {
	query := `
	DELETE FROM refresh_tokens 
	WHERE expires_at < ?`

	_, err := s.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup refresh tokens: %w", err)
	}

	return nil
}

func (s *service) StoreStrategy(user models.User, strategy models.Strategy) (int, error) {
	query := `
	INSERT INTO strategies (user_id, created_by, name, description) 
	VALUES (?, ?, ?, ?) 
	RETURNING id`

	var id int
	err := s.db.QueryRow(query, user.ID, user.Name, strategy.Name, strategy.Description).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create strategy: %w", err)
	}

	return id, nil
}

func (s *service) RetrieveStrategy(id string) (*models.Strategy, error) {
	query := `
	SELECT user_id, created_by, name, description, atlas, public, created_at, updated_at 
	FROM strategies 
	WHERE id = ?`

	// TODO: create DTO for strategy
	var strategy models.Strategy

	err := s.db.QueryRow(query, id).Scan(&strategy.UserID, &strategy.CreatedBy, &strategy.Name, &strategy.Description, &strategy.Atlas, &strategy.Public, &strategy.CreatedAt, &strategy.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve strategy: %w", err)
	}

	return &strategy, nil
}
