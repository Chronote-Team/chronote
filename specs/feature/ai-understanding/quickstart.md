# Quickstart: Postcard AI Understanding

## 1. Review The Planning Inputs

Read these files before implementation:

- [spec.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding/spec.md)
- [plan.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding/plan.md)
- [research.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding/research.md)
- [data-model.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding/data-model.md)
- [internal-workflow.md](/home/bowen/Coding/chronote-refactor/specs/feature/ai-understanding/contracts/internal-workflow.md)
- [docs/postcard-ai-understanding-final.md](/home/bowen/Coding/chronote-refactor/docs/postcard-ai-understanding-final.md)

## 2. Build In Slice Order

Use this implementation sequence:

1. Domain types and migrations for jobs/results
2. Application services and dependency interfaces
3. Provider gateway, prompts, schemas, and validation
4. Worker runner and platform wiring
5. Postcard/media enqueue hooks
6. Private retry/backfill hooks
7. Contract, integration, unit, and privacy verification

## 3. Configuration

Plan for these runtime settings:

```text
AI_ENABLED
AI_PROVIDER
AI_ENDPOINT_TYPE
AI_MODEL
AI_TIMEOUT
OPENAI_API_KEY
```

Rules:

- API keys come from deployment secrets or environment configuration only.
- Do not commit API keys, provider payload examples containing user content, signed URLs, or raw postcard/media data.
- Keep AI disabled by default in local/test app wiring unless a test explicitly enables a fake provider.
- For local tests, use `postcardaiapp.NoopAIClient` and `postcardaiapp.NoopStorage` or fakes through `postcardaiapp.NewService`.
- Public HTTP test apps should keep `AI_ENABLED=false` unless the test is specifically exercising internal worker behavior.

## 4. Migrations

Add durable storage for:

```text
ai_analysis_jobs
media_ai_analysis
postcard_ai_analysis
```

Required constraints:

- Unique media analysis key: media ID, media version, prompt version, schema version, model version
- Unique postcard analysis key: postcard ID, postcard version, prompt version, schema version, model version
- Status fields support pending, processing, succeeded, failed, unavailable, and stale where applicable

## 5. Verification Commands

Run focused tests as slices land:

```bash
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/modules/postcardai/... -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/modules/postcards/... ./internal/modules/media/... -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./tests/contract/... -run 'Postcard|Media|Analysis' -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./tests/integration/... -run 'PostcardAI|Analysis|Worker' -v
```

Before handoff, run:

```bash
env GOCACHE=/tmp/go-build GOPROXY=off go test ./... -v
```

## 6. Public API Guardrails

- Do not add `POST /postcards/:id/analyze`.
- Do not add `GET /postcards/:id/analysis`.
- Do not add analysis status/result fields to normal postcard responses.
- Do not make normal client app reads or writes wait for AI provider calls.
- Preserve public Chronote API response shape, status behavior, and default error text.

## 7. Privacy Guardrails

- Do not log raw postcard text.
- Do not log raw image bytes.
- Do not log full AI provider payloads.
- Do not log API keys.
- Do not log temporary private media links.
- Do not store temporary private media links in database rows.
