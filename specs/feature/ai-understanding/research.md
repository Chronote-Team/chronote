# Research: Postcard AI Understanding

## Decision 1: Use A Dedicated Internal `postcardai` Module

- **Decision**: Add a dedicated `internal/modules/postcardai` module for analysis jobs, result persistence, AI workflow orchestration, provider boundaries, and validation rules.
- **Rationale**: The feature is related to postcards and media, but it introduces separate domain concepts: jobs, analysis results, prompt/schema/model versions, confidence, stale-result prevention, and provider failure categories. Keeping them in a dedicated module prevents public postcard/media handlers from accumulating AI workflow logic.
- **Alternatives considered**:
  - Put the workflow inside `internal/modules/postcards`: rejected because it would mix postcard CRUD behavior with internal AI job orchestration.
  - Put the workflow inside `internal/modules/media`: rejected because final postcard understanding combines text and media facts and is not media-only.
  - Put the workflow in `internal/platform`: rejected because it is business workflow, not concrete wiring.

## Decision 2: PostgreSQL Is Durable Source Of Truth, Redis Is Coordination Only

- **Decision**: Store analysis jobs and results durably in PostgreSQL. Use Redis only for short-lived queue wake-up, locking, rate-limit counters, idempotency hints, or retry coordination.
- **Rationale**: The spec requires jobs to be recoverable after worker interruption and requires durable image/postcard analysis results. Redis availability must not decide whether analysis state exists.
- **Alternatives considered**:
  - Redis-only queue/results: rejected because job state would be less durable and harder to inspect after outages.
  - PostgreSQL-only coordination: acceptable as a fallback, but Redis remains useful for distributed locks and worker wake-up when available.

## Decision 3: OpenAI Responses-Compatible Gateway Behind `AIClient`

- **Decision**: Use an OpenAI Responses-compatible provider adapter behind an application-level `AIClient` interface for text/image input and structured JSON output.
- **Rationale**: The source plan names the Responses API as the preferred provider path. An interface keeps provider details out of business logic and allows fakes in tests.
- **Alternatives considered**:
  - Direct provider calls from the worker use case: rejected because it makes app logic depend on provider payloads and credentials.
  - Generic HTTP client exposed to app logic: rejected because it leaks provider request construction into orchestration code.
  - Multiple provider adapters in v1: deferred until there is a concrete second provider requirement.

## Decision 4: Versioned Prompts, Schemas, Models, And Content Keys

- **Decision**: Store prompt version, schema version, model version, media version, and postcard version with every job/result. Use those values in uniqueness and reuse decisions.
- **Rationale**: The feature requires cache reuse for unchanged media and traceability when prompts, schemas, or models change. Version keys prevent stale or mismatched output from being treated as current understanding.
- **Alternatives considered**:
  - Reuse by media ID only: rejected because changed media or changed schemas would incorrectly reuse old output.
  - Store only provider model: rejected because prompt/schema changes also affect output meaning.

## Decision 5: Short-Lived Private Media Access Inside Workers Only

- **Decision**: Generate short-lived signed media access only inside backend workers, send it to the AI provider when needed, discard it after the call, and regenerate it on retry.
- **Rationale**: Media remains private, signed links should not become durable data, and the normal client app must not receive analysis-only media links.
- **Alternatives considered**:
  - Store signed URLs in analysis jobs: rejected because links expire and create unnecessary privacy risk.
  - Make media public for analysis: rejected because it violates the private-media boundary.

## Decision 6: Validate Output Before Successful Storage With One Repair Retry

- **Decision**: Validate image and postcard structured output before storing it as successful. If output is malformed, perform at most one schema-repair retry and then store either a valid result or a non-successful outcome.
- **Rationale**: The spec requires invalid structured output not to be stored as successful understanding. A single repair attempt addresses common formatting failures without allowing uncontrolled retry loops.
- **Alternatives considered**:
  - Store provider output as-is: rejected because future features need dependable structured data.
  - Retry until valid: rejected because it risks runaway provider usage and delayed job completion.

## Decision 7: Preserve Public API Non-Change With Contract Tests

- **Decision**: Add contract tests that assert no client-facing analysis endpoints or analysis fields are introduced and existing postcard/media response shapes remain unchanged.
- **Rationale**: The highest-risk product regression is leaking internal workflow state into normal client behavior. Contract tests directly guard the explicit out-of-scope boundary from the index section.
- **Alternatives considered**:
  - Rely on implementation review only: rejected because endpoint drift is easy to miss.
  - Add public status endpoints now for observability: rejected because the current phase explicitly excludes them.
