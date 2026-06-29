// __vaultrPathAcCreate — shared directory autocomplete factory.
// Creates a self-contained autocomplete controller bound to a specific
// input element and dropdown list element.
//
// opts:
//   getInput()            → HTMLInputElement | HTMLTextAreaElement
//   getList()             → HTMLUListElement  (the dropdown)
//   parseCtx(val, caret)  → { dirPath, partial, replaceStart } | null
//   onApply(el, newVal, caretPos) → void  (update element value + cursor)
//   escKey                → string key for __vaultrEscPush / __vaultrEscPop
//
function __vaultrPathAcCreate(opts) {
  var st = { seq: 0, abort: null, tick: null, active: -1, fetchedDir: null, cachedDirs: [] };

  function close() {
    if (st.abort) { st.abort.abort(); st.abort = null; }
    clearTimeout(st.tick); st.tick = null;
    var list = opts.getList();
    if (list) {
      list.classList.remove('open'); list.hidden = true;
      list.setAttribute('aria-expanded', 'false'); list.innerHTML = '';
    }
    st.active = -1;
    if (window.__vaultrEscPop && opts.escKey) window.__vaultrEscPop(opts.escKey);
  }

  function listItems() {
    var list = opts.getList();
    return list ? list.querySelectorAll('li[role="option"]') : [];
  }

  function setActive(ix) {
    var items = listItems(); if (!items.length) return;
    if (ix < 0) ix = 0; if (ix >= items.length) ix = items.length - 1;
    st.active = ix;
    items.forEach(function(el, j) { el.setAttribute('aria-selected', j === ix ? 'true' : 'false'); });
    var sel = null;
    items.forEach(function(el) { if (el.getAttribute('aria-selected') === 'true') sel = el; });
    if (sel) sel.scrollIntoView({ block: 'nearest' });
  }

  function render(filtered, emptyMsg) {
    var list = opts.getList(); if (!list) return;
    list.innerHTML = '';
    if (!filtered.length) {
      var li0 = document.createElement('li');
      li0.className = 'path-ac-muted'; li0.textContent = emptyMsg || 'No folders';
      li0.setAttribute('role', 'presentation'); list.appendChild(li0);
      st.active = -1; return;
    }
    filtered.forEach(function(name, i) {
      var li = document.createElement('li');
      li.setAttribute('role', 'option'); li.setAttribute('data-name', name);
      li.setAttribute('aria-selected', i === 0 ? 'true' : 'false');
      li.textContent = name + '/';
      li.addEventListener('mousedown', function(ev) { ev.preventDefault(); apply(name); });
      list.appendChild(li);
    });
    st.active = 0;
  }

  function doFilter(ctx) {
    var pref = (ctx.partial || '').toLowerCase();
    var filtered = st.cachedDirs.filter(function(d) { return !pref || d.toLowerCase().indexOf(pref) === 0; });
    var noMatch = !st.cachedDirs.length ? 'No folders' : 'No match — new folders are created on Publish';
    render(filtered, filtered.length ? '' : noMatch);
    var list = opts.getList();
    if (list) { list.classList.add('open'); list.hidden = false; list.setAttribute('aria-expanded', 'true'); }
    if (window.__vaultrEscPush && opts.escKey) window.__vaultrEscPush(opts.escKey, close);
  }

  async function doFetch(ctx0) {
    var mySeq = st.seq;
    st.abort = new AbortController();
    try {
      var resp = await fetch('/api/vault/list-dirs', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: ctx0.dirPath }), signal: st.abort.signal,
      });
      if (mySeq !== st.seq) return;
      if (!resp.ok) { close(); return; }
      var data = await resp.json();
      st.fetchedDir = typeof data.path === 'string' ? data.path : ctx0.dirPath;
      st.cachedDirs = Array.isArray(data.dirs) ? data.dirs : [];
      var input = opts.getInput(); if (!input) return;
      var ctxNow = opts.parseCtx(input.value, input.selectionStart);
      if (!ctxNow || ctxNow.dirPath !== st.fetchedDir) return;
      doFilter(ctxNow);
    } catch(e) { if (e.name !== 'AbortError') close(); }
  }

  function schedule() {
    if (st.abort) { st.abort.abort(); st.abort = null; }
    clearTimeout(st.tick); st.seq++;
    st.tick = setTimeout(function() {
      st.tick = null;
      var input = opts.getInput(); if (!input) return;
      var ctx = opts.parseCtx(input.value, input.selectionStart);
      if (!ctx) { close(); return; }
      void doFetch(ctx);
    }, 160);
  }

  function refresh() {
    var input = opts.getInput(); if (!input) return;
    var ctx = opts.parseCtx(input.value, input.selectionStart);
    if (!ctx) { close(); return; }
    if (st.fetchedDir !== null && ctx.dirPath === st.fetchedDir) { doFilter(ctx); return; }
    schedule();
  }

  function apply(name) {
    var input = opts.getInput(); if (!input) return;
    var ctx = opts.parseCtx(input.value, input.selectionStart);
    if (!ctx) { close(); return; }
    var prefix = input.value.slice(0, ctx.replaceStart);
    var suffix = input.value.slice(input.selectionStart);
    var insert = name + '/';
    var nextCaret = prefix.length + insert.length;
    var newVal = prefix + insert + suffix;
    if (opts.onApply) {
      opts.onApply(input, newVal, nextCaret);
    } else {
      input.value = newVal;
      if (input.setSelectionRange) input.setSelectionRange(nextCaret, nextCaret);
      input.focus();
    }
    close();
    refresh();
  }

  function isOpen() {
    var list = opts.getList();
    return !!(list && list.classList.contains('open'));
  }

  // Call from a keydown handler. Returns true if the event was consumed.
  function handleKeydown(ev) {
    var open = isOpen();
    var items = open ? listItems() : [];
    if (open && items.length) {
      if (ev.key === 'ArrowDown') {
        ev.preventDefault();
        var nDown = st.active < 0 ? 0 : st.active + 1;
        setActive(nDown >= items.length ? 0 : nDown); return true;
      }
      if (ev.key === 'ArrowUp') {
        ev.preventDefault();
        var pUp = st.active < 0 ? items.length - 1 : st.active - 1;
        setActive(pUp < 0 ? items.length - 1 : pUp); return true;
      }
      if (ev.key === 'Tab') {
        ev.preventDefault();
        var pickT = items[st.active < 0 ? 0 : st.active];
        var nmT = pickT && pickT.getAttribute('data-name');
        if (nmT) apply(nmT); return true;
      }
      if (ev.key === 'Escape') { close(); return true; }
      if (ev.key === 'Enter') {
        var pickE = items[st.active < 0 ? 0 : st.active];
        var nmE = pickE && pickE.getAttribute('data-name');
        if (nmE) { ev.preventDefault(); apply(nmE); return true; }
      }
    }
    return false;
  }

  return { close: close, refresh: refresh, apply: apply, schedule: schedule, handleKeydown: handleKeydown, isOpen: isOpen };
}
