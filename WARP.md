# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Commands

- Backend
  - Build: `go build -o bin/law-oa-server ./cmd/server`
  - Run (dev): `go run cmd/server/main.go`
  - Lint/format: `gofumpt -w -s . && go vet ./... && golangci-lint run`
  - Unit tests (all): `go test -v -race ./...`
  - Run a single test: `go test -v ./internal/handlers -run '^TestAuthHandler_Login$'`
  - Coverage: `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`
  - Benchmarks: `go test -bench=. -benchmem ./...`
  - Fuzz: `go test -fuzz=Fuzz -fuzztime=30s ./internal/...`
  - Integration tests: `go test -v -race -tags=integration ./tests/...`
  - DB migrations (MySQL/PostgreSQL via env): `go run cmd/migrate/main.go -command up -migrations ./migrations`
  - Create migration files: `go run cmd/migrate/main.go -command create <name> -migrations ./migrations`

- Frontend (`frontend/`)
  - Dev server: `npm run dev`
  - Build: `npm run build`
  - Lint: `npm run lint` (auto-fix: `npm run lint:fix`)
  - Type check: `npm run type-check`
  - Tests (Jest): `npm run test` | watch: `npm run test:watch` | coverage: `npm run test:coverage`
  - Run a single test: `npm run test -- src/pages/__tests__/LoginPage.integration.test.tsx -t "logs in"`

- Orchestrated flows (recommended)
  - Start full dev stack (Docker): `bash scripts/start-dev.sh start`
  - Unified tests (select type, coverage, fuzz, etc.): `bash scripts/run_tests.sh -t unit --coverage` (see `-h` for options)
  - Basic API smoke: `bash scripts/test_api.sh`
  - Integration smoke (URLs, DB, Redis, ES): `bash scripts/test-integration.sh test`

## Architecture (big picture)

- Go monolith with layered modules under `internal/`:
  - API layer: `internal/handlers` (HTTP), `internal/middleware` (auth, rate limit, metrics), `internal/router` (routes), `internal/api/response.go` (uniform responses)
  - Services: `internal/services` encapsulate business logic; coordinate repositories and transactions
  - Repositories: `internal/repositories` provide data access (GORM, SQL builders, Redis, Elasticsearch adapters); interfaces in `interfaces.go`
  - Models: `internal/models` (domain entities, JSON types, RBAC, analytics, legal statutes)
  - Cross-cutting: `internal/common` (errors/response), `internal/security` (JWT, audit, rate limiter), `internal/monitoring` (Prometheus), `internal/logging`, `internal/tracing`, `internal/concurrency`
  - Runtime: `internal/server` (HTTP server/bootstrap), `internal/config` (env-config), `internal/database` (pooling, migrator, query builder)
- Entrypoints
  - HTTP server: `cmd/server/main.go`
  - DB migrations CLI: `cmd/migrate/main.go` (env-driven) and `migrate/main.go` (DSN flags)
- Data/infra
  - SQL migrations: `migrations/*.up.sql` / `*.down.sql`
  - Config samples: `configs/environments/*` and `.env.*`
  - Docker/K8s/Helm: `docker-compose.yml`, `deployments/`, `helm/law-oa-go/`, `k8s/`
- Frontend (React + TS, Vite)
  - App in `frontend/` with Jest tests, ESLint/Prettier; dev server on 3003; communicates with backend at 8080
- Additional service
  - `services/document-service/`: standalone Go microservice (rich document editing, JWT, search); independent build/run

## Important conventions and docs

- Spec-driven changes (OpenSpec)
  - For proposals, architecture shifts, or capability additions, use OpenSpec under `openspec/`
  - See `openspec/AGENTS.md` (quick checklist, change workflow, validation commands)
  - Typical commands: `openspec list`, `openspec show <item>`, `openspec validate --strict`
  - Bug fixes and non-breaking maintenance usually do not require proposals

- CI expectations (reference)
  - Go: format with gofumpt; vet; `golangci-lint run`; tests with race/coverage
  - Frontend: `npm run lint`, `npm run type-check`, Jest coverage
  - Security: `gosec` and `govulncheck` used in workflows

## Ports, health, and local endpoints

- Backend: http://localhost:8080 (health: `/health`, metrics: `/metrics`)
- Frontend: http://localhost:3003
- Swagger (if enabled): `http://localhost:8080/swagger/index.html`

## Test directories (orientation)

- Go unit/integration: `internal/**`, `tests/go/**`, `tests/integration/**`, `tests/performance/**`
- JS/TS e2e/integration: `tests/e2e/**`, `tests/javascript/**`, `frontend/tests/**`

## Notes

- Go 1.23+ and Node 18+ assumed.
- PostgreSQL and MySQL are both supported; driver selection via env (`DB_DRIVER`).
