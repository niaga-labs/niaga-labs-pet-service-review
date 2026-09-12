# service-review

Ratings and reviews for runners, shops and owners on Kilat Pet Delivery. Go 1.24 · Gin · GORM ·
PostgreSQL · Kafka. Listens on port **8007**.

## Running the Service

```bash
# Install dependencies
go mod download

# Point at the shared dev-infra stack (cd ~/Documents/dev-infra; ./dev.ps1 up kilat)
export DB_HOST=localhost DB_PORT=5432 DB_USER=kilat DB_PASSWORD=kilat_secret
export DB_NAME=kilat_review DB_SSL_MODE=disable
export KAFKA_BROKERS=localhost:9092

# Apply the SQL migrations -- run from the repository root, the migration
# source is resolved relative to the working directory
go run ./cmd/migrate

# Start the service
go run ./cmd/server
```

## Migrations

`migrations/` holds golang-migrate files and is the single source of truth for the schema, in every
environment including development. `cmd/migrate` applies them and exits; `cmd/server` applies them at
startup too, so a fresh boot is self-sufficient.

There is deliberately no GORM `AutoMigrate` path. Until KPD-58 the server auto-migrated `ReviewModel`
in development and ran golang-migrate everywhere else — but there was no `migrations/` directory, so
every non-development boot died at startup. One path for all environments removes that class of drift.

## Database schema

`reviews` — one row per review.

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | primary key |
| `booking_id` | UUID | indexed; unique together with `reviewer_id` |
| `reviewer_id` | UUID | indexed; the user writing the review |
| `reviewee_id` | UUID | indexed together with `reviewee_type` |
| `reviewee_type` | VARCHAR(20) | `runner` · `shop` · `owner` |
| `rating` | INTEGER | 1–5 |
| `comment` | TEXT | optional |
| `photo_urls` | JSONB | defaults to `[]` |
| `created_at` / `updated_at` | TIMESTAMPTZ | |

The CHECK and UNIQUE constraints mirror the invariants `NewReview` enforces in
`internal/domain/review/review.go`: rating 1–5, a known reviewee type, no self-review, and one review
per booking per reviewer.
