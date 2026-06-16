package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hardhacker/vaultr/internal/plugin"
	"github.com/hardhacker/vaultr/internal/util"
)

const (
	// vaultInternalDir holds metadata DB, search index, and other
	// vault-internal data. User notes live directly under the vault root.
	vaultInternalDir = ".vaultr"

	vaultDBFileName = "meta.db"
)

func vaultInternalPath(root string) string {
	return filepath.Join(root, vaultInternalDir)
}

func vaultDBPath(root string) string {
	return filepath.Join(vaultInternalPath(root), vaultDBFileName)
}

// IsVaultInitialized reports whether dir contains a .vaultr directory.
// It does not create any paths on disk.
func IsVaultInitialized(dir string) (bool, error) {
	expanded, err := expandHome(dir)
	if err != nil {
		return false, fmt.Errorf("vault: expand path: %w", err)
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return false, fmt.Errorf("vault: resolve path: %w", err)
	}
	st, err := os.Stat(vaultInternalPath(abs))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("vault: stat %s: %w", vaultInternalPath(abs), err)
	}
	if !st.IsDir() {
		return false, nil
	}
	return true, nil
}

// EventHook is an optional callback invoked after every successful vault mutation.
// It must not block.
type EventHook func(e plugin.Event)

// Vault is the root of the Vaultr personal database.
//
// All note operations take explicit (dir, name string) coordinates:
//   - dir  — vault-absolute slash path of the containing directory, e.g. "/journal/2026".
//     The vault root is represented as "/".
//   - name — base filename, e.g. "april.md".
//
// All directory operations take a single dir string using the same convention.
//
// Vault is safe for concurrent use: read operations share a read lock and
// write operations hold an exclusive write lock.
//
// Every write operation follows a simple protocol:
//  1. Mutate the filesystem.
//  2. Update the metadata database.
type Vault struct {
	mu          sync.RWMutex
	root        string              // absolute, cleaned OS path
	db          *sql.DB             // SQLite metadata index; always non-nil after New()
	shortsDir   string              // configurable; defaults to ShortDir const
	onEvent     EventHook           // optional; called after each successful mutation
	ensureWatch func(absDir string) // set by StartWatcher; sync fsnotify watches after MkdirAll
}

// New creates a Vault rooted at dir.
// dir is expanded (~ → home dir) and created if it does not yet exist.
func New(dir string) (*Vault, error) {
	expanded, err := expandHome(dir)
	if err != nil {
		return nil, fmt.Errorf("vault: expand path: %w", err)
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return nil, fmt.Errorf("vault: resolve path: %w", err)
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, fmt.Errorf("vault: create root %q: %w", abs, err)
	}
	if err := os.MkdirAll(vaultInternalPath(abs), 0o750); err != nil {
		return nil, fmt.Errorf("vault: create internal dir: %w", err)
	}
	db, err := openDB(abs)
	if err != nil {
		return nil, fmt.Errorf("vault: open metadata db: %w", err)
	}
	return &Vault{root: abs, db: db, shortsDir: ShortDir}, nil
}

// Close releases the vault's underlying resources (SQLite DB).
func (g *Vault) Close() error {
	return g.db.Close()
}

// Root returns the absolute path of the vault root on disk.
func (g *Vault) Root() string { return g.root }

// SetShortsDir overrides the default shorts directory (used when callers pass an empty dir).
func (g *Vault) SetShortsDir(dir string) {
	if dir != "" {
		g.shortsDir = strings.TrimRight(dir, "/")
	}
}

// ── compile plugin hooks ──────────────────────────────────────────────────────

// MarkNoteAsKnowledge sets origin = plugin:compile, compile_count, and tags on the note.
// title is optional; pass an empty string to leave the existing title unchanged.
func (g *Vault) MarkNoteAsKnowledge(p Path, title string, compileCount int, tags []string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return dbMarkKnowledgeOrigin(g.db, p, title, compileCount, tags)
}

// MarkNoteCompiled sets compile_count = 1 on the raw note at p, marking it as
// submitted for compilation. Idempotent; always writes 1, never increments.
func (g *Vault) MarkNoteCompiled(p Path) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return dbSetCompileCount(g.db, p, 1)
}

// MarkNoteAsIndex sets origin = plugin:index on the note.
// title is optional; pass an empty string to leave the existing title unchanged.
func (g *Vault) MarkNoteAsIndex(p Path, title string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return dbMarkIndexOrigin(g.db, p, title)
}


// GetKnowledgeDeps returns all source note paths that the given knowledge note depends on.
func (g *Vault) GetKnowledgeDeps(knowledge Path) ([]Path, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbGetKnowledgeDeps(g.db, knowledge)
}

// GetSourceKnowledges returns all knowledge note paths that list the given note as a source.
func (g *Vault) GetSourceKnowledges(source Path) ([]Path, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbGetSourceKnowledges(g.db, source)
}

// GetNotesByPaths returns metadata for each path in the slice.
// Missing rows are silently skipped.
func (g *Vault) GetNotesByPaths(paths []Path) ([]Note, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbGetByPaths(g.db, paths)
}

// SetKnowledgeDeps replaces all knowledge_deps rows for the given knowledge note
// with the supplied source paths, inside a single transaction.
// Passing an empty sources slice clears all existing dependencies.
func (g *Vault) SetKnowledgeDeps(knowledge Path, sources []Path) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return dbReplaceKnowledgeDeps(g.db, knowledge, sources)
}

// SetIndexDeps replaces all index_deps rows for the given index note with the
// supplied knowledge paths, inside a single transaction.
// Passing an empty knowledges slice clears all existing entries.
func (g *Vault) SetIndexDeps(index Path, knowledges []Path) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return dbReplaceIndexDeps(g.db, index, knowledges)
}

// GetIndexDeps returns all knowledge note paths listed by the given index note.
func (g *Vault) GetIndexDeps(index Path) ([]Path, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbGetIndexDeps(g.db, index)
}

// GetAllIndexDepCounts returns a map of index note PathString → knowledge dep count
// for every index note that has at least one dep.
func (g *Vault) GetAllIndexDepCounts() (map[string]int, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbGetAllIndexDepCounts(g.db)
}

// GetKnowledgeIndexes returns all index note paths that include the given knowledge note.
func (g *Vault) GetKnowledgeIndexes(knowledge Path) ([]Path, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbGetKnowledgeIndexes(g.db, knowledge)
}

// SetNoteTitle sets the title column for the note at p.
func (g *Vault) SetNoteTitle(p Path, title string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return dbSetTitle(g.db, p, title)
}

func (g *Vault) IsNoteTracked(p Path) (bool, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var id int
	err := g.db.QueryRow(`SELECT id FROM notes WHERE dir = ? AND name = ?`, p.Dir(), p.Base()).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// ── event hook ────────────────────────────────────────────────────────────────

// SetEventHook registers a callback that is invoked (non-blocking) after every
// successful note or directory mutation. Pass nil to remove a registered hook.
func (g *Vault) SetEventHook(fn EventHook) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.onEvent = fn
}

// emit fires the event hook for a standard mutation event.
// Must be called after the write-lock has been released.
func (g *Vault) emit(typ, path string, isDir bool, t time.Time) {
	g.emitEvent(plugin.Event{Type: plugin.EventType(typ), Path: path, IsDir: isDir, Time: t})
}

// emitEvent fires the event hook with a fully-populated Event.
func (g *Vault) emitEvent(e plugin.Event) {
	g.mu.RLock()
	fn := g.onEvent
	g.mu.RUnlock()
	if fn != nil {
		fn(e)
	}
}

// ── note read operations ──────────────────────────────────────────────────────

func (g *Vault) StatNote(p Path) (Note, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	abs, absErr := g.osPath(p)
	if absErr != nil {
		return Note{}, absErr
	}

	n, dbErr := dbGet(g.db, p)
	if dbErr != nil {
		// DB is authoritative: not found there means not found.
		// Check FS only to surface ErrIsDir when the path is a directory.
		if errors.Is(dbErr, ErrNotFound) {
			if info, statErr := os.Stat(abs); statErr == nil && info.IsDir() {
				return Note{}, ErrIsDir
			}
		}
		return Note{}, dbErr
	}

	// Verify the file still exists on disk (guard against stale DB rows).
	if _, statErr := os.Stat(abs); statErr != nil {
		return Note{}, mapOSError(statErr)
	}
	return n, nil
}

func (g *Vault) ReadNote(p Path) ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.readNoteLocked(p)
}

func (g *Vault) ReadNoteStream(p Path) (io.ReadCloser, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if err := validateNoteName(p.Base()); err != nil {
		return nil, err
	}
	abs, err := g.osPath(p)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(abs)
	if err != nil {
		return nil, mapOSError(err)
	}
	return f, nil
}

// ── note write operations ─────────────────────────────────────────────────────

func (g *Vault) WriteNote(p Path, data []byte, origin Origin) error {
	return g.WriteNoteWithMeta(p, data, Note{Origin: origin})
}

// WriteNoteWithMeta writes data to p and persists meta fields (Origin, Title)
// atomically in the same DB upsert. Use this instead of WriteNote
// followed by separate Set* calls when the metadata is known up front.
func (g *Vault) WriteNoteWithMeta(p Path, data []byte, meta Note) error {
	g.mu.Lock()
	isNew, modTime, err := g.writeNoteLocked(p, data, meta)
	g.mu.Unlock()
	if err != nil {
		return err
	}
	if isNew {
		g.emit("vault_create", p.String(), false, modTime)
	} else {
		g.emit("vault_write", p.String(), false, modTime)
	}
	return nil
}

// PrependNote inserts incoming into the note at p immediately after the first
// H1 heading. When heading is non-empty the content is inserted at the start
// of the first section whose heading matches (case-insensitive), falling back
// to MdPrepend if not found. An empty heading always uses MdPrepend.
func (g *Vault) PrependNote(p Path, incoming []byte, heading string) error {
	if err := validateNoteName(p.Base()); err != nil {
		return err
	}
	if !util.IsValidText(incoming) {
		return ErrBinaryContent
	}
	g.mu.Lock()
	existing, err := g.readNoteLocked(p)
	if err != nil && !errors.Is(err, ErrNotFound) {
		g.mu.Unlock()
		return err
	}
	var merged []byte
	if heading != "" {
		merged = util.MdPrependSection(existing, incoming, heading)
	} else {
		merged = util.MdPrepend(existing, incoming)
	}
	isNew, modTime, err := g.writeNoteLocked(p, merged, Note{Origin: OriginAPI})
	g.mu.Unlock()
	if err != nil {
		return err
	}
	if isNew {
		g.emit("vault_create", p.String(), false, modTime)
	} else {
		g.emit("vault_write", p.String(), false, modTime)
	}
	return nil
}

// AppendNote appends incoming to the note at p. When heading is non-empty the
// content is inserted after the last section whose heading matches
// (case-insensitive), falling back to end-of-file if not found. An empty
// heading always appends to the end of the file.
func (g *Vault) AppendNote(p Path, incoming []byte, heading string) error {
	if err := validateNoteName(p.Base()); err != nil {
		return err
	}
	if !util.IsValidText(incoming) {
		return ErrBinaryContent
	}
	g.mu.Lock()
	existing, err := g.readNoteLocked(p)
	if err != nil && !errors.Is(err, ErrNotFound) {
		g.mu.Unlock()
		return err
	}
	var merged []byte
	if heading != "" {
		merged = util.MdAppendSection(existing, incoming, heading)
	} else {
		merged = util.MdAppend(existing, incoming)
	}
	isNew, modTime, err := g.writeNoteLocked(p, merged, Note{Origin: OriginAPI})
	g.mu.Unlock()
	if err != nil {
		return err
	}
	if isNew {
		g.emit("vault_create", p.String(), false, modTime)
	} else {
		g.emit("vault_write", p.String(), false, modTime)
	}
	return nil
}

// DeleteNote permanently removes the note at p from the vault.
// If the note is a knowledge note with a linked raw note, the raw note's
// compiled mark is cleared so it can be re-compiled.
func (g *Vault) DeleteNote(p Path) error {
	g.mu.Lock()
	absSrc, err := g.osPath(p)
	if err != nil {
		g.mu.Unlock()
		return err
	}

	if rmErr := os.Remove(absSrc); rmErr != nil && !errors.Is(rmErr, fs.ErrNotExist) {
		g.mu.Unlock()
		return mapOSError(rmErr)
	}
	if dbErr := dbDelete(g.db, p); dbErr != nil {
		g.mu.Unlock()
		return fmt.Errorf("vault: metadata delete %q: %w", p.String(), dbErr)
	}

	g.mu.Unlock()
	g.emit("vault_delete", p.String(), false, time.Now())
	return nil
}

// AppendShort appends content as a new short note entry to today's daily file
// inside shortsDir (defaults to ShortDir when empty). The file is named
// YYYY-MM-DD.md and lives at /<shortsDir>/YYYY-MM-DD.md.
//
// Each entry is prefixed with a sequence heading ("## Short Note N") and a
// creation timestamp. Entries are separated by "---". Any "---" lines present
// in the incoming content are replaced with blank lines so they cannot break
// the entry structure.
//
// The note is stored with origin "short" so it can be filtered distinctly.
// Returns the vault path of the daily file and the note's metadata.
func (g *Vault) AppendShort(content []byte, shortsDir string) (Path, Note, error) {
	if !util.IsValidText(content) {
		return "", Note{}, ErrBinaryContent
	}
	if shortsDir == "" {
		shortsDir = g.shortsDir
	}

	sanitized := shortSanitize(content)

	now := time.Now()
	p := Path("/" + shortsDir + "/" + now.Format("2006-01-02") + ".md")
	timestamp := now.Format("2006-01-02 15:04:05")

	g.mu.Lock()

	existing, err := g.readNoteLocked(p)
	if err != nil && !errors.Is(err, ErrNotFound) {
		g.mu.Unlock()
		return "", Note{}, err
	}

	entry := fmt.Sprintf("###### Short Note: %s\n\n%s", timestamp, strings.TrimSpace(string(sanitized)))

	// Strip any frontmatter (or absence thereof) to get the raw entry body,
	// then always reconstruct with a fresh frontmatter header.
	_, existingBody := util.ParseFrontmatter(existing)
	rawBody := strings.TrimSpace(string(existingBody))

	fm := util.FormatShortFrontmatter(now)
	var merged string
	if rawBody != "" {
		merged = fm + rawBody + "\n\n---\n\n" + entry
	} else {
		merged = fm + entry
	}

	isNew, modTime, writeErr := g.writeNoteLocked(p, []byte(merged), Note{Origin: OriginShort})
	g.mu.Unlock()
	if writeErr != nil {
		return "", Note{}, writeErr
	}

	if isNew {
		g.emit("vault_create", p.String(), false, modTime)
	} else {
		g.emit("vault_write", p.String(), false, modTime)
	}
	g.emitEvent(plugin.Event{
		Type:    plugin.EventVaultShortAppend,
		Path:    p.String(),
		Time:    modTime,
		Content: strings.TrimSpace(string(sanitized)),
	})

	g.mu.RLock()
	note, dbErr := dbGet(g.db, p)
	g.mu.RUnlock()
	if dbErr != nil {
		return p, Note{}, dbErr
	}
	return p, note, nil
}

// shortSanitize replaces any line that is exactly "---" with a blank line so
// that user content cannot accidentally inject an entry separator.
func shortSanitize(content []byte) []byte {
	lines := strings.Split(string(content), "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) == "---" {
			lines[i] = ""
		}
	}
	return []byte(strings.Join(lines, "\n"))
}

// ParseShortEntries parses the raw content of a short daily file and returns
// individual entries, newest first. DailyPath is set on every returned entry.
func ParseShortEntries(content []byte, dailyPath string) []ShortEntry {
	raw := strings.TrimSpace(string(content))
	if raw == "" {
		return nil
	}
	// Strip YAML frontmatter (kind: short daily files carry it).
	if fm, body := util.ParseFrontmatter([]byte(raw)); fm.HasMeta() {
		raw = strings.TrimSpace(string(body))
		if raw == "" {
			return nil
		}
	}
	chunks := strings.Split(raw, "\n\n---\n\n")
	entries := make([]ShortEntry, 0, len(chunks))
	for _, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		lines := strings.SplitN(chunk, "\n", 3)
		const prefix = "###### Short Note: "
		if !strings.HasPrefix(lines[0], prefix) {
			continue
		}
		t, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimPrefix(lines[0], prefix), time.Local)
		if err != nil {
			continue
		}
		body := ""
		if len(lines) >= 3 {
			body = strings.TrimSpace(lines[2])
		}
		entries = append(entries, ShortEntry{Content: body, CreatedAt: t, DailyPath: dailyPath})
	}
	// Reverse to return newest-first (file stores oldest-first).
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries
}

// ListShortEntries returns individual short note entries, newest first, by
// reading and parsing the daily aggregation files in the shorts directory.
func (g *Vault) ListShortEntries(opts ShortListOptions) ([]ShortEntry, error) {
	d := strings.Trim(opts.Dir, "/")
	if d == "" {
		d = g.shortsDir
	}
	listOpts := ListOptions{
		OnlyOrigins: []Origin{OriginShort},
		SortByTime: true,
		After:      opts.After,
		Before:     opts.Before,
	}
	g.mu.RLock()
	dailyNotes, err := dbListByDir(g.db, "/"+d, listOpts)
	g.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	var all []ShortEntry
	for _, daily := range dailyNotes {
		g.mu.RLock()
		content, readErr := g.readNoteLocked(daily.Path())
		g.mu.RUnlock()
		if readErr != nil {
			continue
		}
		for _, e := range ParseShortEntries(content, daily.PathString()) {
			if !opts.After.IsZero() && e.CreatedAt.Before(opts.After) {
				continue
			}
			if !opts.Before.IsZero() && !e.CreatedAt.Before(opts.Before) {
				continue
			}
			all = append(all, e)
			if opts.Limit > 0 && len(all) >= opts.Limit {
				return all, nil
			}
		}
	}
	return all, nil
}

func (g *Vault) MarkNoteIndexed(p Path) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return dbMarkIndexed(g.db, p)
}

// ResetAllIndexed sets indexed = 0 for every note in the metadata database.
// Call this before deleting and rebuilding the search index so that the next
// Backfill re-processes every note with the new index schema.
func (g *Vault) ResetAllIndexed() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return dbResetAllIndexed(g.db)
}

// ── dir operations ────────────────────────────────────────────────────────────

func (g *Vault) StatDir(p Path) (Dir, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	abs, err := g.osPath(p)
	if err != nil {
		return Dir{}, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return Dir{}, mapOSError(err)
	}
	if !info.IsDir() {
		return Dir{}, ErrIsFile
	}
	return dirFromFileInfo(p.String(), info), nil
}

func (g *Vault) ListDir(p Path, opts ListOptions) ([]Note, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, err := g.osPath(p); err != nil {
		return nil, err
	}
	if err := checkListOpts(opts); err != nil {
		return nil, err
	}
	return dbListByDir(g.db, p.String(), opts)
}

// ListChildDirs returns the immediate subdirectory names under p as they appear
// on disk. Hidden directories (names starting with ".") are omitted. Ordering
// is lexicographic.
func (g *Vault) ListChildDirs(p Path) ([]string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	abs, err := g.osPath(p)
	if err != nil {
		return nil, err
	}
	fi, statErr := os.Stat(abs)
	if statErr != nil {
		return nil, mapOSError(statErr)
	}
	if !fi.IsDir() {
		return nil, ErrIsFile
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, mapOSError(err)
	}

	var dirs []string
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		name := ent.Name()
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
			continue
		}
		dirs = append(dirs, name)
	}
	sort.Strings(dirs)
	return dirs, nil
}

// CountNotes returns the total number of notes tracked in the metadata database.
func (g *Vault) CountNotes() (int, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbCount(g.db)
}

// CountKnowledgeNotes returns the total number of compiled knowledge notes.
func (g *Vault) CountKnowledgeNotes() (int, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbCountByOrigin(g.db, string(PluginOrigin("compile")), "")
}

// CountRawNotes returns the number of non-knowledge notes in the vault.
func (g *Vault) CountRawNotes() (int, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbCountByOrigin(g.db, "", string(PluginOrigin("compile")))
}

// ListAllNotes returns every note in the vault regardless of directory.
// Intended for vault-wide operations such as the search index backfill.
func (g *Vault) ListAllNotes(opts ListOptions) ([]Note, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if err := checkListOpts(opts); err != nil {
		return nil, err
	}
	return dbListAll(g.db, opts)
}

func checkListOpts(opts ListOptions) error {
	if len(opts.OnlyOrigins) > 0 && len(opts.ExcludeOrigins) > 0 {
		return fmt.Errorf("storage: OnlyOrigins and ExcludeOrigins cannot both be set")
	}
	return nil
}

// GetNotesByNames returns metadata for every note whose filename matches any
// entry in names. Results are ordered by updated_at descending.
func (g *Vault) GetNotesByNames(names []string) ([]Note, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbGetByNames(g.db, names)
}

// GetNotesByName returns metadata for every note whose filename (name column) equals
// the given base name. Name must satisfy validateNoteName (e.g. "today.md").
// Results are ordered by updated_at descending, then dir ascending.
func (g *Vault) GetNotesByName(name string) ([]Note, error) {
	if err := validateNoteName(name); err != nil {
		return nil, err
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbGetByName(g.db, name)
}

// SetNotePinned sets or clears the pinned flag for the note at p.
func (g *Vault) SetNotePinned(p Path, pinned bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, err := dbGet(g.db, p); err != nil {
		return err
	}
	return dbSetPinned(g.db, p, pinned)
}

// ListPinnedNotes returns all pinned notes ordered by updated_at descending.
func (g *Vault) ListPinnedNotes() ([]Note, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbListPinned(g.db)
}

// ListDirNotes returns notes whose dir column exactly matches dir, queried from
// the metadata DB only (no filesystem check). Use for dirs that are known to
// exist in the DB (e.g. returned by ListDirs). dir should be "/" for the vault
// root or begin with "/" for sub-directories.
func (g *Vault) ListDirNotes(dir string, opts ListOptions) ([]Note, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if err := checkListOpts(opts); err != nil {
		return nil, err
	}
	return dbListByDir(g.db, dir, opts)
}

// ListDirs returns distinct directories that contain notes, sorted alphabetically.
// Directories are read from the DB (not the filesystem). The vault root is "/".
// Any directory whose path contains an underscore-prefixed segment (e.g. "/_shorts",
// "/_clips/2026") is excluded, along with all its children.
func (g *Vault) ListDirs() ([]DirSummary, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	all, err := dbListDirs(g.db)
	if err != nil {
		return nil, err
	}
	out := all[:0]
	for _, d := range all {
		if !dirHasUnderscoreSegment(d.Dir) {
			out = append(out, d)
		}
	}
	return out, nil
}

// ListAllDirs is like ListDirs but includes underscore-prefixed system directories.
func (g *Vault) ListAllDirs() ([]DirSummary, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbListDirs(g.db)
}

// dirHasUnderscoreSegment reports whether any path segment in dir starts with "_".
func dirHasUnderscoreSegment(dir string) bool {
	for _, seg := range strings.Split(dir, "/") {
		if strings.HasPrefix(seg, "_") {
			return true
		}
	}
	return false
}

// ListRecentShortNotes returns metadata for the most recently created short
// notes (origin "short"), ordered by created_at descending.
// When limit <= 0, a default of 7 is used.
func (g *Vault) ListRecentShortNotes(limit int) ([]Note, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbListRecentByOriginByCreatedDesc(g.db, OriginShort, limit)
}

// ErrShortDailyWrongOrigin means a note exists at the expected short path but
// metadata origin is not "short".
var ErrShortDailyWrongOrigin = errors.New("storage: note at short path does not have short origin")

// ShortDailyVaultPath returns the vault-absolute path for the aggregated short file
// for local calendar date day under shortsDir (vault-relative without leading slash;
// empty uses ShortDir).
func ShortDailyVaultPath(shortsDir string, day time.Time) (Path, error) {
	d := strings.Trim(shortsDir, "/")
	if d == "" {
		d = ShortDir
	}
	base := day.Format("2006-01-02") + ".md"
	sp, ok := ParsePath(JoinPath("/"+d, base))
	if !ok {
		return "", fmt.Errorf("storage: invalid shorts path")
	}
	return sp, nil
}

// GetShortDailyNote returns metadata for /{shortsDir}/YYYY-MM-DD.md when the DB row
// exists and origin is "short".
func (g *Vault) GetShortDailyNote(shortsDir string, day time.Time) (Note, error) {
	sp, err := ShortDailyVaultPath(shortsDir, day)
	if err != nil {
		return Note{}, err
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	n, err := dbGet(g.db, sp)
	if err != nil {
		return Note{}, err
	}
	if n.Origin != OriginShort {
		return Note{}, ErrShortDailyWrongOrigin
	}
	return n, nil
}


// ── vault initialisation ──────────────────────────────────────────────────────

// ScanAndRegister walks the vault root recursively and upserts a metadata row
// for every markdown file found on disk.
// Hidden directories (any component beginning with ".") are skipped, which
// excludes .vaultr and editor temp files.
// Returns the number of files registered (newly inserted or updated).
func (g *Vault) ScanAndRegister() (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	count := 0
	err := filepath.WalkDir(g.root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // skip inaccessible entries
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !util.IsMarkdownPath(d.Name()) {
			return nil
		}
		rel, relErr := filepath.Rel(g.root, path)
		if relErr != nil {
			return nil
		}
		slashRel := "/" + filepath.ToSlash(rel)
		dir, name := PathParts(slashRel)
		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}
		if dbErr := dbUpsert(g.db, Note{
			Dir:       dir,
			Name:      name,
			Size:      info.Size(),
			UpdatedAt: info.ModTime(),
			Origin:    OriginFS,
		}); dbErr != nil {
			return fmt.Errorf("storage: register %q: %w", slashRel, dbErr)
		}
		count++
		return nil
	})
	return count, err
}

// ScanAndRegisterFull is the init-time variant of ScanAndRegister.
// It always starts from a clean slate (clears the notes table) and correctly
// distinguishes raw notes from knowledge notes:
//
//   - Notes whose vault path is under knowledgeOutputDir and whose YAML frontmatter
//     contains a "compiled:" field are inserted with origin "plugin:compile".
//
//   - All other notes are inserted with origin "fs".
//
// knowledgeOutputDir is normalised to a vault-absolute slash path (leading "/"
// is added when absent). An empty string defaults to "/_knowledge".
func (g *Vault) ScanAndRegisterFull(knowledgeOutputDir string) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Normalise the knowledge output directory.
	if knowledgeOutputDir == "" {
		knowledgeOutputDir = "/_knowledge"
	}
	if !strings.HasPrefix(knowledgeOutputDir, "/") {
		knowledgeOutputDir = "/" + knowledgeOutputDir
	}
	distillPrefix := knowledgeOutputDir + "/"

	// Clear all existing rows so init is always a clean rebuild.
	if err := dbClearAll(g.db); err != nil {
		return 0, fmt.Errorf("storage: clear notes table: %w", err)
	}
	if _, err := g.db.Exec(`DELETE FROM knowledge_deps`); err != nil {
		return 0, fmt.Errorf("storage: clear knowledge_deps: %w", err)
	}
	if _, err := g.db.Exec(`DELETE FROM index_deps`); err != nil {
		return 0, fmt.Errorf("storage: clear index_deps: %w", err)
	}

	// kDep records a knowledge note's source_notes list for post-walk dep resolution.
	type kDep struct {
		path        Path
		sourceNotes []string
	}
	var depsToRecord []kDep

	// iDep records an index note's parsed body for post-walk index_deps population.
	type iDep struct {
		path Path
		body []byte
	}
	var indexDepsToRecord []iDep

	count := 0
	err := filepath.WalkDir(g.root, func(absPath string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !util.IsMarkdownPath(d.Name()) {
			return nil
		}

		rel, relErr := filepath.Rel(g.root, absPath)
		if relErr != nil {
			return nil
		}
		slashRel := "/" + filepath.ToSlash(rel)
		dir, name := PathParts(slashRel)

		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}

		n := Note{
			Dir:       dir,
			Name:      name,
			Size:      info.Size(),
			UpdatedAt: info.ModTime(),
			CreatedAt: info.ModTime(),
			Origin:    OriginFS,
		}

		// Detect knowledge notes by path prefix and `kind: knowledge` frontmatter,
		// matching the runtime check in the compile plugin.
		isUnderKnowledgeDir := slashRel == knowledgeOutputDir ||
			strings.HasPrefix(slashRel, distillPrefix)
		if isUnderKnowledgeDir {
			data, readErr := os.ReadFile(absPath)
			if readErr == nil {
				fm, body := util.ParseFrontmatter(data)
				if fm.HasMeta() {
					var kind, title, domain string
					var compileCount int
					var srcNotes []string
					for _, e := range fm.All {
						switch e.Key {
						case "kind":
							kind = e.Value
						case "title":
							title = e.Value
						case "domain":
							domain = e.Value
						case "compile_count":
							if cv, err := strconv.Atoi(strings.TrimSpace(e.Value)); err == nil {
								compileCount = cv
							}
						case "source_notes":
							srcNotes = e.List
						}
					}
					switch {
					case strings.EqualFold(kind, "knowledge"):
						n.Origin = PluginOrigin("compile")
						n.Title = title
						n.CompileCount = compileCount
						n.Tags = fm.Tags
						if len(srcNotes) > 0 {
							depsToRecord = append(depsToRecord, kDep{
								path:        Path(slashRel),
								sourceNotes: srcNotes,
							})
						}
					case strings.EqualFold(kind, "index"):
						n.Origin = PluginOrigin("index")
						n.Title = domain
						n.Tags = fm.Tags
						indexDepsToRecord = append(indexDepsToRecord, iDep{
							path: Path(slashRel),
							body: body,
						})
					}
				}
			}
		}

		// Detect short note daily files by `kind: short` frontmatter under the
		// configured shorts directory.
		shortsVaultPrefix := "/" + strings.Trim(g.shortsDir, "/") + "/"
		if n.Origin == OriginFS && strings.HasPrefix(slashRel, shortsVaultPrefix) {
			data, readErr := os.ReadFile(absPath)
			if readErr == nil && util.IsShortNote(data) {
				n.Origin = OriginShort
			}
		}

		if dbErr := dbInsertFull(g.db, n); dbErr != nil {
			return fmt.Errorf("storage: register %q: %w", slashRel, dbErr)
		}
		count++
		return nil
	})
	if err != nil {
		return count, err
	}

	// Populate knowledge_deps from source_notes frontmatter of knowledge notes.
	// Mirrors syncSourceNotes in the compile plugin. Entries with "/" are treated
	// as vault paths; plain names are batch-looked up (ambiguous matches skipped).
	if len(depsToRecord) > 0 {
		var allNames []string
		for _, dep := range depsToRecord {
			for _, sn := range dep.sourceNotes {
				if !strings.Contains(sn, "/") {
					name := sn
					if !strings.HasSuffix(strings.ToLower(name), ".md") {
						name += ".md"
					}
					allNames = append(allNames, name)
				}
			}
		}
		byName := make(map[string][]Note)
		if len(allNames) > 0 {
			if notes, err := dbGetByNames(g.db, allNames); err == nil {
				for _, n := range notes {
					byName[n.Name] = append(byName[n.Name], n)
				}
			}
		}
		for _, dep := range depsToRecord {
			var resolved []Path
			for _, sn := range dep.sourceNotes {
				if strings.Contains(sn, "/") {
					if sp, ok := ParsePath(sn); ok {
						resolved = append(resolved, sp)
					}
				} else {
					name := sn
					if !strings.HasSuffix(strings.ToLower(name), ".md") {
						name += ".md"
					}
					if matches := byName[name]; len(matches) == 1 {
						resolved = append(resolved, matches[0].Path())
					}
				}
			}
			if len(resolved) > 0 {
				_ = dbReplaceKnowledgeDeps(g.db, dep.path, resolved)
			}
		}
	}

	// Populate index_deps by parsing the markdown table in each index note body.
	for _, dep := range indexDepsToRecord {
		var resolved []Path
		for _, kp := range scanIndexTablePaths(dep.body) {
			if sp, ok := ParsePath(kp); ok {
				resolved = append(resolved, sp)
			}
		}
		if len(resolved) > 0 {
			_ = dbReplaceIndexDeps(g.db, dep.path, resolved)
		}
	}

	return count, nil
}

// scanIndexTablePaths extracts vault-absolute knowledge note paths from the
// body of an index note. It looks at every markdown table row and treats the
// last non-empty cell as a path when it starts with "/" and ends with ".md".
func scanIndexTablePaths(body []byte) []string {
	var paths []string
	for _, line := range strings.Split(string(body), "\n") {
		s := strings.TrimSpace(line)
		if !strings.HasPrefix(s, "|") {
			continue
		}
		parts := strings.Split(s, "|")
		for i := len(parts) - 1; i >= 0; i-- {
			v := strings.TrimSpace(parts[i])
			if v == "" {
				continue
			}
			if strings.HasPrefix(v, "/") && strings.HasSuffix(v, ".md") {
				paths = append(paths, util.StripMarkdownEscapes(v))
			}
			break
		}
	}
	return paths
}

// ── internal helpers ──────────────────────────────────────────────────────────

func (g *Vault) readNoteLocked(p Path) ([]byte, error) {
	if err := validateNoteName(p.Base()); err != nil {
		return nil, err
	}
	abs, err := g.osPath(p)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, mapOSError(err)
	}
	return data, nil
}

// writeNoteLocked writes data to the note at p and syncs the metadata DB.
// meta carries the origin and any extra fields (Title) to persist
// atomically with the write. Dir, Name, Size, UpdatedAt are always derived from
// p and the file stat and do not need to be set by the caller.
// Returns isNew=true when the file did not exist on disk before this call, and
// modTime set to the file's actual modification time after the write.
// Must be called with g.mu held for writing.
func (g *Vault) writeNoteLocked(p Path, data []byte, meta Note) (isNew bool, modTime time.Time, err error) {
	if err = validateNoteName(p.Base()); err != nil {
		return false, time.Time{}, err
	}
	if !util.IsValidText(data) {
		return false, time.Time{}, ErrBinaryContent
	}
	abs, absErr := g.osPath(p)
	if absErr != nil {
		return false, time.Time{}, absErr
	}

	// Probe before writing so callers can emit the correct vault event type.
	_, statErr := os.Stat(abs)
	isNew = errors.Is(statErr, fs.ErrNotExist)

	parent := filepath.Dir(abs)
	if err = os.MkdirAll(parent, 0o750); err != nil {
		return false, time.Time{}, fmt.Errorf("vault: mkdir for %q: %w", p.String(), err)
	}
	if g.ensureWatch != nil {
		g.ensureWatch(parent)
	}

	// ── Phase 1: write filesystem ─────────────────────────────────────────────
	if err = os.WriteFile(abs, data, 0o644); err != nil {
		return false, time.Time{}, fmt.Errorf("vault: write %q: %w", p.String(), err)
	}

	// ── Phase 2: sync metadata ────────────────────────────────────────────────
	info, statErr2 := os.Stat(abs)
	if statErr2 != nil {
		return false, time.Time{}, fmt.Errorf("vault: stat after write %q: %w", p.String(), statErr2)
	}
	meta.Dir = p.Dir()
	meta.Name = p.Base()
	meta.Size = info.Size()
	meta.UpdatedAt = info.ModTime()
	if meta.Origin == "" {
		meta.Origin = OriginAPI
	}
	if fm, _ := util.ParseFrontmatter(data); len(fm.Tags) > 0 {
		meta.Tags = fm.Tags
	}
	if dbErr := dbUpsert(g.db, meta); dbErr != nil {
		return false, time.Time{}, fmt.Errorf("vault: metadata upsert %q: %w", p.String(), dbErr)
	}

	return isNew, info.ModTime(), nil
}

func validateNoteName(name string) error {
	if !util.IsMarkdownPath(name) {
		return fmt.Errorf("%w: %q", ErrUnsupportedType, name)
	}
	return nil
}

func (g *Vault) osPath(p Path) (string, error) {
	vaultPath := p.String()
	if !strings.HasPrefix(vaultPath, "/") {
		return "", fmt.Errorf("%w: %q is not a vault-absolute path (must start with '/')", ErrInvalidPath, vaultPath)
	}
	stripped := vaultPath[1:] // remove leading "/"
	joined := filepath.Join(g.root, filepath.FromSlash(stripped))
	if !strings.HasPrefix(joined, g.root+string(os.PathSeparator)) && joined != g.root {
		return "", ErrInvalidPath
	}
	return joined, nil
}

func mapOSError(err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		return ErrFileNotFound
	}
	return err
}

func expandHome(p string) (string, error) {
	if !strings.HasPrefix(p, "~") {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, p[1:]), nil
}

func toSlash(p string) string { return filepath.ToSlash(p) }
