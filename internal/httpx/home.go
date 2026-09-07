package httpx

import "net/http"

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Virtual Art Gallery</title><link rel="stylesheet" href="/static/style.css"></head><body><header class="site-header"><a class="brand" href="/">Virtual Art Gallery</a></header><main><section class="hero"><p class="eyebrow">A quiet place for looking</p><h1>Art, in three dimensions.</h1><p class="lede">Browse the collection quickly, follow it through time, or step into a living museum.</p></section><nav class="experiences" aria-label="Gallery experiences"><a class="experience" href="/gallery/"><span class="eyebrow">01 / Browse</span><h2>Gallery Lite</h2><p>A responsive, image-first collection.</p></a><a class="experience" href="/timeline/"><span class="eyebrow">02 / Follow</span><h2>Timeline</h2><p>Move through the work chronologically.</p></a><a class="experience" href="/museum/"><span class="eyebrow">03 / Enter</span><h2>Museum</h2><p>Explore a generated three-dimensional exhibition.</p></a></nav></main></body></html>`))
}
