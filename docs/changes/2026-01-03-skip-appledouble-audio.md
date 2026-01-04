# Skip AppleDouble files during scan

## What changed
- Skip macOS `._` AppleDouble files during music scans.

## Why it changed
- Prevent ffprobe/tag errors on `._` files from aborting scans.

## New conventions/decisions
- None.

## Follow-ups / TODOs
- None.
