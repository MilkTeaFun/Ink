# Ink Auth Service

This service provides the first Go backend for Ink account authentication.

## Endpoints

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/change-password`
- `POST /api/v1/admin/users`
- `GET /api/v1/workspace`
- `PUT /api/v1/workspace`
- `GET /healthz`

## Local development

### Recommended flow

1. `make dev-db`
2. `make migrate-up`
3. `make seed-dev`
4. `make dev-api`

### One-command bootstrap

If you just want the fastest path, run `make dev-api`.

It will:

- auto-create `server/.env` on first launch
- ensure PostgreSQL is running
- apply unapplied migrations
- seed the development admin account
- start the API server

### Command reference

- `make dev-db`: start PostgreSQL in Docker
- `make migrate-up`: apply SQL migrations from [`server/migrations/`](./migrations)
- `make seed-dev`: create the development admin account if it does not exist
- `make bootstrap-api`: run env setup, migrations, and seed without starting the server
- `make dev-api`: bootstrap and then start the API server
- `make reset-db`: remove the local PostgreSQL volume for a clean rebuild

This is intentionally split so development, CI, and future production deployment can run migrations and seeds explicitly instead of relying on container startup side effects.

The development seed creates this account on first bootstrap:

- login: `admin`
- password: randomly generated once, printed to the terminal, and saved to `server/.dev-admin-password`

After signing in as the development admin, you can create additional member accounts from Ink settings or the `POST /api/v1/admin/users` endpoint. Each account gets its own isolated workspace snapshot on first load.

If you keep the existing PostgreSQL volume, rerunning `make dev-api` will not reset the admin password. Run `make reset-db` to rebuild the database and generate a new initial password.
