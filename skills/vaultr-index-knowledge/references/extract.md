# Extract — Partial Note Reading

`vaultr extract` reads parts of a note without loading the full content. Use it whenever the full note is not needed.

## Subcommands

**Outline** — heading structure only:
```bash
vaultr extract outline <path-or-name>
```

**Section** — one named section and its content (ends at the next heading of equal or higher level):
```bash
vaultr extract section <path-or-name> "<heading>"   # case-insensitive
vaultr extract section today.md "## meeting notes"
```

**Segment** — line range:
```bash
vaultr extract segment <path-or-name> --head 20        # first N lines
vaultr extract segment <path-or-name> --tail 10        # last N lines
vaultr extract segment <path-or-name> --start 5 --end 30  # specific range (1-based, inclusive)
```

**Tags** — front-matter tags only:
```bash
vaultr extract tag <path-or-name>
```

**Code blocks** — all fenced code blocks:
```bash
vaultr extract code <path-or-name>
```

**Lists** — all bullet and numbered lists:
```bash
vaultr extract list <path-or-name>
```

**Links** — all links (wikilinks and markdown links); each result shows type and target:
```bash
vaultr extract link <path-or-name>
```

## Path rules

Same as `vaultr read`: accepts vault-absolute paths (`/journal/today.md`) or bare filenames (resolves to most recently updated match).
