# Knowledge

The knowledge base is three layers deep:

```
Domain Index  →  Knowledge Units  →  Raw Notes (sources)
```

- **Domain Index** — one index per domain; lists all knowledge units in that domain; the entry point.
- **Knowledge Units** — compiled, synthesized notes distilled from raw notes; usually sufficient to answer a question. Each unit contains links to its source raw notes.
- **Raw Notes** — original source notes; follow the source links inside a knowledge unit to read them directly when more detail is needed.

Retrieval path: `list-indexes` → read domain index → read matching knowledge units → follow source links to raw notes → `vaultr search` as last-resort fallback only.

Knowledge notes are auto-generated and are read-only (except for explicit deletes).

## Retrieval Workflow

1. `vaultr knowledge list-indexes --table` — identify relevant domains
2. `vaultr knowledge read <index-path>` — read the domain index to find matching knowledge units
3. `vaultr knowledge read <unit-path>` — read the knowledge unit; usually sufficient
4. If more detail is needed: follow the source links inside the knowledge unit to read the original raw notes directly (`vaultr read <source-path>`)
5. Last resort only: `vaultr search "X"` — use when source links are absent or the above steps don't cover the question

If the domain is unclear, skip steps 1–2 and go straight to `vaultr knowledge search "X"`.

## Command Reference

### List domain indexes

```bash
vaultr knowledge list-indexes           # all domains with vault path (JSON)
vaultr knowledge list-indexes --table   # table format
```

### List knowledge notes

```bash
vaultr knowledge list --kind knowledge                                          # knowledge notes, newest first
vaultr knowledge list --kind knowledge --latest 7                               # updated in the last 7 days
vaultr knowledge list --kind knowledge --start 2026-01-01 --end 2026-01-31      # date range
vaultr knowledge list --kind knowledge --limit 20                               # capped
```

### Read a knowledge note

```bash
vaultr knowledge read "/_knowledge/summary.md"   # vault-absolute path
vaultr knowledge read summary.md                  # bare filename — most recent match
```

### Search knowledge notes

```bash
vaultr knowledge search "machine learning"         # name + content (default)
vaultr knowledge search topic --field name         # filename only
vaultr knowledge search "key idea" --field content # content only
vaultr knowledge search "project" --limit 10
```

Query syntax: bare words are OR-joined; `"quoted phrase"` matches exactly.