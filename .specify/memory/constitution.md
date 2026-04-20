<!--
Sync Impact Report
Version change: template -> 1.0.0
Modified principles:
- Template Principle 1 -> I. Vertical Slice Boundaries
- Template Principle 2 -> II. Stable API Contracts
- Template Principle 3 -> III. Dependency Isolation and Platform Wiring
- Template Principle 4 -> IV. Replacement-First Delivery
- Template Principle 5 -> V. Mandatory Testability and Critical Flow Coverage
Added sections:
- Architectural Constraints
- Delivery Workflow and Review Gates
Removed sections:
- None
Templates requiring updates:
- ✅ .specify/templates/plan-template.md
- ✅ .specify/templates/spec-template.md
- ✅ .specify/templates/tasks-template.md
- ✅ .specify/templates/agent-file-template.md
- ⚠ Pending by absence: .specify/templates/commands/*.md does not exist in this repository
Follow-up TODOs:
- None
-->
# Chronote Refactor Constitution

## Core Principles

### I. Vertical Slice Boundaries
Every business capability MUST live inside `internal/modules/<module>` and keep
transport, application orchestration, domain rules, and infrastructure concerns
separate. Handlers MUST handle HTTP concerns only, `app` MUST orchestrate use
cases, `domain` MUST own invariants and domain errors, and `infra` MUST implement
repositories or gateways. Business logic MUST NOT be added to Gin handlers,
GORM models, or storage adapters. This rule exists to stop controller/service
mixing and keep the refactor extensible.

### II. Stable API Contracts
Each endpoint MUST define a deterministic request schema, response schema, and
error schema before implementation. Validation and input normalization MUST be
explicit and centralized. Persistence models MUST NOT leak into public request
or response DTOs, and internal storage-only fields MUST remain private. Status
codes and error categories MUST stay consistent for equivalent failures. This
rule exists because API inconsistency and string-based error handling are known
sources of instability in the legacy code.

### III. Dependency Isolation and Platform Wiring
Domain code MUST NOT depend on Gin, GORM, Redis, S3, JWT libraries, or package
globals. Application code MAY depend only on domain types plus interfaces for
repositories, gateways, clock, and token services. Concrete implementations MUST
be wired in `internal/platform`. Cross-module collaboration MUST happen through
application interfaces or shared contracts, never through HTTP handlers or
direct access to another module's infrastructure. This rule exists to remove
hidden runtime coupling and make core rules testable.

### IV. Replacement-First Delivery
This repository is the only allowed write scope for the refactor, and work MUST
not modify the legacy codebase in `/home/bowen/Coding/chronote`. The refactor
MUST optimize for clean replacement of the old internal layout rather than
incremental coexistence with it. Backward compatibility is required at the API
contract and persisted data boundary when the new branch takes over; internal
package parity with legacy code is not required. Branch divergence is allowed
when it improves the target architecture. This rule exists because the project
goal is a rebuild-oriented replacement, not a dual-track maintenance effort.

### V. Mandatory Testability and Critical Flow Coverage
Core business rules MUST be unit-testable without real PostgreSQL, Redis, or
S3 dependencies. Every new or materially changed module MUST ship with tests for
its main behavior. Integration tests MUST cover critical request flows for auth,
postcards, media, and any new API surface that changes contracts or orchestration
across dependencies. Contract tests MUST be added when request, response, or
error schemas change. This rule exists because the refactor only succeeds if it
gains a reliable safety net that the legacy code lacks.

## Architectural Constraints

- Project identity: Chronote is a Go backend service for accounts,
  authentication, postcards, postcard media management, and dependency health
  checks.
- Runtime stack: Go, Gin, GORM, PostgreSQL, Redis, and S3-compatible object
  storage.
- Preferred root layout MUST follow this structure unless an approved plan
  records an equivalent boundary-preserving alternative:
  - `cmd/api` for the application entrypoint only.
  - `internal/platform` for configuration, runtime wiring, auth wiring, and HTTP
    bootstrap.
  - `internal/shared` for technical utilities with no business ownership.
  - `internal/modules/users`
  - `internal/modules/auth`
  - `internal/modules/postcards`
  - `internal/modules/media`
  - `internal/modules/health`
  - `migrations`
  - `tests`
  - `docs`
- Each business module SHOULD use `domain`, `app`, `infra`, and `http`
  subpackages whenever that module spans those concerns.
- Current route groups that MUST remain intentionally managed are `/health`,
  `/health/details`, `/user/*`, and `/v1/postcards/*`.
- Protected endpoints MUST use JWT-based auth. Optional auth MAY be used only
  for explicitly documented postcard read paths.
- Domain invariants that MUST be preserved unless a spec and plan explicitly
  amend them:
  - usernames and emails are unique per user.
  - passwords are stored hashed, never plain text.
  - access tokens and refresh tokens have distinct semantics and MUST NOT be
    mixed.
  - each postcard has one author.
  - postcard visibility is `public` or `private`.
  - each media item belongs to one postcard.
  - media groups are `header`, `gallery`, and `bgm`.
  - `header` accepts image media only.
  - `bgm` accepts audio media only.
  - media ordering is meaningful and MUST be preserved.
- Quality standards that MUST hold across the codebase:
  - Naming is English-first and consistent across packages and APIs.
  - Typed or domain-level errors are preferred over scattered raw strings.
  - Each package has one clear responsibility.
  - Large files are split by responsibility, not arbitrary line counts.
  - Business logic does not depend on Gin context objects.
  - Infrastructure adapters do not make domain decisions.

## Delivery Workflow and Review Gates

- Work MUST be planned through the spec-kit flow so that `spec.md`, `plan.md`,
  and `tasks.md` record scope, architecture, and execution order before large
  implementation changes.
- Every implementation plan MUST pass a constitution check that verifies module
  boundaries, DTO separation, dependency injection, domain invariant handling,
  and required test coverage.
- Feature specs MUST call out endpoint changes, validation rules, error contract
  changes, and preserved or amended invariants whenever affected.
- Task lists MUST include architecture work, test work, and contract work for
  each impacted user story. Tests are not optional for production code under
  this refactor.
- Reviews MUST reject work that introduces global dependency access into business
  logic, leaks persistence models into APIs, bypasses platform wiring, or omits
  tests for changed behavior.
- Commit messages MUST follow Conventional Commits and remain human-readable.
- Prioritize architectural clarity over short-term speed when those goals
  conflict.

## Governance

This constitution supersedes local habits and planning shortcuts for the
refactor branch. Amendments require a documented update to this file, an
explanation of affected principles or constraints, and synchronized updates to
dependent templates in `.specify/templates/`. Compliance review is mandatory at
plan approval time, during implementation review, and before merge or branch
handoff.

Versioning policy follows semantic versioning for governance documents:
- MAJOR: remove or redefine a principle in a backward-incompatible way.
- MINOR: add a principle, add a mandatory section, or materially expand project
  obligations.
- PATCH: clarify wording, fix ambiguity, or make non-semantic editorial changes.

When a change affects domain invariants, API contracts, or required repository
layout, the corresponding spec, plan, and tasks artifacts MUST be updated in the
same change. If a required supporting template or command file does not exist,
that absence MUST be recorded in the sync impact report until resolved.

**Version**: 1.0.0 | **Ratified**: 2026-04-20 | **Last Amended**: 2026-04-20
