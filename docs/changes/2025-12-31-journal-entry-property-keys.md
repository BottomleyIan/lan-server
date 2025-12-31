# Journal entry property keys

## What changed
- Added `property_keys` JSON array column on `journal_entries` and populate it during journal sync.
- Parsed Logseq `key:: value` lines to capture property keys per entry and property tags/values.
- Added endpoints to list property keys and property values.
- Property value parsing now drops empty bracket-only values like `[[]]`.
- Exposed `property_keys` on the journal entry DTO and refreshed Swagger docs.

## Why it changed
- Enable querying which property fields exist across journal entries and list values for a given key without a separate properties table.

## New conventions/decisions
- Property keys are normalized to lowercase and stored as a JSON array string in `journal_entries.property_keys`.
- Property values are inferred from entry content when listing values.

## Follow-ups / TODOs
- None.
