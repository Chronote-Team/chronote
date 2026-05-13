# Quickstart: Random Accessible Postcard

## Prerequisites

- Work from `/home/bowen/Coding/chronote-refactor`.
- Keep `/home/bowen/Coding/chronote` read-only.
- Use the existing `feature/random` branch and feature directory `specs/feature/random`.

## Implementation Order

1. Update `internal/modules/postcards/app/repository.go` with `FindRandomAccessible(userID uint)`.
2. Add `Service.GetRandom(userID uint)` in `internal/modules/postcards/app/service.go`.
3. Add random accessible selection to the in-memory repository in `service.go`.
4. Add random accessible selection to `internal/modules/postcards/infra/gorm_repository.go`.
5. Add `GetRandomPostcard` in `internal/modules/postcards/http/handler.go`.
6. Register `GET /random` before `GET /:id` in `internal/modules/postcards/http/routes.go`.
7. Add app-level random filtering tests in `internal/modules/postcards/app/service_test.go`.
8. Add contract tests in `tests/contract/postcard_contract_test.go`.

## Contract Checks

- Anonymous caller with at least one public postcard receives `200 OK` and message `获取随机明信片成功`.
- Anonymous caller never receives a private postcard.
- Signed-in caller can receive their own private postcard when it is the only accessible candidate.
- No accessible postcard returns `404 Not Found` and message `明信片不存在`.
- `/v1/postcards/random` is not handled by the detail route.

## Verification Commands

Focused postcard check:

```bash
env GOCACHE=/tmp/go-build GOPROXY=off go test ./internal/modules/postcards/... ./tests/contract/... -run 'Postcard|Random' -v
```

Full offline-safe check:

```bash
env GOCACHE=/tmp/go-build GOPROXY=off go test ./... -v
```

## Expected Route Behavior

```http
GET /v1/postcards/random
Authorization: Bearer <access_token>
```

The `Authorization` header is optional. A valid access token expands the candidate set to include the caller's own postcards. Missing or invalid optional auth context must not expose private postcards owned by other users.
