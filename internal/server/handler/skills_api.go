package handler

import (
	"net/http"
	"strings"

	"github.com/hardhacker/vaultr/internal/skills"
)

// SkillsHTTP handles /api/skills endpoints.
type SkillsHTTP struct {
	mgr *skills.Manager
}

// NewSkillsHTTP constructs the handler.
func NewSkillsHTTP(mgr *skills.Manager) *SkillsHTTP {
	return &SkillsHTTP{mgr: mgr}
}

// List handles GET /api/skills.
func (h *SkillsHTTP) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.mgr.List()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"skills": list})
}

// Enable handles POST /api/skills/{name}/enable.
func (h *SkillsHTTP) Enable(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !validSkillName(name) {
		respondJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid skill name"})
		return
	}
	if err := h.mgr.Enable(name); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"ok": true, "name": name, "enabled": true})
}

// Disable handles POST /api/skills/{name}/disable.
func (h *SkillsHTTP) Disable(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !validSkillName(name) {
		respondJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid skill name"})
		return
	}
	if err := h.mgr.Disable(name); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"ok": true, "name": name, "enabled": false})
}

func validSkillName(name string) bool {
	return name != "" &&
		!strings.Contains(name, "/") &&
		!strings.Contains(name, "\\") &&
		!strings.Contains(name, "..") &&
		!strings.HasPrefix(name, ".")
}
