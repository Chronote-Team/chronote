# Research: Random Accessible Postcard

## Decision: Keep the feature in the existing postcard vertical slice

**Rationale**: The feature is a postcard read capability under `/v1/postcards/*`. The existing slice already separates HTTP handling, application orchestration, domain shape, and repository implementation. Keeping the work inside `internal/modules/postcards` satisfies the constitution and avoids unnecessary new module boundaries.

**Alternatives considered**: Creating a feed or discovery module was rejected because the source scope explicitly excludes personalized ranking, cursor feed behavior, weighting, and recently seen filtering.

## Decision: Use optional access-token behavior from existing public postcard reads

**Rationale**: Random reads must work for anonymous and authenticated callers. Existing public postcard routes already run behind optional access-token middleware, which sets user context when a valid access token is present and otherwise continues anonymously.

**Alternatives considered**: Requiring authentication was rejected because anonymous callers are in scope. Treating invalid tokens as hard failures was rejected because the current optional auth middleware continues anonymously for invalid optional credentials, and the feature spec preserves existing optional read behavior.

## Decision: Put random access selection behind the postcard repository interface

**Rationale**: The app service owns the use case and error mapping, while repository implementations own data retrieval. Adding `FindRandomAccessible(userID uint)` gives production storage a direct random query and gives the in-memory repository a matching behavior for test app wiring.

**Alternatives considered**: Reusing `List()` and randomizing in the app layer was rejected for production because it would load unnecessary rows. Putting access logic only in the storage adapter was rejected because the service still owns use-case errors and relation attachment.

## Decision: Preserve existing relation attachment after selection

**Rationale**: The returned postcard must match detail/list response shape and include author and media relations the same way existing reads do. Calling the existing relation attachment path after a random postcard is selected keeps response behavior consistent.

**Alternatives considered**: Eager-loading relations directly in the random storage query was rejected because it would duplicate existing response assembly behavior and risk drifting from list/detail semantics.

## Decision: Register `/random` before `/:id`

**Rationale**: The public postcard routes include a dynamic detail route. Registering the literal route first ensures `/v1/postcards/random` reaches the random handler instead of being parsed as an invalid postcard ID.

**Alternatives considered**: Changing the detail route shape was rejected because preserving existing postcard detail paths is a non-negotiable compatibility requirement.

## Decision: Cover behavior with focused app tests and contract tests

**Rationale**: The feature changes a public route and a business access rule. App tests can validate candidate filtering without external dependencies, while contract tests verify endpoint messages, status behavior, response visibility, and route order.

**Alternatives considered**: Relying only on repository tests was rejected because it would miss response contract and route-order regressions. Relying only on contract tests was rejected because app-level filtering is easier to isolate directly.
