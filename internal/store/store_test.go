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
