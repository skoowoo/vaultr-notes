package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver
)

const currentDBVersion = 23

// schema is the notes table DDL.
//
// Each row represents one markdown note.
//
//	dir         — vault-absolute unix path of the containing directory (e.g. "/journal/2026")
//	             Root-level notes use dir = "/".
//	name        — base filename (e.g. "april.md")
//	(dir, name) — composite unique key; primary key is an auto-increment id.
//	size        — file size in bytes
//	created_at  — first-write Unix nanoseconds (never overwritten on UPDATE)
//	updated_at  — last-write  Unix nanoseconds
//	indexed     — 1 after the bleve search index has successfully indexed this note
const schema = `
CREATE TABLE IF NOT EXISTS notes (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    dir           TEXT    NOT NULL,
    name          TEXT    NOT NULL,
    size          INTEGER NOT NULL DEFAULT 0,
    created_at    INTEGER NOT NULL,
    updated_at    INTEGER NOT NULL,
    indexed       INTEGER NOT NULL DEFAULT 0,
    kind          TEXT    NOT NULL DEFAULT '',
    title         TEXT    NOT NULL DEFAULT '',
    pinned        INTEGER NOT NULL DEFAULT 0,
    compile_count INTEGER NOT NULL DEFAULT 0,
    tags          TEXT    NOT NULL DEFAULT '',
    UNIQUE(dir, name)
);
`

// knowledgeDepsSchema creates the knowledge_deps join table.
// Each row records that a knowledge note (identified by knowledge_dir+knowledge_name)
// was compiled from a specific source note (source_dir+source_name).
// The relationship is many-to-many: one knowledge note may aggregate many source notes,
// and one source note may feed into many knowledge notes.
const knowledgeDepsSchema = `
CREATE TABLE IF NOT EXISTS knowledge_deps (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    knowledge_dir   TEXT    NOT NULL,
    knowledge_name  TEXT    NOT NULL,
    source_dir      TEXT    NOT NULL,
    source_name     TEXT    NOT NULL,
    created_at      INTEGER NOT NULL,
    UNIQUE(knowledge_dir, knowledge_name, source_dir, source_name)
);
`

// indexDepsSchema creates the index_deps join table.
// Each row records that an index note (identified by index_dir+index_name)
// lists a specific knowledge note (knowledge_dir+knowledge_name) in its table.
// The relationship is many-to-many: one index may cover many knowledge notes,
// and one knowledge note may appear in many indexes.
const indexDepsSchema = `
CREATE TABLE IF NOT EXISTS index_deps (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    index_dir       TEXT    NOT NULL,
    index_name      TEXT    NOT NULL,
    knowledge_dir   TEXT    NOT NULL,
    knowledge_name  TEXT    NOT NULL,
    created_at      INTEGER NOT NULL,
    UNIQUE(index_dir, index_name, knowledge_dir, knowledge_name)
);
`

// knowledgeLinksSchema creates the knowledge_links join table.
// Each row records a directed wikilink edge from one knowledge note to another.
// source → target means the source note's body contains [[target]] (or [[target|…]]).
const knowledgeLinksSchema = `
CREATE TABLE IF NOT EXISTS knowledge_links (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    source_dir          TEXT    NOT NULL,
    source_name         TEXT    NOT NULL,
    target_dir          TEXT    NOT NULL,
    target_name         TEXT    NOT NULL,
    source_entity_type  TEXT    NOT NULL DEFAULT '',
    created_at          INTEGER NOT NULL,
    UNIQUE(source_dir, source_name, target_dir, target_name)
);
`

// noteAssetsSchema stores extracted resources attached to a note
// (cover now; body images / audio / video later). filename is a vault-wide
// unique basename, matching wiki-link / /api/images/serve?name= lookup.
// ord orders multiple rows of the same kind (cover uses 0).
const noteAssetsSchema = `
CREATE TABLE IF NOT EXISTS note_assets (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    note_dir   TEXT    NOT NULL,
    note_name  TEXT    NOT NULL,
    kind       TEXT    NOT NULL,
    filename   TEXT    NOT NULL DEFAULT '',
    source_url TEXT    NOT NULL DEFAULT '',
    ord        INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    UNIQUE(note_dir, note_name, kind, filename)
);
`

const noteAssetLookupBatchSize = 400

// imagesSchema creates the images metadata table.
// dir  — vault-absolute path of the directory containing the image (e.g. "/_assets/202501")
// name — filename with extension (e.g. "photo.png")
// ext  — lowercase extension including dot (e.g. ".png")
const imagesSchema = `
CREATE TABLE IF NOT EXISTS images (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    dir          TEXT    NOT NULL,
    name         TEXT    NOT NULL,
    ext          TEXT    NOT NULL,
    size         INTEGER NOT NULL DEFAULT 0,
    created_at   INTEGER NOT NULL,
    updated_at   INTEGER NOT NULL,
    linked_notes TEXT    NOT NULL DEFAULT '',
    UNIQUE(dir, name)
);
`

// openDB opens (or creates) the SQLite database at <vaultRoot>/.vaultr/meta.db
// and applies the full schema on every open (all DDL uses IF NOT EXISTS).
func openDB(vaultRoot string) (*sql.DB, error) {
	dbPath := vaultDBPath(vaultRoot)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("storage: open db: %w", err)
	}
	// Single writer avoids SQLITE_BUSY under concurrent goroutines.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("storage: enable WAL: %w", err)
	}

	var version int
	db.QueryRow("PRAGMA user_version").Scan(&version) //nolint:errcheck

	if version < currentDBVersion {
		if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", currentDBVersion)); err != nil {
			db.Close()
			return nil, fmt.Errorf("storage: set schema version: %w", err)
		}
	}

	for _, step := range []struct {
		ddl string
		msg string
	}{
		{schema, "apply schema"},
		{imagesSchema, "create images table"},
		{noteAssetsSchema, "create note_assets table"},
		{knowledgeDepsSchema, "create knowledge_deps table"},
		{indexDepsSchema, "create index_deps table"},
		{knowledgeLinksSchema, "create knowledge_links table"},
		{`CREATE INDEX IF NOT EXISTS idx_kl_source ON knowledge_links(source_dir, source_name)`, "create idx_kl_source"},
		{`CREATE INDEX IF NOT EXISTS idx_kl_target ON knowledge_links(target_dir, target_name)`, "create idx_kl_target"},
		{`CREATE INDEX IF NOT EXISTS idx_kd_knowledge ON knowledge_deps(knowledge_dir, knowledge_name)`, "create idx_kd_knowledge"},
		{`CREATE INDEX IF NOT EXISTS idx_kd_source ON knowledge_deps(source_dir, source_name)`, "create idx_kd_source"},
		{`CREATE INDEX IF NOT EXISTS idx_id_index ON index_deps(index_dir, index_name)`, "create idx_id_index"},
		{`CREATE INDEX IF NOT EXISTS idx_id_knowledge ON index_deps(knowledge_dir, knowledge_name)`, "create idx_id_knowledge"},
		{`CREATE INDEX IF NOT EXISTS idx_note_assets_lookup ON note_assets(kind, note_dir, note_name, ord, id)`, "create idx_note_assets_lookup"},
		{`CREATE INDEX IF NOT EXISTS idx_note_assets_filename ON note_assets(filename)`, "create idx_note_assets_filename"},
	} {
		if _, err := db.Exec(step.ddl); err != nil {
			db.Close()
			return nil, fmt.Errorf("storage: %s: %w", step.msg, err)
		}
	}

	return db, nil
}

// dbListByDir returns notes whose dir exactly matches the given dir.
// dir must be a vault-absolute path (e.g. "/" or "/journal/2026").
// Ordering and pagination are controlled by opts.
func dbListByDir(db *sql.DB, dir string, opts ListOptions) ([]Note, error) {

	orderBy := "name ASC"
	if opts.SortByTime {
		orderBy = "updated_at DESC"
	}

	var where []string
	var args []any

	where = append(where, "dir = ?")
	args = append(args, dir)

	if opts.OnlyUnindexed {
		where = append(where, "indexed = 0")
	}
	if len(opts.OnlyKinds) > 0 {
		placeholders := strings.Repeat("?,", len(opts.OnlyKinds))
		placeholders = placeholders[:len(placeholders)-1]
		where = append(where, "kind IN ("+placeholders+")")
		for _, k := range opts.OnlyKinds {
			args = append(args, string(k))
		}
	}
	if len(opts.ExcludeKinds) > 0 {
		placeholders := strings.Repeat("?,", len(opts.ExcludeKinds))
		placeholders = placeholders[:len(placeholders)-1]
		where = append(where, "kind NOT IN ("+placeholders+")")
		for _, k := range opts.ExcludeKinds {
			args = append(args, string(k))
		}
	}
	if !opts.After.IsZero() {
		where = append(where, "updated_at >= ?")
		args = append(args, opts.After.UnixNano())
	}
	if !opts.Before.IsZero() {
		where = append(where, "updated_at < ?")
		args = append(args, opts.Before.UnixNano())
	}

	query := fmt.Sprintf(`
		SELECT dir, name, size, created_at, updated_at, indexed, kind, title, pinned, compile_count, tags
		FROM notes
		WHERE %s
		ORDER BY %s`, strings.Join(where, " AND "), orderBy)

	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("storage: list dir %q: %w", dir, err)
	}
	defer rows.Close()

	return dbScan_(rows)
}

// dbListAll returns every note in the vault regardless of dir.
// Used for vault-wide operations such as the search index backfill.
// Ordering and pagination are controlled by opts.
func dbListAll(db *sql.DB, opts ListOptions) ([]Note, error) {
	orderBy := "dir ASC, name ASC"
	if opts.SortByTime {
		orderBy = "updated_at DESC"
	}

	var where []string
	var args []any

	if opts.OnlyUnindexed {
		where = append(where, "indexed = 0")
	}
	if len(opts.OnlyKinds) > 0 {
		placeholders := strings.Repeat("?,", len(opts.OnlyKinds))
		placeholders = placeholders[:len(placeholders)-1]
		where = append(where, "kind IN ("+placeholders+")")
		for _, k := range opts.OnlyKinds {
			args = append(args, string(k))
		}
	}
	if len(opts.ExcludeKinds) > 0 {
		placeholders := strings.Repeat("?,", len(opts.ExcludeKinds))
		placeholders = placeholders[:len(placeholders)-1]
		where = append(where, "kind NOT IN ("+placeholders+")")
		for _, k := range opts.ExcludeKinds {
			args = append(args, string(k))
		}
	}
	if !opts.After.IsZero() {
		where = append(where, "updated_at >= ?")
		args = append(args, opts.After.UnixNano())
	}
	if !opts.Before.IsZero() {
		where = append(where, "updated_at < ?")
		args = append(args, opts.Before.UnixNano())
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT dir, name, size, created_at, updated_at, indexed, kind, title, pinned, compile_count, tags
		FROM notes %s
		ORDER BY %s`, whereClause, orderBy)

	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("storage: list all notes: %w", err)
	}
	defer rows.Close()

	return dbScan_(rows)
}

// dbListRecentByKindByCreatedDesc returns up to limit notes with the given
// kind, ordered by created_at descending (newest first).
func dbListRecentByKindByCreatedDesc(db *sql.DB, kind Kind, limit int) ([]Note, error) {
	if limit <= 0 {
		limit = 7
	}
	rows, err := db.Query(`
		SELECT dir, name, size, created_at, updated_at, indexed, kind, title, pinned, compile_count, tags
		FROM notes
		WHERE kind = ?
		ORDER BY created_at DESC
		LIMIT ?`, string(kind), limit)
	if err != nil {
		return nil, fmt.Errorf("storage: list recent notes by kind %q: %w", kind, err)
	}
	defer rows.Close()
	return dbScan_(rows)
}

// dbGet returns the single metadata row for a note path.
func dbGet(db *sql.DB, p Path) (Note, error) {
	var n Note
	var size, createdNs, updNs int64
	var indexed, pinned int
	var tagsRaw string
	err := db.QueryRow(`
		SELECT dir, name, size, created_at, updated_at, indexed, kind, title, pinned, compile_count, tags
		FROM notes WHERE dir = ? AND name = ?`, p.Dir(), p.Base()).
		Scan(&n.Dir, &n.Name, &size, &createdNs, &updNs, &indexed, &n.Kind, &n.Title, &pinned, &n.CompileCount, &tagsRaw)
	if errors.Is(err, sql.ErrNoRows) {
		return Note{}, ErrMetadataNotFound
	}
	if err != nil {
		return Note{}, err
	}
	n.Size = size
	n.CreatedAt = time.Unix(0, createdNs)
	n.UpdatedAt = time.Unix(0, updNs)
	n.Indexed = indexed == 1
	n.Pinned = pinned == 1
	n.Tags = splitTags(tagsRaw)
	return n, nil
}

// dbGetByNames returns all notes whose filename matches any entry in names.
// Results are ordered by updated_at DESC.
func dbGetByNames(db *sql.DB, names []string) ([]Note, error) {
	if len(names) == 0 {
		return nil, nil
	}
	placeholders := strings.Repeat("?,", len(names))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(names))
	for i, n := range names {
		args[i] = n
	}
	rows, err := db.Query(`
		SELECT dir, name, size, created_at, updated_at, indexed, kind, title, pinned, compile_count, tags
		FROM notes WHERE name IN (`+placeholders+`) ORDER BY updated_at DESC`, args...)
	if err != nil {
		return nil, fmt.Errorf("storage: notes by names: %w", err)
	}
	defer rows.Close()
	return dbScan_(rows)
}

// dbGetByPaths returns the metadata row for each (dir, name) pair in paths.
// Missing rows are silently skipped. Results are ordered by updated_at DESC.
func dbGetByPaths(db *sql.DB, paths []Path) ([]Note, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	// Build: WHERE (dir=? AND name=?) OR (dir=? AND name=?) …
	clauses := make([]string, len(paths))
	args := make([]any, 0, len(paths)*2)
	for i, p := range paths {
		clauses[i] = "(dir = ? AND name = ?)"
		args = append(args, p.Dir(), p.Base())
	}
	rows, err := db.Query(`
		SELECT dir, name, size, created_at, updated_at, indexed, kind, title, pinned, compile_count, tags
		FROM notes WHERE `+strings.Join(clauses, " OR ")+`
		ORDER BY updated_at DESC`, args...)
	if err != nil {
		return nil, fmt.Errorf("storage: notes by paths: %w", err)
	}
	defer rows.Close()
	return dbScan_(rows)
}

// dbGetByName returns every row whose name column equals the given filename
func dbGetByName(db *sql.DB, name string) ([]Note, error) {
	rows, err := db.Query(`
		SELECT dir, name, size, created_at, updated_at, indexed, kind, title, pinned, compile_count, tags
		FROM notes
		WHERE name = ?
		ORDER BY updated_at DESC, dir ASC`, name)
	if err != nil {
		return nil, fmt.Errorf("storage: notes by name %q: %w", name, err)
	}
	defer rows.Close()

	return dbScan_(rows)
}

// dbScan_ reads all rows from a notes query result into a []Note slice.
func dbScan_(rows *sql.Rows) ([]Note, error) {
	var notes []Note
	for rows.Next() {
		var n Note
		var createdNs, updNs int64
		var indexed, pinned int
		var tagsRaw string
		if err := rows.Scan(&n.Dir, &n.Name, &n.Size, &createdNs, &updNs, &indexed, &n.Kind, &n.Title, &pinned, &n.CompileCount, &tagsRaw); err != nil {
			return nil, fmt.Errorf("storage: scan note: %w", err)
		}
		n.CreatedAt = time.Unix(0, createdNs)
		n.UpdatedAt = time.Unix(0, updNs)
		n.Indexed = indexed == 1
		n.Pinned = pinned == 1
		n.Tags = splitTags(tagsRaw)
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// dbUpsert inserts or updates the metadata row for a note.
// On conflict, title is only overwritten when the incoming
// value is non-empty, so plain content updates never clobber plugin-set metadata.
func dbUpsert(db *sql.DB, n Note) error {
	now := time.Now().UnixNano()
	updNs := now
	if !n.UpdatedAt.IsZero() {
		updNs = n.UpdatedAt.UnixNano()
	}
	_, err := db.Exec(`
		INSERT INTO notes(dir, name, size, created_at, updated_at, indexed, kind, title, tags)
		VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?)
		ON CONFLICT(dir, name) DO UPDATE SET
		    size        = excluded.size,
		    updated_at  = excluded.updated_at,
		    indexed     = CASE WHEN excluded.updated_at > notes.updated_at THEN 0 ELSE notes.indexed END,
		    title       = CASE WHEN excluded.title != '' THEN excluded.title ELSE notes.title END,
		    tags        = excluded.tags`,
		n.Dir, n.Name, n.Size, now, updNs, string(n.Kind), n.Title, joinTags(n.Tags),
	)
	return err
}

// dbMarkIndexed sets indexed = 1 for the note path
func dbMarkIndexed(db *sql.DB, p Path) error {
	_, err := db.Exec(`UPDATE notes SET indexed = 1 WHERE dir = ? AND name = ?`, p.Dir(), p.Base())
	return err
}

// dbResetAllIndexed sets indexed = 0 for every note, so the next backfill
// re-indexes the entire vault (used when the search index is rebuilt from scratch).
func dbResetAllIndexed(db *sql.DB) error {
	_, err := db.Exec(`UPDATE notes SET indexed = 0`)
	return err
}

// dbListDirs returns every distinct dir in the notes table together with its
// note count, ordered alphabetically.
func dbListDirs(db *sql.DB) ([]DirSummary, error) {
	rows, err := db.Query(`SELECT dir, COUNT(*) FROM notes GROUP BY dir ORDER BY dir ASC`)
	if err != nil {
		return nil, fmt.Errorf("storage: list dirs: %w", err)
	}
	defer rows.Close()
	var dirs []DirSummary
	for rows.Next() {
		var d DirSummary
		if err := rows.Scan(&d.Dir, &d.Count); err != nil {
			return nil, fmt.Errorf("storage: scan dir: %w", err)
		}
		dirs = append(dirs, d)
	}
	return dirs, rows.Err()
}

// dbCount returns the total number of notes in the metadata database.
func dbCount(db *sql.DB) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM notes`).Scan(&n)
	return n, err
}

func dbCountByKind(db *sql.DB, onlyKind, excludeKind string) (int, error) {
	var n int
	var err error
	switch {
	case onlyKind != "":
		err = db.QueryRow(`SELECT COUNT(*) FROM notes WHERE kind = ?`, onlyKind).Scan(&n)
	case excludeKind != "":
		err = db.QueryRow(`SELECT COUNT(*) FROM notes WHERE kind != ?`, excludeKind).Scan(&n)
	default:
		err = db.QueryRow(`SELECT COUNT(*) FROM notes`).Scan(&n)
	}
	return n, err
}

// dbDelete removes the metadata row for the note at p and cascades to every
// relation table it might participate in, regardless of its kind:
// note_assets (owned resources), knowledge_deps (as the knowledge side or
// the source side), index_deps (as the index side or the knowledge side),
// and knowledge_links (as source or target of a wikilink edge). It never
// deletes another note — only join-table rows that reference p.
//
// A source note left with no remaining knowledge dependents has its
// compile_count reset to 0 so it becomes eligible for re-compilation.
//
// Everything runs in one transaction so a partial failure can't leave
// orphaned relation rows or a stale compile_count behind.
func dbDelete(db *sql.DB, p Path) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("storage: delete: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Sources this note depends on (if it's a knowledge note) — recorded before
	// the dependency rows are deleted, so we can recheck each one afterwards.
	rows, err := tx.Query(
		`SELECT source_dir, source_name FROM knowledge_deps WHERE knowledge_dir = ? AND knowledge_name = ?`,
		p.Dir(), p.Base(),
	)
	if err != nil {
		return fmt.Errorf("storage: delete: knowledge_deps lookup: %w", err)
	}
	sources, err := dbScanPaths(rows)
	if err != nil {
		return fmt.Errorf("storage: delete: knowledge_deps scan: %w", err)
	}

	if _, err := tx.Exec(`DELETE FROM note_assets WHERE note_dir = ? AND note_name = ?`, p.Dir(), p.Base()); err != nil {
		return fmt.Errorf("storage: delete: note_assets: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM notes WHERE dir = ? AND name = ?`, p.Dir(), p.Base()); err != nil {
		return fmt.Errorf("storage: delete: notes: %w", err)
	}
	if _, err := tx.Exec(
		`DELETE FROM knowledge_deps WHERE knowledge_dir = ? AND knowledge_name = ?`,
		p.Dir(), p.Base(),
	); err != nil {
		return fmt.Errorf("storage: delete: knowledge_deps (as knowledge): %w", err)
	}
	if _, err := tx.Exec(
		`DELETE FROM knowledge_deps WHERE source_dir = ? AND source_name = ?`,
		p.Dir(), p.Base(),
	); err != nil {
		return fmt.Errorf("storage: delete: knowledge_deps (as source): %w", err)
	}
	if _, err := tx.Exec(
		`DELETE FROM index_deps WHERE index_dir = ? AND index_name = ?`,
		p.Dir(), p.Base(),
	); err != nil {
		return fmt.Errorf("storage: delete: index_deps (as index): %w", err)
	}
	if _, err := tx.Exec(
		`DELETE FROM index_deps WHERE knowledge_dir = ? AND knowledge_name = ?`,
		p.Dir(), p.Base(),
	); err != nil {
		return fmt.Errorf("storage: delete: index_deps (as knowledge): %w", err)
	}
	if _, err := tx.Exec(
		`DELETE FROM knowledge_links WHERE (source_dir = ? AND source_name = ?) OR (target_dir = ? AND target_name = ?)`,
		p.Dir(), p.Base(), p.Dir(), p.Base(),
	); err != nil {
		return fmt.Errorf("storage: delete: knowledge_links: %w", err)
	}

	for _, src := range sources {
		var remaining int
		if err := tx.QueryRow(
			`SELECT COUNT(*) FROM knowledge_deps WHERE source_dir = ? AND source_name = ?`,
			src.Dir(), src.Base(),
		).Scan(&remaining); err != nil {
			return fmt.Errorf("storage: delete: recheck source %q: %w", src.String(), err)
		}
		if remaining == 0 {
			if _, err := tx.Exec(
				`UPDATE notes SET compile_count = 0 WHERE dir = ? AND name = ?`,
				src.Dir(), src.Base(),
			); err != nil {
				return fmt.Errorf("storage: delete: reset compile_count %q: %w", src.String(), err)
			}
		}
	}

	return tx.Commit()
}

// dbMove updates the metadata row for the note at old to live under newDir
// (same filename) and cascades the directory change to every relation table
// it might participate in, regardless of its kind — the same table list
// dbDelete cascades to, but UPDATE instead of DELETE since nothing is being
// removed: note_assets, knowledge_deps (as the knowledge side or the source
// side), index_deps (as the index side or the knowledge side), and
// knowledge_links (as source or target).
//
// Returns ErrAlreadyExists if a note already occupies (newDir, old.Base())
// without touching anything.
//
// This only keeps the DB in sync for the moved note itself. A dependent
// knowledge/index note's frontmatter or table may still hold old's full path
// as a literal string (see skills/vaultr-compile-note and
// skills/vaultr-index-knowledge) — the compile plugin resyncs knowledge_deps/
// index_deps from that text on every save, so leaving it stale would
// eventually overwrite this update. Vault.MoveNote handles that separately by
// rewriting the literal path in whichever dependent notes reference old.
func dbMove(db *sql.DB, old Path, newDir string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("storage: move: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	oldDir, name := old.Dir(), old.Base()

	var occupied int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM notes WHERE dir = ? AND name = ?`, newDir, name,
	).Scan(&occupied); err != nil {
		return fmt.Errorf("storage: move: destination check: %w", err)
	}
	if occupied > 0 {
		return ErrAlreadyExists
	}

	if _, err := tx.Exec(`UPDATE notes SET dir = ? WHERE dir = ? AND name = ?`, newDir, oldDir, name); err != nil {
		return fmt.Errorf("storage: move: notes: %w", err)
	}
	if _, err := tx.Exec(
		`UPDATE note_assets SET note_dir = ? WHERE note_dir = ? AND note_name = ?`,
		newDir, oldDir, name,
	); err != nil {
		return fmt.Errorf("storage: move: note_assets: %w", err)
	}
	if _, err := tx.Exec(
		`UPDATE knowledge_deps SET knowledge_dir = ? WHERE knowledge_dir = ? AND knowledge_name = ?`,
		newDir, oldDir, name,
	); err != nil {
		return fmt.Errorf("storage: move: knowledge_deps (as knowledge): %w", err)
	}
	if _, err := tx.Exec(
		`UPDATE knowledge_deps SET source_dir = ? WHERE source_dir = ? AND source_name = ?`,
		newDir, oldDir, name,
	); err != nil {
		return fmt.Errorf("storage: move: knowledge_deps (as source): %w", err)
	}
	if _, err := tx.Exec(
		`UPDATE index_deps SET index_dir = ? WHERE index_dir = ? AND index_name = ?`,
		newDir, oldDir, name,
	); err != nil {
		return fmt.Errorf("storage: move: index_deps (as index): %w", err)
	}
	if _, err := tx.Exec(
		`UPDATE index_deps SET knowledge_dir = ? WHERE knowledge_dir = ? AND knowledge_name = ?`,
		newDir, oldDir, name,
	); err != nil {
		return fmt.Errorf("storage: move: index_deps (as knowledge): %w", err)
	}
	if _, err := tx.Exec(
		`UPDATE knowledge_links SET source_dir = ? WHERE source_dir = ? AND source_name = ?`,
		newDir, oldDir, name,
	); err != nil {
		return fmt.Errorf("storage: move: knowledge_links (as source): %w", err)
	}
	if _, err := tx.Exec(
		`UPDATE knowledge_links SET target_dir = ? WHERE target_dir = ? AND target_name = ?`,
		newDir, oldDir, name,
	); err != nil {
		return fmt.Errorf("storage: move: knowledge_links (as target): %w", err)
	}

	return tx.Commit()
}

// dbClearAll removes every row from the notes table.
// Used by ScanAndRegisterFull to start from a clean slate.
func dbClearAll(db *sql.DB) error {
	_, err := db.Exec(`DELETE FROM notes`)
	return err
}

// dbInsertFull inserts a note row with all fields explicitly set.
// Uses ON CONFLICT DO UPDATE to preserve user-set fields (pinned) across vault rescans.
func dbInsertFull(db *sql.DB, n Note) error {
	createdNs := n.UpdatedAt.UnixNano()
	if !n.CreatedAt.IsZero() {
		createdNs = n.CreatedAt.UnixNano()
	}
	updNs := n.UpdatedAt.UnixNano()
	_, err := db.Exec(`
		INSERT INTO notes(dir, name, size, created_at, updated_at, indexed, kind, title, compile_count, tags)
		VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?, ?)
		ON CONFLICT(dir, name) DO UPDATE SET
		    size          = excluded.size,
		    created_at    = excluded.created_at,
		    updated_at    = excluded.updated_at,
		    indexed       = 0,
		    kind          = excluded.kind,
		    title         = excluded.title,
		    compile_count = excluded.compile_count,
		    tags          = excluded.tags`,
		n.Dir, n.Name, n.Size, createdNs, updNs, string(n.Kind), n.Title, n.CompileCount, joinTags(n.Tags),
	)
	return err
}

// dbSetPinned sets the pinned flag for the note at p.
func dbSetPinned(db *sql.DB, p Path, pinned bool) error {
	v := 0
	if pinned {
		v = 1
	}
	_, err := db.Exec(`UPDATE notes SET pinned = ? WHERE dir = ? AND name = ?`, v, p.Dir(), p.Base())
	return err
}

// dbListPinned returns all notes where pinned = 1, ordered by updated_at DESC.
func dbListPinned(db *sql.DB) ([]Note, error) {
	rows, err := db.Query(`
		SELECT dir, name, size, created_at, updated_at, indexed, kind, title, pinned, compile_count, tags
		FROM notes WHERE pinned = 1
		ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("storage: list pinned notes: %w", err)
	}
	defer rows.Close()
	return dbScan_(rows)
}

// dbSetTitle sets the title column for the note path.
func dbSetTitle(db *sql.DB, p Path, title string) error {
	_, err := db.Exec(`UPDATE notes SET title = ? WHERE dir = ? AND name = ?`, title, p.Dir(), p.Base())
	return err
}

// dbSetCompileCount sets the compile_count column to the given value for the note at p.
func dbSetCompileCount(db *sql.DB, p Path, count int) error {
	_, err := db.Exec(`UPDATE notes SET compile_count = ? WHERE dir = ? AND name = ?`, count, p.Dir(), p.Base())
	return err
}

// dbSetKindKnowledge sets kind = 'knowledge', compile_count, tags, and optionally title for the note at p.
// title may be empty to leave the column unchanged.
func dbSetKindKnowledge(db *sql.DB, p Path, title string, compileCount int, tags []string) error {
	if title != "" {
		_, err := db.Exec(
			`UPDATE notes SET kind = ?, title = ?, compile_count = ?, tags = ? WHERE dir = ? AND name = ?`,
			string(KindKnowledge), title, compileCount, joinTags(tags), p.Dir(), p.Base(),
		)
		return err
	}
	_, err := db.Exec(
		`UPDATE notes SET kind = ?, compile_count = ?, tags = ? WHERE dir = ? AND name = ?`,
		string(KindKnowledge), compileCount, joinTags(tags), p.Dir(), p.Base(),
	)
	return err
}

// dbSetKindIndex sets kind = 'index' and optionally title for the note at p.
// title may be empty to leave the column unchanged.
func dbSetKindIndex(db *sql.DB, p Path, title string) error {
	if title != "" {
		_, err := db.Exec(
			`UPDATE notes SET kind = ?, title = ? WHERE dir = ? AND name = ?`,
			string(KindIndex), title, p.Dir(), p.Base(),
		)
		return err
	}
	_, err := db.Exec(
		`UPDATE notes SET kind = ? WHERE dir = ? AND name = ?`,
		string(KindIndex), p.Dir(), p.Base(),
	)
	return err
}

// dbUpsertKnowledgeNote inserts or updates a compile-plugin knowledge note, always
// overwriting kind, title, size, and updated_at on conflict.
func dbUpsertKnowledgeNote(db *sql.DB, n Note) error {
	now := time.Now().UnixNano()
	updNs := now
	if !n.UpdatedAt.IsZero() {
		updNs = n.UpdatedAt.UnixNano()
	}
	_, err := db.Exec(`
		INSERT INTO notes(dir, name, size, created_at, updated_at, indexed, kind, title, tags)
		VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?)
		ON CONFLICT(dir, name) DO UPDATE SET
		    size       = excluded.size,
		    updated_at = excluded.updated_at,
		    indexed    = 0,
		    kind       = excluded.kind,
		    title      = excluded.title,
		    tags       = excluded.tags`,
		n.Dir, n.Name, n.Size, now, updNs, string(n.Kind), n.Title, joinTags(n.Tags),
	)
	return err
}

// ── note_assets DB helpers ────────────────────────────────────────────────────

func dbUpsertNoteAsset(db *sql.DB, a NoteAsset) error {
	now := time.Now().UnixNano()
	if !a.CreatedAt.IsZero() {
		now = a.CreatedAt.UnixNano()
	}
	_, err := db.Exec(`
		INSERT INTO note_assets(note_dir, note_name, kind, filename, source_url, ord, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(note_dir, note_name, kind, filename) DO UPDATE SET
		    source_url = excluded.source_url,
		    ord        = excluded.ord`,
		a.NoteDir, a.NoteName, string(a.Kind), a.Filename, a.SourceURL, a.Ord, now,
	)
	return err
}

func dbReplaceNoteAssets(db *sql.DB, p Path, kind NoteAssetKind, assets []NoteAsset) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("storage: note_assets replace: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(
		`DELETE FROM note_assets WHERE note_dir = ? AND note_name = ? AND kind = ?`,
		p.Dir(), p.Base(), string(kind),
	); err != nil {
		return fmt.Errorf("storage: note_assets replace: delete: %w", err)
	}

	now := time.Now().UnixNano()
	for i, a := range assets {
		createdNs := now
		if !a.CreatedAt.IsZero() {
			createdNs = a.CreatedAt.UnixNano()
		}
		if a.Kind == "" {
			a.Kind = kind
		}
		if a.Ord == 0 {
			a.Ord = i
		}
		if _, err := tx.Exec(`
			INSERT INTO note_assets(note_dir, note_name, kind, filename, source_url, ord, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			p.Dir(), p.Base(), string(a.Kind), a.Filename, a.SourceURL, a.Ord, createdNs,
		); err != nil {
			return fmt.Errorf("storage: note_assets replace: insert %q: %w", a.Filename, err)
		}
	}
	return tx.Commit()
}

func dbDeleteNoteAssetsByFilename(db *sql.DB, filename string) error {
	_, err := db.Exec(`DELETE FROM note_assets WHERE filename = ?`, filename)
	return err
}

func dbScanNoteAsset(scan func(dest ...any) error) (NoteAsset, error) {
	var a NoteAsset
	var createdNs int64
	var kindRaw string
	err := scan(&a.NoteDir, &a.NoteName, &kindRaw, &a.Filename, &a.SourceURL, &a.Ord, &createdNs)
	if err != nil {
		return NoteAsset{}, err
	}
	a.Kind = NoteAssetKind(kindRaw)
	a.CreatedAt = time.Unix(0, createdNs)
	return a, nil
}

func dbGetNoteAsset(db *sql.DB, p Path, kind NoteAssetKind) (NoteAsset, error) {
	row := db.QueryRow(`
		SELECT note_dir, note_name, kind, filename, source_url, ord, created_at
		FROM note_assets
		WHERE note_dir = ? AND note_name = ? AND kind = ? AND filename != ''
		ORDER BY ord ASC, id ASC
		LIMIT 1`,
		p.Dir(), p.Base(), string(kind),
	)
	a, err := dbScanNoteAsset(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return NoteAsset{}, ErrMetadataNotFound
	}
	if err != nil {
		return NoteAsset{}, err
	}
	return a, nil
}

func dbListNoteAssets(db *sql.DB, p Path, kind NoteAssetKind) ([]NoteAsset, error) {
	rows, err := db.Query(`
		SELECT note_dir, note_name, kind, filename, source_url, ord, created_at
		FROM note_assets
		WHERE note_dir = ? AND note_name = ? AND kind = ? AND filename != ''
		ORDER BY ord ASC, id ASC`,
		p.Dir(), p.Base(), string(kind),
	)
	if err != nil {
		return nil, fmt.Errorf("storage: list note_assets: %w", err)
	}
	defer rows.Close()
	var out []NoteAsset
	for rows.Next() {
		a, err := dbScanNoteAsset(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func dbGetNoteAssetsByKind(db *sql.DB, paths []Path, kind NoteAssetKind) (map[string]string, error) {
	out := make(map[string]string, len(paths))
	for start := 0; start < len(paths); start += noteAssetLookupBatchSize {
		end := start + noteAssetLookupBatchSize
		if end > len(paths) {
			end = len(paths)
		}
		batch := paths[start:end]
		clauses := make([]string, len(batch))
		args := make([]any, 0, len(batch)*2+1)
		for i, p := range batch {
			clauses[i] = "(na.note_dir = ? AND na.note_name = ?)"
			args = append(args, p.Dir(), p.Base())
		}
		args = append(args, string(kind))
		rows, err := db.Query(`
			SELECT na.note_dir, na.note_name, na.filename FROM note_assets na
			WHERE (`+strings.Join(clauses, " OR ")+`) AND na.kind = ? AND na.filename != ''
			  AND EXISTS (SELECT 1 FROM images i WHERE i.name = na.filename)
			ORDER BY na.ord ASC, na.id ASC`, args...)
		if err != nil {
			return nil, fmt.Errorf("storage: note_assets by kind: %w", err)
		}
		for rows.Next() {
			var dir, name, filename string
			if err := rows.Scan(&dir, &name, &filename); err != nil {
				rows.Close()
				return nil, err
			}
			key := JoinPath(dir, name)
			if _, exists := out[key]; !exists {
				out[key] = filename
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	return out, nil
}

// ── image DB helpers ──────────────────────────────────────────────────────────

// dbImageUpsert inserts or updates the metadata row for an image.
// linked_notes is preserved on conflict — only explicit calls to dbImageSetLinkedNotes change it.
func dbImageUpsert(db *sql.DB, img Image) error {
	now := time.Now().UnixNano()
	updNs := now
	if !img.UpdatedAt.IsZero() {
		updNs = img.UpdatedAt.UnixNano()
	}
	createdNs := updNs
	if !img.CreatedAt.IsZero() {
		createdNs = img.CreatedAt.UnixNano()
	}
	_, err := db.Exec(`
		INSERT INTO images(dir, name, ext, size, created_at, updated_at, linked_notes)
		VALUES (?, ?, ?, ?, ?, ?, '')
		ON CONFLICT(dir, name) DO UPDATE SET
		    ext        = excluded.ext,
		    size       = excluded.size,
		    updated_at = excluded.updated_at`,
		img.Dir, img.Name, img.Ext, img.Size, createdNs, updNs,
	)
	return err
}

// dbImageGetByName returns all images whose name column equals the given filename,
// ordered by updated_at DESC.
func dbImageGetByName(db *sql.DB, name string) ([]Image, error) {
	rows, err := db.Query(`
		SELECT dir, name, ext, size, created_at, updated_at, linked_notes
		FROM images
		WHERE name = ?
		ORDER BY updated_at DESC`, name)
	if err != nil {
		return nil, fmt.Errorf("storage: images by name %q: %w", name, err)
	}
	defer rows.Close()
	return dbImageScan(rows)
}

// dbImageSetLinkedNotes sets the linked_notes column for every image with the given name.
// notes is a newline-separated list of note basenames (without .md extension).
func dbImageSetLinkedNotes(db *sql.DB, name, notes string) error {
	_, err := db.Exec(`UPDATE images SET linked_notes = ? WHERE name = ?`, notes, name)
	return err
}

// dbImageClearAllLinkedNotes resets linked_notes to ” for every image row.
func dbImageClearAllLinkedNotes(db *sql.DB) error {
	_, err := db.Exec(`UPDATE images SET linked_notes = ''`)
	return err
}

// dbImageDelete removes the metadata row for the image identified by (dir, name).
func dbImageDelete(db *sql.DB, dir, name string) error {
	_, err := db.Exec(`DELETE FROM images WHERE dir = ? AND name = ?`, dir, name)
	return err
}

// dbImageClearAll removes every row from the images table.
func dbImageClearAll(db *sql.DB) error {
	_, err := db.Exec(`DELETE FROM images`)
	return err
}

// dbImageListPaged returns images ordered by updated_at DESC with optional cursor.
// If beforeNs > 0, only images with updated_at < beforeNs are returned.
func dbImageListPaged(db *sql.DB, beforeNs int64, limit int) ([]Image, error) {
	query := `SELECT dir, name, ext, size, created_at, updated_at, linked_notes FROM images`
	var args []any
	if beforeNs > 0 {
		query += ` WHERE updated_at < ?`
		args = append(args, beforeNs)
	}
	query += ` ORDER BY updated_at DESC`
	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("storage: list images paged: %w", err)
	}
	defer rows.Close()
	return dbImageScan(rows)
}

// dbImageCount returns the total number of images in the vault.
func dbImageCount(db *sql.DB) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM images`).Scan(&n)
	return n, err
}

// ── knowledge_deps DB helpers ─────────────────────────────────────────────────

// dbReplaceKnowledgeDeps replaces, within a single transaction, all dependency
// rows for the given knowledge note with the supplied source paths.
// Passing an empty sources slice clears all existing dependencies.
func dbReplaceKnowledgeDeps(db *sql.DB, knowledge Path, sources []Path) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("storage: knowledge_deps replace: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(
		`DELETE FROM knowledge_deps WHERE knowledge_dir = ? AND knowledge_name = ?`,
		knowledge.Dir(), knowledge.Base(),
	); err != nil {
		return fmt.Errorf("storage: knowledge_deps replace: delete: %w", err)
	}

	now := time.Now().UnixNano()
	for _, src := range sources {
		if _, err := tx.Exec(`
			INSERT INTO knowledge_deps(knowledge_dir, knowledge_name, source_dir, source_name, created_at)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(knowledge_dir, knowledge_name, source_dir, source_name) DO NOTHING`,
			knowledge.Dir(), knowledge.Base(), src.Dir(), src.Base(), now,
		); err != nil {
			return fmt.Errorf("storage: knowledge_deps replace: insert (%s/%s): %w", src.Dir(), src.Base(), err)
		}
	}
	return tx.Commit()
}

// dbGetKnowledgeDeps returns all source paths that the given knowledge note depends on.
func dbGetKnowledgeDeps(db *sql.DB, knowledge Path) ([]Path, error) {
	rows, err := db.Query(`
		SELECT source_dir, source_name FROM knowledge_deps
		WHERE knowledge_dir = ? AND knowledge_name = ?
		ORDER BY source_dir ASC, source_name ASC`,
		knowledge.Dir(), knowledge.Base())
	if err != nil {
		return nil, fmt.Errorf("storage: knowledge_deps get: %w", err)
	}
	defer rows.Close()
	return dbScanPaths(rows)
}

// dbGetSourceKnowledges returns all knowledge note paths that depend on the given source note.
func dbGetSourceKnowledges(db *sql.DB, source Path) ([]Path, error) {
	rows, err := db.Query(`
		SELECT knowledge_dir, knowledge_name FROM knowledge_deps
		WHERE source_dir = ? AND source_name = ?
		ORDER BY knowledge_dir ASC, knowledge_name ASC`,
		source.Dir(), source.Base())
	if err != nil {
		return nil, fmt.Errorf("storage: knowledge_deps reverse get: %w", err)
	}
	defer rows.Close()
	return dbScanPaths(rows)
}

// ── index_deps DB helpers ─────────────────────────────────────────────────────

// dbReplaceIndexDeps replaces, within a single transaction, all dependency
// rows for the given index note with the supplied knowledge paths.
// Passing an empty knowledges slice clears all existing entries.
func dbReplaceIndexDeps(db *sql.DB, index Path, knowledges []Path) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("storage: index_deps replace: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(
		`DELETE FROM index_deps WHERE index_dir = ? AND index_name = ?`,
		index.Dir(), index.Base(),
	); err != nil {
		return fmt.Errorf("storage: index_deps replace: delete: %w", err)
	}

	now := time.Now().UnixNano()
	for _, k := range knowledges {
		if _, err := tx.Exec(`
			INSERT INTO index_deps(index_dir, index_name, knowledge_dir, knowledge_name, created_at)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(index_dir, index_name, knowledge_dir, knowledge_name) DO NOTHING`,
			index.Dir(), index.Base(), k.Dir(), k.Base(), now,
		); err != nil {
			return fmt.Errorf("storage: index_deps replace: insert (%s/%s): %w", k.Dir(), k.Base(), err)
		}
	}
	return tx.Commit()
}

// dbGetAllIndexDepCounts returns the number of knowledge deps for every index note,
// keyed by the index note's PathString (dir+"/"+name).
func dbGetAllIndexDepCounts(db *sql.DB) (map[string]int, error) {
	rows, err := db.Query(`
		SELECT index_dir, index_name, COUNT(*)
		FROM index_deps
		GROUP BY index_dir, index_name`)
	if err != nil {
		return nil, fmt.Errorf("storage: index_deps counts: %w", err)
	}
	defer rows.Close()
	counts := make(map[string]int)
	for rows.Next() {
		var dir, name string
		var n int
		if err := rows.Scan(&dir, &name, &n); err != nil {
			return nil, fmt.Errorf("storage: scan index dep count: %w", err)
		}
		counts[dir+"/"+name] = n
	}
	return counts, rows.Err()
}

// dbGetIndexDeps returns all knowledge note paths listed by the given index note.
func dbGetIndexDeps(db *sql.DB, index Path) ([]Path, error) {
	rows, err := db.Query(`
		SELECT knowledge_dir, knowledge_name FROM index_deps
		WHERE index_dir = ? AND index_name = ?
		ORDER BY knowledge_dir ASC, knowledge_name ASC`,
		index.Dir(), index.Base())
	if err != nil {
		return nil, fmt.Errorf("storage: index_deps get: %w", err)
	}
	defer rows.Close()
	return dbScanPaths(rows)
}

// dbGetKnowledgeIndexes returns all index note paths that list the given knowledge note.
func dbGetKnowledgeIndexes(db *sql.DB, knowledge Path) ([]Path, error) {
	rows, err := db.Query(`
		SELECT index_dir, index_name FROM index_deps
		WHERE knowledge_dir = ? AND knowledge_name = ?
		ORDER BY index_dir ASC, index_name ASC`,
		knowledge.Dir(), knowledge.Base())
	if err != nil {
		return nil, fmt.Errorf("storage: index_deps reverse get: %w", err)
	}
	defer rows.Close()
	return dbScanPaths(rows)
}

// dbScanPaths reads (dir, name) column pairs from rows into a []Path slice.
func dbScanPaths(rows *sql.Rows) ([]Path, error) {
	var paths []Path
	for rows.Next() {
		var dir, name string
		if err := rows.Scan(&dir, &name); err != nil {
			return nil, fmt.Errorf("storage: scan path: %w", err)
		}
		if p, ok := ParsePath(dir + "/" + name); ok {
			paths = append(paths, p)
		}
	}
	return paths, rows.Err()
}

// ── knowledge_links DB helpers ────────────────────────────────────────────────

// dbReplaceKnowledgeLinks replaces, within a single transaction, all outgoing
// wikilink edges for the given source knowledge note.
// entityType is the source note's frontmatter entity_type (may be empty).
// Passing an empty targets slice clears all existing edges from source.
func dbReplaceKnowledgeLinks(db *sql.DB, source Path, entityType string, targets []Path) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("storage: knowledge_links replace: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(
		`DELETE FROM knowledge_links WHERE source_dir = ? AND source_name = ?`,
		source.Dir(), source.Base(),
	); err != nil {
		return fmt.Errorf("storage: knowledge_links replace: delete: %w", err)
	}

	now := time.Now().UnixNano()
	if len(targets) == 0 && entityType != "" {
		// Leaf node: no outgoing links but has an entity_type. Insert a self-loop
		// so the entity_type survives in the DB; graph API filters self-loops from edges.
		if _, err := tx.Exec(`
			INSERT INTO knowledge_links(source_dir, source_name, target_dir, target_name, source_entity_type, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(source_dir, source_name, target_dir, target_name) DO NOTHING`,
			source.Dir(), source.Base(), source.Dir(), source.Base(), entityType, now,
		); err != nil {
			return fmt.Errorf("storage: knowledge_links replace: insert self-loop: %w", err)
		}
	}
	for _, t := range targets {
		if _, err := tx.Exec(`
			INSERT INTO knowledge_links(source_dir, source_name, target_dir, target_name, source_entity_type, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(source_dir, source_name, target_dir, target_name) DO NOTHING`,
			source.Dir(), source.Base(), t.Dir(), t.Base(), entityType, now,
		); err != nil {
			return fmt.Errorf("storage: knowledge_links replace: insert (%s/%s): %w", t.Dir(), t.Base(), err)
		}
	}
	return tx.Commit()
}

// KnowledgeEdge is a directed edge between two knowledge notes.
type KnowledgeEdge struct {
	Source           Path
	Target           Path
	SourceEntityType string
}

// dbGetAllKnowledgeLinks returns every directed wikilink edge between knowledge notes.
func dbGetAllKnowledgeLinks(db *sql.DB) ([]KnowledgeEdge, error) {
	rows, err := db.Query(`
		SELECT source_dir, source_name, target_dir, target_name, source_entity_type
		FROM knowledge_links
		ORDER BY source_dir, source_name, target_dir, target_name`)
	if err != nil {
		return nil, fmt.Errorf("storage: knowledge_links get all: %w", err)
	}
	defer rows.Close()
	return dbScanEdges(rows)
}

// dbGetKnowledgeLinksForIndex returns all wikilink edges where both source and
// target are listed under the given index note in index_deps.
func dbGetKnowledgeLinksForIndex(db *sql.DB, index Path) ([]KnowledgeEdge, error) {
	rows, err := db.Query(`
		SELECT kl.source_dir, kl.source_name, kl.target_dir, kl.target_name, kl.source_entity_type
		FROM knowledge_links kl
		JOIN index_deps id1
		  ON id1.index_dir = ? AND id1.index_name = ?
		 AND id1.knowledge_dir = kl.source_dir AND id1.knowledge_name = kl.source_name
		JOIN index_deps id2
		  ON id2.index_dir = ? AND id2.index_name = ?
		 AND id2.knowledge_dir = kl.target_dir AND id2.knowledge_name = kl.target_name
		ORDER BY kl.source_dir, kl.source_name, kl.target_dir, kl.target_name`,
		index.Dir(), index.Base(), index.Dir(), index.Base())
	if err != nil {
		return nil, fmt.Errorf("storage: knowledge_links for index: %w", err)
	}
	defer rows.Close()
	return dbScanEdges(rows)
}

// dbScanEdges reads edge rows (source_dir, source_name, target_dir, target_name, source_entity_type)
// into a []KnowledgeEdge slice.
func dbScanEdges(rows *sql.Rows) ([]KnowledgeEdge, error) {
	var edges []KnowledgeEdge
	for rows.Next() {
		var sDir, sName, tDir, tName, entityType string
		if err := rows.Scan(&sDir, &sName, &tDir, &tName, &entityType); err != nil {
			return nil, fmt.Errorf("storage: scan edge: %w", err)
		}
		sp, sok := ParsePath(sDir + "/" + sName)
		tp, tok := ParsePath(tDir + "/" + tName)
		if sok && tok {
			edges = append(edges, KnowledgeEdge{Source: sp, Target: tp, SourceEntityType: entityType})
		}
	}
	return edges, rows.Err()
}

// dbImageScan reads all rows from an images query result into an []Image slice.
func dbImageScan(rows *sql.Rows) ([]Image, error) {
	var imgs []Image
	for rows.Next() {
		var img Image
		var createdNs, updNs int64
		var linkedNotes string
		if err := rows.Scan(&img.Dir, &img.Name, &img.Ext, &img.Size, &createdNs, &updNs, &linkedNotes); err != nil {
			return nil, fmt.Errorf("storage: scan image: %w", err)
		}
		img.CreatedAt = time.Unix(0, createdNs)
		img.UpdatedAt = time.Unix(0, updNs)
		if linkedNotes != "" {
			img.LinkedNotes = strings.Split(linkedNotes, "\n")
		}
		imgs = append(imgs, img)
	}
	return imgs, rows.Err()
}

// joinTags serialises a []string tag slice to a newline-separated string for DB storage.
func joinTags(tags []string) string { return strings.Join(tags, "\n") }

// splitTags deserialises a newline-separated tag string from DB into a []string.
// Returns nil for empty input so that json.Marshal omits the field.
func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
