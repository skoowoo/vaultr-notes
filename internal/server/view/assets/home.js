
// ── Search result selection: open in editor drawer ───────────────────────
window.handleSearchResultSelection = function (el) {
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
  } catch (e) { return false; }
  return false;
};

// ── Home partial refresh via HTMX ────────────────────────────────────────
function homeRefresh() {
  htmx.ajax('GET', '/home/refresh', { target: document.body, swap: 'none' });
}

// ── Home Alpine controller ────────────────────────────────────────────────
function homeCtrl() {
  return Object.assign(drawerCtrl(), {
    init() {
      this.initDrawer();
      window._homeData = this;
      // Home is always safe: partial HTMX refresh is non-destructive.
      window.__vaultrShellSafeForBackgroundReload = function () { return true; };
      // Electron main calls this instead of wc.reload() when syncing sections.
      window.__vaultrBackgroundRefresh = homeRefresh;
    },
    refresh() { homeRefresh(); },
  });
}

// ── Compose entry cards ───────────────────────────────────────────────────
(function () {
  var newCard = document.getElementById('home-new-card');
  var importCard = document.getElementById('home-import-card');
  var fileInput = document.getElementById('home-compose-file-input');
  if (!newCard || !importCard || !fileInput) return;
  newCard.onclick = function () {
    if (window.__vaultrDrawer) void window.__vaultrDrawer.openNewInDrawer('', '');
  };
  importCard.onclick = function () { fileInput.click(); };
  fileInput.onchange = function (e) {
    var file = e.target.files && e.target.files[0];
    if (!file) return;
    var reader = new FileReader();
    reader.onload = function (ev) {
      if (window.__vaultrDrawer)
        void window.__vaultrDrawer.openNewInDrawer(ev.target.result, file.name);
    };
    reader.readAsText(file, 'UTF-8');
    fileInput.value = '';
  };
})();

// ── Hero background image ─────────────────────────────────────────────────
(function () {
  var wrapper = document.querySelector('.home-hero-wrapper');
  var pickBtn = document.getElementById('hero-bg-pick');
  var clearBtn = document.getElementById('hero-bg-clear');
  var fileInput = document.getElementById('hero-bg-input');
  if (!wrapper || !pickBtn || !clearBtn || !fileInput) return;

  var KEY = 'vaultr-hero-bg';
  var KEY_Y = 'vaultr-hero-bg-y';
  var hasBg = false;
  var isLocked = true;
  var offsetY = 0;
  var naturalW = 0, naturalH = 0;
  var isDragging = false, dragStartY = 0, dragStartOffset = 0;

  var lockBtn = document.getElementById('hero-bg-lock');
  var lockSvg = document.getElementById('hero-bg-lock-svg');
  var unlockSvg = document.getElementById('hero-bg-unlock-svg');

  function clampOffset(val) {
    var sectionH = wrapper.offsetHeight;
    var sectionW = wrapper.offsetWidth;
    var scaledH = naturalW > 0 ? Math.round((naturalH / naturalW) * sectionW) : sectionH;
    var min = Math.min(0, -(scaledH - sectionH));
    return Math.max(min, Math.min(0, val));
  }

  function setPosition(y) {
    offsetY = y;
    wrapper.style.backgroundPosition = 'center ' + y + 'px';
  }

  function updateCursor() {
    wrapper.style.cursor = (hasBg && !isLocked) ? 'grab' : '';
  }

  function setLocked(locked) {
    isLocked = locked;
    lockSvg.style.display = locked ? '' : 'none';
    unlockSvg.style.display = locked ? 'none' : '';
    lockBtn.title = locked ? 'Unlock to drag background' : 'Lock background position';
    if (locked) lockBtn.classList.remove('is-active');
    else lockBtn.classList.add('is-active');
    updateCursor();
  }

  function applyBg(dataUrl, savedY) {
    var preload = document.getElementById('hero-bg-preload');
    if (preload) preload.remove();
    if (dataUrl) {
      var y = typeof savedY === 'number' ? savedY : 0;
      wrapper.style.backgroundImage = 'url(' + dataUrl + ')';
      wrapper.style.backgroundSize = '100% auto';
      wrapper.style.backgroundRepeat = 'no-repeat';
      wrapper.style.backgroundPosition = 'center ' + y + 'px';
      hasBg = true;
      lockBtn.style.display = '';
      clearBtn.style.display = '';
      var img = new Image();
      img.onload = function () {
        naturalW = img.naturalWidth;
        naturalH = img.naturalHeight;
        setPosition(clampOffset(y));
      };
      img.src = dataUrl;
    } else {
      wrapper.style.backgroundImage = '';
      wrapper.style.backgroundSize = '';
      wrapper.style.backgroundRepeat = '';
      wrapper.style.backgroundPosition = '';
      hasBg = false;
      naturalW = naturalH = offsetY = 0;
      lockBtn.style.display = 'none';
      clearBtn.style.display = 'none';
      setLocked(true);
    }
    updateCursor();
  }

  var saved = localStorage.getItem(KEY);
  var savedY = parseFloat(localStorage.getItem(KEY_Y)) || 0;
  if (saved) applyBg(saved, savedY);

  pickBtn.onclick = function () { fileInput.click(); };

  lockBtn.onclick = function () { setLocked(!isLocked); };

  clearBtn.onclick = function () {
    localStorage.removeItem(KEY);
    localStorage.removeItem(KEY_Y);
    applyBg(null);
  };

  fileInput.onchange = function (e) {
    var file = e.target.files && e.target.files[0];
    if (!file) return;
    var reader = new FileReader();
    reader.onload = function (ev) {
      var dataUrl = ev.target.result;
      localStorage.setItem(KEY, dataUrl);
      localStorage.removeItem(KEY_Y);
      applyBg(dataUrl, 0);
    };
    reader.readAsDataURL(file);
    fileInput.value = '';
  };

  // ── Vertical drag to reposition background ───────────────────
  wrapper.addEventListener('mousedown', function (e) {
    if (!hasBg || isLocked) return;
    if (e.target.closest('.hero-bg-zone')) return;
    isDragging = true;
    dragStartY = e.clientY;
    dragStartOffset = offsetY;
    wrapper.style.cursor = 'grabbing';
    e.preventDefault();
  });

  document.addEventListener('mousemove', function (e) {
    if (!isDragging) return;
    setPosition(clampOffset(dragStartOffset + (e.clientY - dragStartY)));
  });

  document.addEventListener('mouseup', function () {
    if (!isDragging) return;
    isDragging = false;
    updateCursor();
    localStorage.setItem(KEY_Y, offsetY);
  });
})();
