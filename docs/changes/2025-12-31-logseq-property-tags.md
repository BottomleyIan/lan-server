# Logseq property tags

## What changed
- Added support for `key:: value` Logseq properties by adding both key and value to journal entry tags.
- Journal tag extraction now also includes Logseq property key/value pairs.
- Property values wrapped in `[[...]]` are normalized to avoid duplicate tags.

## Why it changed
- Keep journal and entry tag sets aligned with Logseq property metadata without introducing a new schema.

## New conventions/decisions
- Property lines contribute two tags: the property key and its value (with Logseq brackets stripped).

## Follow-ups / TODOs
- None.
