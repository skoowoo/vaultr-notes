
  function shortDialogCtrl() {
    return window.vaultrOverlay('short', {
      content: '',
      hint: '',
      visible: false,
      closing: false,
      saving: false,
      isSaved: false,
      showSaveConfirm: false,
      fadeTop: false,

      init() {
        var isMac = /Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent);
        this.hint = isMac ? '⌘↵ save  ·  esc exit' : 'Ctrl+Enter save  ·  Esc exit';
        var _self = this;
        window.openShortDialog = function() { _self.openShort(); };
        window.__vaultrHotkeys.register('short', '.', function() {
          if (_self.open) { _self.tryClose(); } else { _self.openShort(); }
        });
      },

      _setTrafficLights(show) {
        if (window.vaultrDesktop && window.vaultrDesktop.setWindowButtonVisibility) {
          window.vaultrDesktop.setWindowButtonVisibility(show);
        }
      },

      openShort() {
        this.closing = false;
        this.visible = true;
        this._setTrafficLights(false);
        var _self = this;
        this.openOverlay(function() { _self.tryClose(); });
        this.$nextTick(function() {
          setTimeout(function() { if (_self.$refs.textarea) _self.$refs.textarea.focus(); }, 30);
        });
      },

      onOverlayClick(e) {
        if (e.target !== this.$refs.textarea) this.$refs.textarea.focus();
      },

      onKeydown(e) {
        if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
          e.preventDefault();
          this.saveShort();
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
      },

      _finishClose() {
        this._setTrafficLights(true);
        this.closing = false;
        this.visible = false;
        this.isSaved = false;
        this.showSaveConfirm = false;
        this.content = '';
        if (this.$refs.textarea) this.$refs.textarea.scrollTop = 0;
        this.fadeTop = false;
        this.saving = false;
      },

      closeShort() {
        this.closeOverlay();
        this.closing = true;
        var _self = this;
        setTimeout(function() { _self._finishClose(); }, 280);
      },

      async tryClose() {
        if (!this.content.trim()) { this.closeShort(); return; }
        var ok = await window.showConfirm({
          title:        'Discard short note?',
          message:      'Your draft will be lost.',
          confirmLabel: 'Discard',
          danger:       true,
        });
        if (ok) this.closeShort();
      },

      async saveShort() {
        var text = this.content.trim();
        if (!text || this.saving) return;
        this.saving = true;
        try {
          var resp = await fetch('/api/vault/shorts', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ content: text }),
          });
          if (!resp.ok) {
            var msg = await resp.text();
            throw new Error(msg || 'Save failed');
          }
          this.isSaved = true;
          this.showSaveConfirm = true;
          this.content = '';
          await new Promise(function(r) { setTimeout(r, 1400); });
          this.closeShort();
          if (window.__vaultrAfterVaultMutation) await window.__vaultrAfterVaultMutation();
        } catch (err) {
          this.saving = false;
          this.isSaved = false;
          this.showSaveConfirm = false;
          window.showError('Failed to save: ' + (err && err.message ? err.message : String(err)), 'Save failed');
        }
      },
    });
  }
