package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

func (s *Server) GetItemsByCategoryHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("categoryID")

		start := time.Now()

		items, err := s.db.GetItemsByCategory(category)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		slog.Info("timed", "since", time.Since(start), "path", r.URL.Path)

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

		items, err := s.db.GetItemsBySubCategory(subCategory)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		slog.Info("timed", "since", time.Since(start), "path", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(items); err != nil {
			slog.Error("encoding items response", "error", err)
		}
	})
}
