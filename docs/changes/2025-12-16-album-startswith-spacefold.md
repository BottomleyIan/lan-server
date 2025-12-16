# 2025-12-16 album startswith ignores spaces

## What changed
- Album list queries now perform case-insensitive prefix matching without removing spaces, so `startswith=a%20` matches `A Dog` or `a cat` but not `Art`.

## Why it changed
- Prefix searches were space-sensitive, making it awkward to find titles when the user included or omitted spaces.

## New conventions / decisions
- `startswith` for albums is case-folded; spaces are preserved so they can be part of the prefix.

## Follow-ups / TODOs
- Consider similar folding (case/diacritics) if user-facing search needs broader normalization.
