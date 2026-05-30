// Package web contiene el servidor HTTP del acortador.
package web

import (
	"context"
	"embed"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/samuelcatalanz123/urlshortener-go/internal/store"
)

//go:embed templates/*.html static/*
var files embed.FS

// Handler sirve las páginas del acortador.
type Handler struct {
	store *store.Store
	tmpl  *template.Template
}

// New crea el Handler con sus plantillas ya parseadas.
func New(s *store.Store) (*Handler, error) {
	tmpl, err := template.ParseFS(files, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Handler{store: s, tmpl: tmpl}, nil
}

// Routes monta las rutas HTTP.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServerFS(files))
	mux.HandleFunc("GET /{$}", h.home)
	mux.HandleFunc("POST /shorten", h.shorten)
	mux.HandleFunc("GET /{code}", h.redirect)
	return mux
}

// homeData son los datos que recibe la plantilla home.html.
type homeData struct {
	Recent []store.Link
	Error  string
}

func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	links, err := h.store.Recent(r.Context(), 10)
	if err != nil {
		http.Error(w, "error del servidor", http.StatusInternalServerError)
		return
	}
	h.render(w, http.StatusOK, homeData{Recent: links})
}

func (h *Handler) render(w http.ResponseWriter, status int, data homeData) {
	w.WriteHeader(status)
	if err := h.tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, "error del servidor", http.StatusInternalServerError)
	}
}

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	long := strings.TrimSpace(r.FormValue("url"))
	if !validURL(long) {
		links, _ := h.store.Recent(r.Context(), 10)
		h.render(w, http.StatusBadRequest, homeData{
			Recent: links,
			Error:  "La URL debe empezar por http:// o https://",
		})
		return
	}
	code, err := h.uniqueCode(r.Context())
	if err != nil {
		http.Error(w, "error del servidor", http.StatusInternalServerError)
		return
	}
	if err := h.store.Save(r.Context(), code, long); err != nil {
		http.Error(w, "error del servidor", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// uniqueCode genera un código que no exista ya (reintenta ante colisión).
func (h *Handler) uniqueCode(ctx context.Context) (string, error) {
	for i := 0; i < 5; i++ {
		code := GenerateCode(6)
		exists, err := h.store.Exists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", errors.New("no se pudo generar un código único")
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	link, err := h.store.Get(r.Context(), code)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "error del servidor", http.StatusInternalServerError)
		return
	}
	_ = h.store.IncrClicks(r.Context(), code)
	http.Redirect(w, r, link.URL, http.StatusMovedPermanently)
}

// validURL acepta solo http(s).
func validURL(u string) bool {
	return strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")
}
