# Metals prices cache + GBP

- What changed
  - Added 1-minute in-memory cache for `GET /prices/metals`.
  - Converted gold/silver prices from USD to GBP using the currency API with a fallback URL.
- Why it changed
  - Reduce repeated upstream calls and return prices in GBP for clients.
- New conventions/decisions
  - Currency conversion uses latest USD->GBP rate; cache TTL is 60 seconds.
- Follow-ups / TODOs
  - Consider sharing a global HTTP client and adding optional stale-cache fallback on upstream errors.
