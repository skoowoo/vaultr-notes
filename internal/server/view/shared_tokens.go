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
// Pixel-art decorations (shared_pixel.go, shared_icons_neo.go) are exempt.
//
// Color tokens below change with theme and must use CSS custom properties.

// appTokensNeo is the sole theme token set. All tokens live under
// html[data-theme="neo"]; the bootstrap script hardcodes that attribute.
// To add a future theme: add a new token block under a new data-theme value.
const appTokensNeo = `
      /* ── Neo shadow tokens ── */
      --px-shadow:rgba(0,0,0,1); --px-d0:1px 1px 0; --px-d1:2px 2px 0; --px-d2:3px 3px 0; --px-d3:4px 4px 0;
      /* ── Neo UI font (Space Grotesk — UI chrome only, content stays Inter) ── */
      --font-ui:"Space Grotesk",-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;
      /* ── Base surface ── */
      --bg:#ffffff; --fg:#111111;
      --muted:#4b5563; --card-bg:#f5f5f5; --surface-soft:#f8f9fa;
      --card-hov:rgba(250,204,21,0.10);
      --canvas:#ffffff; --hairline:#e5e7eb; --muted-soft:#898989; --body:#374151;
      /* ── Borders & dividers — opaque black (Neo-Brutalism) ── */
      --hr:#333333;
      --card-bd:#000000;
      /* ── Code surfaces ── */
      --code-bg:rgba(17,17,17,0.055); --code-bd:rgba(17,17,17,0.10);
      /* ── Navigation — brand yellow ── */
      --nav-bg:#facc15; --nav-act:rgba(0,0,0,0.1);
      /* ── Brand accent — yellow ── */
      --accent:#facc15; --accent-rgb:250,204,21; --accent-hov:#eab308;
      /* ── Link color — prose hyperlinks only; do not use in UI chrome ── */
      --link:#ec4899; --link-hov:#db2777; --link-rgb:236,72,153;
      --link-ul:rgba(236,72,153,0.32); --link-ul-hov:rgba(236,72,153,0.55);
      /* ── UI primary — interactive states, active elements, buttons (not prose links) ── */
      --ui-accent:#facc15; --ui-accent-hov:#eab308;
      /* ── Prose headings ── */
      --cover-dir:#6b7280;
      --h1:#111111; --h2:#1f2937; --h3:#374151; --h4:#6b7280;
      /* ── Prose body ── */
      --prose-body:#374151;
      --prose-strong:#111111; --prose-em:#374151;
      --pre-bg:#f5f5f5; --pre-bd:#e5e7eb; --pre-tx:#111111;
      --code-tx:#374151;
      --bq-bd:rgba(250,204,21,0.55); --bq-tx:rgba(55,65,81,0.88);
      --ul-mk:rgba(17,17,17,0.22); --ol-mk:rgba(17,17,17,0.32);
      --th-bg:rgba(17,17,17,0.04); --th-tx:#9ca3af; --td-tx:#374151; --tbl-bd:#e5e7eb;
      --cm-md-muted:#9ca3af; --cm-active-line:rgba(17,17,17,0.04); --cm-selection-bg:rgba(250,204,21,0.12);
      --p0:#22d3ee; --p1:#f472b6; --p2:#a78bfa; --p3:#34d399;
      /* ── Overlays ── */
      --overlay-bg:rgba(17,17,17,0.38);
      /* ── Typography (theme-independent) ── */
      --font-sans:"Inter",-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;
      --font-mono:"JetBrains Mono",ui-monospace,monospace;
      --text-xs:0.75rem; --text-sm:0.8125rem; --text-base:0.875rem; --text-body:1rem;
      --text-title-sm:1rem; --text-title-lg:1.375rem;
      --text-2xs:0.6875rem;
      --fw-regular:400; --fw-medium:500; --fw-semibold:600;
      --lh-tight:1.2; --lh-snug:1.3; --lh-normal:1.4; --lh-relaxed:1.5;
      --ls-cap:0.08em;
      /* ── Spacing ── */
      --space-xxs:4px; --space-xs:8px; --space-sm:12px; --space-md:16px; --space-lg:24px;
      /* ── Radius (zeroed globally by neo.css * rule) ── */
      --radius-xs:4px; --radius-sm:6px; --radius-md:8px; --radius-lg:12px; --radius-pill:9999px;
      /* ── Component sizes ── */
      --topbar-h:40px; --nav-w:60px; --nav-item-sz:40px; --action-btn-sz:28px;
      --btn-h:40px; --btn-h-sm:36px; --btn-h-xs:28px;
      --card-note-h:96px; --card-note-w-max:210px;
      --cnt-bg:rgba(17,17,17,0.07); --cnt-tx:#6b7280;
      /* ── Primary button — brand yellow ── */
      --btn-primary-bg:#facc15; --btn-primary-fg:#111111;
      --btn-primary-hover:#eab308; --btn-primary-active:#ca8a04;
      /* ── Semantic state colors ── */
      --s-ok:#059669; --s-ok-bg:rgba(16,185,129,0.08); --s-ok-bd:rgba(16,185,129,0.35);
      --s-err:#dc2626; --s-err-bg:rgba(239,68,68,0.06); --s-err-bd:rgba(239,68,68,0.25);
      --s-warn:#d97706; --s-warn-bg:rgba(217,119,6,0.07); --s-warn-bd:rgba(217,119,6,0.26);
      --input-focus-ring:rgba(250,204,21,0.08);
      --drawer-overlay-bg:rgba(0,0,0,0.22);
      --scrollbar-thumb:rgba(17,17,17,0.12); --scrollbar-thumb-hov:rgba(17,17,17,0.26);
      /* ── Interaction tints — brand yellow on white bg ── */
      --icon-hov:rgba(250,204,21,0.10);
      --tint-soft:rgba(250,204,21,0.06);
      --tint-md:rgba(250,204,21,0.10);
      --tint-strong:rgba(250,204,21,0.14);
      /* ── Segmented / toggle control active state ── */
      --seg-act-bg:#000000; --seg-act-fg:#ffffff;
      /* ── Knowledge graph entity type chip ── */
      --entity-type-bg:var(--p2);
      /* ── Search overlay ── */
      --srch-bg:var(--bg);
      --srch-ic:var(--cover-dir); --srch-ph:var(--cover-dir); --srch-av:var(--accent); --srch-backdrop:var(--overlay-bg);
      --sr-dir:#6b7280; --sr-tm:#6b7280; --sr-ic:#6b7280; --sr-em:#6b7280;
      --srch-kbd-fg:var(--fg);
      --srch-panel-bd:#000000; --srch-row-bd:#333333; --srch-kbd-bd:#000000;
      --srch-panel-shadow:4px 4px 0 rgba(0,0,0,1);
      --srch-row-bg:var(--bg);
      --srch-hover-bar:rgba(0,0,0,0.22); --srch-hover-bg:rgba(0,0,0,0.08);`

// appTokensCSS is the single-theme CSS block — drop it inside a <style> tag.
const appTokensCSS = `    html[data-theme="neo"] {` + appTokensNeo + `
    }`
