# Tasks: Chronote Backend Contract-Preserving Refactor

**Input**: Design documents from `/specs/refactor/all/`
**Prerequisites**: [plan.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/plan.md), [spec.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/spec.md), [research.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/research.md), [data-model.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/data-model.md), [http-api.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/contracts/http-api.md)

**Tests**: Tests are REQUIRED by the refactor constitution. Include unit tests for domain and application rules, integration tests for critical request flows, and contract tests whenever request, response, or error schemas change.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Entrypoint**: `refactor/cmd/api/`
- **Platform wiring**: `refactor/internal/platform/`
- **Shared technical utilities**: `refactor/internal/shared/`
- **Business modules**: `refactor/internal/modules/<module>/{domain,app,infra,http}`
- **Migrations**: `refactor/migrations/`
- **Tests**: `refactor/tests/{unit,integration,contract}/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize the isolated replacement workspace and planning-aligned directory structure.

- [ ] T001 Create the replacement workspace directories under `refactor/cmd/api`, `refactor/internal/platform`, `refactor/internal/shared`, `refactor/internal/modules`, `refactor/migrations`, and `refactor/tests`
- [ ] T002 Initialize the Go module and baseline dependencies in `refactor/go.mod`
- [ ] T003 [P] Create application bootstrap skeleton in `refactor/cmd/api/main.go`, `refactor/internal/platform/app/app.go`, and `refactor/internal/platform/http/server.go`
- [ ] T004 [P] Create router bootstrap skeleton in `refactor/internal/platform/http/router.go` and `refactor/tests/integration/test_app.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Build shared primitives and platform adapters that block all user stories.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T005 Create shared response envelope helpers in `refactor/internal/shared/response/envelope.go`
- [ ] T006 [P] Create typed application errors and HTTP mapping in `refactor/internal/shared/errs/errors.go` and `refactor/internal/shared/errs/http_mapper.go`
- [ ] T007 [P] Create shared pagination helpers in `refactor/internal/shared/pagination/pagination.go`
- [ ] T008 Create shared tests for envelope and HTTP error mapping in `refactor/internal/shared/response/envelope_test.go` and `refactor/internal/shared/errs/http_mapper_test.go`
- [ ] T009 Create configuration loading in `refactor/internal/platform/config/config.go`
- [ ] T010 [P] Create Postgres and Redis client wiring in `refactor/internal/platform/db/postgres.go` and `refactor/internal/platform/redis/client.go`
- [ ] T011 [P] Create S3, JWT, and password platform services in `refactor/internal/platform/s3/client.go`, `refactor/internal/platform/auth/jwt_service.go`, and `refactor/internal/platform/auth/password_service.go`
- [ ] T012 Create platform adapter tests in `refactor/internal/platform/auth/jwt_service_test.go` and `refactor/internal/platform/auth/password_service_test.go`

**Checkpoint**: Foundation ready. User story implementation can now begin.

---

## Phase 3: User Story 1 - Preserve Existing Client Flows (Priority: P1) 🎯 MVP

**Goal**: Preserve health, user, and auth endpoint behavior so existing clients can keep using the backend without changing request paths, payload handling, or response handling.

**Independent Test**: Run health and user/auth contract flows against the replacement router and confirm compatible endpoint paths, methods, envelopes, status codes, and key success/failure messages.

### Tests for User Story 1

- [ ] T013 [P] [US1] Create health contract tests in `refactor/tests/contract/health_contract_test.go`
- [ ] T014 [P] [US1] Create health unit tests in `refactor/internal/modules/health/app/service_test.go`
- [ ] T015 [P] [US1] Create user/auth contract tests in `refactor/tests/contract/user_auth_contract_test.go`
- [ ] T016 [P] [US1] Create user and auth unit tests in `refactor/internal/modules/users/app/service_test.go` and `refactor/internal/modules/auth/app/service_test.go`
- [ ] T017 [P] [US1] Create user/auth integration flow tests in `refactor/tests/integration/user_auth_flow_test.go`

### Implementation for User Story 1

- [ ] T018 [P] [US1] Implement health application service and dependency interfaces in `refactor/internal/modules/health/app/service.go`
- [ ] T019 [US1] Implement health handlers and route registration in `refactor/internal/modules/health/http/handler.go` and `refactor/internal/modules/health/http/routes.go`
- [ ] T020 [P] [US1] Implement user domain model and repository contracts in `refactor/internal/modules/users/domain/user.go` and `refactor/internal/modules/users/app/repository.go`
- [ ] T021 [P] [US1] Implement auth service contracts and blacklist abstraction in `refactor/internal/modules/auth/app/service.go` and `refactor/internal/modules/auth/app/token_blacklist.go`
- [ ] T022 [US1] Implement user application service logic in `refactor/internal/modules/users/app/service.go`
- [ ] T023 [US1] Implement user persistence and token blacklist adapters in `refactor/internal/modules/users/infra/gorm_repository.go` and `refactor/internal/modules/auth/infra/redis_blacklist.go`
- [ ] T024 [P] [US1] Implement user and auth DTOs and handlers in `refactor/internal/modules/users/http/dto.go`, `refactor/internal/modules/users/http/handler.go`, `refactor/internal/modules/auth/http/handler.go`, and `refactor/internal/modules/auth/http/middleware.go`
- [ ] T025 [US1] Register `/health` and `/user/*` routes and wire dependencies in `refactor/internal/modules/users/http/routes.go`, `refactor/internal/modules/auth/http/routes.go`, and `refactor/internal/platform/http/router.go`
- [ ] T026 [US1] Preserve health, auth, and user contract messages through HTTP mapping and shared envelopes in `refactor/internal/shared/errs/http_mapper.go` and `refactor/internal/shared/response/envelope.go`

**Checkpoint**: User Story 1 should be functional and testable as the MVP.

---

## Phase 4: User Story 2 - Preserve Postcard and Media Behavior (Priority: P2)

**Goal**: Preserve postcard CRUD, visibility, media grouping, ownership, and ordering behavior for authenticated and anonymous Chronote users.

**Independent Test**: Exercise postcard and media create/read/update/delete/reorder flows against the replacement backend and confirm visibility, ownership, media-group, and ordering rules match the current contract.

### Tests for User Story 2

- [ ] T027 [P] [US2] Create postcard contract tests in `refactor/tests/contract/postcard_contract_test.go`
- [ ] T028 [P] [US2] Create media contract tests in `refactor/tests/contract/media_contract_test.go`
- [ ] T029 [P] [US2] Create postcard unit tests in `refactor/internal/modules/postcards/app/service_test.go`
- [ ] T030 [P] [US2] Create media unit tests in `refactor/internal/modules/media/app/service_test.go`
- [ ] T031 [P] [US2] Create postcard/media integration flow tests in `refactor/tests/integration/postcards_media_flow_test.go`

### Implementation for User Story 2

- [ ] T032 [P] [US2] Implement postcard domain model and repository contracts in `refactor/internal/modules/postcards/domain/postcard.go` and `refactor/internal/modules/postcards/app/repository.go`
- [ ] T033 [P] [US2] Implement postcard policy and application services in `refactor/internal/modules/postcards/app/policy.go` and `refactor/internal/modules/postcards/app/service.go`
- [ ] T034 [P] [US2] Implement postcard persistence adapter in `refactor/internal/modules/postcards/infra/gorm_repository.go`
- [ ] T035 [US2] Implement postcard DTOs, handlers, and routes in `refactor/internal/modules/postcards/http/dto.go`, `refactor/internal/modules/postcards/http/handler.go`, and `refactor/internal/modules/postcards/http/routes.go`
- [ ] T036 [P] [US2] Implement media domain model and repository/storage contracts in `refactor/internal/modules/media/domain/media.go`, `refactor/internal/modules/media/app/repository.go`, `refactor/internal/modules/media/app/storage.go`, and `refactor/internal/modules/media/app/image_processor.go`
- [ ] T037 [US2] Implement media application service and ordering rules in `refactor/internal/modules/media/app/service.go`
- [ ] T038 [P] [US2] Implement media persistence and storage adapters in `refactor/internal/modules/media/infra/gorm_repository.go`, `refactor/internal/modules/media/infra/s3_storage.go`, and `refactor/internal/modules/media/infra/image_processor.go`
- [ ] T039 [US2] Implement media DTOs, handlers, and routes in `refactor/internal/modules/media/http/dto.go`, `refactor/internal/modules/media/http/handler.go`, and `refactor/internal/modules/media/http/routes.go`
- [ ] T040 [US2] Wire `/v1/postcards*` and media routes with optional-auth and owner-only policies in `refactor/internal/platform/http/router.go`

**Checkpoint**: User Stories 1 and 2 should now both work independently.

---

## Phase 5: User Story 3 - Replace the Legacy Backend Safely (Priority: P3)

**Goal**: Prove the refactor can become the new backend source of truth while preserving supported data semantics, dependency health semantics, and cutover confidence.

**Independent Test**: Run the replacement backend against supported data fixtures and the full contract/integration suite, validating health degradation behavior and route compatibility before cutover.

### Tests for User Story 3

- [ ] T041 [P] [US3] Create cutover compatibility integration tests in `refactor/tests/integration/cutover_compatibility_test.go`
- [ ] T042 [P] [US3] Create supported-data fixture coverage in `refactor/tests/integration/fixtures.go` and `refactor/tests/integration/assertions.go`
- [ ] T043 [P] [US3] Create full-suite verification task runner coverage in `refactor/tests/integration/full_stack_verification_test.go`

### Implementation for User Story 3

- [ ] T044 [P] [US3] Implement shared integration test app bootstrap in `refactor/tests/integration/test_app.go`
- [ ] T045 [US3] Add compatibility fixture loading and supported-data assertions in `refactor/tests/integration/fixtures.go` and `refactor/tests/integration/assertions.go`
- [ ] T046 [US3] Create migration placeholder and schema-compatibility notes in `refactor/migrations/README.md`
- [ ] T047 [US3] Update deployment cutover entrypoints for the replacement app in `Dockerfile`, `docker-compose.yml`, and `docker-compose.dev.yml`
- [ ] T048 [US3] Document approved compatibility exceptions and cutover readiness notes in `specs/refactor/all/contracts/http-api.md` and `specs/refactor/all/quickstart.md`

**Checkpoint**: All user stories should now be independently functional and cutover-ready.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final hardening, verification, and documentation updates across all stories.

- [ ] T049 [P] Run the full Go test suite from `refactor/` and record verification notes in `specs/refactor/all/quickstart.md`
- [ ] T050 [P] Review API message/status mismatches and update compatibility documentation in `specs/refactor/all/contracts/http-api.md`
- [ ] T051 Validate branch/workflow helper behavior for implementation commands in `.specify/scripts/bash/common.sh` and `.specify/scripts/bash/tests/common-branch-resolution-test.sh`
- [ ] T052 [P] Update agent guidance and developer workflow notes in `AGENTS.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies, can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion and blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational completion.
- **User Story 2 (Phase 4)**: Depends on Foundational completion and should build on shared auth/platform work from US1 for optional auth and ownership checks.
- **User Story 3 (Phase 5)**: Depends on completion of US1 and US2 because cutover readiness requires end-to-end route coverage.
- **Polish (Phase 6)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Starts after Foundational and delivers the MVP.
- **User Story 2 (P2)**: Starts after Foundational, but practical execution follows US1 because postcards/media reuse auth and shared primitives.
- **User Story 3 (P3)**: Starts after US1 and US2 because cutover validation depends on all route groups existing.

### Within Each User Story

- Tests MUST be written and fail before implementation.
- Domain and repository contracts come before application orchestration.
- Application orchestration comes before handlers and routes.
- Infrastructure adapters come before final router wiring and end-to-end verification.

### Parallel Opportunities

- Phase 1 tasks marked `[P]` can run in parallel after directory and module initialization.
- Phase 2 shared utilities and platform adapters marked `[P]` can run in parallel by file group.
- Within each user story, contract tests, unit tests, and integration tests marked `[P]` can be authored in parallel.
- Domain models and repository contracts for postcards and media can be built in parallel before service wiring.

---

## Parallel Example: User Story 1

```bash
# Write story tests in parallel
Task: "Create health contract tests in refactor/tests/contract/health_contract_test.go"
Task: "Create user/auth contract tests in refactor/tests/contract/user_auth_contract_test.go"
Task: "Create user and auth unit tests in refactor/internal/modules/users/app/service_test.go and refactor/internal/modules/auth/app/service_test.go"

# Build story internals in parallel after tests exist
Task: "Implement health application service in refactor/internal/modules/health/app/service.go"
Task: "Implement user domain model and repository contracts in refactor/internal/modules/users/domain/user.go and refactor/internal/modules/users/app/repository.go"
Task: "Implement auth service contracts and blacklist abstraction in refactor/internal/modules/auth/app/service.go and refactor/internal/modules/auth/app/token_blacklist.go"
```

---

## Parallel Example: User Story 2

```bash
# Write story tests in parallel
Task: "Create postcard contract tests in refactor/tests/contract/postcard_contract_test.go"
Task: "Create media contract tests in refactor/tests/contract/media_contract_test.go"
Task: "Create postcard/media integration flow tests in refactor/tests/integration/postcards_media_flow_test.go"

# Build domain and infra in parallel after tests exist
Task: "Implement postcard domain model and repository contracts in refactor/internal/modules/postcards/domain/postcard.go and refactor/internal/modules/postcards/app/repository.go"
Task: "Implement media domain model and repository/storage contracts in refactor/internal/modules/media/domain/media.go, refactor/internal/modules/media/app/repository.go, refactor/internal/modules/media/app/storage.go, and refactor/internal/modules/media/app/image_processor.go"
Task: "Implement media persistence and storage adapters in refactor/internal/modules/media/infra/gorm_repository.go, refactor/internal/modules/media/infra/s3_storage.go, and refactor/internal/modules/media/infra/image_processor.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1.
4. Validate health and user/auth flows independently with contract and integration tests.
5. Use this as the first demonstrable compatibility milestone.

### Incremental Delivery

1. Finish Setup and Foundational work once.
2. Deliver User Story 1 as the first working replacement slice.
3. Add User Story 2 and validate postcard/media behavior independently.
4. Add User Story 3 to prove cutover readiness and full compatibility.
5. Finish with Polish tasks and full-suite verification.

### Parallel Team Strategy

1. One engineer handles shared/platform primitives while another prepares failing contract tests.
2. After Foundational is complete:
   - Engineer A: Health plus users/auth
   - Engineer B: Postcards
   - Engineer C: Media and storage adapters once postcard contracts stabilize
3. After US1 and US2 land, one engineer focuses on cutover/integration hardening while others fix compatibility gaps.
