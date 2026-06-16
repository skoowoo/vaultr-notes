package view

// Typography scale — aligned to Design.md (base 16px, theme-independent):
//
// Display (Cal Sans → Inter 600 substitute, negative letter-spacing):
//
//	display-xl  4rem     64px  lh 1.05  ls -0.03125em  homepage h1
//	display-lg  3rem     48px  lh 1.1   ls -0.031em    section heads
//	display-md  2.25rem  36px  lh 1.15  ls -0.028em    sub-section heads
//	display-sm  1.75rem  28px  lh 1.2   ls -0.02em     CTA-band heads, cover h1
//
// Title (Inter 600, light or no letter-spacing):
//
//	title-lg  1.375rem  22px  lh 1.3  ls -0.014em  pricing plan names, modal titles
//	title-md  1.125rem  18px  lh 1.4  ls 0          feature card titles
//	title-sm  1rem      16px  lh 1.4  ls 0          small card titles, list labels
//
// Body / UI (Inter 400–600, 0 letter-spacing):
//
//	body-md   1rem      16px  lh 1.75  wt 400  running-text (prose, reading)
//	body-sm   0.875rem  14px  lh 1.5   wt 400  footer body, secondary content
//	caption   0.8125rem 13px  lh 1.4   wt 500  badge labels, timestamps, captions
//	code      0.875rem  14px  lh 1.5   wt 400  JetBrains Mono
//	button    0.875rem  14px  lh 1.0   wt 600  button labels
//	nav-link  0.875rem  14px  lh 1.4   wt 500  top-nav menu items
//
// Minimum: xs = 0.75rem (12px) — section labels, status badges, micro-labels only.
// Pixel-art decorations (shared_pixel.go, shared_icons_pixel.go) are exempt.
//
// Color tokens below change with theme and must use CSS custom properties.

// appTokensDark is the :root CSS custom-properties for the dark theme.
// Based on the Cal.com-inspired neutral system: near-black surfaces, light text,
// coral accent (#cc785c), and badge pastels for folder/pin tints.
const appTokensDark = `
      --bg:#0f0f0f; --fg:#f4f4f5; --hr:rgba(244,244,245,0.07);
      --muted:#71717a; --card-bg:#1a1a1a; --surface-soft:#141414;
      --card-bd:rgba(244,244,245,0.09); --card-hov:#252525;
      --canvas:#ffffff; --hairline:#e5e7eb; --muted-soft:#52525b; --body:#d4d4d8; --body-strong:#f4f4f5;
      --nav-bg:#101010; --nav-act:rgba(244,244,245,0.07);
      --code-bg:rgba(244,244,245,0.055); --code-bd:rgba(244,244,245,0.08);
      --accent:#cc785c; --accent-rgb:204,120,92; --accent-hov:#b3694a;
      --link:var(--accent); --tab-act:var(--link);
      /* reading / noteSharedCSS (.prose — search preview, agent chat, shorts) */
      --cover-dir:#71717a;
      --h1:#f4f4f5; --h2:#e4e4e7; --h3:#d4d4d8; --h4:#71717a;
      --prose-body:#d4d4d8; --prose-lead:#e4e4e7;
      --prose-strong:#f4f4f5; --prose-em:#d4d4d8;
      --link-hov:var(--accent-hov);
      --link-ul:rgba(var(--accent-rgb),0.28); --link-ul-hov:rgba(var(--accent-rgb),0.50);
      --pre-bg:#0a0a0a; --pre-bd:rgba(244,244,245,0.06); --pre-tx:#d4d4d8;
      --code-tx:#d4d4d8;
      --bq-bd:rgba(var(--accent-rgb),0.42); --bq-tx:rgba(212,212,216,0.88);
      --h2-rule:rgba(244,244,245,0.07);
      --ul-mk:rgba(244,244,245,0.20); --ol-mk:rgba(244,244,245,0.30);
      --th-bg:rgba(244,244,245,0.03); --th-tx:#71717a; --td-tx:#d4d4d8; --tbl-bd:rgba(244,244,245,0.08);
      --cm-md-muted:#52525b; --cm-active-line:rgba(244,244,245,0.038); --cm-selection-bg:rgba(var(--accent-rgb),0.13);
      --p0:#3b82f6; --p1:#ec4899; --p2:#8b5cf6; --p2-rgb:139,92,246; --p3:#34d399;
      /* overlay / modal backdrop */
      --overlay-bg:rgba(0,0,0,0.54);
      /* ── Typography (theme-independent) ── */
      --font-sans:"Inter",-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;
      --font-mono:"JetBrains Mono",ui-monospace,monospace;
      --text-xs:0.75rem; --text-sm:0.8125rem; --text-base:0.875rem; --text-body:1rem;
      --text-title-sm:1rem; --text-title-md:1.125rem; --text-title-lg:1.375rem;
      --text-display-sm:1.75rem; --text-display-md:2.25rem;
      --fw-regular:400; --fw-medium:500; --fw-semibold:600;
      --lh-tight:1.2; --lh-snug:1.3; --lh-normal:1.4; --lh-relaxed:1.5; --lh-prose:1.75;
      --ls-display-sm:-0.02em; --ls-display-md:-0.028em;
      --ls-title-lg:-0.014em; --ls-cap:0.08em;
      /* ── Spacing ── */
      --space-xxs:4px; --space-xs:8px; --space-sm:12px; --space-md:16px;
      --space-lg:24px; --space-xl:32px; --space-xxl:48px; --space-section:96px;
      /* ── Radius ── */
      --radius-xs:4px; --radius-sm:6px; --radius-md:8px; --radius-lg:12px;
      --radius-xl:16px; --radius-pill:9999px;
      /* ── Component sizes ── */
      --topbar-h:40px; --nav-w:60px; --nav-item-sz:40px; --action-btn-sz:28px;
      --btn-h:40px; --btn-h-sm:36px; --btn-h-xs:28px;
      --card-note-h:96px; --card-note-w-max:210px;
      /* ── Primary button ── */
      --btn-primary-bg:#111111; --btn-primary-fg:#ffffff;
      --btn-primary-hover:#1a1a1a; --btn-primary-active:#242424;
      /* ── Semantic state colors ── */
      --text-2xs:0.6875rem;
      --s-ok:#10b981; --s-ok-bg:rgba(52,211,153,0.12); --s-ok-bd:rgba(52,211,153,0.42);
      --s-err:#ef4444; --s-err-bg:rgba(239,68,68,0.08); --s-err-bd:rgba(239,68,68,0.32);
      --s-warn:#f59e0b; --s-warn-bg:rgba(245,158,11,0.08); --s-warn-bd:rgba(245,158,11,0.28);
      --input-focus-ring:rgba(var(--accent-rgb),0.10);
      --drawer-overlay-bg:rgba(0,0,0,0.42);
      --scrollbar-thumb:rgba(244,244,245,0.15); --scrollbar-thumb-hov:rgba(244,244,245,0.34);
      /* ── Search overlay ── */
      --srch-bg:var(--card-bg); --srch-bg-glass:rgba(15,15,15,0.96);
      --srch-ic:var(--muted); --srch-ph:var(--muted-soft); --srch-av:var(--card-hov); --srch-backdrop:var(--overlay-bg);
      --sr-dir:var(--muted-soft); --sr-tm:var(--muted-soft); --sr-ic:var(--muted-soft); --sr-em:var(--muted-soft);
      --srch-btn-fg:var(--muted-soft); --srch-btn-fh:var(--muted); --srch-btn-bh:var(--muted-soft); --srch-btn-bgh:var(--card-hov);
      --srch-kbd-fg:var(--muted-soft);
      --srch-panel-bd:rgba(244,244,245,0.22); --srch-row-bd:rgba(244,244,245,0.10); --srch-kbd-bd:rgba(244,244,245,0.16);
      --srch-panel-shadow:0 0 0 1px rgba(244,244,245,0.06),0 16px 48px rgba(0,0,0,0.52),0 4px 16px rgba(0,0,0,0.28);`

// appTokensLight is the light-mode counterpart.
// White canvas (#ffffff), black ink (#111111), coral accent (#cc785c),
// light-gray cards (#f5f5f5), always-dark nav sidebar (#101010).
const appTokensLight = `
      --bg:#ffffff; --fg:#111111; --hr:#e5e7eb;
      --muted:#6b7280; --card-bg:#f5f5f5; --surface-soft:#f8f9fa;
      --card-bd:#e5e7eb; --card-hov:#ebebeb;
      --canvas:#ffffff; --hairline:#e5e7eb; --muted-soft:#898989; --body:#374151; --body-strong:#111111;
      --nav-bg:#f5f5f5; --nav-act:rgba(17,17,17,0.06);
      --code-bg:rgba(17,17,17,0.055); --code-bd:rgba(17,17,17,0.10);
      --accent:#cc785c; --accent-rgb:204,120,92; --accent-hov:#b3694a;
      --link:var(--accent); --tab-act:var(--link);
      --cover-dir:#9ca3af;
      --h1:#111111; --h2:#1f2937; --h3:#374151; --h4:#6b7280;
      --prose-body:#374151; --prose-lead:#111111;
      --prose-strong:#111111; --prose-em:#374151;
      --link-hov:var(--accent-hov);
      --link-ul:rgba(var(--accent-rgb),0.28); --link-ul-hov:rgba(var(--accent-rgb),0.50);
      --pre-bg:#f5f5f5; --pre-bd:#e5e7eb; --pre-tx:#111111;
      --code-tx:#374151;
      --bq-bd:rgba(var(--accent-rgb),0.42); --bq-tx:rgba(55,65,81,0.88);
      --h2-rule:#e5e7eb;
      --ul-mk:rgba(17,17,17,0.22); --ol-mk:rgba(17,17,17,0.32);
      --th-bg:rgba(17,17,17,0.04); --th-tx:#9ca3af; --td-tx:#374151; --tbl-bd:#e5e7eb;
      --cm-md-muted:#9ca3af; --cm-active-line:rgba(17,17,17,0.04); --cm-selection-bg:rgba(var(--accent-rgb),0.12);
      --p0:#3b82f6; --p1:#ec4899; --p2:#8b5cf6; --p2-rgb:139,92,246; --p3:#34d399;
      /* overlay / modal backdrop */
      --overlay-bg:rgba(17,17,17,0.38);
      /* ── Semantic state colors (light overrides) ── */
      --s-ok:#059669; --s-ok-bg:rgba(16,185,129,0.08); --s-ok-bd:rgba(16,185,129,0.35);
      --s-err:#dc2626; --s-err-bg:rgba(239,68,68,0.06); --s-err-bd:rgba(239,68,68,0.25);
      --s-warn:#d97706; --s-warn-bg:rgba(217,119,6,0.07); --s-warn-bd:rgba(217,119,6,0.26);
      --input-focus-ring:rgba(var(--accent-rgb),0.08);
      --drawer-overlay-bg:rgba(0,0,0,0.22);
      --scrollbar-thumb:rgba(17,17,17,0.12); --scrollbar-thumb-hov:rgba(17,17,17,0.26);
      /* ── Search overlay ── */
      --srch-bg:var(--bg); --srch-bg-glass:rgba(255,255,255,0.97);
      --srch-ic:var(--cover-dir); --srch-ph:var(--cover-dir); --srch-av:var(--card-bg); --srch-backdrop:var(--overlay-bg);
      --sr-dir:var(--cover-dir); --sr-tm:var(--cover-dir); --sr-ic:var(--cover-dir); --sr-em:var(--cover-dir);
      --srch-btn-fg:var(--muted); --srch-btn-fh:var(--fg); --srch-btn-bh:var(--hr); --srch-btn-bgh:var(--card-bg);
      --srch-kbd-fg:var(--cover-dir);
      --srch-panel-bd:rgba(17,17,17,0.20); --srch-row-bd:rgba(17,17,17,0.10); --srch-kbd-bd:#d1d5db;
      --srch-panel-shadow:0 0 0 1px rgba(17,17,17,0.04),0 8px 32px rgba(0,0,0,0.14),0 2px 8px rgba(0,0,0,0.08);`

// appTokensCSS is the full dark+light CSS block — drop it inside a <style> tag.
const appTokensCSS = `    :root {` + appTokensDark + `
    }
    html[data-theme="light"] {` + appTokensLight + `
    }`
