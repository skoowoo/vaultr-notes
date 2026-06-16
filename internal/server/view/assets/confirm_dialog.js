
  function confirmDialogCtrl() {
    return {
      open: false,
      title: '',
      titleHTML: '',
      message: '',
      confirmLabel: 'Confirm',
      danger: true,
      threeButton: false,
      altLabel: '',
      altDanger: true,
      _resolve: null,

      init() { window._confirmDialogCtrl = this; },

      show(opts) {
        opts = opts || {};
        this.titleHTML     = opts.titleHTML     || '';
        this.title         = this.titleHTML ? '' : (opts.title || 'Are you sure?');
        this.message       = opts.message       || '';
        this.confirmLabel  = opts.confirmLabel  || 'Confirm';
        this.altLabel      = opts.altLabel      || '';
        this.threeButton   = !!this.altLabel;
        this.altDanger     = opts.altDanger !== false;
        if (this.threeButton) {
          this.danger = opts.danger === true;
        } else {
          this.danger = opts.danger !== false;
        }
        this.open = true;
        var _self = this;
        if (window.__vaultrEscPush) window.__vaultrEscPush('confirm', function() { _self.cancel(); });
        return new Promise(function(resolve) {
          window._confirmDialogCtrl._resolve = resolve;
        });
      },

      confirm() {
        if (window.__vaultrEscPop) window.__vaultrEscPop('confirm');
        this.open = false;
        if (this._resolve) {
          this._resolve(this.threeButton ? 'confirm' : true);
          this._resolve = null;
        }
      },
      chooseAlt() {
        if (window.__vaultrEscPop) window.__vaultrEscPop('confirm');
        this.open = false;
        if (this._resolve) { this._resolve('alt'); this._resolve = null; }
      },
      cancel() {
        if (window.__vaultrEscPop) window.__vaultrEscPop('confirm');
        this.open = false;
        if (this._resolve) { this._resolve(false); this._resolve = null; }
      },
    };
  }

  // Global helper — call from any page that includes confirmDialogHTML.
  window.showConfirm = function(opts) {
    if (!window._confirmDialogCtrl) return Promise.resolve(false);
    return window._confirmDialogCtrl.show(opts);
  };
