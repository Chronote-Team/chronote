# chronote-refactor Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-05-10

## Active Technologies
- Go 1.25 + Gin, GORM, PostgreSQL driver, existing Chronote shared response/error helpers, existing auth middleware (feature/random)
- PostgreSQL for production postcard data; in-memory repositories for test app wiring (feature/random)

- Go 1.25 + Gin, GORM, PostgreSQL driver, Redis client, AWS SDK v2 S3 client, JWT library, bcrypt/password hashing helpers (refactor/all)
- Go 1.25 + Gin, GORM, PostgreSQL driver, Redis client, AWS SDK v2 S3 client, OpenAI Responses-compatible HTTP client, JSON schema validation (feature/ai-understanding)

## Project Structure

```text
specs/
  refactor/all/
  feature/ai-understanding/
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
- feature/random: Added Go 1.25 + Gin, GORM, PostgreSQL driver, existing Chronote shared response/error helpers, existing auth middleware

- refactor/all: Added Go 1.25 + Gin, GORM, PostgreSQL driver, Redis client, AWS SDK v2 S3 client, JWT library, bcrypt/password hashing helpers
- feature/ai-understanding: Added backend-internal postcard AI understanding workflow planning with durable jobs/results, Redis coordination, S3 private media presigning, OpenAI-compatible provider boundary, and structured output validation

<!-- MANUAL ADDITIONS START -->
- The current verified root-level replacement slice covers health, users/auth, postcards, and media with in-memory test wiring in `internal/platform/app/app.go`.
- Prefer the offline-safe test command above in this workspace because network access may be restricted while Go modules are already available in the local cache.
- US3 cutover verification uses isolated Podman containers on a dedicated network such as `chronote_us3_test`; do not run those tests on the default bridge or alongside unrelated local service containers.
<!-- MANUAL ADDITIONS END -->

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
<!-- SPECKIT END -->
