# Feature Specification: Postcard AI Understanding Worker

**Feature Branch**: `feature/ai-understanding-worker`

**Created**: 2026-05-15

**Status**: Draft

**Input**: User description: "Read docs/postcard-ai-understanding-worker.md, use the index and Create The Spec sections, and create the feature specification for branch feature/ai-understanding-worker."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Consume Queued Understanding Work (Priority: P1)

As a Chronote operator, I want postcard understanding work that is already queued after postcard or media changes to be consumed automatically, so analysis work does not remain pending forever.

**Why this priority**: This is the minimum useful feature. Without automatic consumption, the existing enqueue behavior produces durable work records but no usable understanding results.

**Independent Test**: Create or update a postcard so an analysis work item is queued, run the background processor, and verify the work item is claimed and reaches a non-pending outcome.

**Acceptance Scenarios**:

1. **Given** a postcard change has queued analysis work, **When** the background processor is running under idle conditions, **Then** the work is claimed within 10 seconds and no longer remains pending forever.
2. **Given** a postcard has only text content, **When** the background processor handles its queued work, **Then** a postcard-level understanding result is stored without requiring image content.
3. **Given** a postcard has image media, **When** the background processor handles its queued work, **Then** image understanding is completed or reused before postcard-level understanding is stored.

---

### User Story 2 - Preserve Normal Chronote Behavior (Priority: P2)

As a normal Chronote user, I want postcard creation, editing, media management, and reading to behave as they do today, so AI understanding work never changes the public user experience or blocks normal interactions.

**Why this priority**: The feature is internal background processing. It must not introduce new public behavior, response fields, or user-facing failure modes.

**Independent Test**: Run the existing public behavior tests while analysis work is queued and while the background processor is active, then verify response shapes and status behavior remain unchanged.

**Acceptance Scenarios**:

1. **Given** a user creates, updates, or reads a postcard, **When** analysis work is queued or being processed, **Then** the normal Chronote response remains unchanged.
2. **Given** analysis succeeds or fails in the background, **When** a user reads a postcard through the normal public surface, **Then** no AI status or AI result fields are exposed.
3. **Given** no background processor is running, **When** users use postcard and media features, **Then** normal user workflows continue to work while analysis work remains queued.

---

### User Story 3 - Handle Retries, Stale Work, and Reuse (Priority: P3)

As an operator, I want the background processor to classify stale work, retry temporary failures, and reuse still-valid image understanding, so the system remains correct and avoids unnecessary provider work.

**Why this priority**: Once basic processing exists, correctness under changing postcards and provider failures determines whether the feature can be trusted in real deployments.

**Independent Test**: Queue work, mutate the target postcard before processing, simulate temporary provider failures, and process unchanged media more than once; verify each outcome is classified correctly.

**Acceptance Scenarios**:

1. **Given** a postcard changes after analysis work is queued, **When** the older work is processed, **Then** it is marked stale and does not store stale postcard understanding.
2. **Given** current image understanding already exists for unchanged media, **When** new postcard-level work is processed, **Then** the existing image understanding is reused instead of repeating image analysis.
3. **Given** the provider is temporarily unavailable, **When** analysis work is processed, **Then** durable state remains recoverable and the work is classified for retry or a clear unavailable outcome.
4. **Given** provider output is malformed, **When** one controlled repair attempt produces valid structured output, **Then** the repaired result is accepted and stored.

---

### User Story 4 - Operate Securely and Observably (Priority: P4)

As an operator, I want enough processing logs to understand job outcomes without exposing private content or credentials, so local and production deployments can be diagnosed safely.

**Why this priority**: Operators need to verify that the processor is alive and making progress, but the analysis flow handles private text, image access links, and provider credentials.

**Independent Test**: Process successful, stale, failed, and retryable work; inspect logs for job identifiers and outcome categories, and verify sensitive content is absent.

**Acceptance Scenarios**:

1. **Given** a work item is claimed, succeeds, becomes stale, fails, or is scheduled for retry, **When** logs are inspected, **Then** the logs identify the work item and outcome category.
2. **Given** postcard text, image data, image access links, provider credentials, or full provider payloads are involved, **When** logs are inspected, **Then** those sensitive values are not present.
3. **Given** the processor receives a stop signal while idle or while processing, **When** shutdown begins, **Then** it exits promptly and leaves durable work state recoverable.

### Edge Cases

- No background processor is running: queued analysis work remains pending, and normal Chronote user workflows continue.
- AI understanding is disabled by configuration: the processor does not crash and does not store false successful results.
- Provider credentials are missing or invalid: affected work is not marked successful, and the outcome is inspectable.
- A custom provider access path is configured: provider calls use that configured path.
- Provider output appears in different successful text shapes: the system extracts supported structured content consistently.
- Provider output is non-structured text: the system classifies it as malformed and performs at most one repair attempt.
- Image access links are not reachable by the provider: image understanding fails gracefully and the postcard outcome remains accurate.
- Shutdown happens during idle wait: the processor exits promptly.
- Shutdown happens during in-flight processing: cancellation is respected and durable state remains recoverable.
- Multiple processor instances run at the same time: each queued work item is processed by at most one instance at a time.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide an operational background processor that repeatedly checks for queued postcard AI understanding work.
- **FR-002**: The processor MUST consume work created by existing postcard text, postcard metadata, and postcard media changes.
- **FR-003**: The processor MUST keep AI provider work outside normal user request completion, so public user interactions are not blocked by provider latency.
- **FR-004**: The system MUST prevent two processor instances from processing the same work item at the same time.
- **FR-005**: The processor MUST mark work as stale when the queued postcard version no longer matches the current postcard version.
- **FR-006**: The processor MUST reuse existing image understanding when the media input and analysis compatibility attributes are still current.
- **FR-007**: The processor MUST request image understanding for eligible media that does not have reusable image understanding.
- **FR-008**: The processor MUST request postcard-level understanding after collecting available postcard text and current image facts.
- **FR-009**: The system MUST convert successful provider responses into the internal structured result shape.
- **FR-010**: The processor MUST validate structured provider output before storing it as successful understanding.
- **FR-011**: The processor MUST perform no more than one controlled repair attempt for malformed structured output.
- **FR-012**: The processor MUST store successful image-level understanding internally.
- **FR-013**: The processor MUST store successful postcard-level understanding internally.
- **FR-014**: The processor MUST update every processed work item to a terminal, stale, unavailable, or retryable state after each attempt.
- **FR-015**: The processor MUST preserve durable work state across temporary provider failures.
- **FR-016**: The processor MUST support graceful shutdown during idle waiting and in-flight processing.
- **FR-017**: Deployment MUST support running the public Chronote service and the background processor together from the same project release.
- **FR-018**: The processor MUST NOT log raw postcard text, raw image bytes, image access links, provider credentials, or full provider payloads.
- **FR-019**: The processor MUST log enough non-sensitive information to confirm work claim, outcome, retry scheduling, stale detection, and provider failure category.
- **FR-020**: The feature MUST be verifiable in a local deployment using the existing local service stack.
- **FR-021**: The normal public Chronote contract MUST remain unchanged, including paths, response shapes, status behavior, and default error-message text.

### Key Entities

- **Analysis Work Item**: A queued request to understand a specific postcard version. It has ownership, target version, attempt state, and final outcome.
- **Postcard Version Snapshot**: The current postcard content state used to decide whether queued work is still valid.
- **Media Understanding Result**: Internal structured facts derived from a postcard media item and its compatible analysis attributes.
- **Postcard Understanding Result**: Internal structured facts derived from postcard text plus available media understanding.
- **Provider Result**: A structured response from an AI provider that must be parsed, validated, and either stored or classified as invalid.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In a local deployment with the background processor enabled, newly queued analysis work is claimed within 10 seconds under idle conditions.
- **SC-002**: A text-only postcard can move from queued work to stored postcard understanding in an automated test using a controlled provider.
- **SC-003**: A postcard with image media can move from queued work to stored media understanding and stored postcard understanding in an automated test using a controlled provider.
- **SC-004**: Public compatibility tests confirm no AI status fields, AI result fields, or new public AI surfaces appear in normal postcard responses.
- **SC-005**: A local deployment can start the public Chronote service and background processor together using the same operational configuration source.
- **SC-006**: Valid live provider output is parsed into validated internal structured results instead of being rejected as unparsed output.
- **SC-007**: Processing logs identify work item IDs and outcome categories while omitting provider credentials, image access links, raw image bytes, raw postcard content, and full provider payloads.
- **SC-008**: Reprocessing unchanged media reuses existing media understanding in automated tests.

## Assumptions

- Existing postcard and media mutations already queue analysis work after successful changes.
- Existing durable storage for queued work and internal analysis results is already available in target environments.
- Existing analysis orchestration remains the source of truth for stale checks, reuse decisions, provider calls, validation, repair, and result storage.
- The normal public Chronote service should remain focused on user request handling and enqueueing work.
- The background processor can share the same operational configuration values as the public Chronote service.
- Real image understanding requires provider-reachable image access links; local-only storage links are suitable for local storage tests but may not be reachable by external providers.
