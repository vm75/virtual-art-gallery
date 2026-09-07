package httpx

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/vm75/virtual-art-gallery/internal/artwork"
)

type GalleryPage struct{ Repository *artwork.Repository }

func (p GalleryPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	items, err := p.Repository.ListPublic(r.Context(), r.URL.Query().Get("tag"), r.URL.Query().Get("surface"), r.URL.Query().Get("medium"), "desc")
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	var cards strings.Builder
	for _, item := range items {
		name := template.HTMLEscapeString(item.Name)
		slug := template.URLQueryEscaper(item.Slug)
		cards.WriteString(`<article class="gallery-card"><a class="gallery-image" href="/artwork/` + slug + `" data-lightbox data-name="` + name + `"><img src="` + template.HTMLEscapeString(item.Image.Thumbnail) + `" srcset="` + template.HTMLEscapeString(item.Image.Thumbnail) + ` 480w, ` + template.HTMLEscapeString(item.Image.Medium) + ` 1200w" sizes="(min-width: 900px) 30vw, 90vw" alt="` + template.HTMLEscapeString(item.AltText) + `" loading="lazy"></a><div class="gallery-card-copy"><h2><a href="/artwork/` + slug + `">` + name + `</a></h2><p>` + template.HTMLEscapeString(item.Date) + `</p></div></article>`)
	}
	if len(items) == 0 {
		cards.WriteString(`<p class="empty-state">No visible artworks match these filters.</p>`)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Gallery Lite — Virtual Art Gallery</title><link rel="stylesheet" href="/static/style.css"><script type="module" src="/static/gallery.js"></script></head><body><header class="site-header"><a class="brand" href="/">Virtual Art Gallery</a><a href="/timeline/">Timeline</a></header><main class="gallery-page"><section class="page-intro"><p class="eyebrow">01 / Browse</p><h1>Gallery Lite</h1><p>Every visible work, simply arranged.</p></section><form class="filters" method="get" action="/gallery/"><label>Tag <input name="tag" value="` + template.HTMLEscapeString(r.URL.Query().Get("tag")) + `"></label><label>Surface <input name="surface" value="` + template.HTMLEscapeString(r.URL.Query().Get("surface")) + `"></label><label>Medium <input name="medium" value="` + template.HTMLEscapeString(r.URL.Query().Get("medium")) + `"></label><button>Filter</button><a href="/gallery/">Clear</a></form><section class="gallery-grid" aria-label="Artwork collection">` + cards.String() + `</section></main><dialog id="lightbox" aria-label="Artwork preview"><button class="dialog-close" autofocus>Close</button><button class="dialog-prev" aria-label="Previous artwork">Previous</button><button class="dialog-next" aria-label="Next artwork">Next</button><img alt=""><p><a class="dialog-detail">View artwork details</a></p></dialog></body></html>`))
}
