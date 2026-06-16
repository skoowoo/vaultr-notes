
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

  function shortsCtrl() {
    return Object.assign(drawerCtrl(), {
      init() {
        this.initDrawer();
        window.__vaultrAfterVaultMutation = () => {
          window.location.href = '/shorts';
        };
      },
      refresh() {
        window.location.href = '/shorts';
      },
    });
  }

  // Intercept link clicks inside short entries.
  // - Wikilinks (/notes?…) → open in drawer editor
  // - External (https?://) → new tab
  document.addEventListener('click', function(e) {
    var a = e.target.closest ? e.target.closest('a') : null;
    if (!a || !a.closest('.shorts-entry-prose')) return;
    var href = a.getAttribute('href');
    if (!href) return;
    e.preventDefault();
    e.stopPropagation();
    if (href.indexOf('/notes?') === 0) {
      try {
        var u = new URL(href, window.location.origin);
        var name = u.searchParams.get('name') || '';
        var path = u.searchParams.get('path') || '';
        if (path && window.__vaultrDrawer) {
          var title = a.textContent.trim() || path.split('/').pop().replace(/\.md$/, '');
          void window.__vaultrDrawer.openNoteInDrawer(path, title, false, false);
        } else if (name) {
          void __vaultrDrawerOpenWikiLink(name.replace(/\.md$/, ''));
        }
      } catch(_) {}
    } else if (/^https?:\/\//.test(href)) {
      window.open(href, '_blank', 'noopener,noreferrer');
    }
  }, true);
