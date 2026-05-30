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
