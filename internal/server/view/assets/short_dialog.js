
  (function() {
    var overlay     = document.getElementById('short-dialog');
    var textarea    = document.getElementById('short-textarea');
    var placeholder = document.getElementById('short-placeholder');
    var phHint      = document.getElementById('short-ph-hint');
    var fadeTop      = document.getElementById('short-fade-top');
    var saveConfirm  = document.getElementById('short-save-confirm');
    if (!overlay) return;

    var isMac = /Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent);
    if (phHint) phHint.textContent = isMac ? '⌘↵ save  ·  esc exit' : 'Ctrl+Enter save  ·  Esc exit';

    var saving = false;
    var isOpen = false;

    function updatePlaceholder() {
      if (placeholder) placeholder.style.display = textarea.value.length === 0 ? '' : 'none';
    }

    function setTrafficLights(visible) {
      if (window.vaultrDesktop && window.vaultrDesktop.setWindowButtonVisibility) {
        window.vaultrDesktop.setWindowButtonVisibility(visible);
      }
    }

    function openShort() {
      isOpen = true;
      overlay.classList.remove('is-closing');
      overlay.style.display = 'flex';
      setTrafficLights(false);
      if (window.__vaultrEscPush) window.__vaultrEscPush('short', tryClose);
      setTimeout(function() { textarea.focus(); }, 30);
      updatePlaceholder();
    }

    function closeShort() {
      isOpen = false;
      if (window.__vaultrEscPop) window.__vaultrEscPop('short');
      overlay.classList.add('is-closing');
      setTimeout(function() {
        setTrafficLights(true);
        overlay.classList.remove('is-closing');
        overlay.style.display = 'none';
        textarea.classList.remove('is-saved');
        if (saveConfirm) saveConfirm.classList.remove('is-visible');
        textarea.value = '';
        textarea.scrollTop = 0;
        if (fadeTop) fadeTop.classList.remove('is-visible');
        saving = false;
        updatePlaceholder();
      }, 280);
    }

    async function tryClose() {
      if (!textarea.value.trim()) { closeShort(); return; }
      var ok = await window.showConfirm({
        title:        'Discard short note?',
        message:      'Your draft will be lost.',
        confirmLabel: 'Discard',
        danger:       true,
      });
      if (ok) closeShort();
    }

    async function saveShort() {
      var content = textarea.value.trim();
      if (!content || saving) return;
      saving = true;
      try {
        var resp = await fetch('/api/vault/shorts', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ content: content }),
        });
        if (!resp.ok) {
          var msg = await resp.text();
          throw new Error(msg || 'Save failed');
        }
        textarea.classList.add('is-saved');
        if (placeholder) placeholder.style.display = 'none';
        if (saveConfirm) saveConfirm.classList.add('is-visible');
        textarea.value = '';
        await new Promise(function(r) { setTimeout(r, 1400); });
        closeShort();
        if (window.__vaultrAfterVaultMutation) await window.__vaultrAfterVaultMutation();
      } catch (err) {
        saving = false;
        textarea.classList.remove('is-saved');
        if (saveConfirm) saveConfirm.classList.remove('is-visible');
        window.showError('Failed to save: ' + (err && err.message ? err.message : String(err)), 'Save failed');
      }
    }

    textarea.addEventListener('input', updatePlaceholder);
    textarea.addEventListener('scroll', function() {
      if (fadeTop) fadeTop.classList.toggle('is-visible', textarea.scrollTop > 8);
    });
    overlay.addEventListener('click', function(e) {
      if (e.target !== textarea) textarea.focus();
    });

    textarea.addEventListener('keydown', function(e) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
        e.preventDefault();
        saveShort();
        return;
      }
      if ((e.metaKey || e.ctrlKey) && !e.shiftKey && (e.key === 'z' || e.key === 'Z')) {
        e.preventDefault();
        document.execCommand('undo');
        return;
      }
      if ((e.metaKey || e.ctrlKey) && e.shiftKey && (e.key === 'z' || e.key === 'Z')) {
        e.preventDefault();
        document.execCommand('redo');
        return;
      }
    });

    window.__vaultrHotkeys.register('short', '.', function() {
      if (isOpen) { tryClose(); } else { openShort(); }
    });

    window.openShortDialog = openShort;
  })();
