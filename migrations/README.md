# Chronote Refactor Migration Notes

The current cutover target preserves the legacy table semantics for:

- `users`
- `postcards`
- `postcard_media`

The US3 integration tests create these tables directly in an isolated verification database to confirm that the replacement backend can read and write supported data without a schema redesign.

Current status:

- No irreversible migration scripts are committed yet.
- The replacement app expects the legacy-compatible schema above.
- Cutover verification uses fixture-managed schema setup before HTTP compatibility checks.

Before production cutover:

1. Convert the verified schema into reviewed migration files.
2. Add idempotent bootstrap for indexes and constraints that must exist in production.
3. Run the full contract and integration suite against a migrated database snapshot.
