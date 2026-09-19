  // ── Editor state ────────────────────────────────────────────────────────────
  // One CodeMirror 6 EditorView (s.view) for the whole editor now — the old
  // Milkdown (WYSIWYG) / CodeMirror (source) split is gone. "inSource"/
  // editorMode still exist (see below) but now mean "decorations
  // Compartment reconfigured to plain syntax highlighting" vs "live-preview
  // decorations active", toggled in place on the same view/doc instead of
  // switching between two separately-mounted editors.
  var __vaultrEditor = {
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

  var CONTENT_PANE_CREATE_KEY = 'vaultr.content-pane-create';
  var __vaultrEditorTabSeq = 0;
  function __vaultrEditorNewTabId() {
    __vaultrEditorTabSeq = (__vaultrEditorTabSeq + 1) % 1000;
    return Date.now() * 1000 + __vaultrEditorTabSeq;
  }

  // ── Autocomplete (path input in create mode) ─────────────────────────────────
  var __CONTENT_PANE_PATH_DBL_ENTER_MS = 2000;
  var __vaultrEditorPathAc = null; // created in __vaultrEditorSetupCreateMode

  function __vaultrEditorAcParseCtx(val, caret) {
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
  function __vaultrEditorSaveStatus(txt) {
    var el = document.getElementById('content-pane-save-status');
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
  function __vaultrEditorClearPendingBaselineSync() {
    var s = __vaultrEditor;
    s.pendingBaselineFromEditor = false;
    if (s.pendingBaselineTimer) { clearTimeout(s.pendingBaselineTimer); s.pendingBaselineTimer = null; }
  }
  function __vaultrEditorMarkPendingBaselineSync() {
    var s = __vaultrEditor;
    s.pendingBaselineFromEditor = true;
    if (s.pendingBaselineTimer) clearTimeout(s.pendingBaselineTimer);
    s.pendingBaselineTimer = setTimeout(function() {
      s.pendingBaselineTimer = null;
      s.pendingBaselineFromEditor = false;
    }, 500); // Milkdown debounces markdownUpdated at 200ms; 500ms gives safe margin
  }
  function __vaultrEditorScheduleSave() {
    __vaultrEditorSaveStatus('●');
    clearTimeout(__vaultrEditor.saveTimer);
    __vaultrEditor.saveTimer = setTimeout(__vaultrEditorDoSave, 800);
  }
  async function __vaultrEditorDoSave() {
    if (!__vaultrEditor.dirty) return;
    var path = __vaultrEditor.currentPath;
    var content = __vaultrEditor.currentMd;
    if (!path) return;
    if (content === __vaultrEditor.baselineMd) {
      __vaultrEditor.dirty = false;
      clearTimeout(__vaultrEditor.saveTimer); __vaultrEditor.saveTimer = null;
      __vaultrEditorSaveStatus('');
      return;
    }
    try {
      var r = await fetch('/api/vault/write', {
        method: 'POST', headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({path: path, content: content}),
      });
      if (r.ok) {
        __vaultrEditor.dirty = false; __vaultrEditor.baselineMd = content; __vaultrEditorSaveStatus('Saved');
      } else {
        var errText = ''; try { errText = await r.text(); } catch(_) {}
        window.showError(errText || 'Server error — your changes may not be saved.', 'Save error');
      }
    } catch(e) {
      window.showError((e && e.message) || 'Network error — your changes may not be saved.', 'Save error');
    }
  }

  // ── Misc helpers ─────────────────────────────────────────────────────────────
  function __vaultrEditorTightenLists(md) {
    return md.replace(
      /(^[ \t]*(?:[-*+]|\d+[.)]) [^\n]*)\n\n(?=[ \t]*(?:[-*+]|\d+[.)]) )/gm, '$1\n');
  }
  function __vaultrEditorFindImageFile(dt) {
    if (!dt) return null;
    var items = Array.from(dt.items || []);
    for (var i = 0; i < items.length; i++) {
      if (items[i].kind === 'file' && items[i].type.indexOf('image/') === 0) return items[i].getAsFile();
    }
    return null;
  }
  async function __vaultrEditorUploadImage(imgFile) {
    var fd = new FormData();
    fd.append('file', imgFile);
    var resp = await fetch('/api/vault/upload-image', {method: 'POST', body: fd});
    if (!resp.ok) throw new Error(await resp.text());
    return (await resp.json()).src;
  }

  // ── Electron draft store helpers ────────────────────────────────────────────
  function __vaultrEditorDraftStore() {
    return window.vaultrDesktop && window.vaultrDesktop.drafts ? window.vaultrDesktop.drafts : null;
  }
  function __vaultrEditorNewDraftId() {
    return 'draft-' + Date.now().toString(36) + '-' + Math.random().toString(36).slice(2, 10);
  }
  function __vaultrEditorActiveTab() {
    var pane = window.__vaultrContentPane;
    return pane ? pane.tabs[pane.activeTab] : null;
  }
  function __vaultrEditorPathInputValue() {
    var pi = document.getElementById('content-pane-path-input');
    return pi ? pi.value : '';
  }
  function __vaultrEditorDraftTitle(pathInput, content) {
    var name = (pathInput || '').trim();
    if (name) return name.replace(/\.md$/i, '').split('/').pop() || 'New note';
    var first = String(content || '').split(/\r?\n/).map(function(line) {
      return line.replace(/^#+\s*/, '').trim();
    }).find(Boolean);
    return first ? first.slice(0, 48) : 'New note';
  }
  function __vaultrEditorEnsureDraftId(tab) {
    if (!tab || tab.path) return '';
    if (!tab.draftId) {
      tab.draftId = __vaultrEditorNewDraftId();
      var pane = window.__vaultrContentPane;
      if (pane) pane._persist();
    }
    return tab.draftId;
  }
  function __vaultrEditorIsActiveTab(tab) {
    var pane = window.__vaultrContentPane;
    return !!(pane && tab && pane.tabs[pane.activeTab] === tab);
  }
  function __vaultrEditorIsActiveTabId(tabId) {
    if (!tabId) return true;
    var pane = window.__vaultrContentPane;
    var tab = pane && pane.tabs[pane.activeTab];
    return !!(tab && tab.id === tabId);
  }
  function __vaultrEditorGetScrollState() {
    var s = __vaultrEditor;
    var scroller = document.querySelector('#content-pane-edit-area .cm-scroller');
    return {scrollTop: scroller ? scroller.scrollTop : 0, inSource: s.inSource};
  }
  function __vaultrEditorCaptureDraft(tab) {
    if (!tab || tab.path) return null;
    var pane = window.__vaultrContentPane;
    var active = __vaultrEditorIsActiveTab(tab);
    var live = !!(active && pane && pane.contentPaneOpen && __vaultrEditor.currentDraftId === tab.draftId);
    var content = live ? __vaultrEditor.currentMd : (tab._draftContent || tab.draftContent || '');
    var pathInput = live ? __vaultrEditorPathInputValue() : (tab._pathVal || '');
    var scroll = active ? __vaultrEditorGetScrollState() : (__vaultrEditorRestoreTabState(tab.id) || {});
    tab._draftContent = content || '';
    tab._pathVal = pathInput || '';
    tab.title = __vaultrEditorDraftTitle(tab._pathVal, tab._draftContent);
    return {
      version: 1,
      draftId: __vaultrEditorEnsureDraftId(tab),
      content: tab._draftContent,
      pathInput: tab._pathVal,
      title: tab.title,
      mode: scroll.inSource ? 'source' : 'wysiwyg',
      scrollTop: scroll.scrollTop || 0,
      createdAt: tab.createdAt || Date.now(),
    };
  }
  function __vaultrEditorClearDraftTimer(tab) {
    if (!__vaultrEditor.draftSaveTimer) return;
    if (!tab || __vaultrEditor.draftSaveTabId === tab.id) {
      clearTimeout(__vaultrEditor.draftSaveTimer);
      __vaultrEditor.draftSaveTimer = null;
      __vaultrEditor.draftSaveTabId = null;
    }
  }
  async function __vaultrEditorFlushDraft(tab) {
    if (!tab || tab.path) return;
    __vaultrEditorClearDraftTimer(tab);
    var store = __vaultrEditorDraftStore();
    if (!store || !store.write) return;
    var data = __vaultrEditorCaptureDraft(tab);
    if (!data || !data.draftId) return;
    try {
      await store.write(data.draftId, data);
      tab.createdAt = data.createdAt;
    } catch(e) {
      console.warn('draft write failed', e);
    }
  }
  async function __vaultrEditorSaveTabForLeave(tab) { await tabStateManager.saveForLeave(tab); }
  function __vaultrEditorScheduleDraftSave(tab) {
    if (!tab || tab.path) return;
    __vaultrEditorCaptureDraft(tab);
    __vaultrEditorClearDraftTimer(tab);
    __vaultrEditor.draftSaveTabId = tab.id;
    __vaultrEditor.draftSaveTimer = setTimeout(function() {
      __vaultrEditor.draftSaveTimer = null;
      __vaultrEditor.draftSaveTabId = null;
      void __vaultrEditorFlushDraft(tab);
    }, 350);
  }
  function __vaultrEditorScheduleActiveDraftSave() {
    var tab = __vaultrEditorActiveTab();
    if (tab && !tab.path) __vaultrEditorScheduleDraftSave(tab);
  }
  async function __vaultrEditorLoadDraft(tab) {
    if (!tab || tab.path) return {content:'', pathInput:'', inSource:false, scrollTop:0};
    var store = __vaultrEditorDraftStore();
    var content = tab._draftContent || tab.draftContent || '';
    var pathInput = tab._pathVal || '';
    var loaded = null;
    __vaultrEditorEnsureDraftId(tab);
    if (store && store.read) {
      try { loaded = await store.read(tab.draftId); } catch(_) { loaded = null; }
    }
    if (loaded) {
      content = typeof loaded.content === 'string' ? loaded.content : '';
      pathInput = typeof loaded.pathInput === 'string' ? loaded.pathInput : (loaded.path || '');
      tab.createdAt = loaded.createdAt || tab.createdAt || Date.now();
      tab.title = loaded.title || __vaultrEditorDraftTitle(pathInput, content);
    } else {
      tab.createdAt = tab.createdAt || Date.now();
      tab.title = __vaultrEditorDraftTitle(pathInput, content);
    }
    tab._draftContent = __vaultrEditorTightenLists(content || '');
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
  function __vaultrEditorDraftEditorState(draft, savedState) {
    return {
      content: draft && typeof draft.content === 'string' ? draft.content : '',
      inSource: savedState && typeof savedState.inSource === 'boolean' ? savedState.inSource : !!(draft && draft.inSource),
      scrollTop: savedState && typeof savedState.scrollTop === 'number' ? savedState.scrollTop : ((draft && draft.scrollTop) || 0),
    };
  }
  async function __vaultrEditorDeleteDraft(tab) {
    if (!tab || !tab.draftId) return;
    __vaultrEditorClearDraftTimer(tab);
    var store = __vaultrEditorDraftStore();
    var id = tab.draftId;
    tab.draftId = '';
    tab._draftContent = '';
    tab._pathVal = '';
    if (__vaultrEditor.currentDraftId === id) __vaultrEditor.currentDraftId = '';
    if (store && store.delete) {
      try { await store.delete(id); } catch(e) { console.warn('draft delete failed', e); }
    }
  }
  function __vaultrEditorHandleContentChange(md, tighten) {
    var s = __vaultrEditor;
    var pane = window.__vaultrContentPane;
    if (pane && !pane.contentPaneOpen) return;
    var next = tighten ? __vaultrEditorTightenLists(md) : md;
    var tab = __vaultrEditorActiveTab();
    if (tab && tab.path && s.pendingBaselineFromEditor && tighten) {
      s.pendingBaselineFromEditor = false;
      if (s.pendingBaselineTimer) { clearTimeout(s.pendingBaselineTimer); s.pendingBaselineTimer = null; }
      s.currentMd = next;
      s.baselineMd = next;
      s.dirty = false;
      clearTimeout(s.saveTimer); s.saveTimer = null;
      __vaultrEditorSaveStatus('');
      return;
    }
    s.currentMd = next;
    if (tab && !tab.path) {
      s.dirty = false;
      tab._draftContent = next;
      __vaultrEditorScheduleDraftSave(tab);
      return;
    }
    if (next !== s.baselineMd) {
      s.dirty = true;
      __vaultrEditorScheduleSave();
    } else {
      s.dirty = false;
      clearTimeout(s.saveTimer); s.saveTimer = null;
      __vaultrEditorSaveStatus('');
    }
  }

  // Open a wiki-link target in the pane.  value is the raw [[…]] inner text.
  async function __vaultrContentPaneOpenWikiLink(value) {
    var pane = window.__vaultrContentPane;
    if (!pane) return;
    // Path-like value (contains /): treat as vault-absolute path directly
    if (value.indexOf('/') !== -1) {
      var p = value.startsWith('/') ? value : '/' + value;
      if (!p.endsWith('.md')) p += '.md';
      await pane.openNoteInContentPane(p, p.split('/').pop().replace(/\.md$/, ''), false, false);
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
      await pane.openNoteInContentPane(notePath, noteTitle, _isK, !!note.pinned, _isI,
        !_isK && !_isI && !note.compile_count);
    } catch(_) {}
  }

  function __vaultrEditorSaveTabState(tabId) { tabStateManager.save(tabId); }
  function __vaultrEditorRestoreTabState(tabId) { return tabStateManager.restore(tabId); }
  function __vaultrEditorClearTabState(tabId) { tabStateManager.clear(tabId); }
  function __vaultrEditorSetSourceActive(active) {
    document.querySelectorAll('.content-pane-view-btn-wysiwyg').forEach(function(btn) {
      btn.classList.toggle('active', !active);
    });
    document.querySelectorAll('.content-pane-view-btn-source').forEach(function(btn) {
      btn.classList.toggle('active', !!active);
    });
    document.querySelectorAll('.content-pane-source-toggle-btn').forEach(function(btn) {
      btn.classList.toggle('active', !!active);
    });
  }

  // ── Lazy editor init ─────────────────────────────────────────────────────────
  async function __vaultrEnsureContentPaneEditor() {
    var s = __vaultrEditor;
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
      s.cmSearch = mod.search; s.cmOpenSearchPanel = mod.openSearchPanel;
      s.cmCloseSearchPanel = __vaultrEditorMakeAnimatedSearchClose(mod.closeSearchPanel);
      s.cmFindNext = mod.findNext; s.cmFindPrev = mod.findPrevious;
      s.cmReplaceNext = mod.replaceNext; s.cmReplaceAll = mod.cmReplaceAll;
      s.SearchQuery = mod.SearchQuery; s.getSearchQuery = mod.getSearchQuery; s.setSearchQuery = mod.setSearchQuery;
      s.livePreviewPlugin = mod.livePreviewPlugin; s.livePreviewAtomicRanges = mod.livePreviewAtomicRanges;
      s.livePreviewTheme = mod.livePreviewTheme; s.codeHighlightStyle = mod.codeHighlightStyle;
      s.wikiMarkdownLanguage = mod.wikiMarkdownLanguage; s.linkClickHandler = mod.linkClickHandler;
      s.horizontalRuleField = mod.horizontalRuleField;
      s.frontmatterCollapseField = mod.frontmatterCollapseField; s.frontmatterHeaderField = mod.frontmatterHeaderField;
      s.frontmatterReadOnly = mod.frontmatterReadOnly; s.allowFrontmatterEdit = mod.allowFrontmatterEdit;

      var editArea = document.getElementById('content-pane-edit-area');

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
        onWikiLinkClick: function(target) { void __vaultrContentPaneOpenWikiLink(target); },
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
      // singleton (__vaultrEnsureContentPaneEditor only ever creates it once),
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
            s.frontmatterHeaderField({ onEditFrontmatter: __vaultrEditorEditFrontmatter }),
            s.decoCompartment.of(s._liveModeExt()),
            s.EditorView.updateListener.of(function(update) {
              if (!update.docChanged || s.loading) return;
              __vaultrEditorHandleContentChange(update.state.doc.toString(), false);
            }),
            s.EditorView.domEventHandlers({
              paste: function(e, view) {
                var imgFile = __vaultrEditorFindImageFile(e.clipboardData);
                if (!imgFile) return false;
                e.preventDefault();
                __vaultrEditorUploadImage(imgFile).then(function(src) {
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

      document.querySelectorAll('.content-pane-view-btn-wysiwyg').forEach(function(btn) {
        btn.addEventListener('click', function() {
          if (__vaultrEditor.inSource) editorMode.exitSource();
        });
      });
      document.querySelectorAll('.content-pane-view-btn-source').forEach(function(btn) {
        btn.addEventListener('click', function() {
          if (!__vaultrEditor.inSource) editorMode.enterSource();
        });
      });
    })();
    return s.initPromise;
  }

  // ── Search panel (custom, top-anchored) ─────────────────────────────────────
  // CodeMirror's Panel API has no declarative x-show/x-transition
  // equivalent (mount() fires once, right after insertion; there's no
  // pre-removal hook), so the slide-in/out that matches the tabs/more
  // menus (content_pane.html, content_pane.css) is done by hand: mount() toggles
  // .is-entering on the .cm-panels-top wrapper CodeMirror already
  // created, and every close goes through this wrapper, which holds the
  // real close call until .is-closing's CSS transition (content_pane.css) has
  // had time to finish.
  function __vaultrEditorMakeAnimatedSearchClose(realClose) {
    return function(view) {
      var panel = document.querySelector('#content-pane-edit-area .cm-panels-top');
      if (!panel) { realClose(view); return; }
      panel.classList.add('is-closing');
      setTimeout(function() { realClose(view); }, 100);
    };
  }
  function __vaultrCreateSearchPanel(view) {
    var s = __vaultrEditor;
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
        var panel = dom.closest('.cm-panels-top');
        if (panel) {
          panel.classList.add('is-entering');
          requestAnimationFrame(function() {
            requestAnimationFrame(function() { panel.classList.remove('is-entering'); });
          });
        }
      },
    };
  }

  // ── Editor mode state machine ────────────────────────────────────────────────
  // Single authority for live-preview ↔ source transitions. Both modes are
  // the same EditorView/doc now — this only reconfigures s.decoCompartment
  // (see __vaultrEnsureContentPaneEditor) and, when the caller is about to show a
  // *different* note's content, syncs s.currentMd into the view first.
  // applySource / applyWysiwyg: low-level, called by applyState (skipFocus=true).
  // enterSource / exitSource / toggle: user-triggered, manage focus themselves.
  var editorMode = (function() {
    function syncContent() {
      var s = __vaultrEditor;
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
        var s = __vaultrEditor;
        __vaultrEditorClearPendingBaselineSync();
        s.loading = true;
        syncContent();
        s.view.dispatch({effects: s.decoCompartment.reconfigure(s._sourceModeExt())});
        s.loading = false;
        s.inSource = true;
        __vaultrEditorSetSourceActive(true);
        if (!(opts && opts.skipFocus)) focusManager.focusEditor();
      },
      applyWysiwyg: function(opts) {
        var s = __vaultrEditor;
        var tab = __vaultrEditorActiveTab();
        if (tab && tab.path) __vaultrEditorMarkPendingBaselineSync();
        s.loading = true;
        syncContent();
        s.view.dispatch({effects: s.decoCompartment.reconfigure(s._liveModeExt())});
        setTimeout(function() { s.loading = false; }, 50);
        s.inSource = false;
        __vaultrEditorSetSourceActive(false);
        if (!(opts && opts.skipFocus)) focusManager.focusEditor();
      },
      // User-triggered: live preview → source
      enterSource: function() { this.applySource(); },
      // User-triggered: source → live preview
      exitSource: function() {
        var s = __vaultrEditor;
        s.currentMd = s.view.state.doc.toString();
        this.applyWysiwyg();
      },
      toggle: function() {
        var s = __vaultrEditor;
        if (s.inSource) this.exitSource(); else this.enterSource();
      },
    };
  })();

  // ── Focus manager ────────────────────────────────────────────────────────────
  // Single authority for all editor focus/blur decisions.
  var focusManager = {
    focusEditor: function() {
      var s = __vaultrEditor;
      if (s.view) s.view.focus();
    },
    // Focus the path input (create-mode toolbar).
    focusPathInput: function() {
      var pi = document.getElementById('content-pane-path-input');
      if (pi) pi.focus();
    },
    // Blur whatever currently has focus.
    blurActive: function() {
      if (document.activeElement && document.activeElement.blur) document.activeElement.blur();
    },
    // True when focus is inside the editor content area.
    isInsideEditor: function() {
      var ae = document.activeElement;
      var ea = document.getElementById('content-pane-edit-area');
      return !!(ea && ea.contains(ae));
    },
  };

  function __vaultrEditorEnterSource() { editorMode.enterSource(); }
  function __vaultrEditorExitSource() { editorMode.exitSource(); }

  // "Metadata" header's pencil button (cm-live/frontmatter-collapse.js) —
  // frontmatter is read-only in live-preview mode (frontmatterReadOnly()),
  // so this dialog + allowFrontmatterEdit-annotated dispatch is the only
  // way to change it there. from/to span the whole node including the
  // "---" delimiters; only the interior YAML is shown/edited.
  function __vaultrEditorEditFrontmatter(view, from, to) {
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
        annotations: __vaultrEditor.allowFrontmatterEdit.of(true),
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
        var s = __vaultrEditor;
        var scroller = document.querySelector('#content-pane-edit-area .cm-scroller');
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
        if (!tab.path) await __vaultrEditorFlushDraft(tab);
      },
    };
  })();

  // ── Content loaders ──────────────────────────────────────────────────────────
  async function __vaultrContentPaneLoadNote(path, tabId, savedState) {
    var s = __vaultrEditor;
    __vaultrEditorClearPendingBaselineSync();
    if (s.dirty && s.currentPath && s.currentPath !== path) {
      clearTimeout(s.saveTimer); s.saveTimer = null; await __vaultrEditorDoSave();
    } else { clearTimeout(s.saveTimer); s.saveTimer = null; }
    if (!__vaultrEditorIsActiveTabId(tabId)) return false;
    __vaultrEditorSaveStatus('');
    var ptEl = document.getElementById('content-pane-path-text');
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
    if (!__vaultrEditorIsActiveTabId(tabId)) return false;
    s.currentPath = path; s.currentDraftId = ''; s.currentMd = __vaultrEditorTightenLists(content); s.baselineMd = s.currentMd; s.dirty = false;
    
    // Apply state (will create new state if no saved state)
    return await __vaultrEditorApplyState(savedState || { inSource: false, scrollTop: 0 }, tabId);
  }

  async function __vaultrContentPaneSetContent(content, tabId, savedState, draftId) {
    var s = __vaultrEditor;
    __vaultrEditorClearPendingBaselineSync();
    if (s.dirty && s.currentPath) { clearTimeout(s.saveTimer); s.saveTimer = null; await __vaultrEditorDoSave(); }
    else { clearTimeout(s.saveTimer); s.saveTimer = null; }
    if (!__vaultrEditorIsActiveTabId(tabId)) return false;
    __vaultrEditorSaveStatus('');
    s.currentPath = ''; s.currentDraftId = draftId || ''; s.currentMd = __vaultrEditorTightenLists(content || ''); s.baselineMd = ''; s.dirty = false;
    
    // Apply state
    return await __vaultrEditorApplyState(savedState || { inSource: false, scrollTop: 0 }, tabId);
  }
  
  async function __vaultrEditorApplyState(state, expectedTabId) {
    var s = __vaultrEditor;
    s.loading = true;
    await __vaultrEnsureContentPaneEditor();
    if (!__vaultrEditorIsActiveTabId(expectedTabId)) {
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
    var getScroller = function() { return document.querySelector('#content-pane-edit-area .cm-scroller'); };
    // Set it once synchronously, in the same tick as the content swap above
    // (dispatch() has already updated the DOM by the time this line runs) —
    // otherwise the scroller keeps the PREVIOUS tab's scrollTop for at
    // least one paint, since the rAF/setTimeout calls below are the only
    // other place this gets set and both are deliberately deferred (see
    // their own comments). That stale-scrollTop-against-new-content frame
    // is the tab-switch flash: this line is what removes it for the common
    // case; the deferred ones stay as-is for the one case that still needs
    // them — the pane's own slide-in animation tearing down a compositing
    // layer and resetting scrollTop out from under this synchronous set.
    var syncEl = getScroller(); if (syncEl) syncEl.scrollTop = targetScroll;
    s.pendingScrollRaf = requestAnimationFrame(function() {
      s.pendingScrollRaf = requestAnimationFrame(function() {
        s.pendingScrollRaf = null;
        if (!__vaultrEditorIsActiveTabId(expectedTabId)) return;
        var el = getScroller(); if (el) el.scrollTop = targetScroll;
        var tabForFocus = __vaultrEditorActiveTab();
        var isDraftTab = tabForFocus && !tabForFocus.path;
        if (!focusManager.isInsideEditor() && !isDraftTab) {
          focusManager.focusEditor();
        }
      });
    });
    s.pendingOpenScroll = setTimeout(function() {
      s.pendingOpenScroll = null;
      if (!__vaultrEditorIsActiveTabId(expectedTabId)) return;
      var el = getScroller(); if (el) el.scrollTop = targetScroll;
    }, 270);
    return true;
  }


  // ── Create mode: path input handlers + Publish ───────────────────────────────
  function __vaultrEditorSetupCreateMode() {
    var pi = document.getElementById('content-pane-path-input');
    var pb = document.getElementById('content-pane-publish-btn');
    if (!pi) return;

    // Build the autocomplete instance via the shared factory.
    var enterFirstTs = 0;
    __vaultrEditorPathAc = __vaultrPathAcCreate({
      getInput: function() { return document.getElementById('content-pane-path-input'); },
      getList:  function() { return document.getElementById('content-pane-path-ac'); },
      parseCtx: __vaultrEditorAcParseCtx,
      onApply: function(input, newVal, caretPos) {
        input.value = newVal;
        input.setSelectionRange(caretPos, caretPos);
        input.focus();
        input.classList.remove('invalid');
        input.placeholder = 'filename.md  ·  or  /folder/note.md';
        __vaultrEditorScheduleActiveDraftSave();
      },
      escKey: 'content-pane-ac',
    });

    pi.addEventListener('input', function() {
      enterFirstTs = 0;
      pi.classList.remove('invalid');
      pi.placeholder = 'filename.md  ·  or  /folder/note.md';
      __vaultrEditorPathAc.refresh();
      __vaultrEditorScheduleActiveDraftSave();
    });
    pi.addEventListener('click', function() { __vaultrEditorPathAc.refresh(); });
    pi.addEventListener('keyup', function(ev) {
      if (ev.key === 'ArrowLeft' || ev.key === 'ArrowRight' || ev.key === 'Home' || ev.key === 'End')
        __vaultrEditorPathAc.refresh();
    });
    pi.addEventListener('blur', function() {
      enterFirstTs = 0;
      setTimeout(function() {
        var acEl = document.getElementById('content-pane-path-ac');
        if (!acEl || !acEl.contains(document.activeElement)) __vaultrEditorPathAc.close();
      }, 180);
    });
    pi.addEventListener('keydown', function(ev) {
      if (__vaultrEditorPathAc.handleKeydown(ev)) return;
      if (ev.key !== 'Enter') return;
      var now = Date.now();
      if (enterFirstTs && (now - enterFirstTs) <= __CONTENT_PANE_PATH_DBL_ENTER_MS) {
        ev.preventDefault(); enterFirstTs = 0;
        var s = __vaultrEditor;
        if (s.view) s.view.focus();
        return;
      }
      enterFirstTs = now;
    });
    if (pb) pb.addEventListener('click', __vaultrContentPanePublish);
  }

  async function __vaultrContentPanePublish() {
    var pi = document.getElementById('content-pane-path-input');
    var pb = document.getElementById('content-pane-publish-btn');
    if (!pi || !pb) return;
    var pane = window.__vaultrContentPane;
    var tab = pane ? pane.tabs[pane.activeTab] : null;
    if (tab && !tab.path) await __vaultrEditorFlushDraft(tab);
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
        body: JSON.stringify({path: apiPath, content: __vaultrEditor.currentMd}),
      });
      if (!resp.ok) {
        var msg = (await resp.text()) || 'Publish failed.';
        window.showError(msg, 'Publish failed');
        pi.classList.add('invalid'); return;
      }
      published = true;
      __vaultrEditor.currentPath = apiPath; __vaultrEditor.dirty = false; __vaultrEditor.baselineMd = __vaultrEditor.currentMd; __vaultrEditorSaveStatus('Saved');
      __vaultrEditor.currentDraftId = '';
      __vaultrEditorPathAc && __vaultrEditorPathAc.close();
      if (pane) {
        if (tab && !tab.path) {
          var oldDraftId = tab.draftId;
          tab.path = apiPath;
          tab.title = apiPath.split('/').pop().replace(/\.md$/, '') || apiPath;
          tab.draftId = '';
          tab._draftContent = '';
          tab._pathVal = '';
          pane._persist();
          var draftStore = __vaultrEditorDraftStore();
          if (oldDraftId && draftStore && draftStore.delete)
            void draftStore.delete(oldDraftId).catch(function(){});
        }
      }
      var ptEl = document.getElementById('content-pane-path-text');
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
    var pane = window.__vaultrContentPane;
    if (pane) void pane.openNoteInContentPane(d.notePath, d.noteTitle || 'Note',
      d.noteIsKnowledge === 'true', d.notePinned === 'true', d.noteIsIndex === 'true',
      d.noteCanCompile === 'true');
  }

  // ── Content-pane controller factory ────────────────────────────────────────────────
  function contentPaneCtrl() {
    return {
      contentPaneOpen: false, tabs: [], activeTab: -1,
      splitRatio: 0.5, isPaneResizing: false, _prevSplitRatio: 0.5,

      // Tab-bar maximize/restore button — same spot/icon the old "focus
      // mode" toggle used, but it just drives splitRatio to/from 0 now
      // instead of a separate expanded state (the sidebar always stays
      // visible; see home.css's .content-pane-open rule).
      toggleMaximizeEditor() {
        if (this.splitRatio > 0.02) {
          this._prevSplitRatio = this.splitRatio;
          this.splitRatio = 0;
        } else {
          this.splitRatio = this._prevSplitRatio || 0.5;
        }
      },

      // Drags the divider between #home-list-pane and the editor pane.
      // Ratio is list-pane's share of the space left over once the
      // sidebar and resizer are accounted for (see home.css's
      // --split-ratio/--split-ratio-inv). Dragging all the way to the
      // sidebar collapses the list pane to 0 — the editor fills 100% of
      // the content area (not the whole window; the sidebar stays put).
      // Dragging the other way is capped so the editor never fully
      // disappears while still open.
      //
      // Perf: mousemove can fire far faster than the display refreshes,
      // and going through Alpine's reactive splitRatio on every one of
      // those events would re-run the whole :style expression (string
      // concat + reparse) well more often than needed. Instead this
      // writes the two CSS vars straight to the element, coalesced to one
      // update per animation frame — self.splitRatio (and its Alpine
      // side-effects: persistence, the maximize-button icon, etc.) is
      // synced exactly once, on mouseup. A full-viewport capture layer
      // owns the drag's mousemove/mouseup for the same reason a native
      // <input type=range> thumb does: without it, a fast drag can carry
      // the pointer over CodeMirror's contenteditable mid-gesture, which
      // would otherwise start a text selection instead of resizing.
      startPaneResize(e) {
        e.preventDefault();
        var self = this;
        var shell = document.querySelector('.home-shell');
        var side = document.getElementById('home-side');
        var resizer = document.getElementById('home-pane-resizer');
        if (!shell || !side || !resizer) return;
        var shellRect = shell.getBoundingClientRect();
        var sideW = side.offsetWidth;
        var resizerW = resizer.offsetWidth;
        var avail = shellRect.width - sideW - resizerW;
        if (avail <= 0) return;

        self.isPaneResizing = true;
        var prevCursor = document.body.style.cursor;
        var prevUserSelect = document.body.style.userSelect;
        document.body.style.cursor = 'col-resize';
        document.body.style.userSelect = 'none';

        var capture = document.createElement('div');
        capture.style.cssText = 'position:fixed;inset:0;z-index:9999;cursor:col-resize;';
        document.body.appendChild(capture);

        var latestRatio = self.splitRatio;
        var rafId = null;
        function applyRatio() {
          rafId = null;
          shell.style.setProperty('--split-ratio', latestRatio);
          shell.style.setProperty('--split-ratio-inv', 1 - latestRatio);
        }
        function onMove(ev) {
          var x = ev.clientX - shellRect.left - sideW - (resizerW / 2);
          var ratio = x / avail;
          latestRatio = Math.min(0.9, Math.max(0, ratio));
          if (rafId == null) rafId = requestAnimationFrame(applyRatio);
        }
        function onUp() {
          if (rafId != null) { cancelAnimationFrame(rafId); applyRatio(); }
          document.removeEventListener('mousemove', onMove);
          document.removeEventListener('mouseup', onUp);
          capture.remove();
          document.body.style.cursor = prevCursor;
          document.body.style.userSelect = prevUserSelect;
          self.isPaneResizing = false;
          self.splitRatio = latestRatio;
          try { localStorage.setItem('vaultr.splitRatio', String(latestRatio)); } catch(_) {}
        }
        document.addEventListener('mousemove', onMove);
        document.addEventListener('mouseup', onUp);
      },

      _persist() {
        try {
          localStorage.setItem('vaultr.content-pane', JSON.stringify({
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
          var raw = localStorage.getItem('vaultr.content-pane'); if (!raw) return;
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

      initContentPane() {
        window.__vaultrContentPane = this;
        this._restore();
        try {
          var savedRatio = parseFloat(localStorage.getItem('vaultr.splitRatio'));
          if (!isNaN(savedRatio) && savedRatio >= 0 && savedRatio <= 0.9) this.splitRatio = savedRatio;
        } catch(_) {}
        __vaultrEditorSetupCreateMode();

        var self = this; var prevOpen = false;
        // Esc still closes whatever's layered on top of the editor (the
        // more/tabs dropdowns, CodeMirror's search panel) — it just no
        // longer closes the editor itself once those are all closed.
        var _paneEscClose = function() {
          var moreMenu = document.querySelector('.content-pane-more-menu');
          if (moreMenu && moreMenu.style.display !== 'none') {
            document.dispatchEvent(new CustomEvent('content-pane:close-more'));
            focusManager.blurActive();
            return;
          }
          var tabsMenu = document.querySelector('.content-pane-tabs-menu');
          if (tabsMenu && tabsMenu.style.display !== 'none') {
            document.dispatchEvent(new CustomEvent('content-pane:close-tabs'));
            focusManager.blurActive();
            return;
          }
          if (document.querySelector('.vaultr-search-panel')) {
            var _s = __vaultrEditor;
            if (_s.view && _s.cmCloseSearchPanel) _s.cmCloseSearchPanel(_s.view);
            return;
          }
        };
        this.$watch('contentPaneOpen', async function(isOpen) {
          if (!isOpen) {
            if (window.__vaultrEscPop) window.__vaultrEscPop('content-pane');
            // Save state BEFORE blur — blur can trigger scrollIntoView which resets scrollTop
            var currentTab = self.tabs[self.activeTab];
            if (currentTab) await __vaultrEditorSaveTabForLeave(currentTab);
            focusManager.blurActive();
            prevOpen = false; return;
          }
          prevOpen = true;
          // The inbox detail panel shares this same pane (see content_pane.html's
          // inbox-detail-panel) — always one or the other.
          if (self.inboxSheetOpen) self.inboxSheetOpen = false;
          var _paneOverlayEl = document.querySelector('.content-pane');
          if (_paneOverlayEl) {
            _paneOverlayEl.classList.add('content-pane-is-opening');
            setTimeout(function() { _paneOverlayEl.classList.remove('content-pane-is-opening'); }, 320);
          }
          if (window.__vaultrEscPush) window.__vaultrEscPush('content-pane', _paneEscClose);
          // Refresh key-behavior config on every open so settings changes take
          // effect immediately without restarting the app. Called before any
          // content loading so all code paths (same note, new note, draft) pick
          // it up. No-op if the editor hasn't been created yet (handled later in
          // __vaultrEnsureContentPaneEditor's initPromise).
          // Always read latest tabs from localStorage before opening — other
          // WebContentsViews (same session, different JS context) may have
          // added tabs since this view last called _restore().
          self._restore();
          var tab = self.tabs[self.activeTab];
          if (!tab) return;

          var savedState = __vaultrEditorRestoreTabState(tab.id);
          
          if (!tab.path) {
            await self._activateDraftTab(tab, savedState);
            return;
          }
          if (__vaultrEditor.currentPath !== tab.path || !__vaultrEditor.dirty) {
            void __vaultrContentPaneLoadNote(tab.path, tab.id, savedState);
          } else if (savedState) {
            // Same note with unsaved edits — restore scroll only. Two passes:
            // 1) rAF: fire immediately so scroll looks right during slide-in animation
            // 2) setTimeout(250): fire after the 240ms CSS transition in case the browser
            //    resets scrollTop when the GPU compositing layer is torn down at animation end
            var s = __vaultrEditor;
            var sc = savedState.scrollTop || 0;
            function __applyContentPaneScroll() {
              var scroller = document.querySelector('#content-pane-edit-area .cm-scroller');
              if (scroller) scroller.scrollTop = sc;
            }
            if (s.pendingScrollRaf) { cancelAnimationFrame(s.pendingScrollRaf); s.pendingScrollRaf = null; }
            clearTimeout(s.pendingOpenScroll);
            s.pendingScrollRaf = requestAnimationFrame(function() {
              s.pendingScrollRaf = null;
              __applyContentPaneScroll();
              if (!focusManager.isInsideEditor()) {
                focusManager.focusEditor();
              }
            });
            s.pendingOpenScroll = setTimeout(function() {
              s.pendingOpenScroll = null;
              __applyContentPaneScroll();
            }, 250);
          }
        });
        this.$watch('tabs', function(){ self._persist(); });
        this.$watch('activeTab', function(i) {
          self._persist();
        });

        // Sync pane state from other section views (each section is a separate
        // WebContentsView with its own JS context; storage events cross view boundaries).
        window.addEventListener('storage', function(e) {
          if (e.key !== 'vaultr.content-pane' || !e.newValue || self.contentPaneOpen) return;
          self._restore();
        });
      },

      // Open an existing note in the pane.
      async openNoteInContentPane(path, title, isKnowledge, pinned, isIndex, canCompile) {
        if (!path) return;
        var prevTab = this.tabs[this.activeTab];
        if (this.contentPaneOpen && prevTab) await __vaultrEditorSaveTabForLeave(prevTab);
        this.upsertTab(path, title, isKnowledge, pinned, isIndex, canCompile);
        __vaultrEditorResetCompileBtn();
        this.contentPaneOpen = true;
        // Already loaded with unsaved edits — don't clobber with a server fetch
        if (__vaultrEditor.currentPath === path && __vaultrEditor.dirty) return;
        var tab = this.tabs[this.activeTab];
        var savedState = tab ? __vaultrEditorRestoreTabState(tab.id) : null;
        await __vaultrContentPaneLoadNote(path, tab ? tab.id : null, savedState);
      },

      // Open the pane in create mode (new note or imported file).
      async openNewInContentPane(content, suggestedName) {
        var s = __vaultrEditor;
        var prevTab = this.tabs[this.activeTab];
        if (this.contentPaneOpen && prevTab) await __vaultrEditorSaveTabForLeave(prevTab);
        if (s.dirty && s.currentPath) { clearTimeout(s.saveTimer); s.saveTimer = null; await __vaultrEditorDoSave(); }
        else { clearTimeout(s.saveTimer); s.saveTimer = null; }
        __vaultrEditorSaveStatus('');

        var rawName = suggestedName || '';
        var title = rawName ? rawName.replace(/\.md$/i,'').split('/').pop() || 'New note' : 'New note';
        var normalizedContent = __vaultrEditorTightenLists(content || '');
        var newTab = {id:__vaultrEditorNewTabId(), title:title, path:'', isKnowledge:false, pinned:false,
                      draftId:__vaultrEditorNewDraftId(), _draftContent:normalizedContent, _pathVal: rawName,
                      createdAt: Date.now()};
        this.tabs.push(newTab);
        this.activeTab = this.tabs.length - 1;
        s.currentPath = ''; s.currentDraftId = newTab.draftId; s.currentMd = normalizedContent; s.dirty = false;
        var initialPi = document.getElementById('content-pane-path-input');
        if (initialPi) initialPi.value = rawName;
        await __vaultrEditorFlushDraft(newTab);
        
        var MAX_TABS = 10;
        if (this.tabs.length > MAX_TABS) {
          var oldestIdx = -1, oldestId = Infinity;
          for (var j = 0; j < this.tabs.length; j++) {
            if (j !== this.activeTab && this.tabs[j].path && this.tabs[j].id < oldestId)
              { oldestIdx = j; oldestId = this.tabs[j].id; }
          }
          if (oldestIdx >= 0) {
            var evictedTab = this.tabs[oldestIdx];
            __vaultrEditorClearTabState(evictedTab.id);
            this.tabs.splice(oldestIdx,1);
            if (oldestIdx < this.activeTab) this.activeTab--;
          }
        }
        this._moveActiveTabToFront();
        this.contentPaneOpen = true;

        s.currentPath = ''; s.currentDraftId = newTab.draftId; s.currentMd = normalizedContent; s.dirty = false;
        
        // Use the new state application method
        await __vaultrEditorApplyState({ content: normalizedContent, inSource: false, scrollTop: 0 }, newTab.id);

        setTimeout(function() {
          var pi = document.getElementById('content-pane-path-input');
          if (pi) { pi.value = rawName; pi.classList.remove('invalid'); pi.placeholder='filename.md  ·  or  /folder/note.md'; __vaultrEditorPathAc && __vaultrEditorPathAc.close(); }
          focusManager.focusPathInput();
        }, 0);
      },

      // Keeps the tabs list ordered most-recently-opened-first, so the
      // tabs panel always shows whatever was just opened/switched to at
      // the top. Called with this.activeTab already pointing at the tab
      // to promote — a plain array move, since callers key everything
      // else off tab object references (nextTab/newTab/etc.), not index.
      _moveActiveTabToFront() {
        var i = this.activeTab;
        if (i <= 0 || i >= this.tabs.length) return;
        var tab = this.tabs.splice(i, 1)[0];
        this.tabs.unshift(tab);
        this.activeTab = 0;
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
          this.tabs.push({id:__vaultrEditorNewTabId(), title:title||'Note', path:path, isKnowledge:!!isKnowledge, pinned:!!pinned, isIndex:!!isIndex, canCompile:!!canCompile});
          this.activeTab = this.tabs.length - 1;
          var MAX_TABS = 10;
          if (this.tabs.length > MAX_TABS) {
            var oldestIdx=-1, oldestId=Infinity;
            for (var j=0; j<this.tabs.length; j++) {
              if (j !== this.activeTab && this.tabs[j].path && this.tabs[j].id < oldestId) { oldestIdx=j; oldestId=this.tabs[j].id; }
            }
            if (oldestIdx >= 0) {
              var evicted = this.tabs[oldestIdx];
              __vaultrEditorClearTabState(evicted.id);
              this.tabs.splice(oldestIdx,1);
              if (oldestIdx < this.activeTab) this.activeTab--;
            }
          }
        }
        this._moveActiveTabToFront();
      },

      markTabCompiled(path) {
        for (var i = 0; i < this.tabs.length; i++) {
          if (this.tabs[i].path === path) { this.tabs[i].canCompile = false; return; }
        }
      },

      // Activate a draft tab: load draft, populate path input, focus it.
      // savedState comes from tabStateManager.restore() — pass null if unavailable.
      async _activateDraftTab(tab, savedState) {
        var draft = await __vaultrEditorLoadDraft(tab);
        if (!__vaultrEditorIsActiveTabId(tab.id)) return;
        var draftState = __vaultrEditorDraftEditorState(draft, savedState);
        await __vaultrContentPaneSetContent(draftState.content, tab.id, draftState, tab.draftId);
        setTimeout(function() {
          if (!__vaultrEditorIsActiveTabId(tab.id)) return;
          var pi = document.getElementById('content-pane-path-input');
          if (pi) { pi.value = draft.pathInput || ''; pi.classList.remove('invalid'); pi.placeholder = 'filename.md  ·  or  /folder/note.md'; }
          focusManager.focusPathInput();
        }, 0);
      },

      async contentPaneSwitchTab(i) {
        focusManager.blurActive();
        if (i === this.activeTab) return;

        var prevTab = this.tabs[this.activeTab];
        var nextTab = this.tabs[i];
        if (!nextTab) return;
        
        // Save current tab's state
        if (prevTab) {
          await __vaultrEditorSaveTabForLeave(prevTab);
          if (!prevTab.path) __vaultrEditorPathAc && __vaultrEditorPathAc.close();
        }
        
        // Switch active tab
        this.activeTab = i;
        this._moveActiveTabToFront();
        __vaultrEditorResetCompileBtn();

        // Load next tab's content with saved state
        var savedState = __vaultrEditorRestoreTabState(nextTab.id);
        
        if (!nextTab.path) {
          await this._activateDraftTab(nextTab, savedState);
        } else {
          await __vaultrContentPaneLoadNote(nextTab.path, nextTab.id, savedState);
        }
      },

      async contentPaneCloseTab(i) {
        if (i < 0 || i >= this.tabs.length) return;
        var wasActive = (i === this.activeTab);
        var closingTab = this.tabs[i];
        var closingCreate = !closingTab.path;
        if (wasActive && closingTab.path && __vaultrEditor.dirty && __vaultrEditor.currentPath === closingTab.path) {
          clearTimeout(__vaultrEditor.saveTimer); __vaultrEditor.saveTimer = null;
          await __vaultrEditorDoSave();
          var liveIdx = this.tabs.indexOf(closingTab);
          if (liveIdx < 0) return;
          i = liveIdx;
          wasActive = (i === this.activeTab);
        }
        
        // Clear saved state for this tab
        __vaultrEditorClearTabState(closingTab.id);
        if (closingCreate) void __vaultrEditorDeleteDraft(closingTab);
        
        this.tabs.splice(i, 1);
        if (closingCreate && wasActive) __vaultrEditorPathAc && __vaultrEditorPathAc.close();
        
        if (this.tabs.length === 0) {
          this.contentPaneOpen = false; this.activeTab = -1;
          clearTimeout(__vaultrEditor.saveTimer); __vaultrEditor.saveTimer = null;
          __vaultrEditor.currentPath = ''; __vaultrEditor.currentDraftId = ''; __vaultrEditor.currentMd = ''; __vaultrEditor.dirty = false;
          __vaultrEditorSaveStatus('');
          if (__vaultrEditor.view) {
            __vaultrEditor.loading = true;
            // Clearing to empty removes the frontmatter range too — needs
            // the same escape hatch as syncContent() above, or a note with
            // frontmatter left behind a leftover (blocked) suppressed range.
            __vaultrEditor.view.dispatch({
              changes: {from: 0, to: __vaultrEditor.view.state.doc.length, insert: ''},
              annotations: __vaultrEditor.allowFrontmatterEdit ? __vaultrEditor.allowFrontmatterEdit.of(true) : undefined,
            });
            setTimeout(function(){ __vaultrEditor.loading = false; }, 50);
          }
          return;
        }
        
        if (i < this.activeTab) { 
          this.activeTab -= 1; 
        } else if (wasActive) {
          this.activeTab = Math.min(i, this.tabs.length-1);
          var t = this.tabs[this.activeTab];
          if (!t) return;
          
          var savedState = __vaultrEditorRestoreTabState(t.id);
          
          if (t.path) {
            void __vaultrContentPaneLoadNote(t.path, t.id, savedState);
          } else {
            void this._activateDraftTab(t, savedState);
          }
        }
      },

      discardNewNote() {
        var i = this.activeTab;
        if (i < 0 || !this.tabs[i] || this.tabs[i].path) return;
        void this.contentPaneCloseTab(i);
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
        clearTimeout(__vaultrEditor.saveTimer); __vaultrEditor.saveTimer = null; __vaultrEditor.dirty = false;
        var cur = this.activeTab;
        var deletedTab = this.tabs[cur];
        if (deletedTab) __vaultrEditorClearTabState(deletedTab.id);
        this.tabs.splice(cur, 1);
        if (this.tabs.length === 0) {
          this.contentPaneOpen = false; this.activeTab = -1;
          __vaultrEditor.currentPath = ''; __vaultrEditor.currentDraftId = ''; __vaultrEditor.currentMd = ''; __vaultrEditorSaveStatus('');
          if (__vaultrEditor.view) {
            __vaultrEditor.loading = true;
            // Clearing to empty removes the frontmatter range too — needs
            // the same escape hatch as syncContent() above, or a note with
            // frontmatter left behind a leftover (blocked) suppressed range.
            __vaultrEditor.view.dispatch({
              changes: {from: 0, to: __vaultrEditor.view.state.doc.length, insert: ''},
              annotations: __vaultrEditor.allowFrontmatterEdit ? __vaultrEditor.allowFrontmatterEdit.of(true) : undefined,
            });
            setTimeout(function(){ __vaultrEditor.loading=false; },50);
          }
        } else {
          this.activeTab = cur > 0 ? cur-1 : 0;
          var nextTab = this.tabs[this.activeTab];
          if (nextTab && nextTab.path) void __vaultrContentPaneLoadNote(nextTab.path, nextTab.id, __vaultrEditorRestoreTabState(nextTab.id));
          else if (nextTab) {
            void this._activateDraftTab(nextTab, __vaultrEditorRestoreTabState(nextTab.id));
          }
        }
        if (window.__vaultrAfterVaultMutation) await window.__vaultrAfterVaultMutation();
      },
    };
  }

  // ── Compile raw note from pane ─────────────────────────────────────────────
  function __vaultrEditorCompileLabel(btn, text) {
    var label = btn.querySelector('.content-pane-compile-label');
    if (label) label.textContent = text;
  }
  function __vaultrEditorCloseMoreMenu() {
    document.dispatchEvent(new CustomEvent('content-pane:close-more'));
  }
  function __vaultrEditorResetCompileBtn() {
    var btn = document.querySelector('.content-pane-compile-btn');
    if (!btn) return;
    btn.classList.remove('is-compiling', 'success');
    btn.disabled = false;
    btn.title = 'Compile to knowledge note';
    __vaultrEditorCompileLabel(btn, 'Compile');
  }

  async function compileContentPaneNote(event) {
    var btn = event && event.currentTarget;
    if (!btn || btn.disabled) return;
    var pane = window.__vaultrContentPane;
    var tab = pane ? pane.tabs[pane.activeTab] : null;
    if (!tab || !tab.path || !tab.canCompile) return;
    var rawPath = tab.path;

    if (__vaultrEditor.dirty && __vaultrEditor.currentPath === rawPath) {
      clearTimeout(__vaultrEditor.saveTimer);
      __vaultrEditor.saveTimer = null;
      await __vaultrEditorDoSave();
    }

    var originalTitle = btn.title;
    btn.disabled = true;
    btn.classList.add('is-compiling');
    btn.title = 'Compiling…';
    __vaultrEditorCompileLabel(btn, 'Compiling…');

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
        __vaultrEditorCompileLabel(btn, 'Compile');
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
            if (pane && pane.markTabCompiled) pane.markTabCompiled(rawPath);
            btn.classList.add('success');
            btn.title = 'Compiled';
            __vaultrEditorCompileLabel(btn, 'Compiled');
            setTimeout(function() {
              btn.disabled = false;
              btn.classList.remove('success');
              __vaultrEditorCompileLabel(btn, 'Compile');
              __vaultrEditorCloseMoreMenu();
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
      __vaultrEditorCompileLabel(btn, 'Compile');
    } finally {
      btn.classList.remove('is-compiling');
    }
  }

  // Global undo/redo called by Electron main process via executeJavaScript.
  // Calls CodeMirror undo directly, bypassing browser native undo.
  window.__vaultrUndo = function() {
    var s = __vaultrEditor;
    if (s.view && s.cmUndo) s.cmUndo(s.view);
  };
  window.__vaultrRedo = function() {
    var s = __vaultrEditor;
    if (s.view && s.cmRedo) s.cmRedo(s.view);
  };

  window.__vaultrEditorShellHref = function() {
    try {
      var seg = location.pathname.replace(/^\/+/,'').split('/')[0];
      return '/'+(seg||'home');
    } catch(_) { return '/home'; }
  };

  window.__vaultrHotkeys.register('content-pane', 'e', function() {
    if (window.__vaultrContentPane) window.__vaultrContentPane.contentPaneOpen = !window.__vaultrContentPane.contentPaneOpen;
  });

  window.__vaultrHotkeys.register('new-note', 'n', function() {
    if (window.__vaultrContentPane) void window.__vaultrContentPane.openNewInContentPane('', '');
  });

  window.__vaultrHotkeys.register('content-pane-maximize', '\\', function() {
    var _pane = window.__vaultrContentPane;
    if (_pane && _pane.contentPaneOpen) _pane.toggleMaximizeEditor();
  });

  window.__vaultrHotkeys.registerRaw('content-pane-scroll', function(e, mod) {
    if (e.key !== 'ArrowUp' && e.key !== 'ArrowDown') return;
    if (mod || e.altKey) return;
    var ae = document.activeElement;
    var tag = ae && ae.tagName;
    if (tag === 'INPUT' || tag === 'TEXTAREA') return;
    var _pane = window.__vaultrContentPane;
    if (!_pane || !_pane.contentPaneOpen) return;
    var _editArea = document.getElementById('content-pane-edit-area');
    if (_editArea && _editArea.contains(ae)) return;
    e.preventDefault();
    var _scroller = document.querySelector('#content-pane-edit-area .cm-scroller');
    if (_scroller) _scroller.scrollBy(0, e.key === 'ArrowDown' ? 80 : -80);
    return true;
  });

  window.__vaultrHotkeys.registerRaw('content-pane-close-tab', function(e, mod) {
    if (!mod || e.shiftKey || e.altKey || e.key.toLowerCase() !== 'w') return;
    var pane = window.__vaultrContentPane;
    if (!pane || !pane.contentPaneOpen || pane.activeTab < 0) return;
    e.preventDefault();
    var at = pane.tabs[pane.activeTab];
    if (at && at.path) pane.contentPaneCloseTab(pane.activeTab);
    return true;
  });

  // Shared by the Mod-F hotkey below and the editor's "more" menu (Find
  // item, content_pane.html) so there's one place that knows how to open it.
  function __vaultrEditorOpenFind() {
    var s = __vaultrEditor;
    if (!s.cmOpenSearchPanel || !s.view) return;
    s.cmOpenSearchPanel(s.view);
  }

  window.__vaultrHotkeys.registerRaw('content-pane-find', function(e, mod) {
    if (!mod || e.shiftKey || e.altKey || e.key.toLowerCase() !== 'f') return;
    var _fContentPane = window.__vaultrContentPane;
    if (!_fContentPane || !_fContentPane.contentPaneOpen) return;
    if (!__vaultrEditor.cmOpenSearchPanel || !__vaultrEditor.view) return;
    e.preventDefault();
    __vaultrEditorOpenFind();
    return true;
  });

  // The tabs dropdown's own open/highlight state lives in its local x-data
  // (content_pane.html's #content-pane-tabs-wrap) rather than on window.__vaultrContentPane —
  // same "small local x-data for a popover" pattern as .content-pane-more-wrap.
  // Alpine.$data bridges into it for these plain-JS hotkey handlers.
  function __vaultrContentPaneTabsMenuScope() {
    var el = document.getElementById('content-pane-tabs-wrap');
    return (el && window.Alpine) ? window.Alpine.$data(el) : null;
  }

  // Mod-T opens/closes the tabs dropdown (mirrors the "switch tabs" meaning
  // Cmd/Ctrl+T carries in most apps).
  window.__vaultrHotkeys.registerRaw('content-pane-tabs-menu-toggle', function(e, mod) {
    if (!mod || e.shiftKey || e.altKey || e.key.toLowerCase() !== 't') return;
    var _pane = window.__vaultrContentPane;
    if (!_pane || !_pane.contentPaneOpen) return;
    var scope = __vaultrContentPaneTabsMenuScope();
    if (!scope) return;
    e.preventDefault();
    scope.toggleTabsMenu();
    if (!scope.tabsMenuOpen) document.dispatchEvent(new CustomEvent('content-pane:close-tabs'));
    return true;
  });

  // While the tabs dropdown is open: ↑/↓ move the highlighted row, Enter
  // switches to it. Registered after content-pane-scroll (below) so it's checked
  // first — see keysJS's reverse-registration-order dispatch — and takes
  // over plain arrow keys instead of letting them scroll the editor.
  window.__vaultrHotkeys.registerRaw('content-pane-tabs-menu-nav', function(e, mod) {
    if (mod || e.shiftKey || e.altKey) return;
    if (e.key !== 'ArrowUp' && e.key !== 'ArrowDown' && e.key !== 'Enter') return;
    var menu = document.querySelector('.content-pane-tabs-menu');
    if (!menu || menu.style.display === 'none') return;
    var scope = __vaultrContentPaneTabsMenuScope();
    var _pane = window.__vaultrContentPane;
    if (!scope || !scope.tabsMenuOpen || !_pane) return;
    var n = _pane.tabs.length;
    if (!n) return;
    e.preventDefault();
    if (e.key === 'ArrowDown') {
      scope.tabsMenuActiveIndex = (scope.tabsMenuActiveIndex + 1 + n) % n;
    } else if (e.key === 'ArrowUp') {
      scope.tabsMenuActiveIndex = (scope.tabsMenuActiveIndex - 1 + n) % n;
    } else {
      var idx = scope.tabsMenuActiveIndex;
      if (idx >= 0 && idx < n) void _pane.contentPaneSwitchTab(idx);
      scope.tabsMenuOpen = false;
    }
    return true;
  });

  // Intercept mouse back/forward buttons (button 3/4) to switch pane tabs.
  window.addEventListener('mousedown', function(e) {
    var _pane = window.__vaultrContentPane;
    if (!_pane || !_pane.contentPaneOpen || _pane.tabs.length <= 1) return;
    if (e.button === 3) {
      e.preventDefault();
      if (_pane.activeTab > 0) void _pane.contentPaneSwitchTab(_pane.activeTab - 1);
    } else if (e.button === 4) {
      e.preventDefault();
      if (_pane.activeTab < _pane.tabs.length - 1) void _pane.contentPaneSwitchTab(_pane.activeTab + 1);
    }
  }, true);

  // Keeps .content-pane-tool-bar's height pixel-matched to whichever
  // #home-list-pane section is currently showing, rather than guessing a
  // CSS constant that has to independently agree with every section's own
  // header (Pinned's .home-list-head grows a hair past --btn-h-xs because
  // its .seg control's own border+padding don't sum to exactly 28px — a
  // fixed height on both sides can't account for that without measuring).
  // Graph has no .home-list-head at all, so the property is cleared and
  // content_pane.css's calc() fallback takes over.
  function __vaultrSyncContentPaneHeaderHeight() {
    var head = document.querySelector('#home-list-pane .home-list-head');
    var root = document.documentElement;
    if (head) {
      var h = head.getBoundingClientRect().height;
      if (h > 0) { root.style.setProperty('--content-header-h', h + 'px'); return; }
    }
    root.style.removeProperty('--content-header-h');
  }
  document.body.addEventListener('htmx:afterSwap', function(e) {
    var target = e.detail && e.detail.target;
    if (target && target.id === 'home-list-pane') __vaultrSyncContentPaneHeaderHeight();
  });
  __vaultrSyncContentPaneHeaderHeight();

