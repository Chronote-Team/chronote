# Quickstart: Postcard AI Understanding Worker

## 1. Read Planning Inputs

Read these files before implementation:

- [spec.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding-worker/spec.md)
- [plan.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding-worker/plan.md)
- [research.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding-worker/research.md)
- [data-model.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding-worker/data-model.md)
- [worker-runtime.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding-worker/contracts/worker-runtime.md)
- [docs/postcard-ai-understanding-worker.md](/home/bowen/Coding/chronote-refactor/docs/postcard-ai-understanding-worker.md)

## 2. Build In Slice Order

Use this implementation sequence:

1. Worker loop options and tests
2. `cmd/worker` signal-aware entrypoint
3. Shared platform analysis construction
4. OpenAI Responses parser tests and parser implementation
5. Dockerfile and Compose worker service
6. Contract and integration verification
7. Local deployment smoke test

## 3. Worker Configuration

Optional worker-specific settings:

```text
AI_WORKER_ID=worker-1
AI_WORKER_IDLE_SLEEP=2s
AI_WORKER_ERROR_SLEEP=5s
AI_WORKER_RUN_ONCE=false
```

Existing AI settings used by both API and worker:

```text
AI_ENABLED=true
AI_PROVIDER=openai
AI_ENDPOINT_TYPE=responses
AI_ENDPOINT=https://api.openai.com/v1/responses
AI_MODEL=gpt-4.1-mini
AI_TIMEOUT=30
OPENAI_API_KEY=your_key_here
```

Rules:

- Do not commit API keys.
- Keep public API tests independent of real providers.
- Use fake providers for deterministic integration tests.
- Local real-provider image tests need provider-reachable image URLs; local RustFS service names are not reachable from OpenAI.

## 4. Expected Local Compose Shape

After implementation, Compose should run:

```text
chronote-postgres
chronote-redis
chronote-rustfs
chronote-s3-init
chronote-app
chronote-worker
```

The API remains reachable through:

```bash
curl http://127.0.0.1:58080/health
```

The worker should not expose a port. Check worker logs with:

```bash
podman compose logs worker
```

## 5. Verification Commands

Run focused tests while implementing:

```bash
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/platform/app -run 'Worker|Analysis' -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/modules/postcardai/... -run 'OpenAI|Responses|Worker|Analysis' -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./tests/contract/... -run 'Postcard|Media|Analysis' -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./tests/integration/... -run 'PostcardAI|Worker|Analysis' -v
```

Before handoff, run:

```bash
env GOCACHE=/tmp/go-build GOPROXY=off go test ./... -v
```

## 6. Local Smoke Test

1. Start the stack:

```bash
podman compose up -d --build
```

2. Confirm API health:

```bash
curl http://127.0.0.1:58080/health
```

3. Register and log in through the existing public API.
4. Create or update multiple postcards so analysis jobs are enqueued.
5. Inspect worker logs for safe job claim/outcome summaries.
6. Inspect the database and confirm new jobs no longer remain `pending` forever.
7. Confirm normal postcard read responses still omit AI status/result fields.

## 7. Public API Guardrails

- Do not add a public analyze endpoint.
- Do not add a public analysis result endpoint.
- Do not add AI status/result fields to normal postcard responses.
- Do not make normal postcard/media requests wait for AI provider calls.
- Preserve public status codes and default error text.

## 8. Privacy Guardrails

- Do not log raw postcard text.
- Do not log raw image bytes.
- Do not log signed URLs.
- Do not log API keys.
- Do not log full provider request or response payloads.
