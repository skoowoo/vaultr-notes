package view

// navCSS is the stylesheet for the shared left navigation strip. Deliberately
// neutral — this rail is pure function (new note, drawer, short note,
// search, settings). Its is-active state uses the same --control-active-bg/fg
// surface lift as every other "currently selected" indicator app-wide (see
// shared_tokens.go); only the compose button breaks from that, as the app's
// one primary CTA. Drop it inside a page's <style> block.
const navCSS = `
    /* ── Left nav ─────────────────────────────────────────────── */
    .lib-nav {
      flex-shrink: 0; width: var(--nav-w); height: 100%;
      background: var(--surface-soft); border-right: var(--bd-w) solid var(--border);
      border-radius: 0; /* edge-docked chrome, flush to the viewport — never rounds */
      display: flex; flex-direction: column; align-items: center;
      padding: 14px 0; gap: 6px;
      view-transition-name: page-nav;
    }
    .nav-item {
      display: flex; align-items: center; justify-content: center;
      width: var(--nav-item-sz); height: var(--nav-item-sz);
      color: var(--muted); border-radius: var(--r-sm);
      text-decoration: none; border: none; background: transparent; padding: 0;
    }
    .nav-item:hover { color: var(--fg); background: var(--card-hov); }
    .nav-item.is-active { background: var(--control-active-bg); color: var(--control-active-fg); }
    .nav-item svg { width: 14px; height: 14px; }
    .nav-spacer { flex: 1; }
    .lib-nav, .nav-item { user-select: none; }
    /* The one deliberately bolder icon — new note is the primary action, so
       it's the app's one button-primary: accent fill, exactly like the
       drawer's publish button and every other primary CTA. Circular and a
       touch smaller than the other nav slots so it reads as a compact
       accent mark rather than a heavy block. */
    .nav-compose-btn {
      display: flex; align-items: center; justify-content: center;
      width: var(--action-btn-sz); height: var(--action-btn-sz);
      background: var(--accent); border-radius: 50%;
      border: none;
      color: var(--accent-fg); cursor: pointer; padding: 0;
      user-select: none;
    }
    .nav-compose-btn:hover { background: var(--accent-hov); }
    .nav-compose-btn svg { width: 14px; height: 14px; flex-shrink: 0; }`

// svgCompose is the Lucide "Plus" icon used for the quick new-note button.
const svgCompose = `<svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M12 5v14M5 12h14"/></svg>`

// svgCheck is the Lucide "Check" icon, used wherever a small smooth checkmark
// is needed (e.g. Inbox's mark-all-read button). Kept distinct from the
// pixelarticons svgPx* set, which is not used for icon-only action buttons.
const svgCheck = `<svg fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M20 6 9 17l-5-5"/></svg>`

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
      padding: 0 1rem; border-bottom: var(--bd-w) solid var(--border); background: var(--bg);
      border-radius: 0; /* edge-docked chrome, flush to the viewport — never rounds */
      user-select: none;
    }
    .lib-topbar-spacer { flex: 1; min-width: 0; }
    .lib-topbar-actions { display: flex; align-items: center; gap: 0.5rem; min-width: 0; }
    .lib-status {
      display: inline-flex; align-items: center; gap: 0.28rem; padding: 0 7px;
      color: var(--muted); font-size: var(--text-sm); line-height: 1;
      border: 1px solid var(--border); height: 22px;
      border-radius: var(--r-full);
    }
    .lib-status-value { color: var(--muted); opacity: 0.8; font-weight: 500; font-variant-numeric: tabular-nums; }
    .lib-status-label { letter-spacing: 0.01em; }
    .lib-action-btn {
      display: inline-flex; align-items: center; justify-content: center;
      width: var(--action-btn-sz); height: var(--action-btn-sz); padding: 0; border: none;
      background: transparent; color: var(--muted); cursor: pointer;
    }
    .lib-action-btn:hover { color: var(--fg); background: var(--icon-hov); }
    .lib-action-btn.is-active { color: var(--control-active-fg); background: var(--control-active-bg); }
    .lib-action-btn svg { width: 15px; height: 15px; flex-shrink: 0; }
    .lib-action-btn .lib-ai-px { display: none; }
    .lib-action-btn.spinning svg { animation: lib-spin 0.6s linear infinite; }
    @keyframes lib-spin { to { transform: rotate(360deg); } }
    html.macos .lib-topbar { -webkit-app-region: drag; padding-left: 76px; cursor: default; }
    html.macos .lib-topbar button,
    html.macos .lib-topbar a { -webkit-app-region: no-drag; }`

// navHTML returns the complete <nav class="lib-nav">…</nav> HTML block. Home
// is the app's only page now — everything else (Inbox, Agent Chat, Graph,
// Images, Shorts, folders) lives behind its sidebar (see home.html) — so this
// no longer takes an "active page" param: the first nav item doesn't navigate
// anywhere, it opens the same search overlay as the topbar's search button.
func navHTML() string {
	return `  <nav class="lib-nav">
    <button type="button" class="nav-compose-btn" title="New note (Ctrl+N)"
            onclick="window.__vaultrDrawer && void window.__vaultrDrawer.openNewInDrawer('','')">
      ` + svgCompose + `
    </button>
    <button type="button" class="nav-item"
            :class="{ 'is-active': drawerOpen }"
            :title="/Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent) ? 'Reading drawer (⌘E)' : 'Reading drawer (Ctrl+E)'"
            @click="drawerOpen = !drawerOpen">
      ` + svgPanel + `
    </button>
    <button type="button" class="nav-item" title="New short note (Ctrl+.)"
            onclick="window.openShortDialog && window.openShortDialog()">
      ` + svgShort + `
    </button>
    <button type="button" class="nav-item"
            :title="/Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent) ? 'Search (⌘K)' : 'Search (Ctrl+K)'"
            @click="window.dispatchEvent(new CustomEvent('open-search'))">
      ` + svgSearch + `
    </button>
    <div class="nav-spacer"></div>
    <button type="button" class="nav-item" title="Settings" @click="$store.settingsModal.open = true; $el.blur()">
      ` + svgSettings + `
    </button>
  </nav>
  <script>
  document.addEventListener('DOMContentLoaded', function() {
    window.__vaultrHotkeys.register('nav-search', '1', function() {
      if (window.__vaultrAnyModalOpen && window.__vaultrAnyModalOpen()) return;
      window.dispatchEvent(new CustomEvent('open-search'));
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
// every page: just reload. The drawer toggle, short-note trigger, and search
// that used to live here have all moved to the left nav rail (see navHTML)
// alongside the rest of the app's global actions.
//
//   - reloadClick: JS expression for the reload @click (e.g. "refresh()", "window.location.reload()")
//   - reloadTitle: tooltip text (e.g. "Refresh", "Refresh home")
//   - reloadExtraClass: Alpine :class value for the reload button; pass "" for none (graph passes "loading && 'spinning'")
func topbarActionsHTML(reloadClick, reloadTitle, reloadExtraClass string) string {
	reloadClassAttr := ""
	if reloadExtraClass != "" {
		reloadClassAttr = ` :class="` + reloadExtraClass + `"`
	}

	return `    <div class="lib-topbar-actions">
      <button type="button" class="lib-action-btn"` + reloadClassAttr + ` title="` + reloadTitle + `" @click="` + reloadClick + `">` + topbarIconReload + `</button>
    </div>`
}
