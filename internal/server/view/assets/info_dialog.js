
  (function() {
    var overlay  = document.getElementById('info-dialog');
    var titleEl  = document.getElementById('info-dialog-title');
    var bodyEl   = document.getElementById('info-dialog-body');
    var closeBtn = document.getElementById('info-dialog-close');

    function closeInfo() {
      if (window.__vaultrEscPop) window.__vaultrEscPop('info');
      overlay.style.display = 'none';
      bodyEl.innerHTML = '';
    }
    closeBtn.addEventListener('click', closeInfo);
    overlay.addEventListener('click', function(e) {
      if (e.target === overlay) closeInfo();
    });

    window.showInfo = function(opts) {
      titleEl.textContent  = opts.title      || '';
      bodyEl.innerHTML     = opts.bodyHTML   || '';
      closeBtn.textContent = opts.closeLabel || 'Got it';
      closeBtn.classList.toggle('error', !!opts.isError);
      overlay.style.display = 'flex';
      if (window.__vaultrEscPush) window.__vaultrEscPush('info', closeInfo);
    };

    window.showError = function(message, title) {
      window.showInfo({
        title:      title   || 'Error',
        bodyHTML:   '<p>' + (message || 'An unexpected error occurred.') + '</p>',
        closeLabel: 'OK',
        isError:    true,
      });
    };
  })();
