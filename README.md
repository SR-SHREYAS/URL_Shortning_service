# Boilerplate

Monorepo boilerplate for a Go backend built with Gin, plus TypeScript workspace tooling at the repo root. The goal of this repository is to provide a production-minded starting point that keeps architecture, observability, validation, and task automation consistent while letting future projects swap business logic without rebuilding the foundation.

## What lives here

- `apps/backend`: the Go API service.
- `packages/*`: shared workspace packages for frontend or utility code.
- `package.json`: root workspace control plane for Bun and Turbo.
- `apps/backend/Taskfile.yml`: backend task automation.
- `apps/backend/.env.sample`: canonical environment sample for the backend.

## Stack

- Go 1.25.
- Gin for HTTP routing and middleware.
- PostgreSQL with `pgx/v5` and Tern migrations.
- Redis for caching, jobs, and background processing.
- Clerk for authentication.
- New Relic for APM, distributed tracing, and custom events.
- Zerolog for structured logs.
- Asynq for background jobs.
- Resend for transactional email.
- Bun and Turbo for monorepo scripts at the root.

## Design Goals

- Keep the application layer thin and framework-specific concerns isolated.
- Centralize configuration, observability, and error handling.
- Make the boilerplate easy to clone for another product without changing the architecture.
- Keep the code production-oriented without forcing every integration to be mandatory on day one.

## Repository Layout

```text
.
├── package.json               # Root monorepo scripts and workspace definition
├── turbo.json                 # Turbo task orchestration
├── apps/
│   └── backend/               # Gin backend service
│       ├── cmd/go-boilerplate # Server entrypoint
│       ├── internal/          # Private application code
│       ├── static/            # OpenAPI HTML and JSON
│       ├── templates/         # Email templates
│       ├── Taskfile.yml       # Backend tasks
│       └── .env.sample        # Environment template
└── packages/                  # Shared workspace packages
```

## Backend Architecture

The backend follows a clean layered shape:

- `cmd`: process startup, graceful shutdown, dependency wiring.
- `internal/config`: environment loading and validation.
- `internal/database`: PostgreSQL pool setup, migrations, and shutdown.
- `internal/server`: application container and lifecycle management.
- `internal/router`: route registration and middleware ordering.
- `internal/middleware`: request ID, auth, tracing, logging, CORS, recovery, rate limiting.
- `internal/handler`: HTTP handlers and the generic handler pipeline.
- `internal/service`: business logic and integration services.
- `internal/repository`: persistence layer placeholders and future repositories.
- `internal/lib`: shared subsystems such as email and background jobs.
- `internal/sqlerr`: Postgres error normalization.
- `internal/validation`: bind-and-validate helpers and request validation.

The app is intentionally organized so each layer has a single job:

- handlers deal with HTTP shape and request/response conversion.
- services coordinate business behavior.
- repositories handle persistence.
- middleware handles request-wide cross-cutting concerns.
- config and observability are initialized once and passed down.

## Request Flow

1. `main.go` loads config from environment variables.
2. Logger, New Relic, database, Redis, job services, repositories, services, and handlers are initialized.
3. Gin router is created and middleware is attached in a fixed order.
4. Requests receive a request ID, structured logging, tracing, auth context, validation, and standardized error formatting.
5. Responses are emitted through the shared handler helpers so JSON, file, and no-content responses behave consistently.

## Gin Interface

The boilerplate is Gin-native across the request lifecycle:

- `*gin.Context` is used throughout handlers and middleware.
- middleware returns `gin.HandlerFunc`.
- response writing uses Gin primitives like `c.JSON`, `c.Data`, `c.AbortWithStatusJSON`, and `c.Writer`.
- tracing uses `nrgin` rather than Echo instrumentation.
- request binding and context access use Gin-compatible helpers.

There is no Echo framework dependency in the backend Go code.

## Configuration

Configuration is loaded from environment variables prefixed with `BOILERPLATE_`.

The sample file is [apps/backend/.env.sample](apps/backend/.env.sample).

### Core Keys

- `BOILERPLATE_PRIMARY.ENV`
- `BOILERPLATE_SERVER.PORT`
- `BOILERPLATE_SERVER.READ_TIMEOUT`
- `BOILERPLATE_SERVER.WRITE_TIMEOUT`
- `BOILERPLATE_SERVER.IDLE_TIMEOUT`
- `BOILERPLATE_SERVER.CORS_ALLOWED_ORIGINS`
- `BOILERPLATE_DATABASE.HOST`
- `BOILERPLATE_DATABASE.PORT`
- `BOILERPLATE_DATABASE.USER`
- `BOILERPLATE_DATABASE.PASSWORD`
- `BOILERPLATE_DATABASE.NAME`
- `BOILERPLATE_DATABASE.SSL_MODE`
- `BOILERPLATE_DATABASE.MAX_OPEN_CONNS`
- `BOILERPLATE_DATABASE.MAX_IDLE_CONNS`
- `BOILERPLATE_DATABASE.CONN_MAX_LIFETIME`
- `BOILERPLATE_DATABASE.CONN_MAX_IDLE_TIME`
- `BOILERPLATE_AUTH.SECRET_KEY`
- `BOILERPLATE_INTEGRATION.RESEND_API_KEY`
- `BOILERPLATE_REDIS.ADDRESS`

### Observability Keys

- `BOILERPLATE_OBSERVABILITY.SERVICE_NAME`
- `BOILERPLATE_OBSERVABILITY.ENVIRONMENT`
- `BOILERPLATE_OBSERVABILITY.LOGGING.LEVEL`
- `BOILERPLATE_OBSERVABILITY.LOGGING.FORMAT`
- `BOILERPLATE_OBSERVABILITY.LOGGING.SLOW_QUERY_THRESHOLD`
- `BOILERPLATE_OBSERVABILITY.NEW_RELIC.LICENSE_KEY`
- `BOILERPLATE_OBSERVABILITY.NEW_RELIC.APP_LOG_FORWARDING_ENABLED`
- `BOILERPLATE_OBSERVABILITY.NEW_RELIC.DISTRIBUTED_TRACING_ENABLED`
- `BOILERPLATE_OBSERVABILITY.NEW_RELIC.DEBUG_LOGGING`
- `BOILERPLATE_OBSERVABILITY.HEALTH_CHECKS.ENABLED`
- `BOILERPLATE_OBSERVABILITY.HEALTH_CHECKS.INTERVAL`
- `BOILERPLATE_OBSERVABILITY.HEALTH_CHECKS.TIMEOUT`
- `BOILERPLATE_OBSERVABILITY.HEALTH_CHECKS.CHECKS`

## Startup Behavior

The backend entrypoint is [apps/backend/cmd/go-boilerplate/main.go](apps/backend/cmd/go-boilerplate/main.go).

Startup sequence:

1. Load and validate config.
2. Initialize the New Relic logger service.
3. Run database migrations outside local development.
4. Create the server container.
5. Build repositories, services, and handlers.
6. Register router and middleware.
7. Start HTTP server.
8. Wait for SIGINT and shut down gracefully.

## Middleware Stack

Middleware is layered to keep behavior predictable:

- CORS.
- recovery.
- request ID injection.
- New Relic request instrumentation.
- tracing enrichment.
- context enrichment with logger and user metadata.
- request logging.
- global error handling.
- rate limiting.

This ordering matters because logging, tracing, and context enrichment depend on earlier middleware having already populated request metadata.

## Taskfile

The backend task file is [apps/backend/Taskfile.yml](apps/backend/Taskfile.yml).

Available tasks:

- `task help`: list all tasks.
- `task run`: start the backend.
- `task test`: run `go test ./...`.
- `task migrations:new name=<name>`: create a new migration file.
- `task migrations:up`: apply all migrations.
- `task migrations:down`: roll back the last migration.
- `task tidy`: format, tidy, and verify Go modules.

Migration tasks use `tern` and the `BOILERPLATE_DB_DSN` environment variable.

## Dependencies

Root workspace tooling:

- Bun for package management.
- Turbo for task orchestration.
- TypeScript for shared workspace packages.

Backend runtime dependencies:

- Gin.
- Clerk.
- New Relic.
- pgx / pgx-zerolog.
- Redis client and Asynq.
- Resend.
- Zerolog.
- Validator.

## Deployment Notes

The boilerplate is designed so a project can be moved toward production without redesigning the foundation:

- set all required environment variables.
- configure CORS origins correctly.
- provide PostgreSQL and Redis connectivity.
- configure Clerk secret key and New Relic license key.
- enable migrations before first start.
- run the backend through the taskfile or a container entrypoint, not by ad hoc commands.

## Testing

Run backend validation with:

```bash
cd apps/backend
go test ./...
```

## Notes for Future Projects

This boilerplate is intentionally opinionated, but the application layer should be replaceable without changing the platform layer. For a new project, keep the architecture and initialization flow, then swap the domain services, handlers, repositories, and migrations as needed.
