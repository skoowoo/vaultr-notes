package inbox

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS inbox_messages (
    id          TEXT    PRIMARY KEY,
    source      TEXT    NOT NULL DEFAULT '',
    level       TEXT    NOT NULL DEFAULT 'info',
    title       TEXT    NOT NULL DEFAULT '',
    body        TEXT    NOT NULL DEFAULT '',
    link        TEXT    NOT NULL DEFAULT '',
    metadata    TEXT    NOT NULL DEFAULT '{}',
    is_read     INTEGER NOT NULL DEFAULT 0,
    read_at     INTEGER NOT NULL DEFAULT 0,
    created_at  INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS inbox_messages_created_idx ON inbox_messages(created_at DESC);
CREATE INDEX IF NOT EXISTS inbox_messages_unread_idx  ON inbox_messages(is_read, created_at DESC);
`

const defaultListLimit = 50

// retentionWindow bounds how long a message stays in the inbox — anything
// older is pruned so the table (and every List/UnreadCount scan) stays
// small indefinitely.
const retentionWindow = 30 * 24 * time.Hour

// Store is the data access point for the generic inbox. It shares the
// caller-provided *sql.DB connection (the mate package's mate.db) rather
// than owning its own file — Open only adds its own table to that
// connection's schema. The caller owns the connection's lifecycle.
type Store struct {
	db  *sql.DB
	bus *Bus
}

// Open adds the inbox_messages table (if missing) to db and returns a Store
// backed by that shared connection.
func Open(db *sql.DB) (*Store, error) {
	if _, err := db.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("inbox: schema: %w", err)
	}
	return &Store{db: db, bus: newBus()}, nil
}

// Notifications returns the bus that fires whenever a new message is
// created, for wiring GET /api/inbox/notifications.
func (s *Store) Notifications() *Bus { return s.bus }

// Create inserts a new inbox message. ID and CreatedAt are assigned if unset.
func (s *Store) Create(m Message) (Message, error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.Level == "" {
		m.Level = LevelInfo
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	metaJSON, err := marshalMetadata(m.Metadata)
	if err != nil {
		return Message{}, fmt.Errorf("inbox: marshal metadata: %w", err)
	}
	_, err = s.db.Exec(
		`INSERT INTO inbox_messages(id, source, level, title, body, link, metadata, is_read, read_at, created_at)
		 VALUES (?,?,?,?,?,?,?,0,0,?)`,
		m.ID, m.Source, string(m.Level), m.Title, m.Body, m.Link, metaJSON, m.CreatedAt.UnixMilli(),
	)
	if err != nil {
		return Message{}, fmt.Errorf("inbox: create message: %w", err)
	}
	s.bus.push(m)
	// Fire-and-forget: prune runs off the request goroutine so it never adds
	// latency to the caller (a background producer today) or, more
	// importantly, to GET /api/inbox reads, which never invoke this at all.
	go s.pruneOld()
	return m, nil
}

// pruneOld deletes messages older than retentionWindow. Best-effort — a
// failed prune just means the next Create tries again.
func (s *Store) pruneOld() {
	cutoff := time.Now().Add(-retentionWindow).UnixMilli()
	if _, err := s.db.Exec(`DELETE FROM inbox_messages WHERE created_at < ?`, cutoff); err != nil {
		slog.Warn("inbox: prune old messages", "err", err)
	}
}

// List returns messages newest-first, narrowed by f.
func (s *Store) List(f ListFilter) ([]Message, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	query := `SELECT id, source, level, title, body, link, metadata, is_read, read_at, created_at
	          FROM inbox_messages WHERE 1=1`
	var args []any
	if f.Source != "" {
		query += ` AND source = ?`
		args = append(args, f.Source)
	}
	if f.UnreadOnly {
		query += ` AND is_read = 0`
	}
	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, f.Offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("inbox: list: %w", err)
	}
	defer rows.Close()
	return scanMessages(rows)
}

// UnreadCount returns the number of unread messages, optionally narrowed to one source.
func (s *Store) UnreadCount(source string) (int, error) {
	query := `SELECT COUNT(*) FROM inbox_messages WHERE is_read = 0`
	var args []any
	if source != "" {
		query += ` AND source = ?`
		args = append(args, source)
	}
	var n int
	err := s.db.QueryRow(query, args...).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("inbox: unread count: %w", err)
	}
	return n, nil
}

// MarkRead marks a single message as read.
func (s *Store) MarkRead(id string) error {
	_, err := s.db.Exec(
		`UPDATE inbox_messages SET is_read = 1, read_at = ? WHERE id = ? AND is_read = 0`,
		time.Now().UnixMilli(), id,
	)
	if err != nil {
		return fmt.Errorf("inbox: mark read: %w", err)
	}
	return nil
}

// MarkAllRead marks every unread message as read, optionally narrowed to one source.
func (s *Store) MarkAllRead(source string) error {
	query := `UPDATE inbox_messages SET is_read = 1, read_at = ? WHERE is_read = 0`
	args := []any{time.Now().UnixMilli()}
	if source != "" {
		query += ` AND source = ?`
		args = append(args, source)
	}
	if _, err := s.db.Exec(query, args...); err != nil {
		return fmt.Errorf("inbox: mark all read: %w", err)
	}
	return nil
}

// Delete removes a single message.
func (s *Store) Delete(id string) error {
	if _, err := s.db.Exec(`DELETE FROM inbox_messages WHERE id = ?`, id); err != nil {
		return fmt.Errorf("inbox: delete: %w", err)
	}
	return nil
}

func marshalMetadata(m map[string]any) (string, error) {
	if len(m) == 0 {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func scanMessages(rows *sql.Rows) ([]Message, error) {
	var out []Message
	for rows.Next() {
		var m Message
		var level, metaJSON string
		var isRead int
		var readAtMs, createdAtMs int64
		if err := rows.Scan(&m.ID, &m.Source, &level, &m.Title, &m.Body, &m.Link, &metaJSON, &isRead, &readAtMs, &createdAtMs); err != nil {
			return nil, fmt.Errorf("inbox: scan message: %w", err)
		}
		m.Level = Level(level)
		m.IsRead = isRead == 1
		if readAtMs > 0 {
			t := time.UnixMilli(readAtMs)
			m.ReadAt = &t
		}
		m.CreatedAt = time.UnixMilli(createdAtMs)
		if metaJSON != "" && metaJSON != "{}" {
			var meta map[string]any
			if err := json.Unmarshal([]byte(metaJSON), &meta); err == nil {
				m.Metadata = meta
			}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
