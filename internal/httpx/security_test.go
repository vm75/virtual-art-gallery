package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestSecurityHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	for _, name := range []string{"Content-Security-Policy", "X-Content-Type-Options", "Referrer-Policy", "Permissions-Policy"} {
		if recorder.Header().Get(name) == "" {
			t.Fatalf("missing %s", name)
		}
	}
}

func TestRevalidatingStaticSetsCacheHeader(t *testing.T) {
	fileSystem := fstest.MapFS{"asset.css": &fstest.MapFile{Data: []byte("body{}")}}
	recorder := httptest.NewRecorder()
	RevalidatingStatic(fileSystem).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/asset.css", nil))
	if got, want := recorder.Header().Get("Cache-Control"), "public, max-age=0, must-revalidate"; got != want {
		t.Fatalf("cache header = %q, want %q", got, want)
	}
}
