// Package inbox stores a generic, source-agnostic stream of messages surfaced
// to the user (e.g. "your background agent run finished"). It knows nothing
// about mates, agents, or any other producer — callers translate their own
// domain events into a Message and hand it to Store.Create. Producer-specific
// data belongs in Metadata, not in new columns.
package inbox

import "time"

// Level is the visual/severity classification of a message.
type Level string

const (
	LevelInfo    Level = "info"
	LevelSuccess Level = "success"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
)

// Message is one entry in the inbox.
type Message struct {
	ID        string         `json:"id"`
	Source    string         `json:"source"` // producer identifier, e.g. "mate_trigger"
	Level     Level          `json:"level"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Link      string         `json:"link,omitempty"`     // optional deep link the client can navigate to
	Metadata  map[string]any `json:"metadata,omitempty"` // opaque producer-specific payload
	IsRead    bool           `json:"isRead"`
	ReadAt    *time.Time     `json:"readAt,omitempty"` // nil until IsRead is set
	CreatedAt time.Time      `json:"createdAt"`
}

// ListFilter narrows ListMessages results.
type ListFilter struct {
	Source     string // empty = all sources
	UnreadOnly bool
	Limit      int // 0 = default page size
	Offset     int
}
