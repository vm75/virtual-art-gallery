package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vm75/virtual-art-gallery/internal/artwork"
	"github.com/vm75/virtual-art-gallery/internal/store"
)

func TestGalleryUsesDerivedLazyImagesAndFilters(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := artwork.NewRepository(db.DB())
	if _, err := r.Create(context.Background(), artwork.Input{Name: "Visible", Date: "2024-01-01", Tags: []string{"blue"}, Surface: "canvas", Medium: "oil", Visible: true}); err != nil {
		t.Fatal(err)
	}
	h := GalleryPage{Repository: r}
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/gallery/?tag=blue", nil))
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || !strings.Contains(body, `loading="lazy"`) || !strings.Contains(body, `data-lightbox`) || !strings.Contains(body, `/artwork/visible`) || !strings.Contains(body, `name="tag"`) {
		t.Fatalf("unexpected gallery: %d %s", recorder.Code, body)
	}
}
