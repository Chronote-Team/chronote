# Implementation Plan: Chronote Backend Contract-Preserving Refactor

**Branch**: `refactor/all` | **Date**: 2026-04-20 | **Spec**: [spec.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/spec.md)
**Input**: Feature specification from `/specs/refactor/all/spec.md` plus implementation direction from `/home/bowen/Coding/chronote/docs/superpowers/plans/2026-04-20-chronote-refactor-replacement.md`

**Note**: This plan adapts the external replacement plan to the local spec-kit workflow and current repository structure. Per user direction, the replacement application is implemented directly at the repository root while preserving the constitution's required `cmd/api`, `internal/platform`, `internal/shared`, `internal/modules`, `migrations`, and `tests` boundaries.

## Summary

Build a replacement Chronote backend at the repository root that preserves the existing API contract, response envelope, status behavior, default error text, and current data semantics while separating transport, application logic, domain rules, and infrastructure adapters. Execution proceeds slice-by-slice in compatibility-first order: bootstrap and shared primitives, health, users/auth, postcards, media, then cutover verification.

## Technical Context

**Language/Version**: Go 1.25  
**Primary Dependencies**: Gin, GORM, PostgreSQL driver, Redis client, AWS SDK v2 S3 client, JWT library, bcrypt/password hashing helpers  
**Storage**: PostgreSQL, Redis, S3-compatible object storage  
**Testing**: Go `testing` package, table-driven unit tests, HTTP contract tests, integration tests  
**Target Platform**: Linux server environment for HTTP API deployment  
**Project Type**: Web service  
**Performance Goals**: Preserve existing endpoint behavior while keeping lightweight health checks responsive and supporting contract verification across all supported routes  
**Constraints**: No client-visible endpoint redesign; preserve `code` / `message` / `data`; preserve existing status-code behavior and default error text; preserve existing schema semantics; keep legacy repo read-only and isolated from implementation changes  
**Scale/Scope**: One replacement backend covering health, user/auth, postcards, media, and dependency health checks across all current supported API routes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Planned structure uses `cmd/api`, `internal/platform`, `internal/shared`, `internal/modules/<module>`, `migrations`, `tests`, and `docs`, or records an equivalent boundary-preserving alternative.
- [x] Each affected module keeps `http`, `app`, `domain`, and `infra` responsibilities separate; business logic is not placed in handlers or adapters.
- [x] API changes define stable request, response, and error contracts with explicit validation and no persistence-model leakage.
- [x] External dependencies are injected behind interfaces where business logic touches them, and concrete wiring stays in `internal/platform`.
- [x] Preserved or changed domain invariants are identified explicitly.
- [x] Unit, integration, and contract test coverage is planned for every changed behavior that requires it under the constitution.

## Project Structure

### Documentation (this feature)

```text
specs/refactor/all/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── http-api.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
└── api/

internal/
├── platform/
│   ├── app/
│   ├── auth/
│   ├── config/
│   ├── db/
│   ├── http/
│   ├── redis/
│   └── s3/
├── shared/
│   ├── errs/
│   ├── pagination/
│   └── response/
└── modules/
    ├── auth/
    │   ├── app/
    │   ├── infra/
    │   └── http/
    ├── health/
    │   ├── app/
    │   └── http/
    ├── media/
    │   ├── app/
    │   ├── domain/
    │   ├── infra/
    │   └── http/
    ├── postcards/
    │   ├── app/
    │   ├── domain/
    │   ├── infra/
    │   └── http/
    └── users/
        ├── app/
        ├── domain/
        ├── infra/
        └── http/

migrations/
tests/
├── contract/
├── integration/
└── unit/
docs/
specs/
```

**Structure Decision**: Development now uses the repository root as the replacement workspace. This matches the constitution's preferred layout directly and removes the extra `refactor/` nesting while still keeping the legacy repository read-only.

## Phase 0: Research Decisions

Phase 0 outputs are captured in [research.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/research.md). Key decisions:

1. Build the replacement service directly at the repository root during implementation.
2. Use contract tests as the primary guardrail for response envelope, status behavior, and message compatibility across route groups.
3. Keep domain/application code transport-neutral and inject persistence, token, blacklist, storage, and dependency-health adapters behind interfaces.
4. Preserve current schema semantics for `users`, `postcards`, and `postcard_media`, deferring schema redesign and versioned migration cleanup until after compatibility is achieved.

## Phase 1: Design Artifacts

### Data Model

- Create [data-model.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/data-model.md) with entities for User, Auth Tokens, Postcard, Postcard Media, and Health Status.
- Capture validation rules, ownership rules, visibility rules, token-type semantics, and media-group restrictions.

### Interface Contracts

- Create [http-api.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/contracts/http-api.md) to record the preserved public HTTP contract.
- Cover route groups, auth expectations, envelope rules, status-code expectations, and compatibility-sensitive message behavior.

### Quickstart

- Create [quickstart.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/quickstart.md) with implementation bootstrap and verification steps for the root-level workspace.

## Implementation Phases

### Phase A: Bootstrap And Shared Primitives

Goal: create the root-level Go module, router bootstrap, shared response envelope helpers, typed application errors, and HTTP error mapping.

- Bootstrap `go.mod`, `cmd/api/main.go`, and `internal/platform/http/*`.
- Add integration smoke tests proving the router can register `/health` without binding network sockets.
- Add shared response and error packages that enforce the current `code` / `message` / `data` contract.

### Phase B: Platform Adapters

Goal: move configuration and external dependency wiring into `internal/platform`.

- Port configuration keys and environment loading from the current Chronote backend.
- Create Postgres, Redis, S3, JWT, and password services behind interfaces.
- Keep application modules free of package globals and direct SDK/client coupling.

### Phase C: Health Slice

Goal: implement `GET /health` and `GET /health/details` first as the smallest complete vertical slice.

- Build transport-neutral health evaluation in `internal/modules/health/app`.
- Preserve lightweight and detailed health status behavior, including degraded S3 semantics.
- Add contract tests for `200`, `207`, and `503` outcomes and component payload compatibility.

### Phase D: Users And Auth Slices

Goal: preserve registration, login, refresh, logout, user info, avatar, display name, and password update behavior.

- Separate user persistence, auth orchestration, token blacklist behavior, and JWT middleware.
- Preserve refresh-token body usage, bearer-token access enforcement, and current message defaults.
- Add unit tests for normalization and validation rules plus contract tests for all `/user/*` routes.

### Phase E: Postcards Slice

Goal: preserve postcard CRUD, visibility, pagination, and owner-only mutation rules.

- Move validation and access policy into postcard domain/application layers.
- Preserve JSON and multipart create flows, optional auth on reads, and current list/detail shapes.
- Add contract tests for all `/v1/postcards` CRUD endpoints.

### Phase F: Media Slice

Goal: preserve media upload, listing, reorder, deletion, and storage integration behavior.

- Implement media-group restrictions, type validation, ordering, and postcard ownership checks.
- Keep storage and image-processing concerns behind interfaces and infra adapters.
- Add contract and integration coverage for `/v1/postcards/:id/media*`.

### Phase G: Cutover Readiness

Goal: prove the refactor can replace the legacy implementation safely.

- Run the full contract and integration test matrix against the replacement workspace.
- Document any approved compatibility exceptions explicitly before cutover.
- Prepare later root-level promotion and deployment-file updates only after parity is demonstrated.

## Test Strategy

- **Unit tests**: domain and application rules for health, users, auth, postcards, and media without real PostgreSQL, Redis, or S3 dependencies.
- **Contract tests**: endpoint-level compatibility for route paths, request/response schemas, envelope shape, status codes, and key message text.
- **Integration tests**: router wiring, middleware behavior, repository integration, storage adapters, and cross-module flows in the replacement app.
- **Regression focus**: auth token semantics, anonymous postcard reads, owner-only mutations, media ordering, and degraded-vs-unavailable health states.

## File Map

### Bootstrap And Platform

- `go.mod`
- `cmd/api/main.go`
- `internal/platform/app/app.go`
- `internal/platform/config/config.go`
- `internal/platform/db/postgres.go`
- `internal/platform/redis/client.go`
- `internal/platform/s3/client.go`
- `internal/platform/http/router.go`
- `internal/platform/http/server.go`
- `internal/platform/auth/jwt_service.go`
- `internal/platform/auth/password_service.go`

### Shared

- `internal/shared/errs/errors.go`
- `internal/shared/errs/http_mapper.go`
- `internal/shared/response/envelope.go`
- `internal/shared/pagination/pagination.go`

### Modules

- `internal/modules/health/...`
- `internal/modules/users/...`
- `internal/modules/auth/...`
- `internal/modules/postcards/...`
- `internal/modules/media/...`

### Verification

- `tests/contract/...`
- `tests/integration/...`
- `tests/unit/...`

## Risks And Mitigations

- **Undocumented behavior drift**: Mitigate with contract tests derived from the current API documentation and spot checks against the legacy code.
- **Error-message incompatibility**: Centralize HTTP mapping and preserve message text at the transport boundary unless a documented exception is approved.
- **Schema assumption mismatch**: Keep the current schema semantics intact and avoid redesign during this feature.
- **Cross-layer leakage**: Enforce DTO separation and interface-driven dependencies in every slice review.
- **Replacement sprawl**: Build slice-by-slice at the root-level architecture boundaries and defer deployment cutover mechanics until parity is verified.

## Complexity Tracking

No constitution violations are currently required. The repository root now follows the constitution's preferred layout directly.
