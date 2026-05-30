# Diseño: CortaURL — Acortador de URLs (web en Go + Redis)

**Fecha:** 2026-05-30
**Estado:** Aprobado para escribir el plan de implementación
**Autor del proyecto:** Samuel (4º proyecto de portafolio)

## Objetivo

Una aplicación web sencilla que acorta URLs: pegas una dirección larga y recibe
una corta; al abrir la corta, redirige a la larga y cuenta los clics. Servida
por un programa en Go con plantillas HTML y **Redis** como almacenamiento.
Objetivo de aprendizaje del usuario: practicar **Redis** desde Go.

## Decisiones tomadas (brainstorming)

| Tema          | Decisión                                                      |
| ------------- | ------------------------------------------------------------- |
| Interfaz      | **Web app en Go**: página con formulario (`html/template`)    |
| Almacenamiento| **Redis** (clave→valor en memoria; rápido para clics)         |
| Pruebas       | **miniredis** (Redis en memoria 100% Go) — sin servidor real  |
| Repo          | Proyecto y repositorio **nuevos** (`urlshortener-go`)         |

## Pantallas y rutas

- **Inicio** — `GET /`: formulario para pegar una URL larga + lista de los
  enlaces recientes (código corto, URL original y nº de clics).
- **Acortar** — `POST /shorten`: valida la URL, genera un código corto, lo
  guarda en Redis y redirige al inicio (mostrando el nuevo enlace).
- **Redirigir** — `GET /{code}`: busca el código en Redis. Si existe, suma +1 a
  los clics y redirige (301) a la URL larga. Si no, 404.

## Arquitectura y estructura

```
urlshortener-go/
  go.mod
  main.go                     Composición: conecta a Redis, monta rutas, sirve.
  internal/store/
    store.go                  Store sobre un *redis.Client: Save, Get, IncrClicks,
                              Recent. Tipos de dominio Link{Code,URL,Clicks}.
    store_test.go             Pruebas con miniredis (Redis en memoria, sin servidor).
  internal/web/
    handler.go                Handler con el Store y plantillas; rutas y validación.
    shortcode.go              GenerateCode(n int) string — código base62 aleatorio.
    shortcode_test.go         Pruebas de longitud y alfabeto del código.
    templates/home.html       Página de inicio (formulario + lista).
    static/style.css          Estilos sencillos.
  README.md
```

- **store.go:** `type Store struct { rdb *redis.Client }`, `New(rdb) *Store`, y:
  - `Save(ctx, code, url string) error` — guarda el hash `short:{code}`
    (`url`, `clicks=0`) y añade el código a la lista `recent`.
  - `Get(ctx, code string) (Link, error)` — devuelve el Link o `ErrNotFound`.
  - `IncrClicks(ctx, code string) error` — `HINCRBY short:{code} clicks 1`.
  - `Recent(ctx, n int) ([]Link, error)` — los últimos n enlaces.
  - `Exists(ctx, code string) (bool, error)` — para evitar colisiones de código.
  - Tipo `Link{ Code, URL string; Clicks int }`.
- **shortcode.go:** `GenerateCode(n int) string` — n caracteres aleatorios del
  alfabeto base62 (`a-zA-Z0-9`), usando `crypto/rand`.
- **handler.go:** `Handler{ store, tmpl }`, `New(store) (*Handler, error)`,
  `Routes() http.Handler`. Genera el código (reintentando si ya existe),
  valida la URL, y maneja la redirección con conteo de clics.

## Datos en Redis

- `short:{code}` → HASH con campos `url` (string) y `clicks` (entero).
- `recent` → LIST de códigos (LPUSH al crear; LRANGE 0 n-1 para mostrarlos).

## Generación del código corto

Código aleatorio de **6 caracteres** base62 (`GenerateCode(6)`). Antes de
guardar, se comprueba con `Exists`; si ya existe, se genera otro (bucle hasta
encontrar uno libre).

## Configuración (variables de entorno, con defaults)

- `REDIS_ADDR` (default `127.0.0.1:6379`) — dirección de Redis.
- `PORT` / `SERVER_ADDR` (default `:8080`) — dirección de escucha.

## Manejo de errores

- URL vacía o que no empieza por `http://` o `https://` → vuelve al inicio con
  un mensaje claro; no guarda.
- Código inexistente en `GET /{code}` → página 404.
- Error de Redis → página 500 sencilla, sin filtrar detalles.

## Pruebas

- **shortcode_test.go:** `GenerateCode(6)` devuelve 6 caracteres, todos del
  alfabeto base62; dos llamadas dan resultados distintos.
- **store_test.go:** con **miniredis** (`github.com/alicebob/miniredis/v2`):
  `Save` + `Get` devuelve el Link; `Get` de un código inexistente → `ErrNotFound`;
  `IncrClicks` sube el contador; `Recent` devuelve los enlaces creados.
- `go build ./...`, `go vet ./...`, `go test ./...` limpios.

## Dependencias

- `github.com/redis/go-redis/v9` — cliente de Redis para Go.
- `github.com/alicebob/miniredis/v2` — Redis en memoria para las pruebas.

## Ejecución (Redis real)

Para *ejecutar* la app hace falta un servidor Redis (`redis-server`). El
asistente lo instala (`brew install redis`) y lo arranca. Las **pruebas no lo
necesitan** (usan miniredis).

## Fuera de alcance (YAGNI)

Cuentas de usuario, expiración de enlaces, códigos personalizados, estadísticas
avanzadas, y el despliegue en vivo (eso después).

## Criterios de éxito

1. Con Redis corriendo, `go run .` sirve en `http://localhost:8080`.
2. Pegar una URL larga devuelve una corta; abrir la corta redirige y suma un clic.
3. La lista de recientes muestra los enlaces con su contador.
4. URL inválida y código inexistente dan mensajes/errores claros.
5. Las pruebas pasan sin servidor Redis (miniredis).
6. El proyecto queda en su propio repositorio, listo para GitHub.
