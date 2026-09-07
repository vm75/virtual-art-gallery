package httpx

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/vm75/virtual-art-gallery/internal/artwork"
)

type TimelinePage struct{ Repository *artwork.Repository }

func (p TimelinePage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	items, err := p.Repository.ListPublic(r.Context(), "", "", "", "asc")
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	var cards strings.Builder
	for _, item := range items {
		cards.WriteString(`<article class="timeline-card"><p class="timeline-date">` + template.HTMLEscapeString(item.Date) + `</p><a href="/artwork/` + template.URLQueryEscaper(item.Slug) + `"><img src="` + template.HTMLEscapeString(item.Image.Medium) + `" srcset="` + template.HTMLEscapeString(item.Image.Thumbnail) + ` 480w, ` + template.HTMLEscapeString(item.Image.Medium) + ` 1200w" sizes="min(78vw, 48rem)" loading="lazy" alt="` + template.HTMLEscapeString(item.AltText) + `"><h2>` + template.HTMLEscapeString(item.Name) + `</h2></a></article>`)
	}
	if len(items) == 0 {
		cards.WriteString(`<p class="empty-state">The timeline is waiting for its first visible artwork.</p>`)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Timeline — Virtual Art Gallery</title><link rel="stylesheet" href="/static/style.css"><script type="module" src="/static/timeline.js"></script></head><body><header class="site-header"><a class="brand" href="/">Virtual Art Gallery</a><a href="/gallery/">Gallery Lite</a></header><main class="timeline-page"><section class="page-intro"><p class="eyebrow">02 / Follow</p><h1>Timeline</h1><p>From earliest mark to latest work.</p></section><section class="timeline" aria-label="Artwork timeline" tabindex="0">` + cards.String() + `</section></main></body></html>`))
}
