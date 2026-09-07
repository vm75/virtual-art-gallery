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

func TestTimelineIsChronologicalAndLazy(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := artwork.NewRepository(db.DB())
	for _, item := range []artwork.Input{{Name: "Later", Date: "2024-01-01", Surface: "canvas", Medium: "oil", Visible: true}, {Name: "Earlier", Date: "2020-01-01", Surface: "canvas", Medium: "oil", Visible: true}} {
		if _, err := r.Create(context.Background(), item); err != nil {
			t.Fatal(err)
		}
	}
	recorder := httptest.NewRecorder()
	(TimelinePage{Repository: r}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/timeline/", nil))
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || strings.Index(body, "Earlier") > strings.Index(body, "Later") || !strings.Contains(body, `loading="lazy"`) || !strings.Contains(body, `class="timeline"`) {
		t.Fatalf("invalid timeline: %d", recorder.Code)
	}
}
