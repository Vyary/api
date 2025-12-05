// Package database provides the data access layer,
// wrapping persistence logic around libSQL using a Service interface.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Vyary/api/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/tursodatabase/go-libsql"
)

type Service interface {
	GetItemsByCategory(ctx context.Context, category string) ([]models.Item, error)
	GetItemsBySubCategory(ctx context.Context, subCategory string) ([]models.Item, error)
	GetPrices(ctx context.Context, period int64) ([]models.Price, error)

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
	db        *sql.DB
	connector *libsql.Connector
}

var (
	instance *tursoDB
	once     sync.Once
	service  = os.Getenv("SERVICE_NAME")
	tracer   = otel.Tracer(service)
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
	primaryURL := os.Getenv("TURSO_URL")
	authToken := os.Getenv("DB_AUTH_TOKEN")

	dir := "./internal/database/local/"

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating directory '%s': %w", dir, err)
	}

	dbPath := filepath.Join(dir, "local.db")

	interval := 60 * time.Minute

	connector, err := libsql.NewEmbeddedReplicaConnector(dbPath, primaryURL, libsql.WithAuthToken(authToken), libsql.WithSyncInterval(interval))
	if err != nil {
		return err
	}

	s.connector = connector
	s.db = sql.OpenDB(connector)

	return nil
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

func (s *tursoDB) GetItemsByCategory(ctx context.Context, category string) ([]models.Item, error) {
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

	_, span := tracer.Start(ctx, "DB.GetItemsByCategory",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "sqlite"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "items"),
		attribute.String("category", category),
	)

	rows, err := s.db.Query(query, category)
	if err != nil {
		span.SetStatus(codes.Error, "executing query")
		span.RecordError(err)
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
			span.SetStatus(codes.Error, "scanning row")
			span.RecordError(err)
			return nil, err
		}
		items = append(items, i)
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("successfully retrieved %d items", len(items)))
	return items, rows.Err()
}

func (s *tursoDB) GetItemsBySubCategory(ctx context.Context, subCategory string) ([]models.Item, error) {
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

	_, span := tracer.Start(ctx, "DB.GetItemsBySubCategory",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "sqlite"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "items"),
		attribute.String("subCategory", subCategory),
	)

	rows, err := s.db.Query(query, subCategory)
	if err != nil {
		span.SetStatus(codes.Error, "executing query")
		span.RecordError(err)
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
			span.SetStatus(codes.Error, "scanning row")
			span.RecordError(err)
			return nil, err
		}
		items = append(items, i)
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("successfully retrieved %d items", len(items)))
	return items, rows.Err()
}

func (s *tursoDB) GetPrices(ctx context.Context, period int64) ([]models.Price, error) {
	query := `
	SELECT item_id, price, currency_id, volume, stock, league, timestamp
	FROM prices
	WHERE timestamp > ?`

	_, span := tracer.Start(ctx, "DB.GetPrices",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "sqlite"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "prices"),
	)

	rows, err := s.db.Query(query, period)
	if err != nil {
		span.SetStatus(codes.Error, "executing query")
		span.RecordError(err)
		return nil, err
	}

	var prices = []models.Price{}

	for rows.Next() {
		p := models.Price{}

		err := rows.Scan(&p.ItemID, &p.Price, &p.CurrencyID, &p.Volume, &p.Stock, &p.League, &p.Timestamp)
		if err != nil {
			span.SetStatus(codes.Error, "scanning row")
			span.RecordError(err)
			return nil, err
		}

		prices = append(prices, p)
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("successfully retrieved %d prices", len(prices)))
	return prices, rows.Err()
}
