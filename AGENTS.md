<claude-mem-context>
# Memory Context

# [law-oa-go] recent context, 2026-06-21 1:03am GMT+8

No previous sessions found.
</claude-mem-context>

# Law OA Go - Codex Working Notes

## Current Project Snapshot
- Stack: Go backend + React/Vite frontend, Gin + GORM, Redis cache, Elasticsearch search, Docker Compose deployment.
- Active Go version is declared in `go.mod` as `go 1.25.0`. Do not rely on older docs that still say Go 1.19 or 1.23 without checking code.
- Main runtime entry is root `main.go`. `cmd/server/main.go` is an alternate server entry and is not the default `Makefile build` target.
- Default local build command is `go build ./...`. Default Makefile build command builds root `.` into `bin/law-oa-go`.
- Frontend package lives in `frontend/`; useful scripts are `npm run dev`, `npm run build`, `npm run type-check`, and `npm run lint`.

## Documentation State
- Root `README.md` is the best high-level status page after the 2026-04-29 cleanup.
- `docs/feature-status-actual.md`, `docs/architecture-overview.md`, `docs/CONFIGURATION.md`, `docs/DEVELOPER_GUIDE.md`, and `docs/API.md` were corrected on 2026-04-29 for major stale facts.
- The docs tree still contains many older report and planning documents from 2025. Treat them as historical unless their contents match current code.

## Important Implementation Facts
- Docker Compose defaults to MySQL 8.0 for the bundled database service; PostgreSQL remains supported by config and migrations.
- `config.Load()` uses `godotenv` + `viper`, reads `.env`, `config.yaml`, `config/config.yaml`, and `/etc/law-oa/config.yaml`, and binds core env vars such as `DB_DRIVER`, `DB_HOST`, `DB_USERNAME`, `DB_PASSWORD`, `DB_DATABASE`, `REDIS_HOST`, `ES_HOST`, and `JWT_SECRET`.
- Root `main.go` registers `/metrics`, `/health*`, `/api/v1/monitor/*`, and `/swagger/*any` in non-production mode, then calls `internal/router.Init`.
- `internal/router/router.go` registers the main `/api/v1` surface: auth, users, clients, case intakes, cases, enhanced cases, legal statutes, notifications/templates, content filter, dashboard, conflict, teams, documents, OnlyOffice, approvals/delegations, waivers, integration, inbox, deadlines, folder templates, ethical wall, finance, trust, and conflict-v2 entity checks.
- Migrations currently run through `000029_approval_e2e_postgresql_compat.*.sql`, plus `001_schema_v2.2.0.sql`.
