package web

import (
	"html/template"
	"net/http"
	"path/filepath"
	"shared/golog"
	"strings"
)

var log = golog.New("Web")

type TemplateRenderer struct {
	templates map[string]*template.Template
}

type PageData struct {
	Title      string
	ActivePage string
	Stats      DashboardStats
	Flags      []FlagRow
}

type DashboardStats struct {
	FlagsSent     int
	FlagsAccepted int
	TeamsActive   int
}

type FlagRow struct {
	Flag    string
	Team    string
	Service string
	Status  string
}

func NewRenderer(templatesDir string) *TemplateRenderer {
	r := &TemplateRenderer{
		templates: make(map[string]*template.Template),
	}

	layout := filepath.Join(templatesDir, "layout.html")

	dashboardPages := []string{"dashboard", "grafici", "flags", "docs", "top"}
	for _, page := range dashboardPages {
		pageFile := filepath.Join(templatesDir, "pages", page+".html")
		tmpl := template.Must(template.ParseFiles(layout, pageFile))
		r.templates[page] = tmpl
	}

	loginFile := filepath.Join(templatesDir, "login.html")
	loginTmpl := template.Must(template.ParseFiles(loginFile))
	r.templates["login"] = loginTmpl

	log.Info("Loaded %d templates from %s", len(r.templates), templatesDir)
	return r
}

func (r *TemplateRenderer) Render(w http.ResponseWriter, page string, data PageData) {
	tmpl, ok := r.templates[page]
	if !ok {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func (r *TemplateRenderer) RenderLogin(w http.ResponseWriter) {
	tmpl, ok := r.templates["login"]
	if !ok {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func LoginHandler(renderer *TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderer.RenderLogin(w)
	}
}

func DashboardHandler(renderer *TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:      "Dashboard",
			ActivePage: "dashboard",
			Stats: DashboardStats{
				FlagsSent:     42,
				FlagsAccepted: 38,
				TeamsActive:   16,
			},
		}
		renderer.Render(w, "dashboard", data)
	}
}

func FlagsHandler(renderer *TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:      "Flags",
			ActivePage: "flags",
			Flags: []FlagRow{
				{Flag: "flag{abc123}", Team: "Team 1", Service: "webapp", Status: "accepted"},
				{Flag: "flag{def456}", Team: "Team 2", Service: "api", Status: "rejected"},
				{Flag: "flag{ghi789}", Team: "Team 3", Service: "db", Status: "pending"},
			},
		}
		renderer.Render(w, "flags", data)
	}
}

func DocsHandler(renderer *TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:      "Documentation",
			ActivePage: "docs",
		}
		renderer.Render(w, "docs", data)
	}
}

func GraficiHandler(renderer *TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:      "Charts",
			ActivePage: "grafici",
		}
		renderer.Render(w, "grafici", data)
	}
}

func TopHandler(renderer *TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:      "Top Attackers",
			ActivePage: "top",
		}
		renderer.Render(w, "top", data)
	}
}

func PageName(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}
