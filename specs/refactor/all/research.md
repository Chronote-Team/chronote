# Research: Chronote Backend Contract-Preserving Refactor

## Decision 1: Develop In The Repository Root Architecture

- **Decision**: Build the replacement backend directly in the repository root using `cmd/api`, `internal/platform`, `internal/shared`, `internal/modules`, `migrations`, and `tests`.
- **Rationale**: The constitution explicitly prefers this root-level layout, and the user directed implementation to happen at the root instead of in an extra `refactor/` subtree.
- **Alternatives considered**:
  - Implement in a nested `refactor/` subtree: rejected because the user explicitly prefers the root-level layout.
  - Create a second repository: rejected because the current repo is the approved write scope and the spec-kit workflow lives here.

## Decision 2: Use Contract Tests As The Primary Compatibility Guardrail

- **Decision**: Treat HTTP contract tests as the main verification mechanism for route preservation, envelope compatibility, status behavior, and default message text.
- **Rationale**: The highest-risk failure mode is client-visible drift. Contract tests directly measure the preserved behavior described in the feature spec and the current API document.
- **Alternatives considered**:
  - Rely only on unit tests: rejected because unit tests do not prove HTTP compatibility.
  - Delay contract testing until the end: rejected because drift would accumulate too far before detection.

## Decision 3: Keep Business Logic Transport-Neutral And Dependency-Injected

- **Decision**: Domain and application logic depend only on domain types and explicit interfaces for repositories, token services, blacklist storage, object storage, dependency health checks, and clocks when needed.
- **Rationale**: This is required by the constitution and directly addresses the current mixing of handlers, services, globals, and infrastructure concerns in the legacy backend.
- **Alternatives considered**:
  - Use direct Gin, GORM, Redis, or S3 access inside services: rejected because it recreates the same coupling the refactor is meant to remove.
  - Centralize all behavior in shared helpers: rejected because it would create a new dumping ground instead of clear module boundaries.

## Decision 4: Preserve Current Data Semantics Before Any Schema Cleanup

- **Decision**: Keep `users`, `postcards`, and `postcard_media` semantics compatible and defer schema redesign, migration hardening, and indexing cleanup until after API parity is reached.
- **Rationale**: The feature scope is contract-preserving replacement, not schema redesign. Changing schema semantics during the refactor would widen risk and reduce confidence in compatibility verification.
- **Alternatives considered**:
  - Introduce explicit versioned schema redesign now: rejected because it changes scope and mixes operational migration work into the compatibility milestone.
  - Ignore schema concerns entirely: rejected because persistence semantics are part of the compatibility boundary and must still be documented and preserved.

## Decision 5: Implement Slices In Compatibility-Risk Order

- **Decision**: Sequence implementation as bootstrap/shared primitives, platform adapters, health, users/auth, postcards, media, and finally cutover readiness checks.
- **Rationale**: This order establishes stable infrastructure first, delivers the smallest full vertical slice early, and then tackles the most compatibility-sensitive business flows before the more complex media/storage slice.
- **Alternatives considered**:
  - Start with media or postcards first: rejected because auth and shared contract primitives are prerequisites for correct behavior.
  - Build every layer for every module at once: rejected because it would hide slice boundaries and delay usable verification.
