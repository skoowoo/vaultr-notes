package view

// navCSS is the stylesheet for the shared left navigation strip.
// Drop it inside a page's <style> block.
const navCSS = `
    /* ── Left nav — theme-aware sidebar ───────────────────────── */
    .lib-nav {
      flex-shrink: 0; width: var(--nav-w); height: 100%;
      background: var(--nav-bg); border-right: 1px solid var(--hr);
      display: flex; flex-direction: column; align-items: center;
      padding: 0.875rem 0; gap: 0.125rem;
    }
    .nav-item {
      display: flex; align-items: center; justify-content: center;
      width: var(--nav-item-sz); height: var(--nav-item-sz); border-radius: var(--radius-md);
      color: rgba(255,255,255,0.38); transition: color 120ms, background 120ms;
      text-decoration: none; border: none; background: transparent; padding: 0;
    }
    .nav-item:hover { color: rgba(255,255,255,0.78); background: rgba(255,255,255,0.06); }
    .nav-item.active { background: var(--link); color: #ffffff; }
    html[data-theme="light"] .nav-item { color: #9ca3af; }
    html[data-theme="light"] .nav-item:hover { color: #111111; background: rgba(17,17,17,0.05); }
    html[data-theme="light"] .nav-item.active { background: var(--link); color: #ffffff; }
    .nav-item svg { width: 19px; height: 19px; }
    .nav-spacer { flex: 1; }
    .lib-nav, .nav-item { user-select: none; }
    .nav-item-wrap { position: relative; display: flex; align-items: center; justify-content: center; }
    @property --_nb-a { syntax: '<angle>'; inherits: false; initial-value: 0deg; }
    .nav-run-badge {
      display: none; position: absolute; top: -4px; right: -4px;
      min-width: 14px; height: 14px; padding: 0 3px;
      background-color: var(--link);
      background-image: conic-gradient(
        from var(--_nb-a) at 50% 50%,
        transparent 0%, transparent 50%,
        rgba(255,255,255,0.90) 65%,
        rgba(255,255,255,0.20) 78%,
        transparent 88%
      );
      color: var(--canvas);
      font-size: var(--text-2xs); font-weight: var(--fw-semibold); line-height: 14px;
      border-radius: var(--radius-pill); text-align: center;
      animation: nav-badge-chase 1.5s linear infinite;
    }
    @keyframes nav-badge-chase { to { --_nb-a: 360deg; } }`

// svgHome is the Lucide "Layers" icon used for the Home nav item.
const svgHome = `<svg fill="none" stroke="currentColor" stroke-width="1.7" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83z"/>
        <path stroke-linecap="round" stroke-linejoin="round" d="M2 12a1 1 0 0 0 .58.91l8.6 3.91a2 2 0 0 0 1.65 0l8.58-3.9A1 1 0 0 0 22 12"/>
        <path stroke-linecap="round" stroke-linejoin="round" d="M2 17a1 1 0 0 0 .58.91l8.6 3.91a2 2 0 0 0 1.65 0l8.58-3.9A1 1 0 0 0 22 17"/>
      </svg>`

// svgAgent is the Lucide "Users" icon used for the Mate nav item.
const svgAgent = `<svg fill="none" stroke="currentColor" stroke-width="1.7" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/>
        <circle cx="9" cy="7" r="4"/>
        <path stroke-linecap="round" stroke-linejoin="round" d="M16 3.128a4 4 0 0 1 0 7.744"/>
        <path stroke-linecap="round" stroke-linejoin="round" d="M22 21v-2a4 4 0 0 0-3-3.87"/>
      </svg>`

// svgSettings is the Lucide "Settings" icon used for the Settings nav item.
const svgSettings = `<svg fill="none" stroke="currentColor" stroke-width="1.7" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M9.671 4.136a2.34 2.34 0 0 1 4.659 0 2.34 2.34 0 0 0 3.319 1.915 2.34 2.34 0 0 1 2.33 4.033 2.34 2.34 0 0 0 0 3.831 2.34 2.34 0 0 1-2.33 4.033 2.34 2.34 0 0 0-3.319 1.915 2.34 2.34 0 0 1-4.659 0 2.34 2.34 0 0 0-3.32-1.915 2.34 2.34 0 0 1-2.33-4.033 2.34 2.34 0 0 0 0-3.831A2.34 2.34 0 0 1 6.35 6.051a2.34 2.34 0 0 0 3.319-1.915"/>
        <circle cx="12" cy="12" r="3"/>
      </svg>`

// topbarCSS is the stylesheet for the shared 40px horizontal title bar.
// Drop it inside a page's <style> block alongside navCSS.
const topbarCSS = `
    /* ── Top bar ─────────────────────────────────────────────── */
    .lib-topbar {
      flex-shrink: 0; height: var(--topbar-h);
      display: flex; align-items: center;
      padding: 0 1rem; border-bottom: 1px solid var(--hr); background: var(--bg);
      user-select: none;
    }
    .lib-topbar-left { display: flex; align-items: center; gap: 0.5rem; min-width: 0; }
    .lib-topbar-spacer { flex: 1; min-width: 0; }
    .lib-topbar-actions { display: flex; align-items: center; gap: 0.5rem; min-width: 0; }
    .lib-status {
      display: inline-flex; align-items: center; gap: 0.28rem; padding: 0 0.25rem;
      color: var(--muted); font-size: var(--text-sm); line-height: 1;
    }
    .lib-status-value { color: var(--muted); opacity: 0.8; font-weight: 500; font-variant-numeric: tabular-nums; }
    .lib-status-label { letter-spacing: 0.01em; }
    .lib-action-btn {
      display: inline-flex; align-items: center; justify-content: center;
      width: var(--action-btn-sz); height: var(--action-btn-sz); padding: 0; border-radius: var(--radius-sm); border: none;
      background: transparent; color: var(--muted); cursor: pointer;
      transition: color 120ms, background 120ms;
    }
    .lib-action-btn:hover, .lib-action-btn.is-active { color: var(--fg); background: var(--nav-act); }
    .lib-action-btn.is-active { color: var(--link); }
    .lib-action-btn svg { width: 15px; height: 15px; flex-shrink: 0; }
    .lib-action-btn .lib-ai-px { display: none; }
    .lib-back-btn .lib-ai-px { display: none; }
    .lib-action-btn.spinning svg { animation: lib-spin 0.6s linear infinite; }
    @keyframes lib-spin { to { transform: rotate(360deg); } }
    .lib-back-btn {
      display: inline-flex; align-items: center; gap: 0.3rem;
      height: var(--action-btn-sz); padding: 0 0.5rem 0 0.35rem;
      border-radius: var(--radius-sm); border: none; background: transparent;
      color: var(--muted); cursor: pointer; text-decoration: none;
      font-size: var(--text-base); font-weight: 500; white-space: nowrap;
      transition: color 120ms, background 120ms;
    }
    .lib-back-btn:hover { color: var(--fg); background: var(--nav-act); }
    .lib-back-btn svg { width: 15px; height: 15px; flex-shrink: 0; }
    html.macos .lib-topbar { -webkit-app-region: drag; padding-left: 76px; cursor: default; }
    html.macos .lib-topbar button,
    html.macos .lib-topbar a { -webkit-app-region: no-drag; }`

// navHTML returns the complete <nav class="lib-nav">…</nav> HTML block.
// active should be the page's own name: "home", "agent", "settings",
// or a sub-page name ("images", "shorts", "dir", "library") — sub-pages highlight their parent.
func navHTML(active string) string {
	// sub-pages of Home highlight the Home nav item
	if active == "images" || active == "shorts" || active == "dir" || active == "library" || active == "folders" {
		active = "home"
	}
	cls := func(page string) string {
		if page == active {
			return `class="nav-item active"`
		}
		return `class="nav-item"`
	}
	return `  <nav class="lib-nav">
    <a href="/home" ` + cls("home") + ` title="Home">
      ` + svgHome + `
    </a>
    <a href="/agent" ` + cls("agent") + ` title="Agent Chat">
      <span class="nav-item-wrap">
        ` + svgAgent + `
        <span class="nav-run-badge" id="_nav-agent-badge"></span>
      </span>
    </a>
    <div class="nav-spacer"></div>
    <button type="button" ` + cls("settings") + ` title="Settings" @click="$store.settingsModal.open = true; $el.blur()">
      ` + svgSettings + `
    </button>
  </nav>
  <script>
  (function(){
    var _t = null;
    function _poll() {
      fetch('/api/runs/active').then(function(r){return r.json();}).then(function(d){
        var b = document.getElementById('_nav-agent-badge');
        if (!b) return;
        var n = (d && d.count) ? d.count : 0;
        b.textContent = n > 9 ? '9+' : String(n);
        b.style.display = n > 0 ? 'block' : 'none';
      }).catch(function(){});
      _t = setTimeout(_poll, 3000);
    }
    document.addEventListener('visibilitychange', function(){
      if (document.hidden) { clearTimeout(_t); _t = null; }
      else { _poll(); }
    });
    if (!document.hidden) { _poll(); }
  })();
  document.addEventListener('DOMContentLoaded', function() {
    window.__vaultrHotkeys.register('nav-home', '1', function() {
      if (window.__vaultrAnyModalOpen && window.__vaultrAnyModalOpen()) return;
      window.location.href = '/home';
    });
    window.__vaultrHotkeys.register('nav-agent', '2', function() {
      if (window.__vaultrAnyModalOpen && window.__vaultrAnyModalOpen()) return;
      window.location.href = '/agent';
    });
    window.__vaultrHotkeys.register('refresh', 'r', function() {
      if (typeof window.__vaultrBackgroundRefresh === 'function') {
        window.__vaultrBackgroundRefresh();
      } else {
        window.location.reload();
      }
    });
  });
  </script>`
}
