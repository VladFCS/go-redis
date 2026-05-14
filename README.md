# Go + Redis HTTP Example

This project is a learning-focused Go + Redis backend example built with:

- `net/http`
- `chi`
- Redis
- PostgreSQL for room storage
- `sqlc` for typed SQL queries

The goal of the project is to practice Redis-backed backend design in Go through one realistic domain entity. We use:

- `room` as stable business data in PostgreSQL
- `reservation` as TTL-sensitive operational state in Redis

## Current stage

The repository is currently bootstrapped with:

- `cmd/main.go` for HTTP server startup
- `chi` router and middleware
- graceful shutdown
- environment-based config
- `docker-compose.yml` for local Redis
- `/healthz` endpoint for a quick smoke test

At this stage:

- Redis is used for reservation state
- PostgreSQL schema and `sqlc` scaffolding are prepared for room storage

## Project structure

```text
.
├── cmd
│   └── main.go
├── db
│   ├── migrations
│   └── query
├── docker-compose.yml
├── go.mod
└── README.md
```

As the project grows, a likely next structure is:

```text
.
├── cmd
│   └── main.go
├── internal
│   └── reservation
│       ├── handler.go
│       ├── keys.go
│       ├── model.go
│       ├── repository.go
│       ├── service.go
│       └── service_test.go
├── docker-compose.yml
├── go.mod
└── README.md
```

## Run Redis locally

```bash
docker compose up -d
```

Redis will be available at `localhost:6379`.
PostgreSQL will be available at `localhost:5432`.

Default PostgreSQL credentials:

- database: `go_redis`
- user: `postgres`
- password: `postgres`

## Run the first migration

Apply [000001_create_rooms.up.sql](/Users/vlad/MySpace/Code/golang/go-redis/db/migrations/000001_create_rooms.up.sql:1) against your local PostgreSQL instance with your preferred migration tool or `psql`.

Example:

```bash
psql postgresql://postgres:postgres@localhost:5432/go_redis -f db/migrations/000001_create_rooms.up.sql
```

## Generate room queries with sqlc

```bash
sqlc generate
```

This reads:

- schema from `db/migrations/`
- queries from `db/query/`

and generates typed Go code into `internal/room/postgres/`.

## Run the API

First, fetch dependencies:

```bash
go mod tidy
```

Then start the server:

```bash
go run ./cmd
```

The server starts on `http://localhost:8080`.

## Endpoints

### GET /

Simple project info endpoint:

```bash
curl http://localhost:8080/
```

### GET /healthz

Health check endpoint:

```bash
curl http://localhost:8080/healthz
```

Expected response:

```json
{"status":"ok"}
```

## Configuration

Optional environment variables:

- `HTTP_ADDR` default: `:8080`
- `REDIS_ADDR` default: `localhost:6379`
- `REDIS_PASSWORD` default: empty
- `REDIS_DB` default: `0`

Example:

```bash
HTTP_ADDR=:8080 REDIS_ADDR=localhost:6379 REDIS_DB=0 go run ./cmd
```

## Why `reservation` for Redis

Compared with a simple CRUD-only entity, `reservation` better demonstrates why Redis is useful in backend systems:

- expiration with TTL
- temporary holds
- atomic counters
- race-condition awareness
- idempotent request handling
- sorted sets and streams later on

That makes it a stronger teaching project than using Redis as just another document store.
