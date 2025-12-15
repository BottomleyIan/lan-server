# 2025-12-15 allow-all CORS middleware

## What changed
- Added a global CORS middleware in `cmd/server/main.go` that allows any origin, common methods, and common headers, and short-circuits `OPTIONS` with 204.

## Why it changed
- Needed to enable cross-origin requests from any frontend while developing; allow-all is acceptable for now.

## New conventions / decisions
- Default CORS policy is allow-all; revisit to tighten origins/headers later.

## Follow-ups / TODOs
- Consider narrowing allowed origins/headers once frontend hosts are known.
