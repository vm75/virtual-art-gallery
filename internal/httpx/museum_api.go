package httpx

import (
	"database/sql"
	"errors"
	"github.com/vm75/virtual-art-gallery/internal/museum"
	"net/http"
)

type MuseumAPI struct{ Service *museum.Service }

func (a MuseumAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	scene, err := a.Service.Scene(r.Context())
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "museum is not published", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	writeJSON(w, scene)
}
