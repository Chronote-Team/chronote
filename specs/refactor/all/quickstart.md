# Quickstart: Chronote Backend Contract-Preserving Refactor

## 1. Review The Planning Inputs

Read the feature definition and baseline references before implementation:

- [spec.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/spec.md)
- [plan.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/plan.md)
- [research.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/research.md)
- [http-api.md](/home/bowen/Coding/chronote-refactor/specs/refactor/all/contracts/http-api.md)
- `/home/bowen/Coding/chronote/ai_generated/api_documentation.md`
- `/home/bowen/Coding/chronote/docs/superpowers/plans/2026-04-20-chronote-refactor-replacement.md`

## 2. Bootstrap The Root-Level Replacement Workspace

From the repository root:

```bash
mkdir -p cmd/api internal/platform internal/shared internal/modules tests migrations
go mod init chronote-refactor
go mod tidy
```

Create the planned directory skeleton at the repository root before adding slice implementations.

## 3. Implement In Slice Order

Work in this sequence:

1. Bootstrap app and shared response/error primitives
2. Platform adapters for config, Postgres, Redis, S3, JWT, and password services
3. Health slice
4. Users and auth slices
5. Postcards slice
6. Media slice
7. Cutover verification

## 4. Verify Continuously

Run narrow tests as each slice lands. In this workspace, prefer the offline-safe form because the module cache is already populated locally:

```bash
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/shared/... -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/platform/... -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/modules/health/... ./tests/contract/... -run 'Health' -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/modules/users/... ./internal/modules/auth/... ./tests/contract/... -run 'User|Auth|Register|Login|Refresh|Logout' -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/modules/postcards/... ./tests/contract/... -run 'Postcard' -v
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/modules/media/... ./tests/contract/... -run 'Media' -v
```

Before cutover readiness, run the full verification suite:

```bash
env GOCACHE=/tmp/go-build GOPROXY=off go test ./... -v
```

For US3 cutover verification with isolated Podman-backed PostgreSQL and Redis:

```bash
podman network create chronote_us3_test
podman run --replace -d --name chronote_us3_pg --network chronote_us3_test -p 55432:5432 \
  -e POSTGRES_DB=chronote -e POSTGRES_USER=chronote -e POSTGRES_PASSWORD=chronote \
  docker.io/library/postgres:16-alpine
podman run --replace -d --name chronote_us3_redis --network chronote_us3_test -p 56379:6379 \
  docker.io/library/redis:8 redis-server --appendonly yes --requirepass chronote

env CHRONOTE_CUTOVER_TESTS=1 \
  POSTGRES_HOST=127.0.0.1 POSTGRES_PORT=55432 POSTGRES_USER=chronote POSTGRES_PASSWORD=chronote POSTGRES_DB=chronote \
  POSTGRES_SSLMODE=disable POSTGRES_TIMEZONE=Asia/Shanghai \
  REDIS_HOST=127.0.0.1 REDIS_PORT=56379 REDIS_PASSWORD=chronote REDIS_DB=0 \
  JWT_SIGNING_KEY=cutover-test-signing-key ACCESS_TOKEN_EXPIRE=7200 REFRESH_TOKEN_EXPIRE=1814400 \
  GOCACHE=/tmp/go-build GOPROXY=off \
  go test ./tests/integration -run 'Cutover|FullStack|HealthDetailsDegrades' -v
```

These container-backed tests should run on a dedicated network such as `chronote_us3_test` to avoid interacting with unrelated local containers.

## 5. Guardrails

- Do not modify `/home/bowen/Coding/chronote`; use it only as a compatibility reference.
- Preserve the public endpoint contract, status behavior, and default error text unless a documented exception is approved.
- Keep handlers transport-only and infrastructure details out of domain/application logic.
- Do not redesign schema semantics during this feature.
