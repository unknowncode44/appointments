# Appointments API (Backend)

## Overview
Go REST API for appointment scheduling. Uses Fiber v2 as HTTP framework with PostgreSQL, JWT auth, and automatic DB migrations on startup.

## Tech Stack
- **Go 1.23** — language
- **Fiber v2** — HTTP framework (fasthttp-based, NOT net/http)
- **PostgreSQL 15** — database via `pgx/v5` driver
- **sqlc** — generates Go code from SQL queries (never edit `internal/db/sqlc/` by hand)
- **golang-migrate** — DB migrations, run automatically on app startup
- **JWT (golang-jwt/jwt v5)** — auth tokens (access: 15m, refresh: 24h)
- **Viper** — config loaded from `app.env` file in project root
- **zerolog** — structured logging (console writer when `APP_DEBUG=true`)
- **Air** — hot reload in development (`.air.toml`)
- **testify** — test assertions

## Project Structure
```
main.go                    — entry point: loads config, connects DB, runs migrations, starts server
internal/
  api/                     — server struct, middleware setup
  db/
    migration/             — SQL migration files (sequential, immutable)
    query/                 — SQL query files (input for sqlc)
    sqlc/                  — generated Go DB code (DO NOT EDIT)
  platform/                — platform-level config (logger, etc.)
  repositories/            — data access layer (uses sqlc-generated code)
  routes/                  — route registration
  services/                — business logic
  token/                   — JWT token creation and validation
  util/                    — config struct, helpers
tests/                     — integration/unit tests
```

## Commands
```bash
# Development (hot reload)
air

# Run directly
go run main.go

# Build
go build -o main .

# Run tests
go test ./...

# Run specific test
go test ./tests/...

# Regenerate DB code from SQL queries (after changing internal/db/query/*.sql)
sqlc generate

# Docker (full stack: app + postgres + redis)
docker-compose up --build

# Docker (detached)
docker-compose up -d
```

## Environment Variables
Copy `app.env.example` to `app.env`:
```
APP_PORT=8080
APP_NAME=appointments
APP_DEBUG=false            # true enables zerolog console writer
DB_CONNECTION=postgres
DB_HOST=postgres           # use 'localhost' outside Docker
DB_PORT=5432
DB_DATABASE=appointments
DB_USERNAME=appointments
DB_PASSWORD=change_me_strong_password
TOKEN_SECRET_KEY=change_me_to_a_32_char_random_string
ACCESS_TOKEN_DURATION=15m
REFRESH_TOKEN_DURATION=24h
MIGRATION_URL=file://migrations
EVO_API_URL=https://evo-api-server.hvdevs.com
EVO_API_KEY=<your-master-key>
```

## Database & Migrations
- Migrations run **automatically** when the app starts via `runDBMigration()` in `main.go`
- Migration files live in `internal/db/migration/` and are **immutable** — never modify existing ones
- To add a change, create a new numbered migration file
- SQL queries go in `internal/db/query/`; run `sqlc generate` to regenerate `internal/db/sqlc/`
- `sqlc.yaml` configures: UUID → `github.com/google/uuid`, timestamptz → `time.Time`

## Code Style
- Standard Go conventions (gofmt, go vet)
- Package names match directory names
- Interfaces defined in `internal/db/sqlc/` (emit_interface: true in sqlc.yaml)
- Use `zerolog` for all logging — no `fmt.Println` in production paths
- Config is accessed via `util.Config` struct throughout; never read env vars directly

## Docker
- `docker-compose.yml` — development stack (postgres:15-alpine, redis:7-alpine)
- `docker-compose.prod.yml` — production variant
- App runs on port `8080`; postgres exposed on `5432`
- DB credentials in docker-compose: user=`root`, password=`secret`, db=`ticket`

## API Reference
See `@API_REFERENCE.md` for full endpoint documentation.

## Gotchas
- `internal/db/sqlc/` is entirely generated — running `sqlc generate` overwrites it
- The module path is `github.com/mousav1/ticket` (historical name, repo is `appointments`)
- Config file must be named `app.env` and placed in the directory where the binary runs
- `DB_HOST` should be `localhost` for local dev without Docker, `postgres` inside Docker network
- Redis is included in docker-compose but integration may be partial — check before adding cache logic
