# Feature Specification: Postcard AI Understanding

**Feature Branch**: `feature/ai-understanding`  
**Created**: 2026-05-13  
**Status**: Ready for Planning  
**Input**: User description: "Read `docs/postcard-ai-understanding-final.md`, use the index for retrieval context, and create the spec from the `Create The Spec` section for branch `feature/ai-understanding`."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Analyze New Or Changed Postcards Internally (Priority: P1)

As the Chronote backend system, I can detect when a postcard is created or meaningfully changed, analyze the postcard text and attached user photos, and store structured understanding without requiring the normal client app user to start or monitor the workflow.

**Why this priority**: This is the core value of the feature. Stored understanding only exists if the backend can create it from normal postcard activity without changing the user-facing workflow.

**Independent Test**: Can be fully tested by creating or updating a postcard with text and attached images, running the internal analysis workflow, and confirming image-level and postcard-level understanding are stored while the normal postcard API behavior remains unchanged.

**Acceptance Scenarios**:

1. **Given** a postcard with written text and two new image attachments, **When** the internal analysis workflow runs, **Then** each image receives structured image analysis and the postcard receives final structured understanding.
2. **Given** a postcard is created or changed by a normal client app user, **When** analysis is pending, processing, succeeded, or failed, **Then** the normal client app response shapes and behavior remain unchanged.

---

### User Story 2 - Reuse Understanding For Unchanged Media (Priority: P2)

As the Chronote backend system, I can reuse existing image analysis for media that has not changed, so repeated postcard analysis does not repeatedly process the same photo.

**Why this priority**: Reuse keeps the workflow efficient, reduces unnecessary AI processing, and makes re-analysis practical when postcard text changes but media remains stable.

**Independent Test**: Can be fully tested by analyzing a postcard, running analysis again without changing media, and confirming the existing image-level understanding is reused while the postcard-level understanding can still be refreshed.

**Acceptance Scenarios**:

1. **Given** a postcard has media already analyzed for the current media version and understanding contract version, **When** the postcard is analyzed again, **Then** the existing media analysis is reused instead of creating a duplicate image analysis result.
2. **Given** postcard text changes while media stays unchanged, **When** the postcard is re-analyzed, **Then** the system reuses unchanged media facts and produces a new postcard-level understanding for the changed postcard content.

---

### User Story 3 - Preserve Uncertainty And Failure States (Priority: P3)

As a future internal reviewer or product workflow, I can distinguish successful understanding, unavailable or failed analysis, and low-confidence interpretations, so stored AI output is useful without being treated as unquestioned truth.

**Why this priority**: Internal analysis must remain inspectable and trustworthy enough for later management review, search, moderation, quality inspection, tagging, recommendation, or analytics features.

**Independent Test**: Can be fully tested by forcing malformed, refused, unavailable, and low-confidence analysis outcomes and confirming the stored records capture status, confidence, uncertainty, and version context without exposing details to the normal client app.

**Acceptance Scenarios**:

1. **Given** structured AI output is malformed, **When** validation fails, **Then** the workflow performs one controlled repair attempt and stores either a valid repaired result or a failed state.
2. **Given** an image cannot be processed, **When** postcard-level understanding can still be produced from text and successful image results, **Then** the stored postcard understanding records the partial nature and confidence of the result.
3. **Given** analysis succeeds, **When** the result is stored, **Then** the stored understanding identifies its status, confidence, and version context for future inspection.

### Edge Cases

- A postcard has written text but no attached images; postcard-level understanding may still be stored from text alone.
- A postcard has images but little or no written text; the workflow stores image-level understanding and only claims postcard-level meaning supported by available evidence.
- One image fails analysis while other images succeed; the workflow preserves the failed image state and may produce partial postcard understanding from available text and successful image facts.
- AI output is malformed or incomplete; invalid output is not stored as successful understanding.
- The AI provider refuses or cannot process an image; the workflow stores an unavailable or failed state without exposing provider details to normal client app users.
- A temporary access link for private media expires during analysis; the workflow retries with a fresh internal access path when the failure is controlled and recoverable.
- Short-lived job coordination is unavailable; durable analysis state remains recoverable for later processing or inspection.
- Postcard text or media changes while analysis is running; stale results must not overwrite understanding for the newer postcard version.
- Prompt, schema, model, or understanding contract versions change; older results remain traceable and eligible for future re-analysis.

## Contract and Invariant Impact *(mandatory)*

### API Contract Impact

- Affected route groups: None for normal client app public routes in this phase.
- Request schema changes: None. Normal client app postcard creation, update, media, and read requests remain unchanged.
- Response schema changes: None. Normal client app postcard responses do not expose analysis status or analysis result fields in this phase.
- Error schema changes: None. Analysis pending, processing, succeeded, failed, or unavailable states must not change normal client app error behavior.

### Domain Invariant Impact

- Preserved invariants:
  - Normal client app users create and edit postcards as usual.
  - Postcard ownership, visibility, media grouping, and media access rules remain unchanged.
  - Private media remains private and is not exposed through public analysis endpoints.
  - Public Chronote API contract remains unchanged unless an explicit exception is approved.
  - AI interpretation is stored as an internal interpretation, not as absolute truth.
- New invariants:
  - Image-level and postcard-level understanding are stored separately.
  - Successful understanding must pass structured validation before it is stored as successful.
  - Stored understanding must include confidence or uncertainty indicators.
  - Stored understanding must include enough version context to identify the prompt, schema, and model family that produced it.
  - Unchanged media analysis can be reused only when the relevant media and understanding contract versions still match.
  - Analysis records for stale postcard or media versions must not overwrite newer-version understanding.
- Amended invariants: None for existing postcard or media user-facing behavior.

### Boundary Impact

- Modules touched: postcards, media, internal postcard understanding workflow, internal job/result storage, and platform integrations needed to run private analysis work.
- Persistence leakage check: Public request and response shapes remain separate from stored analysis records and provider output.
- Dependency isolation check: Business rules for eligibility, validation, reuse, status transitions, and stale-result prevention remain independent from transport handlers and external provider details.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST run postcard understanding as a backend-internal workflow without requiring normal client app participation.
- **FR-002**: The system MUST NOT expose client-facing analyze endpoints or client-facing analysis result endpoints in this phase.
- **FR-003**: The system MUST be able to start analysis from backend events or internal service calls, including postcard creation, postcard update, media change, scheduled backfill, and private operational retry.
- **FR-004**: The system MUST determine whether a postcard version is eligible for analysis before work is stored as active analysis.
- **FR-005**: The system MUST analyze attached images before producing final postcard-level understanding when images are present.
- **FR-006**: The system MUST produce structured image understanding that includes image type, caption, visible text when present, location clues when present, mood, notable details, confidence, and food/scenery-specific fields when applicable.
- **FR-007**: The system MUST combine original postcard text, image understanding results, visible image text, and useful postcard/media metadata before producing final postcard-level understanding.
- **FR-008**: The system MUST produce structured postcard understanding that includes summary, suggested title, tone or mood, topics or tags, people clues, place clues, food or travel memory meaning when supported, search keywords, and confidence.
- **FR-009**: The system MUST validate structured AI output before storing it as successful understanding.
- **FR-010**: The system MUST perform at most one controlled schema-repair retry when structured output is malformed.
- **FR-011**: The system MUST store image-level understanding and postcard-level understanding as separate records.
- **FR-012**: The system MUST reuse image understanding for unchanged media when the media version and understanding contract version still match.
- **FR-013**: The system MUST preserve normal client app postcard request and response behavior unchanged while analysis is pending, processing, succeeded, failed, or unavailable.
- **FR-014**: The system MUST store enough version metadata to identify the prompt version, schema version, and model family used for each result.
- **FR-015**: The system MUST avoid writing raw postcard text, raw photos, secret credentials, temporary private media links, and full provider payloads to normal application logs.
- **FR-016**: The system MUST store confidence and uncertainty information with AI interpretations and MUST avoid storing unsupported exact claims as facts.
- **FR-017**: The system MUST preserve durable status for pending, processing, succeeded, failed, unavailable, and stale analysis outcomes.
- **FR-018**: The system MUST prevent older analysis results from overwriting newer postcard or media understanding when content changes during analysis.

### Key Entities *(include if feature involves data)*

- **Analysis Job**: Durable record of internal work for a postcard version, including status, retry state, failure category, and the target postcard content version.
- **Media Analysis**: Structured understanding for one media item and media version, including observed facts, confidence, status, and version context.
- **Postcard Analysis**: Final structured understanding for one postcard version, derived from postcard text, media facts, visible image text, and useful metadata.
- **Prompt Version**: The versioned instruction set used to guide AI interpretation.
- **Schema Version**: The versioned structured output contract used to validate stored understanding.
- **Model Version**: The AI model family or model identifier recorded for traceability.
- **Analysis Status**: The lifecycle state of analysis work or results, such as pending, processing, succeeded, failed, unavailable, or stale.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of normal client app postcard request and response contracts remain unchanged while postcard understanding is pending, processing, succeeded, failed, or unavailable.
- **SC-002**: At least 95% of eligible postcards with written text and one to five attached food, scenery, people, object, or memory-related photos produce stored postcard-level understanding in a controlled test set.
- **SC-003**: 100% of malformed structured AI outputs are either repaired into valid structured understanding or stored as non-successful analysis outcomes.
- **SC-004**: 100% of successful stored results identify their status, confidence, prompt version, schema version, and model version.
- **SC-005**: 100% of repeated analysis attempts for unchanged media reuse existing image understanding when the media version and understanding contract version still match.
- **SC-006**: 0 raw AI provider payloads, secret credentials, temporary private media links, or raw image content appear in normal application logs during representative success and failure scenarios.
- **SC-007**: After a worker interruption or restart, pending and processing analysis jobs can be recovered, inspected, or safely retried without losing durable status.

## Assumptions

- Management-web result viewing will be designed later and is out of scope for this phase.
- Normal client app behavior must stay unchanged in this phase.
- No public analyze endpoint, public analysis status endpoint, or public analysis result endpoint will be added in this phase.
- Existing postcard and media data provides enough text, media identity, media version, object location, and ownership metadata for backend-internal analysis.
- User-uploaded postcard photos are private unless existing product visibility rules say otherwise.
- Stored AI output is for internal interpretation and future product use; it is not treated as a source of absolute truth.
- Future search, moderation, recommendation, analytics, and management review features may read stored understanding later, but those user-facing behaviors are out of scope for this phase.
