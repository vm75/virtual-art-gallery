package httpx

import "net/http"

func MuseumPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/museum/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Museum — Virtual Art Gallery</title><link rel="stylesheet" href="/static/style.css"><script type="module" src="/static/museum.js"></script></head><body><main class="museum-page"><header class="museum-header"><a class="brand" href="/">Virtual Art Gallery</a><p class="eyebrow">03 / Enter</p><h1>Museum</h1><p>Walk with W A S D or arrow keys. Drag the scene to look. Press Escape to release focus.</p></header><section class="museum-stage" aria-label="Three-dimensional museum"><canvas id="museum-canvas" tabindex="0" aria-label="Interactive 3D museum"></canvas><div id="museum-fallback"><p>3D museum mode is unavailable here. Browse the artworks instead.</p><a href="/gallery/">Open Gallery Lite</a></div></section><nav class="museum-controls" aria-label="Museum movement"><button data-move="forward">Forward</button><button data-move="left">Left</button><button data-move="back">Back</button><button data-move="right">Right</button></nav><dialog id="museum-info"><button class="info-close" autofocus>Close</button><h2></h2><dl><dt>Date</dt><dd data-field="date"></dd><dt>Surface</dt><dd data-field="surface"></dd><dt>Medium</dt><dd data-field="medium"></dd><dt>Tags</dt><dd data-field="tags"></dd></dl><p><a data-field="link">Open canonical artwork page</a></p></dialog><p><a href="/gallery/">Use the accessible gallery view</a></p></main></body></html>`))
}
