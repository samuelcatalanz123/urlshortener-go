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
