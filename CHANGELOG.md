# Changelog

All notable changes to service-review are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- The service could not start in any environment other than `APP_ENV=development`.
  `cmd/server` called `RunMigrations("migrations")`, but the repository had no
  `migrations/` directory, so `migrate.New` failed and the process exited during
  startup. (KPD-58)

### Added

- `migrations/001_create_reviews.{up,down}.sql`: the `reviews` schema, with CHECK
  and UNIQUE constraints mirroring the invariants `NewReview` enforces -- rating
  1-5, known reviewee type, no self-review, one review per booking per reviewer.
- `cmd/migrate`: applies the migrations and exits.
- `README.md`: the repository had none. Documents the run and migrate commands
  against the shared dev-infra stack, and the schema.
- `CHANGELOG.md`: this file. Partially advances KPD-52.

### Changed

- `cmd/server`: the development-only GORM `AutoMigrate` branch is gone. The SQL
  migrations now own the schema in every environment, so development can no
  longer drift from the rest. (KPD-58)
