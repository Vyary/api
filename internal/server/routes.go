package server

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Vyary/api/internal/models"
	"github.com/andybalholm/brotli"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /categories/{categoryID}", CompressMiddleware(s.GetItemsByCategoryHandler()))
	mux.Handle("GET /subcategories/{subcategoryID}", CompressMiddleware(s.GetItemsBySubCategoryHandler()))

	mux.HandleFunc("GET /info", s.InfoHandler)

	mux.HandleFunc("POST /auth/poe/exchange", s.ExchangeHandler)
	mux.HandleFunc("POST /auth/poe/refresh", s.TokenRefreshHandler)
	mux.HandleFunc("POST /auth/poe/logout", s.LogoutHandler)
	mux.HandleFunc("POST /auth/poe/logout-all", s.LogoutAllHandler)

	mux.HandleFunc("POST /v1/strategies", s.CreateStrategyHandler)
	// mux.HandleFunc("GET /v1/strategies", s.ListPublicStrategiesHandler)
	// mux.HandleFunc("GET /v1/strategies/featured", s.ListFeaturedStrategiesHandler)

	mux.HandleFunc("GET /v1/strategies/{strategy_id}", s.GetStrategyHandler)
	// mux.HandleFunc("PUT /v1/strategies/{strategy_id}", s.UpdateStrategyHandler)
	// mux.HandleFunc("DELETE /v1/strategies/{strategy_id}", s.DeleteStrategyHandler)

	mux.HandleFunc("POST /v1/strategies/{strategy_id}/tables", s.CreateStrategyTableHandler)
	// mux.HandleFunc("GET /v1/strategies/{strategy_id}/tables", s.ListStrategyTablesHandler)
	// mux.HandleFunc("PUT /v1/strategies/{strategy_id}/tables/{table_id}", s.UpdateStrategyTableHandler)
	// mux.HandleFunc("DELETE /v1/strategies/{strategy_id}/tables/{table_id}", s.DeleteStrategyTableHandler)

	mux.HandleFunc("POST /v1/strategies/{strategy_id}/items", s.AddStrategyItemHandler)
	// mux.HandleFunc("GET /v1/strategies/{strategy_id}/tables/{table_id}/items", s.ListStrategyItemHandler)
	// mux.HandleFunc("PUT /v1/strategies/{strategy_id}/tables/{table_id}/items/{item_id}", s.UpdateStrategyItemHandler)
	// mux.HandleFunc("DELETE /v1/strategies/{strategy_id}/tables/{table_id}/items/{item_id}", s.DeleteStrategyItemHandler)

	return mux
}

func (s *Server) AddStrategyItemHandler(w http.ResponseWriter, r *http.Request) {
	strategyID := r.PathValue("strategy_id")

	// user, err := GetUser(r)
	// if err != nil {
	// 	writeError(w, http.StatusUnauthorized, err.Error())
	// 	return
	// }

	var strategyItem models.StrategyItem
	if err := json.NewDecoder(r.Body).Decode(&strategyItem); err != nil {
		slog.Error("failed to decode request body", "error", err)
		writeError(w, http.StatusBadRequest, "Invalid JSON format in request body")
		return
	}

	// TODO: validate strategy item
	// TODO: check if user can add items to strategy

	if err := s.db.StoreStrategyItem(strategyID, strategyItem); err != nil {
		slog.Error("failed to store strategy item", "error", err)
		writeError(w, http.StatusInternalServerError, "Unable to save strategy item, try again later.")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func CompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enc := strings.ToLower(r.Header.Get("Accept-Encoding"))
		var writer io.WriteCloser

		switch {
		case strings.Contains(enc, "br"):
			w.Header().Set("Content-Encoding", "br")
			writer = brotli.NewWriterLevel(w, brotli.BestSpeed)
		case strings.Contains(enc, "gzip"):
			w.Header().Set("Content-Encoding", "gzip")
			var err error
			writer, err = gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
		}

		if writer != nil {
			w.Header().Add("Vary", "Accept-Encoding")
			defer writer.Close()
			w = &compressResponseWriter{ResponseWriter: w, Writer: writer}
		}

		next.ServeHTTP(w, r)
	})
}

type compressResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *compressResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
