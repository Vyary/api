package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Vyary/api/internal/models"
)

func (s *Server) CreateStrategyHandler(w http.ResponseWriter, r *http.Request) {
	var strategy models.Strategy
	statusCode, err := DecodeJSON(r, &strategy)
	if err != nil {
		writeError(w, statusCode, err.Error())
		return
	}

	// TODO: Validate strategy

	user, err := GetUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	storedStrategy, err := s.db.StoreStrategy(*user, strategy)
	if err != nil {
		slog.Error("failed to store strategy", "error", err)
		writeError(w, http.StatusInternalServerError, "Unable to save strategy, try again later.")
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/strategies/%d", storedStrategy.ID))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(storedStrategy); err != nil {
		slog.Error("failed to encode success response", "error", err)
	}
}

func (s *Server) GetStrategyHandler(w http.ResponseWriter, r *http.Request) {
	strategyID := r.PathValue("strategy_id")

	user, err := GetUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	strategy, err := s.db.RetrieveStrategy(strategyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No strategy found with this ID")
			return
		}

		slog.Error("failed to retrieve strategy", "error", err)
		writeError(w, http.StatusInternalServerError, "Unable to retrieve strategy")
		return
	}

	if !strategy.Public && strategy.UserID != user.ID {
		writeError(w, http.StatusUnauthorized, "Strategy is private.")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(strategy); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func (s *Server) CreateStrategyTableHandler(w http.ResponseWriter, r *http.Request) {
	strategyID := r.PathValue("strategy_id")

	var table models.StrategyTable
	if err := json.NewDecoder(r.Body).Decode(&table); err != nil {
		slog.Error("failed to decode request body", "error", err)
		writeError(w, http.StatusBadRequest, "Invalid JSON format in request body")
		return
	}

	storedTable, err := s.db.StoreStrategyTable(strategyID, table)
	if err != nil {
		slog.Error("failed to store strategy table", "error", err)
		writeError(w, http.StatusInternalServerError, "Unable to save strategy table, try again later.")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(storedTable); err != nil {
		slog.Error("failed to encode success response", "error", err)
	}
}
