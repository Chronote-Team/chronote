# Chronote Refactor Migration Notes

The current cutover target preserves the legacy table semantics for:

- `users`
- `postcards`
- `postcard_media`
- `ai_analysis_jobs`
- `media_ai_analysis`
- `postcard_ai_analysis`

The US3 integration tests create these tables directly in an isolated verification database to confirm that the replacement backend can read and write supported data without a schema redesign.

AI understanding tables:

- `202605130001_create_ai_analysis_tables.sql` adds durable internal analysis job and result storage.
- `ai_analysis_jobs` stores recoverable pending/processing/succeeded/failed/unavailable/stale workflow state.
- `media_ai_analysis` stores image-level structured understanding keyed by media, media version, prompt version, schema version, and model version.
- `postcard_ai_analysis` stores final postcard-level structured understanding keyed by postcard, postcard version, prompt version, schema version, and model version.
- These tables are internal only and do not change normal client app postcard response contracts.

Current status:

- The replacement app expects the legacy-compatible schema above plus the internal AI understanding tables when AI is enabled.
- Cutover verification uses fixture-managed schema setup before HTTP compatibility checks.

Before production cutover:

1. Convert the verified schema into reviewed migration files.
2. Add idempotent bootstrap for indexes and constraints that must exist in production.
3. Run the full contract and integration suite against a migrated database snapshot.
