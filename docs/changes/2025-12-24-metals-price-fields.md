# Metals price fields

- What changed
  - Metals price response now returns `gbp` and `usd` fields instead of a single `price` field.
- Why it changed
  - Expose both currencies explicitly in the API response.
- New conventions/decisions
  - `gbp` is derived from the USD price using the latest USD→GBP rate.
- Follow-ups / TODOs
  - Consider adding a currency code field if additional currencies are added.
