package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vm75/virtual-art-gallery/internal/artwork"
	"github.com/vm75/virtual-art-gallery/internal/auth"
	"github.com/vm75/virtual-art-gallery/internal/images"
	"github.com/vm75/virtual-art-gallery/internal/museum"
	"html/template"
	"net/http"
	"strings"
)

type AdminHandler struct {
	Auth     *auth.Manager
	Artworks *artwork.Repository
	Images   images.Pipeline
	Museum   *museum.Service
}

func (h AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin/logout" && r.Method == http.MethodPost {
		if !h.Auth.Authenticate(r.Context(), r) || !h.Auth.ValidateCSRF(r.Context(), r) {
			http.Error(w, "forbidden", 403)
			return
		}
		h.Auth.Logout(r.Context(), r)
		h.Auth.ClearCookies(w)
		http.Redirect(w, r, "/admin/", 303)
		return
	}
	needsSetup, err := h.Auth.NeedsSetup(r.Context())
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	if r.URL.Path == "/admin/setup" {
		if !needsSetup {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodPost {
			h.setup(w, r)
			return
		}
		renderAdmin(w, "Set up administrator", "setup", "")
		return
	}
	if r.URL.Path == "/admin/login" {
		if needsSetup {
			http.Redirect(w, r, "/admin/setup", 303)
			return
		}
		if r.Method == http.MethodPost {
			h.login(w, r)
			return
		}
		renderAdmin(w, "Administrator login", "login", "")
		return
	}
	if !h.Auth.Authenticate(r.Context(), r) {
		if needsSetup {
			http.Redirect(w, r, "/admin/setup", 303)
		} else {
			http.Redirect(w, r, "/admin/login", 303)
		}
		return
	}
	if strings.HasPrefix(r.URL.Path, "/admin/artworks") {
		h.artworks(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/admin/museum") {
		h.museum(w, r)
		return
	}
	csrf := csrfFrom(r)
	items, err := h.Artworks.ListAll(r.Context())
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	var list strings.Builder
	list.WriteString("<ul>")
	for _, item := range items {
		status := "hidden"
		if item.Visible {
			status = "visible"
		}
		list.WriteString("<li><a href=\"/admin/artworks/edit?slug=" + template.URLQueryEscaper(item.Slug) + "\">" + template.HTMLEscapeString(item.Name) + "</a> — " + status + "</li>")
	}
	list.WriteString("</ul>")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><title>Admin</title><main><h1>Administrator</h1><p><a href="/admin/artworks/new">Add artwork</a></p>` + list.String() + `<form method="post" action="/admin/logout"><input type="hidden" name="csrf_token" value="` + template.HTMLEscapeString(csrf) + `"><button>Log out</button></form></main>`))
}

func (h AdminHandler) museum(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if !h.Auth.ValidateCSRF(r.Context(), r) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		var set museum.RuleSet
		if err := json.Unmarshal([]byte(r.FormValue("rules_json")), &set); err != nil {
			renderMuseumAdmin(w, csrfFrom(r), r.FormValue("rules_json"), "Rules must be valid JSON")
			return
		}
		if err := h.Museum.SaveDraft(r.Context(), set); err != nil {
			renderMuseumAdmin(w, csrfFrom(r), r.FormValue("rules_json"), err.Error())
			return
		}
		if r.URL.Path == "/admin/museum/publish" {
			if err := h.Museum.Publish(r.Context()); err != nil {
				renderMuseumAdmin(w, csrfFrom(r), r.FormValue("rules_json"), err.Error())
				return
			}
		}
		http.Redirect(w, r, "/admin/museum", http.StatusSeeOther)
		return
	}
	set, err := h.Museum.Draft(r.Context())
	if err != nil {
		set = museum.RuleSet{Version: 1, Seed: 1, Groups: []museum.Group{{ID: "all", Name: "All works"}}, Rules: []museum.Rule{{ID: "all", Priority: 1, All: []museum.Condition{{Field: "medium", Op: "contains", Value: "oil"}}, Group: "all"}}}
	}
	data, _ := json.MarshalIndent(set, "", "  ")
	assignment, plan, previewErr := h.Museum.Preview(r.Context(), set)
	preview := "Preview unavailable"
	if previewErr == nil {
		preview = fmt.Sprintf("Preview: %d unclassified works; %d layout errors", len(assignment.Unclassified), len(plan.Errors))
	}
	renderMuseumAdmin(w, csrfFrom(r), string(data), "", preview)
}

func renderMuseumAdmin(w http.ResponseWriter, csrf, rules, errorText string, preview ...string) {
	errorHTML := ""
	if errorText != "" {
		errorHTML = `<p role="alert">` + template.HTMLEscapeString(errorText) + `</p>`
	}
	previewHTML := ""
	if len(preview) > 0 && preview[0] != "" {
		previewHTML = `<p>` + template.HTMLEscapeString(preview[0]) + `</p>`
	}
	html := `<!doctype html><title>Museum rules</title><main><a href="/admin/">Admin</a><h1>Museum rules</h1>` + errorHTML + previewHTML + `<p>Draft rules are validated before saving. Publish explicitly to change the public museum.</p><form method="post" action="/admin/museum/save"><input type="hidden" name="csrf_token" value="` + template.HTMLEscapeString(csrf) + `"><label>Rules JSON <textarea name="rules_json" rows="24" cols="80">` + template.HTMLEscapeString(rules) + `</textarea></label><button>Save draft</button></form><form method="post" action="/admin/museum/publish"><input type="hidden" name="csrf_token" value="` + template.HTMLEscapeString(csrf) + `"><input type="hidden" name="rules_json" value="` + template.HTMLEscapeString(rules) + `"><button>Publish draft</button></form></main>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

func (h AdminHandler) artworks(w http.ResponseWriter, r *http.Request) {
	csrf := csrfFrom(r)
	if r.URL.Path == "/admin/artworks/new" {
		if r.Method == http.MethodPost {
			h.createArtwork(w, r)
			return
		}
		h.renderArtworkForm(w, csrf, artwork.Input{}, "", true, "")
		return
	}
	slug := r.URL.Query().Get("slug")
	item, err := h.Artworks.Get(r.Context(), slug, false)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodPost {
		if !h.Auth.ValidateCSRF(r.Context(), r) {
			http.Error(w, "forbidden", 403)
			return
		}
		in := inputFromRequest(r)
		if _, err := h.Artworks.Update(r.Context(), slug, in); err != nil {
			h.renderArtworkForm(w, csrf, in, err.Error(), false, slug)
			return
		}
		http.Redirect(w, r, "/admin/", 303)
		return
	}
	h.renderArtworkForm(w, csrf, artwork.Input{Name: item.Name, Date: item.Date, Tags: item.Tags, Surface: item.Surface, Medium: item.Medium, AltText: item.AltText, Visible: item.Visible}, "", false, slug)
}
func (h AdminHandler) createArtwork(w http.ResponseWriter, r *http.Request) {
	if !h.Auth.ValidateCSRF(r.Context(), r) {
		http.Error(w, "forbidden", 403)
		return
	}
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		h.renderArtworkForm(w, csrfFrom(r), inputFromRequest(r), "Upload is too large or malformed", true, "")
		return
	}
	in := inputFromRequest(r)
	item, err := h.Artworks.Create(r.Context(), in)
	if err != nil {
		h.renderArtworkForm(w, csrfFrom(r), in, err.Error(), true, "")
		return
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		_ = h.Artworks.Delete(r.Context(), item.Slug)
		h.renderArtworkForm(w, csrfFrom(r), in, "An image is required", true, "")
		return
	}
	result, err := h.Images.Process(file, item.ID)
	file.Close()
	if err != nil {
		_ = h.Artworks.Delete(r.Context(), item.Slug)
		h.renderArtworkForm(w, csrfFrom(r), in, err.Error(), true, "")
		return
	}
	if err := h.Artworks.SetImages(r.Context(), item.Slug, result); err != nil {
		_ = h.Artworks.Delete(r.Context(), item.Slug)
		h.renderArtworkForm(w, csrfFrom(r), in, "Unable to save image", true, "")
		return
	}
	http.Redirect(w, r, "/admin/", 303)
}
func inputFromRequest(r *http.Request) artwork.Input {
	return artwork.Input{Name: r.FormValue("name"), Date: r.FormValue("date"), Tags: strings.Split(r.FormValue("tags"), ","), Surface: r.FormValue("surface"), Medium: r.FormValue("medium"), AltText: r.FormValue("alt_text"), Visible: r.FormValue("visible") == "on"}
}
func csrfFrom(r *http.Request) string {
	if c, err := r.Cookie("gallery_csrf"); err == nil {
		return c.Value
	}
	return ""
}
func (h AdminHandler) renderArtworkForm(w http.ResponseWriter, csrf string, in artwork.Input, errorText string, create bool, slug string) {
	surfaces, mediums, _ := h.Artworks.Taxonomy(context.Background())
	renderArtworkFormHTML(w, csrf, in, errorText, create, slug, surfaces, mediums)
}

func renderArtworkFormHTML(w http.ResponseWriter, csrf string, in artwork.Input, errorText string, create bool, slug string, surfaces, mediums []string) {
	action, title := "/admin/artworks/new", "New artwork"
	enctype := ` enctype="multipart/form-data"`
	image := `<label>Image <input type="file" name="image" accept="image/jpeg,image/png,image/gif" required></label>`
	if !create {
		action, title, enctype, image = "/admin/artworks/edit?slug="+template.URLQueryEscaper(slug), "Edit artwork", "", ""
	}
	checked := ""
	if in.Visible {
		checked = " checked"
	}
	e := ""
	if errorText != "" {
		e = `<p role="alert">` + template.HTMLEscapeString(errorText) + `</p>`
	}
	options := `<datalist id="surfaces">`
	for _, value := range surfaces {
		options += `<option value="` + template.HTMLEscapeString(value) + `">`
	}
	options += `</datalist><datalist id="mediums">`
	for _, value := range mediums {
		options += `<option value="` + template.HTMLEscapeString(value) + `">`
	}
	options += `</datalist>`
	html := `<!doctype html><title>` + title + `</title><main><a href="/admin/">Admin</a><h1>` + title + `</h1>` + e + `<form method="post" action="` + action + `"` + enctype + `><input type="hidden" name="csrf_token" value="` + template.HTMLEscapeString(csrf) + `"><label>Name <input name="name" required value="` + template.HTMLEscapeString(in.Name) + `"></label><label>Date <input type="date" name="date" required value="` + template.HTMLEscapeString(in.Date) + `"></label><label>Tags <input name="tags" value="` + template.HTMLEscapeString(strings.Join(in.Tags, ", ")) + `" placeholder="comma separated"></label><label>Surface <input name="surface" list="surfaces" required value="` + template.HTMLEscapeString(in.Surface) + `"></label><label>Medium <input name="medium" list="mediums" required value="` + template.HTMLEscapeString(in.Medium) + `"></label><label>Alt text <textarea name="alt_text" maxlength="1000">` + template.HTMLEscapeString(in.AltText) + `</textarea></label>` + image + `<label>Visible <input type="checkbox" name="visible"` + checked + `></label><button>Save artwork</button></form>` + options + `</main>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}
func (h AdminHandler) setup(w http.ResponseWriter, r *http.Request) {
	if err := h.Auth.Setup(r.Context(), r.FormValue("username"), r.FormValue("password")); err != nil {
		renderAdmin(w, "Set up administrator", "setup", err.Error())
		return
	}
	h.login(w, r)
}
func (h AdminHandler) login(w http.ResponseWriter, r *http.Request) {
	session, csrf, err := h.Auth.Login(r.Context(), r.FormValue("username"), r.FormValue("password"), r.RemoteAddr)
	if err != nil {
		renderAdmin(w, "Administrator login", "login", "Login failed")
		return
	}
	h.Auth.SetCookies(w, session, csrf)
	http.Redirect(w, r, "/admin/", 303)
}
func renderAdmin(w http.ResponseWriter, title, mode, errorText string) {
	action, label := "/admin/setup", "Create administrator"
	if mode == "login" {
		action, label = "/admin/login", "Sign in"
	}
	e := ""
	if errorText != "" {
		e = `<p role="alert">` + template.HTMLEscapeString(errorText) + `</p>`
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><title>` + template.HTMLEscapeString(title) + `</title><main><h1>` + template.HTMLEscapeString(title) + `</h1>` + e + `<form method="post" action="` + action + `"><label>Username <input name="username" autocomplete="username" required></label><label>Password <input type="password" name="password" autocomplete="new-password" required></label><button>` + label + `</button></form></main>`))
}
