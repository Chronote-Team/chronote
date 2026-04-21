# HTTP API Contract: Chronote Backend Contract-Preserving Refactor

## Contract Scope

This document records the public HTTP contract that the replacement backend must preserve while implementation moves into the `refactor/` application tree.

## Global Contract Rules

- Response envelopes remain JSON with `code`, `message`, and optional `data`.
- Success and failure status codes remain compatible with the current backend for equivalent requests.
- Default error-message text remains compatible unless an approved exception is documented.
- Request and response DTOs remain separate from persistence entities and storage-only fields.

## Route Groups

### Health

- `GET /health`
- `GET /health/details`

Compatibility requirements:
- `/health` returns lightweight overall readiness only.
- `/health/details` returns component details and preserves `200`, `207`, and `503` semantics.
- Database and Redis influence overall availability; S3 may be degraded without forcing full unavailability alone.

### Users And Auth

- `POST /user/register`
- `POST /user/login`
- `POST /user/refresh`
- `GET /user/info`
- `POST /user/logout`
- `POST /user/avatar`
- `PUT /user/update/displayname`
- `PUT /user/update/password`

Compatibility requirements:
- Public routes remain register, login, and refresh.
- Protected routes require `Authorization: Bearer <access_token>`.
- Refresh and logout preserve current refresh-token request semantics where applicable.
- Registration preserves username/email uniqueness, lowercase email normalization, display-name defaulting, and password rules.

### Postcards

- `POST /v1/postcards`
- `GET /v1/postcards`
- `GET /v1/postcards/:id`
- `PUT /v1/postcards/:id`
- `DELETE /v1/postcards/:id`

Compatibility requirements:
- Create preserves both JSON and multipart request paths.
- Read endpoints preserve anonymous access to public postcards only.
- Mutations remain owner-only.
- Pagination, sorting, and detail/list payload shapes remain compatible.

### Media

- `POST /v1/postcards/:id/media`
- `GET /v1/postcards/:id/media`
- `PUT /v1/postcards/:id/media/reorder`
- `DELETE /v1/postcards/:id/media/:media_id`

Compatibility requirements:
- Read access follows postcard visibility rules.
- Upload, reorder, and delete remain owner-only.
- Media groups remain `header`, `gallery`, and `bgm`.
- `header` remains image-only, and `bgm` remains audio-only.
- Reorder semantics preserve `position` ordering behavior.

## Compatibility-Sensitive Error Behavior

The replacement backend must preserve key error categories and current message text defaults for:

- Missing or malformed bearer tokens
- Revoked or expired tokens
- Validation failures on registration and profile mutation
- Credential mismatch on login
- Visibility and ownership enforcement on postcards and media
- Dependency-health degradation and unavailability responses

## Verification Expectations

- Contract tests must exist for every route group.
- Each route group must cover at least one success path and one important failure, authorization, or degradation path.
- Any approved contract deviation must be documented here before release.
