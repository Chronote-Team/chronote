# Implementation Plan: Postcard AI Understanding Worker

**Branch**: `feature/ai-understanding-worker` | **Date**: 2026-05-15 | **Spec**: [spec.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding-worker/spec.md)
**Input**: Feature specification from `/specs/feature/ai-understanding-worker/spec.md` plus implementation direction from `docs/postcard-ai-understanding-worker.md`, using the index section for scope/retrieval context and the "Create A Technical Implementation Plan" section for technical direction.

## Summary

Complete the backend-internal postcard AI understanding feature by adding the missing deployable worker runtime and live provider response parsing. The current code already enqueues durable AI jobs and implements `RunNextAnalysisJob`, but deployment only starts the HTTP API process, so queued jobs remain pending. This plan adds a separate worker command/service, shared platform construction for analysis dependencies, OpenAI Responses parsing into validated `AIResult` values, Compose wiring, and tests proving jobs are consumed while public Chronote API behavior remains unchanged.

## Technical Context

**Language/Version**: Go 1.25  
**Primary Dependencies**: Gin, GORM, PostgreSQL driver, Redis client, AWS SDK v2 S3 client, Go standard `net/http`, OpenAI Responses-compatible HTTP client, existing `postcardai` app/domain/infra packages  
**Storage**: Existing PostgreSQL tables for `ai_analysis_jobs`, `media_ai_analysis`, and `postcard_ai_analysis`; S3-compatible storage for private media objects; Redis for existing auth/health usage and possible future coordination only  
**Testing**: Go `testing` package, table-driven unit tests, HTTP contract tests, integration tests with fakes, offline-safe `env GOCACHE=/tmp/go-build GOPROXY=off go test ./... -v`  
**Target Platform**: Linux backend containers run by Docker Compose or Podman Compose, including rootless Podman local deployment  
**Project Type**: Backend web service plus background worker process  
**Performance Goals**: Claim newly queued work within 10 seconds under idle local conditions; keep public request/response flows independent of AI provider latency; process immediately after successful job completion until the queue is empty  
**Constraints**: Preserve public Chronote API contract; do not add public analysis endpoints or response fields; keep business logic out of command entrypoints, Gin handlers, GORM models, storage adapters, and provider adapters; never log raw postcard text, raw image bytes, signed URLs, API keys, or full provider payloads  
**Scale/Scope**: One new worker runtime path, one shared platform wiring path for analysis dependencies, provider response parsing, Compose worker service, and tests for worker loop behavior, parser behavior, public contract non-change, and job consumption using controlled fakes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Planned structure preserves `cmd/api`, adds `cmd/worker` as an entrypoint-only process, keeps concrete wiring in `internal/platform`, keeps business logic in `internal/modules/postcardai/app`, and keeps data rules in `internal/modules/postcardai/domain`.
- [x] No business logic is planned for Gin handlers, command entrypoints, GORM models, storage adapters, or provider adapters.
- [x] Public API contracts remain stable because this feature adds no public endpoints and no public response fields.
- [x] PostgreSQL, Redis, S3, and AI provider access remain injected behind app-level interfaces; concrete construction stays in `internal/platform`.
- [x] Domain invariants for postcards, media ownership, media groups, media ordering, and private media access are preserved.
- [x] Unit, integration, deployment smoke, and contract verification are planned for changed behavior.

## Project Structure

### Documentation (this feature)

```text
specs/feature/ai-understanding-worker/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── worker-runtime.md
├── checklists/
│   └── requirements.md
└── tasks.md              # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
cmd/
├── api/
└── worker/

internal/
├── platform/
│   ├── app/
│   │   ├── app.go
│   │   └── analysis_worker.go
│   ├── config/
│   ├── db/
│   ├── http/
│   ├── redis/
│   └── s3/
└── modules/
    └── postcardai/
        ├── app/
        ├── domain/
        └── infra/
            └── ai/

tests/
├── contract/
├── integration/
└── unit/
```

**Structure Decision**: Add `cmd/worker` for process concerns only and add `internal/platform/app/analysis_worker.go` to share analysis dependency construction with the API runtime. Keep `RunNextAnalysisJob` and all job/result decisions inside the existing `postcardai/app.Service`. Update deployment files so the same image contains both API and worker binaries, with Compose starting the worker as a separate service.

## Phase 0: Research Decisions

Phase 0 outputs are captured in [research.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding-worker/research.md). Key decisions:

1. Run AI processing as a separate worker process instead of adding provider calls to the HTTP API process.
2. Wrap `RunNextAnalysisJob` in a small loop runner with worker identity, idle sleep, error sleep, and run-once options.
3. Extract shared analysis dependency construction into platform wiring so API enqueue hooks and worker consumption use the same repositories/provider/storage configuration.
4. Parse successful OpenAI Responses API bodies from common text output locations into `AIResult` and preserve existing failure classification.
5. Build both `/app/chronote` and `/app/chronote-worker` into the same container image and start the worker with a Compose `command`.
6. Keep current database schema for this branch unless implementation proves an additive field is required.

## Phase 1: Design Artifacts

### Data Model

- Create [data-model.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding-worker/data-model.md) to document the existing job/result tables, worker runtime options, worker outcome state, provider parsed result, and deployment service relationship.
- Confirm no new table is planned in the first worker branch.

### Interface Contracts

- Create [worker-runtime.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding-worker/contracts/worker-runtime.md) for process contracts, public API non-change guarantees, worker loop behavior, OpenAI response parsing expectations, Compose runtime contract, and verification boundaries.
- No public client-facing HTTP contract is added for this feature.

### Quickstart

- Create [quickstart.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding-worker/quickstart.md) with implementation sequence, worker configuration, Compose verification, real-provider caveats, and offline test commands.

## Implementation Phases

### Phase A: Worker Runtime Shape

Goal: add the process and loop that consume one existing analysis job per service call.

- Add `cmd/worker/main.go` with config loading, signal-aware context, worker ID defaults, and process-level logging.
- Add worker options for `AI_WORKER_ID`, `AI_WORKER_IDLE_SLEEP`, `AI_WORKER_ERROR_SLEEP`, and `AI_WORKER_RUN_ONCE`.
- Add a loop helper that calls `RunNextAnalysisJob(ctx, workerID)`, sleeps only when there is no work or an error, and exits cleanly on cancellation.
- Add tests for canceled context, idle sleep, error sleep, run-once mode, and repeated job processing.

### Phase B: Shared Platform Construction

Goal: prevent duplicated dependency wiring between API and worker.

- Extract analysis service construction from `internal/platform/app/app.go` into a shared platform helper.
- Keep HTTP router construction unchanged.
- Build the worker runtime from the same PostgreSQL repository, postcard/media source adapters, S3 presigner, and AI provider configuration used by API enqueue wiring.
- Add construction tests where practical to catch missing AI endpoint/provider/S3 config wiring.

### Phase C: OpenAI Responses Parsing

Goal: make live provider success responses usable instead of returning the current parsing-not-configured error.

- Add parser fixtures for successful OpenAI Responses shapes, including `output[].content[].text`, text variants in content entries, and `output_text`.
- Add tests for fenced JSON, missing output, malformed JSON, and response bodies that exceed the bounded reader.
- Implement bounded response reading, text extraction, fence trimming, JSON validation, confidence/uncertainty extraction when present, and `AIResult` return.
- Preserve existing provider error classification for 429, 5xx, 400, and 403 responses.
- Do not log full response bodies in normal operation.

### Phase D: Container And Compose Deployment

Goal: make local Podman/Docker deployment run API and worker together.

- Update `Dockerfile` to build `/app/chronote` from `cmd/api` and `/app/chronote-worker` from `cmd/worker`.
- Keep the image default entrypoint as `/app/chronote`.
- Add a `worker` service to `docker-compose.yml` using the same image and environment as `app`, with `command: ["/app/chronote-worker"]`.
- Make the worker depend on healthy PostgreSQL, healthy Redis, started RustFS, and completed bucket initialization.
- Do not expose worker HTTP ports.

### Phase E: End-To-End Verification

Goal: prove pending jobs move and the public API remains stable.

- Add integration coverage using fakes to prove an enqueued job is consumed by the worker runtime path.
- Add or preserve contract tests proving there are no public analysis endpoints and no analysis fields in normal postcard responses.
- Run the full offline-safe Go test suite.
- Start local Compose stack, create/update postcards through the real API, and confirm `ai_analysis_jobs` moves out of `pending` when the worker is enabled.
- Document real-provider image caveat: local RustFS URLs must be reachable by the external provider or image analysis will fail gracefully.

## Test Strategy

- **Unit tests**: worker loop cancellation, run-once mode, idle sleep, error sleep, repeated work, safe logging helpers if introduced, and OpenAI Responses parser fixtures.
- **Application tests**: existing `RunNextAnalysisJob` behavior remains the central job claim/process/status workflow; add focused coverage only where the worker loop changes invocation behavior.
- **Contract tests**: normal postcard and media public endpoints keep their current request/response/error shapes; no public analyze/status/result endpoints are introduced.
- **Integration tests**: controlled provider/storage fakes prove queued jobs can be consumed through the runtime path and that unchanged media analysis can be reused.
- **Deployment smoke tests**: Compose starts `app` plus `worker`; worker has no exposed port; `/health` remains served only by the API service.
- **Privacy checks**: logs include job IDs/outcomes but exclude API keys, signed URLs, raw postcard content, raw image bytes, and full provider payloads.

## File Map

### New Files

- `cmd/worker/main.go`
- `internal/platform/app/analysis_worker.go`
- `internal/platform/app/analysis_worker_test.go`
- `internal/modules/postcardai/infra/ai/openai_responses_client_test.go`
- `specs/feature/ai-understanding-worker/contracts/worker-runtime.md`

### Modified Files

- `Dockerfile`
- `docker-compose.yml`
- `internal/platform/app/app.go`
- `internal/platform/config/config.go`
- `internal/modules/postcardai/infra/ai/openai_responses_client.go`
- `AGENTS.md`

### Possible Files If Tests Need Them

- `tests/contract/postcard_contract_test.go`
- `tests/integration/postcard_ai_worker_test.go`
- `.env.example`
- `docs/local-real-api-test-guide.md`

## Risks And Mitigations

- **Worker starts but cannot consume jobs**: Mitigate by testing the runtime path with a fake provider and by wiring the worker from the same repositories used by enqueue.
- **Provider success parsing still rejects real responses**: Mitigate with representative OpenAI Responses fixtures before implementation and preserve malformed-output classification for unsupported shapes.
- **Platform wiring drifts between API and worker**: Mitigate by extracting shared analysis construction instead of duplicating config logic.
- **Real image analysis fails in local deployment because provider cannot fetch RustFS URLs**: Mitigate by documenting that text-only real-provider tests can run locally, while image tests need provider-reachable object URLs or a tunnel.
- **Public API changes accidentally**: Mitigate with contract tests and by keeping worker code outside HTTP router construction.
- **Sensitive data leaks in worker logs**: Mitigate by logging only job IDs, statuses, timing, and categorized errors.

## Complexity Tracking

No constitution violations are currently required. The feature adds one process entrypoint and one platform wiring helper while preserving existing module boundaries and public API contracts.

## Post-Design Constitution Check

- [x] Structure preserves the established root layout while adding `cmd/worker` for process startup only.
- [x] Business workflow remains in `internal/modules/postcardai/app`; provider parsing stays in `infra/ai`; concrete wiring stays in `internal/platform`.
- [x] Public API contracts remain stable because this phase adds no client-facing endpoints and no response fields.
- [x] Dependency injection remains intact through app-level interfaces and platform construction.
- [x] Required tests cover new runtime behavior, parser behavior, deployment behavior, and public contract non-change.
