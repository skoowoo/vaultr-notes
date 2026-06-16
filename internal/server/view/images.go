package view

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/hardhacker/vaultr/internal/storage"
	"github.com/hardhacker/vaultr/internal/util"
)

const imagesPageSize = 30

type imageItem struct {
	Dir             string
	Name            string
	Ext             string
	Size            string
	UpdatedAt       string
	ThumbURL        string
	CursorNs        int64
	LinkedNotes     []string
	LinkedNotesJSON string // JSON array for data attribute, e.g. ["note1","note2"]
	// Caption is linked note titles joined for the card (2-line clamp in CSS); empty → use Name in template.
	Caption string
}

type imagesPageData struct {
	Images      []imageItem
	ImagesCount int
	NextNs      int64
}

type imagesGridData struct {
	Images []imageItem
	NextNs int64
}

func imageItemFrom(img storage.Image) imageItem {
	notes := img.LinkedNotes
	if notes == nil {
		notes = []string{}
	}
	notesJSON, _ := json.Marshal(notes)
	var captionParts []string
	for _, n := range notes {
		n = strings.TrimSpace(n)
		if n != "" {
			captionParts = append(captionParts, n)
		}
	}
	caption := strings.Join(captionParts, " · ")
	return imageItem{
		Dir:             img.Dir,
		Name:            img.Name,
		Ext:             img.Ext,
		Size:            util.FormatSize(img.Size),
		UpdatedAt:       formatRelativeTime(img.UpdatedAt),
		ThumbURL:        imageThumbURL(img),
		CursorNs:        img.UpdatedAt.UnixNano(),
		LinkedNotes:     notes,
		LinkedNotesJSON: string(notesJSON),
		Caption:         caption,
	}
}

func imageThumbURL(img storage.Image) string {
	if strings.HasPrefix(img.Dir, "/_assets") {
		return img.Dir + "/" + url.PathEscape(img.Name)
	}
	q := url.Values{}
	q.Set("dir", img.Dir)
	q.Set("name", img.Name)
	return "/api/images/at?" + q.Encode()
}

// Images handles GET /images — the full gallery page.
func (vh *ViewHandler) Images(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	imgs, err := vh.vault.ListImages(0, imagesPageSize+1)
	if err != nil {
		http.Error(w, "images: "+err.Error(), http.StatusInternalServerError)
		return
	}

	count, _ := vh.vault.CountImages()

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
	if err := imagesPageTemplate.Execute(&buf, imagesPageData{
		Images:      items,
		ImagesCount: count,
		NextNs:      nextNs,
	}); err != nil {
		http.Error(w, fmt.Sprintf("render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

// ImagesGrid handles GET /images/grid?before=NS — HTMX fragment for scroll pagination.
func (vh *ViewHandler) ImagesGrid(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/images") {
		return
	}

	var beforeNs int64
	if s := r.URL.Query().Get("before"); s != "" {
		beforeNs, _ = strconv.ParseInt(s, 10, 64)
	}

	imgs, err := vh.vault.ListImages(beforeNs, imagesPageSize+1)
	if err != nil {
		http.Error(w, "images grid: "+err.Error(), http.StatusInternalServerError)
		return
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := imagesGridTemplate.Execute(w, imagesGridData{Images: items, NextNs: nextNs}); err != nil {
		http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── templates ─────────────────────────────────────────────────────────────────

var imagesGridTemplate = template.Must(template.New("imagesgrid").Parse(imagesGridHTML))

var imagesPageHTML = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
` + headHTML(headOpts{title: "Images — Vaultr", withFonts: true, withTW: true, withAlpine: true, withHTMX: true}) + `  <style>
    :root {` + appTokensDark + `}
    html[data-theme="light"] {` + appTokensLight + `}
` + navCSS + pixelCSS + topbarCSS + imagesCSS + drawerCSS + noteSharedCSS + noteEditorCSS + searchOverlayStyles + confirmDialogCSS + shortDialogCSS + settingsModalCSS + `
  </style>
</head>
<body x-data="imgCtrl()" :class="{ 'select-mode': selectMode }">
` + searchOnlyOverlayHTML + confirmDialogHTML + shortDialogHTML + settingsModalHTML() + `

  <!-- ── Lightbox ──────────────────────────────────────────────────── -->
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

      <!-- Left: image -->
      <div class="lb-viewer">
        <img class="lb-viewer-img"
             :src="lightbox ? lightbox.src : ''"
             :alt="lightbox ? lightbox.name : ''"
             draggable="false">
      </div>

      <!-- Right: sidebar -->
      <aside class="lb-sidebar">
        <div class="lb-sidebar-head">
          <span class="lb-sidebar-title" x-text="lightbox ? lightbox.name : ''"></span>
          <button type="button" class="lb-delete-btn" title="Delete image"
                  x-show="lightbox"
                  @click="deleteLightboxImage()">
            <svg fill="none" stroke="currentColor" stroke-width="1.9" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M3 6h18"/><path stroke-linecap="round" stroke-linejoin="round" d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><path stroke-linecap="round" stroke-linejoin="round" d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"/><path stroke-linecap="round" stroke-linejoin="round" d="M10 11v6"/><path stroke-linecap="round" stroke-linejoin="round" d="M14 11v6"/></svg>
          </button>
          <button type="button" class="lb-close-btn" @click="closeLightbox()" title="Close (Esc)">
            <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" d="M18 6 6 18"/><path stroke-linecap="round" d="m6 6 12 12"/></svg>
          </button>
        </div>

        <div class="lb-sidebar-body">

          <!-- Linked notes — shown prominently first -->
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

          <!-- Basic info -->
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
  </div>

  <header class="lib-topbar">
    <div class="lib-topbar-left">
      <a href="/home" class="lib-back-btn" title="Back to Home">` + topbarIconBack + `<span class="lib-back-label">back</span></a>
    </div>
    <div class="lib-topbar-spacer"></div>
    <div class="lib-topbar-actions">
      ` + shortTriggerButton + `
      <button type="button" class="lib-action-btn" title="Refresh" @click="refresh()">` + topbarIconReload + `</button>
      <button type="button" class="lib-action-btn"
              :class="{ 'is-active': drawerOpen }"
              :title="/Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent) ? 'Reading drawer (⌘E)' : 'Reading drawer (Ctrl+E)'"
              @click="drawerOpen = !drawerOpen">` + topbarIconPanel + `</button>
      <button type="button" class="lib-action-btn"
              :title="/Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent) ? 'Search (⌘K)' : 'Search (Ctrl+K)'"
              @click="window.dispatchEvent(new CustomEvent('open-search'))">` + topbarIconSearch + `</button>
    </div>
  </header>

  <div class="img-body">
` + navHTML("images") + imagesMainHTML + `
  </div>

` + drawerHTML + `

  <script>
  document.addEventListener('alpine:init', () => {
` + themeStoreScript + `
  });

` + keysJS + drawerScript + searchOverlayScript + confirmDialogJS + shortDialogJS + settingsCtrlJS + imagesJS + `
  </script>
` + noteSharedJS + `
</body>
</html>
`

var imagesPageTemplate = template.Must(template.New("images").Parse(imagesPageHTML))
