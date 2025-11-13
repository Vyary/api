package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/Vyary/api/internal/models"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type CacheValue struct {
	Result    any
	Timestamp time.Time
}

type PriceMap struct {
	Map       map[models.ItemID]map[models.League]models.Prices
	Timestamp time.Time
}

type agg struct {
	pricePoints   float64
	weightedSum   float64
	weightedTotal float64
	volumeTotal   float64
	stockTotal    float64
}

type currency map[string]*agg

var (
	pricesMap = PriceMap{Map: map[models.ItemID]map[models.League]models.Prices{}}
	cache     = make(map[string]CacheValue)
	service   = os.Getenv("SERVICE_NAME")
	tracer    = otel.Tracer(service)
)

func (s *Server) GetItemsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("categoryID")
		subCategory := r.PathValue("subcategoryID")
		if subCategory != "" {
			category = subCategory
		}

		start := time.Now()

		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, "GetItems", trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.path", r.URL.Path),
			attribute.String("category", category),
			attribute.String("subCategory", subCategory),
		)

		if cache, ok := cache[category]; ok && time.Since(cache.Timestamp) < time.Hour {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(cache.Result); err != nil {
				slog.Error("encoding items response", "error", err)

				span.SetStatus(codes.Error, "encoding cache response")
				span.RecordError(err)
			}

			dur := time.Since(start).String()
			slog.Info(fmt.Sprintf("Cache hit: %s - %s", category, dur), "function", "GetItemsHandler", "duration", dur)

			span.SetStatus(codes.Ok, "")
			return
		}

		var items []models.Item
		var err error

		if subCategory != "" {
			items, err = s.db.GetItemsBySubCategory(ctx, category)
		} else {
			items, err = s.db.GetItemsByCategory(ctx, category)
		}
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			span.SetStatus(codes.Error, "DB query")
			span.RecordError(err)
			return
		}

		if err := s.calculatePrices(ctx); err != nil {
			slog.Error(err.Error())

			span.RecordError(err)
		}

		for i := range items {
			if p, ok := pricesMap.Map[items[i].ID]; ok {
				items[i].Prices = p
			}
		}

		if len(items) > 0 {
			cache[category] = CacheValue{Result: items, Timestamp: time.Now()}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(items); err != nil {
			slog.Error("encoding items response", "error", err)

			span.SetStatus(codes.Error, "encoding response failed")
			span.RecordError(err)
		}

		dur := time.Since(start).String()
		slog.Info(fmt.Sprintf("Cache miss: %s - %s", category, dur), "function", "GetItemsHandler", "duration", dur)

		span.SetStatus(codes.Ok, "")
	})
}

func (s *Server) calculatePrices(ctx context.Context) error {
	if time.Since(pricesMap.Timestamp) < time.Hour {
		return nil
	}

	start := time.Now()

	_, span := tracer.Start(ctx, "server.calculatePrices",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	prices, err := s.db.GetPrices(time.Now().Add(-24 * time.Hour).UTC().Unix())
	if err != nil {
		span.SetStatus(codes.Error, "retrieving prices")
		span.RecordError(err)
		return err
	}

	var priceWeights = map[models.ItemID]map[models.League]currency{}
	now := time.Now().UTC().Unix()

	for _, p := range prices {
		hoursAgo := float64(now-p.Timestamp) / 3600.0
		weight := float64(p.Volume) * math.Exp(-hoursAgo/1.0)

		if priceWeights[p.ItemID] == nil {
			priceWeights[p.ItemID] = map[models.League]currency{}
		}

		if priceWeights[p.ItemID][p.League] == nil {
			priceWeights[p.ItemID][p.League] = currency{}
		}

		if priceWeights[p.ItemID][p.League][p.CurrencyID] == nil {
			priceWeights[p.ItemID][p.League][p.CurrencyID] = &agg{}
		}

		agg := priceWeights[p.ItemID][p.League][p.CurrencyID]
		agg.pricePoints += 1
		agg.weightedSum += p.Price * weight
		agg.weightedTotal += weight
		agg.volumeTotal += p.Volume
		agg.stockTotal += p.Stock
	}

	for itemID, weights := range priceWeights {
		for league, wv := range weights {
			for currency, v := range wv {
				if pricesMap.Map[itemID] == nil {
					pricesMap.Map[itemID] = map[models.League]models.Prices{}
				}

				if pricesMap.Map[itemID][league] == nil {
					pricesMap.Map[itemID][league] = models.Prices{}
				}

				pDTO := models.PriceDTO{
					Price:  v.weightedSum / v.weightedTotal,
					Volume: v.volumeTotal / v.pricePoints,
					Stock:  v.stockTotal / v.pricePoints,
				}

				pricesMap.Map[itemID][league][models.Currency(currency)] = pDTO
			}
		}
	}

	pricesMap.Timestamp = time.Now()

	dur := time.Since(start).String()
	slog.Info(fmt.Sprintf("RUN calculatePrices - %s", dur), "duration", dur)

	span.SetStatus(codes.Ok, "")
	return nil
}
