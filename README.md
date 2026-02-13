# Deals & Promotions MVP (Go + Fiber + PostgreSQL)

## Stack
- Go 1.22
- Fiber
- Server-side rendering with templ components (Go-based templ API) + htmx partial updates
- PostgreSQL
- golang-migrate migrations
- Session cookie auth
- Local uploads in `static/uploads`
- i18n from JSON files (`i18n/en.json`, `i18n/si.json`, `i18n/ta.json`)

## Quick Start
1. Start Postgres:
   ```bash
   docker compose up -d
   ```
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Run migrations + seed:
   ```bash
   make migrate-up
   ```
4. Run app:
   ```bash
   make run
   ```
5. Open http://localhost:3000

## Default users
- Admin: `admin@example.com` / `password123`
- Submitter: `owner@example.com` / `password123`

## Routes
Public:
- `/:countryCode`
- `/:countryCode/deals`
- `/:countryCode/city/:city`
- `/:countryCode/category/:categorySlug`
- `/:countryCode/deal/:dealSlug`

Account:
- `/account/register`
- `/account/login`
- `/account/logout`
- `/account/submissions`

Admin:
- `/admin`
- `/admin/moderation`
- `/admin/users`
- `/admin/config`
- `/admin/master`

Language:
- Query parameter `?lang=en|si|ta`, persisted to cookie.

## Environment variables
- `PORT` (default `3000`)
- `DATABASE_URL` (default `postgres://postgres:postgres@localhost:5432/deals?sslmode=disable`)
- `SESSION_COOKIE_SECURE` (`false` for local)
- `MAX_UPLOAD_MB` (default `5`)

## Project structure
- `cmd/server` application entrypoint
- `cmd/migrate` migration runner
- `internal/config`, `internal/db`
- `internal/repo`, `internal/service`
- `internal/http` (middleware + handlers)
- `internal/views` (templ components)
- `migrations`
- `static`
