# Internal Workflow Contract: Postcard AI Understanding

## Contract Scope

This feature adds backend-internal workflow contracts only. It does not add public client-facing HTTP endpoints, public analysis status endpoints, or public analysis result fields.

## Public API Non-Change Contract

The following must remain true for normal client app behavior:

- No `POST /postcards/:id/analyze` endpoint is added.
- No `GET /postcards/:id/analysis` endpoint is added.
- No analysis status or result fields are added to normal postcard read/write responses.
- Existing postcard and media request schemas remain unchanged.
- Existing postcard and media response schemas remain unchanged.
- Existing postcard and media error behavior remains unchanged when analysis is pending, processing, succeeded, failed, unavailable, or stale.

## Internal Application Methods

### EnqueuePostcardAnalysis

**Purpose**: Request analysis for a postcard after postcard creation, postcard update, media change, scheduled backfill, or private operational retry.

**Input**:
- `postcard_id`
- `reason`: create, update, media_change, backfill, retry
- `requested_by`: system or private operator context

**Output**:
- Existing active job when an equivalent job already exists
- New pending job when analysis is eligible and not already queued
- No-op result when analysis is disabled or the postcard is ineligible

**Rules**:
- Must not block the normal client app mutation on AI provider calls.
- Must create durable state before relying on short-lived coordination.
- Must target the current postcard version.

### RunNextAnalysisJob

**Purpose**: Let an internal worker claim and process one due job.

**Input**:
- Worker identity
- Current time
- Optional processing limit context

**Output**:
- No due work
- Succeeded job
- Failed/unavailable/stale job
- Pending retry with next run time

**Rules**:
- Must acquire coordination before processing when coordination is available.
- Must recover safely if coordination is unavailable and durable state indicates due work.
- Must not write successful results for stale postcard or media versions.
- Must reuse successful media analysis when the reusable version key matches.

### RetryAnalysisJob

**Purpose**: Schedule a private operational retry for a failed or unavailable job.

**Input**:
- `job_id`
- Retry reason
- Private operator or scheduled maintenance context

**Output**:
- Pending job when retry is accepted
- Rejected retry when job state, version, or retry policy does not allow retry

**Rules**:
- Must not expose retry behavior through normal client app endpoints.
- Must preserve prior failure records for inspection.

## Provider Boundary

### AIClient

**Purpose**: Analyze image inputs and postcard context through a replaceable AI provider gateway.

**Required capabilities**:
- Analyze one image into structured image understanding.
- Analyze one postcard context into structured postcard understanding.
- Return safe error categories for refusal, unavailable provider, timeout, malformed output, and permanent input failure.

**Rules**:
- Application logic must not construct provider-specific HTTP payloads.
- Provider credentials must come from deployment secrets/configuration.
- Provider request/response payloads must not be logged in full.
- Malformed structured output may trigger at most one repair attempt.

## Storage Boundary

### Private Media Access

**Purpose**: Provide temporary access to private media for AI image understanding.

**Rules**:
- Generate temporary media access only inside backend workers.
- Never return analysis temporary links to the normal client app.
- Never store temporary links in durable job/result records.
- Regenerate temporary links on retry.
- Keep media storage private.

## Persistence Contract

### Durable Job State

Jobs must preserve:
- Target postcard version
- Status
- Attempts
- Next run time
- Lock/processing metadata
- Safe categorized error code

### Result State

Media and postcard results must preserve:
- Target media or postcard version
- Prompt version
- Schema version
- Model version
- Status
- Valid structured result when successful
- Confidence and uncertainty
- Safe categorized error code when unsuccessful

### Uniqueness

- Media analysis uniqueness is based on `media_id`, `media_version`, `prompt_version`, `schema_version`, and `model_version`.
- Postcard analysis uniqueness is based on `postcard_id`, `postcard_version`, `prompt_version`, `schema_version`, and `model_version`.

## Verification Expectations

- Contract tests must prove existing public postcard/media endpoints still match prior request and response behavior.
- Contract tests must prove no public analysis endpoints exist in this phase.
- Integration tests must prove durable jobs/results can be recovered after worker interruption.
- Privacy tests must prove logs exclude raw postcard text, raw image content, API keys, temporary private media links, and full provider payloads.
