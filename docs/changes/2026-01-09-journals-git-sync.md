# Journals git sync endpoint

- What changed: added a journals sync endpoint that stages/commits changes, pulls, and pushes the journals git repo with outputs returned; wired the route and Swagger docs.
- Why it changed: the journals folder is a git repository and needs a simple server-triggered sync flow.
- New conventions/decisions: treat a clean working tree as a skipped commit and proceed with pull/push.
- Follow-ups / TODOs: none.
