package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func (s *Server) GetItemsByCategoryHandler(w http.ResponseWriter, r *http.Request) {
	category := r.PathValue("category")

	items, err := s.db.GetItemsByCategory(category)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(items); err != nil {
		slog.Error("encoding items response", "error", err)
	}
}
