# CortaURL — Plan de implementación

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Un acortador de URLs web en Go que guarda enlaces y cuenta clics en Redis.

**Architecture:** Servidor `net/http` con plantillas `html/template`. Capa `store` sobre Redis (`go-redis`) con tipos de dominio; capa `web` con handlers, generación de código y plantillas embebidas. Pruebas con `miniredis` (sin servidor real) y `httptest`.

**Tech Stack:** Go 1.26 · `github.com/redis/go-redis/v9` · `github.com/alicebob/miniredis/v2` · html/template · go:embed.

Módulo: `github.com/samuelcatalanz123/urlshortener-go`

**Commits:** cada commit con autoría de Samuel:
`git -c user.name="Samuel Catalán" -c user.email="samuelcatalanz123@gmail.com" commit -m "..."`

Verificación tras cada tarea con código Go: `go build ./... && go vet ./... && go test ./...`

---

## File Structure

```
urlshortener-go/
  go.mod / go.sum
  main.go                       Arranque: conecta a Redis, monta rutas, sirve.
  .gitignore
  README.md
  .github/workflows/ci.yml
  internal/store/
    store.go                    Store sobre Redis: Save, Get, IncrClicks, Recent, Exists.
    store_test.go               Pruebas con miniredis.
  internal/web/
    shortcode.go                GenerateCode(n).
    shortcode_test.go
    handler.go                  Handler: rutas, formulario, redirección.
    handler_test.go             Pruebas con miniredis + httptest.
    templates/home.html
    static/style.css
```

---

### Task 1: Scaffold del proyecto (módulo, git, dependencias)

**Files:**
- Create: `go.mod`, `.gitignore`

- [ ] **Step 1: Inicializar git y el módulo**

```bash
cd /Users/mqr93ea/Repos/urlshortener-go
git init
go mod init github.com/samuelcatalanz123/urlshortener-go
```

- [ ] **Step 2: Crear `.gitignore`**

```
/urlshortener-go
*.test
/tmp/
.DS_Store
```

- [ ] **Step 3: Añadir dependencias**

```bash
go get github.com/redis/go-redis/v9
go get github.com/alicebob/miniredis/v2
```

Expected: `go.mod` y `go.sum` con ambas dependencias.

- [ ] **Step 4: Commit**

```bash
git add -A
git -c user.name="Samuel Catalán" -c user.email="samuelcatalanz123@gmail.com" commit -m "chore: inicializar módulo y dependencias"
```

---

### Task 2: Generación del código corto (TDD)

**Files:**
- Create: `internal/web/shortcode.go`, `internal/web/shortcode_test.go`

- [ ] **Step 1: Escribir el test que falla**

`internal/web/shortcode_test.go`:
```go
package web

import (
	"strings"
	"testing"
)

func TestGenerateCodeLength(t *testing.T) {
	code := GenerateCode(6)
	if len(code) != 6 {
		t.Fatalf("esperaba longitud 6, obtuve %d (%q)", len(code), code)
	}
}

func TestGenerateCodeAlphabet(t *testing.T) {
	code := GenerateCode(20)
	for _, r := range code {
		if !strings.ContainsRune(alphabet, r) {
			t.Errorf("carácter inválido %q en %q", r, code)
		}
	}
}

func TestGenerateCodeDiffers(t *testing.T) {
	if GenerateCode(8) == GenerateCode(8) {
		t.Error("dos códigos consecutivos no deberían ser iguales")
	}
}
```

- [ ] **Step 2: Ejecutar y ver que falla**

Run: `go test ./internal/web/`
Expected: FAIL (undefined: GenerateCode / alphabet)

- [ ] **Step 3: Implementar**

`internal/web/shortcode.go`:
```go
package web

import "crypto/rand"

// alphabet es el conjunto base62 (sin caracteres ambiguos extra).
const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateCode devuelve un código aleatorio de n caracteres del alfabeto base62.
func GenerateCode(n int) string {
	b := make([]byte, n)
	// crypto/rand.Read no falla en la práctica en estos sistemas.
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}
```

- [ ] **Step 4: Ejecutar y ver que pasa**

Run: `go test ./internal/web/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add -A
git -c user.name="Samuel Catalán" -c user.email="samuelcatalanz123@gmail.com" commit -m "feat: generación de código corto base62"
```

---

### Task 3: Capa store sobre Redis (TDD con miniredis)

**Files:**
- Create: `internal/store/store.go`, `internal/store/store_test.go`

- [ ] **Step 1: Escribir el test que falla**

`internal/store/store_test.go`:
```go
package store

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("no se pudo arrancar miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return New(rdb)
}

func TestSaveAndGet(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if err := s.Save(ctx, "abc123", "https://example.com"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	link, err := s.Get(ctx, "abc123")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if link.URL != "https://example.com" {
		t.Errorf("URL = %q", link.URL)
	}
	if link.Clicks != 0 {
		t.Errorf("Clicks = %d, esperaba 0", link.Clicks)
	}
}

func TestGetNotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Get(context.Background(), "noexiste")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("esperaba ErrNotFound, obtuve %v", err)
	}
}

func TestIncrClicks(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	_ = s.Save(ctx, "abc123", "https://example.com")
	if err := s.IncrClicks(ctx, "abc123"); err != nil {
		t.Fatalf("IncrClicks: %v", err)
	}
	_ = s.IncrClicks(ctx, "abc123")
	link, _ := s.Get(ctx, "abc123")
	if link.Clicks != 2 {
		t.Errorf("Clicks = %d, esperaba 2", link.Clicks)
	}
}

func TestExists(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if ok, _ := s.Exists(ctx, "abc123"); ok {
		t.Error("no debería existir todavía")
	}
	_ = s.Save(ctx, "abc123", "https://example.com")
	if ok, _ := s.Exists(ctx, "abc123"); !ok {
		t.Error("debería existir tras Save")
	}
}

func TestRecent(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	_ = s.Save(ctx, "aaa", "https://a.com")
	_ = s.Save(ctx, "bbb", "https://b.com")
	links, err := s.Recent(ctx, 10)
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("esperaba 2 enlaces, obtuve %d", len(links))
	}
	if links[0].Code != "bbb" {
		t.Errorf("el más reciente debería ir primero, obtuve %q", links[0].Code)
	}
}
```

- [ ] **Step 2: Ejecutar y ver que falla**

Run: `go test ./internal/store/`
Expected: FAIL (undefined: Store, New, ErrNotFound)

- [ ] **Step 3: Implementar**

`internal/store/store.go`:
```go
// Package store guarda y recupera enlaces acortados en Redis.
package store

import (
	"context"
	"errors"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// ErrNotFound se devuelve cuando un código no existe.
var ErrNotFound = errors.New("código no encontrado")

// Link es un enlace acortado.
type Link struct {
	Code   string
	URL    string
	Clicks int
}

// Store guarda enlaces en Redis.
type Store struct {
	rdb *redis.Client
}

// New crea un Store sobre el cliente de Redis dado.
func New(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

// key es la clave del hash de un código.
func key(code string) string { return "short:" + code }

// Save guarda un enlace nuevo (clicks = 0) y lo añade a la lista de recientes.
func (s *Store) Save(ctx context.Context, code, url string) error {
	if err := s.rdb.HSet(ctx, key(code), "url", url, "clicks", 0).Err(); err != nil {
		return err
	}
	return s.rdb.LPush(ctx, "recent", code).Err()
}

// Exists indica si un código ya está en uso.
func (s *Store) Exists(ctx context.Context, code string) (bool, error) {
	n, err := s.rdb.Exists(ctx, key(code)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// Get devuelve el enlace de un código, o ErrNotFound si no existe.
func (s *Store) Get(ctx context.Context, code string) (Link, error) {
	m, err := s.rdb.HGetAll(ctx, key(code)).Result()
	if err != nil {
		return Link{}, err
	}
	if len(m) == 0 {
		return Link{}, ErrNotFound
	}
	clicks, _ := strconv.Atoi(m["clicks"])
	return Link{Code: code, URL: m["url"], Clicks: clicks}, nil
}

// IncrClicks suma 1 al contador de clics de un código.
func (s *Store) IncrClicks(ctx context.Context, code string) error {
	return s.rdb.HIncrBy(ctx, key(code), "clicks", 1).Err()
}

// Recent devuelve los últimos n enlaces creados (el más reciente primero).
func (s *Store) Recent(ctx context.Context, n int) ([]Link, error) {
	codes, err := s.rdb.LRange(ctx, "recent", 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}
	links := make([]Link, 0, len(codes))
	for _, code := range codes {
		link, err := s.Get(ctx, code)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}
```

- [ ] **Step 4: Ejecutar y ver que pasa**

Run: `go test ./internal/store/`
Expected: PASS (5 tests)

- [ ] **Step 5: Commit**

```bash
git add -A
git -c user.name="Samuel Catalán" -c user.email="samuelcatalanz123@gmail.com" commit -m "feat: capa store sobre Redis con pruebas (miniredis)"
```

---

### Task 4: Handler web, plantillas y estáticos (TDD con httptest)

**Files:**
- Create: `internal/web/handler.go`, `internal/web/handler_test.go`, `internal/web/templates/home.html`, `internal/web/static/style.css`

- [ ] **Step 1: Crear la plantilla `internal/web/templates/home.html`**

```html
<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>CortaURL — Acortador de enlaces</title>
  <link rel="stylesheet" href="/static/style.css">
</head>
<body>
  <main>
    <h1>🔗 CortaURL</h1>
    <p class="sub">Pega una URL larga y obtén un enlace corto.</p>

    {{if .Error}}<p class="error">{{.Error}}</p>{{end}}

    <form method="post" action="/shorten">
      <input type="text" name="url"
             placeholder="https://ejemplo.com/una-direccion-muy-larga" autofocus>
      <button type="submit">Acortar</button>
    </form>

    <h2>Enlaces recientes</h2>
    {{if .Recent}}
    <ul class="links">
      {{range .Recent}}
      <li>
        <a class="short" href="/{{.Code}}">/{{.Code}}</a>
        <span class="target">{{.URL}}</span>
        <span class="clicks">{{.Clicks}} clics</span>
      </li>
      {{end}}
    </ul>
    {{else}}
    <p class="empty">Todavía no hay enlaces. ¡Crea el primero!</p>
    {{end}}
  </main>
</body>
</html>
```

- [ ] **Step 2: Crear los estilos `internal/web/static/style.css`**

```css
:root { --accent: #4f46e5; --bg: #f8fafc; --card: #fff; --text: #1f2933; }
* { box-sizing: border-box; }
body {
  font-family: system-ui, -apple-system, "Segoe UI", sans-serif;
  background: var(--bg); color: var(--text); margin: 0; padding: 40px 16px;
}
main { max-width: 640px; margin: 0 auto; background: var(--card);
  padding: 32px; border-radius: 16px; box-shadow: 0 8px 30px rgba(0,0,0,.06); }
h1 { margin: 0 0 4px; }
.sub { color: #64748b; margin: 0 0 24px; }
form { display: flex; gap: 8px; margin-bottom: 28px; }
input[name=url] { flex: 1; padding: 12px 14px; border: 1px solid #cbd5e1;
  border-radius: 10px; font-size: 15px; }
input[name=url]:focus { outline: none; border-color: var(--accent); }
button { padding: 12px 20px; background: var(--accent); color: #fff; border: 0;
  border-radius: 10px; font-size: 15px; font-weight: 600; cursor: pointer; }
button:hover { background: #4338ca; }
h2 { font-size: 15px; text-transform: uppercase; letter-spacing: .05em;
  color: #64748b; }
.error { background: #fef2f2; color: #b91c1c; padding: 10px 14px;
  border-radius: 10px; }
.empty { color: #94a3b8; }
ul.links { list-style: none; padding: 0; }
ul.links li { display: flex; flex-direction: column; gap: 2px;
  padding: 12px 0; border-bottom: 1px solid #eef2f7; }
.short { color: var(--accent); font-weight: 600; text-decoration: none; }
.target { color: #475569; font-size: 13px; word-break: break-all; }
.clicks { color: #94a3b8; font-size: 12px; }
```

- [ ] **Step 3: Escribir el test que falla `internal/web/handler_test.go`**

```go
package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/samuelcatalanz123/urlshortener-go/internal/store"
)

func newTestHandler(t *testing.T) (http.Handler, *store.Store) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	s := store.New(rdb)
	h, err := New(s)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return h.Routes(), s
}

func postForm(path, value string) *http.Request {
	form := url.Values{"url": {value}}
	req := httptest.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestHomeRenders(t *testing.T) {
	h, _ := newTestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, esperaba 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "CortaURL") {
		t.Error("la página no contiene el título")
	}
}

func TestShortenCreatesLink(t *testing.T) {
	h, s := newTestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, postForm("/shorten", "https://example.com"))
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, esperaba 303", rec.Code)
	}
	links, _ := s.Recent(context.Background(), 10)
	if len(links) != 1 || links[0].URL != "https://example.com" {
		t.Fatalf("no se guardó el enlace correctamente: %+v", links)
	}
}

func TestShortenRejectsInvalidURL(t *testing.T) {
	h, s := newTestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, postForm("/shorten", "no-es-una-url"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("código = %d, esperaba 400", rec.Code)
	}
	links, _ := s.Recent(context.Background(), 10)
	if len(links) != 0 {
		t.Errorf("no debería guardar nada, hay %d", len(links))
	}
}

func TestRedirectIncrementsClicks(t *testing.T) {
	h, s := newTestHandler(t)
	_ = s.Save(context.Background(), "abc123", "https://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/abc123", nil))
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("código = %d, esperaba 301", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "https://example.com" {
		t.Errorf("Location = %q", loc)
	}
	link, _ := s.Get(context.Background(), "abc123")
	if link.Clicks != 1 {
		t.Errorf("Clicks = %d, esperaba 1", link.Clicks)
	}
}

func TestRedirectNotFound(t *testing.T) {
	h, _ := newTestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/noexiste", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("código = %d, esperaba 404", rec.Code)
	}
}
```

- [ ] **Step 4: Ejecutar y ver que falla**

Run: `go test ./internal/web/`
Expected: FAIL (undefined: New, Routes)

- [ ] **Step 5: Implementar `internal/web/handler.go`**

```go
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
```

- [ ] **Step 6: Ejecutar y ver que pasa**

Run: `go test ./internal/web/`
Expected: PASS (todos los tests de web)

- [ ] **Step 7: Commit**

```bash
git add -A
git -c user.name="Samuel Catalán" -c user.email="samuelcatalanz123@gmail.com" commit -m "feat: handler web, plantillas y estáticos con pruebas"
```

---

### Task 5: Punto de entrada `main.go`

**Files:**
- Create: `main.go`

- [ ] **Step 1: Implementar `main.go`**

```go
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samuelcatalanz123/urlshortener-go/internal/store"
	"github.com/samuelcatalanz123/urlshortener-go/internal/web"
)

func main() {
	redisAddr := envOr("REDIS_ADDR", "127.0.0.1:6379")
	addr := envOr("SERVER_ADDR", ":8080")

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("no se pudo conectar a Redis en %s: %v", redisAddr, err)
	}

	h, err := web.New(store.New(rdb))
	if err != nil {
		log.Fatalf("no se pudo crear el handler: %v", err)
	}

	slog.Info("servidor iniciado", "addr", addr, "redis", redisAddr)
	if err := http.ListenAndServe(addr, h.Routes()); err != nil {
		log.Fatal(err)
	}
}

// envOr devuelve la variable de entorno key, o def si está vacía.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
```

- [ ] **Step 2: Verificar todo**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: compila, vet limpio, todos los tests PASS.

- [ ] **Step 3: Commit**

```bash
git add -A
git -c user.name="Samuel Catalán" -c user.email="samuelcatalanz123@gmail.com" commit -m "feat: punto de entrada main con conexión a Redis"
```

---

### Task 6: README y CI

**Files:**
- Create: `README.md`, `.github/workflows/ci.yml`

- [ ] **Step 1: Crear `README.md`**

```markdown
# CortaURL — Acortador de URLs (Go + Redis)

[![CI](https://github.com/samuelcatalanz123/urlshortener-go/actions/workflows/ci.yml/badge.svg)](https://github.com/samuelcatalanz123/urlshortener-go/actions/workflows/ci.yml)

Aplicación web que acorta URLs y cuenta los clics. Servidor en **Go**
(`net/http`, `html/template`) que guarda los enlaces en **Redis**.

## Uso

Necesitas un servidor Redis corriendo (por defecto en `127.0.0.1:6379`):

```bash
redis-server &       # o: brew services start redis
go run .
```

Abre **http://localhost:8080**. Pega una URL larga y obtén un enlace corto.
Al abrir el enlace corto, te redirige a la original y suma un clic.

Variables de entorno: `REDIS_ADDR` (default `127.0.0.1:6379`),
`SERVER_ADDR` (default `:8080`).

## Cómo funciona

- Cada enlace se guarda en Redis como un hash `short:{código}` con la URL y
  los clics; los códigos recientes van en una lista `recent`.
- El código corto son 6 caracteres aleatorios (base62) generados con
  `crypto/rand`.
- Las plantillas y el CSS van embebidos en el binario (`go:embed`).

## Estructura

```
main.go                 arranque (conecta a Redis, sirve)
internal/store/         acceso a datos (Redis) + pruebas con miniredis
internal/web/           handlers, plantillas (templates/) y estilos (static/)
```

## Pruebas

```bash
go test ./...
```

Las pruebas usan **miniredis** (un Redis en memoria 100% Go): no necesitan
un servidor Redis real.

## Stack

Go (net/http, html/template, go:embed, crypto/rand) ·
Redis (github.com/redis/go-redis/v9) · pruebas con miniredis.
```

- [ ] **Step 2: Crear `.github/workflows/ci.yml`**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Build
        run: go build ./...

      - name: Vet
        run: go vet ./...

      - name: Test
        run: go test ./...
```

> Nota: la CI **no** instala Redis porque las pruebas usan miniredis.

- [ ] **Step 3: Commit**

```bash
git add -A
git -c user.name="Samuel Catalán" -c user.email="samuelcatalanz123@gmail.com" commit -m "docs: README y CI con GitHub Actions"
```

---

### Task 7: Prueba manual con Redis real

- [ ] **Step 1: Instalar y arrancar Redis**

```bash
brew install redis
redis-server --daemonize yes
redis-cli ping   # -> PONG
```

- [ ] **Step 2: Arrancar la app y probar**

```bash
go run . &
sleep 1
# Crear un enlace y comprobar que redirige:
curl -s -X POST -d "url=https://example.com" http://localhost:8080/shorten -i | head -1   # 303
curl -s http://localhost:8080/ | grep -o '/[A-Za-z0-9]\{6\}' | head -1                    # /XXXXXX
```

Abrir http://localhost:8080 en el navegador, pegar una URL, ver el enlace y
los clics. Parar con `kill %1` cuando termine.

---

## Self-Review (autor del plan)

- **Cobertura del spec:** rutas (Task 4), Redis store (Task 3), código corto
  (Task 2), errores URL inválida/404 (Task 4 tests), miniredis (Task 3/4),
  config por entorno (Task 5), README+CI (Task 6). ✔
- **Sin placeholders:** todo el código está escrito. ✔
- **Consistencia de tipos:** `Link{Code,URL,Clicks}`, `Store.Save/Get/Exists/
  IncrClicks/Recent`, `web.New→*Handler`, `Handler.Routes()`,
  `GenerateCode(n)` — usados igual en tests e implementación. ✔
```

