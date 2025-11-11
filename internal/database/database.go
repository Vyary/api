// Package database provides the data access layer,
// wrapping persistence logic around libSQL using a Service interface.
package database

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/Vyary/api/internal/models"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Service interface {
	GetItemsByCategory(category string) ([]models.Item, error)
	GetItemsBySubCategory(category string) ([]models.Item, error)
	GetPrices(period int64) ([]models.Price, error)

	StoreOAuthToken(id string, token models.OAuthToken) error
	RemoveOAuthToken(id string) error

	StoreRefreshToken(userID string, tokenID string, expiration time.Duration) error
	IsRefreshTokenValid(tokenID string) bool
	RevokeRefreshToken(userID string, tokenID string) error
	RevokeAllRefreshTokens(userID string) error

	StoreStrategy(user models.UserProfile, strategy models.Strategy) (*models.StrategyDTO, error)
	StoreStrategyTable(strategyID string, table models.StrategyTable) (*models.StrategyTable, error)
	StoreStrategyItem(strategyID string, item models.StrategyItem) error

	RetrieveStrategy(id string) (*models.Strategy, error)

	Close() error
}

type tursoDB struct {
	db *sql.DB
}

var (
	dbURL    = os.Getenv("DB_URL")
	instance *tursoDB
	once     sync.Once
)

func Get() Service {
	once.Do(func() {
		db := tursoDB{}

		if err := db.connect(); err != nil {
			log.Fatal("connecting to db: %w", err)
		}

		instance = &db
	})

	return instance
}

func (s *tursoDB) connect() error {
	db, err := sql.Open("libsql", dbURL)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(30 * time.Minute)

	s.db = db

	return nil
}

func (s *tursoDB) Close() error {
	var errs []error

	// if s.connector != nil {
	// 	if err := s.connector.Close(); err != nil {
	// 		errs = append(errs, err)
	// 	}
	// }

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

func (s *tursoDB) GetItemsByCategory(category string) ([]models.Item, error) {
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
		socketed_items,
		properties,
		requirements,
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
		desecrated
	FROM items
	WHERE category = ?`

	slog.Info("getting sub items")

	rows, err := s.db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items = []models.Item{}

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
			&i.SocketedItems,
			&i.Properties,
			&i.Requirements,
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
		)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	return items, rows.Err()
}

func (s *tursoDB) GetItemsBySubCategory(subCategory string) ([]models.Item, error) {
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
		socketed_items,
		properties,
		requirements,
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
		desecrated
	FROM items
	WHERE sub_category = ?`

	slog.Info("getting sub items")

	rows, err := s.db.Query(query, subCategory)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items = []models.Item{}

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
			&i.SocketedItems,
			&i.Properties,
			&i.Requirements,
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
		)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	return items, rows.Err()
}

func (s *tursoDB) GetPrices(period int64) ([]models.Price, error) {
	query := `
	SELECT item_id, price, currency_id, volume, stock, league, timestamp
	FROM prices
	WHERE timestamp > ?`

	rows, err := s.db.Query(query, period)
	if err != nil {
		return nil, err
	}

	var prices = []models.Price{}

	for rows.Next() {
		p := models.Price{}

		err := rows.Scan(&p.ItemID, &p.Price, &p.CurrencyID, &p.Volume, &p.Stock, &p.League, &p.Timestamp)
		if err != nil {
			return nil, err
		}

		prices = append(prices, p)
	}

	return prices, rows.Err()
}
