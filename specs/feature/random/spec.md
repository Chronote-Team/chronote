# Feature Specification: Random Accessible Postcard

**Feature Branch**: `feature/random`  
**Created**: 2026-05-10  
**Status**: Ready for Planning  
**Input**: User description: "Read docs/v1-postcards-random.md, first read the index, then use the Create The Spec section to create the feature specification."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Discover a Random Postcard (Priority: P1)

As a Chronote user on the main page, I can refresh and receive a random postcard I am allowed to view so the feed feels social and variable instead of repeatedly showing the same deterministic list item.

**Why this priority**: This is the smallest usable social-discovery behavior for the main page and delivers the primary user value of the feature.

**Independent Test**: Can be fully tested by requesting a random postcard as an anonymous visitor and as a signed-in user, then confirming that exactly one accessible postcard is returned when eligible content exists.

**Acceptance Scenarios**:

1. **Given** at least one public postcard exists, **When** an anonymous visitor requests a random postcard, **Then** the response succeeds with one public postcard.
2. **Given** a signed-in user owns private postcards and public postcards exist, **When** the user requests a random postcard, **Then** the random candidate set includes public postcards and that user's own postcards.
3. **Given** only private postcards owned by other users exist, **When** an anonymous visitor requests a random postcard, **Then** the response uses the existing not-found postcard behavior.
4. **Given** the random postcard route exists alongside postcard detail routes, **When** a caller requests the random postcard route, **Then** the system treats it as the random postcard request and not as a postcard identifier lookup.

### Edge Cases

- No postcards exist.
- Only private postcards exist.
- The caller does not provide authentication.
- The caller provides invalid authentication.
- No accessible postcard is available after applying visibility and ownership rules.
- The selected postcard has no media.
- The selected postcard's author relationship cannot be resolved; existing postcard relation-loading behavior is preserved.

## Contract and Invariant Impact *(mandatory)*

### API Contract Impact

- Affected route groups: `/v1/postcards/*`
- Request schema changes: Adds `GET /v1/postcards/random` with no request body and no query parameters. Authentication remains optional for this read path.
- Response schema changes: The success payload uses the same postcard response shape as existing postcard detail and list responses, including existing author and media relation fields.
- Error schema changes: If no accessible postcard exists, the route uses the existing not-found postcard status and message behavior. Invalid optional authentication does not expand anonymous visibility.

### Domain Invariant Impact

- Preserved invariants:
  - Anonymous postcard reads remain limited to public postcards.
  - Signed-in users may read public postcards and their own postcards.
  - Private postcards owned by other users remain inaccessible.
  - Each response contains at most one postcard.
  - Existing postcard author and media relationship behavior remains unchanged.
- New invariants:
  - A random postcard request selects from only the postcards accessible to the current caller.
  - The random postcard route is reserved and cannot be interpreted as a postcard identifier.
- Amended invariants: None.

### Boundary Impact

- Modules touched: postcard read behavior and route contract.
- Persistence leakage check: Public postcard responses remain separate from stored postcard records.
- Dependency isolation check: Access rules and random selection behavior remain part of the postcard business flow, not embedded in transport handling.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose `GET /v1/postcards/random`.
- **FR-002**: The random postcard route MUST accept requests without authentication.
- **FR-003**: Anonymous callers MUST only receive postcards whose visibility is public.
- **FR-004**: Signed-in callers MUST receive a random postcard from the set of public postcards plus postcards authored by that signed-in user.
- **FR-005**: The random postcard route MUST return at most one postcard per request.
- **FR-006**: The returned postcard MUST use the same response fields as existing postcard detail and list responses.
- **FR-007**: The returned postcard MUST include author and media relationships using the same behavior as existing postcard detail and list responses.
- **FR-008**: When no accessible postcard exists, the route MUST return the existing not-found postcard behavior.
- **FR-009**: The random postcard route MUST be distinguishable from postcard detail lookup so `random` is not treated as a postcard identifier.
- **FR-010**: Invalid optional authentication MUST NOT grant access to private postcards owned by other users.
- **FR-011**: The feature MUST NOT add personalized ranking, recently seen filtering, owner exclusion, weighted randomness, cursor-based feed behavior, or frontend changes.

### Key Entities *(include if feature involves data)*

- **Postcard**: A user-authored content item with visibility that determines who can read it.
- **Caller**: The visitor requesting a random postcard, either anonymous or signed in.
- **Author**: The user who owns a postcard and whose ownership affects private postcard accessibility.
- **Postcard Media**: Media related to a postcard and returned through the existing postcard response shape.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of anonymous random postcard requests return only public postcards when successful.
- **SC-002**: 100% of signed-in random postcard requests choose only from public postcards and postcards authored by the signed-in user.
- **SC-003**: 100% of requests with no accessible postcard return the existing not-found postcard behavior.
- **SC-004**: 100% of successful random postcard responses match the existing postcard response shape for shared fields and relationships.
- **SC-005**: Existing postcard list and detail behaviors continue to pass their current compatibility checks after the random postcard route is added.

## Assumptions

- The main page is the initial consumer of the random postcard behavior.
- Authentication is optional and follows the existing public postcard read behavior.
- Randomness does not need to avoid postcards the caller recently saw.
- Randomness does not need to exclude postcards authored by the current signed-in user.
- Randomness does not need personalized ranking or weighting.
- Frontend integration is out of scope for this feature.
