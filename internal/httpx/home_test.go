package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHomeLinksOnlyPublicExperiences(t *testing.T) {
	recorder := httptest.NewRecorder()
	HomeHandler(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	body := recorder.Body.String()
	for _, path := range []string{"/gallery/", "/timeline/", "/museum/"} {
		if !strings.Contains(body, path) {
			t.Fatalf("home missing %s", path)
		}
	}
	if strings.Contains(body, "/admin") {
		t.Fatal("home exposes admin")
	}
}
