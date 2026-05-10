# Implementation Plan: Random Accessible Postcard

**Branch**: `feature/random` | **Date**: 2026-05-10 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/home/bowen/Coding/chronote-refactor/specs/feature/random/spec.md`

**Setup Note**: `.specify/scripts/bash/setup-plan.sh --json` was run from the repository root, but it rejected the active branch name `feature/random` even though the script's shared branch validator accepts conventional `feature/*` branches. The plan continues against the user-requested feature directory `/home/bowen/Coding/chronote-refactor/specs/feature/random`.

## Summary

Add `GET /v1/postcards/random` as a public postcard read path that supports optional access-token context. The endpoint returns one randomly selected postcard from the caller's accessible set, preserves the existing postcard response envelope and relation loading behavior, and registers `/random` before `/:id` so the literal route is not treated as a postcard identifier.

Implementation follows the existing postcard vertical slice:

```text
internal/modules/postcards/http
  -> internal/modules/postcards/app
  -> internal/modules/postcards/infra
```

The handler reads optional auth context and writes the response, the app service owns access behavior and relation attachment, and repository implementations own random candidate selection.

## Technical Context

**Language/Version**: Go 1.25  
**Primary Dependencies**: Gin, GORM, PostgreSQL driver, existing Chronote shared response/error helpers, existing auth middleware  
**Storage**: PostgreSQL for production postcard data; in-memory repositories for test app wiring  
**Testing**: `env GOCACHE=/tmp/go-build GOPROXY=off go test ./... -v`, plus focused postcard app and contract tests  
**Target Platform**: Linux backend service  
**Project Type**: Web service API  
**Performance Goals**: Return one eligible postcard in a single read flow without adding extra client round trips; preserve existing list/detail response relation behavior  
**Constraints**: Preserve public API contract style, optional auth behavior, exact default messages from the source plan, module boundaries, and route ordering before `/:id`  
**Scale/Scope**: One new postcard read endpoint, one repository interface addition, production and in-memory repository support, focused unit and contract coverage

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Planned structure uses `cmd/api`, `internal/platform`, `internal/shared`,
      `internal/modules/<module>`, `migrations`, `tests`, and `docs`, or records
      an equivalent boundary-preserving alternative.
- [x] Each affected module keeps `http`, `app`, `domain`, and `infra`
      responsibilities separate; business logic is not placed in handlers or
      adapters.
- [x] API changes define stable request, response, and error contracts with
      explicit validation and no persistence-model leakage.
- [x] External dependencies are injected behind interfaces where business logic
      touches them, and concrete wiring stays in `internal/platform`.
- [x] Preserved or changed domain invariants are identified explicitly.
- [x] Unit, integration, and contract test coverage is planned for every changed
      behavior that requires it under the constitution.

## Project Structure

### Documentation (this feature)

```text
specs/feature/random/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── random-postcard.openapi.yaml
├── checklists/
│   └── requirements.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/
├── platform/
│   ├── app/
│   │   └── app.go                    # existing wiring remains compatible
│   └── http/
│       └── router.go                 # existing OptionalAccessToken postcard group
└── modules/
    └── postcards/
        ├── domain/
        │   └── postcard.go           # no domain shape change expected
        ├── app/
        │   ├── repository.go         # add FindRandomAccessible
        │   ├── service.go            # add GetRandom and memory repository support
        │   ├── policy.go             # reuse access policy
        │   └── service_test.go       # add random access filtering coverage
        ├── infra/
        │   └── gorm_repository.go    # add random accessible PostgreSQL-backed query
        └── http/
            ├── handler.go            # add GetRandomPostcard
            ├── routes.go             # register /random before /:id
            └── dto.go                # reuse existing PostcardResponse

tests/
└── contract/
    └── postcard_contract_test.go     # add anonymous, authenticated, no-access, route-order checks

docs/
└── v1-postcards-random.md            # source build documentation
```

**Structure Decision**: Use the existing postcard module only. The endpoint is a read capability under `/v1/postcards/*`, so no new module, migration, platform service, or cross-module direct dependency is required.

## Technical Design

### Backend Workflow

```text
GET /v1/postcards/random
  -> OptionalAccessToken middleware
  -> postcardshttp.Handler.GetRandomPostcard
  -> postcardsapp.Service.GetRandom(userID)
  -> postcardsapp.Repository.FindRandomAccessible(userID)
  -> random accessible PostgreSQL query or in-memory selection
  -> Service.attachRelations(postcard)
  -> response.Write(...)
```

### Interface Changes

Update `internal/modules/postcards/app/repository.go`:

```go
type Repository interface {
	Create(postcard *postcardsdomain.Postcard) error
	FindByID(id uint) (*postcardsdomain.Postcard, error)
	FindRandomAccessible(userID uint) (*postcardsdomain.Postcard, error)
	List() ([]postcardsdomain.Postcard, error)
	Update(postcard *postcardsdomain.Postcard) error
	Delete(id uint) error
}
```

### App Service

Add `GetRandom(userID uint) (*postcardsdomain.Postcard, error)` to `internal/modules/postcards/app/service.go`.

Behavior:

- Call `s.repo.FindRandomAccessible(userID)`.
- Map repository errors to `errs.Internal("获取随机明信片失败")`.
- If repository returns `nil`, return `errs.NotFound("明信片不存在")`.
- Call `s.attachRelations(postcard)` before returning.
- Preserve relation-loading failure behavior by mapping relation errors to `errs.Internal("获取随机明信片失败")`.

### Repository Behavior

Production repository:

- Anonymous caller (`userID == 0`): select public postcards only.
- Authenticated caller: select postcards where visibility is public or author matches `userID`.
- Use database random ordering with a limit of one.
- Return `(nil, nil)` when no row is found.
- Return domain postcard on success and raw error on unexpected storage failure.

Memory repository:

- Build an accessible candidate list using the same visibility and author rules.
- Pick one candidate randomly.
- Return a copy of the selected postcard.
- Return `(nil, nil)` if no candidate exists.

### HTTP Handler and Routes

Add `GetRandomPostcard` to `internal/modules/postcards/http/handler.go`:

- Read optional `userID` from Gin context if present.
- Call `h.postcards.GetRandom(userID)`.
- Map errors through `errs.MapHTTP`.
- Return `200 OK` with message `获取随机明信片成功` and `newPostcardResponse(postcard)`.

Update public postcard route registration:

```go
func RegisterPublicRoutes(group *gin.RouterGroup, handler *Handler) {
	group.GET("/random", handler.GetRandomPostcard)
	group.GET("", handler.GetPostcards)
	group.GET("/:id", handler.GetPostcardDetail)
}
```

`/random` must be registered before `/:id`.

## Testing Plan

- Add app-level tests for random access filtering:
  - anonymous candidates include public postcards only.
  - authenticated candidates include public postcards and owned private postcards.
  - no accessible candidates returns not found from the service.
- Add contract tests in `tests/contract/postcard_contract_test.go`:
  - anonymous random postcard returns `200 OK`, message `获取随机明信片成功`, and public visibility.
  - authenticated random postcard can return the caller's private postcard when it is the only accessible candidate.
  - no accessible postcard returns `404 Not Found` and message `明信片不存在`.
  - `/v1/postcards/random` is handled by the random route rather than the `/:id` detail route.
- Re-run existing postcard contract tests to confirm list/detail behavior remains unchanged.
- Run the offline-safe full suite before completion:

```bash
env GOCACHE=/tmp/go-build GOPROXY=off go test ./... -v
```

## Post-Design Constitution Check

- [x] Planned structure still stays inside the existing postcard vertical slice.
- [x] Handler, service, and repository responsibilities remain separated.
- [x] Contract artifact defines request, response, and error behavior.
- [x] No new external dependency or platform wiring is introduced.
- [x] Visibility and ownership invariants are preserved and extended only by random candidate selection.
- [x] Unit and contract tests cover the changed behavior and route-order risk.

## Complexity Tracking

No constitution violations.
