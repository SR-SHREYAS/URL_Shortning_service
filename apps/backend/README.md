# Backend

This directory contains the Gin backend service for the URL Service monorepo.

The canonical project documentation lives in [../../README.md](../../README.md).

## Quick Facts

- Go 1.25 backend.
- Gin router and middleware stack.
- PostgreSQL, Redis, Clerk, New Relic, Resend, Asynq.
- Clean architecture with handler, service, repository, and middleware layers.

## Useful Commands

```bash
go test ./...
task run
task test
task migrations:up
task migrations:down
task tidy
```

## Environment

Use [./.env.sample](./.env.sample) as the source of truth for backend configuration.

## Architecture

- `cmd/go-boilerplate`: application entrypoint.
- `internal/config`: environment loading and validation.
- `internal/server`: application lifecycle and dependency container.
- `internal/router`: Gin routing and middleware wiring.
- `internal/handler`: HTTP handlers and response helpers.
- `internal/middleware`: request ID, auth, logging, tracing, recovery, security, and rate limiting.
- `internal/service`: domain and integration services.
- `internal/repository`: persistence abstractions.
- `internal/database`: database pool and migrations.
- `internal/lib`: email and background job subsystems.

## Notes

The backend is intentionally framework-specific to Gin, but the architecture is meant to stay reusable for future projects.

## URL Shortener Features

The service implements Clerk/API-key backed link ownership, link CRUD and export,
password-protected redirects, QR PNG generation, asynchronous click enrichment,
and dashboard/link analytics. Configure the `BOILERPLATE_SERVICE.*` and
`BOILERPLATE_RATE_LIMIT.*` variables in `.env.sample` before running locally.

Public redirects use `GET /:shortCode`; all management and analytics endpoints
live below `/api/v1` and accept either a verified Clerk session or `X-API-Key`.
