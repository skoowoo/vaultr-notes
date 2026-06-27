package view

// navCSS is the stylesheet for the shared left navigation strip.
// Drop it inside a page's <style> block.
const navCSS = `
    /* ── Left nav ─────────────────────────────────────────────── */
    .lib-nav {
      flex-shrink: 0; width: var(--nav-w); height: 100%;
      background: var(--nav-bg); border-right: 2px solid var(--hr);
      display: flex; flex-direction: column; align-items: center;
      padding: 0.875rem 0; gap: 0.375rem;
      view-transition-name: page-nav;
    }
    .nav-item {
      display: flex; align-items: center; justify-content: center;
      width: var(--nav-item-sz); height: var(--nav-item-sz);
      color: rgba(0,0,0,0.52);
      text-decoration: none; border: none; background: transparent; padding: 0;
    }
    .nav-item:hover { color: var(--fg); background: rgba(0,0,0,0.1); }
    .nav-item.active { background: var(--seg-act-bg); color: var(--seg-act-fg); box-shadow: var(--px-d1) var(--px-shadow); }
    .nav-item svg { width: 19px; height: 19px; }
    .nav-spacer { flex: 1; }
    .lib-nav, .nav-item { user-select: none; }
    .nav-item-wrap { position: relative; display: flex; align-items: center; justify-content: center; }
    .nav-compose-btn {
      display: flex; align-items: center; justify-content: center;
      width: 30px; height: 30px;
      background: var(--nav-act);
      border: none;
      color: rgba(0,0,0,0.55); cursor: pointer; padding: 0;
      user-select: none;
    }
    .nav-compose-btn:hover { background: var(--seg-act-bg); color: var(--seg-act-fg); }
    .nav-compose-btn svg { width: 17px; height: 17px; flex-shrink: 0; }
    .nav-run-badge {
      display: none; position: absolute; top: -4px; right: -4px;
      min-width: 14px; height: 14px; padding: 0 3px;
      background-color: var(--seg-act-bg);
      color: var(--accent);
      font-size: var(--text-2xs); font-weight: var(--fw-semibold); line-height: 14px;
      text-align: center;
    }`

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

// svgCompose is the Lucide "Plus" icon used for the quick new-note button.
const svgCompose = `<svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M12 5v14M5 12h14"/></svg>`

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
      padding: 0 1rem; border-bottom: 2px solid var(--hr); background: var(--bg);
      user-select: none;
    }
    .lib-topbar-left { display: flex; align-items: center; gap: 0.5rem; min-width: 0; }
    .lib-topbar-spacer { flex: 1; min-width: 0; }
    .lib-topbar-actions { display: flex; align-items: center; gap: 0.5rem; min-width: 0; }
    .lib-status {
      display: inline-flex; align-items: center; gap: 0.28rem; padding: 0 7px;
      color: var(--muted); font-size: var(--text-sm); line-height: 1;
      border: 1px solid var(--hr); height: 22px;
    }
    .lib-status-value { color: var(--muted); opacity: 0.8; font-weight: 500; font-variant-numeric: tabular-nums; }
    .lib-status-label { letter-spacing: 0.01em; }
    .lib-action-btn {
      display: inline-flex; align-items: center; justify-content: center;
      width: var(--action-btn-sz); height: var(--action-btn-sz); padding: 0; border: none;
      background: transparent; color: var(--muted); cursor: pointer;
    }
    .lib-action-btn:hover, .lib-action-btn.is-active { color: var(--fg); background: var(--icon-hov); }
    .lib-action-btn.is-active { color: var(--ui-accent); }
    .lib-action-btn svg { width: 15px; height: 15px; flex-shrink: 0; }
    .lib-action-btn .lib-ai-px { display: none; }
    .lib-back-btn .lib-ai-px { display: none; }
    .lib-action-btn.spinning svg { animation: lib-spin 0.6s linear infinite; }
    @keyframes lib-spin { to { transform: rotate(360deg); } }
    .lib-back-btn {
      display: inline-flex; align-items: center; gap: 0.3rem;
      height: var(--action-btn-sz); padding: 0 0.5rem 0 0.35rem;
      border: none; background: transparent;
      color: var(--muted); cursor: pointer; text-decoration: none;
      font-size: var(--text-base); font-weight: 500; white-space: nowrap;
    }
    .lib-back-btn:hover { color: var(--fg); background: var(--icon-hov); }
    .lib-back-btn svg { width: 15px; height: 15px; flex-shrink: 0; }
    html.macos .lib-topbar { -webkit-app-region: drag; padding-left: 76px; cursor: default; }
    html.macos .lib-topbar button,
    html.macos .lib-topbar a { -webkit-app-region: no-drag; }`

// navHTML returns the complete <nav class="lib-nav">…</nav> HTML block.
// active should be the page's own name: "home", "agent", "graph", "settings",
// or a sub-page name ("images", "shorts", "dir", "library") — sub-pages highlight their parent.
func navHTML(active string) string {
	// sub-pages of Home highlight the Home nav item
	if active == "images" || active == "shorts" || active == "dir" || active == "library" || active == "folders" || active == "graph" {
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
    <button type="button" class="nav-compose-btn" title="New note (Ctrl+N)"
            style="margin-top:0.75rem"
            onclick="window.__vaultrDrawer && void window.__vaultrDrawer.openNewInDrawer('','')">
      ` + svgCompose + `
    </button>
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

// topbarActionsHTML returns the shared right-side action button group used on
// every page: drawer toggle → short note → search → reload.
//
//   - reloadClick: JS expression for the reload @click (e.g. "refresh()", "window.location.reload()")
//   - reloadTitle: tooltip text (e.g. "Refresh", "Refresh home")
//   - reloadExtraClass: Alpine :class value for the reload button; pass "" for none (graph passes "loading && 'spinning'")
//   - searchExtraDetail: extra detail object for the open-search CustomEvent; pass "" for none (graph passes "{ mode: 'knowledge' }")
func topbarActionsHTML(reloadClick, reloadTitle, reloadExtraClass, searchExtraDetail string) string {
	const macCheck = `/Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent)`

	reloadClassAttr := ""
	if reloadExtraClass != "" {
		reloadClassAttr = ` :class="` + reloadExtraClass + `"`
	}

	searchEvent := "new CustomEvent('open-search')"
	if searchExtraDetail != "" {
		searchEvent = "new CustomEvent('open-search', { detail: " + searchExtraDetail + " })"
	}

	return `    <div class="lib-topbar-actions">
      <button type="button" class="lib-action-btn"
              :class="{ 'is-active': drawerOpen }"
              :title="` + macCheck + ` ? 'Reading drawer (⌘E)' : 'Reading drawer (Ctrl+E)'"
              @click="drawerOpen = !drawerOpen">` + topbarIconPanel + `</button>
      ` + shortTriggerButton + `
      <button type="button" class="lib-action-btn"` + reloadClassAttr + ` title="` + reloadTitle + `" @click="` + reloadClick + `">` + topbarIconReload + `</button>
      <button type="button" class="lib-action-btn"
              :title="` + macCheck + ` ? 'Search (⌘K)' : 'Search (Ctrl+K)'"
              @click="window.dispatchEvent(` + searchEvent + `)">` + topbarIconSearch + `</button>
    </div>`
}
