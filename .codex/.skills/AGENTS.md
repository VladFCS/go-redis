# Repository Guidelines

## Project Structure & Module Organization
`cmd/main.go` is the application entrypoint: it loads config, connects Redis, and starts the HTTP server. Domain code lives under `internal/reservation/`:

- `handler.go` for HTTP handlers and JSON responses
- `service.go` for validation and business rules
- `repository.go` and `keys.go` for Redis persistence and key naming
- `model.go` for request/response and entity types

Keep new features inside `internal/<domain>/` and wire them from `cmd/main.go`. Local Redis infrastructure is defined in `docker-compose.yml`.

## Build, Test, and Development Commands
- `docker compose up -d`: start Redis on `localhost:6379`.
- `go run ./cmd`: run the API locally on `:8080`.
- `go test ./...`: run all Go tests; currently passes with no test files.
- `go test ./... -cover`: check coverage when tests are added.
- `go mod tidy`: clean up module dependencies after adding or removing imports.

Use `curl http://localhost:8080/healthz` for a quick smoke test.

## Coding Style & Naming Conventions
Follow standard Go formatting: run `gofmt` on edited files before submitting changes. Keep package names short and lowercase (`reservation`), exported identifiers in `CamelCase`, and unexported helpers in `camelCase`. Preserve the current layering: handler -> service -> repository. Prefer explicit request/response structs such as `CreateReservationRequest`, and keep Redis key helpers centralized in `keys.go`.

## Testing Guidelines
Place tests next to the code they cover, using `*_test.go` files, for example `internal/reservation/service_test.go`. Favor table-driven tests for validation, TTL behavior, and repository edge cases. Add coverage for new handlers and service rules with every feature or bug fix; do not leave business logic untested.

## Commit & Pull Request Guidelines
Recent history uses Conventional Commit style prefixes such as `feat:`. Continue with concise messages like `fix: validate reservation quantity` or `test: add service validation cases`.

PRs should include:
- a short description of the change and why it is needed
- linked issue or task reference when applicable
- test evidence (`go test ./...`, manual `curl`, or both)
- sample request/response output for API behavior changes

## Configuration & Security Tips
Configuration is environment-driven: `HTTP_ADDR`, `REDIS_ADDR`, `REDIS_PASSWORD`, and `REDIS_DB`. Do not commit secrets or hardcode credentials; use local environment variables for development.
