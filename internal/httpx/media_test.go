package httpx

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestMediaHandlerServesDerivativeOnly(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "images", "7-token", "thumbnail.jpg")
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("jpeg"), 0o640); err != nil {
		t.Fatal(err)
	}
	handler := MediaHandler(root)
	recorder := httptest.NewRecorder()
	handler(recorder, httptest.NewRequest(http.MethodGet, "/media/images/7-token/thumbnail.jpg", nil))
	if recorder.Code != http.StatusOK || recorder.Header().Get("Cache-Control") == "" {
		t.Fatalf("unexpected media response: %d", recorder.Code)
	}
	recorder = httptest.NewRecorder()
	handler(recorder, httptest.NewRequest(http.MethodGet, "/media/images/7-token/original.jpg", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("original status = %d", recorder.Code)
	}
}
