package server

import (
	"context"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Vyary/api/internal/models"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type CacheValue struct {
	Result    any
	Timestamp time.Time
}

type PriceMap struct {
	Map       map[models.ItemID]map[models.League]models.Prices
	Timestamp time.Time
	mu        sync.Mutex
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
	pricesMap = PriceMap{}
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

		if cache, ok := cache[category]; ok && time.Since(cache.Timestamp) < time.Hour {
			WriteJSON(r.Context(), w, http.StatusOK, cache.Result)
			return
		}

		var items []models.Item
		var err error

		if subCategory != "" {
			items, err = s.db.GetItemsBySubCategory(r.Context(), category)
		} else {
			items, err = s.db.GetItemsByCategory(r.Context(), category)
		}
		if err != nil {
			NewInternalError(r.Context(), w, "quering db", err, r.URL.Path)
			return
		}

		if err := s.calculatePrices(r.Context()); err != nil {
			CaptureError(r.Context(), "calculating prices", err)
		}

		for i := range items {
			if p, ok := pricesMap.Map[items[i].ID]; ok {
				items[i].Prices = p
			}
		}

		if len(items) > 0 {
			cache[category] = CacheValue{Result: items, Timestamp: time.Now()}
		}

		WriteJSON(r.Context(), w, http.StatusOK, items)
	})
}

func (s *Server) calculatePrices(ctx context.Context) error {
	pricesMap.mu.Lock()
	defer pricesMap.mu.Unlock()

	if time.Since(pricesMap.Timestamp) < time.Hour {
		return nil
	}

	ctx, span := tracer.Start(ctx, "S.calculatePrices",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	prices, err := s.db.GetPrices(ctx, time.Now().Add(-24*time.Hour).UTC().Unix())
	if err != nil {
		CaptureError(ctx, "retrieving prices", err)
		return err
	}

	newPricesMap := make(map[models.ItemID]map[models.League]models.Prices)
	var priceWeights = make(map[models.ItemID]map[models.League]currency)
	now := time.Now().UTC().Unix()

	for _, p := range prices {
		hoursAgo := float64(now-p.Timestamp) / 3600.0
		weight := float64(p.Volume) * math.Exp(-hoursAgo/1.0)

		if priceWeights[p.ItemID] == nil {
			priceWeights[p.ItemID] = make(map[models.League]currency)
		}
		if priceWeights[p.ItemID][p.League] == nil {
			priceWeights[p.ItemID][p.League] = make(currency)
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
			if newPricesMap[itemID] == nil {
				newPricesMap[itemID] = make(map[models.League]models.Prices)
			}

			if newPricesMap[itemID][league] == nil {
				newPricesMap[itemID][league] = make(models.Prices)
			}

			for currency, v := range wv {
				pDTO := models.PriceDTO{
					Price:  v.weightedSum / v.weightedTotal,
					Volume: v.volumeTotal / v.pricePoints,
					Stock:  v.stockTotal / v.pricePoints,
				}

				newPricesMap[itemID][league][models.Currency(currency)] = pDTO
			}
		}
	}

	pricesMap.Map = newPricesMap
	pricesMap.Timestamp = time.Now()

	span.SetStatus(codes.Ok, "")
	return nil
}
