package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/Vyary/api/internal/models"
	"go.opentelemetry.io/otel"
)

var (
	service = os.Getenv("SERVICE_NAME")
	tracer  = otel.Tracer(service)
)

type ItemsDTO struct {
	Items  []models.Item `json:"items"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
	Total  int           `json:"total"`
}

func (s *Server) GetItemsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")
		search := r.URL.Query().Get("search")
		order := r.URL.Query().Get("order")
		league := r.URL.Query().Get("league")

		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			limit = 10
		}

		offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil {
			offset = 0
		}

		if league != "chc" {
			league = "csc"
		}

		if order != "asc" {
			order = "desc"
		}

		searchPattern := "%" + search + "%"
		orderBy := fmt.Sprintf("%s_value %s", league, order)

		items, total, err := s.db.GetItems(r.Context(), category, searchPattern, orderBy, limit, offset, league)
		if err != nil {
			NewInternalError(r.Context(), w, "quering db", err, r.URL.Path)
			return
		}

		if len(items) == 0 {
			http.Error(w, "no items found", http.StatusNotFound)
			return
		}

		result := ItemsDTO{Items: items, Limit: limit, Offset: offset, Total: total}

		WriteJSON(r.Context(), w, http.StatusOK, result)
	})
}
