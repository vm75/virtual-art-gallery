package httpx

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/vm75/virtual-art-gallery/internal/artwork"
)

type ArtworkAPI struct{ Repository *artwork.Repository }

func (api ArtworkAPI) List(w http.ResponseWriter, r *http.Request) {
	items, err := api.Repository.ListPublic(r.Context(), r.URL.Query().Get("tag"), r.URL.Query().Get("surface"), r.URL.Query().Get("medium"), r.URL.Query().Get("order"))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, items)
}

func (api ArtworkAPI) Detail(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/artworks/")
	item, err := api.Repository.Get(r.Context(), slug, true)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, item)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(value)
}
