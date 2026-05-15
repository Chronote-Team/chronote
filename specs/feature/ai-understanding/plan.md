# Implementation Plan: Postcard AI Understanding

**Branch**: `feature/ai-understanding` | **Date**: 2026-05-13 | **Spec**: [spec.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding/spec.md)
**Input**: Feature specification from `/specs/feature/ai-understanding/spec.md` plus implementation direction from `docs/postcard-ai-understanding-final.md`, using the index section for scope/retrieval context and the "Create A Technical Implementation Plan" section for technical direction.

**Note**: `.specify/scripts/bash/setup-plan.sh --json` was attempted first, but the spec-kit branch validator rejects `feature/ai-understanding` even though the user explicitly required this branch name. This plan therefore uses the same resolved feature directory manually: `/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding`.

## Summary

Build a backend-internal postcard AI understanding workflow that reacts to postcard creation, postcard updates, media changes, scheduled backfills, or private retries. The workflow stores durable analysis jobs, analyzes attached images before final postcard understanding, validates structured output, stores image-level and postcard-level results separately, reuses unchanged media analysis, and preserves all normal client app postcard API behavior unchanged.

## Technical Context

**Language/Version**: Go 1.25  
**Primary Dependencies**: Gin, GORM, PostgreSQL driver, Redis client, AWS SDK v2 S3 client, OpenAI Responses-compatible HTTP client, JSON schema validation  
**Storage**: PostgreSQL for durable jobs/results, Redis for short-lived coordination, S3-compatible object storage for private media  
**Testing**: Go `testing` package, table-driven unit tests, HTTP contract tests, integration tests, offline-safe `env GOCACHE=/tmp/go-build GOPROXY=off go test ./... -v`  
**Target Platform**: Linux backend service  
**Project Type**: Backend web service with internal worker workflow  
**Performance Goals**: Do not block normal client-app postcard create/update/read flows on AI calls; reuse cached media analysis whenever possible; recover due jobs after worker interruption  
**Constraints**: Preserve public Chronote API contract; no client-facing analysis endpoints or response fields in this phase; keep business logic out of Gin handlers, GORM models, storage adapters, and AI provider adapters; never log raw postcard text, raw photos, API keys, signed URLs, or full provider payloads  
**Scale/Scope**: Current phase covers backend input, analysis eligibility, internal job enqueue/retry/backfill hooks, image understanding, postcard understanding, structured output validation, durable storage, stale-result prevention, and operational inspection of internal state

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
specs/feature/ai-understanding/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── internal-workflow.md
├── checklists/
│   └── requirements.md
└── tasks.md              # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
└── api/

internal/
├── platform/
│   ├── app/
│   ├── config/
│   ├── db/
│   ├── http/
│   ├── redis/
│   └── s3/
├── shared/
│   ├── errs/
│   └── response/
└── modules/
    ├── postcards/
    │   ├── app/
    │   ├── domain/
    │   ├── infra/
    │   └── http/
    ├── media/
    │   ├── app/
    │   ├── domain/
    │   ├── infra/
    │   └── http/
    └── postcardai/
        ├── app/
        ├── domain/
        └── infra/

migrations/
tests/
├── contract/
├── integration/
└── unit/
docs/
```

**Structure Decision**: Add a new `internal/modules/postcardai` vertical slice for internal analysis jobs, AI understanding domain rules, provider gateway interfaces, and persistence adapters. Existing `postcards` and `media` modules remain owners of user-facing postcard/media behavior; they call postcard AI application interfaces for internal enqueue hooks only. Concrete OpenAI-compatible HTTP, Redis coordination, PostgreSQL repositories, and S3 presigning are wired in `internal/platform` and implemented behind interfaces.

## Phase 0: Research Decisions

Phase 0 outputs are captured in [research.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding/research.md). Key decisions:

1. Create a dedicated `postcardai` module rather than placing AI workflow logic in postcard or media handlers.
2. Use PostgreSQL as the durable source of truth for jobs and results; Redis is only coordination.
3. Use an OpenAI Responses-compatible provider gateway behind an `AIClient` interface.
4. Generate short-lived private media access only inside backend workers and never store or return those links.
5. Validate structured AI output before successful storage and allow at most one schema-repair retry.
6. Record prompt, schema, model, media, and postcard versions to support reuse, traceability, and stale-result prevention.

## Phase 1: Design Artifacts

### Data Model

- Create [data-model.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding/data-model.md) with entities for Analysis Job, Media Analysis, Postcard Analysis, prompt/schema/model versions, analysis status, and provider call metadata.
- Capture uniqueness rules, version matching, confidence/uncertainty storage, and state transitions for pending, processing, succeeded, failed, unavailable, and stale outcomes.

### Interface Contracts

- Create [internal-workflow.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding/contracts/internal-workflow.md) to record internal service contracts, trigger points, provider boundaries, storage rules, and public API non-change guarantees.
- No public client-facing HTTP contract is added for analysis in this phase.

### Quickstart

- Create [quickstart.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding/quickstart.md) with implementation sequence, configuration, migration, and verification steps.

## Implementation Phases

### Phase A: Domain And Storage Shape

Goal: define internal AI workflow invariants and durable persistence without touching public postcard response shapes.

- Add `internal/modules/postcardai/domain` types for analysis status, job identity, version keys, confidence, provider error category, media understanding, and postcard understanding.
- Add migrations for `ai_analysis_jobs`, `media_ai_analysis`, and `postcard_ai_analysis`.
- Add uniqueness constraints for media and postcard result version keys.
- Add tests for status transitions, stale version prevention, and version-key equality.

### Phase B: Application Services And Interfaces

Goal: keep orchestration testable and dependency-injected.

- Add `postcardai/app` service interfaces for postcard lookup, media lookup, result repository, job repository, lock/queue coordination, storage presigning, clock, and AI provider calls.
- Implement `EnqueuePostcardAnalysis`, `RunNextAnalysisJob`, and `RetryAnalysisJob` use cases.
- Add idempotency rules for duplicate enqueue requests and unchanged media reuse.
- Add unit tests with in-memory fakes for enqueue, cache reuse, retry, stale detection, and validation failures.

### Phase C: Provider Gateway, Prompts, And Schemas

Goal: isolate AI-provider details and keep accepted output deterministic.

- Add `postcardai/infra/ai` provider adapter for OpenAI Responses-compatible text/image requests behind an `AIClient` interface.
- Store versioned image and postcard prompts under the postcard AI infra boundary.
- Store versioned structured output schema definitions paired with backend validation.
- Implement one schema-repair retry for malformed output.
- Add provider adapter tests with recorded or fake responses that do not contain raw user content.

### Phase D: Worker And Platform Wiring

Goal: run internal analysis jobs without blocking normal postcard flows.

- Add worker runner wiring in `internal/platform/app` or an adjacent platform worker bootstrap.
- Wire PostgreSQL repositories, Redis coordination, S3 presigning, and AI provider configuration in `internal/platform`.
- Ensure worker failures leave durable recoverable state in PostgreSQL.
- Add integration tests for persistence, recovery after interrupted jobs, Redis-unavailable behavior, and signed URL regeneration.

### Phase E: Postcard And Media Trigger Hooks

Goal: enqueue analysis from existing backend events while preserving public behavior.

- Call `EnqueuePostcardAnalysis` from postcard creation/update and media change application paths after the user-facing mutation succeeds.
- Do not add public analyze/status/result routes.
- Do not add analysis fields to normal client app responses.
- Add contract tests proving existing public postcard and media endpoints remain unchanged.

### Phase F: Operational Backfill And Retry Hooks

Goal: make analysis repairable internally without broad public API surface.

- Add private operational service methods for scheduled backfill and retry.
- Keep future admin/management endpoints out of scope unless a later spec adds them.
- Add tests for backfill eligibility, retry limits, failure classification, and stale-job handling.

## Test Strategy

- **Unit tests**: domain status transitions, version-key matching, enqueue idempotency, media analysis cache selection, schema validation success/failure, retry decision rules, stale version prevention, prompt/schema/model propagation, and safe logging helpers.
- **Contract tests**: existing public postcard and media endpoints remain unchanged; no client-facing analysis endpoint is introduced; normal responses do not expose analysis status/result fields.
- **Integration tests**: PostgreSQL job/result persistence, uniqueness constraints, worker recovery, Redis lock behavior when available, Redis-unavailable recovery path, storage presign regeneration, and provider gateway behavior using fakes.
- **Privacy tests**: representative success and failure logs exclude raw postcard text, raw image data, API keys, signed URLs, and full provider payloads.

## File Map

### New Module

- `internal/modules/postcardai/domain/status.go`
- `internal/modules/postcardai/domain/version.go`
- `internal/modules/postcardai/domain/analysis.go`
- `internal/modules/postcardai/app/service.go`
- `internal/modules/postcardai/app/worker.go`
- `internal/modules/postcardai/app/repositories.go`
- `internal/modules/postcardai/app/provider.go`
- `internal/modules/postcardai/app/storage.go`
- `internal/modules/postcardai/infra/gorm_repository.go`
- `internal/modules/postcardai/infra/redis_coordination.go`
- `internal/modules/postcardai/infra/ai/openai_responses_client.go`
- `internal/modules/postcardai/infra/ai/prompts/image_understanding_v1.tmpl`
- `internal/modules/postcardai/infra/ai/prompts/postcard_understanding_v1.tmpl`
- `internal/modules/postcardai/infra/ai/schemas.go`

### Existing Modules And Platform

- `internal/modules/postcards/app/service.go`
- `internal/modules/media/app/service.go`
- `internal/platform/app/app.go`
- `internal/platform/config/config.go`
- `internal/platform/s3/client.go`
- `internal/platform/redis/client.go`

### Migrations And Tests

- `migrations/*_create_ai_analysis_tables.sql`
- `tests/contract/postcard_contract_test.go`
- `tests/contract/media_contract_test.go`
- `tests/integration/postcard_ai_worker_test.go`
- `tests/unit/postcardai/*`

## Risks And Mitigations

- **Public API drift**: Mitigate with contract tests proving no public analysis endpoints or response fields are added.
- **Provider payload leakage**: Mitigate with safe logging rules, redaction tests, and provider adapter boundaries.
- **Duplicate or stale results**: Mitigate with version-key uniqueness, stale write checks, and job/result state transitions.
- **Workflow blocking user requests**: Mitigate by enqueueing internal jobs after normal mutations and running AI work outside request handling.
- **Redis dependency becoming a hard blocker**: Mitigate by keeping durable job state in PostgreSQL and treating Redis as coordination only.
- **Malformed AI output**: Mitigate with structured validation and at most one schema-repair retry before non-successful storage.

## Complexity Tracking

No constitution violations are currently required. The feature adds an internal module and durable storage tables while preserving existing public API contracts and root-level architecture boundaries.

## Post-Design Constitution Check

- [x] Planned structure preserves `cmd/api`, `internal/platform`, `internal/shared`, `internal/modules/<module>`, `migrations`, `tests`, and `docs`.
- [x] The new `postcardai` module keeps `domain`, `app`, and `infra` responsibilities separate; no HTTP package is planned because this phase has no public or admin analysis endpoint.
- [x] Public API contracts remain stable because this phase adds no client-facing endpoints and no response fields.
- [x] PostgreSQL, Redis, S3, and AI provider access are injected behind interfaces; concrete wiring stays in `internal/platform`.
- [x] Preserved and new invariants are explicitly recorded in the spec and data model.
- [x] Unit, integration, contract, and privacy-focused tests are planned for changed behavior.
