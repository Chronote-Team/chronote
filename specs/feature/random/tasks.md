# Tasks: Random Accessible Postcard

**Input**: Design documents from `/home/bowen/Coding/chronote-refactor/specs/feature/random/`  
**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/random-postcard.openapi.yaml](./contracts/random-postcard.openapi.yaml), [quickstart.md](./quickstart.md)

**Setup Note**: `.specify/scripts/bash/check-prerequisites.sh --json` was run from `/home/bowen/Coding/chronote-refactor` but rejected the active branch `feature/random`. Tasks are generated against the user-requested feature directory `/home/bowen/Coding/chronote-refactor/specs/feature/random`.

**Tests**: Tests are required by the Chronote refactor constitution and by the implementation plan. Write the US1 tests first and confirm they fail before implementation.

**Organization**: Tasks are grouped by user story so the P1 MVP can be implemented and tested independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel after its phase prerequisites are satisfied because it touches different files or is read-only.
- **[Story]**: User-story task label, required only inside user-story phases.
- Every task includes exact file paths.

## Phase 1: Setup (Shared Context)

**Purpose**: Confirm the feature contract, local source layout, and current route/test locations before editing.

- [X] T001 Review the random postcard endpoint source requirements in `/home/bowen/Coding/chronote-refactor/docs/v1-postcards-random.md` and `/home/bowen/Coding/chronote-refactor/specs/feature/random/spec.md`
- [X] T002 [P] Review the generated implementation plan and OpenAPI contract in `/home/bowen/Coding/chronote-refactor/specs/feature/random/plan.md` and `/home/bowen/Coding/chronote-refactor/specs/feature/random/contracts/random-postcard.openapi.yaml`
- [X] T003 [P] Inspect existing postcard route and handler patterns in `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/http/routes.go` and `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/http/handler.go`
- [X] T004 [P] Inspect existing postcard app and repository patterns in `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/app/repository.go`, `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/app/service.go`, and `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/infra/gorm_repository.go`

---

## Phase 2: Foundational (Blocking Prerequisite)

**Purpose**: Add the shared app repository contract required by all random postcard behavior.

**Critical**: Complete this phase before starting US1 implementation tasks.

- [X] T005 Add `FindRandomAccessible(userID uint) (*postcardsdomain.Postcard, error)` to the postcard repository interface in `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/app/repository.go`

**Checkpoint**: Repository contract is ready for test and implementation work.

---

## Phase 3: User Story 1 - Discover a Random Postcard (Priority: P1)

**Goal**: A main-page caller can request one random postcard that is accessible under the caller's anonymous or signed-in visibility rules.

**Independent Test**: Request `GET /v1/postcards/random` as anonymous and signed-in callers, then confirm exactly one accessible postcard is returned when eligible content exists and existing not-found behavior is returned when no eligible content exists.

### Tests for User Story 1

- [X] T006 [P] [US1] Add failing app tests for anonymous public-only selection, signed-in owned-private selection, and no-accessible not-found behavior in `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/app/service_test.go`
- [X] T007 [P] [US1] Add failing contract tests for anonymous success, signed-in owned-private success, no-accessible `404`, and `/random` route-order behavior in `/home/bowen/Coding/chronote-refactor/tests/contract/postcard_contract_test.go`

### Implementation for User Story 1

- [X] T008 [US1] Implement `Service.GetRandom(userID uint)` with repository error mapping, not-found mapping, and relation attachment in `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/app/service.go`
- [X] T009 [US1] Implement `memoryRepository.FindRandomAccessible(userID uint)` using the same public-or-owner access rules and returning a copied postcard in `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/app/service.go`
- [X] T010 [P] [US1] Implement `Repository.FindRandomAccessible(userID uint)` with random accessible selection and not-found handling in `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/infra/gorm_repository.go`
- [X] T011 [US1] Implement `Handler.GetRandomPostcard` with optional `userID`, `errs.MapHTTP`, success message `获取随机明信片成功`, and `newPostcardResponse` in `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/http/handler.go`
- [X] T012 [US1] Register `GET /random` before `GET /:id` in `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/http/routes.go`
- [X] T013 [US1] Run the focused postcard tests for `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/app/service_test.go` and `/home/bowen/Coding/chronote-refactor/tests/contract/postcard_contract_test.go`

**Checkpoint**: User Story 1 is fully functional and independently testable.

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Format, regression-test, and verify the implementation against the feature quickstart.

- [X] T014 Format modified Go files with `gofmt` for `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/app/repository.go`, `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/app/service.go`, `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/infra/gorm_repository.go`, `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/http/handler.go`, `/home/bowen/Coding/chronote-refactor/internal/modules/postcards/http/routes.go`, and `/home/bowen/Coding/chronote-refactor/tests/contract/postcard_contract_test.go`
- [X] T015 Run the offline-safe full test command from `/home/bowen/Coding/chronote-refactor/AGENTS.md` and verify `env GOCACHE=/tmp/go-build GOPROXY=off go test ./... -v`
- [X] T016 Validate the implementation against `/home/bowen/Coding/chronote-refactor/specs/feature/random/quickstart.md` and `/home/bowen/Coding/chronote-refactor/specs/feature/random/contracts/random-postcard.openapi.yaml`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1: Setup** has no dependencies.
- **Phase 2: Foundational** depends on Phase 1 and blocks all US1 implementation.
- **Phase 3: User Story 1** depends on Phase 2.
- **Phase 4: Polish** depends on User Story 1 implementation being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on other user stories. This is the complete MVP.

### Within User Story 1

- Write tests T006 and T007 first and confirm they fail.
- Complete app service and memory repository work T008-T009 before handler verification.
- Complete production repository work T010 before full suite verification.
- Complete handler and route work T011-T012 before contract tests can pass.
- Run focused tests T013 before polish and full-suite verification.

### Parallel Opportunities

- T002, T003, and T004 are read-only setup tasks and can run in parallel after T001.
- T006 and T007 can run in parallel because they touch separate test files.
- T010 can run in parallel with T008-T009 after T005 because it touches the infrastructure adapter only.

## Parallel Example: User Story 1

```bash
# After T005, write failing tests in parallel:
Task T006: "Add app tests in /home/bowen/Coding/chronote-refactor/internal/modules/postcards/app/service_test.go"
Task T007: "Add contract tests in /home/bowen/Coding/chronote-refactor/tests/contract/postcard_contract_test.go"

# After T005, implement independent adapter work in parallel with app service work:
Task T008/T009: "Update /home/bowen/Coding/chronote-refactor/internal/modules/postcards/app/service.go"
Task T010: "Update /home/bowen/Coding/chronote-refactor/internal/modules/postcards/infra/gorm_repository.go"
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 setup review.
2. Complete Phase 2 repository interface contract.
3. Write failing US1 app and contract tests.
4. Implement US1 app, repository, handler, and route changes.
5. Run focused tests and stop to validate the random postcard endpoint.

### Incremental Delivery

1. Add app repository contract.
2. Add tests that capture access rules and route contract.
3. Implement app and repository behavior.
4. Add HTTP handler and route registration.
5. Verify focused tests, then full offline-safe suite.

### Task Count Summary

- Total tasks: 16
- Setup tasks: 4
- Foundational tasks: 1
- User Story 1 tasks: 8
- Polish tasks: 3
- Parallel opportunities: 6 tasks marked `[P]`
