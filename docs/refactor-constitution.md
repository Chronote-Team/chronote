  # Chronote Refactor Constitution

  ## Basic Information

  - Older codes of project is stored in "/home/bowen/Coding/chronote"
  - Newer location is "/home/bowen/Coding/chronote-refactor", all changes shouldn't cross the line of this folder.  

  ## Project Identity
  Chronote is a Go backend service for user accounts, authentication,
  postcards, and postcard media management.
  Its current runtime stack is:
  - Go
  - Gin HTTP server
  - GORM
  - PostgreSQL
  - Redis
  - S3-compatible object storage

  ## Current Domain Scope
  The system currently owns these business areas:
  - User registration, login, refresh token, logout
  - User profile info, avatar upload, display name update, password update
  - Postcard CRUD
  - Postcard visibility control (`public`, `private`)
  - Postcard media upload, list, delete, reorder
  - Health check endpoints for database, redis, and s3 dependencies

  ## Current API Surface
  Current main route groups are:
  - `/health`
  - `/health/details`
  - `/user/*`
  - `/v1/postcards/*`

  Protected endpoints use JWT-based auth.
  Optional auth is used for some postcard read endpoints.

  ## Current Problems Driving Refactor
  The current codebase is functional but is trending toward chaos because:
  - controller, service, validation, and response concerns are still too mixed
  - API behavior and naming are not fully consistent
  - error handling is string-based and scattered
  - global runtime dependencies are used directly (`global.Db`, redis, s3
  config)
  - side effects and persistence logic are tightly coupled
  - route structure and endpoint semantics need cleanup
  - testing exists only in limited areas and is not yet a full safety net
  - old code will be abandoned immediately once the new refactor path starts

  ## Refactor Intent
  This refactor is a rebuild-oriented refactor on a dedicated branch.
  Old code is not to be incrementally maintained.
  A new code structure may be created separately and will become the future
  source of truth.
  When complete, this branch is intended to replace the current dev branch.

  ## Preferred Refactor Layout
  The refactor should follow a vertical-slice structure with shared
  infrastructure separated from business modules.

  Preferred root layout:
  - `cmd/api` for application entrypoint only
  - `internal/platform` for config, db, redis, s3, auth wiring, HTTP bootstrap
  - `internal/shared` for shared technical utilities with no business ownership
  - `internal/modules/users`
  - `internal/modules/auth`
  - `internal/modules/postcards`
  - `internal/modules/media`
  - `internal/modules/health`
  - `migrations`
  - `tests`
  - `docs`

  Each business module should prefer this internal shape when needed:
  - `domain` for entities, invariants, domain errors, business rules
  - `app` for use cases and orchestration
  - `infra` for repository and gateway implementations
  - `http` for handlers, request DTOs, response DTOs, and route registration

  ## Dependency Direction Rules
  - `http` may depend on `app`
  - `app` may depend on `domain` and repository or gateway interfaces
  - `infra` may depend on `domain` and implement interfaces required by `app`
  - `platform` wires concrete implementations together
  - `domain` must not depend on Gin, GORM, Redis, S3, or transport concerns
  - business modules must not depend on each other through HTTP handlers
  - cross-module collaboration should happen through application interfaces or
  shared contracts, not direct transport-layer calls

  ## Non-Negotiable Principles
  1. Clear layer boundaries
  - Transport/API layer handles HTTP only
  - Application/service layer handles use cases
  - Domain layer holds business rules and invariants
  - Repository/infrastructure layer handles database, redis, and s3 access

  2. No persistence models leaking into API contracts
  - Request/response DTOs must remain separate from database entities
  - Internal storage fields must never leak into public responses

  3. Deterministic API contracts
  - Each endpoint must have a stable request schema, response schema, and error
  schema
  - Validation rules must be explicit and centralized
  - Status codes must be consistent by error category

  4. Dependency isolation
  - No direct global dependency access inside business logic where avoidable
  - Database, redis, object storage, token service, and clock should be
  injectable behind interfaces where practical

  5. Predictable business rules
  - Postcard visibility rules must be centralized
  - Media group rules must be centralized
  - Auth token rules must be centralized
  - Input normalization must happen in one clear place

  6. Refactor for replacement, not coexistence
  - The new architecture does not need to preserve the old internal layout
  - Compatibility matters at the API contract level, not at the package-layout level

  7. Testability first
  - Core business rules must be unit-testable without real postgres, redis, or s3
  - Integration tests should cover critical request flows
  - New modules should not be accepted without tests for their main behavior

  ## Domain Invariants To Preserve
  - Users have unique username and unique email
  - Passwords are stored hashed, never plain text
  - Access tokens and refresh tokens have different meanings and must not be
  mixed
  - A postcard belongs to one author
  - Postcard visibility is currently `public` or `private`
  - Media belongs to one postcard
  - Media groups are currently `header`, `gallery`, `bgm`
  - `header` must be image only
  - `bgm` must be audio only
  - Media ordering is meaningful and must be preserved

  ## Quality Standards
  - Naming must be consistent and English-first in code structure
  - Error handling should use typed/domain errors instead of scattered raw strings where possible
  - Each package must have one clear responsibility
  - Large files should be split by responsibility, not by arbitrary size alone
  - Business logic must not depend on Gin context objects
  - Infrastructure adapters must not contain domain decisions

  ## Delivery Constraints
  - Work will be done with SDD on a dedicated refactor branch
  - The branch is allowed to diverge because old code is being abandoned
  - The goal is a clean replacement, not a minimal patch set
  - Prioritize architectural clarity over short-term speed
  - Commits informations should follow Conventional Commits, keep human-readable

  ## Success Criteria
  The refactor is successful when:
  - the new codebase has stable architectural boundaries
  - API behavior is deterministic and documented
  - auth, postcard, and media flows are covered by tests
  - external dependencies are isolated behind clear interfaces
  - the codebase is easier to extend without controller/service chaos
  
