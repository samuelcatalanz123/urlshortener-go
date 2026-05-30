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
