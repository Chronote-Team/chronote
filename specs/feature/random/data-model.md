# Data Model: Random Accessible Postcard

## Entity: Postcard

**Purpose**: User-authored content item returned by the random postcard endpoint.

**Relevant Fields**:

- `id`: Stable postcard identifier.
- `title`: Postcard title.
- `content`: Structured postcard content.
- `visibility`: Access level, either `public` or `private`.
- `author_id`: Identifier of the user who owns the postcard.
- `author`: Existing author relationship attached to postcard responses.
- `medias`: Existing media relationship attached to postcard responses.
- `created_at`: Existing creation timestamp.
- `updated_at`: Existing update timestamp.

**Validation and Invariants**:

- Visibility remains limited to `public` and `private`.
- Anonymous callers can receive only `public` postcards.
- Signed-in callers can receive `public` postcards and postcards where `author_id` matches the caller.
- The random endpoint returns at most one postcard.
- No stored postcard schema change is planned.

## Entity: Caller

**Purpose**: The visitor requesting a random postcard.

**Relevant Fields**:

- `user_id`: Optional signed-in user identifier. A zero or absent value means anonymous access.
- `auth_state`: Anonymous or signed in through an existing valid access token.

**Validation and Invariants**:

- Missing auth context is valid and results in anonymous candidate rules.
- Invalid optional auth context must not grant private access.
- Valid signed-in context expands candidates to include the caller's own postcards.

## Entity: Author

**Purpose**: Existing user relationship shown with postcard responses and used for ownership checks.

**Relevant Fields**:

- `id`: Author identifier.
- `username`: Public username.
- `display_name`: Public display name.
- `avatar`: Optional public avatar URL.

**Validation and Invariants**:

- A postcard has exactly one author identifier.
- Author relation output follows existing postcard detail/list behavior.

## Entity: Postcard Media

**Purpose**: Existing media relationship shown with postcard responses.

**Relevant Fields**:

- `id`: Media identifier.
- `postcard_id`: Owning postcard identifier.
- `type`: Media type.
- `url`: Media URL.
- `thumbnail_url`: Optional thumbnail URL.
- `original_width`: Optional image width.
- `original_height`: Optional image height.
- `duration`: Optional media duration.
- `file_size`: Media file size.
- `position`: Display order.
- `group`: Media group.
- `created_at`: Existing creation timestamp.
- `updated_at`: Existing update timestamp.

**Validation and Invariants**:

- Media response shape follows existing postcard detail/list behavior.
- A postcard with no media is still a valid random postcard response.

## State and Selection Rules

```text
anonymous caller
  -> candidate postcards: visibility == public
  -> no candidates: existing postcard not-found behavior
  -> one or more candidates: return one random candidate

signed-in caller
  -> candidate postcards: visibility == public OR author_id == caller user_id
  -> no candidates: existing postcard not-found behavior
  -> one or more candidates: return one random candidate
```
