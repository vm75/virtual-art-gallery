package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vm75/virtual-art-gallery/internal/artwork"
	"github.com/vm75/virtual-art-gallery/internal/store"
)

func TestArtworkAPIHidesDraftAndNotFound(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := artwork.NewRepository(db.DB())
	item, err := r.Create(context.Background(), artwork.Input{Name: "<Gallery>", Date: "2024-01-01", Surface: "paper", Medium: "ink", Visible: true})
	if err != nil {
		t.Fatal(err)
	}
	api := ArtworkAPI{Repository: r}
	recorder := httptest.NewRecorder()
	api.Detail(recorder, httptest.NewRequest(http.MethodGet, "/api/artworks/"+item.Slug, nil))
	if recorder.Code != http.StatusOK || recorder.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("unexpected detail response: %d %s", recorder.Code, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	api.Detail(recorder, httptest.NewRequest(http.MethodGet, "/api/artworks/missing", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d", recorder.Code)
	}
}
