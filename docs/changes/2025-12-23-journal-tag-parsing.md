# Journal tag parsing

- What changed
  - Updated journal tag extraction to parse `[[tag]]` patterns instead of `#tag`.
- Why it changed
  - Journal files use Logseq-style double-bracket tags like `[[dog]][[cat]]`.
- New conventions/decisions
  - Tags are extracted from `[[...]]` sequences and stored as plain tag strings.
- Follow-ups / TODOs
  - Consider trimming or normalizing tag casing if needed.
