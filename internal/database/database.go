// Package database provides the data access layer,
// wrapping persistence logic around libSQL using a Service interface.
package database

import (
	"database/sql"
	"log/slog"
	"os"
	"time"

	"github.com/Vyary/api/internal/models"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Service interface {
	GetItemsByCategory(category string) (*[]models.Item, error)

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
	dbURL := os.Getenv("DB_URL")

	if dbURL == "" {
		slog.Error("DB_URL env variable is required")
		os.Exit(1)
	}

	db, err := sql.Open("libsql", dbURL)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	return &service{db: db}
}

func (s *service) GetItemsByCategory(category string) (*[]models.Item, error) {
	query := `
	SELECT
		id,
		COALESCE(realm, ''),
		COALESCE(w, 0),
		COALESCE(h, 0),
		COALESCE(icon, ''),
		COALESCE(name, ''),
		COALESCE(base_type, ''),
		COALESCE(category, ''),
		COALESCE(sub_category, ''),
		COALESCE(rarity, ''),
		COALESCE(support, 0),
		COALESCE(desecrated, 0),
		COALESCE(properties, ''),
		COALESCE(requirements, ''),
		COALESCE(enchant_mods, ''),
		COALESCE(rune_mods, ''),
		COALESCE(implicit_mods, ''),
		COALESCE(explicit_mods, ''),
		COALESCE(fractured_mods, ''),
		COALESCE(desecrated_mods, ''),
		COALESCE(flavour_text, ''),
		COALESCE(descr_text, ''),
		COALESCE(sec_descr_text, ''),
		COALESCE(icon_tier_text, ''),
		COALESCE(gem_sockets, 0)
	FROM items
	WHERE category = ?`

	rows, err := s.db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item

	for rows.Next() {
		var i models.Item
		err := rows.Scan(
			&i.ID, &i.Realm, &i.W, &i.H, &i.Icon, &i.Name,
			&i.BaseType, &i.Category, &i.SubCategory, &i.Rarity, &i.Support, &i.Desecrated,
			&i.Properties, &i.Requirements, &i.EnchantMods, &i.RuneMods,
			&i.ImplicitMods, &i.ExplicitMods, &i.FracturedMods, &i.DesecratedMods,
			&i.FlavourText, &i.DescrText, &i.SecDescrText, &i.IconTierText,
			&i.GemSocketsCount,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &items, nil
}
