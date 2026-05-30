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
	addr := serverAddr()

	rdb, err := newRedisClient()
	if err != nil {
		log.Fatalf("configuración de Redis inválida: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("no se pudo conectar a Redis: %v", err)
	}

	h, err := web.New(store.New(rdb))
	if err != nil {
		log.Fatalf("no se pudo crear el handler: %v", err)
	}

	slog.Info("servidor iniciado", "addr", addr)
	if err := http.ListenAndServe(addr, h.Routes()); err != nil {
		log.Fatal(err)
	}
}

// newRedisClient crea el cliente de Redis. En la nube se usa REDIS_URL (una
// dirección completa tipo redis://usuario:clave@host:puerto); en local basta
// REDIS_ADDR (host:puerto), con valor por defecto 127.0.0.1:6379.
func newRedisClient() (*redis.Client, error) {
	if url := os.Getenv("REDIS_URL"); url != "" {
		opt, err := redis.ParseURL(url)
		if err != nil {
			return nil, err
		}
		return redis.NewClient(opt), nil
	}
	return redis.NewClient(&redis.Options{Addr: envOr("REDIS_ADDR", "127.0.0.1:6379")}), nil
}

// serverAddr decide en qué dirección escuchar. Muchos servicios en la nube
// indican el puerto en la variable PORT; en local se usa SERVER_ADDR (:8080).
func serverAddr() string {
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return envOr("SERVER_ADDR", ":8080")
}

// envOr devuelve la variable de entorno key, o def si está vacía.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
