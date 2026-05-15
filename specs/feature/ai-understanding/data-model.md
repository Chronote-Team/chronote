# Data Model: Postcard AI Understanding

## Analysis Job

**Description**: Durable record of internal work to analyze one postcard content version.

**Core fields**:
- `id`: Stable job identifier
- `postcard_id`: Target postcard
- `postcard_version`: Version of postcard content/media set this job targets
- `status`: Pending, processing, succeeded, failed, unavailable, or stale
- `attempts`: Number of processing attempts
- `next_run_at`: Earliest time the job may run again
- `locked_at`: Time a worker claimed the job, if currently processing
- `last_error_code`: Safe categorized error value, if any
- `created_at`
- `updated_at`

**Validation rules**:
- A job targets exactly one postcard version.
- Jobs for stale postcard versions must not write successful current results.
- Retryable failures must respect retry/backoff policy.
- Schema-repair retry is limited to one repair attempt per malformed output event.

**Relationships**:
- Belongs to one postcard.
- Produces zero or one current postcard analysis for the targeted postcard version.
- May reference many media analyses through the postcard's media set.

**State transitions**:
- `pending` -> `processing`
- `processing` -> `succeeded`
- `processing` -> `failed`
- `processing` -> `unavailable`
- `processing` -> `stale`
- `processing` -> `pending` when a controlled retry is scheduled
- `failed` or `unavailable` -> `pending` only through private operational retry/backfill

## Media Analysis

**Description**: Structured understanding result for one media item and media version.

**Core fields**:
- `id`: Stable analysis identifier
- `media_id`: Target media item
- `media_version`: Version of the media content analyzed
- `prompt_version`: Image understanding prompt version
- `schema_version`: Image understanding schema version
- `model_version`: AI model version used for the result
- `status`: Succeeded, failed, unavailable, or stale
- `result`: Structured image understanding when successful
- `confidence`: Overall confidence score or band
- `uncertainty`: Notes about ambiguity or unsupported inferences
- `error_code`: Safe categorized error value, if any
- `created_at`
- `updated_at`

**Structured result fields**:
- `image_type`
- `caption`
- `visible_text`
- `location_clues`
- `mood`
- `notable_details`
- `food_details` when applicable
- `scenery_details` when applicable
- `confidence`

**Validation rules**:
- Successful media analysis must validate against the active image understanding schema.
- A media analysis is reusable only when media version, prompt version, schema version, and model version match the requested analysis context.
- Raw image bytes, signed URLs, and full provider payloads are never stored in this entity.

**Relationships**:
- Belongs to one media item.
- May contribute to many postcard analysis runs when reused.

**Uniqueness**:
- `media_id + media_version + prompt_version + schema_version + model_version` must uniquely identify the reusable analysis context.

## Postcard Analysis

**Description**: Final structured understanding for one postcard version, derived from postcard text, image facts, visible image text, and useful metadata.

**Core fields**:
- `id`: Stable analysis identifier
- `postcard_id`: Target postcard
- `postcard_version`: Version of the postcard content/media set analyzed
- `prompt_version`: Postcard understanding prompt version
- `schema_version`: Postcard understanding schema version
- `model_version`: AI model version used for the result
- `status`: Succeeded, failed, unavailable, or stale
- `result`: Structured postcard understanding when successful
- `confidence`: Overall confidence score or band
- `uncertainty`: Notes about ambiguity, partial evidence, or unsupported inferences
- `error_code`: Safe categorized error value, if any
- `created_at`
- `updated_at`

**Structured result fields**:
- `summary`
- `suggested_title`
- `tone_or_mood`
- `topics_or_tags`
- `people_clues`
- `place_clues`
- `food_memory_meaning` when supported
- `travel_memory_meaning` when supported
- `search_keywords`
- `confidence`

**Validation rules**:
- Successful postcard analysis must validate against the active postcard understanding schema.
- Final understanding may be partial when one or more image analyses failed, but uncertainty must identify the partial evidence.
- Unsupported exact claims must not be stored as facts.
- Stale analysis must not overwrite current postcard-version analysis.

**Relationships**:
- Belongs to one postcard.
- References the image facts available for the analyzed postcard version.

**Uniqueness**:
- `postcard_id + postcard_version + prompt_version + schema_version + model_version` must uniquely identify the final analysis context.

## Prompt Version

**Description**: Versioned instruction set used to guide image or postcard understanding.

**Core fields**:
- `name`
- `version`
- `purpose`: Image understanding or postcard understanding
- `content_hash`

**Validation rules**:
- Stored analysis must record the prompt version that produced it.
- Prompt changes create a new reusable analysis context.

## Schema Version

**Description**: Versioned structured output contract accepted by the backend.

**Core fields**:
- `name`
- `version`
- `purpose`: Image understanding or postcard understanding
- `content_hash`

**Validation rules**:
- Successful output must validate against the schema version recorded with the result.
- Schema changes create a new reusable analysis context.

## Model Version

**Description**: AI model family or provider model identifier used for traceability.

**Core fields**:
- `provider`
- `model`
- `version_label`

**Validation rules**:
- Stored analysis must record the model version that produced it.
- Model changes create a new reusable analysis context unless explicitly approved otherwise.

## Analysis Status

**Description**: Lifecycle state for jobs and stored results.

**Allowed values**:
- `pending`: Work is waiting to run.
- `processing`: A worker has claimed the work.
- `succeeded`: Valid structured understanding is stored.
- `failed`: Controlled failure occurred and no successful result was produced.
- `unavailable`: Provider or input refusal/unavailability prevented analysis.
- `stale`: The target postcard or media version is no longer current.

**Validation rules**:
- Only `succeeded` results may be reused as successful understanding.
- `failed`, `unavailable`, and `stale` states must preserve safe error categories without exposing raw provider payloads.

## Provider Call Metadata

**Description**: Safe operational metadata for observability and debugging.

**Core fields**:
- `job_id`
- `postcard_id`
- `media_ids`
- `provider`
- `model_version`
- `prompt_version`
- `schema_version`
- `status`
- `latency_ms`
- `retry_count`
- `usage_summary` when safe and available
- `error_code`

**Validation rules**:
- Must not include API keys, signed URLs, raw postcard text, raw image bytes, full AI provider payloads, or raw provider error bodies that may contain user data.
