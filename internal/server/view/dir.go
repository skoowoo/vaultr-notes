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

	"github.com/hardhacker/vaultr/internal/storage"
)

const dirPageSize = 20

type dirPageData struct {
	DirPath  string // vault-absolute dir path, e.g. "/" or "/journal/2026"
	DirLabel string // display label: "/" for root, else path without leading slash
	Notes    []noteItem
	NextNs   int64
	Count    int
	BackURL  string // back button destination, defaults to /home
}

type dirNotesFragData struct {
	DirPath string
	Notes   []noteItem
	NextNs  int64
}

// Dir handles GET /dir?path=DIR — full directory detail page.
func (vh *ViewHandler) Dir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dirPath := r.URL.Query().Get("path")
	if dirPath == "" {
		dirPath = "/"
	}
	if !strings.HasPrefix(dirPath, "/") {
		http.Error(w, "path must start with /", http.StatusBadRequest)
		return
	}
	notes, nextNs := vh.listDirNoteItems(dirPath, 0, dirPageSize)

	dirs, _ := vh.vault.ListAllDirs()
	var count int
	for _, d := range dirs {
		if d.Dir == dirPath {
			count = d.Count
			break
		}
	}

	label := dirPath
	if dirPath != "/" {
		label = strings.TrimPrefix(dirPath, "/")
	}

	backURL := "/home"
	if r.URL.Query().Get("from") == "folders" {
		backURL = "/folders"
	}

	data := dirPageData{
		DirPath:  dirPath,
		DirLabel: label,
		Notes:    notes,
		NextNs:   nextNs,
		Count:    count,
		BackURL:  backURL,
	}

	var buf bytes.Buffer
	if err := dirPageTemplate.Execute(&buf, data); err != nil {
		http.Error(w, fmt.Sprintf("render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

// DirRefresh handles GET /dir/refresh?path=DIR — HTMX OOB refresh of the notes grid.
func (vh *ViewHandler) DirRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/dir") {
		return
	}

	dirPath := r.URL.Query().Get("path")
	if dirPath == "" {
		dirPath = "/"
	}
	if !strings.HasPrefix(dirPath, "/") {
		http.Error(w, "path must start with /", http.StatusBadRequest)
		return
	}
	notes, nextNs := vh.listDirNoteItems(dirPath, 0, dirPageSize)

	dirs, _ := vh.vault.ListAllDirs()
	var count int
	for _, d := range dirs {
		if d.Dir == dirPath {
			count = d.Count
			break
		}
	}

	data := dirPageData{
		DirPath: dirPath,
		Notes:   notes,
		NextNs:  nextNs,
		Count:   count,
	}

	var buf bytes.Buffer
	if err := dirRefreshTemplate.Execute(&buf, data); err != nil {
		http.Error(w, fmt.Sprintf("dir/refresh: render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

// DirNotes handles GET /dir/notes?path=DIR&before=NS — HTMX scroll fragment.
func (vh *ViewHandler) DirNotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/dir") {
		return
	}

	dirPath := r.URL.Query().Get("path")
	if dirPath == "" || !strings.HasPrefix(dirPath, "/") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	var beforeNs int64
	if s := r.URL.Query().Get("before"); s != "" {
		beforeNs, _ = strconv.ParseInt(s, 10, 64)
	}

	notes, nextNs := vh.listDirNoteItems(dirPath, beforeNs, dirPageSize)
	data := dirNotesFragData{DirPath: dirPath, Notes: notes, NextNs: nextNs}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := dirNotesFragTemplate.Execute(w, data); err != nil {
		http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
	}
}

func (vh *ViewHandler) listDirNoteItems(dir string, beforeNs int64, limit int) ([]noteItem, int64) {
	opts := storage.ListOptions{
		SortByTime: true,
		Limit:      limit + 1,
	}
	if beforeNs > 0 {
		opts.Before = time.Unix(0, beforeNs)
	}

	notes, err := vh.vault.ListDirNotes(dir, opts)
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

// dirHasUnderscoreView mirrors storage.dirHasUnderscoreSegment for the view layer.
func dirHasUnderscoreView(dir string) bool {
	for _, seg := range strings.Split(dir, "/") {
		if strings.HasPrefix(seg, "_") {
			return true
		}
	}
	return false
}

// ─── templates ────────────────────────────────────────────────────────────────

var dirTemplateFuncs = template.FuncMap{
	"label": func(item noteItem) string {
		if item.Title != "" {
			return item.Title
		}
		return item.Name
	},
	"encdir": url.QueryEscape,
}

var dirNotesFragTemplate = template.Must(template.New("dirnotesfrag").Funcs(dirTemplateFuncs).Parse(dirNotesFragHTML))

var dirRefreshTemplate = template.Must(template.New("dirrefresh").Funcs(dirTemplateFuncs).Parse(dirRefreshHTML))

var dirPageHTML = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
` + headHTML(headOpts{title: "Folder — Vaultr", withFonts: true, withTW: true, withAlpine: true, withHTMX: true}) + `  <style>
    :root {` + appTokensDark + `}
    html[data-theme="light"] {` + appTokensLight + `}
` + infoDialogCSS + navCSS + pixelCSS + topbarCSS + dirCSS + drawerCSS + noteSharedCSS + noteEditorCSS + searchOverlayStyles + confirmDialogCSS + shortDialogCSS + settingsModalCSS + `
  </style>
  <script>
  /* Apply hero background synchronously before first paint to avoid flash */
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
<body x-data="dirCtrl()">
` + searchOnlyOverlayHTML + confirmDialogHTML + infoDialogHTML + shortDialogHTML + settingsModalHTML() + `
  <header class="lib-topbar">
    <div class="lib-topbar-left">
      <a href="{{.BackURL}}" class="lib-back-btn" title="Back">` + topbarIconBack + `<span class="lib-back-label">back</span></a>
    </div>
    <div class="lib-topbar-spacer"></div>
    <div class="lib-topbar-actions">
      ` + shortTriggerButton + `
      <button type="button" class="lib-action-btn" title="Refresh" @click="window.location.reload()">` + topbarIconReload + `</button>
      <button type="button" class="lib-action-btn"
              :class="{ 'is-active': drawerOpen }"
              :title="/Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent) ? 'Reading drawer (⌘E)' : 'Reading drawer (Ctrl+E)'"
              @click="drawerOpen = !drawerOpen">` + topbarIconPanel + `</button>
      <button type="button" class="lib-action-btn"
              :title="/Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent) ? 'Search (⌘K)' : 'Search (Ctrl+K)'"
              @click="window.dispatchEvent(new CustomEvent('open-search'))">` + topbarIconSearch + `</button>
    </div>
  </header>

  <div class="lib-body">
` + navHTML("dir") + dirMainHTML + `
  </div>

` + drawerHTML + `

  <script>
  document.addEventListener('alpine:init', () => {
` + themeStoreScript + `
  });

` + keysJS + drawerScript + searchOverlayScript + confirmDialogJS + infoDialogJS + shortDialogJS + settingsCtrlJS + dirJS + `
  </script>
` + noteSharedJS + `
</body>
</html>`

var dirPageTemplate = template.Must(template.New("dir").Funcs(dirTemplateFuncs).Parse(dirPageHTML))
