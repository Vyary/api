package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Vyary/api/internal/models"
	_ "github.com/joho/godotenv/autoload"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /info", s.InfoHandler())
	mux.Handle("POST /auth/poe/exchange", s.ExchangeHandler())
	mux.Handle("POST /auth/poe/refresh", s.TokenRefreshHandler())
	mux.Handle("POST /auth/poe/logout", s.LogoutHandler())
	mux.Handle("POST /auth/poe/logout-all", s.LogoutAllHandler())

	mux.Handle("POST /strategies", s.CreateStrategyHandler())
	mux.Handle("GET /strategies/{id}", s.GetStrategy())

	return mux
}

func (s *Server) CreateStrategyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			slog.Warn("invalid content type", "content_type", r.Header.Get("Content-Type"), "expected", "application/json")
			writeError(w, "Content-Type must be application/json", http.StatusBadRequest)
			return
		}

		var strategy models.Strategy
		if err := json.NewDecoder(r.Body).Decode(&strategy); err != nil {
			slog.Error("failed to decode request body", "error", err)
			writeError(w, "Invalid JSON format in request body", http.StatusBadRequest)
			return
		}

		// TODO: get user from jwt
		// TODO: validate strategy
		u := models.User{
			ID:   "123",
			Name: "Vyary",
		}

		id, err := s.db.StoreStrategy(u, strategy)
		if err != nil {
			slog.Error("failed to store strategy", "error", err)
			writeError(w, "Unable to save strategy, try again later.", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"strategyID": id,
		}

		if err = json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("failed to encode success response", "error", err)
		}
	})
}

func (s *Server) GetStrategy() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		strategyID := r.PathValue("id")

		strategy, err := s.db.RetrieveStrategy(strategyID)
		if err != nil {
			slog.Error("failed to retrieve strategy", "error", err)
			writeError(w, "Unable to retrieve strategy", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(strategy); err != nil {
			slog.Error("failed to encode response", "error", err)
		}
	})
}
