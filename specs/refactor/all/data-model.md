# Data Model: Chronote Backend Contract-Preserving Refactor

## User Account

**Description**: Represents a Chronote user who can authenticate, manage a profile, and own postcards.

**Core fields**:
- `id`: Stable user identifier
- `username`: Unique account name, limited to letters, digits, and underscores
- `display_name`: User-facing name; defaults to username when not provided
- `email`: Unique email address persisted in lowercase form
- `password_hash`: Hashed password value, never exposed in public responses
- `avatar_url`: Optional profile image location when avatar upload is supported
- `created_at`
- `updated_at`

**Validation rules**:
- Username is required, unique, length-limited, and format-restricted.
- Display name is optional but length-limited.
- Email is required, unique, format-validated, and normalized to lowercase before persistence.
- Password is required for registration and password changes, length-limited, and cannot be blank-only input.

**Relationships**:
- One user owns many postcards.

## Auth Session Tokens

**Description**: Logical credentials used for authorization and session refresh.

**Core fields**:
- `access_token`
- `refresh_token`
- `token_type`: Distinguishes access vs refresh semantics
- `subject_user_id`
- `expires_at`

**Validation rules**:
- Access and refresh tokens are not interchangeable.
- Protected routes accept bearer access tokens only.
- Refresh and logout flows use refresh-token request-body semantics where currently defined.

**State transitions**:
- Issued on successful login.
- Refreshed through the refresh flow.
- Refresh tokens can be blacklisted on logout.

## Postcard

**Description**: A user-owned content object that can be listed, viewed, created, updated, or deleted.

**Core fields**:
- `id`
- `author_id`
- `title`
- `content`
- `visibility`
- `created_at`
- `updated_at`

**Validation rules**:
- Each postcard belongs to exactly one author.
- Visibility is restricted to `public` or `private`.
- Anonymous reads can access only `public` postcards.
- Mutations are restricted to the owning user.

**Relationships**:
- One postcard belongs to one user.
- One postcard has many media items.

## Postcard Media Item

**Description**: A media asset associated with a postcard and ordered within a media group.

**Core fields**:
- `id`
- `postcard_id`
- `group`
- `media_type`
- `position`
- `storage_key`
- `public_url`
- `created_at`

**Validation rules**:
- Each media item belongs to exactly one postcard.
- Allowed groups are `header`, `gallery`, and `bgm`.
- `header` accepts image media only.
- `bgm` accepts audio media only.
- `position` ordering is meaningful and must be preserved during reorder operations.
- Mutation operations require postcard ownership.

**Relationships**:
- Many media items belong to one postcard.

## Health Status Snapshot

**Description**: A transport-neutral representation of current service readiness and dependency health.

**Core fields**:
- `healthy`: Overall boolean health indicator
- `overall_status`: Logical result such as operational, degraded, or unavailable
- `components`: Collection of component health entries

## Health Component Status

**Description**: Dependency-level health information returned by the detailed health endpoint.

**Core fields**:
- `name`
- `status`
- `message`
- `latency_ms`

**Validation rules**:
- Database and Redis determine overall availability.
- S3 can be reported as degraded without forcing total service unavailability by itself.
