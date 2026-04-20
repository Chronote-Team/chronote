---

description: "Task list template for feature implementation"
---

# Tasks: [FEATURE NAME]

**Input**: Design documents from `/specs/[###-feature-name]/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are REQUIRED by the refactor constitution. Include unit tests
for domain and application rules, integration tests for critical request flows,
and contract tests whenever request, response, or error schemas change.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Entrypoint**: `cmd/api/`
- **Platform wiring**: `internal/platform/`
- **Shared technical utilities**: `internal/shared/`
- **Business modules**: `internal/modules/<module>/{domain,app,infra,http}`
- **Migrations**: `migrations/`
- **Tests**: `tests/{unit,integration,contract}/`

<!-- 
  ============================================================================
  IMPORTANT: The tasks below are SAMPLE TASKS for illustration purposes only.
  
  The /speckit.tasks command MUST replace these with actual tasks based on:
  - User stories from spec.md (with their priorities P1, P2, P3...)
  - Feature requirements from plan.md
  - Entities from data-model.md
  - Endpoints from contracts/
  
  Tasks MUST be organized by user story so each story can be:
  - Implemented independently
  - Tested independently
  - Delivered as an MVP increment
  
  DO NOT keep these sample tasks in the generated tasks.md file.
  ============================================================================
-->

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and boundary-preserving structure

- [ ] T001 Create project structure per implementation plan and constitution
- [ ] T002 Initialize [language] project with [framework] dependencies
- [ ] T003 [P] Configure linting and formatting tools

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

Examples of foundational tasks (adjust based on your project):

- [ ] T004 Setup database schema and migrations framework
- [ ] T005 [P] Scaffold `internal/platform` dependency wiring and configuration
- [ ] T006 [P] Setup API routing and middleware structure in module `http`
      packages
- [ ] T007 Create shared domain contracts and typed error patterns needed across
      stories
- [ ] T008 Configure logging, error translation, and response shaping
- [ ] T009 Setup environment configuration management

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - [Title] (Priority: P1) 🎯 MVP

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T010 [P] [US1] Unit tests for domain/app rules in tests/unit/[name]_test.go
- [ ] T011 [P] [US1] Contract test for [endpoint] in tests/contract/[name]_test.go
- [ ] T012 [P] [US1] Integration test for [user journey] in tests/integration/[name]_test.go

### Implementation for User Story 1

- [ ] T013 [P] [US1] Define or update entities and invariants in
      internal/modules/[module]/domain/
- [ ] T014 [P] [US1] Implement use cases in internal/modules/[module]/app/
- [ ] T015 [P] [US1] Implement repositories or gateways in
      internal/modules/[module]/infra/
- [ ] T016 [US1] Implement handlers, DTOs, and route registration in
      internal/modules/[module]/http/
- [ ] T017 [US1] Wire dependencies in internal/platform/ and verify no globals
      leak into business logic
- [ ] T018 [US1] Add validation, normalization, and typed error mapping for the
      story contract

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - [Title] (Priority: P2)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 2

- [ ] T019 [P] [US2] Unit tests for domain/app rules in tests/unit/[name]_test.go
- [ ] T020 [P] [US2] Contract test for [endpoint] in tests/contract/[name]_test.go
- [ ] T021 [P] [US2] Integration test for [user journey] in tests/integration/[name]_test.go

### Implementation for User Story 2

- [ ] T022 [P] [US2] Define or update entities and invariants in
      internal/modules/[module]/domain/
- [ ] T023 [US2] Implement use cases in internal/modules/[module]/app/
- [ ] T024 [US2] Implement repositories, gateways, or adapters in
      internal/modules/[module]/infra/
- [ ] T025 [US2] Implement handlers, DTOs, and route registration in
      internal/modules/[module]/http/
- [ ] T026 [US2] Integrate with shared contracts or other module interfaces
      without transport-layer coupling

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - [Title] (Priority: P3)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 3

- [ ] T027 [P] [US3] Unit tests for domain/app rules in tests/unit/[name]_test.go
- [ ] T028 [P] [US3] Contract test for [endpoint] in tests/contract/[name]_test.go
- [ ] T029 [P] [US3] Integration test for [user journey] in tests/integration/[name]_test.go

### Implementation for User Story 3

- [ ] T030 [P] [US3] Define or update entities and invariants in
      internal/modules/[module]/domain/
- [ ] T031 [US3] Implement use cases in internal/modules/[module]/app/
- [ ] T032 [US3] Implement repositories, gateways, or adapters in
      internal/modules/[module]/infra/
- [ ] T033 [US3] Implement handlers, DTOs, and route registration in
      internal/modules/[module]/http/

**Checkpoint**: All user stories should now be independently functional

---

[Add more user story phases as needed, following the same pattern]

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] TXXX [P] Documentation updates in docs/
- [ ] TXXX Code cleanup and refactoring
- [ ] TXXX Performance optimization across all stories
- [ ] TXXX [P] Additional unit and integration test hardening in tests/
- [ ] TXXX Security hardening
- [ ] TXXX Validate quickstart.md or equivalent verification flow

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US1 but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - May integrate with US1/US2 but should be independently testable

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Domain invariants before application orchestration
- Application orchestration before transport handlers
- Infrastructure adapters before final platform wiring
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Models within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (if tests requested):
Task: "Contract test for [endpoint] in tests/contract/test_[name].py"
Task: "Integration test for [user journey] in tests/integration/test_[name].py"

# Launch all models for User Story 1 together:
Task: "Create [Entity1] model in src/models/[entity1].py"
Task: "Create [Entity2] model in src/models/[entity2].py"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, persistence model leakage, same file conflicts,
  transport-layer cross-module calls, and cross-story dependencies that break
  independence
