  // ── Editor state ────────────────────────────────────────────────────────────
  // One CodeMirror 6 EditorView (s.view) for the whole editor now — the old
  // Milkdown (WYSIWYG) / CodeMirror (source) split is gone. "inSource"/
  // editorMode still exist (see below) but now mean "decorations
  // Compartment reconfigured to plain syntax highlighting" vs "live-preview
  // decorations active", toggled in place on the same view/doc instead of
  // switching between two separately-mounted editors.
  var __vaultrDE = {
    view: null,
    initPromise: null, loading: false, dirty: false,
    pendingBaselineFromEditor: false, pendingBaselineTimer: null,
    currentPath: '', currentDraftId: '', currentMd: '', baselineMd: '',
    saveTimer: null, draftSaveTimer: null, draftSaveTabId: null,
    EditorView: null, EditorState: null, Compartment: null, keymap: null,
    defaultKeymap: null, historyKeymap: null, history: null,
    markdown: null, HighlightStyle: null, syntaxHighlighting: null, tags: null,
    cmUndo: null, cmRedo: null,
    cmSearch: null, cmOpenSearchPanel: null, cmCloseSearchPanel: null,
    cmFindNext: null, cmFindPrev: null, cmReplaceNext: null, cmReplaceAll: null,
    SearchQuery: null, getSearchQuery: null, setSearchQuery: null,
    livePreviewPlugin: null, livePreviewAtomicRanges: null,
    livePreviewTheme: null, wikiMarkdownLanguage: null,
    decoCompartment: null, sharedLanguage: null,
    inSource: false, pendingScrollRaf: null, pendingOpenScroll: null,
    pendingTabScrollRaf: null, pendingTabScrollTimer: null,
  };

  // ── Editor visual effects plugin system ─────────────────────────────────────
  window.__vaultrEditorEffects = (function() {
    var KEY = 'vaultr-editor-effect';
    var effects = {
      none: { label: 'None', desc: 'No effect' },
      particles: {
        label: 'Particles', desc: 'Colorful dots burst from cursor',
        fn: function(c) {
          var colors = ['var(--accent)','var(--p2)','var(--p3)','var(--p1)','var(--p0)','var(--accent-hov)'];
          for (var i = 0; i < 7; i++) {
            var p = document.createElement('div');
            p.className = 'vaultr-ep';
            var angle = (i / 7) * Math.PI * 2 - Math.PI / 2;
            var dist = 18 + Math.random() * 22;
            p.style.cssText = 'left:'+c.left+'px;top:'+c.top+'px;background:'+colors[i%colors.length]+';--ex:'+(Math.cos(angle)*dist).toFixed(1)+'px;--ey:'+(Math.sin(angle)*dist).toFixed(1)+'px';
            document.body.appendChild(p);
            setTimeout(function(el) { el.remove(); }, 620, p);
          }
        },
      },
    };
    return {
      all: function() {
        return Object.keys(effects).map(function(k) {
          return { key: k, label: effects[k].label, desc: effects[k].desc };
        });
      },
      get current() { return localStorage.getItem(KEY) || 'particles'; },
      set: function(key) { localStorage.setItem(KEY, key); },
      trigger: function(view) {
        var eff = effects[this.current];
        if (!eff || !eff.fn) return;
        try {
          var c = view.coordsAtPos(view.state.selection.from);
          eff.fn(c);
        } catch(_) {}
      },
    };
  })();

  var DRAWER_CREATE_KEY = 'vaultr.drawer-create';
  var __vaultrDETabSeq = 0;
  function __vaultrDENewTabId() {
    __vaultrDETabSeq = (__vaultrDETabSeq + 1) % 1000;
    return Date.now() * 1000 + __vaultrDETabSeq;
  }

  // ── Autocomplete (path input in create mode) ─────────────────────────────────
  var __DRAWER_PATH_DBL_ENTER_MS = 2000;
  var __vaultrDEPathAc = null; // created in __vaultrDESetupCreateMode

  function __vaultrDEAcParseCtx(val, caret) {
    val = typeof val === 'string' ? val : '';
    if (caret == null || caret > val.length) caret = val.length;
    var left = val.slice(0, caret);
    if (left.indexOf('/') === -1) return null;
    var slash = left.lastIndexOf('/');
    var partial = left.slice(slash + 1);
    if (partial.indexOf('.') !== -1) return null;
    return { dirPath: slash === 0 ? '/' : left.slice(0, slash), partial: partial, replaceStart: slash + 1 };
  }

  // ── Save helpers ─────────────────────────────────────────────────────────────
  function __vaultrDESaveStatus(txt) {
    var el = document.getElementById('drawer-save-status');
    if (!el) return;
    clearTimeout(el._ssiTimer);
    if (txt === '●') {
      el.dataset.state = 'pending';
    } else if (txt === 'Saved') {
      el.dataset.state = 'saved';
      el._ssiTimer = setTimeout(function() { el.dataset.state = ''; }, 2000);
    } else {
      el.dataset.state = '';
    }
  }
  function __vaultrDEClearPendingBaselineSync() {
    var s = __vaultrDE;
    s.pendingBaselineFromEditor = false;
    if (s.pendingBaselineTimer) { clearTimeout(s.pendingBaselineTimer); s.pendingBaselineTimer = null; }
  }
  function __vaultrDEMarkPendingBaselineSync() {
    var s = __vaultrDE;
    s.pendingBaselineFromEditor = true;
    if (s.pendingBaselineTimer) clearTimeout(s.pendingBaselineTimer);
    s.pendingBaselineTimer = setTimeout(function() {
      s.pendingBaselineTimer = null;
      s.pendingBaselineFromEditor = false;
    }, 500); // Milkdown debounces markdownUpdated at 200ms; 500ms gives safe margin
  }
  function __vaultrDEScheduleSave() {
    __vaultrDESaveStatus('●');
    clearTimeout(__vaultrDE.saveTimer);
    __vaultrDE.saveTimer = setTimeout(__vaultrDEDoSave, 800);
  }
  async function __vaultrDEDoSave() {
    if (!__vaultrDE.dirty) return;
    var path = __vaultrDE.currentPath;
    var content = __vaultrDE.currentMd;
    if (!path) return;
    if (content === __vaultrDE.baselineMd) {
      __vaultrDE.dirty = false;
      clearTimeout(__vaultrDE.saveTimer); __vaultrDE.saveTimer = null;
      __vaultrDESaveStatus('');
      return;
    }
    try {
      var r = await fetch('/api/vault/write', {
        method: 'POST', headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({path: path, content: content}),
      });
      if (r.ok) {
        __vaultrDE.dirty = false; __vaultrDE.baselineMd = content; __vaultrDESaveStatus('Saved');
      } else {
        var errText = ''; try { errText = await r.text(); } catch(_) {}
        window.showError(errText || 'Server error — your changes may not be saved.', 'Save error');
      }
    } catch(e) {
      window.showError((e && e.message) || 'Network error — your changes may not be saved.', 'Save error');
    }
  }

  // ── Misc helpers ─────────────────────────────────────────────────────────────
  function __vaultrDETightenLists(md) {
    return md.replace(
      /(^[ \t]*(?:[-*+]|\d+[.)]) [^\n]*)\n\n(?=[ \t]*(?:[-*+]|\d+[.)]) )/gm, '$1\n');
  }
  function __vaultrDEFindImageFile(dt) {
    if (!dt) return null;
    var items = Array.from(dt.items || []);
    for (var i = 0; i < items.length; i++) {
      if (items[i].kind === 'file' && items[i].type.indexOf('image/') === 0) return items[i].getAsFile();
    }
    return null;
  }
  async function __vaultrDEUploadImage(imgFile) {
    var fd = new FormData();
    fd.append('file', imgFile);
    var resp = await fetch('/api/vault/upload-image', {method: 'POST', body: fd});
    if (!resp.ok) throw new Error(await resp.text());
    return (await resp.json()).src;
  }

  // ── Electron draft store helpers ────────────────────────────────────────────
  function __vaultrDEDraftStore() {
    return window.vaultrDesktop && window.vaultrDesktop.drafts ? window.vaultrDesktop.drafts : null;
  }
  function __vaultrDENewDraftId() {
    return 'draft-' + Date.now().toString(36) + '-' + Math.random().toString(36).slice(2, 10);
  }
  function __vaultrDEActiveTab() {
    var drawer = window.__vaultrDrawer;
    return drawer ? drawer.tabs[drawer.activeTab] : null;
  }
  function __vaultrDEPathInputValue() {
    var pi = document.getElementById('drawer-path-input');
    return pi ? pi.value : '';
  }
  function __vaultrDEDraftTitle(pathInput, content) {
    var name = (pathInput || '').trim();
    if (name) return name.replace(/\.md$/i, '').split('/').pop() || 'New note';
    var first = String(content || '').split(/\r?\n/).map(function(line) {
      return line.replace(/^#+\s*/, '').trim();
    }).find(Boolean);
    return first ? first.slice(0, 48) : 'New note';
  }
  function __vaultrDEEnsureDraftId(tab) {
    if (!tab || tab.path) return '';
    if (!tab.draftId) {
      tab.draftId = __vaultrDENewDraftId();
      var drawer = window.__vaultrDrawer;
      if (drawer) drawer._persist();
    }
    return tab.draftId;
  }
  function __vaultrDEIsActiveTab(tab) {
    var drawer = window.__vaultrDrawer;
    return !!(drawer && tab && drawer.tabs[drawer.activeTab] === tab);
  }
  function __vaultrDEIsActiveTabId(tabId) {
    if (!tabId) return true;
    var drawer = window.__vaultrDrawer;
    var tab = drawer && drawer.tabs[drawer.activeTab];
    return !!(tab && tab.id === tabId);
  }
  function __vaultrDEGetScrollState() {
    var s = __vaultrDE;
    var scroller = document.querySelector('#drawer-edit-area .cm-scroller');
    return {scrollTop: scroller ? scroller.scrollTop : 0, inSource: s.inSource};
  }
  function __vaultrDECaptureDraft(tab) {
    if (!tab || tab.path) return null;
    var drawer = window.__vaultrDrawer;
    var active = __vaultrDEIsActiveTab(tab);
    var live = !!(active && drawer && drawer.drawerOpen && __vaultrDE.currentDraftId === tab.draftId);
    var content = live ? __vaultrDE.currentMd : (tab._draftContent || tab.draftContent || '');
    var pathInput = live ? __vaultrDEPathInputValue() : (tab._pathVal || '');
    var scroll = active ? __vaultrDEGetScrollState() : (__vaultrDERestoreTabState(tab.id) || {});
    tab._draftContent = content || '';
    tab._pathVal = pathInput || '';
    tab.title = __vaultrDEDraftTitle(tab._pathVal, tab._draftContent);
    return {
      version: 1,
      draftId: __vaultrDEEnsureDraftId(tab),
      content: tab._draftContent,
      pathInput: tab._pathVal,
      title: tab.title,
      mode: scroll.inSource ? 'source' : 'wysiwyg',
      scrollTop: scroll.scrollTop || 0,
      createdAt: tab.createdAt || Date.now(),
    };
  }
  function __vaultrDEClearDraftTimer(tab) {
    if (!__vaultrDE.draftSaveTimer) return;
    if (!tab || __vaultrDE.draftSaveTabId === tab.id) {
      clearTimeout(__vaultrDE.draftSaveTimer);
      __vaultrDE.draftSaveTimer = null;
      __vaultrDE.draftSaveTabId = null;
    }
  }
  async function __vaultrDEFlushDraft(tab) {
    if (!tab || tab.path) return;
    __vaultrDEClearDraftTimer(tab);
    var store = __vaultrDEDraftStore();
    if (!store || !store.write) return;
    var data = __vaultrDECaptureDraft(tab);
    if (!data || !data.draftId) return;
    try {
      await store.write(data.draftId, data);
      tab.createdAt = data.createdAt;
    } catch(e) {
      console.warn('draft write failed', e);
    }
  }
  async function __vaultrDESaveTabForLeave(tab) { await tabStateManager.saveForLeave(tab); }
  function __vaultrDEScheduleDraftSave(tab) {
    if (!tab || tab.path) return;
    __vaultrDECaptureDraft(tab);
    __vaultrDEClearDraftTimer(tab);
    __vaultrDE.draftSaveTabId = tab.id;
    __vaultrDE.draftSaveTimer = setTimeout(function() {
      __vaultrDE.draftSaveTimer = null;
      __vaultrDE.draftSaveTabId = null;
      void __vaultrDEFlushDraft(tab);
    }, 350);
  }
  function __vaultrDEScheduleActiveDraftSave() {
    var tab = __vaultrDEActiveTab();
    if (tab && !tab.path) __vaultrDEScheduleDraftSave(tab);
  }
  async function __vaultrDELoadDraft(tab) {
    if (!tab || tab.path) return {content:'', pathInput:'', inSource:false, scrollTop:0};
    var store = __vaultrDEDraftStore();
    var content = tab._draftContent || tab.draftContent || '';
    var pathInput = tab._pathVal || '';
    var loaded = null;
    __vaultrDEEnsureDraftId(tab);
    if (store && store.read) {
      try { loaded = await store.read(tab.draftId); } catch(_) { loaded = null; }
    }
    if (loaded) {
      content = typeof loaded.content === 'string' ? loaded.content : '';
      pathInput = typeof loaded.pathInput === 'string' ? loaded.pathInput : (loaded.path || '');
      tab.createdAt = loaded.createdAt || tab.createdAt || Date.now();
      tab.title = loaded.title || __vaultrDEDraftTitle(pathInput, content);
    } else {
      tab.createdAt = tab.createdAt || Date.now();
      tab.title = __vaultrDEDraftTitle(pathInput, content);
    }
    tab._draftContent = __vaultrDETightenLists(content || '');
    tab._pathVal = pathInput || '';
    if (!loaded && store && store.write) {
      try {
        await store.write(tab.draftId, {
          version: 1,
          draftId: tab.draftId,
          content: tab._draftContent,
          pathInput: tab._pathVal,
          title: tab.title,
          mode: 'wysiwyg',
          scrollTop: 0,
          createdAt: tab.createdAt,
        });
      } catch(e) {
        console.warn('draft write failed', e);
      }
    }
    return {
      content: tab._draftContent,
      pathInput: tab._pathVal,
      inSource: loaded && loaded.mode === 'source',
      scrollTop: loaded ? (loaded.scrollTop || 0) : 0,
    };
  }
  function __vaultrDEDraftEditorState(draft, savedState) {
    return {
      content: draft && typeof draft.content === 'string' ? draft.content : '',
      inSource: savedState && typeof savedState.inSource === 'boolean' ? savedState.inSource : !!(draft && draft.inSource),
      scrollTop: savedState && typeof savedState.scrollTop === 'number' ? savedState.scrollTop : ((draft && draft.scrollTop) || 0),
    };
  }
  async function __vaultrDEDeleteDraft(tab) {
    if (!tab || !tab.draftId) return;
    __vaultrDEClearDraftTimer(tab);
    var store = __vaultrDEDraftStore();
    var id = tab.draftId;
    tab.draftId = '';
    tab._draftContent = '';
    tab._pathVal = '';
    if (__vaultrDE.currentDraftId === id) __vaultrDE.currentDraftId = '';
    if (store && store.delete) {
      try { await store.delete(id); } catch(e) { console.warn('draft delete failed', e); }
    }
  }
  function __vaultrDEHandleContentChange(md, tighten) {
    var s = __vaultrDE;
    var drawer = window.__vaultrDrawer;
    if (drawer && !drawer.drawerOpen) return;
    var next = tighten ? __vaultrDETightenLists(md) : md;
    var tab = __vaultrDEActiveTab();
    if (tab && tab.path && s.pendingBaselineFromEditor && tighten) {
      s.pendingBaselineFromEditor = false;
      if (s.pendingBaselineTimer) { clearTimeout(s.pendingBaselineTimer); s.pendingBaselineTimer = null; }
      s.currentMd = next;
      s.baselineMd = next;
      s.dirty = false;
      clearTimeout(s.saveTimer); s.saveTimer = null;
      __vaultrDESaveStatus('');
      return;
    }
    s.currentMd = next;
    if (tab && !tab.path) {
      s.dirty = false;
      tab._draftContent = next;
      __vaultrDEScheduleDraftSave(tab);
      return;
    }
    if (next !== s.baselineMd) {
      s.dirty = true;
      __vaultrDEScheduleSave();
    } else {
      s.dirty = false;
      clearTimeout(s.saveTimer); s.saveTimer = null;
      __vaultrDESaveStatus('');
    }
  }

  // Open a wiki-link target in the drawer.  value is the raw [[…]] inner text.
  async function __vaultrDrawerOpenWikiLink(value) {
    var drawer = window.__vaultrDrawer;
    if (!drawer) return;
    // Path-like value (contains /): treat as vault-absolute path directly
    if (value.indexOf('/') !== -1) {
      var p = value.startsWith('/') ? value : '/' + value;
      if (!p.endsWith('.md')) p += '.md';
      await drawer.openNoteInDrawer(p, p.split('/').pop().replace(/\.md$/, ''), false, false);
      return;
    }
    // Bare name: resolve through the server
    var nm = value.endsWith('.md') ? value : value + '.md';
    try {
      var r = await fetch('/api/notes/resolve', {
        method: 'POST', headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({name: nm}),
      });
      if (!r.ok) {
        if (window.showError) window.showError('Could not resolve note: ' + value, 'Note not found');
        return;
      }
      var data = await r.json();
      if (!data.matches || !data.matches.length) {
        if (window.showError) window.showError('「' + value + '」 does not exist in the vault.', 'Note not found');
        return;
      }
      var note = data.matches[0];
      var notePath = note.dir === '/' ? '/' + note.name : note.dir + '/' + note.name;
      var noteTitle = note.title || note.name.replace(/\.md$/, '');
      var _isK = note.origin === 'plugin:compile';
      var _isI = note.origin === 'plugin:index';
      await drawer.openNoteInDrawer(notePath, noteTitle, _isK, !!note.pinned, _isI,
        !_isK && !_isI && !note.compile_count);
    } catch(_) {}
  }

  function __vaultrDESaveTabState(tabId) { tabStateManager.save(tabId); }
  function __vaultrDERestoreTabState(tabId) { return tabStateManager.restore(tabId); }
  function __vaultrDEClearTabState(tabId) { tabStateManager.clear(tabId); }
  function __vaultrDECenterActiveTab() {
    var drawer = window.__vaultrDrawer;
    if (!drawer || drawer.activeTab < 0) return false;
    var overlay = document.querySelector('.drawer-overlay');
    var container = overlay && overlay.querySelector('.drawer-tabs');
    var tab = overlay && overlay.querySelectorAll('.drawer-tab')[drawer.activeTab];
    if (!container || !tab) return false;
    var containerRect = container.getBoundingClientRect();
    var tabRect = tab.getBoundingClientRect();
    if (!containerRect.width || !tabRect.width) return false;
    var delta = (tabRect.left + tabRect.width / 2) - (containerRect.left + containerRect.width / 2);
    var maxLeft = Math.max(0, container.scrollWidth - container.clientWidth);
    var nextLeft = Math.max(0, Math.min(maxLeft, container.scrollLeft + delta));
    if (Math.abs(container.scrollLeft - nextLeft) > 0.5) {
      container.scrollTo({left: nextLeft, behavior: 'smooth'});
    }
    return true;
  }
  function __vaultrDEScheduleActiveTabScroll() {
    var s = __vaultrDE;
    if (s.pendingTabScrollRaf) {
      cancelAnimationFrame(s.pendingTabScrollRaf);
      s.pendingTabScrollRaf = null;
    }
    clearTimeout(s.pendingTabScrollTimer);
    clearTimeout(s.pendingTabScrollDelay);
    s.pendingTabScrollDelay = setTimeout(function() {
      s.pendingTabScrollDelay = null;
      s.pendingTabScrollRaf = requestAnimationFrame(function() {
        s.pendingTabScrollRaf = requestAnimationFrame(function() {
          s.pendingTabScrollRaf = null;
          __vaultrDECenterActiveTab();
        });
      });
    }, 80);
    s.pendingTabScrollTimer = setTimeout(function() {
      s.pendingTabScrollTimer = null;
      __vaultrDECenterActiveTab();
    }, 400);
  }
  function __vaultrDESetSourceActive(active) {
    document.querySelectorAll('.drawer-view-btn-wysiwyg').forEach(function(btn) {
      btn.classList.toggle('active', !active);
    });
    document.querySelectorAll('.drawer-view-btn-source').forEach(function(btn) {
      btn.classList.toggle('active', !!active);
    });
    document.querySelectorAll('.drawer-source-toggle-btn').forEach(function(btn) {
      btn.classList.toggle('active', !!active);
    });
  }

  // ── Lazy editor init ─────────────────────────────────────────────────────────
  async function __vaultrEnsureDrawerEditor() {
    var s = __vaultrDE;
    if (s.view) return;
    if (s.initPromise) return s.initPromise;
    s.initPromise = (async function() {
      var mod = await import('/static/editor.js');
      s.EditorView = mod.EditorView; s.EditorState = mod.EditorState; s.Compartment = mod.Compartment;
      s.Transaction = mod.Transaction;
      s.keymap = mod.keymap;
      s.defaultKeymap = mod.defaultKeymap; s.historyKeymap = mod.historyKeymap; s.history = mod.history;
      s.listIndentExtension = mod.listIndentExtension;
      s.markdown = mod.markdown; s.HighlightStyle = mod.HighlightStyle;
      s.syntaxHighlighting = mod.syntaxHighlighting; s.tags = mod.tags;
      s.cmUndo = mod.cmUndo; s.cmRedo = mod.cmRedo;
      s.cmSearch = mod.search; s.cmOpenSearchPanel = mod.openSearchPanel; s.cmCloseSearchPanel = mod.closeSearchPanel;
      s.cmFindNext = mod.findNext; s.cmFindPrev = mod.findPrevious;
      s.cmReplaceNext = mod.replaceNext; s.cmReplaceAll = mod.cmReplaceAll;
      s.SearchQuery = mod.SearchQuery; s.getSearchQuery = mod.getSearchQuery; s.setSearchQuery = mod.setSearchQuery;
      s.livePreviewPlugin = mod.livePreviewPlugin; s.livePreviewAtomicRanges = mod.livePreviewAtomicRanges;
      s.livePreviewTheme = mod.livePreviewTheme; s.codeHighlightStyle = mod.codeHighlightStyle;
      s.wikiMarkdownLanguage = mod.wikiMarkdownLanguage; s.linkClickHandler = mod.linkClickHandler;
      s.horizontalRuleField = mod.horizontalRuleField;
      s.frontmatterCollapseField = mod.frontmatterCollapseField; s.frontmatterHeaderField = mod.frontmatterHeaderField;
      s.frontmatterReadOnly = mod.frontmatterReadOnly; s.allowFrontmatterEdit = mod.allowFrontmatterEdit;

      var editArea = document.getElementById('drawer-edit-area');

      var cmTheme = s.EditorView.theme({
        '&': {height:'100%',color:'var(--prose-body)',background:'transparent'},
        '&.cm-focused': {outline:'none'},
        '.cm-content': {caretColor:'var(--accent)'},
        '.cm-cursor,.cm-dropCursor': {borderLeftColor:'var(--accent)'},
        '.cm-selectionBackground': {background:'var(--cm-selection-bg) !important'},
        '&.cm-focused .cm-selectionBackground': {background:'var(--cm-selection-bg)'},
        '.cm-activeLine': {background:'var(--cm-active-line)'},
        '.cm-gutters': {display:'none'},
      });
      // Raw-markdown ("Source") mode's syntax highlighting — the live-preview
      // decorations (s.livePreviewPlugin etc.) replace this in "Live preview"
      // mode; see editorMode below for how the two are swapped.
      var cmHighlight = s.HighlightStyle.define([
        {tag:s.tags.heading1,color:'var(--h1)',fontWeight:'600'},
        {tag:s.tags.heading2,color:'var(--h2)',fontWeight:'600'},
        {tag:s.tags.heading3,color:'var(--h3)',fontWeight:'600'},
        {tag:s.tags.heading4,color:'var(--h4)',fontWeight:'500'},
        {tag:s.tags.emphasis,fontStyle:'italic',color:'var(--prose-em)'},
        {tag:s.tags.strong,fontWeight:'600',color:'var(--prose-strong)'},
        {tag:s.tags.link,color:'var(--p1)'},{tag:s.tags.url,color:'var(--p1)',opacity:'0.72'},
        {tag:s.tags.monospace,color:'var(--code-tx)'},{tag:s.tags.meta,color:'var(--cm-md-muted)'},
        {tag:s.tags.punctuation,color:'var(--cm-md-muted)'},
        {tag:s.tags.processingInstruction,color:'var(--cm-md-muted)'},
        {tag:s.tags.strikethrough,color:'var(--muted)',textDecoration:'line-through'},
      ]);

      // Shared by both modes so toggling reconfigures s.decoCompartment
      // around the same parse instead of re-parsing from scratch.
      s.sharedLanguage = s.wikiMarkdownLanguage();

      var liveOptions = {
        resolveImageSrc: function(filename) {
          return '/api/images/serve?name=' + encodeURIComponent(filename);
        },
        onWikiLinkClick: function(target) { void __vaultrDrawerOpenWikiLink(target); },
      };
      // Exposed on s so editorMode (a separate closure) can reconfigure
      // s.decoCompartment without rebuilding these from scratch. Frontmatter
      // now renders through s.livePreviewPlugin like everything else — see
      // cm-live/index.js's livePreviewExtensions() comment for why it no
      // longer needs its own opt-in field.
      s._liveModeExt = function() {
        return [
          s.livePreviewPlugin.of(liveOptions),
          s.livePreviewAtomicRanges(),
          s.horizontalRuleField(),
          // Live-preview only: source mode is the raw-YAML editing surface,
          // so frontmatter stays freely editable there. The Metadata
          // header's edit dialog is the only path in this mode.
          s.frontmatterReadOnly(),
          s.livePreviewTheme,
          // Fenced-code token colors (keyword/string/comment/...) — see
          // cm-live/theme.js's codeHighlightStyle comment for why this is a
          // separate, code-scoped style rather than cm-live/index.js's
          // livePreviewExtensions() bundling it in directly: this file
          // hand-picks pieces instead of calling that function (see the
          // comment above s._liveModeExt) so wikiMarkdownLanguage() can stay
          // shared across both modes outside s.decoCompartment — anything
          // livePreviewExtensions() adds has to be mirrored here too.
          s.syntaxHighlighting(s.codeHighlightStyle),
        ];
      };
      s._sourceModeExt = function() {
        return [s.syntaxHighlighting(cmHighlight)];
      };
      s.decoCompartment = new s.Compartment();
      // Own compartment so editorMode.syncContent() (below) can wipe the
      // undo stack on note/tab switch — this EditorView is a session-long
      // singleton (__vaultrEnsureDrawerEditor only ever creates it once),
      // so without this reset Ctrl+Z after switching notes walks back into
      // the PREVIOUS note's edit history against the new note's document.
      s.historyCompartment = new s.Compartment();

      s.view = new s.EditorView({
        parent: editArea,
        state: s.EditorState.create({
          doc: s.currentMd || '',
          extensions: [
            s.sharedLanguage,
            s.historyCompartment.of(s.history()),
            // listIndentExtension is Prec.highest internally (see
            // cm-live/list-indent.js) — @codemirror/lang-markdown's own
            // language support registers a high-precedence Enter binding for
            // "continue list markup" that a plain keymap.of(...) here would
            // lose to regardless of array position. It handles Tab/Shift-Tab
            // (not bound by defaultKeymap at all — every outliner-style
            // editor claims them, same as Cmd+]/Cmd+[) and Enter on an empty
            // list item specifically (the built-in path is supposed to
            // outdent/exit there but unreliably just inserts a blank line
            // with the marker left dangling instead); a non-empty item's
            // Enter isn't handled here and falls through to defaultKeymap.
            s.listIndentExtension,
            s.keymap.of([...s.defaultKeymap, ...s.historyKeymap]),
            s.cmSearch({ top: true, createPanel: __vaultrCreateSearchPanel }),
            s.EditorView.lineWrapping, cmTheme,
            s.EditorView.contentAttributes.of({spellcheck: 'false'}),
            s.linkClickHandler(), // click a collapsed link to open it, Obsidian-style — works in both modes, not compartmented
            // Outside the compartment so the collapsed flag survives a
            // source/live-preview toggle — reconfigure() tears down and
            // recreates any StateField that was only inside _liveModeExt().
            s.frontmatterCollapseField,
            s.frontmatterHeaderField({ onEditFrontmatter: __vaultrDEEditFrontmatter }),
            s.decoCompartment.of(s._liveModeExt()),
            s.EditorView.updateListener.of(function(update) {
              if (!update.docChanged || s.loading) return;
              __vaultrDEHandleContentChange(update.state.doc.toString(), false);
            }),
            s.EditorView.domEventHandlers({
              paste: function(e, view) {
                var imgFile = __vaultrDEFindImageFile(e.clipboardData);
                if (!imgFile) return false;
                e.preventDefault();
                __vaultrDEUploadImage(imgFile).then(function(src) {
                  var filename = src.split('/').pop();
                  var ins = '![[' + filename + ']]'; var sel = view.state.selection.main;
                  view.dispatch({changes:{from:sel.from,to:sel.to,insert:ins},selection:{anchor:sel.from+ins.length}});
                }).catch(function(e) {
                  window.showError((e && e.message) || 'Image upload failed.', 'Upload error');
                });
                return true;
              },
              keydown: function(e) {
                if (e.key === 'Enter' && !e.isComposing) window.__vaultrEditorEffects.trigger(s.view);
                return false;
              },
            }),
          ],
        }),
      });

      document.querySelectorAll('.drawer-view-btn-wysiwyg').forEach(function(btn) {
        btn.addEventListener('click', function() {
          if (__vaultrDE.inSource) editorMode.exitSource();
        });
      });
      document.querySelectorAll('.drawer-view-btn-source').forEach(function(btn) {
        btn.addEventListener('click', function() {
          if (!__vaultrDE.inSource) editorMode.enterSource();
        });
      });
    })();
    return s.initPromise;
  }

  // ── Search panel (custom, top-anchored) ─────────────────────────────────────
  function __vaultrCreateSearchPanel(view) {
    var s = __vaultrDE;
    var dom = document.createElement('div');
    dom.className = 'vaultr-search-panel';

    // Row 1: Find
    var findRow = document.createElement('div');
    findRow.className = 'vaultr-sr-row';
    var findWrap = document.createElement('div');
    findWrap.className = 'vaultr-sr-input-wrap';

    var findInput = document.createElement('input');
    findInput.type = 'text'; findInput.placeholder = 'Find';
    findInput.className = 'vaultr-sr-input'; findInput.setAttribute('main-field', '');
    findInput.setAttribute('aria-label', 'Find');

    var findInset = document.createElement('div');
    findInset.className = 'vaultr-sr-inset-btns';

    var prevBtn = document.createElement('button');
    prevBtn.type = 'button'; prevBtn.className = 'vaultr-sr-ibtn'; prevBtn.title = 'Previous (Shift+Enter)';
    prevBtn.innerHTML = '<svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m18 15-6-6-6 6"/></svg>';

    var nextBtn = document.createElement('button');
    nextBtn.type = 'button'; nextBtn.className = 'vaultr-sr-ibtn'; nextBtn.title = 'Next (Enter)';
    nextBtn.innerHTML = '<svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>';

    var caseBtn = document.createElement('button');
    caseBtn.type = 'button'; caseBtn.className = 'vaultr-sr-ibtn vaultr-sr-toggle'; caseBtn.title = 'Match case';
    caseBtn.textContent = 'Aa';

    var closeBtn = document.createElement('button');
    closeBtn.type = 'button'; closeBtn.className = 'vaultr-sr-ibtn'; closeBtn.setAttribute('aria-label', 'Close');
    closeBtn.innerHTML = '<svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>';

    findInset.append(prevBtn, nextBtn, caseBtn);
    findWrap.append(findInput, findInset);
    findRow.append(findWrap, closeBtn);

    // Row 2: Replace
    var replaceRow = document.createElement('div');
    replaceRow.className = 'vaultr-sr-row';
    var replaceWrap = document.createElement('div');
    replaceWrap.className = 'vaultr-sr-input-wrap';

    var replaceInput = document.createElement('input');
    replaceInput.type = 'text'; replaceInput.placeholder = 'Replace';
    replaceInput.className = 'vaultr-sr-input'; replaceInput.setAttribute('aria-label', 'Replace');

    var replaceInset = document.createElement('div');
    replaceInset.className = 'vaultr-sr-inset-btns';

    var replaceBtn = document.createElement('button');
    replaceBtn.type = 'button'; replaceBtn.className = 'vaultr-sr-ibtn'; replaceBtn.title = 'Replace (Enter)';
    replaceBtn.innerHTML = '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 4v7a4 4 0 0 1-4 4H4"/><path d="m9 10-5 5 5 5"/></svg>';

    var replaceAllBtn = document.createElement('button');
    replaceAllBtn.type = 'button'; replaceAllBtn.className = 'vaultr-sr-ibtn'; replaceAllBtn.title = 'Replace All';
    replaceAllBtn.innerHTML = '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 17 5-5-5-5"/><path d="m13 17 5-5-5-5"/></svg>';

    replaceInset.append(replaceBtn, replaceAllBtn);
    replaceWrap.append(replaceInput, replaceInset);
    replaceRow.append(replaceWrap);
    dom.append(findRow, replaceRow);

    // State
    var caseSensitive = false;

    function buildQuery() {
      return new s.SearchQuery({ search: findInput.value, caseSensitive: caseSensitive, replace: replaceInput.value });
    }
    function commit() { view.dispatch({ effects: s.setSearchQuery.of(buildQuery()) }); }

    findInput.addEventListener('input', commit);
    replaceInput.addEventListener('input', commit);

    caseBtn.addEventListener('click', function() {
      caseSensitive = !caseSensitive;
      caseBtn.classList.toggle('active', caseSensitive);
      commit(); findInput.focus();
    });
    findInput.addEventListener('keydown', function(e) {
      if (e.key === 'Enter') { e.preventDefault(); if (e.shiftKey) s.cmFindPrev(view); else s.cmFindNext(view); }
      if (e.key === 'Escape') { e.preventDefault(); s.cmCloseSearchPanel(view); }
    });
    replaceInput.addEventListener('keydown', function(e) {
      if (e.key === 'Enter') { e.preventDefault(); s.cmReplaceNext(view); }
      if (e.key === 'Escape') { e.preventDefault(); s.cmCloseSearchPanel(view); }
    });
    prevBtn.addEventListener('click', function() { s.cmFindPrev(view); });
    nextBtn.addEventListener('click', function() { s.cmFindNext(view); });
    replaceBtn.addEventListener('click', function() { s.cmReplaceNext(view); });
    replaceAllBtn.addEventListener('click', function() { s.cmReplaceAll(view); });
    closeBtn.addEventListener('click', function() { s.cmCloseSearchPanel(view); });

    return {
      dom: dom,
      top: true,
      mount: function() {
        var sel = view.state.selection.main;
        if (!sel.empty) {
          var txt = view.state.sliceDoc(sel.from, sel.to);
          if (txt && !txt.includes('\n')) { findInput.value = txt; commit(); }
        }
        findInput.focus(); findInput.select();
      },
    };
  }

  // ── Editor mode state machine ────────────────────────────────────────────────
  // Single authority for live-preview ↔ source transitions. Both modes are
  // the same EditorView/doc now — this only reconfigures s.decoCompartment
  // (see __vaultrEnsureDrawerEditor) and, when the caller is about to show a
  // *different* note's content, syncs s.currentMd into the view first.
  // applySource / applyWysiwyg: low-level, called by applyState (skipFocus=true).
  // enterSource / exitSource / toggle: user-triggered, manage focus themselves.
  var editorMode = (function() {
    function syncContent() {
      var s = __vaultrDE;
      if (s.view.state.doc.toString() !== s.currentMd) {
        // Whole-document swap (mode toggle, tab/note switch), not a user
        // edit to frontmatter text — frontmatterReadOnly()'s changeFilter
        // computes its blocked range from the *old* doc's frontmatter and
        // would otherwise clip this replace, silently dropping everything
        // after the old frontmatter (the exact "only frontmatter shows"
        // bug). allowFrontmatterEdit is the filter's one escape hatch.
        //
        // Also resets historyCompartment and excludes this swap from
        // history itself, in the same transaction — s.view is a session-
        // long singleton reused across every note, so without this a note
        // switch is just another undoable edit: Ctrl+Z right after opening
        // a different note would revert its content to the PREVIOUS note's
        // text while s.currentPath/s.currentMd still track the new note,
        // risking that stale content gets autosaved under the new path.
        var annotations = [s.Transaction.addToHistory.of(false)];
        if (s.allowFrontmatterEdit) annotations.push(s.allowFrontmatterEdit.of(true));
        s.view.dispatch({
          changes: {from: 0, to: s.view.state.doc.length, insert: s.currentMd},
          effects: s.historyCompartment.reconfigure(s.history()),
          annotations: annotations,
        });
      }
    }
    return {
      applySource: function(opts) {
        var s = __vaultrDE;
        __vaultrDEClearPendingBaselineSync();
        s.loading = true;
        syncContent();
        s.view.dispatch({effects: s.decoCompartment.reconfigure(s._sourceModeExt())});
        s.loading = false;
        s.inSource = true;
        __vaultrDESetSourceActive(true);
        if (!(opts && opts.skipFocus)) focusManager.focusEditor();
      },
      applyWysiwyg: function(opts) {
        var s = __vaultrDE;
        var tab = __vaultrDEActiveTab();
        if (tab && tab.path) __vaultrDEMarkPendingBaselineSync();
        s.loading = true;
        syncContent();
        s.view.dispatch({effects: s.decoCompartment.reconfigure(s._liveModeExt())});
        setTimeout(function() { s.loading = false; }, 50);
        s.inSource = false;
        __vaultrDESetSourceActive(false);
        if (!(opts && opts.skipFocus)) focusManager.focusEditor();
      },
      // User-triggered: live preview → source
      enterSource: function() { this.applySource(); },
      // User-triggered: source → live preview
      exitSource: function() {
        var s = __vaultrDE;
        s.currentMd = s.view.state.doc.toString();
        this.applyWysiwyg();
      },
      toggle: function() {
        var s = __vaultrDE;
        if (s.inSource) this.exitSource(); else this.enterSource();
      },
    };
  })();

  // ── Focus manager ────────────────────────────────────────────────────────────
  // Single authority for all editor focus/blur decisions.
  var focusManager = {
    focusEditor: function() {
      var s = __vaultrDE;
      if (s.view) s.view.focus();
    },
    // Focus the path input (create-mode toolbar).
    focusPathInput: function() {
      var pi = document.getElementById('drawer-path-input');
      if (pi) pi.focus();
    },
    // Blur whatever currently has focus.
    blurActive: function() {
      if (document.activeElement && document.activeElement.blur) document.activeElement.blur();
    },
    // True when focus is inside the editor content area.
    isInsideEditor: function() {
      var ae = document.activeElement;
      var ea = document.getElementById('drawer-edit-area');
      return !!(ea && ea.contains(ae));
    },
  };

  function __vaultrDEEnterSource() { editorMode.enterSource(); }
  function __vaultrDEExitSource() { editorMode.exitSource(); }

  // "Metadata" header's pencil button (cm-live/frontmatter-collapse.js) —
  // frontmatter is read-only in live-preview mode (frontmatterReadOnly()),
  // so this dialog + allowFrontmatterEdit-annotated dispatch is the only
  // way to change it there. from/to span the whole node including the
  // "---" delimiters; only the interior YAML is shown/edited.
  function __vaultrDEEditFrontmatter(view, from, to) {
    var raw = view.state.doc.sliceString(from, to);
    var lines = raw.split('\n');
    var hasClose = lines.length > 1 && lines[lines.length - 1].trim() === '---';
    var interior = (hasClose ? lines.slice(1, lines.length - 1) : lines.slice(1)).join('\n');
    if (!window.__vaultrEditFrontmatter) return;
    window.__vaultrEditFrontmatter(interior, function(newInterior) {
      var body = newInterior.replace(/\s+$/, '');
      var newBlock = '---\n' + (body ? body + '\n' : '') + '---';
      view.dispatch({
        changes: { from: from, to: to, insert: newBlock },
        annotations: __vaultrDE.allowFrontmatterEdit.of(true),
      });
    });
  }

  // ── Tab state manager ─────────────────────────────────────────────────────────
  // Owns the per-tab saved state (scroll position, source mode).
  var tabStateManager = (function() {
    var _states = new Map();  // tabId → { scrollTop, inSource }
    return {
      save: function(tabId) {
        if (!tabId) return;
        var s = __vaultrDE;
        var scroller = document.querySelector('#drawer-edit-area .cm-scroller');
        _states.set(tabId, { scrollTop: scroller ? scroller.scrollTop : 0, inSource: s.inSource });
      },
      restore: function(tabId) {
        if (!tabId) return null;
        return _states.get(tabId) || null;
      },
      clear: function(tabId) {
        if (!tabId) return;
        _states.delete(tabId);
      },
      saveForLeave: async function(tab) {
        if (!tab) return;
        this.save(tab.id);
        if (!tab.path) await __vaultrDEFlushDraft(tab);
      },
    };
  })();

  // ── Content loaders ──────────────────────────────────────────────────────────
  async function __vaultrDrawerLoadNote(path, tabId, savedState) {
    var s = __vaultrDE;
    __vaultrDEClearPendingBaselineSync();
    if (s.dirty && s.currentPath && s.currentPath !== path) {
      clearTimeout(s.saveTimer); s.saveTimer = null; await __vaultrDEDoSave();
    } else { clearTimeout(s.saveTimer); s.saveTimer = null; }
    if (!__vaultrDEIsActiveTabId(tabId)) return false;
    __vaultrDESaveStatus('');
    var ptEl = document.getElementById('drawer-path-text');
    if (ptEl) ptEl.textContent = path;

    // Load from server
    var resp = await fetch('/api/vault/read', {
      method:'POST', headers:{'Content-Type':'application/json'},
      body: JSON.stringify({path: path}),
    });
    if (!resp.ok) {
      if (resp.status === 404 && window.showError) {
        window.showError('「' + path.split('/').pop().replace(/\.md$/, '') + '」 does not exist in the vault.', 'Note not found');
      }
      return false;
    }
    var content = await resp.text();
    if (!__vaultrDEIsActiveTabId(tabId)) return false;
    s.currentPath = path; s.currentDraftId = ''; s.currentMd = __vaultrDETightenLists(content); s.baselineMd = s.currentMd; s.dirty = false;
    
    // Apply state (will create new state if no saved state)
    return await __vaultrDEApplyState(savedState || { inSource: false, scrollTop: 0 }, tabId);
  }

  async function __vaultrDrawerSetContent(content, tabId, savedState, draftId) {
    var s = __vaultrDE;
    __vaultrDEClearPendingBaselineSync();
    if (s.dirty && s.currentPath) { clearTimeout(s.saveTimer); s.saveTimer = null; await __vaultrDEDoSave(); }
    else { clearTimeout(s.saveTimer); s.saveTimer = null; }
    if (!__vaultrDEIsActiveTabId(tabId)) return false;
    __vaultrDESaveStatus('');
    s.currentPath = ''; s.currentDraftId = draftId || ''; s.currentMd = __vaultrDETightenLists(content || ''); s.baselineMd = ''; s.dirty = false;
    
    // Apply state
    return await __vaultrDEApplyState(savedState || { inSource: false, scrollTop: 0 }, tabId);
  }
  
  async function __vaultrDEApplyState(state, expectedTabId) {
    var s = __vaultrDE;
    s.loading = true;
    await __vaultrEnsureDrawerEditor();
    if (!__vaultrDEIsActiveTabId(expectedTabId)) {
      s.loading = false;
      return false;
    }

    var targetInSource = state.inSource || false;
    var targetScroll = state.scrollTop || 0;

    if (targetInSource) {
      editorMode.applySource({skipFocus: true});
    } else {
      editorMode.applyWysiwyg({skipFocus: true});
    }

    if (s.pendingScrollRaf) { cancelAnimationFrame(s.pendingScrollRaf); s.pendingScrollRaf = null; }
    clearTimeout(s.pendingOpenScroll);
    var getScroller = function() { return document.querySelector('#drawer-edit-area .cm-scroller'); };
    // Set it once synchronously, in the same tick as the content swap above
    // (dispatch() has already updated the DOM by the time this line runs) —
    // otherwise the scroller keeps the PREVIOUS tab's scrollTop for at
    // least one paint, since the rAF/setTimeout calls below are the only
    // other place this gets set and both are deliberately deferred (see
    // their own comments). That stale-scrollTop-against-new-content frame
    // is the tab-switch flash: this line is what removes it for the common
    // case; the deferred ones stay as-is for the one case that still needs
    // them — the drawer's own slide-in animation tearing down a compositing
    // layer and resetting scrollTop out from under this synchronous set.
    var syncEl = getScroller(); if (syncEl) syncEl.scrollTop = targetScroll;
    s.pendingScrollRaf = requestAnimationFrame(function() {
      s.pendingScrollRaf = requestAnimationFrame(function() {
        s.pendingScrollRaf = null;
        if (!__vaultrDEIsActiveTabId(expectedTabId)) return;
        var el = getScroller(); if (el) el.scrollTop = targetScroll;
        var tabForFocus = __vaultrDEActiveTab();
        var isDraftTab = tabForFocus && !tabForFocus.path;
        if (!focusManager.isInsideEditor() && !isDraftTab) {
          focusManager.focusEditor();
        }
      });
    });
    s.pendingOpenScroll = setTimeout(function() {
      s.pendingOpenScroll = null;
      if (!__vaultrDEIsActiveTabId(expectedTabId)) return;
      var el = getScroller(); if (el) el.scrollTop = targetScroll;
    }, 270);
    return true;
  }


  // ── Create mode: path input handlers + Publish ───────────────────────────────
  function __vaultrDESetupCreateMode() {
    var pi = document.getElementById('drawer-path-input');
    var pb = document.getElementById('drawer-publish-btn');
    if (!pi) return;

    // Build the autocomplete instance via the shared factory.
    var enterFirstTs = 0;
    __vaultrDEPathAc = __vaultrPathAcCreate({
      getInput: function() { return document.getElementById('drawer-path-input'); },
      getList:  function() { return document.getElementById('drawer-path-ac'); },
      parseCtx: __vaultrDEAcParseCtx,
      onApply: function(input, newVal, caretPos) {
        input.value = newVal;
        input.setSelectionRange(caretPos, caretPos);
        input.focus();
        input.classList.remove('invalid');
        input.placeholder = 'filename.md  ·  or  /folder/note.md';
        __vaultrDEScheduleActiveDraftSave();
      },
      escKey: 'drawer-ac',
    });

    pi.addEventListener('input', function() {
      enterFirstTs = 0;
      pi.classList.remove('invalid');
      pi.placeholder = 'filename.md  ·  or  /folder/note.md';
      __vaultrDEPathAc.refresh();
      __vaultrDEScheduleActiveDraftSave();
    });
    pi.addEventListener('click', function() { __vaultrDEPathAc.refresh(); });
    pi.addEventListener('keyup', function(ev) {
      if (ev.key === 'ArrowLeft' || ev.key === 'ArrowRight' || ev.key === 'Home' || ev.key === 'End')
        __vaultrDEPathAc.refresh();
    });
    pi.addEventListener('blur', function() {
      enterFirstTs = 0;
      setTimeout(function() {
        var acEl = document.getElementById('drawer-path-ac');
        if (!acEl || !acEl.contains(document.activeElement)) __vaultrDEPathAc.close();
      }, 180);
    });
    pi.addEventListener('keydown', function(ev) {
      if (__vaultrDEPathAc.handleKeydown(ev)) return;
      if (ev.key !== 'Enter') return;
      var now = Date.now();
      if (enterFirstTs && (now - enterFirstTs) <= __DRAWER_PATH_DBL_ENTER_MS) {
        ev.preventDefault(); enterFirstTs = 0;
        var s = __vaultrDE;
        if (s.view) s.view.focus();
        return;
      }
      enterFirstTs = now;
    });
    if (pb) pb.addEventListener('click', __vaultrDrawerPublish);
  }

  async function __vaultrDrawerPublish() {
    var pi = document.getElementById('drawer-path-input');
    var pb = document.getElementById('drawer-publish-btn');
    if (!pi || !pb) return;
    var drawer = window.__vaultrDrawer;
    var tab = drawer ? drawer.tabs[drawer.activeTab] : null;
    if (tab && !tab.path) await __vaultrDEFlushDraft(tab);
    var name = pi.value.trim();
    if (!name) {
      window.showError('A file name is required to publish.', 'Cannot publish');
      pi.classList.add('invalid'); pi.focus(); return;
    }
    var apiPath = name.startsWith('/') ? name : '/' + name;
    if (!apiPath.endsWith('.md')) apiPath += '.md';
    pb.disabled = true; pb.textContent = 'Publishing…';
    var published = false;
    try {
      var statResp = await fetch('/api/vault/stat', {
        method: 'POST', headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({path: apiPath}),
      });
      if (statResp.ok) {
        window.showError('"' + apiPath + '" already exists — rename the file.', 'Cannot publish');
        pi.classList.add('invalid'); return;
      }
      var baseName = apiPath.split('/').pop();
      var resolveResp = await fetch('/api/notes/resolve', {
        method: 'POST', headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({name: baseName}),
      });
      if (resolveResp.ok) {
        var resolveData = await resolveResp.json();
        if (resolveData.count > 0) {
          window.showError('"' + baseName.replace(/\.md$/i, '') + '" already exists in the vault — rename the file.', 'Cannot publish');
          pi.classList.add('invalid'); return;
        }
      }
      var resp = await fetch('/api/vault/write', {
        method:'POST', headers:{'Content-Type':'application/json'},
        body: JSON.stringify({path: apiPath, content: __vaultrDE.currentMd}),
      });
      if (!resp.ok) {
        var msg = (await resp.text()) || 'Publish failed.';
        window.showError(msg, 'Publish failed');
        pi.classList.add('invalid'); return;
      }
      published = true;
      __vaultrDE.currentPath = apiPath; __vaultrDE.dirty = false; __vaultrDE.baselineMd = __vaultrDE.currentMd; __vaultrDESaveStatus('Saved');
      __vaultrDE.currentDraftId = '';
      __vaultrDEPathAc && __vaultrDEPathAc.close();
      if (drawer) {
        if (tab && !tab.path) {
          var oldDraftId = tab.draftId;
          tab.path = apiPath;
          tab.title = apiPath.split('/').pop().replace(/\.md$/, '') || apiPath;
          tab.draftId = '';
          tab._draftContent = '';
          tab._pathVal = '';
          drawer._persist();
          var draftStore = __vaultrDEDraftStore();
          if (oldDraftId && draftStore && draftStore.delete)
            void draftStore.delete(oldDraftId).catch(function(){});
        }
      }
      var ptEl = document.getElementById('drawer-path-text');
      if (ptEl) ptEl.textContent = apiPath;
      pb.classList.add('success'); pb.textContent = '✓ Published';
      setTimeout(function(){ pb.disabled = false; pb.textContent = 'Publish'; pb.classList.remove('success'); }, 1200);
      if (window.__vaultrAfterVaultMutation) await window.__vaultrAfterVaultMutation();
    } catch(e) {
      window.showError((e && e.message) || 'Network error.', 'Publish failed');
      pi.classList.add('invalid');
    } finally { if (!published) { pb.disabled = false; pb.textContent = 'Publish'; } }
  }

  // ── Global note-card helper ───────────────────────────────────────────────────
  function __vaultrOpenNote(el) {
    var d = el && el.dataset; if (!d || !d.notePath) return;
    var drawer = window.__vaultrDrawer;
    if (drawer) void drawer.openNoteInDrawer(d.notePath, d.noteTitle || 'Note',
      d.noteIsKnowledge === 'true', d.notePinned === 'true', d.noteIsIndex === 'true',
      d.noteCanCompile === 'true');
  }

  // ── Drawer controller factory ─────────────────────────────────────────────────
  function drawerCtrl() {
    return {
      drawerOpen: false, drawerExpanded: false, tabs: [], activeTab: -1,

      _persist() {
        try {
          localStorage.setItem('vaultr.drawer', JSON.stringify({
            tabs: this.tabs.map(function(t){
              var obj = {id:t.id, title:t.title, path:t.path, isKnowledge:!!t.isKnowledge, pinned:!!t.pinned, isIndex:!!t.isIndex, canCompile:!!t.canCompile};
              if (!t.path) {
                obj.draftId = t.draftId || '';
                if (!t.draftId && (t.draftContent || t._pathVal)) {
                  obj.draftContent = t.draftContent || '';
                  obj._pathVal = t._pathVal || '';
                }
              }
              return obj;
            }),
            activeTab: this.activeTab,
          }));
        } catch(_) {}
      },

      _restore() {
        try {
          var raw = localStorage.getItem('vaultr.drawer'); if (!raw) return;
          var data = JSON.parse(raw);
          if (!data || !Array.isArray(data.tabs) || !data.tabs.length) return;
          var tabs = data.tabs;
          var MAX_TABS = 10;
          if (tabs.length > MAX_TABS) {
            var publishedByAge = tabs.filter(function(t){ return !!t.path; }).sort(function(a,b){return a.id-b.id;});
            var removeCount = Math.min(tabs.length-MAX_TABS, publishedByAge.length);
            var remove = new Set(publishedByAge.slice(0, removeCount).map(function(t){return t.id;}));
            tabs = tabs.filter(function(t){return !remove.has(t.id);});
          }
          this.tabs = tabs.map(function(t) {
            var path = t.path || '';
            if (!path && t.fragmentUrl) { try { path = new URL(t.fragmentUrl, location.origin).searchParams.get('path') || ''; } catch(_){} }
            if (!path && t.pageUrl)     { try { path = new URL(t.pageUrl,     location.origin).searchParams.get('path') || ''; } catch(_){} }
            var tab = {id:t.id, title:t.title||'Note', path:path, isKnowledge:!!t.isKnowledge, pinned:!!t.pinned, isIndex:!!t.isIndex, canCompile:!!t.canCompile};
            if (!path) {
              tab.draftId = t.draftId || '';
              tab.draftContent = t.draftContent || '';
              tab._draftContent = t.draftContent || '';
              tab._pathVal = t._pathVal || '';
            }
            return tab;
          });
          if (!this.tabs.length) return;
          this.activeTab = Math.min(Math.max(data.activeTab||0, 0), this.tabs.length-1);
        } catch(_) {}
      },

      initDrawer() {
        window.__vaultrDrawer = this;
        this._restore();
        __vaultrDESetupCreateMode();

        var self = this; var prevOpen = false;
        var _drawerEscClose = function() {
          var moreMenu = document.querySelector('.drawer-more-menu');
          if (moreMenu && moreMenu.style.display !== 'none') {
            document.dispatchEvent(new CustomEvent('drawer:close-more'));
            focusManager.blurActive();
            return;
          }
          if (document.querySelector('.vaultr-search-panel')) {
            var _s = __vaultrDE;
            if (_s.view && _s.cmCloseSearchPanel) _s.cmCloseSearchPanel(_s.view);
            return;
          }
          self.drawerOpen = false;
        };
        this.$watch('drawerExpanded', function(expanded) {
          if (window.vaultrDesktop && window.vaultrDesktop.setWindowButtonVisibility) {
            window.vaultrDesktop.setWindowButtonVisibility(!expanded);
          }
          if (expanded) {
            if (window.__vaultrEscPop) window.__vaultrEscPop('drawer');
          } else if (self.drawerOpen) {
            if (window.__vaultrEscPush) window.__vaultrEscPush('drawer', _drawerEscClose);
          }
        });
        this.$watch('drawerOpen', async function(isOpen) {
          if (!isOpen) {
            if (window.__vaultrEscPop) window.__vaultrEscPop('drawer');
            if (self.drawerExpanded) {
              self.drawerExpanded = false;
              if (window.vaultrDesktop && window.vaultrDesktop.setWindowButtonVisibility) {
                window.vaultrDesktop.setWindowButtonVisibility(true);
              }
            }
            // Save state BEFORE blur — blur can trigger scrollIntoView which resets scrollTop
            var currentTab = self.tabs[self.activeTab];
            if (currentTab) await __vaultrDESaveTabForLeave(currentTab);
            focusManager.blurActive();
            prevOpen = false; return;
          }
          prevOpen = true;
          var _drawerOverlayEl = document.querySelector('.drawer-overlay');
          if (_drawerOverlayEl) {
            _drawerOverlayEl.classList.add('drawer-is-opening');
            setTimeout(function() { _drawerOverlayEl.classList.remove('drawer-is-opening'); }, 320);
          }
          if (window.__vaultrEscPush) window.__vaultrEscPush('drawer', _drawerEscClose);
          // Refresh key-behavior config on every open so settings changes take
          // effect immediately without restarting the app. Called before any
          // content loading so all code paths (same note, new note, draft) pick
          // it up. No-op if the editor hasn't been created yet (handled later in
          // __vaultrEnsureDrawerEditor's initPromise).
          // Always read latest tabs from localStorage before opening — other
          // WebContentsViews (same session, different JS context) may have
          // added tabs since this view last called _restore().
          self._restore();
          var tab = self.tabs[self.activeTab];
          if (!tab) return;
          __vaultrDEScheduleActiveTabScroll();
          
          var savedState = __vaultrDERestoreTabState(tab.id);
          
          if (!tab.path) {
            await self._activateDraftTab(tab, savedState);
            return;
          }
          if (__vaultrDE.currentPath !== tab.path || !__vaultrDE.dirty) {
            void __vaultrDrawerLoadNote(tab.path, tab.id, savedState);
          } else if (savedState) {
            // Same note with unsaved edits — restore scroll only. Two passes:
            // 1) rAF: fire immediately so scroll looks right during slide-in animation
            // 2) setTimeout(250): fire after the 240ms CSS transition in case the browser
            //    resets scrollTop when the GPU compositing layer is torn down at animation end
            var s = __vaultrDE;
            var sc = savedState.scrollTop || 0;
            function __applyDrawerScroll() {
              var scroller = document.querySelector('#drawer-edit-area .cm-scroller');
              if (scroller) scroller.scrollTop = sc;
            }
            if (s.pendingScrollRaf) { cancelAnimationFrame(s.pendingScrollRaf); s.pendingScrollRaf = null; }
            clearTimeout(s.pendingOpenScroll);
            s.pendingScrollRaf = requestAnimationFrame(function() {
              s.pendingScrollRaf = null;
              __applyDrawerScroll();
              if (!focusManager.isInsideEditor()) {
                focusManager.focusEditor();
              }
            });
            s.pendingOpenScroll = setTimeout(function() {
              s.pendingOpenScroll = null;
              __applyDrawerScroll();
            }, 250);
          }
        });
        this.$watch('tabs', function(){ self._persist(); });
        this.$watch('activeTab', function(i) {
          self._persist();
          self.$nextTick(function() { __vaultrDEScheduleActiveTabScroll(); });
        });

        // Sync drawer state from other section views (each section is a separate
        // WebContentsView with its own JS context; storage events cross view boundaries).
        window.addEventListener('storage', function(e) {
          if (e.key !== 'vaultr.drawer' || !e.newValue || self.drawerOpen) return;
          self._restore();
        });
      },

      // Open an existing note in the drawer.
      async openNoteInDrawer(path, title, isKnowledge, pinned, isIndex, canCompile) {
        if (!path) return;
        var prevTab = this.tabs[this.activeTab];
        if (this.drawerOpen && prevTab) await __vaultrDESaveTabForLeave(prevTab);
        this.upsertTab(path, title, isKnowledge, pinned, isIndex, canCompile);
        __vaultrDEResetCompileBtn();
        this.drawerOpen = true;
        this.$nextTick(function() { __vaultrDEScheduleActiveTabScroll(); });
        // Already loaded with unsaved edits — don't clobber with a server fetch
        if (__vaultrDE.currentPath === path && __vaultrDE.dirty) return;
        var tab = this.tabs[this.activeTab];
        var savedState = tab ? __vaultrDERestoreTabState(tab.id) : null;
        await __vaultrDrawerLoadNote(path, tab ? tab.id : null, savedState);
      },

      // Open the drawer in create mode (new note or imported file).
      async openNewInDrawer(content, suggestedName) {
        var s = __vaultrDE;
        var prevTab = this.tabs[this.activeTab];
        if (this.drawerOpen && prevTab) await __vaultrDESaveTabForLeave(prevTab);
        if (s.dirty && s.currentPath) { clearTimeout(s.saveTimer); s.saveTimer = null; await __vaultrDEDoSave(); }
        else { clearTimeout(s.saveTimer); s.saveTimer = null; }
        __vaultrDESaveStatus('');

        var rawName = suggestedName || '';
        var title = rawName ? rawName.replace(/\.md$/i,'').split('/').pop() || 'New note' : 'New note';
        var normalizedContent = __vaultrDETightenLists(content || '');
        var newTab = {id:__vaultrDENewTabId(), title:title, path:'', isKnowledge:false, pinned:false,
                      draftId:__vaultrDENewDraftId(), _draftContent:normalizedContent, _pathVal: rawName,
                      createdAt: Date.now()};
        this.tabs.push(newTab);
        this.activeTab = this.tabs.length - 1;
        s.currentPath = ''; s.currentDraftId = newTab.draftId; s.currentMd = normalizedContent; s.dirty = false;
        var initialPi = document.getElementById('drawer-path-input');
        if (initialPi) initialPi.value = rawName;
        await __vaultrDEFlushDraft(newTab);
        
        var MAX_TABS = 10;
        if (this.tabs.length > MAX_TABS) {
          var oldestIdx = -1, oldestId = Infinity;
          for (var j = 0; j < this.tabs.length; j++) {
            if (j !== this.activeTab && this.tabs[j].path && this.tabs[j].id < oldestId)
              { oldestIdx = j; oldestId = this.tabs[j].id; }
          }
          if (oldestIdx >= 0) { 
            var evictedTab = this.tabs[oldestIdx];
            __vaultrDEClearTabState(evictedTab.id);
            this.tabs.splice(oldestIdx,1); 
            if (oldestIdx < this.activeTab) this.activeTab--; 
          }
        }
        this.drawerOpen = true;
        this.$nextTick(function() { __vaultrDEScheduleActiveTabScroll(); });

        s.currentPath = ''; s.currentDraftId = newTab.draftId; s.currentMd = normalizedContent; s.dirty = false;
        
        // Use the new state application method
        await __vaultrDEApplyState({ content: normalizedContent, inSource: false, scrollTop: 0 }, newTab.id);

        setTimeout(function() {
          var pi = document.getElementById('drawer-path-input');
          if (pi) { pi.value = rawName; pi.classList.remove('invalid'); pi.placeholder='filename.md  ·  or  /folder/note.md'; __vaultrDEPathAc && __vaultrDEPathAc.close(); }
          focusManager.focusPathInput();
        }, 0);
      },

      upsertTab(path, title, isKnowledge, pinned, isIndex, canCompile) {
        var idx = -1;
        for (var i = 0; i < this.tabs.length; i++) { if (this.tabs[i].path === path) { idx=i; break; } }
        if (idx >= 0) {
          this.tabs[idx].title = title || this.tabs[idx].title;
          if (isKnowledge !== undefined) this.tabs[idx].isKnowledge = !!isKnowledge;
          if (pinned !== undefined) this.tabs[idx].pinned = !!pinned;
          if (isIndex !== undefined) this.tabs[idx].isIndex = !!isIndex;
          if (canCompile !== undefined) this.tabs[idx].canCompile = !!canCompile;
          this.activeTab = idx;
        } else {
          this.tabs.push({id:__vaultrDENewTabId(), title:title||'Note', path:path, isKnowledge:!!isKnowledge, pinned:!!pinned, isIndex:!!isIndex, canCompile:!!canCompile});
          this.activeTab = this.tabs.length - 1;
          var MAX_TABS = 10;
          if (this.tabs.length > MAX_TABS) {
            var oldestIdx=-1, oldestId=Infinity;
            for (var j=0; j<this.tabs.length; j++) {
              if (j !== this.activeTab && this.tabs[j].path && this.tabs[j].id < oldestId) { oldestIdx=j; oldestId=this.tabs[j].id; }
            }
            if (oldestIdx >= 0) {
              var evicted = this.tabs[oldestIdx];
              __vaultrDEClearTabState(evicted.id);
              this.tabs.splice(oldestIdx,1);
              if (oldestIdx < this.activeTab) this.activeTab--;
            }
          }
        }
      },

      markTabCompiled(path) {
        for (var i = 0; i < this.tabs.length; i++) {
          if (this.tabs[i].path === path) { this.tabs[i].canCompile = false; return; }
        }
      },

      // Activate a draft tab: load draft, populate path input, focus it.
      // savedState comes from tabStateManager.restore() — pass null if unavailable.
      async _activateDraftTab(tab, savedState) {
        var draft = await __vaultrDELoadDraft(tab);
        if (!__vaultrDEIsActiveTabId(tab.id)) return;
        var draftState = __vaultrDEDraftEditorState(draft, savedState);
        await __vaultrDrawerSetContent(draftState.content, tab.id, draftState, tab.draftId);
        setTimeout(function() {
          if (!__vaultrDEIsActiveTabId(tab.id)) return;
          var pi = document.getElementById('drawer-path-input');
          if (pi) { pi.value = draft.pathInput || ''; pi.classList.remove('invalid'); pi.placeholder = 'filename.md  ·  or  /folder/note.md'; }
          focusManager.focusPathInput();
        }, 0);
      },

      async drawerSwitchTab(i) {
        focusManager.blurActive();
        if (i === this.activeTab) return;

        var prevTab = this.tabs[this.activeTab];
        var nextTab = this.tabs[i];
        if (!nextTab) return;
        
        // Save current tab's state
        if (prevTab) {
          await __vaultrDESaveTabForLeave(prevTab);
          if (!prevTab.path) __vaultrDEPathAc && __vaultrDEPathAc.close();
        }
        
        // Switch active tab
        this.activeTab = i;
        __vaultrDEResetCompileBtn();

        // Load next tab's content with saved state
        var savedState = __vaultrDERestoreTabState(nextTab.id);
        
        if (!nextTab.path) {
          await this._activateDraftTab(nextTab, savedState);
        } else {
          await __vaultrDrawerLoadNote(nextTab.path, nextTab.id, savedState);
        }
      },

      async drawerCloseTab(i) {
        if (i < 0 || i >= this.tabs.length) return;
        var wasActive = (i === this.activeTab);
        var closingTab = this.tabs[i];
        var closingCreate = !closingTab.path;
        if (wasActive && closingTab.path && __vaultrDE.dirty && __vaultrDE.currentPath === closingTab.path) {
          clearTimeout(__vaultrDE.saveTimer); __vaultrDE.saveTimer = null;
          await __vaultrDEDoSave();
          var liveIdx = this.tabs.indexOf(closingTab);
          if (liveIdx < 0) return;
          i = liveIdx;
          wasActive = (i === this.activeTab);
        }
        
        // Clear saved state for this tab
        __vaultrDEClearTabState(closingTab.id);
        if (closingCreate) void __vaultrDEDeleteDraft(closingTab);
        
        this.tabs.splice(i, 1);
        if (closingCreate && wasActive) __vaultrDEPathAc && __vaultrDEPathAc.close();
        
        if (this.tabs.length === 0) {
          this.drawerOpen = false; this.activeTab = -1;
          clearTimeout(__vaultrDE.saveTimer); __vaultrDE.saveTimer = null;
          __vaultrDE.currentPath = ''; __vaultrDE.currentDraftId = ''; __vaultrDE.currentMd = ''; __vaultrDE.dirty = false;
          __vaultrDESaveStatus('');
          if (__vaultrDE.view) {
            __vaultrDE.loading = true;
            // Clearing to empty removes the frontmatter range too — needs
            // the same escape hatch as syncContent() above, or a note with
            // frontmatter left behind a leftover (blocked) suppressed range.
            __vaultrDE.view.dispatch({
              changes: {from: 0, to: __vaultrDE.view.state.doc.length, insert: ''},
              annotations: __vaultrDE.allowFrontmatterEdit ? __vaultrDE.allowFrontmatterEdit.of(true) : undefined,
            });
            setTimeout(function(){ __vaultrDE.loading = false; }, 50);
          }
          return;
        }
        
        if (i < this.activeTab) { 
          this.activeTab -= 1; 
        } else if (wasActive) {
          this.activeTab = Math.min(i, this.tabs.length-1);
          this.$nextTick(function() { __vaultrDEScheduleActiveTabScroll(); });
          var t = this.tabs[this.activeTab]; 
          if (!t) return;
          
          var savedState = __vaultrDERestoreTabState(t.id);
          
          if (t.path) {
            void __vaultrDrawerLoadNote(t.path, t.id, savedState);
          } else {
            void this._activateDraftTab(t, savedState);
          }
        }
      },

      discardNewNote() {
        var i = this.activeTab;
        if (i < 0 || !this.tabs[i] || this.tabs[i].path) return;
        void this.drawerCloseTab(i);
      },

      async togglePinActiveNote() {
        var tab = this.tabs[this.activeTab];
        if (!tab || !tab.path) return;
        if (tab.pinned) await this.unpinActiveNote(); else await this.pinActiveNote();
      },

      async pinActiveNote() {
        var tab = this.tabs[this.activeTab]; if (!tab || !tab.path || tab.pinned) return;
        var resp = await fetch('/api/vault/pin', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({path:tab.path,pinned:true})});
        if (resp.status === 409) { window.showError((await resp.text()).trim() || 'Pinned notes limit reached.', 'Cannot pin'); return; }
        if (!resp.ok) return;
        tab.pinned = true;
        if (window.__vaultrAfterVaultMutation) await window.__vaultrAfterVaultMutation();
      },

      async unpinActiveNote() {
        var tab = this.tabs[this.activeTab]; if (!tab || !tab.path || !tab.pinned) return;
        var resp = await fetch('/api/vault/pin', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({path:tab.path,pinned:false})});
        if (!resp.ok) return;
        tab.pinned = false;
        if (window.__vaultrAfterVaultMutation) await window.__vaultrAfterVaultMutation();
      },

      async deleteActiveNote() {
        var tab = this.tabs[this.activeTab]; if (!tab || !tab.path) return;
        var confirmed = await window.showConfirm({titleHTML:'<span class="confirm-title-icon"><svg width="13" height="13" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 4h10M5 4V3a1 1 0 011-1h2a1 1 0 011 1v1M12 4l-1 8H3L2 4"/><path d="M6 7v3M8 7v3"/></svg></span>Delete note',message:'"'+tab.title+'" will be permanently deleted.',confirmLabel:'Delete',danger:true});
        if (!confirmed) return;
        var reqBody = {path:tab.path};
        var resp = await fetch('/api/vault/delete', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(reqBody)});
        if (!resp.ok) return;
        clearTimeout(__vaultrDE.saveTimer); __vaultrDE.saveTimer = null; __vaultrDE.dirty = false;
        var cur = this.activeTab;
        var deletedTab = this.tabs[cur];
        if (deletedTab) __vaultrDEClearTabState(deletedTab.id);
        this.tabs.splice(cur, 1);
        if (this.tabs.length === 0) {
          this.drawerOpen = false; this.activeTab = -1;
          __vaultrDE.currentPath = ''; __vaultrDE.currentDraftId = ''; __vaultrDE.currentMd = ''; __vaultrDESaveStatus('');
          if (__vaultrDE.view) {
            __vaultrDE.loading = true;
            // Clearing to empty removes the frontmatter range too — needs
            // the same escape hatch as syncContent() above, or a note with
            // frontmatter left behind a leftover (blocked) suppressed range.
            __vaultrDE.view.dispatch({
              changes: {from: 0, to: __vaultrDE.view.state.doc.length, insert: ''},
              annotations: __vaultrDE.allowFrontmatterEdit ? __vaultrDE.allowFrontmatterEdit.of(true) : undefined,
            });
            setTimeout(function(){ __vaultrDE.loading=false; },50);
          }
        } else {
          this.activeTab = cur > 0 ? cur-1 : 0;
          this.$nextTick(function() { __vaultrDEScheduleActiveTabScroll(); });
          var nextTab = this.tabs[this.activeTab];
          if (nextTab && nextTab.path) void __vaultrDrawerLoadNote(nextTab.path, nextTab.id, __vaultrDERestoreTabState(nextTab.id));
          else if (nextTab) {
            void this._activateDraftTab(nextTab, __vaultrDERestoreTabState(nextTab.id));
          }
        }
        if (window.__vaultrAfterVaultMutation) await window.__vaultrAfterVaultMutation();
      },
    };
  }

  // ── Compile raw note from drawer ─────────────────────────────────────────────
  function __vaultrDECompileLabel(btn, text) {
    var label = btn.querySelector('.drawer-compile-label');
    if (label) label.textContent = text;
  }
  function __vaultrDECloseMoreMenu() {
    document.dispatchEvent(new CustomEvent('drawer:close-more'));
  }
  function __vaultrDEResetCompileBtn() {
    var btn = document.querySelector('.drawer-compile-btn');
    if (!btn) return;
    btn.classList.remove('is-compiling', 'success');
    btn.disabled = false;
    btn.title = 'Compile to knowledge note';
    __vaultrDECompileLabel(btn, 'Compile');
  }

  async function compileDrawerNote(event) {
    var btn = event && event.currentTarget;
    if (!btn || btn.disabled) return;
    var drawer = window.__vaultrDrawer;
    var tab = drawer ? drawer.tabs[drawer.activeTab] : null;
    if (!tab || !tab.path || !tab.canCompile) return;
    var rawPath = tab.path;

    if (__vaultrDE.dirty && __vaultrDE.currentPath === rawPath) {
      clearTimeout(__vaultrDE.saveTimer);
      __vaultrDE.saveTimer = null;
      await __vaultrDEDoSave();
    }

    var originalTitle = btn.title;
    btn.disabled = true;
    btn.classList.add('is-compiling');
    btn.title = 'Compiling…';
    __vaultrDECompileLabel(btn, 'Compiling…');

    try {
      var resp = await fetch('/api/compile/trigger', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({path: rawPath}),
      });
      if (!resp.ok && resp.status !== 409) {
        var msg = 'Compile failed';
        try { var data = await resp.json(); if (data && data.error) msg = data.error; } catch (_) {}
        throw new Error(msg);
      }
      if (resp.status === 409) {
        btn.disabled = false;
        btn.title = originalTitle;
        __vaultrDECompileLabel(btn, 'Compile');
        return;
      }

      var pollURL = '/api/runs/by-ref?path=' + encodeURIComponent(rawPath);
      var deadline = Date.now() + 10 * 60 * 1000;
      await new Promise(function(r) { setTimeout(r, 600); });
      while (Date.now() < deadline) {
        var pr = await fetch(pollURL);
        if (pr.ok) {
          var st = await pr.json();
          if (st.status === 'succeeded') {
            if (drawer && drawer.markTabCompiled) drawer.markTabCompiled(rawPath);
            btn.classList.add('success');
            btn.title = 'Compiled';
            __vaultrDECompileLabel(btn, 'Compiled');
            setTimeout(function() {
              btn.disabled = false;
              btn.classList.remove('success');
              __vaultrDECompileLabel(btn, 'Compile');
              __vaultrDECloseMoreMenu();
            }, 1500);
            if (window.__vaultrAfterVaultMutation) await window.__vaultrAfterVaultMutation();
            return;
          }
          if (st.status === 'failed' || st.status === 'canceled') {
            throw new Error('Compile agent ' + st.status);
          }
        }
        await new Promise(function(r) { setTimeout(r, 1500); });
      }
      throw new Error('Compile timed out');
    } catch (err) {
      btn.disabled = false;
      btn.title = err && err.message ? err.message : 'Compile failed';
      __vaultrDECompileLabel(btn, 'Compile');
    } finally {
      btn.classList.remove('is-compiling');
    }
  }

  // Global undo/redo called by Electron main process via executeJavaScript.
  // Calls CodeMirror undo directly, bypassing browser native undo.
  window.__vaultrUndo = function() {
    var s = __vaultrDE;
    if (s.view && s.cmUndo) s.cmUndo(s.view);
  };
  window.__vaultrRedo = function() {
    var s = __vaultrDE;
    if (s.view && s.cmRedo) s.cmRedo(s.view);
  };

  window.__vaultrEditorShellHref = function() {
    try {
      var seg = location.pathname.replace(/^\/+/,'').split('/')[0];
      return '/'+(seg||'home');
    } catch(_) { return '/home'; }
  };

  window.__vaultrHotkeys.register('drawer', 'e', function() {
    if (window.__vaultrDrawer) window.__vaultrDrawer.drawerOpen = !window.__vaultrDrawer.drawerOpen;
  });

  window.__vaultrHotkeys.register('new-note', 'n', function() {
    if (window.__vaultrDrawer) void window.__vaultrDrawer.openNewInDrawer('', '');
  });

  window.__vaultrHotkeys.registerRaw('drawer-scroll', function(e, mod) {
    if (e.key !== 'ArrowUp' && e.key !== 'ArrowDown') return;
    if (mod || e.altKey) return;
    var ae = document.activeElement;
    var tag = ae && ae.tagName;
    if (tag === 'INPUT' || tag === 'TEXTAREA') return;
    var _drawer = window.__vaultrDrawer;
    if (!_drawer || !_drawer.drawerOpen) return;
    var _editArea = document.getElementById('drawer-edit-area');
    if (_editArea && _editArea.contains(ae)) return;
    e.preventDefault();
    var _scroller = document.querySelector('#drawer-edit-area .cm-scroller');
    if (_scroller) _scroller.scrollBy(0, e.key === 'ArrowDown' ? 80 : -80);
    return true;
  });

  window.__vaultrHotkeys.register('drawer-focus', '\\', function() {
    if (window.__vaultrDrawer && window.__vaultrDrawer.drawerOpen) {
      window.__vaultrDrawer.drawerExpanded = !window.__vaultrDrawer.drawerExpanded;
    }
  });

  window.__vaultrHotkeys.registerRaw('drawer-close-tab', function(e, mod) {
    if (!mod || e.shiftKey || e.altKey || e.key.toLowerCase() !== 'w') return;
    var drawer = window.__vaultrDrawer;
    if (!drawer || !drawer.drawerOpen || drawer.activeTab < 0) return;
    e.preventDefault();
    var at = drawer.tabs[drawer.activeTab];
    if (at && at.path) drawer.drawerCloseTab(drawer.activeTab);
    return true;
  });

  window.__vaultrHotkeys.registerRaw('drawer-find', function(e, mod) {
    if (!mod || e.shiftKey || e.altKey || e.key.toLowerCase() !== 'f') return;
    var _fDrawer = window.__vaultrDrawer;
    if (!_fDrawer || !_fDrawer.drawerOpen) return;
    var _de = __vaultrDE;
    if (!_de.cmOpenSearchPanel || !_de.view) return;
    e.preventDefault();
    _de.cmOpenSearchPanel(_de.view);
    return true;
  });

  // Mod-T toggles the selection-format tooltip (see tooltip.js) — replaces
  // the old auto-popup-on-select behavior, which read as too noisy.
  window.__vaultrHotkeys.registerRaw('drawer-format-tooltip', function(e, mod) {
    if (!mod || e.shiftKey || e.altKey || e.key.toLowerCase() !== 't') return;
    var _drawer = window.__vaultrDrawer;
    if (!_drawer || !_drawer.drawerOpen) return;
    var ea = document.getElementById('drawer-edit-area');
    var ae = document.activeElement;
    if (!ea || !ae || !ea.contains(ae)) return;
    if (!window.__vaultrToggleFormatTooltip) return;
    e.preventDefault();
    window.__vaultrToggleFormatTooltip();
    return true;
  });

  // Intercept mouse back/forward buttons (button 3/4) to switch drawer tabs.
  window.addEventListener('mousedown', function(e) {
    var _drawer = window.__vaultrDrawer;
    if (!_drawer || !_drawer.drawerOpen || _drawer.tabs.length <= 1) return;
    if (e.button === 3) {
      e.preventDefault();
      if (_drawer.activeTab > 0) void _drawer.drawerSwitchTab(_drawer.activeTab - 1);
    } else if (e.button === 4) {
      e.preventDefault();
      if (_drawer.activeTab < _drawer.tabs.length - 1) void _drawer.drawerSwitchTab(_drawer.activeTab + 1);
    }
  }, true);

