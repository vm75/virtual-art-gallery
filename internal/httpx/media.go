package httpx

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func MediaHandler(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		relative := strings.TrimPrefix(r.URL.Path, "/media/")
		if relative == "" || filepath.Ext(relative) != ".jpg" || strings.Contains(relative, "..") || strings.HasSuffix(relative, "/original.jpg") {
			http.NotFound(w, r)
			return
		}
		path := filepath.Join(root, filepath.FromSlash(relative))
		file, err := os.Open(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("Content-Type", "image/jpeg")
		http.ServeContent(w, r, filepath.Base(path), info.ModTime(), file)
	}
}
