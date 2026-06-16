package gitsync

import (
	"encoding/json"
	"net/http"
)

// Syncer is satisfied by the gitsync.Plugin (and any future plugin that
// supports on-demand sync).  Defined here as a local interface to keep the
// handler layer decoupled from the concrete plugin type.
type Syncer interface {
	TriggerSync()
}

// Handler handles requests to manually trigger a git sync cycle.
type Handler struct {
	syncer Syncer
}

// NewHandler returns a handler backed by the given Syncer.
func NewHandler(s Syncer) *Handler {
	return &Handler{syncer: s}
}

// Sync handles POST /api/git/sync.
// It enqueues an immediate pull+push and returns 202 Accepted.
func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	h.syncer.TriggerSync()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "sync requested"}) //nolint:errcheck
}
