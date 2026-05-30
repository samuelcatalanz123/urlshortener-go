# CortaURL — Acortador de URLs (Go + Redis)

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
