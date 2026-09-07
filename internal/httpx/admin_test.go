package httpx

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vm75/virtual-art-gallery/internal/artwork"
	"github.com/vm75/virtual-art-gallery/internal/auth"
	"github.com/vm75/virtual-art-gallery/internal/images"
	"github.com/vm75/virtual-art-gallery/internal/store"
)

func TestAdminArtworkCreateEditAndProtectedAccess(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := artwork.NewRepository(db.DB())
	manager := auth.NewManager(db.DB(), false)
	if err := manager.Setup(context.Background(), "admin", "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	h := AdminHandler{Auth: manager, Artworks: r, Images: images.Pipeline{Root: t.TempDir()}}
	unauth := httptest.NewRecorder()
	h.ServeHTTP(unauth, httptest.NewRequest(http.MethodGet, "/admin/", nil))
	if unauth.Code != http.StatusSeeOther {
		t.Fatalf("unauthenticated status = %d", unauth.Code)
	}
	session, csrf, err := manager.Login(context.Background(), "admin", "correct horse battery staple", "test")
	if err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("name", "Admin Work")
	_ = writer.WriteField("date", "2024-01-01")
	_ = writer.WriteField("tags", "one, two")
	_ = writer.WriteField("surface", "canvas")
	_ = writer.WriteField("medium", "oil")
	_ = writer.WriteField("visible", "on")
	_ = writer.WriteField("csrf_token", csrf)
	file, _ := writer.CreateFormFile("image", "work.png")
	_ = png.Encode(file, image.NewRGBA(image.Rect(0, 0, 10, 10)))
	writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/admin/artworks/new", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-CSRF-Token", csrf)
	req.AddCookie(&http.Cookie{Name: "gallery_session", Value: session})
	req.AddCookie(&http.Cookie{Name: "gallery_csrf", Value: csrf})
	response := httptest.NewRecorder()
	h.ServeHTTP(response, req)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, body=%s", response.Code, response.Body.String())
	}
	item, err := r.Get(context.Background(), "admin-work", false)
	if err != nil || !item.Visible || item.Image.Thumbnail == "" {
		t.Fatalf("created artwork = %+v, err=%v", item, err)
	}
	edit := httptest.NewRequest(http.MethodPost, "/admin/artworks/edit?slug=admin-work", bytes.NewBufferString("name=Edited&date=2024-01-01&surface=canvas&medium=oil&csrf_token="+csrf))
	edit.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	edit.Header.Set("X-CSRF-Token", csrf)
	edit.AddCookie(&http.Cookie{Name: "gallery_session", Value: session})
	edit.AddCookie(&http.Cookie{Name: "gallery_csrf", Value: csrf})
	response = httptest.NewRecorder()
	h.ServeHTTP(response, edit)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("edit status = %d", response.Code)
	}
}
