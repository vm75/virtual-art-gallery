package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMuseumPageLoadsOnlyMuseumEnhancementAndFallback(t *testing.T) {
	recorder := httptest.NewRecorder()
	MuseumPage(recorder, httptest.NewRequest(http.MethodGet, "/museum/", nil))
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || !strings.Contains(body, "/static/museum.js") || !strings.Contains(body, "/gallery/") || strings.Contains(body, "ARTIC") {
		t.Fatalf("unexpected museum page: %s", body)
	}
}
