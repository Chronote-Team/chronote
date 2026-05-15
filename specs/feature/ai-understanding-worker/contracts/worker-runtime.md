# Worker Runtime Contract: Postcard AI Understanding Worker

## Contract Scope

This feature adds backend-internal runtime contracts only. It does not add public client-facing HTTP endpoints, public analysis status endpoints, or public analysis result fields.

## Public API Non-Change Contract

The following must remain true:

- No public analyze endpoint is added.
- No public analysis read endpoint is added.
- Normal postcard and media request schemas remain unchanged.
- Normal postcard and media response schemas remain unchanged.
- Normal postcard and media error behavior remains unchanged when background analysis is pending, processing, succeeded, failed, unavailable, or stale.
- The API service continues to serve `/health`; the worker exposes no HTTP port.

## Worker Process Contract

### Startup

**Inputs**:
- Existing database, Redis, S3, JWT, media, and AI configuration values
- Worker-specific optional settings:
  - `AI_WORKER_ID`
  - `AI_WORKER_IDLE_SLEEP`
  - `AI_WORKER_ERROR_SLEEP`
  - `AI_WORKER_RUN_ONCE`

**Rules**:
- The worker loads the same project configuration mechanism as the API service.
- The worker builds the existing postcard AI application service through platform wiring.
- The worker must fail fast on unrecoverable dependency construction errors.
- The worker must not start an HTTP router.

### Loop

**Behavior**:

```text
while context is active:
  call RunNextAnalysisJob(ctx, worker_id)
  if context canceled:
    exit
  if an error occurred:
    log safe summary
    sleep error interval
    continue
  if no job was claimed:
    sleep idle interval
    continue
  log safe job outcome
  continue immediately
```

**Rules**:
- One call processes at most one job.
- Sleep occurs only after no due work or after an error.
- Successful, stale, unavailable, failed, and retryable outcomes are logged safely.
- `run_once` exits after one loop attempt.
- Shutdown on `SIGINT` or `SIGTERM` cancels the loop promptly.

## Provider Parsing Contract

### Successful Responses

The OpenAI Responses-compatible adapter must extract structured JSON from supported response locations including:

- `output[].content[].text`
- text or JSON-like variants inside `output[].content[]`
- `output_text`

**Rules**:
- Response body reading must be bounded.
- Markdown JSON code fences are trimmed.
- Extracted output must parse as valid JSON.
- Parsed output is returned as the internal AI result with confidence and uncertainty extracted when present.
- Missing or malformed structured output is classified as malformed output.

### Error Responses

The existing provider classification must be preserved:

- `429` and `5xx`: temporary provider unavailability
- `400` and `403`: provider refusal or permanent input failure
- unsupported success shape: malformed output

Full provider request or response payloads must not be logged.

## Compose Runtime Contract

The local Compose deployment must contain:

- Existing `app` service running the API binary
- New `worker` service using the same project image
- Worker command set to the worker binary
- Same database, Redis, S3, and AI environment configuration as `app`
- Dependencies on healthy PostgreSQL, healthy Redis, started object storage, and completed bucket initialization
- No worker port mapping

## Verification Contract

Required verification:

- A queued job moves out of `pending` when the worker is running.
- Text-only postcard analysis can succeed with a controlled provider.
- Image postcard analysis can store media and postcard results with a controlled provider.
- Public API contract tests still prove no analysis endpoints or fields exist.
- Logs show job IDs and outcomes without secrets or raw user/provider data.
