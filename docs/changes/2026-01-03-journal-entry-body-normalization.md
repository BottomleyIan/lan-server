# Journal entry body normalization

## What changed
- Normalized journal entry request bodies to drop duplicated first-line content, de-indent Logseq body lines, and remove scheduled/deadline lines before rendering.

## Why it changed
- Prevent updates from merging two entries when clients send back the API's `body` field verbatim.

## New conventions/decisions
- Request `body` is treated as the indented Logseq body content, not including the entry's first line.

## Follow-ups / TODOs
- None.
