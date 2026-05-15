# Data Model: Postcard AI Understanding Worker

## Analysis Job

**Description**: Existing durable work item queued when postcard or media content changes.

**Existing fields used by worker**:
- `id`: Stable job identifier used in logs and status updates
- `postcard_id`: Target postcard
- `postcard_version`: Target content/media version
- `status`: Pending, processing, succeeded, failed, unavailable, or stale
- `attempts`: Number of processing attempts
- `next_run_at`: Earliest eligible retry/processing time
- `locked_at`: Claim timestamp when processing
- `worker_id`: Worker identity that claimed the job when available
- `last_error_code`: Safe categorized error code
- `created_at`
- `updated_at`

**Worker rules**:
- Only due pending work may be claimed.
- A claimed job must either reach a terminal/stale/unavailable state or be scheduled for retry.
- Work for an old postcard version must become stale and must not store current postcard results.
- Multiple workers must not process the same job at the same time.

## Worker Runtime Options

**Description**: Process-level settings for the background worker loop.

**Fields**:
- `worker_id`: Stable identifier used for claims and safe logs; default `worker-1`
- `idle_sleep`: Delay after no due job is found; default small interval such as `2s`
- `error_sleep`: Delay after unexpected processing error; default larger interval such as `5s`
- `run_once`: Boolean that processes at most one loop iteration for deterministic tests or operational checks

**Validation rules**:
- Empty worker ID falls back to a safe default.
- Non-positive sleep values fall back to defaults.
- `run_once` must exit after one attempt regardless of whether work was found.

## Worker Outcome

**Description**: Safe process-level summary returned after one call to the existing job processor.

**Fields**:
- `job_id`: Processed job identifier, when a job was claimed
- `status`: Resulting job status, when a job was claimed
- `has_job`: Derived from whether `job_id` is present
- `error_category`: Safe categorized process/provider/storage error for logs when available

**Rules**:
- Logs may include job ID, status, worker ID, and safe error category.
- Logs must not include raw postcard text, raw image bytes, signed URLs, API keys, or full provider payloads.

## Media Analysis

**Description**: Existing image-level internal result reused or created by `RunNextAnalysisJob`.

**Fields used by worker**:
- `media_id`
- `media_version`
- `prompt_version`
- `schema_version`
- `model_version`
- `status`
- `result`
- `confidence`
- `uncertainty`
- `error_code`
- `created_at`
- `updated_at`

**Worker rules**:
- Successful reusable analysis is selected only when media version, prompt version, schema version, and model version match.
- Failed/unavailable image analysis may contribute partial uncertainty to postcard-level analysis.
- Signed URLs and raw image bytes are never stored.

## Postcard Analysis

**Description**: Existing postcard-level internal result created after postcard text and available media facts are analyzed.

**Fields used by worker**:
- `postcard_id`
- `postcard_version`
- `prompt_version`
- `schema_version`
- `model_version`
- `status`
- `result`
- `confidence`
- `uncertainty`
- `error_code`
- `created_at`
- `updated_at`

**Worker rules**:
- Successful output must validate against the active postcard understanding schema before storage.
- Partial evidence must be represented through uncertainty rather than unsupported facts.
- Stale postcard versions must not overwrite current analysis.

## Provider Parsed Result

**Description**: In-memory structured output returned by the provider adapter to application logic.

**Fields**:
- `json`: Valid JSON document extracted from a successful provider response
- `confidence`: Numeric confidence when present in the structured output, otherwise default `0`
- `uncertainty`: Uncertainty text when present in the structured output, otherwise empty

**Validation rules**:
- Provider output must be valid JSON before becoming a parsed result.
- Markdown code fences around JSON are trimmed before validation.
- Missing supported text output is classified as malformed output.
- Full provider response bodies are not logged or stored.

## Deployment Service

**Description**: Runtime service definition for the worker in local container deployment.

**Fields**:
- `image`: Same project image as the API service
- `command`: Worker binary path
- `environment`: Same database, Redis, S3, and AI configuration surface used by API
- `depends_on`: Database, Redis, object storage, and bucket initialization dependencies
- `ports`: None

**Rules**:
- Worker service must not expose HTTP ports.
- API service remains responsible for `/health`.
- Worker can run as one or more replicas if job claim semantics stay exclusive.
