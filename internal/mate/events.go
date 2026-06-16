package mate

import (
	"time"

	"github.com/hardhacker/vaultr/internal/plugin"
)

// MateEventType is a high-level semantic event exposed to trigger matching.
type MateEventType string

const (
	MateEventNoteCreated      MateEventType = "note_created"
	MateEventNoteUpdated      MateEventType = "note_updated"
	MateEventNoteDeleted      MateEventType = "note_deleted"
	MateEventShortNoteCreated MateEventType = "short_note_created"
	MateEventScheduled        MateEventType = "scheduled"
	MateEventWechatMessage    MateEventType = "wechat_message"
	MateEventCompileRequested MateEventType = "compile_requested"
)

// EventDef describes a built-in mate event for display in the UI.
type EventDef struct {
	Type        MateEventType `json:"type"`
	Label       string        `json:"label"`
	Description string        `json:"description"`
}

// BuiltinEvents is the authoritative list of mate events the system supports.
var BuiltinEvents = []EventDef{
	{MateEventNoteCreated, "Note Created", "Fires when any note is created in the vault"},
	{MateEventNoteUpdated, "Note Updated", "Fires when any existing note is modified"},
	{MateEventNoteDeleted, "Note Deleted", "Fires when any note is deleted"},
	{MateEventShortNoteCreated, "Short Note Created", "Fires each time a short note entry is appended"},
	{MateEventScheduled, "Scheduled", "Fires on a configured interval or daily time"},
	{MateEventWechatMessage, "WeChat Message", "Fires when a WeChat direct message is received"},
	{MateEventCompileRequested, "Compile Requested", "Fires when the user manually triggers knowledge compilation for a note — path carries the note to compile"},
}

// EventLabel returns the display label for a mate event type.
func EventLabel(t MateEventType) string {
	for _, e := range BuiltinEvents {
		if e.Type == t {
			return e.Label
		}
	}
	return string(t)
}

// MateEvent is the enriched semantic event passed to trigger matching and prompt rendering.
type MateEvent struct {
	Type         MateEventType
	Path         string
	Content      string    // populated for short_note_created and wechat_message
	FiredAt      time.Time // populated for scheduled: when the trigger fired
	WechatUserID string    // populated for wechat_message

	// Reply is propagated from plugin.Event; invoked after the trigger agent run finishes.
	Reply plugin.ReplyFunc
}

// Translate converts a low-level plugin.Event to the matching set of MateEvents.
func Translate(e plugin.Event) []MateEvent {
	if e.IsDir {
		return nil
	}
	var out []MateEvent
	switch e.Type {
	case plugin.EventVaultCreate:
		out = append(out, MateEvent{Type: MateEventNoteCreated, Path: e.Path})
	case plugin.EventVaultWrite:
		out = append(out, MateEvent{Type: MateEventNoteUpdated, Path: e.Path})
	case plugin.EventVaultDelete:
		out = append(out, MateEvent{Type: MateEventNoteDeleted, Path: e.Path})
	case plugin.EventVaultShortAppend:
		out = append(out, MateEvent{Type: MateEventShortNoteCreated, Path: e.Path, Content: e.Content})
	case plugin.EventWechatMessage:
		out = append(out, MateEvent{
			Type:         MateEventWechatMessage,
			Path:         e.Path,
			Content:      e.Content,
			WechatUserID: e.WechatUserID,
			Reply:        e.Reply,
		})
	case plugin.EventCompileRequested:
		out = append(out, MateEvent{
			Type:  MateEventCompileRequested,
			Path:  e.Path,
			Reply: e.Reply,
		})
	}
	return out
}
