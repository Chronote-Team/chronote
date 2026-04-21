# Quickstart: Chronote Backend Contract-Preserving Refactor

## 1. Review The Planning Inputs

Read the feature definition and baseline references before implementation:

- [spec.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/spec.md)
- [plan.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/plan.md)
- [research.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/research.md)
- [http-api.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/contracts/http-api.md)
- `/home/bowen/Coding/chronote/ai_generated/api_documentation.md`
- `/home/bowen/Coding/chronote/docs/superpowers/plans/2026-04-20-chronote-refactor-replacement.md`

## 2. Bootstrap The Root-Level Replacement Workspace

From the repository root:

```bash
mkdir -p cmd/api internal/platform internal/shared internal/modules tests migrations
go mod init chronote-refactor
go mod tidy
```

Create the planned directory skeleton at the repository root before adding slice implementations.

## 3. Implement In Slice Order

Work in this sequence:

1. Bootstrap app and shared response/error primitives
2. Platform adapters for config, Postgres, Redis, S3, JWT, and password services
3. Health slice
4. Users and auth slices
5. Postcards slice
6. Media slice
7. Cutover verification

## 4. Verify Continuously

Run narrow tests as each slice lands. In this workspace, prefer the offline-safe form because the module cache is already populated locally:

```bash
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/shared/... -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/platform/... -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/modules/health/... ./tests/contract/... -run 'Health' -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/modules/users/... ./internal/modules/auth/... ./tests/contract/... -run 'User|Auth|Register|Login|Refresh|Logout' -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/modules/postcards/... ./tests/contract/... -run 'Postcard' -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/modules/media/... ./tests/contract/... -run 'Media' -v
```

Before cutover readiness, run the full verification suite:

```bash
env GOCACHE=/tmp/go-build GOPROXY=off go test ./... -v
```

## 5. Guardrails

- Do not modify `/home/bowen/Coding/chronote`; use it only as a compatibility reference.
- Preserve the public endpoint contract, status behavior, and default error text unless a documented exception is approved.
- Keep handlers transport-only and infrastructure details out of domain/application logic.
- Do not redesign schema semantics during this feature.
