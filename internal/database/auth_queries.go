package database

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Vyary/api/internal/models"
)

func (s *service) StoreOAuthToken(id string, token models.OAuthToken) error {
	query := `
	INSERT INTO users (id, username, access_token, expires_in, token_type, scope, sub)
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
	INSERT INTO refresh_tokens (user_id, token_id, expires_at) 
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
