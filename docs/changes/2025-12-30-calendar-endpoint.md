# Calendar day endpoint

- What changed: added GET /calendar?year=&month=&day= returning tasks and notes for a day; added DayViewDTO and required date query parsing.
- Why it changed: provide a single day view payload for the app instead of separate tasks and notes calls.
- New conventions/decisions: /calendar requires year, month, and day query parameters and returns tasks filtered by journal/scheduled/deadline dates and notes by journal date.
- Follow-ups / TODOs: none.
