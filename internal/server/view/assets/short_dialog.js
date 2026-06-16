
  (function() {
    var overlay   = document.getElementById('short-dialog');
    var textarea  = document.getElementById('short-textarea');
    var saveBtn   = document.getElementById('short-save-btn');
    var cancelBtn = document.getElementById('short-cancel-btn');
    var closeBtn  = document.getElementById('short-close-btn');
    var charEl    = document.getElementById('short-charcount');
    var hintEl    = document.getElementById('short-kbd-hint');
    if (!overlay) return;

    var isMac = /Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent);
    if (hintEl) hintEl.textContent = isMac ? '⌘+Return to save' : 'Ctrl+Enter to save';

    function updateCount() {
      var n = textarea.value.length;
      charEl.textContent = n === 1 ? '1 char' : n + ' chars';
      saveBtn.disabled = n === 0;
    }

    function openShort() {
      overlay.style.display = 'flex';
      if (window.__vaultrEscPush) window.__vaultrEscPush('short', tryClose);
      setTimeout(function() { textarea.focus(); }, 30);
      updateCount();
    }

    function closeShort() {
      if (window.__vaultrEscPop) window.__vaultrEscPop('short');
      overlay.style.display = 'none';
      textarea.value = '';
      updateCount();
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
      if (!content || saveBtn.disabled) return;
      var prevLabel = saveBtn.textContent;
      saveBtn.disabled = true;
      saveBtn.textContent = 'Saving…';
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
        closeShort();
        if (window.__vaultrAfterVaultMutation) await window.__vaultrAfterVaultMutation();
      } catch (err) {
        saveBtn.disabled = false;
        saveBtn.textContent = prevLabel;
        window.showError('Failed to save: ' + (err && err.message ? err.message : String(err)), 'Save failed');
      }
    }

    textarea.addEventListener('input', updateCount);
    closeBtn.addEventListener('click', tryClose);
    cancelBtn.addEventListener('click', tryClose);
    saveBtn.addEventListener('click', saveShort);

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
      var isOpen = overlay.style.display !== 'none';
      if (isOpen) { tryClose(); } else { openShort(); }
    });

    window.openShortDialog = openShort;
  })();
