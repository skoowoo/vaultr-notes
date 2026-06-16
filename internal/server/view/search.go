package view

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/hardhacker/vaultr/internal/plugins/search"
	"github.com/hardhacker/vaultr/internal/storage"
)

// searchResultItem is passed to the search fragment template.
type searchResultItem struct {
	Name        string
	Dir         string
	DirLabel    string // first path segment of Dir for display (empty for vault root)
	UpdatedAt   string
	URL         string
	FocusURL    string
	PreviewPath string
	IsKnowledge bool
	IsIndex     bool
	CanCompile  bool
}

var searchFragTmpl = template.Must(template.New("searchfrag").Parse(`
{{- if .}}
{{- range .}}
<a href="{{.URL}}" data-focus-url="{{.FocusURL}}" data-preview-path="{{.PreviewPath}}" data-note-is-knowledge="{{.IsKnowledge}}" data-note-is-index="{{.IsIndex}}" data-note-can-compile="{{.CanCompile}}" class="flex shrink-0 items-center gap-3 px-3.5 py-2.5 rounded-md cursor-pointer overflow-hidden">
  <svg class="sr-icon w-[14px] h-[14px] flex-shrink-0" fill="none" stroke="currentColor" stroke-width="1.75" viewBox="0 0 24 24">
    <path stroke-linecap="round" stroke-linejoin="round" d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/>
    <path stroke-linecap="round" stroke-linejoin="round" d="M14 2v4a1 1 0 0 0 1 1h4"/>
    <path stroke-linecap="round" d="M10 9H8"/>
    <path stroke-linecap="round" d="M16 13H8"/>
    <path stroke-linecap="round" d="M16 17H8"/>
  </svg>
  <div class="flex-1 min-w-0">
    <div class="sr-name truncate">{{.Name}}</div>
    {{- if or .DirLabel .UpdatedAt}}
    <div class="sr-meta mt-[2px] grid grid-cols-[minmax(0,1fr)_auto] gap-x-2 items-baseline min-w-0">
      <span class="sr-dir truncate min-w-0">{{.DirLabel}}</span>
      {{- if .UpdatedAt}}<span class="sr-time">{{.UpdatedAt}}</span>{{end}}
    </div>
    {{- end}}
  </div>
</a>
{{- end}}
{{- else}}
<div class="sr-empty px-4 py-8 text-center">No results found</div>
{{- end}}
`))

// SearchFragment handles GET /notes/search?q=…&field=… and returns an HTML
// fragment with search results for HTMX to swap into the dropdown list.
// field may be "name", "content", "tag", or "" (search all, default).
func (vh *ViewHandler) SearchFragment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/home") {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" || vh.searcher == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		return
	}

	field := r.URL.Query().Get("field")
	switch field {
	case "", "name", "content", "tag":
	default:
		field = ""
	}

	kind := r.URL.Query().Get("kind")
	switch kind {
	case "", "raw", "knowledge", "short", "index":
	default:
		kind = ""
	}
	// kind-filter modes search name+content, not a single field
	if kind != "" {
		field = ""
	}

	results, err := vh.searcher.Search(q, search.SearchOptions{
		Limit: 20,
		Type:  field,
		Kind:  kind,
	})
	if err != nil {
		http.Error(w, "search: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// In mixed/name mode, surface filename matches above pure content hits.
	// In tag/content mode bleve scoring is already meaningful; skip the re-sort.
	if field == "" || field == "name" {
		sort.SliceStable(results, func(i, j int) bool {
			return results[i].NameMatch && !results[j].NameMatch
		})
	}

	// Batch-fetch compile_count for all results so we can compute CanCompile accurately.
	// Only raw/short notes need checking; knowledge and index are always non-compilable.
	paths := make([]storage.Path, 0, len(results))
	for _, res := range results {
		if res.Kind != "knowledge" && res.Kind != "index" {
			if p, ok := storage.ParsePath(storage.JoinPath(res.Dir, res.Name)); ok {
				paths = append(paths, p)
			}
		}
	}
	compileCountByPath := make(map[string]int, len(paths))
	if len(paths) > 0 {
		if notes, err2 := vh.vault.GetNotesByPaths(paths); err2 == nil {
			for _, n := range notes {
				compileCountByPath[n.PathString()] = n.CompileCount
			}
		}
	}

	items := make([]searchResultItem, 0, len(results))
	for _, res := range results {
		name := strings.TrimSuffix(res.Name, ".md")
		updatedAt := ""
		if !res.UpdatedAt.IsZero() {
			updatedAt = formatRelativeTime(res.UpdatedAt)
		}
		notePath := storage.JoinPath(res.Dir, res.Name)
		q := url.Values{}
		q.Set("name", res.Name)
		focusQuery := url.Values{}
		focusQuery.Set("path", notePath)
		isKnowledge := res.Kind == "knowledge"
		isIndex := res.Kind == "index"
		canCompile := !isKnowledge && !isIndex && compileCountByPath[notePath] == 0
		items = append(items, searchResultItem{
			Name:        name,
			Dir:         res.Dir,
			DirLabel:    firstDirSegment(res.Dir),
			UpdatedAt:   updatedAt,
			URL:         "/notes?" + q.Encode(),
			FocusURL:    "/library/focus?" + focusQuery.Encode(),
			PreviewPath: notePath,
			IsKnowledge: isKnowledge,
			IsIndex:     isIndex,
			CanCompile:  canCompile,
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := searchFragTmpl.Execute(w, items); err != nil {
		http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
	}
}

// firstDirSegment returns the top-level directory name for a vault-abs dir path.
// "/" (vault root) is shown as "/".
func firstDirSegment(dir string) string {
	dir = strings.TrimSpace(dir)
	if dir == "/" {
		return "/"
	}
	if dir == "" {
		return ""
	}
	dir = strings.Trim(dir, "/")
	if dir == "" {
		return ""
	}
	if i := strings.Index(dir, "/"); i >= 0 {
		return dir[:i]
	}
	return dir
}

func formatRelativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}
