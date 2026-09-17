package view

// Token architecture (DESIGN.md):
//   1. Primitives — raw tuples / theme-invariant literals
//   2. Semantic — what components consume (--bg/--fg/…); dark on :root,
//      light via data-theme or prefers-color-scheme; ink/accent RGB re-tints rgba()
//   3. Visual-regime — flat by default; floating layers opt into --shadow-*
//   4. Component-scoped — search, lightbox, etc.
// New state classes: is-* (legacy .active/.open/.sel stay as-is).
const appTokensShared = `
      /* ═══ Layer 1: primitives ═══ */
      --accent-rgb:94,106,210;
      --inverse-canvas:#ffffff; --inverse-ink:#000000;
      /* Brand accent — CTA / focus / links only. */
      --accent:#5e6ad2; --accent-fg:#ffffff;
      /* Selected surface: surface-2 + --fg, not accent fill. */
      --control-active-bg:var(--surface-2); --control-active-fg:var(--fg);
      /* Modal scrim — shared by dialogs/settings/search/drawer. Lightbox has its own. */
      --overlay-bg:rgba(0,0,0,0.25);
      --card-hov:rgba(var(--accent-rgb),0.10);
      --icon-hov:rgba(var(--accent-rgb),0.10);
      --tint-soft:rgba(var(--accent-rgb),0.06);
      --tint-md:rgba(var(--accent-rgb),0.10);
      --tint-strong:rgba(var(--accent-rgb),0.14);
      --input-focus-ring:rgba(var(--accent-rgb),0.16);
      --cm-selection-bg:rgba(var(--accent-rgb),0.16);
      --bq-bd:rgba(var(--accent-rgb),0.55);
      --ul-mk:var(--accent); --ol-mk:var(--accent);
      --code-bg:rgba(var(--ink-rgb),0.055); --code-bd:rgba(var(--ink-rgb),0.10);
      /* --hairline: secondary seam (bg already differs, or inside a bordered card).
         --border/--border-strong: sole separator or floating outer edge. */
      --hairline:rgba(var(--ink-rgb),0.13);
      --nav-fg:rgba(var(--ink-rgb),0.82);
      --th-bg:rgba(var(--ink-rgb),0.04); --tbl-bd:rgba(var(--ink-rgb),0.12);
      --cm-active-line:rgba(var(--ink-rgb),0.04);
      --cnt-bg:rgba(var(--ink-rgb),0.07);
      --scrollbar-thumb:rgba(var(--ink-rgb),0.12); --scrollbar-thumb-hov:rgba(var(--ink-rgb),0.26);
      --nav-act:rgba(var(--ink-rgb),0.1);
      /* Content-identity chips — theme-invariant. */
      --p0:#22d3ee; --p1:#f472b6; --p2:#a78bfa; --p3:#34d399;
      --entity-type-bg:var(--p2);

      /* ═══ Layer 3: visual-regime ═══ */
      --shadow-xs:0 1px 2px; --shadow-sm:0 2px 5px; --shadow-md:0 5px 14px; --shadow-lg:0 14px 34px;
      --accent-focus:#5e69d1;
      --r-xs:4px; --r-sm:6px; --r-md:8px; --r-lg:12px; --r-xl:16px; --r-full:999px;
      /* fast=hover, base=small toggle, slow=panel/scrim. */
      --motion-fast:100ms; --motion-base:160ms; --motion-slow:220ms;
      --bd-w:1px; /* whole px — fractional breaks AA at radius arcs */
      --glass-scrim-filter:blur(3px);

      /* ═══ Layer 4: component-scoped ═══ */
      --srch-ic:var(--muted); --srch-ph:var(--muted-soft); --srch-av:var(--control-active-bg); --srch-backdrop:var(--overlay-bg);
      --sr-dir:var(--cover-dir); --sr-tm:var(--muted); --sr-ic:var(--muted); --sr-em:var(--muted);
      --srch-kbd-fg:var(--fg);
      --srch-panel-bd:var(--border-strong); --srch-row-bd:var(--border);

      /* Lightbox/images — fixed veil over arbitrary photos. */
      --lightbox-overlay-bg:rgba(0,0,0,0.4); --lightbox-viewer-bg:#0c0c0c;
      --img-chip-bg:rgba(17,17,17,0.72);
      --img-select-border:rgba(255,255,255,0.72); --img-select-bg:rgba(0,0,0,0.28);

      --font-ui:"Inter",-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;
      --font-sans:"Inter",-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;
      --font-mono:"JetBrains Mono",ui-monospace,monospace;
      --text-xs:0.75rem; --text-sm:0.8125rem; --text-base:0.875rem; --text-body:1rem;
      --text-title-sm:1rem; --text-title-lg:1.375rem;
      --text-2xs:0.6875rem; --text-3xs:0.625rem;
      --fw-regular:400; --fw-medium:500; --fw-semibold:600;
      --lh-tight:1.2; --lh-snug:1.3; --lh-normal:1.4; --lh-relaxed:1.5;
      --ls-cap:0.08em;
      --space-xxs:4px; --space-xs:8px; --space-sm:12px; --space-md:16px; --space-lg:24px;
      --space-xl:32px; --space-xxl:48px;
      --topbar-h:40px; --action-btn-sz:28px;
      /* Sidebar width — drawer/inbox dock flush to this edge. */
      --home-side-w:248px;
      /* --btn-h primary; --btn-h-xs compact secondary; --btn-h-sm chrome-bar. */
      --btn-h:32px; --btn-h-sm:36px; --btn-h-xs:28px;
      --card-note-h:96px; --card-note-w-max:210px;`

// Default palette (bare :root). Only tokens that differ from light.
const appTokensDark = `
      --ink-rgb:238,240,244;
      --bg:#18191e; --fg:#eef0f4;
      --muted:#8d92a0; --surface-soft:#212229; --surface-2:#292a33;
      /* Extra lift when container already sits on --surface-soft. */
      --control-active-bg-on-soft:#3d3e49;
      /* macOS vibrancy wash — pull native material toward --surface-soft. */
      --sidebar-glass-bg:rgba(33,34,41,0.6); --sidebar-glass-border:rgba(255,255,255,0.1);
      --sidebar-glass-active:rgba(255,255,255,0.08);
      --sidebar-glass-scrollbar-thumb:rgba(255,255,255,0.10); --sidebar-glass-scrollbar-thumb-hov:rgba(255,255,255,0.20);
      --canvas:#ffffff; --muted-soft:#656a78; --body:#d0d6e0;
      --border:#35363f; --border-strong:#44454f;
      --accent-hov:#7b86e8;
      --link:#eef0f4;
      --link-ul:var(--accent); --link-ul-hov:var(--accent-hov);
      --cover-dir:#8d92a0;
      --h1:#eef0f4; --h2:#d0d6e0; --h3:#adb2bc; --h4:#8d92a0;
      --prose-body:#ccd1db;
      /* Editor-only; reader still uses --h4. --lp-h4=--h3 so h4 can't outrank h3. */
      --lp-h4:var(--h3); --lp-h5:var(--h4);
      --prose-strong:#eef0f4; --prose-em:#ccd1db;
      --pre-bg:rgba(var(--ink-rgb),0.055); --pre-bd:rgba(var(--ink-rgb),0.13); --pre-tx:#e9ebf0;
      --code-tx:#ccd1db;
      --bq-tx:rgba(208,214,224,0.85);
      --th-tx:#9a9fa8; --td-tx:#ccd1db; --tbl-bd:rgba(var(--ink-rgb),0.12);
      --cm-md-muted:#8d92a0;
      --cnt-tx:#9a9fa8;
      --s-ok:#34c481; --s-ok-bg:rgba(39,166,68,0.16); --s-ok-bd:rgba(39,166,68,0.4);
      --s-err:#f87171; --s-err-bg:rgba(248,113,113,0.14); --s-err-bd:rgba(248,113,113,0.35);
      --s-warn:#eeae3a; --s-warn-bg:rgba(251,191,36,0.14); --s-warn-bd:rgba(251,191,36,0.36);
      /* Solid status fills need dark ink (these hues are too light for white text). */
      --s-ok-fg:var(--inverse-ink); --s-err-fg:var(--inverse-ink); --s-warn-fg:var(--inverse-ink);
      --shadow-color:rgba(0,0,0,0.55);
      /* Code-fence syntax highlighting — Tokyo Night. */
      --syn-keyword:#bb9af7; --syn-const:#ff9e64; --syn-string:#9ece6a;
      --syn-escape:#b4f9f8; --syn-def:#7aa2f7; --syn-param:#e0af68;
      --syn-type:#2ac3de; --syn-class:#73daca; --syn-builtin:#f7768e;
      --syn-property:#73daca; --syn-comment:#565f89; --syn-invalid:#db4b4b;`

// Light adaptation — cool near-neutral grays.
const appTokensLight = `
      --ink-rgb:24,24,27;
      --bg:#f9f9fb; --fg:#18181b;
      --muted:#6d7080; --surface-soft:#f5f5f7; --surface-2:#e3e3e9;
      --control-active-bg-on-soft:var(--surface-2);
      /* Brighter canvas needs a more visible thumb than the shared (dark-tuned) value. */
      --scrollbar-thumb:rgba(var(--ink-rgb),0.16); --scrollbar-thumb-hov:rgba(var(--ink-rgb),0.30);
      /* Vibrancy reads dull alone — lean lighter while staying translucent. */
      --sidebar-glass-bg:rgba(250,250,251,0.72); --sidebar-glass-border:rgba(0,0,0,0.08);
      --sidebar-glass-active:rgba(0,0,0,0.06);
      /* Over vibrancy, flat --scrollbar-thumb reads too dark — keep lower. */
      --sidebar-glass-scrollbar-thumb:rgba(0,0,0,0.11); --sidebar-glass-scrollbar-thumb-hov:rgba(0,0,0,0.20);
      --canvas:#ffffff; --muted-soft:#9b9eac; --body:#3f4147;
      --border:#d2d2d9; --border-strong:#bcbcc5;
      --accent-hov:#4c56c8;
      --link:#111111;
      --link-ul:var(--accent); --link-ul-hov:var(--accent-hov);
      --cover-dir:#6b7280;
      --h1:#111111; --h2:#1f2937; --h3:#374151; --h4:#6b7280;
      --prose-body:#374151;
      --lp-h4:var(--h3); --lp-h5:var(--h4);
      --prose-strong:#111111; --prose-em:#374151;
      --pre-bg:rgba(var(--ink-rgb),0.055); --pre-bd:rgba(var(--ink-rgb),0.13); --pre-tx:#111111;
      --code-tx:#374151;
      --bq-tx:rgba(55,65,81,0.88);
      --th-tx:#9ca3af; --td-tx:#374151; --tbl-bd:rgba(var(--ink-rgb),0.12);
      --cm-md-muted:#9ca3af;
      --cnt-tx:#6b7280;
      --s-ok:#059669; --s-ok-bg:rgba(16,185,129,0.08); --s-ok-bd:rgba(16,185,129,0.35);
      --s-err:#dc2626; --s-err-bg:rgba(239,68,68,0.06); --s-err-bd:rgba(239,68,68,0.25);
      --s-warn:#d97706; --s-warn-bg:rgba(217,119,6,0.07); --s-warn-bd:rgba(217,119,6,0.26);
      --s-ok-fg:var(--canvas); --s-err-fg:var(--canvas); --s-warn-fg:var(--canvas);
      --shadow-color:rgba(15,15,25,0.10);
      /* Code-fence syntax highlighting — Tokyo Night Light. */
      --syn-keyword:#5a4a78; --syn-const:#965027; --syn-string:#385f0d;
      --syn-escape:#0f4b6e; --syn-def:#34548a; --syn-param:#8f5e15;
      --syn-type:#166775; --syn-class:#33635c; --syn-builtin:#8c4351;
      --syn-property:#33635c; --syn-comment:#848cb1; --syn-invalid:#c53b53;`

// Drop into a <style> tag. Dark on :root; light via data-theme or OS preference.
const appTokensCSS = `    :root {` + appTokensShared + appTokensDark + `
    }
    :root[data-theme="light"] {` + appTokensLight + `
    }
    @media (prefers-color-scheme: light) {
      :root:not([data-theme="dark"]) {` + appTokensLight + `
      }
    }
    *:focus-visible {
      outline: 2px solid color-mix(in srgb, var(--accent-focus) 50%, transparent);
      outline-offset: 1px;
      box-shadow: none;
    }
    /* Floor — component :active still wins by specificity. */
    button:not(:disabled):active {
      opacity: 0.85;
    }`
