# Metals prices endpoint

- What changed
  - Added `GET /prices/metals` to return current gold (XAU) and silver (XAG) prices from gold-api.com.
  - Added DTOs for metals prices and wired the new route.
- Why it changed
  - Provide a server-side endpoint for current metal prices that can be expanded later.
- New conventions/decisions
  - Endpoint returns `{gold, silver}` with the upstream payload shape.
- Follow-ups / TODOs
  - Consider caching or retry behavior if upstream latency becomes an issue.
