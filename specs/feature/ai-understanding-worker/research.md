# Research: Postcard AI Understanding Worker

## Decision 1: Use A Separate Worker Runtime

- **Decision**: Add a `cmd/worker` process that runs postcard AI understanding work separately from `cmd/api`.
- **Rationale**: The current API process already enqueues durable analysis jobs after postcard/media changes. Running provider calls inside the API request path would block normal Chronote behavior and violate the internal-only design. A separate worker process lets jobs move out of `pending` without changing public HTTP behavior.
- **Alternatives considered**:
  - Call `RunNextAnalysisJob` from API startup in a background goroutine: rejected because it couples HTTP process lifecycle and background work in one binary and makes deployment scaling less explicit.
  - Process one job synchronously after each mutation: rejected because provider latency and failures would affect normal user requests.
  - Add a public analyze endpoint: rejected because public analysis endpoints are explicitly out of scope.

## Decision 2: Wrap Existing `RunNextAnalysisJob`

- **Decision**: Build the worker loop around the existing `postcardai/app.Service.RunNextAnalysisJob(ctx, workerID)` method.
- **Rationale**: That method already owns job claim, stale checks, media reuse, provider calls, validation/repair, result storage, and status transitions. The new process should own only lifecycle concerns: config, dependency construction, signal handling, worker identity, sleep intervals, and safe outcome logging.
- **Alternatives considered**:
  - Reimplement job processing in the worker command: rejected because it would duplicate business rules and break module boundaries.
  - Move job processing into platform code: rejected because platform code should wire dependencies, not own domain workflow.

## Decision 3: Add Minimal Worker Options

- **Decision**: Support `WorkerID`, `IdleSleep`, `ErrorSleep`, and `RunOnce` options with environment-backed defaults.
- **Rationale**: Operators need a stable worker identifier for job claims/logs, idle sleep prevents busy loops, error sleep prevents tight failure loops, and run-once mode makes deterministic tests and operational debugging easier. Defaults should be enough for normal local deployment.
- **Alternatives considered**:
  - Hard-code all options: rejected because logs and multi-worker testing need distinct worker IDs and tunable sleeps.
  - Add a broad scheduler configuration surface: rejected because the first worker branch only needs a simple durable queue consumer.

## Decision 4: Share Analysis Platform Construction

- **Decision**: Extract analysis dependency construction into `internal/platform/app` so both API and worker use the same repositories, source adapters, S3 presigner, and AI provider config.
- **Rationale**: `internal/platform/app/app.go` currently wires the analysis service for enqueue hooks, but the worker also needs the same service. Shared construction avoids drift and keeps concrete dependencies in platform code as required by the constitution.
- **Alternatives considered**:
  - Duplicate wiring in `cmd/worker`: rejected because command packages should be thin entrypoints and dependency drift would be likely.
  - Move construction into `postcardai/app`: rejected because concrete PostgreSQL/S3/OpenAI dependencies do not belong in application logic.

## Decision 5: Parse OpenAI Responses Success Bodies In The Adapter

- **Decision**: Complete the OpenAI Responses adapter so successful HTTP responses are parsed into `postcardaiapp.AIResult`.
- **Rationale**: The current adapter sends requests but returns `response parsing not configured for live provider` for successful responses. Real API verification requires successful provider responses to become validated internal JSON while preserving provider-specific parsing inside the infra adapter.
- **Alternatives considered**:
  - Parse provider payloads in application service: rejected because app logic should not depend on provider response shapes.
  - Store raw provider responses for later parsing: rejected because it leaks provider payloads and may include user content.

## Decision 6: Use Existing Database Schema For This Branch

- **Decision**: Do not plan new database tables for the worker branch. Use existing `ai_analysis_jobs`, `media_ai_analysis`, and `postcard_ai_analysis`.
- **Rationale**: The missing behavior is runtime consumption and provider parsing, not data persistence. Existing result and job tables already model durable queue state and internal analysis output.
- **Alternatives considered**:
  - Add a separate worker heartbeat table: deferred until operational monitoring needs are proven.
  - Add a separate provider-call audit table: rejected for this branch because privacy constraints and current requirements only need safe logs/outcomes.

## Decision 7: Build One Image With Two Binaries

- **Decision**: Build `/app/chronote` and `/app/chronote-worker` into the same container image and run the worker through Compose `command`.
- **Rationale**: API and worker must run from the same source revision and config surface. One image keeps deployment simple and lets Compose start separate services without maintaining parallel image tags.
- **Alternatives considered**:
  - Separate Dockerfiles/images: rejected as unnecessary operational complexity for the current branch.
  - Replace the API entrypoint with a multi-mode binary: rejected because separate commands are clearer and preserve the current API entrypoint.

## Decision 8: Preserve Public API By Contract Tests

- **Decision**: Keep all AI understanding status/results internal and verify public non-change with contract tests.
- **Rationale**: The user-facing Chronote app should not see analysis fields or endpoints in this branch. Contract tests guard against accidental router or DTO drift while worker code changes platform wiring.
- **Alternatives considered**:
  - Add temporary debug HTTP endpoints: rejected because they would violate the explicit public scope boundary.
  - Rely on manual inspection: rejected because response-shape drift is easy to miss.
