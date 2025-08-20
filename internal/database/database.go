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

	StoreStrategy(user models.User, strategy models.Strategy) (*models.StrategyDTO, error)
	StoreStrategyTable(strategyID string, table models.StrategyTable) (*models.StrategyTable, error)
	StoreStrategyItem(strategyID string, item models.StrategyItem) error

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

func (s *service) StoreStrategy(user models.User, strategy models.Strategy) (*models.StrategyDTO, error) {
	query := `
	INSERT INTO strategies (user_id, created_by, name, description, atlas, public) 
	VALUES (?, ?, ?, ?, ?, ?) 
	RETURNING id, created_by, name, description, atlas, public, created_at, updated_at`

	var strategyDTO models.StrategyDTO
	if err := s.db.QueryRow(query, user.ID, user.Name, strategy.Name, strategy.Description, strategy.Atlas, strategy.Public).Scan(&strategyDTO.ID, &strategyDTO.CreatedBy, &strategyDTO.Name, &strategyDTO.Description, &strategyDTO.Atlas, &strategyDTO.Public, &strategyDTO.CreatedAt, &strategyDTO.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to create strategy: %w", err)
	}

	return &strategyDTO, nil
}

func (s *service) StoreStrategyTable(strategyID string, table models.StrategyTable) (*models.StrategyTable, error) {
	query := `
	INSERT INTO strategy_tables (strategy_id, type, title)
	VALUES (?, ?, ?)
	RETURNING id, strategy_id, type, title`

	var st models.StrategyTable
	if err := s.db.QueryRow(query, strategyID, table.Type, table.Title).Scan(&st.ID, &st.StrategyID, &st.Type, &st.Title); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &st, nil
}

func (s *service) RetrieveStrategy(id string) (*models.Strategy, error) {
	strategyQuery := `
	SELECT *
	FROM strategies 
	WHERE id = ?`

	// TODO: create DTO for strategy
	var strategy models.Strategy

	err := s.db.QueryRow(strategyQuery, id).Scan(&strategy.UserID, &strategy.CreatedBy, &strategy.Name, &strategy.Description, &strategy.Atlas, &strategy.Public, &strategy.CreatedAt, &strategy.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve strategy: %w", err)
	}

	itemsQuery := `
	SELECT si.*, i.*
	FROM strategy_items si
	JOIN items i ON si.item_id = i.id
	WHERE strategy_id = ?
	`

	rows, err := s.db.Query(itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve strategy items: %w", err)
	}

	for rows.Next() {
		var item models.SItem

		if err := rows.Scan(&item.ID, &item.StrategyID, &item.Icon, &item.TableID, &item.Amount, &item.Role, &item.Pair, &item.DropChance, &item.Icon, &item.Name, &item.Base, &item.Category, &item.Value, &item.Currency, &item.Listed, &item.UserID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}


	}

	return &strategy, nil
}

func (s *service) StoreStrategyItem(strategyID string, item models.StrategyItem) error {
	return nil
}
