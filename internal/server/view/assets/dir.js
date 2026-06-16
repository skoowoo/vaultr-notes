
  window.handleSearchResultSelection = function(el) {
    if (!el || !el.dataset) return false;
    var focusURL = el.dataset.focusUrl || '';
    if (!focusURL) return false;
    try {
      var u = new URL(focusURL, window.location.origin);
      var path = u.searchParams.get('path');
      if (!path) return false;
      var nameEl = el.querySelector('.sr-name');
      var title = nameEl ? nameEl.textContent.trim() : (path.split('/').pop().replace(/\.md$/, '') || 'Note');
      if (window.__vaultrDrawer) {
        void window.__vaultrDrawer.openNoteInDrawer(path, title,
          el.dataset.noteIsKnowledge === 'true', false, el.dataset.noteIsIndex === 'true',
          el.dataset.noteCanCompile === 'true');
        return true;
      }
    } catch(e) { return false; }
    return false;
  };

  // ── Dir compose entry cards ───────────────────────────────────────────────
  (function() {
    var newCard    = document.getElementById('dir-new-card');
    var importCard = document.getElementById('dir-import-card');
    var fileInput  = document.getElementById('dir-compose-file-input');
    if (!newCard || !importCard || !fileInput) return;
    newCard.onclick = function() {
      if (window.__vaultrDrawer) void window.__vaultrDrawer.openNewInDrawer('', '');
    };
    importCard.onclick = function() { fileInput.click(); };
    fileInput.onchange = function(e) {
      var file = e.target.files && e.target.files[0];
      if (!file) return;
      var reader = new FileReader();
      reader.onload = function(ev) {
        if (window.__vaultrDrawer)
          void window.__vaultrDrawer.openNewInDrawer(ev.target.result, file.name);
      };
      reader.readAsText(file, 'UTF-8');
      fileInput.value = '';
    };
  })();

  async function dirRefresh() {
    var p = new URLSearchParams(window.location.search).get('path') || '/';
    try {
      var resp = await fetch('/dir/refresh?path=' + encodeURIComponent(p), {
        headers: { 'HX-Request': 'true' }
      });
      if (!resp.ok) return;
      var text = await resp.text();
      var doc = (new DOMParser()).parseFromString('<html><body>' + text + '</body></html>', 'text/html');

      var newCountEl = doc.getElementById('dir-note-count-head');
      var oldCountEl = document.getElementById('dir-note-count-head');
      var newGrid    = doc.getElementById('dir-notes-grid');
      var oldGrid    = document.getElementById('dir-notes-grid');
      if (!newCountEl || !oldCountEl || !newGrid || !oldGrid) return;

      var newCount = newCountEl.textContent.trim();
      var oldCount = oldCountEl.textContent.trim();

      // Always sync the count badge (text-only, no layout impact)
      if (newCount !== oldCount) oldCountEl.textContent = newCount;

      // Skip grid replacement entirely when count is unchanged
      if (newCount === oldCount) return;

      // Fade out → innerHTML swap → fade in (avoids outerHTML reflow)
      oldGrid.style.transition = 'opacity 150ms ease';
      oldGrid.style.opacity = '0';
      setTimeout(function() {
        oldGrid.innerHTML = newGrid.innerHTML;
        htmx.process(oldGrid);
        requestAnimationFrame(function() { oldGrid.style.opacity = '1'; });
      }, 160);
    } catch(e) {}
  }

  function dirCtrl() {
    return Object.assign(drawerCtrl(), {
      init() {
        this.initDrawer();
        window.__vaultrShellSafeForBackgroundReload = function() { return true; };
        window.__vaultrBackgroundRefresh = dirRefresh;
      },
    });
  }
