# Regular Notes — CRUD

All note writes (create, modify, delete) must go through the CLI — never touch note files directly.

## List notes

```bash
vaultr list                                          # all notes, sorted by most recently updated
vaultr list /journal                                 # scope to a vault directory
vaultr list --latest 7                               # updated in the last 7 days
vaultr list --start 2026-01-01 --end 2026-01-31      # date range
vaultr list --limit 20                               # cap results
```

## Search notes

```bash
vaultr search "meeting notes"          # search filename + content (default)
vaultr search april --field name       # filename only
vaultr search "TODO" --field content   # content only
vaultr search "project" --limit 10
```

Query syntax: bare words are OR-joined; `"quoted phrase"` matches exactly.

## Read a note

```bash
vaultr read /journal/today.md    # vault-absolute path
vaultr read today.md             # bare filename — resolves to most recently updated match
```

If multiple notes share the same name, use the full vault path to be unambiguous.

## Preview / partial read (save tokens)

Before reading a note in full, use `vaultr extract` to inspect only what you need.

**Decide whether to read the whole note:**
```bash
vaultr extract outline today.md          # heading structure — see if the note is relevant
vaultr extract segment today.md --head 20  # first 20 lines — quick content preview
vaultr extract tag today.md              # front-matter tags only
```

**Read only the part you need:**
```bash
vaultr extract section today.md "Goals"         # extract a named section (case-insensitive)
vaultr extract section today.md "## meeting notes"
vaultr extract segment today.md --start 10 --end 30   # specific line range
vaultr extract segment today.md --tail 15        # last 15 lines
```

**Extract specific content types:**
```bash
vaultr extract code today.md     # all fenced code blocks
vaultr extract list today.md     # all bullet/numbered lists
vaultr extract link today.md     # all links
```

**Recommended workflow for unknown or long notes:**
1. `vaultr extract outline <note>` — scan the structure
2. `vaultr extract section <note> "<heading>"` — read only the relevant section(s)
3. `vaultr read <note>` — fall back to full read only when the above isn't enough

## Create a note

```bash
vaultr create /path/note.md --content "# Title\n\nContent here."
vaultr create /path/note.md --file local.md       # copy from a local file
```

Path rules:
- Must be vault-absolute and start with `/`
- Use a meaningful folder: `/journal/`, `/projects/`, `/research/`, etc.
- Use `.md` extension

## Modify a note

**Overwrite** (replace entire content):
```bash
vaultr create /path/note.md --content "updated content" --force
```

**Append** to the end:
```bash
vaultr append /path/note.md --content "new paragraph"
```

**Prepend** to the top:
```bash
vaultr prepend /path/note.md --content "header line"
```

Read the note first (`vaultr read`) if you need its current content before modifying.

## Delete a note

```bash
vaultr delete /path/note.md
```

Path must be vault-absolute. Use `vaultr resolve <filename>` first if you only know the filename.
