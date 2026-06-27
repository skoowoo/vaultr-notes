package view

const searchOverlayStyles = `
    /* ── Search UI ─────────────────────────────────────── */
    .srch-panel {
      border-color: var(--srch-panel-bd) !important;
      border-width: 2px;
      box-shadow: var(--srch-panel-shadow);
    }
    .srch-row   { border-color: var(--srch-row-bd) !important; border-bottom-width: 2px; background: var(--srch-row-bg); }
    .srch-icon  { color: var(--srch-ic); transition: color 150ms; }
    .srch-row:focus-within .srch-icon { color: var(--ui-accent); }
    .srch-input {
      color: var(--fg);
      caret-color: var(--fg);
      font-size: var(--text-body);
      font-family: var(--font-ui);
      letter-spacing: 0;
    }
    .srch-input::placeholder { color: var(--srch-ph); }
    .srch-kbd {
      color: var(--srch-kbd-fg);
      border-color: var(--srch-kbd-bd);
      font-family: var(--font-ui);
      font-size: var(--text-2xs);
      letter-spacing: 0.03em;
      line-height: 1;
    }
    .srch-btn {
      color: var(--muted);
      border-color: var(--srch-panel-bd);
      font-family: var(--font-ui);
      box-shadow: var(--px-d1) var(--px-shadow);
    }
    .srch-btn:hover {
      color: var(--fg);
      border-color: var(--srch-panel-bd);
      background: var(--card-bg);
      box-shadow: var(--px-d0) var(--px-shadow);
      transform: translate(1px, 1px);
    }
    .srch-btn:active {
      box-shadow: none;
      transform: translate(2px, 2px);
    }
    .srch-shortcut {
      font-family: var(--font-ui);
      font-size: var(--text-2xs);
      letter-spacing: 0.04em;
      color: var(--muted);
    }
    #search-results a {
      position: relative;
      border: 2px solid transparent;
    }
    #search-results a::before {
      display: none;
    }
    #search-results a.is-active {
      background: var(--srch-av);
      border-color: var(--card-bd);
      box-shadow: var(--px-d1) var(--px-shadow);
    }
    .sr-name  { color: var(--fg); font-size: var(--text-base); font-weight: 500; letter-spacing: 0; }
    .sr-dir   { color: var(--sr-dir); font-size: var(--text-xs); }
    .sr-time  {
      color: var(--sr-tm);
      font-size: var(--text-xs);
      font-family: var(--font-mono);
      letter-spacing: 0.01em;
      font-variant-numeric: tabular-nums;
      min-width: 10ch;
      text-align: right;
      justify-self: end;
    }
    .sr-icon  { color: var(--sr-ic); transition: color 100ms; }
    #search-results a.is-active .sr-name,
    #search-results a.is-active .sr-dir,
    #search-results a.is-active .sr-time,
    #search-results a.is-active .sr-icon { color: var(--fg); }
    .sr-empty {
      color: var(--sr-em);
      font-size: var(--text-sm);
      font-style: italic;
      font-family: var(--font-ui);
    }
    .srch-footer {
      border-color: var(--hr) !important;
      border-top-width: 2px;
    }
    .srch-hint {
      font-size: var(--text-xs);
      font-family: var(--font-ui);
      color: var(--sr-tm);
      letter-spacing: 0.01em;
    }
    .srch-hint kbd {
      display: inline-flex; align-items: center; justify-content: center;
      font-family: var(--font-ui);
      font-size: var(--text-2xs);
      color: var(--srch-kbd-fg);
      border: 2px solid var(--srch-kbd-bd);
      padding: 1px 5px;
      line-height: 1.4;
    }
    #search-results::-webkit-scrollbar { display: none; }
    /* ── Preview panel ─────────────────────────────────────── */
    .srch-preview-pane {
      padding: 1.25rem 1.5rem 1.75rem;
    }
    .srch-preview-pane .frag-cover {
      margin-bottom: 1rem;
      padding-bottom: 0.75rem;
      border-bottom: 2px solid var(--hr);
    }
    .srch-preview-pane .frag-dir,
    .srch-preview-pane .cover-dir {
      font-size: var(--text-xs);
      font-weight: 500;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      color: var(--sr-dir);
      margin: 0 0 0.3rem;
    }
    .srch-preview-pane .frag-title {
      font-family: inherit;
      font-size: var(--text-title-lg);
      font-weight: 600;
      line-height: 1.3;
      letter-spacing: -0.014em;
      color: var(--fg);
      margin: 0;
    }
    .srch-preview-pane .frag-fm-wrap { display: none; }
    .srch-preview-pane .prose { font-size: var(--text-sm); }
    .srch-prev-hint {
      display: flex; height: 100%;
      align-items: center; justify-content: center;
      color: var(--sr-dir); font-size: var(--text-sm); font-style: italic;
    }
    .srch-prev-spin {
      padding: 1.25rem 1.5rem;
      color: var(--sr-dir); font-size: var(--text-sm);
    }
    /* ── Mode system ─────────────────────────────────────── */
    .srch-icon-area {
      cursor: pointer;
      padding: 3px 2px;
      min-width: 24px;
    }
    .srch-icon-area:hover { background: var(--icon-hov); }
    .srch-icon-area.has-mode .srch-icon { color: var(--fg); }
    .srch-mode-chip {
      font-size: var(--text-2xs);
      font-family: var(--font-ui);
      font-weight: 600;
      color: var(--fg);
      letter-spacing: 0.06em;
      text-transform: uppercase;
      white-space: nowrap;
      line-height: 1;
    }
    .srch-mode-chip::after {
      content: ' ›';
      font-size: var(--text-2xs);
      font-weight: 400;
      opacity: 0.7;
    }
    .srch-mode-menu { border-color: var(--hr) !important; border-bottom-width: 2px; }
    .srch-mode-item {
      color: var(--fg);
      font-family: var(--font-ui);
    }
    .srch-mode-item.is-active { background: var(--seg-act-bg); color: var(--seg-act-fg); }
    .srch-mode-name { font-size: var(--text-sm); font-weight: 500; min-width: 5rem; }
    .srch-mode-desc { font-size: var(--text-xs); color: var(--sr-dir); }
    .srch-mode-footer { border-color: var(--hr) !important; border-top-width: 2px; }
    .srch-hint-sep { color: var(--sr-tm); opacity: 0.4; }
    .srch-key {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--fg);
      opacity: 0.65;
      letter-spacing: 0;
    }
`

const searchOverlayPanelHTML = `
    <!-- Overlay panel -->
    <div x-show="show"
         x-cloak
         x-transition:enter="transition ease-out duration-200"
         x-transition:enter-start="opacity-0 -translate-y-2 scale-[0.97]"
         x-transition:enter-end="opacity-100 translate-y-0 scale-100"
         x-transition:leave="transition ease-in duration-120"
         x-transition:leave-start="opacity-100 translate-y-0 scale-100"
         x-transition:leave-end="opacity-0 -translate-y-1 scale-[0.98]"
         class="pointer-events-auto absolute left-1/2 -translate-x-1/2 w-[calc(100%-2rem)] max-w-[860px]"
         style="top:calc(50% - 284px)">

      <div class="srch-panel rounded-xl overflow-hidden border"
           style="background:var(--srch-bg)">

        <!-- Input row -->
        <div class="srch-row flex items-center gap-3 px-4 border-b" style="height:68px">
          <input type="hidden" id="srch-field" name="field" :value="mode.field">
          <input type="hidden" id="srch-kind"  name="kind"  :value="mode.kind">
          <button @click.prevent="openModeMenu()"
                  tabindex="-1"
                  :class="{'has-mode': mode.field || mode.kind}"
                  class="srch-icon-area shrink-0 flex items-center gap-1.5 justify-center">
            <svg class="srch-icon w-[18px] h-[18px] shrink-0"
                 fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <circle cx="11" cy="11" r="8"/><path stroke-linecap="round" d="m21 21-4.35-4.35"/>
            </svg>
            <span x-show="mode.field || mode.kind"
                  class="srch-mode-chip"
                  x-text="mode.label"></span>
          </button>
          <input type="text"
                 x-ref="input"
                 :placeholder="mode.placeholder"
                 @input="onInput($event)"
                 @keydown="onInputKeydown($event)"
                 hx-get="/notes/search"
                 hx-trigger="input changed delay:300ms, search"
                 hx-target="#search-results"
                 hx-swap="innerHTML"
                 hx-include="#srch-field,#srch-kind"
                 name="q"
                 autocomplete="off"
                 class="srch-input flex-1 bg-transparent outline-none">
          <span class="srch-hint flex items-center gap-1.5 shrink-0">
            <span x-show="!mode.field && !mode.kind && !hasQuery" class="flex items-center gap-1.5">
              <kbd>/</kbd> mode
              <span class="srch-hint-sep">·</span>
            </span>
            <kbd>esc</kbd> close
          </span>
        </div>

        <!-- Mode selector -->
        <div x-show="showModes"
             x-transition:enter="transition ease-out duration-100"
             x-transition:enter-start="opacity-0"
             x-transition:enter-end="opacity-100"
             x-transition:leave="transition ease-in duration-100"
             x-transition:leave-start="opacity-100"
             x-transition:leave-end="opacity-0"
             class="srch-mode-menu border-b">
          <div class="flex flex-col gap-0.5 p-2" @mousemove="onModeHover($event)">
            <template x-for="(m, i) in modes" :key="m.key">
              <button @click="selectMode(m)"
                      :class="{'is-active': i === modeIdx}"
                      class="srch-mode-item flex items-center gap-3 px-3 py-2 rounded-lg w-full text-left">
                <span class="srch-mode-name" x-text="m.label"></span>
                <span class="srch-mode-desc flex-1" x-text="m.desc"></span>
                <span x-show="m.key === mode.key" class="ml-auto shrink-0" style="color:var(--ui-accent);font-size:var(--text-xs)">✓</span>
              </button>
            </template>
          </div>
          <div class="srch-mode-footer flex items-center gap-4 px-4 border-t" style="height:var(--btn-h-sm)">
            <span class="srch-hint flex items-center gap-1.5"><kbd>↑</kbd><kbd>↓</kbd> navigate</span>
            <span class="srch-hint flex items-center gap-1.5"><kbd>↵</kbd> select</span>
            <span class="srch-hint flex items-center gap-1.5"><kbd>esc</kbd> cancel</span>
          </div>
        </div>

        <!-- Body: two-column split (results left, preview right) — shown only when there is a query -->
        <div class="flex overflow-hidden"
             x-show="hasQuery"
             x-transition:enter="transition ease-out duration-150"
             x-transition:enter-start="opacity-0"
             x-transition:enter-end="opacity-100"
             style="height:480px">

          <!-- Left: results list -->
          <div id="search-results" @click="onResultClick($event)" @mousemove="onResultHover($event)"
               class="flex flex-col gap-0.5 p-2 overflow-y-auto flex-shrink-0 [scrollbar-width:none]"
               style="width:300px; border-right:2px solid var(--hr); padding:0.5rem; gap:0"></div>

          <!-- Right: note preview -->
          <div id="search-preview"
               class="flex-1 overflow-y-auto
                      [scrollbar-width:thin] [scrollbar-color:var(--card-bd)_transparent]">
            <div class="srch-prev-hint">Navigate results to preview</div>
          </div>
        </div>

        <!-- Footer hints — shown only when body is visible -->
        <div class="srch-footer flex items-center gap-4 px-4 border-t" x-show="hasQuery" style="height:var(--btn-h-sm)">
          <span class="srch-hint flex items-center gap-1.5">
            <span class="srch-key">↑↓</span> navigate
          </span>
          <span class="srch-hint flex items-center gap-1.5">
            <span class="srch-key">↵</span> open
          </span>
          <span class="srch-hint flex items-center gap-1.5">
            <span class="srch-key" x-text="isMac ? '⌘↵' : 'Ctrl+↵'"></span> copy path
          </span>
        </div>
      </div>
    </div>

    <!-- Backdrop -->
    <div x-show="show" x-cloak @click="close()"
         x-transition:enter="transition-opacity ease-out duration-200"
         x-transition:enter-start="opacity-0"
         x-transition:enter-end="opacity-100"
         x-transition:leave="transition-opacity ease-in duration-150"
         x-transition:leave-start="opacity-100"
         x-transition:leave-end="opacity-0"
         :class="show ? 'pointer-events-auto' : 'pointer-events-none'"
         class="fixed inset-0 -z-10"
         style="background:var(--srch-backdrop)"></div>
`

const searchOnlyOverlayHTML = `
  <!-- Search -->
  <div x-data="searchOverlay()"
       tabindex="-1"
       @keydown.window="onKey($event)"
       @open-search.window="openWithMode($event.detail)"
       @htmx:after-swap.window="onResultsSwap($event)"
       class="fixed inset-0 z-[10000] pointer-events-none outline-none">
` + searchOverlayPanelHTML + `
  </div>
`

const searchOverlayScript = `
  function searchOverlay() {
    return {
      show: false,
      hasQuery: false,
      activeIdx: -1,
      _previewCtrl: null,
      isMac: /Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent),
      modes: [
        { key: 'all',       label: 'All',       desc: 'Search everywhere',              field: '',        kind: '',          placeholder: 'Search notes…' },
        { key: 'name',      label: 'Name',      desc: 'Search by filename',             field: 'name',    kind: '',          placeholder: 'Search by filename…' },
        { key: 'content',   label: 'Content',   desc: 'Search in file content',         field: 'content', kind: '',          placeholder: 'Search in content…' },
        { key: 'tag',       label: 'Tag',       desc: 'Search by tag',                  field: 'tag',     kind: '',          placeholder: 'Search by tag…' },
        { key: 'knowledge', label: 'Knowledge', desc: 'Knowledge notes only',           field: '',        kind: 'knowledge', placeholder: 'Search knowledge notes…' },
        { key: 'short',     label: 'Short',     desc: 'Short notes only',               field: '',        kind: 'short',     placeholder: 'Search short notes…' },
        { key: 'raw',       label: 'Raw',       desc: 'Raw notes only',                 field: '',        kind: 'raw',       placeholder: 'Search raw notes…' },
        { key: 'index',     label: 'Index',     desc: 'Index notes only',               field: '',        kind: 'index',     placeholder: 'Search index notes…' },
      ],
      mode: { key: 'all', label: 'All', desc: 'Search everywhere', field: '', kind: '', placeholder: 'Search notes…' },
      showModes: false,
      modeIdx: 0,
      init() {
        this.mode = this.modes[0];
        window.__vaultrHotkeys.register('search', 'k', () => {
          this.show ? this.close() : this.open();
        });
      },
      _items() { return Array.from(document.querySelectorAll('#search-results a')); },
      _highlight(items) {
        items.forEach((el, i) => el.classList.toggle('is-active', i === this.activeIdx));
        if (this.activeIdx >= 0) {
          items[this.activeIdx].scrollIntoView({ block: 'nearest' });
          this._loadPreview(items[this.activeIdx]);
        }
      },
      _loadPreview(el) {
        const path = el && el.dataset.previewPath;
        if (!path) return;
        if (this._previewCtrl) this._previewCtrl.abort();
        this._previewCtrl = new AbortController();
        const panel = document.getElementById('search-preview');
        if (!panel) return;
        panel.innerHTML = '<div class="srch-prev-spin">…</div>';
        fetch('/notes/fragment?path=' + encodeURIComponent(path), { signal: this._previewCtrl.signal, headers: { 'HX-Request': 'true' } })
          .then(r => r.ok ? r.text() : Promise.reject())
          .then(html => { panel.innerHTML = '<div class="srch-preview-pane">' + html + '</div>'; })
          .catch(() => {});
      },
      _clearPreview() {
        if (this._previewCtrl) { this._previewCtrl.abort(); this._previewCtrl = null; }
        const p = document.getElementById('search-preview');
        if (p) p.innerHTML = '<div class="srch-prev-hint">Navigate results to preview</div>';
      },
      _clearQuery() {
        if (this.$refs.input) this.$refs.input.value = '';
        this.hasQuery = false;
        this.activeIdx = -1;
        this._clearPreview();
        const sr = document.getElementById('search-results');
        if (sr) sr.innerHTML = '';
        if (this.$refs.input) this.$refs.input.focus();
      },
      open() {
        this.show = true;
        window.__vaultrSearchOpen = true;
        this.activeIdx = -1;
        if (window.__vaultrEscPush) window.__vaultrEscPush('search', () => {
          if (this.showModes) { this.closeModeMenu(); }
          else if (this.hasQuery) { this._clearQuery(); }
          else { this.close(); }
        });
        this.$nextTick(() => this.$refs.input && this.$refs.input.focus());
      },
      openWithMode(detail) {
        this.open();
        if (detail && detail.mode) {
          const m = this.modes.find(x => x.key === detail.mode);
          if (m) this.$nextTick(() => this.selectMode(m));
        }
      },
      close() {
        if (window.__vaultrEscPop) window.__vaultrEscPop('search');
        this.show = false;
        window.__vaultrSearchOpen = false;
        this.activeIdx = -1;
        if (this._previewCtrl) { this._previewCtrl.abort(); this._previewCtrl = null; }
        setTimeout(() => {
          if (this.show) return;
          this.hasQuery = false;
          this.showModes = false;
          this.mode = this.modes[0];
          if (this.$refs.input) {
            this.$refs.input.value = '';
            const sr = document.getElementById('search-results');
            if (sr) sr.innerHTML = '';
          }
          const p = document.getElementById('search-preview');
          if (p) p.innerHTML = '<div class="srch-prev-hint">Navigate results to preview</div>';
        }, 150);
      },
      openModeMenu() {
        this.showModes = true;
        this.modeIdx = this.modes.findIndex(m => m.key === this.mode.key);
        if (this.modeIdx < 0) this.modeIdx = 0;
      },
      closeModeMenu() {
        this.showModes = false;
      },
      selectMode(m) {
        this.mode = m;
        this.showModes = false;
        this.$nextTick(() => {
          if (this.$refs.input) this.$refs.input.focus();
          if (this.hasQuery) htmx.trigger(this.$refs.input, 'search');
        });
      },
      onInput(e) {
        const val = e.target.value;
        if (val === '/' && !this.hasQuery) {
          e.target.value = '';
          this.openModeMenu();
          return;
        }
        this.hasQuery = val.trim().length > 0;
      },
      onInputKeydown(e) {
        if (this.showModes) {
          if (e.key === 'ArrowDown') {
            e.preventDefault(); e.stopPropagation();
            this.modeIdx = Math.min(this.modeIdx + 1, this.modes.length - 1);
          } else if (e.key === 'ArrowUp') {
            e.preventDefault(); e.stopPropagation();
            this.modeIdx = Math.max(this.modeIdx - 1, 0);
          } else if (e.key === 'Enter') {
            e.preventDefault(); e.stopPropagation();
            this.selectMode(this.modes[this.modeIdx]);
          } else if (e.key === 'Escape') {
            e.preventDefault(); e.stopPropagation();
            this.closeModeMenu();
          } else if (e.key.length === 1 && !e.metaKey && !e.ctrlKey) {
            e.preventDefault();
            const idx = this.modes.findIndex(m =>
              m.label.toLowerCase().startsWith(e.key.toLowerCase())
            );
            if (idx >= 0) this.modeIdx = idx;
          }
          return;
        }
        if (e.key === 'ArrowDown') { e.preventDefault(); this.moveDown(); }
        else if (e.key === 'ArrowUp') { e.preventDefault(); this.moveUp(); }
        else if (e.key === 'Enter') { e.preventDefault(); this.handleInputEnter(e); }
      },
      moveDown() {
        const items = this._items();
        if (!items.length) return;
        this.activeIdx = Math.min(this.activeIdx + 1, items.length - 1);
        this._highlight(items);
      },
      moveUp() {
        const items = this._items();
        if (!items.length) return;
        this.activeIdx = Math.max(this.activeIdx - 1, 0);
        this._highlight(items);
      },
      confirm() {
        if (this.activeIdx < 0) return;
        const el = this._items()[this.activeIdx];
        if (el) this.selectResult(el);
      },
      confirmPath() {
        if (this.activeIdx < 0) return;
        const el = this._items()[this.activeIdx];
        const path = el && el.dataset.previewPath;
        if (!path) return;
        try { navigator.clipboard.writeText(path); } catch(_) {}
        window.dispatchEvent(new CustomEvent('vaultr:insert-path', { detail: { path } }));
        this.close();
      },
      handleInputEnter(e) {
        if (e.metaKey || e.ctrlKey) { this.confirmPath(); } else { this.confirm(); }
      },
      selectResult(el) {
        if (!el) return;
        if (typeof window.handleSearchResultSelection === 'function' &&
            window.handleSearchResultSelection(el)) {
          this.close();
          return;
        }
        window.location.href = el.href;
      },
      onModeHover(e) {
        const el = e.target && e.target.closest ? e.target.closest('.srch-mode-item') : null;
        if (!el) return;
        const items = Array.from(document.querySelectorAll('.srch-mode-item'));
        const idx = items.indexOf(el);
        if (idx >= 0 && idx !== this.modeIdx) this.modeIdx = idx;
      },
      onResultHover(e) {
        const el = e.target && e.target.closest ? e.target.closest('#search-results a') : null;
        if (!el) return;
        const items = this._items();
        const idx = items.indexOf(el);
        if (idx >= 0 && idx !== this.activeIdx) {
          this.activeIdx = idx;
          this._highlight(items);
        }
      },
      onResultClick(e) {
        const el = e.target && e.target.closest ? e.target.closest('#search-results a') : null;
        if (!el) return;
        if (typeof window.handleSearchResultSelection === 'function' &&
            window.handleSearchResultSelection(el)) {
          e.preventDefault();
          this.close();
        }
      },
      onResultsSwap(e) {
        if (e.detail && e.detail.target && e.detail.target.id === 'search-results') {
          this.activeIdx = -1;
          this._clearPreview();
          this.$nextTick(() => {
            const items = this._items();
            if (items.length) {
              this.activeIdx = 0;
              this._highlight(items);
            }
          });
        }
      },
      onKey(e) {
        if (!this.show) return;
        // When focus leaves the input (e.g. user scrolled the preview panel),
        // still handle nav keys so the results list stays navigable.
        if (document.activeElement === this.$refs.input) return;
        const mod = this.isMac ? e.metaKey : e.ctrlKey;
        if (e.key === 'ArrowDown')  { e.preventDefault(); this.moveDown(); }
        else if (e.key === 'ArrowUp')   { e.preventDefault(); this.moveUp(); }
        else if (e.key === 'Enter' && mod) { e.preventDefault(); this.confirmPath(); }
        else if (e.key === 'Enter')     { e.preventDefault(); this.confirm(); }
      }
    }
  }
`
