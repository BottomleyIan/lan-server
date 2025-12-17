# Fix stuck folder scan status

## What changed
- Added folder availability detection in the scanner and return a typed `ErrFolderUnavailable` when the root is missing or not a directory.
- Updated the scan handler to always finish scans in the database, marking success, unavailable, or error so `last_scan_status` no longer gets stuck at `running`.
- Reset folder statuses in `data.sqlite`: `~/Downloads` marked `ok`; `/Volumes/SAMSUNG/MUSIC` marked `skipped_unavailable` with an error note and `available=0`.

## Why it changed
- Folder scans were left in `running` indefinitely because we never recorded completion or failure, and unmounted volumes were not detected up front.

## New conventions/decisions
- Scan requests record completion state (ok/error/skipped_unavailable) immediately in the handler, even when the request fails.
- Treat missing/unmounted scan roots as “folder unavailable” and surface that explicitly.

## Follow-ups / TODOs
- Track progress and remediation in Beads issue `musicserver-3un` (investigate stuck folder scans).
