// Package database provides the data access layer,
// wrapping persistence logic around libSQL using a Service interface.
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Vyary/api/internal/models"
	"github.com/tursodatabase/go-libsql"
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

	Close() error
}

type tursoDB struct {
	db        *sql.DB
	connector *libsql.Connector
}

var (
	instance *tursoDB
	once     sync.Once
)

func Get() *tursoDB {
	once.Do(func() {
		instance = connect()
	})

	return instance
}

func connect() *tursoDB {
	primaryURL := os.Getenv("TURSO_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")

	dbName := "local.db"

	dir, err := os.MkdirTemp("", "libsql-*")
	if err != nil {
		fmt.Println("Error creating temporary directory:", err)
		os.Exit(1)
	}

	dbPath := filepath.Join(dir, dbName)

	syncInterval := 5 * time.Minute

	connector, err := libsql.NewEmbeddedReplicaConnector(dbPath, primaryURL,
		libsql.WithAuthToken(authToken),
		libsql.WithSyncInterval(syncInterval),
	)
	if err != nil {
		fmt.Println("Error creating connector:", err)
		os.Exit(1)
	}

	db := sql.OpenDB(connector)

	return &tursoDB{db: db, connector: connector}
}

func (s *tursoDB) Close() error {
	var errs []error

	if s.connector != nil {
		if err := s.connector.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("service close errors: %v", errs)
	}

	return nil
}

func (s *tursoDB) GetItemsByCategory(category string) (*[]models.Item, error) {
	query := `
	SELECT
		id,
		realm,
		category,
		sub_category,
		icon,
		icon_tier_text,
		name,
		base_type,
		rarity,
		w,
		h,
		ilvl,
		sockets_count,
		properties,
		requirements,
		enchant_mods,
		rune_mods,
		implicit_mods,
		explicit_mods,
		fractured_mods,
		desecrated_mods,
		flavour_text,
		descr_text,
		sec_descr_text,
		support,
		duplicated,
		corrupted,
		sanctified,
		desecrated,
		buy,
		sell
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
			&i.ID,
			&i.Realm,
			&i.Category,
			&i.SubCategory,
			&i.Icon,
			&i.IconTierText,
			&i.Name,
			&i.BaseType,
			&i.Rarity,
			&i.W,
			&i.H,
			&i.Ilvl,
			&i.SocketsCount,
			&i.Properties,
			&i.Requirements,
			&i.EnchantMods,
			&i.RuneMods,
			&i.ImplicitMods,
			&i.ExplicitMods,
			&i.FracturedMods,
			&i.DesecratedMods,
			&i.FlavourText,
			&i.DescrText,
			&i.SecDescrText,
			&i.Support,
			&i.Duplicated,
			&i.Corrupted,
			&i.Sanctified,
			&i.Desecrated,
			&i.Buy,
			&i.Sell,
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
