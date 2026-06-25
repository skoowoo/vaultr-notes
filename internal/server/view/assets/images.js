
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

  // ── Lightbox ─────────────────────────────────────────────────────────────────

  var _lbOpen = null; // set by imgCtrl.init; receives the lightbox data object

  window._imgSelectMode = false;
  window._imgUpdateSelected = null;

  window.openImageLightbox = function(el) {
    if (window._imgSelectMode) {
      el.classList.toggle('is-selected');
      if (window._imgUpdateSelected) window._imgUpdateSelected();
      return;
    }
    var d = el.dataset;
    var notes = [];
    try {
      var parsed = JSON.parse(d.imgNotes || '[]');
      if (Array.isArray(parsed)) notes = parsed;
    } catch(_) {}
    if (_lbOpen) _lbOpen({
      src: d.imgSrc, name: d.imgName, dir: d.imgDir,
      size: d.imgSize, time: d.imgTime, ext: d.imgExt,
      notes: notes,
    });
  };

  function extToType(ext) {
    var m = { '.jpg':'JPEG Image', '.jpeg':'JPEG Image', '.png':'PNG Image',
               '.gif':'GIF Image', '.webp':'WebP Image', '.avif':'AVIF Image', '.svg':'SVG Vector' };
    return m[(ext||'').toLowerCase()] || ((ext||'').toUpperCase().replace('.','') + ' Image');
  }

  function imgCtrl() {
    return Object.assign(drawerCtrl(), {
      lightbox: null,
      selectMode: false,
      selectedCount: 0,
      init() {
        this.initDrawer();
        _lbOpen = (data) => { this.lightbox = data; };
        window._imgUpdateSelected = () => {
          this.selectedCount = document.querySelectorAll('.img-card.is-selected').length;
        };
      },
      enterSelectMode() {
        this.lightbox = null;
        this.selectMode = true;
        window._imgSelectMode = true;
      },
      exitSelectMode() {
        this.selectMode = false;
        window._imgSelectMode = false;
        document.querySelectorAll('.img-card.is-selected').forEach(c => c.classList.remove('is-selected'));
        this.selectedCount = 0;
      },
      async deleteSelected() {
        var cards = Array.from(document.querySelectorAll('.img-card.is-selected'));
        if (!cards.length) return;
        var n = cards.length;
        var ok = await window.showConfirm({
          titleHTML: '<span class="confirm-title-icon"><svg width="13" height="13" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 4h10M5 4V3a1 1 0 011-1h2a1 1 0 011 1v1M12 4l-1 8H3L2 4"/><path d="M6 7v3M8 7v3"/></svg></span>Delete ' + n + ' image' + (n > 1 ? 's' : ''),
          message: 'Permanently delete ' + n + ' image' + (n > 1 ? 's' : '') + '? This cannot be undone.',
          confirmLabel: 'Delete',
          danger: true,
        });
        if (!ok) return;
        var results = await Promise.allSettled(cards.map(card =>
          fetch('/api/images/delete', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ dir: card.dataset.imgDir, name: card.dataset.imgName }),
          }).then(r => ({ ok: r.ok, card }))
        ));
        var removed = 0;
        results.forEach(r => {
          if (r.status === 'fulfilled' && r.value.ok) { r.value.card.remove(); removed++; }
        });
        this.exitSelectMode();
        var stat = document.querySelector('.lib-status-value');
        if (stat && removed > 0) {
          var cur = parseInt(stat.textContent, 10);
          if (!isNaN(cur)) stat.textContent = String(Math.max(0, cur - removed));
        }
        var grid = document.getElementById('img-grid');
        if (grid && !grid.querySelector('.img-card') && !grid.querySelector('.img-sentinel')) {
          var empty = document.createElement('div');
          empty.className = 'img-empty';
          empty.textContent = 'No images found';
          grid.appendChild(empty);
        }
      },
      closeLightbox() { this.lightbox = null; },
      extToType,
      async deleteLightboxImage() {
        var lb = this.lightbox;
        if (!lb || !lb.name || lb.dir == null || lb.dir === undefined) return;
        var linkHint = (lb.notes && lb.notes.length)
          ? (' It is still referenced from ' + lb.notes.length + ' note(s); those embeds will break.')
          : '';
        var ok = await window.showConfirm({
          titleHTML: '<span class="confirm-title-icon"><svg width="13" height="13" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 4h10M5 4V3a1 1 0 011-1h2a1 1 0 011 1v1M12 4l-1 8H3L2 4"/><path d="M6 7v3M8 7v3"/></svg></span>Delete image',
          message: 'Permanently delete "' + lb.name + '" from the vault?' + linkHint,
          confirmLabel: 'Delete',
          danger: true,
        });
        if (!ok) return;
        try {
          var resp = await fetch('/api/images/delete', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ dir: lb.dir, name: lb.name }),
          });
          if (!resp.ok) {
            var msg = (await resp.text()).trim() || 'Delete failed.';
            window.showError(msg, 'Delete failed');
            return;
          }
          var dir = lb.dir, name = lb.name;
          this.closeLightbox();
          var grid = document.getElementById('img-grid');
          if (grid) {
            grid.querySelectorAll('.img-card').forEach(function(c) {
              if (c.dataset.imgDir === dir && c.dataset.imgName === name) c.remove();
            });
            var stat = document.querySelector('.lib-status-value');
            if (stat) {
              var n = parseInt(stat.textContent, 10);
              if (!isNaN(n) && n > 0) stat.textContent = String(n - 1);
            }
            if (!grid.querySelector('.img-card') && !grid.querySelector('.img-sentinel')) {
              var hadEmpty = grid.querySelector('.img-empty');
              if (!hadEmpty) {
                var empty = document.createElement('div');
                empty.className = 'img-empty';
                empty.textContent = 'No images found';
                grid.appendChild(empty);
              }
            }
          }
        } catch (e) {
          window.showError((e && e.message) ? e.message : 'Delete failed.', 'Delete failed');
        }
      },
      // Close lightbox first, then open the note in the drawer after the
      // leave animation (160 ms) finishes so the two panels don't collide.
      async openLinkedNote(noteName) {
        this.lightbox = null;
        try {
          var resp = await fetch('/api/notes/resolve', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: noteName + '.md' }),
          });
          if (!resp.ok) return;
          var data = await resp.json();
          var matches = data.matches;
          if (!Array.isArray(matches) || matches.length === 0) return;
          var n = matches[0];
          var notePath = n.dir === '/' ? '/' + n.name : n.dir + '/' + n.name;
          await new Promise(function(r) { setTimeout(r, 180); });
          if (window.__vaultrDrawer) {
            void window.__vaultrDrawer.openNoteInDrawer(notePath, noteName, false, !!n.pinned);
          }
        } catch(e) {}
      },
      refresh() { window.location.reload(); },
    });
  }
