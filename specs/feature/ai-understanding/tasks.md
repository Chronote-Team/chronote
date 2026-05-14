# Tasks: Postcard AI Understanding

**Input**: Design documents from `/specs/feature/ai-understanding/`
**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/internal-workflow.md](contracts/internal-workflow.md), [quickstart.md](quickstart.md)

**Tests**: Tests are REQUIRED by the refactor constitution. Include unit tests for domain and application rules, integration tests for worker/persistence behavior, privacy tests for safe logging, and contract tests proving public API non-change.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested as an independent increment after shared foundation work.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files and does not depend on incomplete tasks.
- **[Story]**: Which user story this task belongs to, only present for user story phases.
- Every task includes exact file paths.

## Path Conventions

- **Entrypoint**: `cmd/api/`
- **Platform wiring**: `internal/platform/`
- **Shared technical utilities**: `internal/shared/`
- **Business modules**: `internal/modules/<module>/{domain,app,infra,http}`
- **Migrations**: `migrations/`
- **Tests**: `tests/{unit,integration,contract}/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the internal module skeleton and configuration placeholders without changing public API behavior.

- [X] T001 Create postcard AI module directories in `internal/modules/postcardai/domain/`, `internal/modules/postcardai/app/`, `internal/modules/postcardai/infra/`, and `internal/modules/postcardai/infra/ai/`
- [X] T002 [P] Create prompt directory and placeholder versioned prompt files in `internal/modules/postcardai/infra/ai/prompts/image_understanding_v1.tmpl` and `internal/modules/postcardai/infra/ai/prompts/postcard_understanding_v1.tmpl`
- [X] T003 [P] Create unit test directory for postcard AI module tests in `tests/unit/postcardai/`
- [X] T004 [P] Create integration test placeholder for worker flow in `tests/integration/postcard_ai_worker_test.go`
- [X] T005 [P] Create privacy test placeholder for safe logging in `tests/unit/postcardai/privacy_test.go`
- [X] T006 Add AI runtime configuration fields and defaults in `internal/platform/config/config.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define shared domain contracts, schema shape, and persistence prerequisites required before user stories can be implemented.

**Critical**: No user story work should begin until this phase is complete.

- [X] T007 Define analysis status constants and transition helpers in `internal/modules/postcardai/domain/status.go`
- [X] T008 [P] Define version key value objects for postcard, media, prompt, schema, and model versions in `internal/modules/postcardai/domain/version.go`
- [X] T009 [P] Define safe provider error categories and retry classifications in `internal/modules/postcardai/domain/errors.go`
- [X] T010 [P] Define media and postcard structured result domain types in `internal/modules/postcardai/domain/analysis.go`
- [X] T011 Define analysis job domain type and invariants in `internal/modules/postcardai/domain/job.go`
- [X] T012 Add database migration for AI job and result tables in `migrations/202605130001_create_ai_analysis_tables.sql`
- [X] T013 Add schema definitions and validation entry points in `internal/modules/postcardai/infra/ai/schemas.go`
- [X] T014 Define application interfaces for repositories, source lookups, locks, storage presigning, clock, and provider calls in `internal/modules/postcardai/app/ports.go`
- [X] T015 Add foundation tests for status transitions, version-key equality, and schema validation in `tests/unit/postcardai/domain_foundation_test.go`

**Checkpoint**: Foundation ready. User story implementation can now begin.

---

## Phase 3: User Story 1 - Analyze New Or Changed Postcards Internally (Priority: P1) - MVP

**Goal**: Detect eligible postcard create/update/media-change events, enqueue internal analysis work, process text and attached images, store image-level and postcard-level results, and keep normal client app behavior unchanged.

**Independent Test**: Create or update a postcard with text and attached images, run the internal worker with fake provider/storage dependencies, and confirm stored image/postcard understanding exists while public postcard responses remain unchanged.

### Tests for User Story 1

- [X] T016 [P] [US1] Add unit tests for enqueue eligibility and idempotent active-job creation in `tests/unit/postcardai/enqueue_test.go`
- [X] T017 [P] [US1] Add unit tests for worker success flow with fake postcard, media, storage, and AI provider ports in `tests/unit/postcardai/worker_success_test.go`
- [X] T018 [P] [US1] Add integration test for PostgreSQL-backed job/result persistence in `tests/integration/postcard_ai_worker_test.go`
- [X] T019 [P] [US1] Add contract tests proving normal postcard create/read responses do not expose analysis fields in `tests/contract/postcard_contract_test.go`
- [X] T020 [P] [US1] Add contract tests proving no client-facing analysis endpoints are registered in `tests/contract/postcard_ai_contract_test.go`

### Implementation for User Story 1

- [X] T021 [P] [US1] Implement job repository and result repository interfaces in `internal/modules/postcardai/app/repositories.go`
- [X] T022 [P] [US1] Implement postcard and media source lookup interfaces in `internal/modules/postcardai/app/sources.go`
- [X] T023 [P] [US1] Implement storage presign and AI provider app interfaces in `internal/modules/postcardai/app/provider.go`
- [X] T024 [US1] Implement `EnqueuePostcardAnalysis` use case with eligibility and durable pending-job creation in `internal/modules/postcardai/app/service.go`
- [X] T025 [US1] Implement worker orchestration for loading postcard text, media, metadata, and image inputs in `internal/modules/postcardai/app/worker.go`
- [X] T026 [US1] Implement image understanding step and image result validation in `internal/modules/postcardai/app/worker.go`
- [X] T027 [US1] Implement postcard understanding step and final result validation in `internal/modules/postcardai/app/worker.go`
- [X] T028 [US1] Implement GORM models and mappers for jobs and analysis results in `internal/modules/postcardai/infra/gorm_models.go`
- [X] T029 [US1] Implement PostgreSQL/GORM repository methods for enqueue, claim, store media result, store postcard result, and complete job in `internal/modules/postcardai/infra/gorm_repository.go`
- [X] T030 [US1] Implement OpenAI Responses-compatible request/response mapping behind `AIClient` in `internal/modules/postcardai/infra/ai/openai_responses_client.go`
- [X] T031 [US1] Implement prompt loading and prompt version metadata in `internal/modules/postcardai/infra/ai/prompts.go`
- [X] T032 [US1] Implement worker dependency wiring with fake-provider support in `internal/platform/app/app.go`
- [X] T033 [US1] Implement AI provider configuration parsing and disabled-by-default local/test behavior in `internal/platform/config/config.go`
- [X] T034 [US1] Add enqueue hook after successful postcard creation/update in `internal/modules/postcards/app/service.go`
- [X] T035 [US1] Add enqueue hook after successful media upload/delete/reorder changes in `internal/modules/media/app/service.go`
- [X] T036 [US1] Wire storage presigning adapter for postcard AI worker through existing S3 platform storage in `internal/platform/s3/client.go`
- [X] T037 [US1] Verify User Story 1 with `env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/modules/postcardai/... ./tests/unit/postcardai/... ./tests/contract/... -run 'Postcard|Analysis|Worker|Enqueue' -v`

**Checkpoint**: User Story 1 is independently functional and testable as the MVP.

---

## Phase 4: User Story 2 - Reuse Understanding For Unchanged Media (Priority: P2)

**Goal**: Reuse successful image-level analysis for unchanged media when media, prompt, schema, and model versions match, while refreshing postcard-level understanding when postcard text changes.

**Independent Test**: Analyze a postcard once, run analysis again without changing media, and confirm media analysis is reused while postcard-level analysis can be regenerated for the updated postcard version.

### Tests for User Story 2

- [X] T038 [P] [US2] Add unit tests for media analysis cache selection by media/prompt/schema/model version key in `tests/unit/postcardai/media_cache_test.go`
- [X] T039 [P] [US2] Add unit tests proving mismatched prompt, schema, model, or media versions do not reuse analysis in `tests/unit/postcardai/media_cache_mismatch_test.go`
- [X] T040 [P] [US2] Add integration test for reuse on repeated worker runs with unchanged media in `tests/integration/postcard_ai_worker_test.go`

### Implementation for User Story 2

- [X] T041 [P] [US2] Add reusable media analysis lookup contract to `internal/modules/postcardai/app/repositories.go`
- [X] T042 [US2] Implement media analysis lookup by full version key in `internal/modules/postcardai/infra/gorm_repository.go`
- [X] T043 [US2] Add uniqueness enforcement and repository conflict handling for media analysis version keys in `internal/modules/postcardai/infra/gorm_repository.go`
- [X] T044 [US2] Update worker orchestration to skip provider image calls when reusable successful media analysis exists in `internal/modules/postcardai/app/worker.go`
- [X] T045 [US2] Update worker context assembly to combine reused media facts with changed postcard text in `internal/modules/postcardai/app/worker.go`
- [X] T046 [US2] Add version metadata propagation from prompt/schema/model definitions into stored results in `internal/modules/postcardai/app/worker.go`
- [X] T047 [US2] Add media version source mapping from media records into postcard AI source adapter in `internal/modules/media/app/service.go`
- [X] T048 [US2] Add postcard version source mapping from postcard records into postcard AI source adapter in `internal/modules/postcards/app/service.go`
- [X] T049 [US2] Verify User Story 2 with `env GOCACHE=/tmp/go-build GOPROXY=off go test ./tests/unit/postcardai/... ./tests/integration/... -run 'MediaCache|Reuse|Version' -v`

**Checkpoint**: User Stories 1 and 2 work independently and together.

---

## Phase 5: User Story 3 - Preserve Uncertainty And Failure States (Priority: P3)

**Goal**: Store failed, unavailable, stale, partial, and low-confidence outcomes with safe error categories, confidence/uncertainty metadata, and no leakage to normal client app users.

**Independent Test**: Force malformed, refused, unavailable, partial, stale, and low-confidence provider outcomes and confirm stored records capture status, confidence, uncertainty, and version context while public responses remain unchanged.

### Tests for User Story 3

- [X] T050 [P] [US3] Add unit tests for one schema-repair retry and malformed-output failure handling in `tests/unit/postcardai/schema_repair_test.go`
- [X] T051 [P] [US3] Add unit tests for partial image failure producing partial postcard understanding with uncertainty in `tests/unit/postcardai/partial_result_test.go`
- [X] T052 [P] [US3] Add unit tests for stale postcard/media version prevention in `tests/unit/postcardai/stale_version_test.go`
- [X] T053 [P] [US3] Add privacy tests proving logs exclude raw text, image bytes, API keys, signed URLs, and provider payloads in `tests/unit/postcardai/privacy_test.go`
- [X] T054 [P] [US3] Add integration tests for worker interruption recovery and failed/unavailable job inspection in `tests/integration/postcard_ai_worker_test.go`

### Implementation for User Story 3

- [X] T055 [P] [US3] Add confidence and uncertainty helpers to result domain types in `internal/modules/postcardai/domain/analysis.go`
- [X] T056 [P] [US3] Add safe logging metadata builder in `internal/modules/postcardai/app/logging.go`
- [X] T057 [US3] Implement schema-repair retry decision logic in `internal/modules/postcardai/app/worker.go`
- [X] T058 [US3] Implement provider refusal, timeout, unavailable, and permanent input failure mapping in `internal/modules/postcardai/infra/ai/openai_responses_client.go`
- [X] T059 [US3] Implement partial-result handling when some media analysis fails but postcard analysis can continue in `internal/modules/postcardai/app/worker.go`
- [X] T060 [US3] Implement stale postcard/media version checks before storing media and postcard analysis in `internal/modules/postcardai/app/worker.go`
- [X] T061 [US3] Implement failed, unavailable, and stale status persistence in `internal/modules/postcardai/infra/gorm_repository.go`
- [X] T062 [US3] Implement private retry and backfill application methods in `internal/modules/postcardai/app/service.go`
- [X] T063 [US3] Verify User Story 3 with `env GOCACHE=/tmp/go-build GOPROXY=off go test ./tests/unit/postcardai/... ./tests/integration/... -run 'Repair|Partial|Stale|Privacy|Recovery|Retry' -v`

**Checkpoint**: All user stories are independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Harden documentation, verification, and cross-feature safety after all selected user stories are complete.

- [X] T064 [P] Update AI understanding implementation notes in `docs/postcard-ai-understanding-final.md`
- [X] T065 [P] Add migration documentation for AI analysis tables in `migrations/README.md`
- [X] T066 [P] Add quickstart verification notes for fake provider configuration in `specs/feature/ai-understanding/quickstart.md`
- [X] T067 Run full offline-safe test suite with `env GOCACHE=/tmp/go-build GOPROXY=off go test ./... -v`
- [X] T068 Run Go formatting on changed Go files with `gofmt` covering `internal/modules/postcardai/`, `internal/modules/postcards/app/service.go`, `internal/modules/media/app/service.go`, `internal/platform/app/app.go`, `internal/platform/config/config.go`, and `internal/platform/s3/client.go`
- [X] T069 Review public API diff to confirm no analysis endpoints or response fields were introduced in `tests/contract/postcard_ai_contract_test.go`
- [X] T070 Review logging and provider adapter code for privacy guardrails in `internal/modules/postcardai/app/logging.go` and `internal/modules/postcardai/infra/ai/openai_responses_client.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup completion and blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational; delivers the MVP workflow.
- **User Story 2 (Phase 4)**: Depends on Foundational and can be developed after US1 contracts are understood, but its cache rules are independently testable with fakes.
- **User Story 3 (Phase 5)**: Depends on Foundational and can be developed after US1 worker shape exists, but its failure-state logic is independently testable with fakes.
- **Polish (Phase 6)**: Depends on all selected user stories being complete.

### User Story Dependencies

- **US1 - Analyze New Or Changed Postcards Internally**: Required MVP; no dependency on US2 or US3.
- **US2 - Reuse Understanding For Unchanged Media**: Builds on the worker/repository shape from US1 but can be verified independently using fake existing media analysis.
- **US3 - Preserve Uncertainty And Failure States**: Builds on worker/provider shape from US1 but can be verified independently using fake provider failures.

### Within Each User Story

- Write story tests first and confirm they fail before implementation.
- Implement domain invariants before application orchestration.
- Implement application orchestration before infrastructure adapters where feasible.
- Keep public HTTP handlers free of AI workflow logic.
- Complete each story's verification command before moving to the next story.

### Parallel Opportunities

- Setup tasks T002-T005 can run in parallel.
- Foundational tasks T008-T010 can run in parallel after T007 starts.
- US1 tests T016-T020 can run in parallel.
- US1 interface tasks T021-T023 can run in parallel.
- US2 tests T038-T040 can run in parallel.
- US3 tests T050-T054 can run in parallel.
- Polish documentation tasks T064-T066 can run in parallel.

---

## Parallel Example: User Story 1

```bash
Task: "T016 [US1] Add unit tests for enqueue eligibility and idempotent active-job creation in tests/unit/postcardai/enqueue_test.go"
Task: "T017 [US1] Add unit tests for worker success flow with fake postcard, media, storage, and AI provider ports in tests/unit/postcardai/worker_success_test.go"
Task: "T018 [US1] Add integration test for PostgreSQL-backed job/result persistence in tests/integration/postcard_ai_worker_test.go"
Task: "T019 [US1] Add contract tests proving normal postcard create/read responses do not expose analysis fields in tests/contract/postcard_contract_test.go"
Task: "T020 [US1] Add contract tests proving no client-facing analysis endpoints are registered in tests/contract/postcard_ai_contract_test.go"
```

## Parallel Example: User Story 2

```bash
Task: "T038 [US2] Add unit tests for media analysis cache selection by media/prompt/schema/model version key in tests/unit/postcardai/media_cache_test.go"
Task: "T039 [US2] Add unit tests proving mismatched prompt, schema, model, or media versions do not reuse analysis in tests/unit/postcardai/media_cache_mismatch_test.go"
Task: "T040 [US2] Add integration test for reuse on repeated worker runs with unchanged media in tests/integration/postcard_ai_worker_test.go"
```

## Parallel Example: User Story 3

```bash
Task: "T050 [US3] Add unit tests for one schema-repair retry and malformed-output failure handling in tests/unit/postcardai/schema_repair_test.go"
Task: "T051 [US3] Add unit tests for partial image failure producing partial postcard understanding with uncertainty in tests/unit/postcardai/partial_result_test.go"
Task: "T052 [US3] Add unit tests for stale postcard/media version prevention in tests/unit/postcardai/stale_version_test.go"
Task: "T053 [US3] Add privacy tests proving logs exclude raw text, image bytes, API keys, signed URLs, and provider payloads in tests/unit/postcardai/privacy_test.go"
Task: "T054 [US3] Add integration tests for worker interruption recovery and failed/unavailable job inspection in tests/integration/postcard_ai_worker_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1.
4. Stop and validate with T037.
5. Confirm public postcard/media contract tests still pass.

### Incremental Delivery

1. Deliver US1 to create and store internal understanding without public API changes.
2. Deliver US2 to reduce repeated media analysis for unchanged media.
3. Deliver US3 to harden failure, uncertainty, stale-result, and privacy behavior.
4. Run Phase 6 polish and full verification.

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup and Foundational phases together.
2. Developer A implements US1 worker and enqueue path.
3. Developer B implements US2 media reuse after repository/version contracts exist.
4. Developer C implements US3 failure-state and privacy tests after worker/provider seams exist.
5. Team reconvenes for full contract, integration, and privacy verification.
