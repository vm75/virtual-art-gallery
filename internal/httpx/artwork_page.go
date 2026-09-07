package httpx

import (
	"database/sql"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/vm75/virtual-art-gallery/internal/artwork"
)

type ArtworkPage struct{ Repository *artwork.Repository }

func (p ArtworkPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/artwork/")
	item, err := p.Repository.Get(r.Context(), slug, true)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	title := template.HTMLEscapeString(item.Name)
	image := template.HTMLEscapeString(item.Image.Large)
	if image == "" {
		image = template.HTMLEscapeString(item.Image.Medium)
	}
	srcset := template.HTMLEscapeString(item.Image.Thumbnail) + " 480w, " + template.HTMLEscapeString(item.Image.Medium) + " 1200w, " + template.HTMLEscapeString(item.Image.Large) + " 2400w"
	var tags strings.Builder
	for _, tag := range item.Tags {
		tags.WriteString(`<a class="chip" href="/gallery/?tag=` + template.URLQueryEscaper(tag) + `">` + template.HTMLEscapeString(tag) + `</a> `)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>` + title + ` — Virtual Art Gallery</title><meta name="description" content="Artwork detail for ` + title + `"><link rel="canonical" href="/artwork/` + template.URLQueryEscaper(item.Slug) + `"><link rel="stylesheet" href="/static/style.css"></head><body><header class="site-header"><a class="brand" href="/">Virtual Art Gallery</a><a href="/gallery/">All artworks</a></header><main class="artwork-detail"><figure><img src="` + image + `" srcset="` + srcset + `" sizes="(min-width: 900px) 70vw, 100vw" alt="` + template.HTMLEscapeString(item.AltText) + `" loading="eager"></figure><article><p class="eyebrow">Artwork</p><h1>` + title + `</h1><dl><dt>Date</dt><dd>` + template.HTMLEscapeString(item.Date) + `</dd><dt>Surface</dt><dd><a href="/gallery/?surface=` + template.URLQueryEscaper(item.Surface) + `">` + template.HTMLEscapeString(item.Surface) + `</a></dd><dt>Medium</dt><dd><a href="/gallery/?medium=` + template.URLQueryEscaper(item.Medium) + `">` + template.HTMLEscapeString(item.Medium) + `</a></dd><dt>Tags</dt><dd>` + tags.String() + `</dd></dl></article></main></body></html>`))
}
