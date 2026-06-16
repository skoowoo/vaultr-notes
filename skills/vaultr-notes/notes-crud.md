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
