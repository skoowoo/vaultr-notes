package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hardhacker/vaultr/internal/storage"
)

type listRequest struct {
	// Path is the vault-relative directory path (must start with "/" if non-empty).
	// Omit or set to "/" for the vault root.
	Path string `json:"path"`
	All  bool   `json:"all"`
	Sort string `json:"sort"`
	Limit  int    `json:"limit"`
	// Start and End are RFC 3339 date strings for updated_at filtering (inclusive start, exclusive end).
	Start string `json:"start"`
	End   string `json:"end"`
	// Latest limits results to notes updated within the last N days.
	Latest int `json:"latest"`
	// Origins, when non-empty, filters results to notes whose origin is in the set (e.g. ["api"], ["plugin:compile","plugin:index"]).
	Origins []string `json:"origins"`
	// ExcludeOrigins excludes notes whose origin is in the set. Mutually exclusive with Origins.
	ExcludeOrigins []string `json:"exclude_origins"`
}

// List handles POST /api/vault/list.
//
// Body fields:
//
//	path            — vault-relative directory path; omit or "/" for root (must start with "/" if set)
//	all             — true: return every note in the vault (path ignored)
//	sort            — "time" to sort by updated_at DESC
//	limit           — maximum number of notes to return
//	origins         — if set, only notes whose origin is in the list (e.g. ["plugin:compile","plugin:index"])
//	exclude_origins — if set, exclude notes whose origin is in the list; mutually exclusive with origins
func (gh *VaultHandler) List(w http.ResponseWriter, r *http.Request) {
	var req listRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Origins) > 0 && len(req.ExcludeOrigins) > 0 {
		http.Error(w, `cannot set both "origins" and "exclude_origins"`, http.StatusBadRequest)
		return
	}
	onlyOrigins := make([]storage.Origin, len(req.Origins))
	for i, o := range req.Origins {
		onlyOrigins[i] = storage.Origin(o)
	}
	excludeOrigins := make([]storage.Origin, len(req.ExcludeOrigins))
	for i, o := range req.ExcludeOrigins {
		excludeOrigins[i] = storage.Origin(o)
	}
	opts := storage.ListOptions{
		SortByTime:     req.Sort == "time",
		Limit:          req.Limit,
		OnlyOrigins:    onlyOrigins,
		ExcludeOrigins: excludeOrigins,
	}
	if req.Latest > 0 {
		opts.After = time.Now().AddDate(0, 0, -req.Latest)
	}
	if req.Start != "" {
		t, err := time.Parse(time.DateOnly, req.Start)
		if err != nil {
			http.Error(w, "invalid start date (use YYYY-MM-DD): "+err.Error(), http.StatusBadRequest)
			return
		}
		opts.After = t
	}
	if req.End != "" {
		t, err := time.Parse(time.DateOnly, req.End)
		if err != nil {
			http.Error(w, "invalid end date (use YYYY-MM-DD): "+err.Error(), http.StatusBadRequest)
			return
		}
		opts.Before = t
	}

	if req.All {
		gh.listAll(w, opts)
		return
	}

	// Resolve and validate path.
	dirStr := req.Path
	if dirStr == "" {
		dirStr = "/"
	}
	p, ok := storage.ParsePath(dirStr)
	if !ok {
		http.Error(w, `path must be absolute (start with "/")`, http.StatusBadRequest)
		return
	}

	gh.listDir(w, p, opts)
}

func (gh *VaultHandler) listAll(w http.ResponseWriter, opts storage.ListOptions) {
	notes, err := gh.vault.ListAllNotes(opts)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"path":  "/",
		"notes": notes,
		"all":   true,
	})
}

func (gh *VaultHandler) listDir(w http.ResponseWriter, p storage.Path, opts storage.ListOptions) {
	dirMeta, err := gh.vault.StatDir(p)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	notes, err := gh.vault.ListDir(p, opts)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"path":  dirMeta.Path,
		"notes": notes,
	})
}
