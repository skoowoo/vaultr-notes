package handler

import (
	"net/http"
)

// StatusResponse is the JSON body returned by POST /api/status.
type StatusResponse struct {
	Notes   int    `json:"notes"`
	Indexed uint64 `json:"indexed"`
}

// NoteCounter is satisfied by storage.Vault.
type NoteCounter interface {
	CountNotes() (int, error)
}

// IndexCounter is satisfied by search.Plugin.
type IndexCounter interface {
	DocCount() (uint64, error)
}

// StatusHandler handles POST /api/status.
type StatusHandler struct {
	notes   NoteCounter
	indexed IndexCounter
}

// NewStatus returns a StatusHandler.
func NewStatus(notes NoteCounter, indexed IndexCounter) *StatusHandler {
	return &StatusHandler{notes: notes, indexed: indexed}
}

// Status handles POST /api/status.
func (h *StatusHandler) Status(w http.ResponseWriter, r *http.Request) {
	notes, err := h.notes.CountNotes()
	if err != nil {
		http.Error(w, "status: count notes: "+err.Error(), http.StatusInternalServerError)
		return
	}
	indexed, err := h.indexed.DocCount()
	if err != nil {
		http.Error(w, "status: count indexed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, StatusResponse{Notes: notes, Indexed: indexed})
}
