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

func TestArtworkPageVisibleEscapesAndHiddenNotFound(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := artwork.NewRepository(db.DB())
	visible, err := r.Create(context.Background(), artwork.Input{Name: `<Night & Day>`, Date: "2020-01-02", Tags: []string{"blue sky"}, Surface: "paper", Medium: "ink", Visible: true})
	if err != nil {
		t.Fatal(err)
	}
	hidden, err := r.Create(context.Background(), artwork.Input{Name: "Hidden", Date: "2020-01-02", Surface: "paper", Medium: "ink"})
	if err != nil {
		t.Fatal(err)
	}
	h := ArtworkPage{Repository: r}
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/artwork/"+visible.Slug, nil))
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || !strings.Contains(body, "&lt;Night &amp; Day&gt;") || !strings.Contains(body, "/gallery/?tag=blue+sky") || !strings.Contains(body, "srcset=") {
		t.Fatalf("unexpected page: %d %s", recorder.Code, body)
	}
	recorder = httptest.NewRecorder()
	h.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/artwork/"+hidden.Slug, nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("hidden status = %d", recorder.Code)
	}
}
