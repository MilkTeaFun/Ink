# Ink API

The Go service in `server/` backs authentication, workspace persistence, printers, plugins, schedules, and feedback flows for Ink.

## Local development

Recommended setup from the repository root:

```bash
make dev-db
make migrate-up
make seed-dev
make dev-api
```

Fast path:

```bash
make dev-api
```

That command creates `server/.env` when missing, ensures PostgreSQL is ready, applies migrations, seeds the development admin account, and starts the API server.

The generated admin credentials are written to `server/.dev-admin-password`.

## Commands

- `make dev-api`: bootstrap and start the API
- `make migrate-up`: apply SQL migrations from `server/migrations/`
- `make seed-dev`: create the development admin account if it does not already exist
- `make check-api`: `gofmt`, Go tests, and backend build
- `make smoke-api`: real API smoke flow for login and workspace persistence
- `make reset-db`: remove the local PostgreSQL volume

## Endpoint groups

- auth: login, refresh, current user, logout, password change
- workspace: read and write persisted workspace state
- admin: member creation and shared AI configuration
- printers and print jobs: device bindings, queue actions, and job lifecycle
- plugins and schedules: installation, binding, validation, and scheduling
- feedback: printable feedback submission for the admin workflow
