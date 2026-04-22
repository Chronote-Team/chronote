# Speckit Specify: Chronote Backend Refactor

  ## Title
  Chronote backend replacement refactor with preserved external API contract

  ## Summary
  Refactor the Chronote backend into a new isolated codebase structure on a
  dedicated refactor branch, while preserving the current public API behavior
  as defined by the existing implementation and the API reference document.

  The refactor is replacement-oriented, not incremental maintenance. Old code
  will be abandoned. The new codebase must become the future source of truth
  once complete.

  ## Reference Sources of Truth
  Primary API contract reference:
  - `/home/bowen/Coding/chronote/ai_generated/api_documentation.md`

  Code-level behavior reference:
  - current router, controllers, services, DTOs, models, and middleware in the
  existing repo

  Schema baseline:
  - current GORM model-defined schema for `users`, `postcards`, and
  `postcard_media`

  ## Problem Statement
  The current backend is functional, but the implementation is drifting into
  instability because HTTP handling, business logic, persistence, validation,
  and side effects are too tightly mixed.

  Current pain points include:
  - controller and service responsibilities are blurred
  - validation and normalization are scattered
  - error handling is string-based and repeated across layers
  - global dependencies are used directly in business logic
  - API behavior is harder to reason about and verify
  - infrastructure concerns influence core logic too early
  - test coverage exists but does not yet protect the full API contract

  The system should be rebuilt into clearer module boundaries without changing
  the externally observed API contract.

  ## Goals
  1. Preserve the current external API contract.
  2. Rebuild the backend into a modular, testable structure.
  3. Isolate domain logic from Gin, GORM, Redis, and S3 details.
  4. Centralize validation, normalization, authorization rules, and error
  mapping.
  5. Make future feature work predictable without reintroducing controller/
  service chaos.

  ## Non-Goals
  1. No redesign of endpoint paths or HTTP methods.
  2. No redesign of request or response envelope structure.
  3. No broad product-scope expansion.
  4. No immediate database schema redesign unless a later explicit migration is
  approved.
  5. No coexistence strategy between old and new implementations inside the
  same runtime path.

  ## Scope
  Included domains:
  - health checks
  - user registration and authentication
  - user profile operations
  - postcard CRUD
  - postcard visibility rules
  - postcard media upload, list, delete, reorder
  - token blacklist behavior
  - storage integration with PostgreSQL, Redis, and S3-compatible object
  storage

  Out of scope:
  - frontend work
  - new business capabilities
  - admin platform
  - analytics features
  - schema redesign beyond compatibility preservation

  ## Compatibility Contract
  The refactor must preserve the following by default:

  1. Endpoint paths and HTTP methods
  2. Request payload shapes
  3. Response payload shapes
  4. Response envelope structure: `code`, `message`, `data`
  5. Status-code behavior
  6. Current error-message text as much as possible
  7. Authentication behavior and token semantics
  8. Database schema and stored data semantics

  Any deviation must be treated as an explicit, documented exception.

  ## Current API Surface To Preserve
  Health:
  - `GET /health`
  - `GET /health/details`

  User:
  - `POST /user/register`
  - `POST /user/login`
  - `POST /user/refresh`
  - `GET /user/info`
  - `POST /user/logout`
  - `POST /user/avatar`
  - `PUT /user/update/displayname`
  - `PUT /user/update/password`

  Postcards:
  - `POST /v1/postcards`
  - `GET /v1/postcards`
  - `GET /v1/postcards/:id`
  - `PUT /v1/postcards/:id`
  - `DELETE /v1/postcards/:id`

  Media:
  - `POST /v1/postcards/:id/media`
  - `GET /v1/postcards/:id/media`
  - `PUT /v1/postcards/:id/media/reorder`
  - `DELETE /v1/postcards/:id/media/:media_id`

  ## Business Rules To Preserve
  1. Users have unique username and unique email.
  2. Email is normalized to lowercase before persistence.
  3. Passwords are stored hashed.
  4. Access token and refresh token semantics are distinct and must not be
  mixed.
  5. Postcards belong to one author.
  6. Postcard visibility is currently `public` or `private`.
  7. Anonymous postcard reads can access only public content.
  8. Media belongs to one postcard.
  9. Media groups are `header`, `gallery`, and `bgm`.
  10. `header` supports image only.
  11. `bgm` supports audio only.
  12. Media ordering is meaningful and preserved through `position`.
  13. Health endpoints must preserve current overall/degraded/unavailable
  behavior.

  ## Database Baseline
  Default decision: preserve the current schema.

  Current schema baseline:
  - `users`
  - `postcards`
  - `postcard_media`

  Known schema risks to document but not redesign in this stage:
  - schema migration is currently driven by `AutoMigrate` at startup rather
  than explicit versioned migrations
  - unique constraints may interact poorly with future soft-delete usage
  - email case-insensitivity is normalized in application logic, not enforced
  by DB-native case-insensitive type/indexing
  - some index strategy may be incomplete for read-heavy patterns
  - some index definitions may be redundant

  These are recorded as risks, not immediate redesign requirements.

  ## Preferred Target Structure
  The new implementation should live in a new isolated refactor code area and
  follow a vertical-slice structure.

  Preferred layout:
  - `cmd/api`
  - `internal/platform`
  - `internal/shared`
  - `internal/modules/users`
  - `internal/modules/auth`
  - `internal/modules/postcards`
  - `internal/modules/media`
  - `internal/modules/health`
  - `migrations`
  - `tests`
  - `docs`

  Preferred module internals:
  - `domain`
  - `app`
  - `infra`
  - `http`

  ## Architecture Requirements
  1. `http` handles transport only.
  2. `app` handles use-case orchestration.
  3. `domain` contains business rules, invariants, and domain errors.
  4. `infra` implements persistence and external service adapters.
  5. `platform` wires runtime dependencies and application bootstrap.
  6. Business logic must not depend on Gin context.
  7. Domain code must not depend on GORM models, Redis clients, or S3 SDK
  types.
  8. Public API DTOs must remain separate from persistence entities.
  9. Cross-module collaboration must happen through application interfaces or
  explicit contracts, not handler-to-handler calls.
  10. Global state access should be removed from core business paths where
  practical.

  ## Functional Requirements
  ### Health
  - preserve lightweight health endpoint behavior
  - preserve detailed health endpoint behavior and component reporting
  semantics
  - preserve database and redis as primary availability signals
  - preserve s3 as informational/degraded signal

  ### Users and Auth
  - preserve register, login, refresh, logout, info, avatar, display name, and
  password flows
  - preserve JWT access-token requirement for protected routes
  - preserve refresh-token body usage on refresh/logout flows
  - preserve blacklist checks where they currently apply
  - preserve current success and failure message behavior unless explicitly
  documented otherwise

  ### Postcards
  - preserve create, list, detail, update, and delete behavior
  - preserve JSON and multipart create behavior
  - preserve current pagination, sorting, and visibility behavior
  - preserve owner-only mutation behavior

  ### Media
  - preserve upload, list, reorder, and delete behavior
  - preserve media-type detection and validation behavior
  - preserve media-group constraints and limits
  - preserve ordering semantics and ownership checks

  ## Quality Requirements
  1. Validation must be centralized per use case or domain policy, not
  scattered across handlers and services.
  2. Error translation must be centralized so domain/app errors map
  consistently to HTTP responses.
  3. Shared helpers must not become a new dumping ground for unrelated logic.
  4. Files should be split by responsibility rather than by arbitrary layer
  size only.
  5. Infrastructure failures must be observable and testable.
  6. Internal storage fields must never leak in public response payloads.

  ## Testing Requirements
  1. Domain and application logic must be unit-testable without real
  PostgreSQL, Redis, or S3.
  2. Contract-level tests must verify preserved API envelopes, status codes,
  and key error messages.
  3. Critical integration tests must cover:
  - user register/login/refresh/logout
  - user info and profile mutation
  - postcard create/list/detail/update/delete
  - media upload/list/reorder/delete
  - health endpoints
  4. Compatibility-focused regression tests should compare new behavior against
  the documented current contract.

  ## Delivery Model
  1. Work happens on a dedicated refactor branch.
  2. Old implementation is abandoned as the forward path once refactor begins.
  3. Refactor may be built in a separate directory structure inside the repo
  during development.
  4. The new implementation becomes the only intended future source of truth.
  5. Replacement of the current dev branch happens only after compatibility and
  verification goals are met.

  ## Risks
  1. Hidden behavior in current handlers may not be fully captured by docs
  alone.
  2. Some clients may rely on exact error strings, including inconsistent ones.
  3. Preserving behavior while changing architecture can expose undocumented
  edge cases.
  4. Startup `AutoMigrate` may conceal schema assumptions that should later be
  made explicit.
  5. Rebuild-oriented refactor can drift if compatibility tests are not
  established early.

  ## Acceptance Criteria
  The refactor is complete when:
  1. the new codebase exposes the same endpoint paths and methods
  2. request and response payloads remain compatible with the documented API
  3. response envelope structure remains compatible
  4. status-code behavior remains compatible
  5. error-message text remains compatible by default, except for documented
  exceptions
  6. current schema remains supported without forced redesign
  7. auth, postcard, media, and health flows are covered by automated tests
  8. core business logic is isolated from transport and infrastructure concerns
  9. the new codebase is clear enough that future endpoint work can be done
  within a single module without cross-layer chaos

  ## Explicit Decisions Locked For This Spec
  - no redesign of API endpoints
  - current API document is the contract reference
  - preserve current response envelope
  - preserve current error-message text by default
  - preserve the current database schema by default
