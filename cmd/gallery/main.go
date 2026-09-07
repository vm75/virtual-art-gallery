package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vm75/virtual-art-gallery/internal/artwork"
	"github.com/vm75/virtual-art-gallery/internal/auth"
	"github.com/vm75/virtual-art-gallery/internal/config"
	"github.com/vm75/virtual-art-gallery/internal/httpx"
	"github.com/vm75/virtual-art-gallery/internal/images"
	"github.com/vm75/virtual-art-gallery/internal/museum"
	"github.com/vm75/virtual-art-gallery/internal/store"
	webassets "github.com/vm75/virtual-art-gallery/web"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg, err := config.FromEnv()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(2)
	}
	database, err := store.Open(context.Background(), cfg.DataDir)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(2)
	}
	defer database.Close()
	artworkAPI := httpx.ArtworkAPI{Repository: artwork.NewRepository(database.DB())}
	artworkPage := httpx.ArtworkPage{Repository: artwork.NewRepository(database.DB())}
	galleryPage := httpx.GalleryPage{Repository: artwork.NewRepository(database.DB())}
	timelinePage := httpx.TimelinePage{Repository: artwork.NewRepository(database.DB())}
	museumService := museum.NewService(database.DB(), artwork.NewRepository(database.DB()))
	adminHandler := httpx.AdminHandler{Auth: auth.NewManager(database.DB(), cfg.SecureCookies), Artworks: artwork.NewRepository(database.DB()), Images: images.Pipeline{Root: cfg.DataDir}, Museum: museumService}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", httpx.HealthHandler)
	mux.HandleFunc("GET /api/artworks", artworkAPI.List)
	mux.HandleFunc("GET /api/artworks/{slug}", artworkAPI.Detail)
	mux.Handle("/artwork/", artworkPage)
	mux.Handle("/gallery/", galleryPage)
	mux.Handle("/timeline/", timelinePage)
	mux.HandleFunc("GET /museum/", httpx.MuseumPage)
	mux.HandleFunc("GET /api/museum", httpx.MuseumAPI{Service: museumService}.ServeHTTP)
	mux.Handle("/media/", httpx.MediaHandler(cfg.DataDir))
	mux.Handle("/admin/", adminHandler)
	mux.HandleFunc("GET /{$}", httpx.HomeHandler)
	staticFS, err := fs.Sub(webassets.Static, "static")
	if err != nil {
		logger.Error("load static assets", "error", err)
		os.Exit(2)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", httpx.ImmutableStatic(staticFS)))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })

	server := &http.Server{Addr: cfg.ListenAddr, Handler: httpx.SecurityHeaders(mux), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		logger.Info("server listening", "address", cfg.ListenAddr, "data_dir", cfg.DataDir)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", serveErr)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}
