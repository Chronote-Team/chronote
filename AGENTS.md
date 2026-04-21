# chronote-refactor Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-04-21

## Active Technologies

- Go 1.25 + Gin, GORM, PostgreSQL driver, Redis client, AWS SDK v2 S3 client, JWT library, bcrypt/password hashing helpers (refactor/all)

## Project Structure

```text
specs/
  refactor/all/
cmd/api/
internal/platform/
internal/shared/
internal/modules/
migrations/
tests/
docs/
```

## Non-Negotiable Constraints

- Preserve the public Chronote API contract, including endpoint paths, request and response shapes, status behavior, and default error-message text unless an explicit exception is approved.
- Keep business logic out of Gin handlers and infrastructure adapters; use `http`, `app`, `domain`, and `infra` boundaries per module.
- Keep concrete dependency wiring in `internal/platform`; inject repositories and external services behind interfaces.
- Do not modify `/home/bowen/Coding/chronote`; use it only as the read-only compatibility reference.
- Treat the repository root as the replacement workspace while keeping `/home/bowen/Coding/chronote` read-only.

## Commands

- `bash .specify/scripts/bash/check-prerequisites.sh --paths-only`
- `bash .specify/scripts/bash/tests/common-branch-resolution-test.sh`
- `env GOCACHE=/tmp/go-build GOPROXY=off go test ./... -v`

## Code Style

Go 1.25: Follow standard conventions

## Recent Changes

- refactor/all: Added Go 1.25 + Gin, GORM, PostgreSQL driver, Redis client, AWS SDK v2 S3 client, JWT library, bcrypt/password hashing helpers

<!-- MANUAL ADDITIONS START -->
- The current verified root-level replacement slice covers health, users/auth, postcards, and media with in-memory test wiring in `internal/platform/app/app.go`.
- Prefer the offline-safe test command above in this workspace because network access may be restricted while Go modules are already available in the local cache.
<!-- MANUAL ADDITIONS END -->
