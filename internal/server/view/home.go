package view

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/hardhacker/vaultr/internal/storage"
)

const homeRecentLimit = 20
const homeRecentShortsLimit = 15
const homeFolderLimit = 15

type homePageData struct {
	PinnedNotes      []noteItem
	Folders          []storage.DirSummary
	RecentShorts     []noteItem
	RecentRaw        []noteItem
	RecentKnowledge  []noteItem
	TotalNotes       int
	KnowledgeNotes   int
}

func (vh *ViewHandler) Home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pinned, err := vh.vault.ListPinnedNotes()
	if err != nil {
		http.Error(w, "home: list pinned: "+err.Error(), http.StatusInternalServerError)
		return
	}

	recentShorts, err := vh.vault.ListAllNotes(storage.ListOptions{
		SortByTime:  true,
		Limit:       homeRecentShortsLimit,
		OnlyKinds:  []storage.Kind{storage.KindShort},
	})
	if err != nil {
		http.Error(w, "home: list recent shorts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	recentKnowledge, err := vh.vault.ListAllNotes(storage.ListOptions{
		SortByTime:  true,
		Limit:       homeRecentLimit,
		OnlyKinds:  []storage.Kind{storage.KindKnowledge},
	})
	if err != nil {
		http.Error(w, "home: list recent knowledge: "+err.Error(), http.StatusInternalServerError)
		return
	}

	recentRaw, err := vh.vault.ListAllNotes(storage.ListOptions{
		SortByTime:     true,
		Limit:          homeRecentLimit,
		ExcludeKinds: []storage.Kind{storage.KindShort, storage.KindKnowledge, storage.KindIndex},
	})
	if err != nil {
		http.Error(w, "home: list recent raw: "+err.Error(), http.StatusInternalServerError)
		return
	}

	folders, err := vh.vault.ListDirs()
	if err != nil {
		http.Error(w, "home: list dirs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(folders) > homeFolderLimit {
		folders = folders[:homeFolderLimit]
	}

	totalNotes, _ := vh.vault.CountNotes()
	knowledgeNotes, _ := vh.vault.CountKnowledgeNotes()

	data := homePageData{
		PinnedNotes:     noteItemsFromNotes(pinned),
		Folders:         folders,
		RecentShorts:    noteItemsFromNotes(recentShorts),
		RecentRaw:       noteItemsFromNotes(recentRaw),
		RecentKnowledge: noteItemsFromNotes(recentKnowledge),
		TotalNotes:      totalNotes,
		KnowledgeNotes:  knowledgeNotes,
	}

	var buf bytes.Buffer
	if err := homePageTemplate.Execute(&buf, data); err != nil {
		http.Error(w, fmt.Sprintf("render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

func noteItemsFromNotes(notes []storage.Note) []noteItem {
	items := make([]noteItem, 0, len(notes))
	for _, n := range notes {
		items = append(items, noteToItem(n))
	}
	return items
}

var homeTemplateFuncs = template.FuncMap{
	"label": func(item noteItem) string {
		if item.Title != "" {
			return item.Title
		}
		return item.Name
	},
	"folderLabel": func(dir string) string {
		if dir == "/" {
			return "/"
		}
		if len(dir) > 1 && dir[0] == '/' {
			return dir[1:]
		}
		return dir
	},
	"folderColorIdx": func(dir string) int {
		sum := 0
		for _, c := range dir {
			sum += int(c)
		}
		return sum % 4
	},
}

var homePageHTML = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
` + headHTML(headOpts{title: "Home — Vaultr", withFonts: true, withTW: true, withAlpine: true, withHTMX: true}) + `  <style>
    :root {` + appTokensDark + `}
    html[data-theme="light"] {` + appTokensLight + `}
` + infoDialogCSS + navCSS + pixelCSS + topbarCSS + homeCSS + drawerCSS + noteSharedCSS + noteEditorCSS + searchOverlayStyles + confirmDialogCSS + shortDialogCSS + settingsModalCSS + `
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
<body x-data="homeCtrl()">
` + searchOnlyOverlayHTML + confirmDialogHTML + infoDialogHTML + shortDialogHTML + settingsModalHTML() + `
  <header class="lib-topbar">
    <div class="lib-topbar-spacer"></div>
    <div class="lib-topbar-actions">
      ` + shortTriggerButton + `
      <button type="button" class="lib-action-btn" title="Refresh home" @click="refresh()">` + topbarIconReload + `</button>
      <button type="button" class="lib-action-btn"
              :class="{ 'is-active': drawerOpen }"
              :title="/Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent) ? 'Reading drawer (⌘E)' : 'Reading drawer (Ctrl+E)'"
              @click="drawerOpen = !drawerOpen">
        ` + topbarIconPanel + `
      </button>
      <button type="button" class="lib-action-btn"
              :title="/Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent) ? 'Search (⌘K)' : 'Search (Ctrl+K)'"
              @click="window.dispatchEvent(new CustomEvent('open-search'))">` + topbarIconSearch + `</button>
    </div>
  </header>

  <div class="lib-body">
` + navHTML("home") + homeMainHTML + `
  </div>

` + drawerHTML + `

  <script>
  document.addEventListener('alpine:init', () => {
` + themeStoreScript + `
  });

` + keysJS + drawerScript + searchOverlayScript + confirmDialogJS + infoDialogJS + shortDialogJS + settingsCtrlJS + homeJS + `
  </script>
` + noteSharedJS + `
</body>
</html>`

var homePageTemplate = template.Must(template.New("home").Funcs(homeTemplateFuncs).Parse(homePageHTML))

// HomeRefresh handles GET /home/refresh — returns HTMX OOB fragments for the
// three dynamic sections (pinned, folders, recent) without a full page reload.
func (vh *ViewHandler) HomeRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/home") {
		return
	}

	pinned, err := vh.vault.ListPinnedNotes()
	if err != nil {
		http.Error(w, "home/refresh: list pinned: "+err.Error(), http.StatusInternalServerError)
		return
	}

	recentShorts, err := vh.vault.ListAllNotes(storage.ListOptions{
		SortByTime:  true,
		Limit:       homeRecentShortsLimit,
		OnlyKinds:  []storage.Kind{storage.KindShort},
	})
	if err != nil {
		http.Error(w, "home/refresh: list recent shorts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	recentKnowledge, err := vh.vault.ListAllNotes(storage.ListOptions{
		SortByTime:  true,
		Limit:       homeRecentLimit,
		OnlyKinds:  []storage.Kind{storage.KindKnowledge},
	})
	if err != nil {
		http.Error(w, "home/refresh: list recent knowledge: "+err.Error(), http.StatusInternalServerError)
		return
	}

	recentRaw, err := vh.vault.ListAllNotes(storage.ListOptions{
		SortByTime:     true,
		Limit:          homeRecentLimit,
		ExcludeKinds: []storage.Kind{storage.KindShort, storage.KindKnowledge, storage.KindIndex},
	})
	if err != nil {
		http.Error(w, "home/refresh: list recent raw: "+err.Error(), http.StatusInternalServerError)
		return
	}

	folders, err := vh.vault.ListDirs()
	if err != nil {
		http.Error(w, "home/refresh: list dirs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(folders) > homeFolderLimit {
		folders = folders[:homeFolderLimit]
	}

	totalNotes, _ := vh.vault.CountNotes()
	knowledgeNotes, _ := vh.vault.CountKnowledgeNotes()

	data := homePageData{
		PinnedNotes:     noteItemsFromNotes(pinned),
		Folders:         folders,
		RecentShorts:    noteItemsFromNotes(recentShorts),
		RecentRaw:       noteItemsFromNotes(recentRaw),
		RecentKnowledge: noteItemsFromNotes(recentKnowledge),
		TotalNotes:      totalNotes,
		KnowledgeNotes:  knowledgeNotes,
	}

	var buf bytes.Buffer
	if err := homeRefreshTemplate.Execute(&buf, data); err != nil {
		http.Error(w, fmt.Sprintf("home/refresh: render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

var homeRefreshTemplate = template.Must(template.New("home-refresh").Funcs(homeTemplateFuncs).Parse(`<p id="home-hero-lede" class="home-hero-lede" hx-swap-oob="true">{{.TotalNotes}} notes &middot; {{.KnowledgeNotes}} knowledge</p>
<span id="home-pinned-count" class="s-count" hx-swap-oob="true">{{len .PinnedNotes}}</span>
<div id="home-pinned-grid" class="card-grid" hx-swap-oob="true">
  {{range .PinnedNotes}}
  <div class="card" data-pi="{{folderColorIdx .Path}}">
    <svg class="card-pin-mark" fill="currentColor" viewBox="0 0 24 24"><path d="M17 3a2 2 0 0 1 2 2v15a1 1 0 0 1-1.496.868l-4.512-2.578a2 2 0 0 0-1.984 0l-4.512 2.578A1 1 0 0 1 5 20V5a2 2 0 0 1 2-2z"/></svg>
    <svg class="card-pin-mark card-pin-mark-px" fill="currentColor" viewBox="0 0 10 12" shape-rendering="crispEdges" style="display:none"><rect x="0" y="0" width="10" height="8"/><rect x="0" y="8" width="4" height="1"/><rect x="0" y="9" width="3" height="1"/><rect x="0" y="10" width="2" height="1"/><rect x="0" y="11" width="1" height="1"/><rect x="6" y="8" width="4" height="1"/><rect x="7" y="9" width="3" height="1"/><rect x="8" y="10" width="2" height="1"/><rect x="9" y="11" width="1" height="1"/></svg>
    <div class="card-a"
         @click="__vaultrOpenNote($event.currentTarget)"
         data-note-path="{{.Path}}"
         data-note-title="{{label .}}"
         data-note-is-knowledge="{{.IsKnowledge}}"
         data-note-is-index="{{.IsIndex}}"
         data-note-can-compile="{{.CanCompile}}"
         data-note-pinned="{{.Pinned}}">
      <span class="card-title">{{label .}}</span>
      <div class="card-foot">
        <span class="card-meta">{{.UpdatedAt}}</span>
        {{if .IsKnowledge}}<span class="card-ai">k</span>{{end}}
        {{if .IsShort}}<span class="card-short">s</span>{{end}}
        {{if .IsIndex}}<span class="card-index">i</span>{{end}}
        {{if .IsCompiled}}<span class="card-compiled"><svg fill="none" stroke="currentColor" stroke-width="4" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24" width="9" height="9"><circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/></svg></span>{{end}}
      </div>
    </div>
  </div>
  {{end}}
  {{if not .PinnedNotes}}<div class="grid-empty">No pinned notes</div>{{end}}
</div>
<span id="home-folders-count" class="s-count" hx-swap-oob="true">{{len .Folders}}</span>
<div id="home-folders-grid" class="card-grid" hx-swap-oob="true">
  {{range .Folders}}
  <div class="folder-card" data-fc="{{folderColorIdx .Dir}}" onclick="location.href='/dir?path='+encodeURIComponent('{{.Dir}}')">
    <svg class="folder-card-watermark" fill="currentColor" viewBox="0 0 24 24">
      <path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/>
    </svg>
    <svg class="folder-card-watermark folder-card-watermark-px" fill="currentColor" viewBox="0 0 24 24" shape-rendering="crispEdges" style="display:none">
      <rect x="0" y="0" width="11" height="3"/>
      <rect x="0" y="3" width="24" height="18"/>
    </svg>
    <div class="folder-card-inner">
      <div class="folder-card-top">
        <div class="folder-card-badge">
          <svg class="folder-card-icon" fill="none" stroke="currentColor" stroke-width="1.6" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/>
          </svg>
          <span class="folder-card-label">Folder</span>
        </div>
        <span class="folder-card-name">{{folderLabel .Dir}}</span>
      </div>
      <div class="folder-card-foot">
        <span class="folder-card-count">{{.Count}} notes</span>
      </div>
    </div>
  </div>
  {{end}}
  {{if not .Folders}}<div class="grid-empty">No folders</div>{{end}}
</div>
<div id="home-recent-shorts-grid" class="card-grid" hx-swap-oob="true">
  {{range .RecentShorts}}
  <div class="card">
    <div class="card-a"
         @click="__vaultrOpenNote($event.currentTarget)"
         data-note-path="{{.Path}}"
         data-note-title="{{label .}}"
         data-note-is-knowledge="{{.IsKnowledge}}"
         data-note-is-index="{{.IsIndex}}"
         data-note-can-compile="{{.CanCompile}}"
         data-note-pinned="{{.Pinned}}">
      <span class="card-title">{{label .}}</span>
      <div class="card-foot">
        <span class="card-meta">{{.UpdatedAt}}</span>
        {{if .IsShort}}<span class="card-short">s</span>{{end}}
      </div>
    </div>
  </div>
  {{end}}
  {{if not .RecentShorts}}<div class="grid-empty">No shorts</div>{{end}}
</div>
<div id="home-recent-raw-grid" class="card-grid" hx-swap-oob="true">
  {{range .RecentRaw}}
  <div class="card">
    <div class="card-a"
         @click="__vaultrOpenNote($event.currentTarget)"
         data-note-path="{{.Path}}"
         data-note-title="{{label .}}"
         data-note-is-knowledge="{{.IsKnowledge}}"
         data-note-is-index="{{.IsIndex}}"
         data-note-can-compile="{{.CanCompile}}"
         data-note-pinned="{{.Pinned}}">
      <span class="card-title">{{label .}}</span>
      <div class="card-foot">
        <span class="card-meta">{{.UpdatedAt}}</span>
        {{if .IsIndex}}<span class="card-index">i</span>{{end}}
        {{if .IsCompiled}}<span class="card-compiled"><svg fill="none" stroke="currentColor" stroke-width="4" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24" width="9" height="9"><circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/></svg></span>{{end}}
      </div>
    </div>
  </div>
  {{end}}
  {{if not .RecentRaw}}<div class="grid-empty">No notes</div>{{end}}
</div>
<div id="home-recent-knowledge-grid" class="card-grid" hx-swap-oob="true">
  {{range .RecentKnowledge}}
  <div class="card">
    <div class="card-a"
         @click="__vaultrOpenNote($event.currentTarget)"
         data-note-path="{{.Path}}"
         data-note-title="{{label .}}"
         data-note-is-knowledge="{{.IsKnowledge}}"
         data-note-is-index="{{.IsIndex}}"
         data-note-can-compile="{{.CanCompile}}"
         data-note-pinned="{{.Pinned}}">
      <span class="card-title">{{label .}}</span>
      <div class="card-foot">
        <span class="card-meta">{{.UpdatedAt}}</span>
        <span class="card-ai">k</span>
        {{if .IsCompiled}}<span class="card-compiled"><svg fill="none" stroke="currentColor" stroke-width="4" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24" width="9" height="9"><circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/></svg></span>{{end}}
      </div>
    </div>
  </div>
  {{end}}
  {{if not .RecentKnowledge}}<div class="grid-empty">No knowledge</div>{{end}}
</div>
`))
