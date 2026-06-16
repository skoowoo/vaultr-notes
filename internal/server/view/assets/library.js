
  var libData = null;

  // doLibraryRefresh resets all selection state back to the initial view.
  // /library/refresh OOB already covers knowledge-col-body and raw-col-body,
  // so no subsequent /library/unfocus call is needed.
  function doLibraryRefresh() {
    history.replaceState(null, '', '/library');
    if (!libData) return;
    _clearKSelection();
    libData.selectedTag = '';
    libData.selectedIndex = '';
    libData.selectedFocus = '';
    clearHighlights(); clearConnectionLines();
    htmx.ajax('GET', '/library/refresh', { target: document.body, swap: 'none' });
  }

  function libCtrl() {
    return Object.assign(drawerCtrl(), {
      selectedMode: 'index',
      selectedTag: '',
      selectedIndex: '',
      selectedFocus: '',
      get dark() { return Alpine.store('theme').dark; },
      refresh() { doLibraryRefresh(); },

      init() {
        this.initDrawer();
        libData = this;

        // Restore selection state from URL on initial page load.
        var params = new URLSearchParams(window.location.search);
        var initTag = params.get('tag');
        var initIndex = params.get('index');
        if (initTag) {
          this.selectedTag = initTag;
          this.selectedMode = 'tags';
          htmx.ajax('GET', '/library/tag?tag=' + encodeURIComponent(initTag),
            {target: document.getElementById('knowledge-col-body'), swap: 'innerHTML'});
        } else if (initIndex) {
          this.selectedIndex = initIndex;
          htmx.ajax('GET', '/library/index/select?path=' + encodeURIComponent(initIndex),
            {target: document.getElementById('knowledge-col-body'), swap: 'innerHTML'});
        }

        window.__vaultrShellSafeForBackgroundReload = function() {
          return !libData || (!libData.selectedTag && !libData.selectedIndex && !libData.selectedFocus);
        };
        window.__vaultrBackgroundRefresh = doLibraryRefresh;
      },
    });
  }

  function scrollColToTop(bodyId) {
    var el = document.getElementById(bodyId);
    if (el) el.scrollTo({ top: 0, behavior: 'smooth' });
  }

  function bindColScrollHandlers() {
    document.querySelectorAll('.col-body').forEach(function(el) {
      if (el._vaultrScrollHandler) {
        el.removeEventListener('scroll', el._vaultrScrollHandler);
      }
      var btnId = el.id ? el.id.replace('col-body', 'back-top') : null;
      el._vaultrScrollHandler = function() {
        if (btnId) {
          var btn = document.getElementById(btnId);
          if (btn) btn.classList.toggle('visible', el.scrollTop > 80);
        }
      };
      el.addEventListener('scroll', el._vaultrScrollHandler, {passive: true});
    });
  }

  document.addEventListener('DOMContentLoaded', function() {
    bindColScrollHandlers();
  });

  // ── K card in-place selection ────────────────────────────────────────────────

  var _selectedKPath = '';

  function _clearKSelection() {
    if (!_selectedKPath) return;
    var prev = document.querySelector('#knowledge-col-body .note-card.note-card-pinned[data-path]');
    if (prev) prev.classList.remove('note-card-pinned');
    _selectedKPath = '';
  }

  function selectKCard(path, focusURL) {
    if (!libData) return;
    if (_selectedKPath === path) {
      // Toggle off: restore both cols to current filter context
      _clearKSelection();
      libData.selectedFocus = '';
      clearConnectionLines();
      clearHighlights();
      if (libData.selectedIndex) {
        htmx.ajax('GET', '/library/index/select?path=' + encodeURIComponent(libData.selectedIndex),
          {target: document.getElementById('knowledge-col-body'), swap: 'innerHTML'});
      } else if (libData.selectedTag) {
        htmx.ajax('GET', '/library/tag?tag=' + encodeURIComponent(libData.selectedTag),
          {target: document.getElementById('knowledge-col-body'), swap: 'innerHTML'});
      } else {
        htmx.ajax('GET', '/library/unfocus',
          {target: document.getElementById('knowledge-col-body'), swap: 'none'});
      }
      return;
    }
    // Deselect previous K card (if any)
    _clearKSelection();
    clearConnectionLines();
    clearHighlights();
    // Highlight new K card in-place
    _selectedKPath = path;
    libData.selectedFocus = focusURL;
    var card = document.querySelector('#knowledge-col-body .note-card[data-path="' + path.replace(/\\/g, '\\\\').replace(/"/g, '\\"') + '"]');
    if (card) card.classList.add('note-card-pinned');
    // Fetch only raw col (knowledge col stays intact)
    var rawOnlyURL = focusURL + (focusURL.indexOf('?') !== -1 ? '&' : '?') + 'raw_only=true';
    htmx.ajax('GET', rawOnlyURL, {target: document.getElementById('raw-col-body'), swap: 'none'});
  }

  // ── Note focus toggle (R cards only) ────────────────────────────────────────

  function focusNote(focusURL) {
    if (!libData) return;
    // Clear any in-place K selection (K col OOB swap will replace DOM anyway)
    _selectedKPath = '';
    if (libData.selectedFocus === focusURL) {
      libData.selectedFocus = '';
      if (libData.selectedIndex) {
        htmx.ajax('GET', '/library/index/select?path=' + encodeURIComponent(libData.selectedIndex),
          {target: document.getElementById('knowledge-col-body'), swap: 'innerHTML'});
      } else if (libData.selectedTag) {
        htmx.ajax('GET', '/library/tag?tag=' + encodeURIComponent(libData.selectedTag),
          {target: document.getElementById('knowledge-col-body'), swap: 'innerHTML'});
      } else {
        htmx.ajax('GET', '/library/unfocus',
          {target: document.getElementById('knowledge-col-body'), swap: 'none'});
      }
    } else {
      libData.selectedFocus = focusURL;
      htmx.ajax('GET', focusURL,
        {target: document.getElementById('knowledge-col-body'), swap: 'none'});
    }
  }

  function refreshFocus(focusURL) {
    if (!libData || !focusURL) return;
    libData.selectedFocus = focusURL;
    htmx.ajax('GET', focusURL,
      {target: document.getElementById('knowledge-col-body'), swap: 'none'});
  }

  async function compileRaw(event) {
    var btn = event && event.currentTarget;
    if (!btn || btn.disabled) return;
    var rawPath = btn.dataset.path;
    var focusURL = btn.dataset.focusUrl;
    if (!rawPath || !focusURL) return;

    var label = btn.querySelector('span');
    var original = label ? label.textContent : '';
    btn.disabled = true;
    btn.classList.add('is-compiling');
    if (label) label.textContent = 'Compiling';

    try {
      var resp = await fetch('/api/compile/trigger', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({path: rawPath})
      });
      if (!resp.ok && resp.status !== 409) {
        var msg = 'Compile failed';
        try {
          var data = await resp.json();
          if (data && data.error) msg = data.error;
        } catch (_) {}
        throw new Error(msg);
      }
      if (resp.status === 409) {
        refreshFocus(focusURL);
        return;
      }

      // Poll run status until the agent finishes.
      var pollURL = '/api/runs/by-ref?path=' + encodeURIComponent(rawPath);
      var deadline = Date.now() + 10 * 60 * 1000;
      await new Promise(function(r) { setTimeout(r, 600); });
      while (Date.now() < deadline) {
        var pr = await fetch(pollURL);
        if (pr.ok) {
          var st = await pr.json();
          if (st.status === 'succeeded') {
            refreshFocus(focusURL);
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
      if (label) label.textContent = 'Retry';
      btn.title = err && err.message ? err.message : 'Compile failed';
    } finally {
      btn.classList.remove('is-compiling');
    }
  }

  window.handleSearchResultSelection = function(el) {
    var focusURL = el && el.dataset ? el.dataset.focusUrl : '';
    if (!focusURL) return false;
    focusNote(focusURL);
    return true;
  };

  // ── Index selection toggle ───────────────────────────────────────────────────

  function switchMode(mode) {
    if (!libData || libData.selectedMode === mode) return;
    var hadFilter = libData.selectedTag || libData.selectedIndex || libData.selectedFocus;
    _clearKSelection();
    libData.selectedMode = mode;
    libData.selectedTag = '';
    libData.selectedIndex = '';
    libData.selectedFocus = '';
    clearHighlights(); clearConnectionLines();
    history.pushState(null, '', '/library');
    if (hadFilter) {
      htmx.ajax('GET', '/library/unfocus', {target: document.body, swap: 'none'});
    }
  }

  function selectIndex(path) {
    if (!libData) return;
    _clearKSelection();
    clearHighlights(); clearConnectionLines();
    if (libData.selectedIndex === path) {
      libData.selectedIndex = '';
      libData.selectedFocus = '';
      history.pushState(null, '', '/library');
      htmx.ajax('GET', '/library/unfocus',
        {target: document.getElementById('knowledge-col-body'), swap: 'none'});
    } else {
      libData.selectedIndex = path;
      libData.selectedTag = '';
      libData.selectedFocus = '';
      history.pushState(null, '', '/library?index=' + encodeURIComponent(path));
      htmx.ajax('GET', '/library/index/select?path=' + encodeURIComponent(path),
        {target: document.getElementById('knowledge-col-body'), swap: 'innerHTML'});
    }
  }

  // ── Tag selection toggle ─────────────────────────────────────────────────────

  function selectTag(name) {
    if (!libData) return;
    _clearKSelection();
    clearHighlights(); clearConnectionLines();
    if (libData.selectedTag === name) {
      libData.selectedTag = '';
      libData.selectedFocus = '';
      history.pushState(null, '', '/library');
      htmx.ajax('GET', '/library/unfocus',
        {target: document.getElementById('knowledge-col-body'), swap: 'none'});
    } else {
      libData.selectedTag = name;
      libData.selectedIndex = '';
      libData.selectedFocus = '';
      history.pushState(null, '', '/library?tag=' + encodeURIComponent(name));
      htmx.ajax('GET', '/library/tag?tag=' + encodeURIComponent(name),
        {target: document.getElementById('knowledge-col-body'), swap: 'innerHTML'});
    }
  }

  // ── Column count badges ──────────────────────────────────────────────────────

  function updateColCounts() {
    var kEl = document.getElementById('knowledge-col-count');
    var rEl = document.getElementById('raw-col-count');
    if (!kEl || !rEl) return;
    var isFiltered = libData && (libData.selectedTag || libData.selectedIndex || libData.selectedFocus);
    if (isFiltered) {
      kEl.textContent = document.querySelectorAll('#knowledge-col-body .note-card').length;
      rEl.textContent = document.querySelectorAll('#raw-col-body .note-card').length;
    } else {
      kEl.textContent = kEl.dataset.total || '';
      rEl.textContent = rEl.dataset.total || '';
    }
  }

  // ── HTMX hooks ───────────────────────────────────────────────────────────────
  document.addEventListener('htmx:afterSwap', function(e) {
    var tid = e.detail.target.id;
    if (tid !== 'knowledge-col-body' && tid !== 'raw-col-body') return;
    bindColScrollHandlers();
    clearTimeout(window._afterSwapTimer);
    window._afterSwapTimer = setTimeout(function() {
      updateColCounts();
      if (document.querySelector('#knowledge-col-body .note-card-pinned') ||
          document.querySelector('#raw-col-body .note-card-pinned')) {
        activateConnectionLines();
      } else {
        clearConnectionLines();
        clearHighlights();
      }
    }, 50);
  });

  // Sync view when browser back/forward navigation changes the URL.
  window.addEventListener('popstate', function() {
    if (!libData) return;
    _clearKSelection();
    clearHighlights(); clearConnectionLines();
    var params = new URLSearchParams(window.location.search);
    var tag = params.get('tag');
    var index = params.get('index');
    libData.selectedTag = '';
    libData.selectedIndex = '';
    libData.selectedFocus = '';
    if (tag) {
      libData.selectedTag = tag;
      libData.selectedMode = 'tags';
      htmx.ajax('GET', '/library/tag?tag=' + encodeURIComponent(tag),
        {target: document.getElementById('knowledge-col-body'), swap: 'innerHTML'});
    } else if (index) {
      libData.selectedIndex = index;
      htmx.ajax('GET', '/library/index/select?path=' + encodeURIComponent(index),
        {target: document.getElementById('knowledge-col-body'), swap: 'innerHTML'});
    } else {
      htmx.ajax('GET', '/library/unfocus',
        {target: document.getElementById('knowledge-col-body'), swap: 'none'});
    }
  });

  // ── SVG connection lines (bidirectional focus mode) ─────────────────────────

  var _connScrollCleanup = null;

  function _drawLine(svg, cr, srcX, srcY, tX, tY, animate) {
    var mid = (srcX + tX) / 2;
    var path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    path.setAttribute('class', 'lib-conn-line');
    path.setAttribute('d',
      'M' + srcX + ',' + srcY +
      ' H' + mid +
      ' V' + tY +
      ' H' + tX
    );
    svg.appendChild(path);
    if (animate) {
      var len = path.getTotalLength();
      path.style.strokeDasharray = len;
      path.style.strokeDashoffset = len;
      requestAnimationFrame(function() {
        path.style.transition = 'stroke-dashoffset 0.45s ease';
        path.style.strokeDashoffset = 0;
      });
    }
  }

  function drawConnectionLinesFromKnowledge(animate) {
    var svg = document.getElementById('lib-connections');
    if (!svg) return;
    svg.innerHTML = '';
    var container = document.querySelector('.lib-content');
    if (!container) return;
    var cr = container.getBoundingClientRect();
    var srcCard = document.querySelector('#knowledge-col-body .note-card-pinned');
    if (!srcCard) return;
    var sr = srcCard.getBoundingClientRect();
    var srcX = sr.right - cr.left;
    var srcY = (sr.top + sr.bottom) / 2 - cr.top;
    document.querySelectorAll('#raw-col-body .note-card').forEach(function(rawCard) {
      var rr = rawCard.getBoundingClientRect();
      if (rr.bottom < cr.top || rr.top > cr.bottom) return;
      _drawLine(svg, cr, srcX, srcY, rr.left - cr.left, (rr.top + rr.bottom) / 2 - cr.top, animate);
    });
  }

  function drawConnectionLinesFromRaw(animate) {
    var svg = document.getElementById('lib-connections');
    if (!svg) return;
    svg.innerHTML = '';
    var container = document.querySelector('.lib-content');
    if (!container) return;
    var cr = container.getBoundingClientRect();
    var srcCard = document.querySelector('#raw-col-body .note-card-pinned');
    if (!srcCard) return;
    var sr = srcCard.getBoundingClientRect();
    var srcX = sr.left - cr.left;
    var srcY = (sr.top + sr.bottom) / 2 - cr.top;
    document.querySelectorAll('#knowledge-col-body .note-card').forEach(function(kCard) {
      var rr = kCard.getBoundingClientRect();
      if (rr.bottom < cr.top || rr.top > cr.bottom) return;
      _drawLine(svg, cr, srcX, srcY, rr.right - cr.left, (rr.top + rr.bottom) / 2 - cr.top, animate);
    });
  }

  function _getActiveDrawFn() {
    if (document.querySelector('#knowledge-col-body .note-card-pinned')) return drawConnectionLinesFromKnowledge;
    if (document.querySelector('#raw-col-body .note-card-pinned')) return drawConnectionLinesFromRaw;
    return null;
  }

  function clearConnectionLines() {
    var svg = document.getElementById('lib-connections');
    if (svg) svg.innerHTML = '';
    if (_connScrollCleanup) { _connScrollCleanup(); _connScrollCleanup = null; }
  }

  function activateConnectionLines() {
    clearConnectionLines();
    var drawFn = _getActiveDrawFn();
    if (!drawFn) return;
    requestAnimationFrame(function() {
      drawFn(true);
      var handler = function() { drawFn(false); };
      var kBody = document.getElementById('knowledge-col-body');
      var rBody = document.getElementById('raw-col-body');
      if (kBody) kBody.addEventListener('scroll', handler, {passive: true});
      if (rBody) rBody.addEventListener('scroll', handler, {passive: true});
      _connScrollCleanup = function() {
        if (kBody) kBody.removeEventListener('scroll', handler);
        if (rBody) rBody.removeEventListener('scroll', handler);
      };
    });
  }

  // ── CSS connection highlighting ───────────────────────────────────────────────

  function clearHighlights() {
    document.querySelectorAll('#raw-col-body .note-card, #knowledge-col-body .note-card').forEach(function(card) {
      card.classList.remove('is-linked', 'is-dim');
    });
  }
