package view

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hardhacker/vaultr/internal/plugins/search"
	"github.com/hardhacker/vaultr/internal/storage"
)

const libraryPageSize = 20

// ─── data types ───────────────────────────────────────────────────────────────

type tagItem struct {
	Name  string
	Count uint64
}

type noteItem struct {
	Name            string
	Title           string // LLM-generated title; non-empty for distill notes
	Dir             string
	Path            string
	UpdatedAt       string
	URL             string // full view page URL
	FragmentURL     string // drawer fragment URL
	FocusURL        string // library focus URL — filters related cards via OOB
	CursorNs        int64  // Unix nanoseconds for pagination cursor
	IsKnowledge     bool   // true for knowledge notes; drives data-knowledge on the read button
	IsShort         bool   // true for short notes (origin = "short")
	IsIndex         bool   // true for index notes (origin = "plugin:index")
	Pinned          bool   // true when the note is pinned by the user
	CanCompile      bool   // true when this note is eligible to be compiled (raw, compile_count==0)
	IsCompiled      bool   // true when raw note has been compiled at least once (compile_count>0, not knowledge/index)
	LinkedPathsJSON string   // JSON array of linked raw paths; set on knowledge cards in tag view
	Tags            []string // frontmatter tags; non-nil on knowledge/index cards that have tags
	DepCount        int      // number of knowledge deps; set on index cards
}

type libraryPageData struct {
	Tags           []tagItem
	IndexNotes     []noteItem
	KnowledgeNotes []noteItem
	RawNotes       []noteItem
	NotesCount     int
	IndexCount     int
	TagCount       int
	KnowledgeCount int
	RawCount       int
	KnowledgeNext  int64 // Unix ns cursor, 0 = no more pages
	RawNext        int64
}

type notesFragData struct {
	Notes       []noteItem
	NextNs      int64
	Type        string
	IsKnowledge bool // drives hx-get in the template (FocusURL vs FragmentURL)
}

type tagKnowledgeData struct {
	KnowledgeNotes []noteItem
	RawNotes       []noteItem
}

type focusOOBData struct {
	Focused             noteItem
	FocusedIsKnowledge  bool
	FocusedCanKnowledge bool
	LinkedKnowledge     []noteItem
	LinkedRaw           []noteItem
	RawOnly             bool // true → skip knowledge-col-body OOB (K card selected in-place)
}

// ─── handlers ─────────────────────────────────────────────────────────────────

// Library handles GET /library — the full overview page.
func (vh *ViewHandler) Library(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tags []tagItem
	if vh.searcher != nil {
		counts, err := vh.searcher.TagDistribution(300)
		if err == nil {
			for _, tc := range counts {
				tags = append(tags, tagItem{Name: tc.Tag, Count: tc.Count})
			}
		}
	}

	indexNotes := vh.listIndexItems()
	knowledgeNotes, knowledgeNext := vh.listNoteItems("knowledge", 0, libraryPageSize)
	rawNotes, rawNext := vh.listNoteItems("raw", 0, libraryPageSize)
	notesCount, err := vh.vault.CountNotes()
	if err != nil {
		http.Error(w, "status: count notes: "+err.Error(), http.StatusInternalServerError)
		return
	}
	knowledgeCount, _ := vh.vault.CountKnowledgeNotes()
	rawCount, _ := vh.vault.CountRawNotes()

	data := libraryPageData{
		Tags:           tags,
		IndexNotes:     indexNotes,
		KnowledgeNotes: knowledgeNotes,
		RawNotes:       rawNotes,
		NotesCount:     notesCount,
		IndexCount:     len(indexNotes),
		TagCount:       len(tags),
		KnowledgeCount: knowledgeCount,
		RawCount:       rawCount,
		KnowledgeNext:  knowledgeNext,
		RawNext:        rawNext,
	}

	var buf bytes.Buffer
	if err := libraryPageTemplate.Execute(&buf, data); err != nil {
		http.Error(w, fmt.Sprintf("render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

// LibraryNotes handles GET /library/notes?type=knowledge|raw&before=UNIX_NS
// and returns an HTMX fragment for scroll-to-load pagination.
func (vh *ViewHandler) LibraryNotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/library") {
		return
	}

	noteType := r.URL.Query().Get("type")
	if noteType != "knowledge" {
		noteType = "raw"
	}
	var beforeNs int64
	if s := r.URL.Query().Get("before"); s != "" {
		beforeNs, _ = strconv.ParseInt(s, 10, 64)
	}

	notes, nextNs := vh.listNoteItems(noteType, beforeNs, libraryPageSize)
	data := notesFragData{Notes: notes, NextNs: nextNs, Type: noteType, IsKnowledge: noteType == "knowledge"}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := notesFragTemplate.Execute(w, data); err != nil {
		http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
	}
}

// LibraryTag handles GET /library/tag?tag=NAME
// Returns an HTMX fragment that replaces the distill column with notes tagged
// NAME, plus an OOB swap that clears the raw column.
func (vh *ViewHandler) LibraryTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/library") {
		return
	}
	tag := strings.TrimSpace(r.URL.Query().Get("tag"))
	if tag == "" || vh.searcher == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tagKnowledgeTemplate.Execute(w, tagKnowledgeData{}) //nolint:errcheck
		return
	}

	results, err := vh.searcher.Search(tag, search.SearchOptions{Type: "tag", Limit: 200})
	var knowledgeItems []noteItem
	var knowledgeNames []string
	if err == nil {
		for _, sr := range results {
			if sr.Kind != "knowledge" && sr.Kind != "index" {
				continue
			}
			knowledgeItems = append(knowledgeItems, searchResultToItem(sr))
			knowledgeNames = append(knowledgeNames, sr.Name)
		}
	}

	// Batch-fetch distill metadata to resolve titles and linked raw notes.
	var rawItems []noteItem
	if len(knowledgeNames) > 0 {
		knowledgeMeta, _ := vh.vault.GetNotesByNames(knowledgeNames)

		titleByName := make(map[string]string, len(knowledgeMeta))
		for _, n := range knowledgeMeta {
			if n.Title != "" {
				titleByName[n.Name] = n.Title
			}
		}

		// Back-fill titles into knowledgeItems (search results don't carry title).
		for i := range knowledgeItems {
			if t := titleByName[knowledgeItems[i].Name+".md"]; t != "" {
				knowledgeItems[i].Title = t
			}
		}

		// Collect raw notes and build knowledge→raw adjacency for connection drawing.
		seen := make(map[string]bool)
		knowledgeToRaws := make(map[string][]string, len(knowledgeMeta))
		for _, n := range knowledgeMeta {
			p, ok := storage.ParsePath(n.PathString())
			if !ok {
				continue
			}
			sourcePaths, _ := vh.vault.GetKnowledgeDeps(p)
			if len(sourcePaths) == 0 {
				continue
			}
			sourceNotes, _ := vh.vault.GetNotesByPaths(sourcePaths)
			rawPaths := make([]string, 0, len(sourceNotes))
			for _, sn := range sourceNotes {
				ps := sn.PathString()
				rawPaths = append(rawPaths, ps)
				if !seen[ps] {
					seen[ps] = true
					rawItems = append(rawItems, noteToItem(sn))
				}
			}
			knowledgeToRaws[n.PathString()] = rawPaths
		}

		// Embed linked raw paths as JSON so the JS can draw actual connections.
		for i := range knowledgeItems {
			knowledgeItems[i].LinkedPathsJSON = pathsToJSON(knowledgeToRaws[knowledgeItems[i].Path])
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tagKnowledgeTemplate.Execute(w, tagKnowledgeData{
		KnowledgeNotes: knowledgeItems,
		RawNotes:       rawItems,
	}); err != nil {
		http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
	}
}

// LibraryFocus handles GET /library/focus?path=PATH
// Returns only OOB swaps (no drawer content — the drawer is opened separately via
// the peek button which calls /notes/fragment):
//   - #knowledge-col-body: the selected note pinned at top + its linked distill notes
//   - #raw-col-body:     the linked raw notes
func (vh *ViewHandler) LibraryFocus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/library") {
		return
	}
	p, ok := storage.ParsePath(r.URL.Query().Get("path"))
	if !ok {
		http.Error(w, `path must be absolute (start with "/")`, http.StatusBadRequest)
		return
	}
	p, _ = storage.ParsePath(storage.JoinPath(p.Dir(), ensureMarkdownName(p.Base())))

	meta, err := vh.vault.StatNote(p)
	if err != nil {
		writeVaultError(w, err)
		return
	}

	isKnowledge := meta.Kind == storage.KindKnowledge
	var linkedKnowledge, linkedRaw []noteItem

	if isKnowledge {
		// Knowledge note focused: show its source raw notes from knowledge_deps.
		sourcePaths, _ := vh.vault.GetKnowledgeDeps(p)
		if len(sourcePaths) > 0 {
			sourceNotes, _ := vh.vault.GetNotesByPaths(sourcePaths)
			for _, n := range sourceNotes {
				linkedRaw = append(linkedRaw, noteToItem(n))
			}
		}
	} else {
		// Raw note focused: show knowledge notes that compiled from this note.
		knowledgePaths, _ := vh.vault.GetSourceKnowledges(p)
		if len(knowledgePaths) > 0 {
			knowledgeNotes, _ := vh.vault.GetNotesByPaths(knowledgePaths)
			for _, n := range knowledgeNotes {
				linkedKnowledge = append(linkedKnowledge, noteToItem(n))
			}
		}
	}

	rawOnly := r.URL.Query().Get("raw_only") == "true"
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	focused := noteToItem(meta)
	if err := focusOOBTemplate.Execute(w, focusOOBData{
		Focused:             focused,
		FocusedIsKnowledge:  isKnowledge,
		FocusedCanKnowledge: !isKnowledge && len(linkedKnowledge) == 0,
		LinkedKnowledge:     linkedKnowledge,
		LinkedRaw:           linkedRaw,
		RawOnly:             rawOnly,
	}); err != nil {
		http.Error(w, "render oob: "+err.Error(), http.StatusInternalServerError)
	}
}

// LibraryUnfocus handles GET /library/unfocus
// Returns OOB swaps that restore both columns to the default initial listing,
// called when the user deselects a tag (with no active tag).
func (vh *ViewHandler) LibraryUnfocus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/library") {
		return
	}
	knowledgeNotes, knowledgeNext := vh.listNoteItems("knowledge", 0, libraryPageSize)
	rawNotes, rawNext := vh.listNoteItems("raw", 0, libraryPageSize)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := unfocusOOBTemplate.Execute(w, libraryPageData{
		KnowledgeNotes: knowledgeNotes,
		RawNotes:       rawNotes,
		KnowledgeNext:  knowledgeNext,
		RawNext:        rawNext,
	}); err != nil {
		http.Error(w, "render oob: "+err.Error(), http.StatusInternalServerError)
	}
}

// LibraryIndexSelect handles GET /library/index/select?path=PATH
// Returns OOB swaps for both columns filtered by the selected index note's deps.
func (vh *ViewHandler) LibraryIndexSelect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/library") {
		return
	}

	p, ok := storage.ParsePath(r.URL.Query().Get("path"))
	if !ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tagKnowledgeTemplate.Execute(w, tagKnowledgeData{}) //nolint:errcheck
		return
	}
	p, _ = storage.ParsePath(storage.JoinPath(p.Dir(), ensureMarkdownName(p.Base())))

	knowledgePaths, _ := vh.vault.GetIndexDeps(p)
	var knowledgeItems []noteItem
	if len(knowledgePaths) > 0 {
		knowledgeNotes, _ := vh.vault.GetNotesByPaths(knowledgePaths)
		for _, n := range knowledgeNotes {
			knowledgeItems = append(knowledgeItems, noteToItem(n))
		}
	}

	seen := make(map[string]bool)
	knowledgeToRaws := make(map[string][]string, len(knowledgeItems))
	var rawItems []noteItem
	for _, ki := range knowledgeItems {
		kPath, ok := storage.ParsePath(ki.Path)
		if !ok {
			continue
		}
		sourcePaths, _ := vh.vault.GetKnowledgeDeps(kPath)
		if len(sourcePaths) == 0 {
			continue
		}
		sourceNotes, _ := vh.vault.GetNotesByPaths(sourcePaths)
		rawPaths := make([]string, 0, len(sourceNotes))
		for _, sn := range sourceNotes {
			ps := sn.PathString()
			rawPaths = append(rawPaths, ps)
			if !seen[ps] {
				seen[ps] = true
				rawItems = append(rawItems, noteToItem(sn))
			}
		}
		knowledgeToRaws[ki.Path] = rawPaths
	}

	for i := range knowledgeItems {
		knowledgeItems[i].LinkedPathsJSON = pathsToJSON(knowledgeToRaws[knowledgeItems[i].Path])
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tagKnowledgeTemplate.Execute(w, tagKnowledgeData{
		KnowledgeNotes: knowledgeItems,
		RawNotes:       rawItems,
	}); err != nil {
		http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
	}
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// noteToItem converts a storage.Note to a noteItem for template rendering.
func noteToItem(n storage.Note) noteItem {
	isKnowledge := n.Kind == storage.KindKnowledge
	isIndex := n.Kind == storage.KindIndex
	q := url.Values{}
	q.Set("path", n.PathString())
	var tags []string
	if (isKnowledge || isIndex) && len(n.Tags) > 0 {
		tags = n.Tags
	}
	return noteItem{
		Name:        strings.TrimSuffix(n.Name, ".md"),
		Title:       n.Title,
		Dir:         n.Dir,
		Path:        n.PathString(),
		UpdatedAt:   formatRelativeTime(n.UpdatedAt),
		Pinned:      n.Pinned,
		IsKnowledge: isKnowledge,
		IsShort:     n.Kind == storage.KindShort,
		IsIndex:     isIndex,
		CanCompile:  !isKnowledge && !isIndex && n.CompileCount == 0,
		IsCompiled:  !isKnowledge && !isIndex && n.CompileCount > 0,
		URL:         "/notes?" + q.Encode(),
		FragmentURL: "/notes/fragment?" + q.Encode(),
		FocusURL:    "/library/focus?" + q.Encode(),
		CursorNs:    n.UpdatedAt.UnixNano(),
		Tags:        tags,
	}
}

// searchResultToItem converts a search.SearchResult to a noteItem.
func searchResultToItem(sr search.SearchResult) noteItem {
	pathStr := storage.JoinPath(sr.Dir, sr.Name)
	q := url.Values{}
	q.Set("path", pathStr)
	return noteItem{
		Name:        strings.TrimSuffix(sr.Name, ".md"),
		Dir:         sr.Dir,
		Path:        pathStr,
		UpdatedAt:   formatRelativeTime(sr.UpdatedAt),
		URL:         "/notes?" + q.Encode(),
		FragmentURL: "/notes/fragment?" + q.Encode(),
		FocusURL:    "/library/focus?" + q.Encode(),
		CursorNs:    sr.UpdatedAt.UnixNano(),
		IsKnowledge: sr.Kind == "knowledge" || sr.Kind == "index",
		IsIndex:     sr.Kind == "index",
		CanCompile:  false, // knowledge/index items used here are never compilable
	}
}

// pathsToJSON encodes a slice of storage paths as a JSON array string.
func pathsToJSON(paths []string) string {
	if len(paths) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteByte('[')
	for i, p := range paths {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('"')
		for _, c := range p {
			if c == '"' || c == '\\' {
				sb.WriteByte('\\')
			}
			sb.WriteRune(c)
		}
		sb.WriteByte('"')
	}
	sb.WriteByte(']')
	return sb.String()
}

func (vh *ViewHandler) listIndexItems() []noteItem {
	notes, err := vh.vault.ListAllNotes(storage.ListOptions{
		OnlyKinds:  []storage.Kind{storage.KindIndex},
		SortByTime: true,
	})
	if err != nil {
		return nil
	}
	depCounts, _ := vh.vault.GetAllIndexDepCounts()
	items := make([]noteItem, 0, len(notes))
	for _, n := range notes {
		item := noteToItem(n)
		if depCounts != nil {
			item.DepCount = depCounts[n.PathString()]
		}
		items = append(items, item)
	}
	return items
}

func (vh *ViewHandler) listNoteItems(noteType string, beforeNs int64, limit int) ([]noteItem, int64) {
	opts := storage.ListOptions{
		SortByTime: true,
		Limit:      limit + 1,
	}
	if beforeNs > 0 {
		opts.Before = time.Unix(0, beforeNs)
	}
	if noteType == "knowledge" {
		opts.OnlyKinds = []storage.Kind{storage.KindKnowledge}
	} else {
		opts.ExcludeKinds = []storage.Kind{
			storage.KindKnowledge,
			storage.KindIndex,
		}
	}

	notes, err := vh.vault.ListAllNotes(opts)
	if err != nil {
		return nil, 0
	}

	hasMore := len(notes) > limit
	if hasMore {
		notes = notes[:limit]
	}

	items := make([]noteItem, 0, len(notes))
	for _, n := range notes {
		items = append(items, noteToItem(n))
	}

	var nextNs int64
	if hasMore && len(items) > 0 {
		nextNs = items[len(items)-1].CursorNs
	}

	return items, nextNs
}

// ─── templates ────────────────────────────────────────────────────────────────

var notesFragTemplate = template.Must(template.New("notesfrag").Parse(libraryNotesFragHTML))

var tagKnowledgeTemplate = template.Must(template.New("tagknowledge").Parse(libraryTagKnowledgeHTML))

var focusOOBTemplate = template.Must(template.New("focusoob").Parse(libraryFocusOOBHTML))

var unfocusOOBTemplate = template.Must(template.New("unfocusoob").Parse(libraryUnfocusOOBHTML))

var libraryPageHTML = `<!DOCTYPE html>
<html lang="en" data-theme="neo">
` + headHTML(headOpts{title: "Library — Vaultr", withFonts: true, withTW: true, withAlpine: true, withHTMX: true}) + `  <style>
` + appTokensCSS + `
` + navCSS + neoCSS + topbarCSS + libraryCSS + drawerCSS + noteSharedCSS + noteEditorCSS + searchOverlayStyles + confirmDialogCSS + shortDialogCSS + settingsModalCSS + `
  </style>
  <script>
  (function(){
    var d=localStorage.getItem('vaultr-hero-bg');
    if(!d) return;
    var y=parseFloat(localStorage.getItem('vaultr-hero-bg-y'))||0;
    var s=document.createElement('style');
    s.id='hero-bg-preload';
    s.textContent='.home-hero-wrapper{background-image:url('+d+');background-size:100% auto;background-repeat:no-repeat;background-position:center '+y+'px}';
    document.head.appendChild(s);
  })();
  </script>
</head>
<body x-data="libCtrl()">
` + searchOnlyOverlayHTML + confirmDialogHTML + shortDialogHTML + settingsModalHTML() + `

  <header class="lib-topbar">
    <div class="lib-topbar-left">
      <a href="/home" class="lib-back-btn" title="Back to Home">` + topbarIconBack + `<span class="lib-back-label">back</span></a>
    </div>
    <div class="lib-topbar-spacer"></div>
` + topbarActionsHTML("refresh()", "Refresh library", "", "") + `
  </header>

  <div class="lib-body">
` + navHTML("library") + libraryMainHTML + `
  </div>

` + drawerHTML + `

  <script>
  document.addEventListener('alpine:init', () => {
` + alpineStoresScript + `
  });

` + keysJS + drawerScript + searchOverlayScript + confirmDialogJS + shortDialogJS + settingsCtrlJS + libraryJS + `
  </script>
` + noteSharedJS + `
</body>
</html>
`

var libraryPageTemplate = template.Must(template.New("library").Parse(libraryPageHTML))

// LibraryRefresh handles GET /library/refresh — returns HTMX OOB fragments for
// tags, distill, and raw columns without a full page reload.
func (vh *ViewHandler) LibraryRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/library") {
		return
	}

	var tags []tagItem
	if vh.searcher != nil {
		counts, err := vh.searcher.TagDistribution(300)
		if err == nil {
			for _, tc := range counts {
				tags = append(tags, tagItem{Name: tc.Tag, Count: tc.Count})
			}
		}
	}

	indexNotes := vh.listIndexItems()
	knowledgeNotes, knowledgeNext := vh.listNoteItems("knowledge", 0, libraryPageSize)
	rawNotes, rawNext := vh.listNoteItems("raw", 0, libraryPageSize)
	notesCount, _ := vh.vault.CountNotes()
	knowledgeCount, _ := vh.vault.CountKnowledgeNotes()
	rawCount, _ := vh.vault.CountRawNotes()

	data := libraryPageData{
		Tags:           tags,
		IndexNotes:     indexNotes,
		KnowledgeNotes: knowledgeNotes,
		RawNotes:       rawNotes,
		NotesCount:     notesCount,
		IndexCount:     len(indexNotes),
		TagCount:       len(tags),
		KnowledgeCount: knowledgeCount,
		RawCount:       rawCount,
		KnowledgeNext:  knowledgeNext,
		RawNext:        rawNext,
	}

	var buf bytes.Buffer
	if err := libraryRefreshTemplate.Execute(&buf, data); err != nil {
		http.Error(w, fmt.Sprintf("library/refresh: render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

var libraryRefreshTemplate = template.Must(template.New("library-refresh").Parse(libraryRefreshHTML))
