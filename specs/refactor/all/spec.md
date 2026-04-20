# Feature Specification: Chronote Backend Contract-Preserving Refactor

**Feature Branch**: `refactor/all`  
**Created**: 2026-04-20  
**Status**: Ready for Planning  
**Input**: User description: "Read `docs/refactor-specify.md` and create the specification for the Chronote backend replacement refactor with preserved external API contract."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Preserve Existing Client Flows (Priority: P1)

As a Chronote client consuming the current backend, I can continue using authentication, profile, postcard, and media endpoints against the replacement backend without changing my existing request paths, request formats, or response handling.

**Why this priority**: Existing clients must keep working without disruption. If this fails, the refactor does not deliver usable value.

**Independent Test**: Can be fully tested by running the current client request flows against the replacement backend and confirming compatible paths, methods, payload shapes, status codes, and response envelopes.

**Acceptance Scenarios**:

1. **Given** a client that already uses the supported Chronote endpoints, **When** it sends the same requests to the replacement backend, **Then** the backend accepts the same endpoint paths and HTTP methods and returns compatible response envelopes.
2. **Given** an existing user account and valid credentials, **When** the user registers, logs in, refreshes tokens, logs out, and retrieves profile information, **Then** the observed success and failure outcomes remain compatible with the baseline behavior.

---

### User Story 2 - Preserve Postcard and Media Behavior (Priority: P2)

As an authenticated or anonymous Chronote user, I experience the same postcard visibility, ownership, media grouping, and ordering rules after the backend replacement.

**Why this priority**: Postcards and media are core product behaviors. Compatibility here protects the main user value of the system.

**Independent Test**: Can be fully tested by creating, viewing, updating, deleting, uploading, reordering, and removing postcard media while validating public/private access rules and media constraints.

**Acceptance Scenarios**:

1. **Given** a public postcard and a private postcard, **When** an anonymous user requests both, **Then** the public postcard remains accessible and the private postcard remains inaccessible.
2. **Given** a postcard owner with existing postcard media, **When** the owner uploads, lists, reorders, and deletes media, **Then** the same media-group constraints, ownership checks, and ordering semantics are preserved.

---

### User Story 3 - Replace the Legacy Backend Safely (Priority: P3)

As a maintainer responsible for the refactor, I can replace the legacy backend with the new backend while preserving supported data semantics, health reporting behavior, and documented compatibility expectations.

**Why this priority**: The refactor only succeeds if it can become the new source of truth without forcing client rewrites or operational ambiguity.

**Independent Test**: Can be fully tested by running compatibility verification across supported routes, validating health-check behavior under dependency degradation, and confirming the replacement backend works with existing supported data.

**Acceptance Scenarios**:

1. **Given** existing supported data and current API contracts, **When** the replacement backend is run against that data, **Then** it supports the same user, postcard, and media semantics without requiring a forced data reset or redesign.
2. **Given** dependency states that affect service health, **When** the health endpoints are requested, **Then** the lightweight and detailed health responses preserve the expected overall, degraded, and unavailable meanings.

### Edge Cases

- Clients may depend on exact legacy error messages or status-code combinations, including inconsistent ones.
- Both JSON and multipart postcard creation paths must remain supported and behave compatibly.
- Anonymous reads must continue to allow only public postcards while preventing private postcard access.
- Media operations must preserve group-specific constraints, including image-only header media and audio-only background music.
- Media reorder and delete requests must handle missing, duplicate, foreign, or out-of-sequence media identifiers without breaking ordering guarantees.
- Detailed health checks must continue distinguishing between fully unavailable core dependencies and degraded informational dependencies.

## Contract and Invariant Impact *(mandatory)*

### API Contract Impact

- Affected route groups: `/health`, `/user/*`, `/v1/postcards/*`
- Request schema changes: None intended. Existing request payload shapes, normalization rules, JSON and multipart creation formats, and token usage patterns remain compatible.
- Response schema changes: None intended. Existing response envelope structure (`code`, `message`, `data`) and public response payloads remain compatible.
- Error schema changes: None intended. Existing status-code behavior and default error-message text remain compatible unless an approved exception is documented.

### Domain Invariant Impact

- Preserved invariants:
  - Usernames and emails remain unique per user.
  - Emails remain normalized before persistence.
  - Passwords remain stored as hashes rather than plain text.
  - Access tokens and refresh tokens remain distinct in purpose and usage.
  - Each postcard remains owned by one author.
  - Postcard visibility remains limited to `public` and `private`.
  - Anonymous postcard reads remain limited to public content.
  - Each media item remains attached to one postcard.
  - Media groups remain `header`, `gallery`, and `bgm`.
  - Header media remains image-only.
  - Background music remains audio-only.
  - Media ordering remains meaningful and preserved.
  - Health endpoints retain their current overall, degraded, and unavailable semantics.
- New invariants: None.
- Amended invariants: None.

### Boundary Impact

- Modules touched: health, users, auth, postcards, media, plus supporting platform and shared boundaries required to isolate cross-cutting concerns
- Persistence leakage check: Public request and response DTOs remain separate from persistence and storage representations.
- Dependency isolation check: Business rules, validation, normalization, authorization, and error handling remain independent from transport and infrastructure-specific runtime concerns.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST preserve the current public endpoint paths and HTTP methods for all supported health, user, postcard, and media operations.
- **FR-002**: The system MUST preserve the current request payload shapes, field meanings, and normalization behavior for all supported routes.
- **FR-003**: The system MUST preserve the current response envelope structure and keep public response payloads compatible for supported routes.
- **FR-004**: The system MUST preserve the current success and failure status-code behavior for equivalent requests.
- **FR-005**: The system MUST preserve current error-message text by default, and any approved deviations MUST be explicitly documented before release.
- **FR-006**: The system MUST preserve user registration, login, refresh, logout, profile retrieval, avatar update, display name update, and password update flows.
- **FR-007**: The system MUST preserve unique username and email rules, lowercase email normalization, hashed password storage, and the distinct semantics of access and refresh tokens.
- **FR-008**: The system MUST preserve protected-route authentication behavior and current token blacklist behavior where it applies.
- **FR-009**: The system MUST preserve postcard create, list, detail, update, and delete behavior, including JSON and multipart creation paths.
- **FR-010**: The system MUST preserve current postcard visibility, pagination, sorting, and owner-only mutation behavior.
- **FR-011**: The system MUST preserve media upload, list, reorder, and delete behavior, including media-group constraints, type restrictions, and ordering semantics.
- **FR-012**: The system MUST preserve supported stored-data semantics for users, postcards, and postcard media without forcing an immediate schema redesign or client-visible migration step.
- **FR-013**: The system MUST preserve the meaning of the lightweight and detailed health endpoints, including compatibility of overall, degraded, and unavailable service reporting.
- **FR-014**: The system MUST support replacement of the legacy backend as the sole future source of truth without requiring both old and new implementations to remain active on the same runtime path.
- **FR-015**: The system MUST make compatibility verification possible for authentication, profile, postcard, media, and health flows before the replacement is considered complete.

### Key Entities *(include if feature involves data)*

- **User Account**: A person using Chronote, identified by unique username and email, with profile attributes and authentication credentials.
- **Auth Session Tokens**: The access and refresh credentials used to authorize protected actions, refresh sessions, and support logout and blacklist behavior.
- **Postcard**: A user-owned content record that can be created, listed, viewed, updated, or deleted and whose visibility can be public or private.
- **Postcard Media Item**: A media record attached to a postcard with a defined group, type constraints, and an ordering position.
- **Health Status Snapshot**: A representation of current service availability used to communicate overall and component-level readiness to clients and operators.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of documented supported routes remain callable at the same endpoint paths and HTTP methods without requiring client-side URL or verb changes.
- **SC-002**: 100% of critical flows for authentication, profile operations, postcard CRUD, media operations, and health checks pass compatibility verification against the baseline contract before replacement is approved.
- **SC-003**: 0 forced client-visible schema redesigns or mandatory data resets are required for supported existing data to work with the replacement backend.
- **SC-004**: At least 95% of sampled baseline success and failure scenarios across each route group match the prior status code, response envelope, and message behavior exactly, and every remaining mismatch is documented and approved before release.
- **SC-005**: Every supported route group has automated coverage for at least one primary success scenario and one primary failure, authorization, or degradation scenario before the legacy backend is retired.

## Assumptions

- The current API reference document and current backend behavior together define the compatibility baseline for this refactor.
- Existing supported Chronote clients are expected to continue working without frontend or client workflow redesign.
- New product capabilities, admin tooling, analytics features, and broad scope expansion are out of scope for this feature.
- Existing supported stored data in user, postcard, and postcard media records remains valid and must continue working without forced redesign during this feature.
- Legacy inconsistencies that clients already rely on are treated as part of the observed contract unless an explicit exception is documented and approved.
- The replacement backend becomes the only intended future implementation once compatibility verification and review gates are satisfied.
