package view

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// Inbox handles GET /agent/inbox — a sub-page of the agent chat page (like
// /graph is a sub-page of /home): no nav-bar icon of its own, reached via the
// inbox entry point in agent_chat.go's topbar, and nested under /agent so the
// Electron shell's per-section WebContentsView routing treats it as part of
// the same view. Stage one only has one producer (mate trigger runs), but the
// page itself renders whatever /api/inbox returns without assuming a source.
func (vh *ViewHandler) Inbox(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var buf bytes.Buffer
	if err := inboxTemplate.Execute(&buf, nil); err != nil {
		http.Error(w, fmt.Sprintf("render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

// ── template ──────────────────────────────────────────────────────────────────

const inboxCSS = `
    /* ── Inbox ───────────────────────────────────────────────── */
    *, *::before, *::after { box-sizing: border-box; }
    html, body { height: 100%; margin: 0; overflow: hidden; }
    body {
      display: flex; flex-direction: column; height: 100vh; overflow: hidden;
      background: var(--bg); color: var(--fg);
      font-family: var(--font-ui);
      -webkit-font-smoothing: antialiased;
    }
    .lib-body { flex: 1; display: flex; min-height: 0; overflow: hidden; }
    .inbox-body { flex: 1; display: flex; min-height: 0; overflow: hidden; }
    .inbox-columns { flex: 1; display: flex; min-height: 0; overflow: hidden; }

    /* ── Sidebar — message list, mirrors .graph-index-col ──────── */
    .inbox-sidebar {
      flex: 0 0 320px; max-width: 320px;
      display: flex; flex-direction: column;
      border-right: 2px solid var(--hr);
      min-height: 0; background: var(--surface-soft);
    }
    .inbox-sidebar-head {
      flex-shrink: 0; min-height: 54px; display: flex; align-items: center; gap: 0.6rem;
      padding: 0.6rem 0.75rem; border-bottom: 2px solid var(--hr); background: var(--surface-soft);
    }
    .inbox-sidebar-title { flex-shrink: 0; font-size: var(--text-sm); font-weight: var(--fw-semibold); color: var(--fg); opacity: 0.82; }
    .inbox-sidebar-count {
      flex-shrink: 0; font-size: var(--text-xs); font-weight: var(--fw-medium); color: var(--cnt-tx);
      background: var(--cnt-bg); padding: 0.07em 0.4em; font-variant-numeric: tabular-nums;
    }
    .inbox-sidebar-spacer { flex: 1; }
    .inbox-mark-all-btn {
      flex-shrink: 0; width: 24px; height: 24px; display: flex; align-items: center; justify-content: center;
      border: none; background: transparent; color: var(--muted); cursor: pointer; padding: 0;
    }
    .inbox-mark-all-btn:hover { color: var(--fg); background: var(--icon-hov); }
    .inbox-mark-all-btn svg { width: 13px; height: 13px; flex-shrink: 0; }

    /* ── Unread / Read filter — mirrors agent_chat.go's .conv-seg (§6.7 segmented control) ── */
    .inbox-filter-seg { flex-shrink: 0; display: inline-flex; background: transparent; border: 2px solid var(--card-bd); padding: 2px; gap: 2px; }
    .inbox-filter-seg-btn {
      display: flex; align-items: center; padding: 0.3rem 0.6rem; border: none; background: transparent;
      font-size: var(--text-sm); font-weight: var(--fw-medium); color: var(--muted); cursor: pointer; white-space: nowrap;
    }
    .inbox-filter-seg-btn:hover { color: var(--fg); background: var(--icon-hov); }
    .inbox-filter-seg-btn.active { background: var(--seg-act-bg); color: var(--seg-act-fg); box-shadow: none; }

    .inbox-list { flex: 1; overflow-y: auto; padding: 0.75rem; display: flex; flex-direction: column; gap: 0.6rem; }
    .inbox-empty { padding: 3rem 1rem; text-align: center; color: var(--muted); font-size: var(--text-sm); }
    .inbox-load-more { padding: 0.5rem 0; text-align: center; color: var(--muted); font-size: var(--text-2xs); }
    .inbox-sidebar-foot {
      flex-shrink: 0; padding: 0.55rem 0.75rem; border-top: 2px solid var(--hr);
      background: var(--surface-soft); color: var(--muted-soft); font-size: var(--text-2xs); text-align: center;
    }

    /* Neo-brutalist message card (mirrors .folder-card): black border always
       on, white fill at rest, no shadow until hover, shadow + yellow fill
       when selected/active. */
    .inbox-row {
      display: flex; align-items: flex-start; gap: 0.5rem;
      padding: 0.9rem 1rem; cursor: pointer;
      border: 2px solid var(--card-bd); background: var(--canvas);
    }
    .inbox-row:hover {
      box-shadow: var(--px-d2) var(--px-shadow);
    }
    .inbox-row.is-selected {
      background: var(--accent); color: var(--fg);
      box-shadow: var(--px-d2) var(--px-shadow);
    }

    /* Dot sits in its own column so the title and snippet below share one left edge */
    .inbox-unread-dot { flex-shrink: 0; width: 7px; height: 7px; margin-top: 0.4rem; background: var(--fg); }
    .inbox-row:not(.is-unread) .inbox-unread-dot { display: none; }
    .inbox-row-content { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.45rem; }
    .inbox-row-head { display: flex; align-items: center; gap: 0.4rem; min-width: 0; }
    .inbox-title {
      flex: 1; min-width: 0; font-size: var(--text-sm); color: var(--fg);
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
    }
    .inbox-row.is-unread .inbox-title { font-weight: var(--fw-semibold); }
    .inbox-time { flex-shrink: 0; font-size: var(--text-2xs); color: var(--muted-soft); font-variant-numeric: tabular-nums; }
    .inbox-row.is-selected .inbox-time { color: var(--fg); opacity: 0.65; }
    .inbox-snippet {
      font-size: var(--text-xs); color: var(--muted); line-height: var(--lh-relaxed);
      overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
    }
    .inbox-row.is-selected .inbox-snippet { color: var(--fg); opacity: 0.75; }

    /* ── Reader pane — right side ───────────────────────────────── */
    .inbox-reader { flex: 1; min-width: 0; display: flex; flex-direction: column; overflow: hidden; background: var(--bg); }
    .inbox-reader-empty {
      flex: 1; display: flex; align-items: center; justify-content: center;
      color: var(--muted); font-size: var(--text-sm);
    }
    .inbox-reader-scroll {
      flex: 1; overflow-y: auto; padding: 1.75rem 2.25rem 3rem;
      scrollbar-width: thin; scrollbar-color: var(--scrollbar-thumb) transparent;
    }
    .inbox-reader-scroll::-webkit-scrollbar { width: 5px; }
    .inbox-reader-scroll::-webkit-scrollbar-track { background: transparent; }
    .inbox-reader-scroll::-webkit-scrollbar-thumb { background: var(--scrollbar-thumb); }
    .inbox-reader-scroll::-webkit-scrollbar-thumb:hover { background: var(--scrollbar-thumb-hov); }
    .inbox-reader-head { display: flex; flex-direction: column; gap: 0.5rem; margin-bottom: 1.25rem; padding-bottom: 1.1rem; border-bottom: 2px solid var(--hr); }
    .inbox-reader-title { font-size: var(--text-title-lg); font-weight: var(--fw-semibold); color: var(--fg); }
    .inbox-reader-meta { display: flex; align-items: center; gap: 0.5rem; font-size: var(--text-xs); color: var(--muted); }

    .inbox-reader-body.prose { font-size: var(--text-base); }`

var inboxMainHTML = `  <main class="inbox-body">
    <div class="inbox-columns">
      <div class="inbox-sidebar">
        <div class="inbox-sidebar-head">
          <span class="inbox-sidebar-title">Inbox</span>
          <span class="inbox-sidebar-count" x-show="unreadCount > 0" x-text="unreadCount"></span>
          <div class="inbox-sidebar-spacer"></div>
          <div class="inbox-filter-seg">
            <button type="button" class="inbox-filter-seg-btn" :class="filter === 'all' ? 'active' : ''" @click="setFilter('all')">All</button>
            <button type="button" class="inbox-filter-seg-btn" :class="filter === 'unread' ? 'active' : ''" @click="setFilter('unread')">Unread</button>
            <button type="button" class="inbox-filter-seg-btn" :class="filter === 'read' ? 'active' : ''" @click="setFilter('read')">Read</button>
          </div>
          <button type="button" class="inbox-mark-all-btn" title="Mark all read" x-show="unreadCount > 0" @click="markAllRead()">` + svgCheck + `</button>
        </div>
        <div class="inbox-list" x-ref="list" @scroll="onListScroll($event)">
          <template x-if="!loading && visibleMessages().length === 0">
            <div class="inbox-empty" x-text="filter === 'unread' ? 'All caught up — no unread messages.' : (filter === 'read' ? 'No read messages yet.' : 'No messages yet.')"></div>
          </template>
          <template x-for="m in visibleMessages()" :key="m.id">
            <div class="inbox-row" :class="{ 'is-unread': !m.isRead, 'is-selected': selected && selected.id === m.id }" @click="select(m)">
              <div class="inbox-unread-dot"></div>
              <div class="inbox-row-content">
                <div class="inbox-row-head">
                  <span class="inbox-title" x-text="m.title || m.source"></span>
                  <span class="inbox-time" x-text="relTime(m.createdAt)"></span>
                </div>
                <div class="inbox-snippet" x-text="m.body"></div>
              </div>
            </div>
          </template>
          <div class="inbox-load-more" x-show="loadingMore">Loading more…</div>
        </div>
        <div class="inbox-sidebar-foot">Messages older than 30 days are removed automatically.</div>
      </div>

      <div class="inbox-reader">
        <template x-if="!selected">
          <div class="inbox-reader-empty">Select a message to read</div>
        </template>
        <template x-if="selected">
          <div class="inbox-reader-scroll">
            <div class="inbox-reader-head">
              <span class="inbox-reader-title" x-text="selected.title || selected.source"></span>
              <div class="inbox-reader-meta">
                <span x-text="selected.source"></span>
                <span>·</span>
                <span x-text="fullTime(selected.createdAt)"></span>
              </div>
            </div>
            <div class="prose inbox-reader-body" x-html="renderMarkdown(selected.body)"></div>
          </div>
        </template>
      </div>
    </div>
  </main>`

const inboxJS = `
  // Search overlay (search_overlay.go) falls back to a real navigation
  // (window.location.href = el.href) when this is undefined, and that href
  // points at /notes?name=…, which has no registered route — 404. Every
  // other page defines this to open the pick in the editor drawer instead;
  // mirrors home.js / agent_chat.js verbatim.
  window.handleSearchResultSelection = function(el) {
    if (!el || !el.dataset) return false;
    var focusURL = el.dataset.focusUrl || '';
    if (!focusURL) return false;
    try {
      var u = new URL(focusURL, window.location.origin);
      var p = u.searchParams.get('path');
      if (!p) return false;
      var nameEl = el.querySelector('.sr-name');
      var title = nameEl ? nameEl.textContent.trim() : (p.split('/').pop().replace(/\.md$/, '') || 'Note');
      if (window.__vaultrDrawer) {
        void window.__vaultrDrawer.openNoteInDrawer(p, title,
          el.dataset.noteIsKnowledge === 'true', false, el.dataset.noteIsIndex === 'true',
          el.dataset.noteCanCompile === 'true');
        return true;
      }
    } catch(e) { return false; }
    return false;
  };

  // Matches internal/inbox.defaultListLimit — the server-side page size used
  // when a request omits ?limit=. Kept in sync manually since the client
  // needs to know a full page was returned to decide whether more exist.
  var INBOX_PAGE_SIZE = 50;

  function inboxCtrl() {
    return Object.assign(drawerCtrl(), {
      messages: [],
      loading: true,
      loadingMore: false,
      hasMore: true,
      unreadCount: 0,
      selected: null,
      filter: 'all',
      init() {
        this.initDrawer();
        this.load();
        this.initStream();
      },
      // Subscribes to /api/inbox/notifications (internal/inbox.Bus) so a
      // message created while this page is open is appended live instead of
      // waiting for a manual refresh or a visibilitychange re-fetch.
      initStream() {
        if (typeof EventSource === 'undefined') return;
        const es = new EventSource('/api/inbox/notifications');
        es.addEventListener('message', (e) => {
          let msg;
          try { msg = JSON.parse(e.data); } catch (err) { return; }
          if (this.messages.some((m) => m.id === msg.id)) return;
          this.messages.unshift(msg);
          if (!msg.isRead) this.unreadCount++;
        });
      },
      visibleMessages() {
        if (this.filter === 'all') return this.messages;
        return this.messages.filter((m) => this.filter === 'unread' ? !m.isRead : m.isRead);
      },
      setFilter(f) {
        this.filter = f;
        // A filter can shrink the rendered list below the container's
        // height, which would never fire a scroll event — pull in more
        // pages until either the list overflows or the server runs dry.
        this.maybeFillList();
      },
      load() {
        this.loading = true;
        this.hasMore = true;
        return fetch('/api/inbox').then(function(r){return r.json();}).then((data) => {
          this.messages = data.messages || [];
          this.hasMore = this.messages.length >= INBOX_PAGE_SIZE;
          if (this.selected) {
            var fresh = this.messages.find((m) => m.id === this.selected.id);
            this.selected = fresh || this.selected;
          }
        }).catch(function(){}).then(() => {
          this.loading = false;
          this.refreshUnreadCount();
          this.maybeFillList();
        });
      },
      // Fetches the next page (offset = however many we already hold — see
      // note below) and appends it. Scroll-triggered, and also chained by
      // maybeFillList() when a page doesn't fill the viewport.
      loadMore() {
        if (this.loadingMore || !this.hasMore) return Promise.resolve();
        this.loadingMore = true;
        // Messages only ever grow at index 0 (SSE prepends new arrivals at
        // the top, matching their position in the server's created_at DESC
        // order), so the count already held is always the correct OFFSET
        // into that same order — no separate cursor needs tracking.
        var offset = this.messages.length;
        return fetch('/api/inbox?offset=' + offset).then(function(r){return r.json();}).then((data) => {
          var page = data.messages || [];
          var seen = new Set(this.messages.map((m) => m.id));
          page.forEach((m) => { if (!seen.has(m.id)) this.messages.push(m); });
          this.hasMore = page.length >= INBOX_PAGE_SIZE;
        }).catch(function(){}).then(() => {
          this.loadingMore = false;
        });
      },
      onListScroll(e) {
        if (this.loadingMore || !this.hasMore) return;
        var el = e.target;
        if (el.scrollTop + el.clientHeight >= el.scrollHeight - 120) {
          this.loadMore();
        }
      },
      async maybeFillList() {
        await this.$nextTick();
        var el = this.$refs.list;
        if (!el || !this.hasMore || this.loadingMore) return;
        if (el.scrollHeight <= el.clientHeight + 4) {
          await this.loadMore();
          this.maybeFillList();
        }
      },
      refreshUnreadCount() {
        return fetch('/api/inbox/unread-count').then(function(r){return r.json();}).then((data) => {
          this.unreadCount = data.count || 0;
        }).catch(function(){});
      },
      select(m) {
        this.selected = m;
        if (!m.isRead) this.markRead(m);
      },
      async markRead(m) {
        m.isRead = true;
        this.unreadCount = Math.max(0, this.unreadCount - 1);
        try { await fetch('/api/inbox/' + m.id + '/read', { method: 'POST' }); } catch (e) { /* ignore */ }
      },
      async markAllRead() {
        this.messages.forEach(m => { m.isRead = true; });
        this.unreadCount = 0;
        try { await fetch('/api/inbox/read-all', { method: 'POST' }); } catch (e) { /* ignore */ }
      },
      relTime(iso) {
        const d = new Date(iso);
        const diffSec = Math.floor((Date.now() - d.getTime()) / 1000);
        if (diffSec < 60) return 'now';
        if (diffSec < 3600) return Math.floor(diffSec / 60) + 'm';
        if (diffSec < 86400) return Math.floor(diffSec / 3600) + 'h';
        return Math.floor(diffSec / 86400) + 'd';
      },
      fullTime(iso) {
        try { return new Date(iso).toLocaleString(); } catch (e) { return iso; }
      },
      // Mirrors agent_chat.js renderMarkdown(): same wiki-link handling and
      // sanitize allowlist, so inbox message bodies render identically to
      // assistant chat replies.
      renderMarkdown(text) {
        if (!text || typeof text !== 'string') return '';
        var processed = text.replace(/\[\[([^\]\[|]+?)(?:\|([^\]\[]+?))?\]\]/g, function(_, target, display) {
          target = target.trim();
          display = (display || target).trim();
          var name = target.endsWith('.md') ? target : target + '.md';
          return '[' + display + '](/notes?name=' + encodeURIComponent(name) + ')';
        });
        processed = processed.replace(/(?<!~)~(?!~)/g, '\\~');
        if (typeof marked === 'undefined') {
          return processed.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
        }
        var html = marked.parse(processed);
        if (typeof DOMPurify !== 'undefined') {
          html = DOMPurify.sanitize(html, {
            ALLOWED_TAGS: ['p','br','strong','em','s','del','code','pre','h1','h2','h3','h4','h5','h6',
                           'ul','ol','li','blockquote','a','hr','table','thead','tbody','tr','th','td','img'],
            ALLOWED_ATTR: ['href','title','src','alt'],
          });
        }
        return html;
      },
    });
  }
  // Electron's per-section WebContentsView can reuse an already-loaded /agent/inbox
  // document (no fresh navigation) when swapped back into view. Page Visibility API
  // still fires in that case, so re-fetch whenever the page becomes visible to avoid
  // showing stale messages until a manual refresh.
  document.addEventListener('visibilitychange', function(){
    if (document.hidden || typeof Alpine === 'undefined') return;
    var data = Alpine.$data(document.body);
    if (data && typeof data.load === 'function') data.load();
  });`

var inboxHTML = `<!DOCTYPE html>
<html lang="en" data-theme="neo">
` + headHTML(headOpts{title: "Inbox — Vaultr", withFonts: true, withTW: true, withAlpine: true, withHTMX: true}) + `  <style>
` + appTokensCSS + `
` + navCSS + neoCSS + topbarCSS + inboxCSS + searchOverlayStyles + confirmDialogCSS + shortDialogCSS + drawerCSS + noteSharedCSS + noteEditorCSS + settingsModalCSS + `
  </style>
  <script src="/static/vendor/marked.min.js"></script>
  <script src="/static/vendor/dompurify.min.js"></script>
  <script>
  if (typeof marked !== 'undefined') { marked.setOptions({ gfm: true, breaks: true }); }
  </script>
</head>
<body x-data="inboxCtrl()">
` + searchOnlyOverlayHTML + confirmDialogHTML + shortDialogHTML + settingsModalHTML() + `

  <header class="lib-topbar">
    <div class="lib-topbar-spacer"></div>
` + topbarActionsHTML("window.location.reload()", "Refresh", "", "") + `
  </header>

  <div class="lib-body">
` + navHTML("inbox") + inboxMainHTML + `
  </div>

` + drawerHTML + `

  <script>
  document.addEventListener('alpine:init', () => {
` + alpineStoresScript + `
  });

` + keysJS + pathAcScript + drawerScript + searchOverlayScript + confirmDialogJS + shortDialogJS + settingsCtrlJS + inboxJS + `
  </script>
` + noteSharedJS + `
</body>
</html>
`

var inboxTemplate = template.Must(template.New("inbox").Parse(inboxHTML))
