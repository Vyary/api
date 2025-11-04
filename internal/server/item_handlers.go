package server

import (
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/Vyary/api/internal/models"
)

type CacheValue struct {
	Result    any
	Timestamp time.Time
}

type PriceMap struct {
	Map       map[models.ItemID]map[models.League]models.Prices
	Timestamp time.Time
}

var pricesMap = PriceMap{Map: map[models.ItemID]map[models.League]models.Prices{}}
var cache = make(map[string]CacheValue)

func (s *Server) GetItemsByCategoryHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("categoryID")
		start := time.Now()

		if cache, ok := cache[category]; ok && time.Since(cache.Timestamp) < time.Hour {
			slog.Info("cache hit")

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(cache.Result); err != nil {
				slog.Error("encoding items response", "error", err)
			}

			slog.Info("took", "time", time.Since(start))
			return
		}

		items, err := s.db.GetItemsByCategory(category)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := s.calculatePrices(); err != nil {
			slog.Error(err.Error())
		}

		for i := range items {
			if p, ok := pricesMap.Map[items[i].ID]; ok {
				items[i].Prices = p
			}
		}

		cache[category] = CacheValue{Result: items, Timestamp: time.Now()}
		slog.Info("cache miss")
		slog.Info("timed", "took", time.Since(start).String(), "path", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(items); err != nil {
			slog.Error("encoding items response", "error", err)
		}
	})
}

func (s *Server) GetItemsBySubCategoryHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		subCategory := r.PathValue("subcategoryID")
		start := time.Now()

		if result, ok := cache[subCategory]; ok && time.Since(result.Timestamp) < time.Hour {
			slog.Info("cache hit")

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(result); err != nil {
				slog.Error("encoding items response", "error", err)
			}

			slog.Info("took", "time", time.Since(start))
			return
		}

		items, err := s.db.GetItemsBySubCategory(subCategory)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := s.calculatePrices(); err != nil {
			slog.Error(err.Error())
		}

		for i := range items {
			if p, ok := pricesMap.Map[items[i].ID]; ok {
				items[i].Prices = p
			}
		}

		cache[subCategory] = CacheValue{Result: items, Timestamp: time.Now()}
		slog.Info("cache miss")
		slog.Info("timed", "since", time.Since(start), "path", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(items); err != nil {
			slog.Error("encoding items response", "error", err)
		}
	})
}

type agg struct {
	weightedSum   float64
	weightedTotal float64
}
type currency map[string]*agg

func (s *Server) calculatePrices() error {
	if time.Since(pricesMap.Timestamp) < time.Hour {
		return nil
	}

	start := time.Now()

	prices, err := s.db.GetPrices(time.Now().Add(-24 * time.Hour).UTC().Unix())
	if err != nil {
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
		agg.weightedSum += p.Price * weight
		agg.weightedTotal += weight
	}

	for itemID, weights := range priceWeights {
		for league, wv := range weights {
			if pricesMap.Map[itemID] == nil {
				pricesMap.Map[itemID] = map[models.League]models.Prices{}
			}

			if pricesMap.Map[itemID][league] == nil {
				pricesMap.Map[itemID][league] = models.Prices{}
			}

			for currency, v := range wv {
				pricesMap.Map[itemID][league][models.Currency(currency)] = v.weightedSum / v.weightedTotal
			}
		}
	}

	pricesMap.Timestamp = time.Now()

	slog.Info("calculatePrices", "took", time.Since(start))

	return nil
}
