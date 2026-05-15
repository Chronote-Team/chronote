# Tasks: Postcard AI Understanding Worker

**Input**: Design documents from `/specs/feature/ai-understanding-worker/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/worker-runtime.md](contracts/worker-runtime.md), [quickstart.md](quickstart.md)

**Tests**: Required by the feature plan and quickstart for worker loop behavior, provider parsing, public API contract non-change, integration job consumption, deployment shape, and privacy logging.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing. Complete Phase 1 and Phase 2 before starting story work.

## Phase 1: Setup

**Purpose**: Prepare the worker branch structure and configuration surface.

- [X] T001 Create worker command directory and placeholder package documentation in `cmd/worker/main.go`
- [X] T002 Create platform worker wiring file with option/runner stubs in `internal/platform/app/analysis_worker.go`
- [X] T003 [P] Create platform worker test file for loop/config behavior in `internal/platform/app/analysis_worker_test.go`
- [X] T004 [P] Create OpenAI Responses parser test file in `internal/modules/postcardai/infra/ai/openai_responses_client_test.go`
- [X] T005 [P] Create worker integration test skeleton in `tests/integration/postcard_ai_worker_test.go`

---

## Phase 2: Foundational

**Purpose**: Add shared configuration and dependency seams that all user stories depend on.

**Critical**: No user story work should begin until this phase is complete.

- [X] T006 Add AI worker config fields and env bindings for `AI_WORKER_ID`, `AI_WORKER_IDLE_SLEEP`, `AI_WORKER_ERROR_SLEEP`, and `AI_WORKER_RUN_ONCE` in `internal/platform/config/config.go`
- [X] T007 [P] Add config default and env override tests for AI worker settings in `internal/platform/config/config_test.go`
- [X] T008 Extract reusable postcard AI service construction from `newProductionApp` into helper functions in `internal/platform/app/analysis_worker.go`
- [X] T009 Refactor `newProductionApp` to call shared analysis construction while preserving router behavior in `internal/platform/app/app.go`
- [X] T010 Add platform construction tests for enabled/disabled AI, custom endpoint, missing API key, and S3 presigner wiring in `internal/platform/app/analysis_worker_test.go`
- [X] T011 Add a small processor interface around `RunNextAnalysisJob` for testable worker loops in `internal/platform/app/analysis_worker.go`
- [X] T012 Verify existing postcard AI unit tests still compile after platform extraction in `tests/unit/postcardai/test_fakes_test.go`

**Checkpoint**: Worker config and shared analysis construction are ready.

---

## Phase 3: User Story 1 - Consume Queued Understanding Work (Priority: P1) MVP

**Goal**: A deployable background worker claims queued postcard AI jobs and moves them out of pending.

**Independent Test**: Create or update a postcard so an analysis job is queued, run the worker runtime, and verify the job is claimed and reaches a non-pending outcome.

### Tests for User Story 1

- [X] T013 [P] [US1] Add failing worker loop tests for run-once, idle sleep, error sleep, cancellation, and repeated processing in `internal/platform/app/analysis_worker_test.go`
- [X] T014 [P] [US1] Add failing command startup test coverage using testable option parsing helpers in `cmd/worker/main_test.go`
- [X] T015 [P] [US1] Add failing integration test for text-only queued job consumption with fake provider/storage in `tests/integration/postcard_ai_worker_test.go`
- [X] T016 [P] [US1] Add failing integration test for image postcard job consumption storing media and postcard analysis with fake provider/storage in `tests/integration/postcard_ai_worker_test.go`
- [X] T017 [P] [US1] Add failing OpenAI Responses parser tests for `output[].content[].text`, `output_text`, fenced JSON, confidence, and uncertainty in `internal/modules/postcardai/infra/ai/openai_responses_client_test.go`

### Implementation for User Story 1

- [X] T018 [US1] Implement `WorkerOptions`, defaulting, and environment-to-options parsing in `internal/platform/app/analysis_worker.go`
- [X] T019 [US1] Implement `AnalysisWorker` construction with shared postcard AI service dependencies in `internal/platform/app/analysis_worker.go`
- [X] T020 [US1] Implement worker loop around `RunNextAnalysisJob(ctx, workerID)` with idle/error sleeps, run-once support, and context cancellation in `internal/platform/app/analysis_worker.go`
- [X] T021 [US1] Implement safe job outcome logging for claimed, no-work, success, stale, unavailable, failed, and retryable outcomes in `internal/platform/app/analysis_worker.go`
- [X] T022 [US1] Implement signal-aware worker process entrypoint with config load, dependency construction, and exit handling in `cmd/worker/main.go`
- [X] T023 [US1] Implement bounded response body reading and successful Responses API text extraction in `internal/modules/postcardai/infra/ai/openai_responses_client.go`
- [X] T024 [US1] Implement JSON fence trimming, JSON validation, confidence extraction, and uncertainty extraction in `internal/modules/postcardai/infra/ai/openai_responses_client.go`
- [X] T025 [US1] Preserve provider error classification for 429, 5xx, 400, 403, missing output, malformed output, and oversized body in `internal/modules/postcardai/infra/ai/openai_responses_client.go`
- [X] T026 [US1] Update integration fakes to support worker runtime execution and fake provider results in `tests/integration/postcard_ai_worker_test.go`
- [X] T027 [US1] Wire text-only and image job integration assertions to check non-pending job status plus stored analysis rows in `tests/integration/postcard_ai_worker_test.go`
- [X] T028 [US1] Run focused worker/parser/integration checks and document any remaining local-only limitations in `specs/feature/ai-understanding-worker/quickstart.md`

**Checkpoint**: US1 is functional. The worker can consume queued work and valid provider success bodies can become internal `AIResult` values.

---

## Phase 4: User Story 2 - Preserve Normal Chronote Behavior (Priority: P2)

**Goal**: Public postcard/media behavior remains unchanged while the worker runs or is absent.

**Independent Test**: Run existing public behavior tests while analysis is queued and while the worker is active, then verify response shapes and status behavior remain unchanged.

### Tests for User Story 2

- [X] T029 [P] [US2] Add contract assertions that normal postcard responses omit AI status/result fields in `tests/contract/postcard_ai_contract_test.go`
- [X] T030 [P] [US2] Add contract assertions that public analyze and analysis-read endpoints do not exist in `tests/contract/postcard_ai_contract_test.go`
- [X] T031 [P] [US2] Add integration assertion that postcard/media flows work when no worker is running and jobs remain pending in `tests/integration/postcard_ai_worker_test.go`

### Implementation for User Story 2

- [X] T032 [US2] Keep router registration free of public analysis endpoints while adding worker code by verifying `internal/platform/http` router usage through `internal/platform/app/app.go`
- [X] T033 [US2] Preserve postcard and media DTOs so no AI fields are added in `internal/modules/postcards/http` and `internal/modules/media/http`
- [X] T034 [US2] Run public contract tests and update only test expectations that document unchanged AI non-exposure in `tests/contract/postcard_ai_contract_test.go`

**Checkpoint**: US2 is functional. Public API behavior is unchanged with or without background worker processing.

---

## Phase 5: User Story 3 - Handle Retries, Stale Work, and Reuse (Priority: P3)

**Goal**: The worker correctly classifies stale work, preserves retryable state, and reuses current media understanding.

**Independent Test**: Queue work, mutate the target postcard before processing, simulate provider failures, and process unchanged media more than once; verify each outcome is classified correctly.

### Tests for User Story 3

- [X] T035 [P] [US3] Add stale queued job integration test using worker runtime path in `tests/integration/postcard_ai_worker_test.go`
- [X] T036 [P] [US3] Add provider unavailable retry integration test using worker runtime path in `tests/integration/postcard_ai_worker_test.go`
- [X] T037 [P] [US3] Add unchanged media reuse integration test proving no duplicate image provider call in `tests/integration/postcard_ai_worker_test.go`
- [X] T038 [P] [US3] Add malformed-output repair parser/provider test coverage in `internal/modules/postcardai/infra/ai/openai_responses_client_test.go`

### Implementation for User Story 3

- [X] T039 [US3] Adjust worker loop error handling so retryable provider errors sleep safely without losing durable state in `internal/platform/app/analysis_worker.go`
- [X] T040 [US3] Ensure `RunNextAnalysisJob` outcomes for stale, unavailable, failed, succeeded, and no-work map to safe worker log categories in `internal/platform/app/analysis_worker.go`
- [X] T041 [US3] Ensure OpenAI parser returns malformed-output errors for missing JSON and repairable malformed JSON without raw body logging in `internal/modules/postcardai/infra/ai/openai_responses_client.go`
- [X] T042 [US3] Reuse existing media analysis assertions through the worker path in `tests/integration/postcard_ai_worker_test.go`

**Checkpoint**: US3 is functional. Stale, retry, malformed output, and media reuse behavior is correct through the worker runtime path.

---

## Phase 6: User Story 4 - Operate Securely and Observably (Priority: P4)

**Goal**: Operators can verify worker progress from safe logs and deploy API plus worker together locally.

**Independent Test**: Process successful, stale, failed, and retryable work; inspect logs for job IDs and outcome categories; verify secrets and raw content are absent; confirm Compose starts API and worker services.

### Tests for User Story 4

- [X] T043 [P] [US4] Add privacy test for worker logs excluding API keys, signed URLs, raw postcard text, raw image bytes, and full provider payloads in `tests/unit/postcardai/privacy_test.go`
- [X] T044 [P] [US4] Add Dockerfile build contract test or documented static check for both binaries in `tests/integration/full_stack_verification_test.go`
- [X] T045 [P] [US4] Add Compose service shape assertion for worker command, dependencies, environment, and no ports in `tests/integration/full_stack_verification_test.go`

### Implementation for User Story 4

- [X] T046 [US4] Update `Dockerfile` to build `/app/chronote` from `cmd/api` and `/app/chronote-worker` from `cmd/worker`
- [X] T047 [US4] Add `worker` service to `docker-compose.yml` using the same image, same DB/Redis/S3/AI env values, worker-specific env values, and `command: ["/app/chronote-worker"]`
- [X] T048 [US4] Configure `docker-compose.yml` worker dependencies on healthy postgres, healthy redis, started rustfs, and completed s3-init with no port mapping
- [X] T049 [US4] Update local smoke-test instructions for `podman compose logs worker` and pending-job verification in `specs/feature/ai-understanding-worker/quickstart.md`

**Checkpoint**: US4 is functional. Operators can run and inspect API plus worker locally without exposing worker ports or sensitive data.

---

## Final Phase: Polish & Cross-Cutting Concerns

**Purpose**: Final verification, documentation, and cleanup across all user stories.

- [X] T050 [P] Update `.env.example` with optional AI worker settings and AI endpoint comments in `.env.example`
- [X] T051 [P] Update real API testing docs with worker-specific verification steps in `docs/local-real-api-test-guide.md`
- [X] T052 Run `gofmt` on changed Go files under `cmd/worker`, `internal/platform/app`, `internal/platform/config`, and `internal/modules/postcardai/infra/ai`
- [X] T053 Run focused tests from `specs/feature/ai-understanding-worker/quickstart.md` and record any unavailable local-provider caveats in `specs/feature/ai-understanding-worker/quickstart.md`
- [X] T054 Run full offline verification `env GOCACHE=/tmp/go-build GOPROXY=off go test ./... -v` from repository root and record pass/fail notes in `specs/feature/ai-understanding-worker/quickstart.md`
- [X] T055 Review `git diff` for public API changes and confirm only worker/runtime/parser/deployment changes are present in `specs/feature/ai-understanding-worker/tasks.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Phase 1 and blocks all user stories.
- **US1 (Phase 3)**: Depends on Phase 2 and is the MVP.
- **US2 (Phase 4)**: Depends on Phase 2; can run after US1 tests exist, but should remain independently verifiable.
- **US3 (Phase 5)**: Depends on Phase 2 and benefits from US1 worker loop/parser completion.
- **US4 (Phase 6)**: Depends on Phase 2 and needs `cmd/worker` from US1 for container wiring.
- **Final Phase**: Depends on all desired user stories.

### User Story Dependencies

- **US1 - Consume Queued Understanding Work**: Required MVP; no dependency on other stories after foundation.
- **US2 - Preserve Normal Chronote Behavior**: Can be verified independently after foundation, but should be re-run after US1 and US4 changes.
- **US3 - Handle Retries, Stale Work, and Reuse**: Depends on worker runtime behavior from US1.
- **US4 - Operate Securely and Observably**: Depends on worker command from US1 and shared config from foundation.

### Within Each User Story

- Write failing tests first.
- Implement the smallest runtime/provider/platform code needed for that story.
- Run the focused tests listed in the story.
- Keep public API assertions passing before moving to the next story.

## Parallel Opportunities

- T003, T004, and T005 can run in parallel after T001 and T002 create target paths.
- T007 can run in parallel with T008 through T011 after T006 defines config fields.
- T013 through T017 can run in parallel because they touch different test concerns.
- T029 through T031 can run in parallel with US1 implementation because they assert public non-change.
- T035 through T038 can run in parallel after US1 worker fixtures exist.
- T043 through T045 can run in parallel after US1 worker logging/command shape exists.
- T050 and T051 can run in parallel during polish.

## Parallel Example: User Story 1

```text
Task: "T013 [P] [US1] Add failing worker loop tests for run-once, idle sleep, error sleep, cancellation, and repeated processing in internal/platform/app/analysis_worker_test.go"
Task: "T015 [P] [US1] Add failing integration test for text-only queued job consumption with fake provider/storage in tests/integration/postcard_ai_worker_test.go"
Task: "T017 [P] [US1] Add failing OpenAI Responses parser tests for output locations and metadata in internal/modules/postcardai/infra/ai/openai_responses_client_test.go"
```

## Parallel Example: User Story 2

```text
Task: "T029 [P] [US2] Add contract assertions that normal postcard responses omit AI status/result fields in tests/contract/postcard_ai_contract_test.go"
Task: "T031 [P] [US2] Add integration assertion that postcard/media flows work when no worker is running in tests/integration/postcard_ai_worker_test.go"
```

## Parallel Example: User Story 3

```text
Task: "T035 [P] [US3] Add stale queued job integration test using worker runtime path in tests/integration/postcard_ai_worker_test.go"
Task: "T037 [P] [US3] Add unchanged media reuse integration test proving no duplicate image provider call in tests/integration/postcard_ai_worker_test.go"
Task: "T038 [P] [US3] Add malformed-output repair parser/provider test coverage in internal/modules/postcardai/infra/ai/openai_responses_client_test.go"
```

## Parallel Example: User Story 4

```text
Task: "T043 [P] [US4] Add privacy test for worker logs excluding secrets and raw content in tests/unit/postcardai/privacy_test.go"
Task: "T044 [P] [US4] Add Dockerfile build contract test or documented static check for both binaries in tests/integration/full_stack_verification_test.go"
Task: "T045 [P] [US4] Add Compose service shape assertion for worker command, dependencies, environment, and no ports in tests/integration/full_stack_verification_test.go"
```

## Implementation Strategy

### MVP First

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 only.
3. Validate that queued jobs are consumed and provider success parsing works.
4. Stop and review before deployment-facing work.

### Incremental Delivery

1. US1 delivers the actual worker consumption path.
2. US2 protects public API compatibility.
3. US3 hardens correctness for stale/retry/reuse/provider-malformed paths.
4. US4 makes the feature deployable and observable in local Compose.

### Final Validation

1. Run focused worker/parser/contract/integration tests from `quickstart.md`.
2. Run full offline-safe Go tests.
3. Confirm local Compose starts API plus worker and `/health` remains API-only.
