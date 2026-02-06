# Lootstash Catalog API - Project Guide

## Overview

Go service that serves static Diablo II game data. Read-only catalog backed by PostgreSQL (via pgx/v5) and Redis for caching.

## Build & Run

```bash
go build ./...
go run . serve          # HTTP server on :8080
go run . import         # Import game data from catalogs/
go run . seed           # Seed initial data
```

Uses Cobra CLI for command management.

## Architecture

```
Handlers (HTTP/Fiber)
    ↓
Services (business logic, caching)
    ↓
Repositories (data access, SQL queries)
    ↓
Models (domain entities)
    ↓
PostgreSQL (pgx/v5 direct) + Redis cache
```

Middleware stack: Recovery → Logger → RequestID → CORS → Handlers

## Key Paths

| Path | Purpose |
|------|---------|
| `main.go` | Entry point, calls `cmd.Execute()` |
| `cmd/serve.go` | HTTP server startup |
| `cmd/import.go` | Import game data |
| `cmd/seed.go` | Seed initial data |
| `internal/api/server.go` | Route registration |
| `internal/api/handlers/items.go` | All HTTP handlers |
| `internal/api/dto/items.go` | Response DTOs, transformation logic |
| `internal/games/d2/entities.go` | Domain models (UniqueItem, SetItem, Runeword, Rune, Gem, etc.) |
| `internal/games/d2/repository.go` | Database queries |
| `internal/games/d2/query.go` | Query builder helpers |
| `internal/games/d2/translator.go` | Property code → human-readable text translation (100+ codes) |
| `internal/games/d2/importer.go` | Data import logic |
| `internal/storage/supabase.go` | S3-compatible icon storage |
| `catalogs/` | 712 D2 data files (TSV format) |

## API Endpoints

```
GET /api/v1/d2/items/search         # Search items by name
GET /api/v1/d2/items/:type/:id      # Generic item lookup
GET /api/v1/d2/items/unique/:id     # Unique item detail
GET /api/v1/d2/items/set/:id        # Set item detail
GET /api/v1/d2/items/runeword/:id   # Runeword detail
GET /api/v1/d2/items/runeword/:id/bases  # Valid bases for runeword
GET /api/v1/d2/items/rune/:id       # Rune detail
GET /api/v1/d2/items/gem/:id        # Gem detail
GET /api/v1/d2/items/base/:id       # Base item detail
GET /api/v1/d2/{runes,gems,bases,uniques,sets,runewords}  # List all of type
```

## Property Translation

`translator.go` maps stat codes to display text with placeholder replacement:
- `"dmg-fire"` → `"Adds {min}-{max} Fire Damage"`
- `"allskills"` → `"+{value} To All Skills"`
- `"ac%"` → `"+{value}% Enhanced Defense"`

Items have a `Properties` array with `Code`, `Min`, `Max`, `Param` fields. The translator uses these to produce `DisplayText`.

## Dependencies (Go 1.21)

- gofiber/fiber/v2 - HTTP framework
- jackc/pgx/v5 - PostgreSQL driver
- redis/go-redis/v9 - Redis client
- spf13/cobra - CLI framework
- PuerkitoBio/goquery - HTML parsing
- aws/aws-sdk-go - S3-compatible storage

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_URL` | Redis connection string |
| `SUPABASE_URL` | Supabase project URL (for storage) |
| `SUPABASE_SERVICE_KEY` | Supabase service role key |

## Docker

```bash
# Build
docker build -t lootstash-catalog-api .

# Run with local Redis
docker-compose up -d redis
docker run -p 8080:8080 --env-file .env lootstash-catalog-api
```

## Fly.io Deployment

```bash
fly launch --no-deploy  # First time setup
fly deploy              # Deploy
fly secrets set DATABASE_URL=xxx REDIS_URL=xxx  # Set secrets
```
