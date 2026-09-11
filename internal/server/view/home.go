package view

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hardhacker/vaultr/internal/storage"
)

const homeListPageSize = 20

type homePageData struct {
	Folders     []storage.DirSummary
	IndexNotes  []noteItem
	SectionHTML template.HTML
}

// homeSectionData drives the right-hand note list — the content shown for
// whichever sidebar item (Pinned / a folder) is active. Shorts is
// a different "view kind" (the rendered shorts stream, not a note list) and
// is rendered separately by renderHomeShortsSection; see renderHomeSectionHTML.
type homeSectionData struct {
	Title    string
	Count    int
	Items    []noteItem
	NextURL  string // hx-get URL for the load-more sentinel; empty = no more pages
	EmptyMsg string
}

func (vh *ViewHandler) Home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	folders, err := vh.vault.ListAllDirs()
	if err != nil {
		http.Error(w, "home: list dirs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sortDirsSystemLast(folders)

	indexNotes := vh.listIndexItems()

	sectionHTML, err := vh.renderHomeSectionHTML(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("home: render section: %s", err), http.StatusInternalServerError)
		return
	}

	data := homePageData{
		Folders:     folders,
		IndexNotes:  indexNotes,
		SectionHTML: sectionHTML,
	}

	var buf bytes.Buffer
	if err := homePageTemplate.Execute(&buf, data); err != nil {
		http.Error(w, fmt.Sprintf("render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

// homeSectionData resolves the query params of a /home, /home/section, or
// /home/section/more request into the note list + head info for the active
// sidebar selection. itemsOnly skips the (slightly more expensive) total-count
// lookups, since /home/section/more only ever re-renders the note rows.
func (vh *ViewHandler) homeSectionData(r *http.Request, itemsOnly bool) (homeSectionData, error) {
	sectionType := r.URL.Query().Get("type")
	var beforeNs int64
	if s := r.URL.Query().Get("before"); s != "" {
		beforeNs, _ = strconv.ParseInt(s, 10, 64)
	}

	var data homeSectionData

	switch sectionType {
	case "folder":
		dirPath := r.URL.Query().Get("path")
		if dirPath == "" || !strings.HasPrefix(dirPath, "/") {
			return data, fmt.Errorf("path must start with /")
		}
		var nextNs int64
		data.Items, nextNs = vh.listDirNoteItems(dirPath, beforeNs, homeListPageSize)
		label := dirPath
		if dirPath != "/" {
			label = strings.TrimPrefix(dirPath, "/")
		}
		data.Title = label
		data.EmptyMsg = "No notes in this folder"
		data.NextURL = homeSectionMoreURL("folder", dirPath, nextNs)
		if !itemsOnly {
			dirs, _ := vh.vault.ListAllDirs()
			for _, d := range dirs {
				if d.Dir == dirPath {
					data.Count = d.Count
					break
				}
			}
		}
	default: // "pinned"
		pinned, err := vh.vault.ListPinnedNotes()
		if err != nil {
			return data, err
		}
		data.Items = noteItemsFromNotes(pinned)
		data.Title = "Pinned"
		data.EmptyMsg = "No pinned notes"
		data.Count = len(data.Items)
	}
	return data, nil
}

func homeSectionMoreURL(typ, path string, nextNs int64) string {
	if nextNs == 0 {
		return ""
	}
	q := url.Values{}
	q.Set("type", typ)
	if path != "" {
		q.Set("path", path)
	}
	q.Set("before", strconv.FormatInt(nextNs, 10))
	return "/home/section/more?" + q.Encode()
}

func renderHomeSectionFull(data homeSectionData) (template.HTML, error) {
	var buf bytes.Buffer
	if err := homeSectionTemplate.ExecuteTemplate(&buf, "full", data); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil //nolint:gosec // server-rendered fragment, not user HTML
}

// renderHomeSectionHTML resolves a /home or /home/section request into the
// #home-list-pane contents for whichever sidebar item is active. This is the
// dispatch point between the right pane's different "view kinds": today a
// generic note list (pinned/folder), the rendered Shorts stream, the image
// gallery grid, the knowledge graph canvas, the inbox message list, or the
// agent chat panel; more kinds can be added here as new cases without
// touching the sidebar's swap mechanism (see home.js
// selectSection/selectFolder).
func (vh *ViewHandler) renderHomeSectionHTML(r *http.Request) (template.HTML, error) {
	switch r.URL.Query().Get("type") {
	case "shorts":
		return vh.renderHomeShortsSection(r)
	case "images":
		return vh.renderHomeImagesSection(r)
	case "graph":
		// The graph canvas is entirely client-rendered (home.js fetches
		// /api/graph/data itself once #graph-canvas lands in the DOM, reading
		// ?index= off the URL that was just loaded — see the htmx:afterSwap
		// listener in home.js), so this markup never varies with the request.
		return template.HTML(homeGraphSectionHTML), nil //nolint:gosec // static markup, not user HTML
	case "inbox":
		// Same story as graph: the message list is entirely client-rendered
		// (home.js fetches /api/inbox itself once #home-inbox-list lands in
		// the DOM — see the htmx:afterSwap listener), so this markup never
		// varies with the request.
		return template.HTML(homeInboxSectionHTML), nil //nolint:gosec // static markup, not user HTML
	case "chat":
		// Same story again: mates/conversation state lives entirely in
		// homeCtrl (home.js fetches /api/mates, /api/conversations, etc.
		// once #chat-scroll lands in the DOM — see the htmx:afterSwap
		// listener), so this markup never varies with the request.
		return template.HTML(homeChatSectionHTML), nil //nolint:gosec // static markup, not user HTML
	}
	data, err := vh.homeSectionData(r, false)
	if err != nil {
		return "", err
	}
	return renderHomeSectionFull(data)
}

// renderHomeShortsSection renders the same rendered Shorts stream (composer,
// date-grouped entries, month rail) as the standalone /shorts page, reusing
// its data loaders (shorts.go). Pagination ("Load earlier") and the month
// rail both stay in place inside #home-list-pane via hx-get, rather than
// navigating to /shorts — unlike the note-list view kind, they don't go
// through /home/section/more: the month rail re-requests this same endpoint
// with &from=, and "Load earlier" hits the existing /shorts/stream endpoint
// directly (its #shorts-stream-groups/#shorts-load-more ids match here too).
func (vh *ViewHandler) renderHomeShortsSection(r *http.Request) (template.HTML, error) {
	before := time.Now().Add(time.Second)
	activeYM := time.Now().Format("2006-01")
	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.ParseInLocation("2006-01", from, time.Local); err == nil {
			before = t.AddDate(0, 1, 0) // first moment of next month = end of selected month
			activeYM = from
		}
	}

	groups, cursor, hasMore, err := vh.loadStreamGroups(before, 50)
	if err != nil {
		return "", err
	}

	data := shortsStreamPageData{
		Groups:  groups,
		Cursor:  cursor,
		HasMore: hasMore,
		Months:  vh.loadMonthsWithEntries(activeYM),
	}

	var buf bytes.Buffer
	if err := homeShortsSectionTemplate.Execute(&buf, data); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil //nolint:gosec // server-rendered fragment, not user HTML
}

// renderHomeImagesSection renders the same image gallery grid (toolbar,
// thumbnail grid, select/delete) as the standalone /images page, reusing its
// data loader and item type (images.go). Pagination ("scroll for more") stays
// in place via the existing /images/grid endpoint — its #img-grid target and
// sentinel markup match here too, so no /home-specific pagination is needed.
// The lightbox overlay this grid opens lives once at the page level (see
// homePageHTML) rather than inside this swappable fragment.
func (vh *ViewHandler) renderHomeImagesSection(r *http.Request) (template.HTML, error) {
	imgs, err := vh.vault.ListImages(0, imagesPageSize+1)
	if err != nil {
		return "", err
	}

	hasMore := len(imgs) > imagesPageSize
	if hasMore {
		imgs = imgs[:imagesPageSize]
	}

	items := make([]imageItem, 0, len(imgs))
	for _, img := range imgs {
		items = append(items, imageItemFrom(img))
	}

	var nextNs int64
	if hasMore && len(items) > 0 {
		nextNs = items[len(items)-1].CursorNs
	}

	var buf bytes.Buffer
	if err := homeImagesSectionTemplate.Execute(&buf, imagesGridData{Images: items, NextNs: nextNs}); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil //nolint:gosec // server-rendered fragment, not user HTML
}

// HomeSection handles GET /home/section?type=pinned|folder|shorts|images|graph|inbox|chat[&path=DIR]
// Returns the full #home-list-pane contents, used when the sidebar selection changes.
func (vh *ViewHandler) HomeSection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/home") {
		return
	}

	html, err := vh.renderHomeSectionHTML(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

// HomeSectionMore handles GET /home/section/more — same params as HomeSection,
// triggered by the load-sentinel at the bottom of the list. Returns only the
// next page of rows + a following sentinel (no head), swapped in via outerHTML.
func (vh *ViewHandler) HomeSectionMore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/home") {
		return
	}

	data, err := vh.homeSectionData(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := homeSectionTemplate.ExecuteTemplate(w, "rows", data); err != nil {
		http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
	}
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
	"encdir": url.QueryEscape,
}

const homeSectionRowsHTML = `{{define "rows"}}{{range .Items}}
<div class="home-note-row" @click="__vaultrOpenNote($event.currentTarget)"
     data-note-path="{{.Path}}" data-note-title="{{label .}}"
     data-note-is-knowledge="{{.IsKnowledge}}" data-note-is-index="{{.IsIndex}}"
     data-note-can-compile="{{.CanCompile}}" data-note-pinned="{{.Pinned}}">
  <div class="home-note-row-top">
    <span class="home-note-row-title">{{label .}}</span>
    <div class="home-note-row-badges">
      {{if .Pinned}}<span class="home-note-badge home-note-badge--pin" title="Pinned"><svg fill="currentColor" viewBox="0 0 24 24"><path d="M17 3a2 2 0 0 1 2 2v15a1 1 0 0 1-1.496.868l-4.512-2.578a2 2 0 0 0-1.984 0l-4.512 2.578A1 1 0 0 1 5 20V5a2 2 0 0 1 2-2z"/></svg></span>{{end}}
      {{if .IsKnowledge}}<span class="home-note-badge home-note-badge--k" title="Knowledge">K</span>{{end}}
      {{if .IsShort}}<span class="home-note-badge home-note-badge--s" title="Short">S</span>{{end}}
      {{if .IsIndex}}<span class="home-note-badge home-note-badge--i" title="Index">I</span>{{end}}
      {{if .IsCompiled}}<span class="home-note-badge home-note-badge--done" title="Compiled"><svg fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24"><path d="M20 6 9 17l-5-5"/></svg></span>{{end}}
    </div>
  </div>
  <div class="home-note-row-meta">
    {{if ne .Dir "/"}}<span class="home-note-row-dir">{{.Dir}}</span>{{end}}
    <span class="home-note-row-time">{{.UpdatedAt}}</span>
  </div>
</div>
{{end}}{{if .NextURL}}
<div class="home-load-sentinel" hx-get="{{.NextURL}}" hx-trigger="intersect once" hx-swap="outerHTML"></div>
{{end}}{{if not .Items}}
<div class="home-list-empty">{{.EmptyMsg}}</div>
{{end}}{{end}}`

const homeSectionFullHTML = `{{define "full"}}<div class="home-list-head">
  <span class="home-list-title" id="home-list-title">{{.Title}}</span>
  <span class="home-list-count" id="home-list-count">{{.Count}}</span>
</div>
<div class="home-list-body" id="home-list-body">{{template "rows" .}}</div>{{end}}`

var homeSectionTemplate = template.Must(
	template.Must(
		template.New("rows").Funcs(homeTemplateFuncs).Parse(homeSectionRowsHTML),
	).New("full").Funcs(homeTemplateFuncs).Parse(homeSectionFullHTML),
)

// homeShortsSectionHTML mirrors shorts.html's .shorts-stream-layout (feed +
// month rail) verbatim, with one difference: the month rail re-requests this
// same #home-list-pane in place (hx-get) instead of navigating to /shorts —
// everything here must keep working when it's /shorts's markup embedded
// inside home rather than the standalone page.
const homeShortsSectionHTML = `<div class="shorts-stream-layout">
  <div class="shorts-stream-wrap">
    <div class="shorts-stream-inner">
      <button type="button" class="shorts-compose"
              onclick="window.openShortDialog && window.openShortDialog()">
        Write a short...
        <svg fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" viewBox="0 0 24 24"><path d="M5 12h14"/><path d="M12 5v14"/></svg>
      </button>
      <div id="shorts-stream-groups">
        {{range .Groups}}
        {{$date := .CompactDate}}
        {{range .Entries}}
        <div class="shorts-entry">
          <span class="shorts-entry-meta">{{$date}} · {{.Time}}</span>
          <div class="prose shorts-entry-prose">{{.HTML}}</div>
        </div>
        {{end}}
        {{else}}
        <div class="shorts-empty-state">
          <p class="shorts-empty-label">No shorts yet</p>
          <button type="button" class="shorts-empty-btn"
                  onclick="window.openShortDialog && window.openShortDialog()">
            <svg fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" viewBox="0 0 24 24"><path d="M5 12h14"/><path d="M12 5v14"/></svg>Write a short
          </button>
        </div>
        {{end}}
      </div>

      <div id="shorts-load-more">
        {{if .HasMore}}
        <div class="shorts-load-more-wrap">
          <button class="shorts-loadmore-btn"
                  hx-get="/shorts/stream?before={{.Cursor}}"
                  hx-target="#shorts-stream-groups"
                  hx-swap="beforeend">
            Load earlier
          </button>
        </div>
        {{end}}
      </div>
    </div>
  </div>

  <nav class="shorts-month-rail">
    {{range .Months}}
    <button type="button" class="shorts-month-item{{if .IsCurrent}} is-current{{end}}{{if .HasData}} has-data{{end}}"
            hx-get="/home/section?type=shorts&from={{.YM}}" hx-target="#home-list-pane" hx-swap="innerHTML">
      <span class="shorts-month-abbr">{{.Abbr}}</span>
      <span class="shorts-month-year">{{.Year}}</span>
    </button>
    {{end}}
  </nav>
</div>`

var homeShortsSectionTemplate = template.Must(template.New("home-shorts-section").Parse(homeShortsSectionHTML))

// homeImagesSectionHTML mirrors images.html's toolbar + scroll + grid
// verbatim — same markup, ids, and hx-get pagination target as the
// standalone /images page, just without its own .img-main wrapper (the
// #home-list-pane flex column already plays that role).
const homeImagesSectionHTML = `<div class="img-toolbar">
  <button type="button" class="toolbar-toggle" x-show="!selectMode" @click="enterSelectMode()">
    <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect width="18" height="18" x="3" y="3" rx="2"/><path stroke-linecap="round" stroke-linejoin="round" d="m9 12 2 2 4-4"/></svg>
    Select
  </button>
  <div class="toolbar-select-group" x-show="selectMode" style="display:none">
    <div class="bulk-count"><strong x-text="selectedCount"></strong> selected</div>
    <button type="button" class="bulk-btn bulk-btn-danger"
            :disabled="!selectedCount"
            @click="deleteSelected()"
            x-text="selectedCount ? 'Delete (' + selectedCount + ')' : 'Delete'">
    </button>
    <button type="button" class="bulk-btn bulk-btn-neutral" @click="exitSelectMode()">Cancel</button>
  </div>
</div>
<div class="img-scroll">
  <div class="img-grid" id="img-grid">
    {{- if .Images}}
    {{- range .Images}}
    <div class="img-card" onclick="openImageLightbox(this)"
         data-img-src="{{.ThumbURL}}" data-img-name="{{.Name}}" data-img-dir="{{.Dir}}"
         data-img-size="{{.Size}}" data-img-time="{{.UpdatedAt}}" data-img-ext="{{.Ext}}"
         data-img-notes="{{.LinkedNotesJSON}}">
      <div class="img-thumb-wrap">
        <img class="img-thumb" src="{{.ThumbURL}}" alt="{{.Name}}" loading="lazy" decoding="async">
        <div class="img-select-check"><svg fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"/></svg></div>
        {{- if .LinkedNotes}}
        <div class="img-link-badge">
          <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M6 22a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h8a2.4 2.4 0 0 1 1.704.706l3.588 3.588A2.4 2.4 0 0 1 20 8v12a2 2 0 0 1-2 2z"/><path stroke-linecap="round" stroke-linejoin="round" d="M14 2v5a1 1 0 0 0 1 1h5"/></svg>
          {{len .LinkedNotes}}
        </div>
        {{- end}}
      </div>
      <div class="img-info">
        <div class="img-caption">{{if .Caption}}{{.Caption}}{{else}}{{.Name}}{{end}}</div>
        <div class="img-time">{{.UpdatedAt}}</div>
      </div>
    </div>
    {{- end}}
    {{- if .NextNs}}
    <div class="img-sentinel"
         hx-get="/images/grid?before={{.NextNs}}"
         hx-trigger="intersect once"
         hx-swap="beforebegin"
         hx-target="this"></div>
    {{- end}}
    {{- else}}
    <div class="img-empty">No images found</div>
    {{- end}}
  </div>
</div>`

var homeImagesSectionTemplate = template.Must(template.New("home-images-section").Parse(homeImagesSectionHTML))

// homeImagesLightboxHTML mirrors images.go's lightbox overlay verbatim. It
// lives once at the page level (spliced into homePageHTML below), not inside
// the swappable #home-list-pane, since it's a full-viewport modal driven by
// homeCtrl's lightbox/selectMode state regardless of which sidebar section
// is currently showing.
const homeImagesLightboxHTML = `
  <div class="lb-overlay"
       x-show="lightbox"
       x-transition:enter="lb-enter"
       x-transition:enter-start="lb-enter-start"
       x-transition:enter-end="lb-enter-end"
       x-transition:leave="lb-leave"
       x-transition:leave-start="lb-leave-start"
       x-transition:leave-end="lb-leave-end"
       @click.self="closeLightbox()"
       @keydown.escape.window="closeLightbox()"
       style="display:none">
    <div class="lb-panel" @click.stop>
      <div class="lb-viewer">
        <img class="lb-viewer-img"
             :src="lightbox ? lightbox.src : ''"
             :alt="lightbox ? lightbox.name : ''"
             draggable="false">
      </div>
      <aside class="lb-sidebar">
        <div class="lb-sidebar-head">
          <span class="lb-sidebar-title" x-text="lightbox ? lightbox.name : ''"></span>
          <button type="button" class="lb-delete-btn" title="Delete image"
                  x-show="lightbox"
                  @click="deleteLightboxImage()">
            <svg fill="none" stroke="currentColor" stroke-width="1.9" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M3 6h18"/><path stroke-linecap="round" stroke-linejoin="round" d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><path stroke-linecap="round" stroke-linejoin="round" d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"/><path stroke-linecap="round" stroke-linejoin="round" d="M10 11v6"/><path stroke-linecap="round" stroke-linejoin="round" d="M14 11v6"/></svg>
          </button>
          <button type="button" class="icon-btn-close icon-btn-close--sm" @click="closeLightbox()" title="Close (Esc)">
            <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" d="M18 6 6 18"/><path stroke-linecap="round" d="m6 6 12 12"/></svg>
          </button>
        </div>
        <div class="lb-sidebar-body">
          <div class="lb-section-label">Linked Notes</div>
          <template x-if="lightbox && lightbox.notes && lightbox.notes.length > 0">
            <div>
              <template x-for="note in lightbox.notes" :key="note">
                <div class="lb-note-card" @click="openLinkedNote(note)">
                  <svg class="lb-note-card-icon" fill="none" stroke="currentColor" stroke-width="1.7" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M6 22a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h8a2.4 2.4 0 0 1 1.704.706l3.588 3.588A2.4 2.4 0 0 1 20 8v12a2 2 0 0 1-2 2z"/><path stroke-linecap="round" stroke-linejoin="round" d="M14 2v5a1 1 0 0 0 1 1h5"/></svg>
                  <span class="lb-note-card-name" x-text="note"></span>
                  <svg class="lb-note-card-arrow" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 18l6-6-6-6"/>
                  </svg>
                </div>
              </template>
            </div>
          </template>
          <template x-if="!lightbox || !lightbox.notes || lightbox.notes.length === 0">
            <div class="lb-no-notes">No linked notes</div>
          </template>
          <div class="lb-divider"></div>
          <div class="lb-section-label">Info</div>
          <div class="lb-field">
            <div class="lb-field-label">Filename</div>
            <div class="lb-field-value mono" x-text="lightbox ? lightbox.name : ''"></div>
          </div>
          <div class="lb-field">
            <div class="lb-field-label">Type</div>
            <div class="lb-field-value" x-text="lightbox ? extToType(lightbox.ext) : ''"></div>
          </div>
          <div class="lb-field">
            <div class="lb-field-label">Size</div>
            <div class="lb-field-value" x-text="lightbox ? lightbox.size : ''"></div>
          </div>
          <div class="lb-field">
            <div class="lb-field-label">Location</div>
            <div class="lb-field-value mono" x-text="lightbox ? lightbox.dir : ''"></div>
          </div>
          <div class="lb-field">
            <div class="lb-field-label">Modified</div>
            <div class="lb-field-value" x-text="lightbox ? lightbox.time : ''"></div>
          </div>
        </div>
      </aside>
    </div>
  </div>`

// homeGraphSectionHTML mirrors graph.html's .graph-main block (canvas, zoom
// controls, node info panel, loading/empty states) verbatim — everything
// except the .graph-index-col sidebar, which is now the sidebar's Graph
// group (see home.html) instead of living inside the swappable pane.
const homeGraphSectionHTML = `<div class="graph-main">
  <div style="position:relative;flex:1;display:flex;flex-direction:column;overflow:hidden">
    <div id="graph-canvas" style="flex:1;width:100%"></div>

    <div class="graph-zoom-controls">
      <button type="button" class="graph-zoom-btn" title="Zoom in" @click="zoomIn()">
        <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" d="M12 5v14M5 12h14"/>
        </svg>
      </button>
      <div class="graph-zoom-divider"></div>
      <button type="button" class="graph-zoom-btn" title="Zoom out" @click="zoomOut()">
        <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" d="M5 12h14"/>
        </svg>
      </button>
      <div class="graph-zoom-divider"></div>
      <button type="button" class="graph-zoom-btn" title="Fit all nodes" @click="zoomFit()">
        <svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M8 3H5a2 2 0 0 0-2 2v3"/>
          <path stroke-linecap="round" stroke-linejoin="round" d="M21 8V5a2 2 0 0 0-2-2h-3"/>
          <path stroke-linecap="round" stroke-linejoin="round" d="M3 16v3a2 2 0 0 0 2 2h3"/>
          <path stroke-linecap="round" stroke-linejoin="round" d="M16 21h3a2 2 0 0 0 2-2v-3"/>
        </svg>
      </button>
    </div>

    <div class="graph-node-panel" :class="{ open: !!nodePanel }">
      <div class="graph-node-panel-row">
        <span class="graph-node-panel-type"
              x-show="nodePanel && nodePanel.entityType"
              x-text="nodePanel ? nodePanel.entityType : ''"></span>
        <span class="graph-node-panel-edges"
              x-text="nodePanel ? (nodePanel.edgeCount + (nodePanel.edgeCount === 1 ? ' link' : ' links')) : ''"></span>
        <button type="button" class="graph-node-panel-close" @click="closeNodePanel()" title="Close">
          <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" d="M18 6 6 18"/>
            <path stroke-linecap="round" d="m6 6 12 12"/>
          </svg>
        </button>
      </div>
      <div class="graph-node-panel-title" x-text="nodePanel ? nodePanel.label : ''"></div>
      <button type="button" class="graph-node-open-btn"
              @click="nodePanel && openNodeInDrawer(nodePanel.path, nodePanel.label, nodePanel.entityType)">
        Open
        <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 3h6v6"/>
          <path stroke-linecap="round" stroke-linejoin="round" d="M10 14 21 3"/>
          <path stroke-linecap="round" stroke-linejoin="round" d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
        </svg>
      </button>
    </div>

    <div class="graph-loading" x-show="loading" x-cloak>
      <span>Loading graph…</span>
    </div>

    <div class="graph-zero" x-show="!loading && empty" x-cloak>
      <div class="graph-zero-icon">
        <svg fill="none" stroke="currentColor" stroke-width="1.25" viewBox="0 0 48 48">
          <circle cx="24" cy="12" r="4.5"/>
          <circle cx="10" cy="36" r="4.5"/>
          <circle cx="38" cy="36" r="4.5"/>
          <line x1="24" y1="16.5" x2="10.8" y2="31.7" stroke-linecap="round"/>
          <line x1="24" y1="16.5" x2="37.2" y2="31.7" stroke-linecap="round"/>
          <line x1="14.5" y1="36" x2="33.5" y2="36" stroke-linecap="round"/>
        </svg>
      </div>
      <div class="graph-zero-title">No knowledge notes yet</div>
      <div class="graph-zero-desc">Knowledge notes and their connections will appear here once the compile agent has run on your raw notes.</div>
    </div>
  </div>
</div>`

// homeInboxSectionHTML is the right-pane inbox view: a filter bar (all/
// unread/read + mark-all-read) over a wide-card message list, mirroring the
// visual language of the pinned/folder note list (.home-note-row) rather
// than the standalone /agent/inbox page's narrow sidebar rows. Entirely
// client-rendered — home.js fetches /api/inbox once this markup lands in the
// DOM (see the htmx:afterSwap listener) and drives the x-for below.
// Selecting a card opens the read-only message sheet (homeInboxSheetHTML),
// not the note-editor drawer.
const homeInboxSectionHTML = `<div class="home-inbox-section">
  <div class="home-inbox-head">
    <span class="home-list-title">Inbox</span>
    <span class="home-list-count" x-show="unreadCount > 0" x-text="unreadCount + ' unread'"></span>
    <div class="home-inbox-head-spacer"></div>
    <div class="home-inbox-filter-seg">
      <button type="button" class="home-inbox-filter-seg-btn" :class="inboxFilter === 'all' ? 'active' : ''" @click="setInboxFilter('all')">All</button>
      <button type="button" class="home-inbox-filter-seg-btn" :class="inboxFilter === 'unread' ? 'active' : ''" @click="setInboxFilter('unread')">Unread</button>
      <button type="button" class="home-inbox-filter-seg-btn" :class="inboxFilter === 'read' ? 'active' : ''" @click="setInboxFilter('read')">Read</button>
    </div>
    <button type="button" class="home-inbox-mark-all-btn" title="Mark all read" x-show="unreadCount > 0" @click="markAllInboxRead()">` + svgCheck + `</button>
  </div>
  <div class="home-inbox-list" id="home-inbox-list" @scroll="onInboxListScroll($event)">
    <template x-if="!inboxLoading && visibleInboxMessages().length === 0">
      <div class="home-list-empty" x-text="inboxFilter === 'unread' ? 'All caught up — no unread messages.' : (inboxFilter === 'read' ? 'No read messages yet.' : 'No messages yet.')"></div>
    </template>
    <template x-for="m in visibleInboxMessages()" :key="m.id">
      <div class="home-inbox-card" :class="{ 'is-unread': !m.isRead }" @click="selectInboxMessage(m)">
        <div class="home-inbox-unread-dot"></div>
        <div class="home-inbox-card-content">
          <div class="home-inbox-card-top">
            <span class="home-inbox-card-title" x-text="m.title || m.source"></span>
            <span class="home-inbox-card-time" x-text="relTime(m.createdAt)"></span>
          </div>
          <div class="home-inbox-card-snippet" x-text="m.body"></div>
        </div>
      </div>
    </template>
    <div class="home-list-empty" x-show="inboxLoadingMore" style="padding:1rem 0">Loading more…</div>
  </div>
</div>`

// homeInboxSheetHTML is the read-only slide-in panel opened by clicking an
// inbox card — distinct from the note-editor drawer (drawerHTML): it only
// ever displays a message's rendered body, never edits anything. Lives once
// at the page level (spliced into homePageHTML below), not inside the
// swappable #home-list-pane, so it survives switching sidebar sections while
// still animating closed via inboxSheetOpen.
const homeInboxSheetHTML = `
  <div class="inbox-sheet-overlay" :class="{ open: inboxSheetOpen }"
       @click.self="closeInboxSheet()"
       @keydown.escape.window="closeInboxSheet()">
    <div class="inbox-sheet-panel">
      <div class="inbox-sheet-head">
        <span class="inbox-sheet-title" x-text="inboxSelected ? (inboxSelected.title || inboxSelected.source) : ''"></span>
        <button type="button" class="inbox-sheet-close" @click="closeInboxSheet()" title="Close (Esc)">
          <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" d="M18 6 6 18"/><path stroke-linecap="round" d="m6 6 12 12"/></svg>
        </button>
      </div>
      <div class="inbox-sheet-meta">
        <span x-text="inboxSelected ? inboxSelected.source : ''"></span>
        <span>·</span>
        <span x-text="inboxSelected ? fullTime(inboxSelected.createdAt) : ''"></span>
      </div>
      <div class="inbox-sheet-scroll">
        <div class="prose inbox-sheet-body" x-html="inboxSelected ? renderMarkdown(inboxSelected.body) : ''"></div>
      </div>
    </div>
  </div>`

// homeChatSectionHTML mirrors agent_chat.html's .chat-main block (mate chips,
// message thread, composer) verbatim — everything except the standalone
// page's own nav rail, which home's sidebar already replaces. Entirely
// client-rendered: home.js fetches /api/mates, /api/conversations, etc. once
// this markup lands in the DOM (see the htmx:afterSwap listener) and drives
// the mates/messages state that's now merged into homeCtrl.
const homeChatSectionHTML = `<div class="chat-main">

  <!-- ── Mate bar (mate selection itself now lives in the sidebar's Chats
       group — see home.html — this keeps only the description, conv-type
       toggle, and new-chat button) ────────────────────────────── -->
  <div class="mate-chips-bar">
    <span class="mate-bar-desc" x-show="selectedMate && selectedMate.description"
      x-text="selectedMate ? selectedMate.description : ''"></span>
    <div class="chat-bar-spacer"></div>

    <!-- ── Conv type segmented control ───────────────────── -->
    <div class="conv-seg" x-show="selectedMateId">
      <template x-for="t in convTypes" :key="t.value">
        <button class="conv-seg-btn" :class="convType === t.value ? 'active' : ''" @click="setConvType(t.value)"
          type="button" x-text="t.label"></button>
      </template>
    </div>

    <button class="new-chat-btn" type="button" title="New chat"
      :disabled="isRunning || !selectedMateId || messages.length === 0 || convType !== 'chat'" @click="newChat()">
      <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 5v14M5 12h14" />
      </svg>
    </button>
  </div>

  <!-- ── Messages ─────────────────────────────────────────────── -->
  <div class="chat-scroll" id="chat-scroll">

    <template x-if="messages.length === 0">
      <div class="chat-empty">
        <template x-if="!selectedMate">
          <div class="chat-empty-text">Select a mate from the sidebar to start</div>
        </template>
        <template x-if="selectedMate">
          <div class="chat-empty-card">
            <div class="chat-empty-name" x-text="selectedMate.name"></div>
            <div class="chat-empty-desc">Ask anything or assign a task</div>
            <div class="chat-empty-hints">
              <span class="chat-empty-hint-key">Insert file</span>
              <div class="chat-empty-hint-val">
                <kbd x-text="isMac ? '⌘' : 'Ctrl'"></kbd><kbd>K</kbd>
                <span>search</span>
                <span class="chat-empty-hint-arrow">→</span>
                <kbd x-text="isMac ? '⌘' : 'Ctrl'"></kbd><kbd>↵</kbd>
                <span>insert path</span>
              </div>
              <span class="chat-empty-hint-key">Insert dir</span>
              <div class="chat-empty-hint-val">
                <kbd>/</kbd>
                <span>autocomplete</span>
              </div>
            </div>
          </div>
        </template>
      </div>
    </template>

    <template x-for="(msg, i) in messages" :key="i">
      <div :class="'msg-row msg-row-' + msg.role">

        <template x-if="msg.role === 'user'">
          <div class="msg-user-wrap">
            <div class="msg-user-bubble" x-text="msg.content"></div>
          </div>
        </template>

        <template x-if="msg.role === 'assistant'">
          <div class="msg-assistant-wrap">
            <div class="msg-agent-header">
              <div class="msg-agent-avatar"
                :style="getMateColor(msg.mateId) ? 'background:' + getMateColor(msg.mateId) : ''"
                x-text="mateInitials(getMateNameForMsg(msg))"></div>
              <div class="msg-agent-name" x-text="getMateNameForMsg(msg)"></div>
              <template x-if="msg.triggerEvent">
                <span class="msg-trigger-badge" x-text="triggerEventLabel(msg.triggerEvent)"></span>
              </template>
              <template x-if="msg._fmtTime">
                <span class="msg-agent-time" x-text="msg._fmtTime"></span>
              </template>
            </div>
            <div class="msg-body">
              <template x-for="(seg, j) in msg.segments" :key="j">
                <div>
                  <template x-if="seg.type === 'text'">
                    <div class="prose" x-html="renderMarkdown(seg.content)"></div>
                  </template>
                  <template x-if="seg.type === 'thinking' && seg.content">
                    <div class="seg-thinking">
                      <button class="seg-thinking-toggle"
                        :class="{'open': seg.open, 'is-streaming': msg.status === 'running' && isLastThinkingInMsg(msg, j)}"
                        @click="seg.open = !seg.open" type="button">
                        <svg fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M9 18l6-6-6-6" />
                        </svg>
                        Thinking
                      </button>
                      <div class="seg-thinking-content" x-show="seg.open" x-text="seg.content"></div>
                    </div>
                  </template>
                  <template x-if="seg.type === 'tool_use'">
                    <div>
                      <template x-if="seg.results.length === 0">
                        <div class="seg-tool-use">
                          <span :class="msg.status === 'running' ? 'seg-tool-spinner' : 'seg-tool-icon-done'"></span>
                          <span x-text="seg.name"></span>
                        </div>
                      </template>
                      <template x-if="seg.results.length > 0">
                        <div class="seg-tool-result">
                          <button class="seg-tool-result-toggle" :class="seg.open ? 'open' : ''"
                            @click="seg.open = !seg.open" type="button">
                            <svg fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" d="M9 18l6-6-6-6" />
                            </svg>
                            <span x-text="seg.count > 1 ? seg.name + ' ×' + seg.count : seg.name"></span>
                          </button>
                          <template x-if="seg.open">
                            <div>
                              <template x-for="(r, ri) in seg.results" :key="ri">
                                <pre class="seg-tool-result-content"
                                  :style="seg.results.length > 1 ? 'margin-top:0.25rem' : ''"
                                  x-text="seg.results.length > 1 ? '[' + (ri + 1) + '] ' + r : r"></pre>
                              </template>
                            </div>
                          </template>
                        </div>
                      </template>
                    </div>
                  </template>
                  <template x-if="seg.type === 'tool_result'">
                    <div class="seg-tool-result">
                      <button class="seg-tool-result-toggle" :class="seg.open ? 'open' : ''"
                        @click="seg.open = !seg.open" type="button">
                        <svg fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M9 18l6-6-6-6" />
                        </svg>
                        Result
                      </button>
                      <pre class="seg-tool-result-content" x-show="seg.open" x-text="seg.content"></pre>
                    </div>
                  </template>
                  <template x-if="seg.type === 'status' && msg.status === 'running'">
                    <div class="seg-status is-running" x-text="seg.label"></div>
                  </template>
                  <template x-if="seg.type === 'console'">
                    <pre class="seg-console" x-text="seg.content"></pre>
                  </template>
                  <template x-if="seg.type === 'error'">
                    <div class="seg-error" x-text="seg.message"></div>
                  </template>
                </div>
              </template>
            </div>
            <div class="msg-footer">
              <template
                x-if="msg.status === 'running' && !msg.segments.some(function(s){ return (s.type==='text' && s.content) || s.type==='status'; })">
                <span class="msg-thinking-label">Thinking<span
                    class="thinking-dots"><span></span><span></span><span></span></span></span>
              </template>
              <template
                x-if="msg.status !== 'running' && msg.segments.some(function(s){ return s.type==='text' && s.content; })">
                <button class="msg-copy-btn" type="button" title="Copy reply" :class="msg.copied ? 'copied' : ''"
                  @click="copyMateText(msg)">
                  <template x-if="!msg.copied">
                    <svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
                      <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
                      <path stroke-linecap="round" stroke-linejoin="round"
                        d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
                    </svg>
                  </template>
                  <template x-if="msg.copied">
                    <svg fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M20 6L9 17l-5-5" />
                    </svg>
                  </template>
                </button>
              </template>
              <template x-if="msg.status === 'failed'">
                <span class="msg-failed">Failed</span>
              </template>
            </div>
          </div>
        </template>

      </div>
    </template>

  </div><!-- .chat-scroll -->

  <!-- ── Input (chat only) ────────────────────────────────────── -->
  <div class="chat-input-outer" x-show="convType === 'chat'">
    <div class="chat-ac-wrap">
      <ul class="chat-path-ac" id="chat-path-ac" role="listbox" aria-expanded="false" hidden></ul>
      <div class="chat-input-card">
        <div class="chat-input-hint" x-show="!inputText" aria-hidden="true">
          <span class="chat-hint-name"
            x-text="selectedMate ? 'Ask ' + selectedMate.name + '…' : 'Select a mate from the sidebar'"></span>
        </div>
        <textarea class="chat-textarea" id="chat-textarea" rows="1" placeholder="" x-model="inputText"
          :disabled="isRunning || !selectedMateId" @keydown="handleKeydown($event)"
          @input="autoResize($event.target)"></textarea>
        <div class="chat-input-bar">
          <span class="chat-shortcut-hint" x-text="isMac ? '⌘↵ send' : 'Ctrl+↵ send'"></span>
          <div class="chat-input-bar-right">
            <button x-show="!isRunning" type="button" class="chat-send-circle"
              :disabled="!inputText.trim() || !selectedMateId" @click="send()">
              <svg fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 19V5M5 12l7-7 7 7" />
              </svg>
            </button>
            <button x-show="isRunning" type="button" class="chat-stop-circle" @click="cancel()">
              <svg fill="currentColor" viewBox="0 0 24 24">
                <rect x="5" y="5" width="14" height="14" rx="2" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div><!-- .chat-ac-wrap -->
  </div>

</div>`

// homeChatToastHTML mirrors agent_chat.html's run-completion toast — a
// position:fixed element, so like homeImagesLightboxHTML/homeInboxSheetHTML
// it lives once at the page level rather than inside the swappable
// #home-list-pane: a background chat run can finish while a different
// sidebar section is showing, and the toast should still surface.
const homeChatToastHTML = `
  <div class="run-toast" :class="toastKind === 'ok' ? 'run-toast-ok' : 'run-toast-err'" x-show="toastVisible" x-transition
    style="display:none">
    <span x-text="toastText"></span>
    <button class="run-toast-dismiss" @click="toastVisible = false" type="button" title="Dismiss">
      <svg fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6" />
        <path stroke-linecap="round" stroke-linejoin="round" d="m6 6 12 12" />
      </svg>
    </button>
  </div>`

// sortDirsSystemLast stably sorts dirs so that underscore-prefixed system
// directories (_knowledge, _shorts, etc. — see dirHasUnderscoreView) sort
// after every regular directory, preserving each group's existing (dir-path
// ascending, from ListAllDirs) relative order.
func sortDirsSystemLast(dirs []storage.DirSummary) {
	sort.SliceStable(dirs, func(i, j int) bool {
		iSys, jSys := dirHasUnderscoreView(dirs[i].Dir), dirHasUnderscoreView(dirs[j].Dir)
		return iSys != jSys && !iSys
	})
}

func noteItemsFromNotes(notes []storage.Note) []noteItem {
	items := make([]noteItem, 0, len(notes))
	for _, n := range notes {
		items = append(items, noteToItem(n))
	}
	return items
}

var homePageHTML = `<!DOCTYPE html>
<html lang="en">
` + headHTML(headOpts{title: "Home — Vaultr", withFonts: true, withTW: true, withAlpine: true, withHTMX: true}) + `
  <script src="/static/vendor/cytoscape.min.js"></script>
  <script src="/static/vendor/layout-base.js"></script>
  <script src="/static/vendor/cose-base.js"></script>
  <script src="/static/vendor/cytoscape-fcose.min.js"></script>
  <script src="/static/vendor/marked.min.js"></script>
  <script src="/static/vendor/dompurify.min.js"></script>
  <script>
  if (typeof marked !== 'undefined') { marked.setOptions({ gfm: true, breaks: true }); }
  </script>
  <style>
` + appTokensCSS + `
` + infoDialogCSS + navCSS + neoCSS + topbarCSS + homeCSS + shortsCSS + imagesCSS + graphCSS + agentChatCSS + drawerCSS + noteSharedCSS + noteEditorCSS + searchOverlayStyles + confirmDialogCSS + shortDialogCSS + settingsModalCSS + `
  </style>
</head>
<body x-data="homeCtrl()" @vaultr:insert-path.window="insertPath($event)">
` + searchOnlyOverlayHTML + confirmDialogHTML + infoDialogHTML + shortDialogHTML + settingsModalHTML() + homeImagesLightboxHTML + homeInboxSheetHTML + homeChatToastHTML + `
  <header class="lib-topbar">
    <div class="lib-topbar-spacer"></div>
` + topbarActionsHTML("refresh()", "Refresh home", "") + `
  </header>

  <div class="lib-body">
` + navHTML() + homeMainHTML + `
  </div>

` + drawerHTML + `
  <div id="graph-tooltip" class="graph-tooltip"></div>

  <script>
  document.addEventListener('alpine:init', () => {
` + alpineStoresScript + `
  });

` + keysJS + pathAcScript + drawerScript + searchOverlayScript + confirmDialogJS + infoDialogJS + shortDialogJS + settingsCtrlJS + homeJS + `
  </script>
` + noteSharedJS + `
</body>
</html>`

var homePageTemplate = template.Must(template.New("home").Funcs(homeTemplateFuncs).Parse(homePageHTML))

// HomeRefresh handles GET /home/refresh — returns HTMX OOB fragments for the
// sidebar's counts + folder list, without a full page reload.
// It does not touch the active note list — the client re-requests that
// separately via homeCtrl.reloadActiveSection (see home.js), since only the
// browser knows which sidebar item is currently selected.
func (vh *ViewHandler) HomeRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/home") {
		return
	}

	folders, err := vh.vault.ListAllDirs()
	if err != nil {
		http.Error(w, "home/refresh: list dirs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sortDirsSystemLast(folders)

	indexNotes := vh.listIndexItems()

	data := homePageData{
		Folders:    folders,
		IndexNotes: indexNotes,
	}

	var buf bytes.Buffer
	if err := homeRefreshTemplate.Execute(&buf, data); err != nil {
		http.Error(w, fmt.Sprintf("home/refresh: render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

var homeRefreshTemplate = template.Must(template.New("home-refresh").Funcs(homeTemplateFuncs).Parse(`<div id="home-side-folders-body" class="home-side-children" x-show="foldersOpen" x-cloak hx-swap-oob="true">
  {{range .Folders}}
  <button type="button" class="home-side-child" :class="{'is-active': activeKey === 'dir:{{.Dir}}'}"
          @click="selectFolder('{{.Dir}}','/home/section?type=folder&path={{encdir .Dir}}')">
    <svg class="home-side-child-icon" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"
         stroke-linejoin="round" viewBox="0 0 24 24">
      <path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z" />
    </svg>
    <span class="home-side-child-name">{{folderLabel .Dir}}</span>
  </button>
  {{end}}
  {{if not .Folders}}<div class="home-side-empty">No folders</div>{{end}}
</div>
<div id="home-side-graph-body" class="home-side-children" x-show="graphOpen" x-cloak hx-swap-oob="true">
  {{range .IndexNotes}}
  <button type="button" class="home-side-child" :class="{'is-active': activeKey === 'graph:{{.Path}}'}"
          @click="selectGraphIndex('{{.Path}}')">
    <span class="home-side-child-name">{{label .}}</span>
  </button>
  {{end}}
</div>
`))
