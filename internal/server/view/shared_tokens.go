package view

// Typography scale — aligned to Design.md (base 16px, theme-independent):
//
// Display (Inter 600, negative letter-spacing — see DESIGN-linear.app.md):
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

// ── Token architecture ──────────────────────────────────────────────────────
//
// Visual language: Linear's marketing-canvas system (DESIGN-linear.app.md) —
// near-black surface ladder, lavender-blue (#5e6ad2) as the single chromatic
// accent, hairline borders instead of heavy shadows, Inter as one continuous
// display+body voice. Dark is the shipped default (matches Linear, which
// documents no light mode); light is this app's own from-scratch adaptation
// of the same hue/radius/spacing system, kept alive because vaultr (unlike a
// marketing page) needs a comfortable reading surface in bright rooms.
//
// Four layers, in the order they appear below:
//
//  1. Primitives — raw color tuples with no meaning of their own, and the two
//     handful of tokens that are genuinely identical in both themes (the
//     accent hue, its on-color text, the true-white/true-black inverse pair,
//     modal scrim). Never referenced directly by component CSS; only by
//     Layer 2 tokens.
//
//  2. Semantic tokens — what components actually consume (--bg, --fg,
//     --border, --accent, etc.). Dark values live on bare :root (the
//     default); :root[data-theme="light"] and the mirroring
//     @media (prefers-color-scheme: light) block (guarded by
//     :not([data-theme="dark"])) override just the tokens that differ for
//     light — most --ink-rgb/--accent-rgb-derived rgba() tokens don't need
//     restating: custom-property var() references resolve against whatever
//     is cascaded on the element at use time, so overriding --ink-rgb alone
//     re-tints every rgba(var(--ink-rgb),…) token automatically. Component
//     CSS should never need to know which theme is active.
//
//  3. Visual-regime tokens — depth. Per DESIGN-linear.app.md's Elevation &
//     Depth table, none of Linear's 4 documented levels use a box-shadow —
//     depth is carried entirely by the surface ladder (Layer 2's --bg/
//     --surface-soft steps) and hairline borders, with a 2px focus ring as
//     the one exception (level 4). --shadow-xs/sm/md/lg keep their offset/
//     blur shapes (component CSS still pairs them with --shadow-color) but
//     --shadow-color itself is transparent in both themes, so every existing
//     `box-shadow: var(--shadow-*) var(--shadow-color)` call renders
//     invisible without having to touch each of those call sites — the "5
//     lines" a future shadow-language swap would touch are these. --r-*
//     (radius) and --bd-w (border width) are NOT regime-specific — they're
//     generic structural scales any visual style would keep using; the
//     values here are Linear's rounded.* scale.
//
//  4. Component-scoped tokens — narrow, single-purpose tokens for one part of
//     the UI (search overlay, count badges, etc.). These may reference
//     Layer 2/3 tokens but rarely need touching when the theme changes.
const appTokensShared = `
      /* ═══ Layer 1: primitives — raw tuples, only consumed via rgba(var(...)) below,
         plus the small set of literal tokens that don't change between themes ═══ */
      --accent-rgb:94,106,210;
      --inverse-canvas:#ffffff; --inverse-ink:#000000;
      /* ── Brand accent — Linear lavender-blue. Scarce by design: brand mark,
         primary CTA, focus ring, link emphasis. Same hue in both themes;
         only --accent-hov (Layer 2) shifts direction per theme. ── */
      --accent:#5e6ad2; --accent-fg:#ffffff;
      /* ── Selected / active surface — Linear's "pricing-tab-selected" /
         "status-badge" pattern (surface-2 lift + full ink text), not an
         accent fill: DESIGN-linear.app.md reserves lavender for brand mark,
         primary CTA, focus ring, and link emphasis only ("Don't use lavender
         as a section background or card fill"). This is the ONE treatment
         for every "this tab/segment/row/toggle is currently selected" state
         app-wide — sidebar items, drawer tabs, segmented controls, search
         result rows, icon-toggle buttons. Declared here (not per-theme)
         because it's the same formula in both: step up to --surface-2,
         text to full --fg. Primary-action buttons (compose, publish, save,
         send) use --accent/--accent-fg directly instead — never this pair. */
      --control-active-bg:var(--surface-2); --control-active-fg:var(--fg);
      /* ── Modal / drawer scrim — always a true-black veil, independent of
         theme (darkening the backdrop reads the same whether the surface
         above it is paper-white or near-black). ── */
      --overlay-bg:rgba(0,0,0,0.55); --drawer-overlay-bg:rgba(0,0,0,0.4);
      /* ── Interaction tints (accent-derived; --accent-rgb is theme-invariant
         so these need no per-theme restating) ── */
      --card-hov:rgba(var(--accent-rgb),0.10);
      --icon-hov:rgba(var(--accent-rgb),0.10);
      --tint-soft:rgba(var(--accent-rgb),0.06);
      --tint-md:rgba(var(--accent-rgb),0.10);
      --tint-strong:rgba(var(--accent-rgb),0.14);
      --input-focus-ring:rgba(var(--accent-rgb),0.16);
      --cm-selection-bg:rgba(var(--accent-rgb),0.16);
      --bq-bd:rgba(var(--accent-rgb),0.55);
      /* List/task-list prefix marks (bullet dot, ordered numeral, checkbox
         border+fill) — accent, not ink-derived: these are the one spot in
         prose where the accent is allowed to recur per-item, since it's
         marking structure (list-ness) rather than decorating content. */
      --ul-mk:var(--accent); --ol-mk:var(--accent);
      /* ── Code surfaces (ink-derived; --ink-rgb flips per theme in Layer 2,
         which re-tints all of these automatically) ── */
      --code-bg:rgba(var(--ink-rgb),0.055); --code-bd:rgba(var(--ink-rgb),0.10);
      --hairline:rgba(var(--ink-rgb),0.13);
      --th-bg:rgba(var(--ink-rgb),0.04); --tbl-bd:rgba(var(--ink-rgb),0.12);
      --cm-active-line:rgba(var(--ink-rgb),0.04);
      --cnt-bg:rgba(var(--ink-rgb),0.07);
      --scrollbar-thumb:rgba(var(--ink-rgb),0.12); --scrollbar-thumb-hov:rgba(var(--ink-rgb),0.26);
      --nav-act:rgba(var(--ink-rgb),0.1);
      /* ── Content-identity palette (entity chips, mate avatar fallback) —
         theme-invariant: these mark "which thing", not "what mode" ── */
      --p0:#22d3ee; --p1:#f472b6; --p2:#a78bfa; --p3:#34d399;
      --entity-type-bg:var(--p2);

      /* ═══ Layer 3: visual-regime tokens — box-shadow is switched off (see
         Layer-3 note above): --shadow-color resolves to transparent in both
         themes, so these offset/blur shapes render invisible wherever
         component CSS still pairs them in (box-shadow: var(--shadow-*)
         var(--shadow-color)). Kept only so that pairing stays valid CSS. */
      --shadow-xs:0 1px 2px; --shadow-sm:0 2px 5px; --shadow-md:0 5px 14px; --shadow-lg:0 14px 34px;
      /* Linear's one exception to "no shadow": the 2px focus ring at 50%
         opacity (Elevation level 4). --accent-focus is the spec's literal
         primary-focus value; used by *:focus-visible below. */
      --accent-focus:#5e69d1;
      /* Corner radius scale, by visual weight — Linear's rounded.* scale:
         xs (chips/badges) → sm (inline tags) → md (buttons/inputs, the
         default) → lg (cards) → xl (modal/panel surfaces, drawer, image
         tiles — Linear's "product-screenshot-card" treatment) → full
         (pills/dots/avatars). Generic, not regime-specific. */
      --r-xs:4px; --r-sm:6px; --r-md:8px; --r-lg:12px; --r-xl:16px; --r-full:999px;
      /* Structural border weight (was a hardcoded 2px everywhere). Kept as a
         whole pixel — a fractional width here visibly breaks the anti-aliasing
         where a straight border segment meets a border-radius arc (soft seam
         right at the tangent point). Generic, not regime-specific. */
      --bd-w:1px;
      /* Glass material — the floating-card treatment (search, confirm/info
         dialogs, settings modal, image lightbox): a translucent surface
         over a blurred, dimmed scrim. Reserved for centered cards with
         visible backdrop margin on all sides; NOT used on edge-docked
         chrome (drawer, inbox sheet) — those stay opaque per Layer 2's
         --bg/--surface-soft, since they occupy most/all of the viewport
         (no backdrop left to blur through) and host long-session reading/
         editing where translucency would hurt legibility. --glass-bg reuses
         --surface-soft's own hue via color-mix() rather than a new literal,
         so it can never drift out of sync with the opaque surface ladder;
         --glass-filter/--glass-scrim-filter are shared verbatim by every
         glass card so no two dialogs' blur amounts can quietly diverge. */
      --glass-bg:color-mix(in srgb, var(--surface-soft) 82%, transparent);
      --glass-filter:blur(24px) saturate(1.5);
      --glass-scrim-filter:blur(8px);

      /* ═══ Layer 4: component-scoped tokens ═══ */
      /* ── Search overlay — icon/placeholder/timestamp/row-icon/empty-state
         all point at --muted (or --muted-soft for the placeholder), matching
         the same roles everywhere else app-wide (nav-item icons, .home-
         inbox-card-time, .lb-note-card-icon, .img-empty, .vaultr-sr-input's
         own placeholder). --sr-dir is the one exception: it's a genuine
         directory-label, same concept as note_frontmatter.css/fragment.go's
         .cover-dir (the preview pane literally reuses that class), so it
         stays tied to --cover-dir rather than the generic muted tone. ── */
      /* Search's panel uses the shared --glass-bg/--glass-filter (Layer 3)
         for its floating-card material — one material for the whole panel,
         input row included, divided only by hairline borders, instead of
         the row carrying its own opaque --bg fill. */
      --srch-ic:var(--muted); --srch-ph:var(--muted-soft); --srch-av:var(--control-active-bg); --srch-backdrop:var(--overlay-bg);
      --sr-dir:var(--cover-dir); --sr-tm:var(--muted); --sr-ic:var(--muted); --sr-em:var(--muted);
      --srch-kbd-fg:var(--fg);
      --srch-panel-bd:var(--border-strong); --srch-row-bd:var(--border); --srch-kbd-bd:var(--border-strong);

      /* ── Images grid & lightbox — chrome drawn on top of arbitrary photo
         content, not app surfaces, so (like the modal scrim above) these
         stay a fixed black/white veil independent of theme. ── */
      --lightbox-overlay-bg:rgba(0,0,0,0.78); --lightbox-viewer-bg:#0c0c0c;
      --img-chip-bg:rgba(17,17,17,0.72);
      --img-select-border:rgba(255,255,255,0.72); --img-select-bg:rgba(0,0,0,0.28);

      /* ── UI font — Inter carries both display and body (Linear treats them
         as one continuous voice; the family change is silent). Content
         prose stays on --font-sans, same family. ── */
      --font-ui:"Inter",-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;
      /* ── Typography (theme-independent) ── */
      --font-sans:"Inter",-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;
      --font-mono:"JetBrains Mono",ui-monospace,monospace;
      --text-xs:0.75rem; --text-sm:0.8125rem; --text-base:0.875rem; --text-body:1rem;
      --text-title-sm:1rem; --text-title-lg:1.375rem;
      --text-2xs:0.6875rem; --text-3xs:0.625rem;
      --fw-regular:400; --fw-medium:500; --fw-semibold:600;
      --lh-tight:1.2; --lh-snug:1.3; --lh-normal:1.4; --lh-relaxed:1.5;
      --ls-cap:0.08em;
      /* ── Spacing — DESIGN-linear.app.md's full scale (4px base unit).
         --space-xl/xxl exist for completeness (dialog/panel-level padding);
         --space-section (96px) isn't declared since nothing in this app's
         chrome-density UI operates at marketing-page section scale. ── */
      --space-xxs:4px; --space-xs:8px; --space-sm:12px; --space-md:16px; --space-lg:24px;
      --space-xl:32px; --space-xxl:48px;
      /* ── Component sizes ── */
      --topbar-h:40px; --nav-w:52px; --nav-item-sz:28px; --action-btn-sz:28px;
      /* --btn-h is the standard text-label button/input height — 32px, which
         is what DESIGN-linear.app.md's own button-primary spec (8px 14px
         padding around 14px/1.2 button type) resolves to, and matches the
         app's form inputs so a button never mismatches the input beside it.
         --btn-h-xs (28px) is the one compact tier, shared with
         --action-btn-sz for icon-only buttons — used for secondary inline
         row actions (card "Delete"/"Edit", grid toolbar toggles), never for
         a primary CTA. --btn-h-sm is unrelated to buttons — it's a chrome
         bar height (search overlay footers). */
      --btn-h:32px; --btn-h-sm:36px; --btn-h-xs:28px;
      --card-note-h:96px; --card-note-w-max:210px;`

// appTokensDark is the default palette (bare :root) — Linear's near-black
// canvas (#010102) with light-gray ink and the lavender accent's hover state
// tuned to lighten (per DESIGN-linear.app.md's --primary-hover). Only the
// tokens that genuinely differ from light live here; everything derived from
// --ink-rgb/--accent-rgb via rgba() in appTokensShared re-tints on its own.
const appTokensDark = `
      --ink-rgb:247,248,248;
      --bg:#010102; --fg:#f7f8f8;
      --muted:#8a8f98; --surface-soft:#0f1011; --surface-2:#141516;
      --canvas:#ffffff; --muted-soft:#62666d; --body:#d0d6e0;
      --border:#23252a; --border-strong:#34343a;
      --accent-hov:#828fff;
      --link:#f7f8f8;
      --link-ul:var(--accent); --link-ul-hov:var(--accent-hov);
      --cover-dir:#8a8f98;
      --h1:#f7f8f8; --h2:#d0d6e0; --h3:#adb2bc; --h4:#8a8f98;
      --prose-body:#ccd1db;
      --prose-strong:#f7f8f8; --prose-em:#ccd1db;
      --pre-bg:rgba(var(--ink-rgb),0.055); --pre-bd:rgba(var(--ink-rgb),0.13); --pre-tx:#f2f3f4;
      --code-tx:#ccd1db;
      --bq-tx:rgba(208,214,224,0.85);
      --th-tx:#9a9fa8; --td-tx:#ccd1db; --tbl-bd:rgba(var(--ink-rgb),0.12);
      --cm-md-muted:#8a8f98;
      --cnt-tx:#9a9fa8;
      --s-ok:#3ddc8f; --s-ok-bg:rgba(39,166,68,0.16); --s-ok-bd:rgba(39,166,68,0.4);
      --s-err:#f87171; --s-err-bg:rgba(248,113,113,0.14); --s-err-bd:rgba(248,113,113,0.35);
      --s-warn:#fbbf24; --s-warn-bg:rgba(251,191,36,0.14); --s-warn-bd:rgba(251,191,36,0.36);
      /* Text color to pair with --s-ok/-err/-warn when THEY are the fill
         (a handful of solid status buttons/pills) rather than the text —
         these semantic colors are tuned bright for legibility as running
         text on the dark canvas, which makes them too light for white text
         on top; dark ink reads correctly there instead. */
      --s-ok-fg:var(--inverse-ink); --s-err-fg:var(--inverse-ink); --s-warn-fg:var(--inverse-ink);
      --shadow-color:transparent;`

// appTokensLight is vaultr's own light adaptation — same accent hue, radius
// and spacing scale as dark, but not part of DESIGN-linear.app.md (its "Known
// Gaps" section notes Linear ships no light mode). Cool near-neutral grays,
// not the old warm cream/pure-black pairing.
const appTokensLight = `
      --ink-rgb:24,24,27;
      --bg:#fcfcfc; --fg:#18181b;
      --muted:#6b6e76; --surface-soft:#f6f6f7; --surface-2:#ececf0;
      --canvas:#ffffff; --muted-soft:#9a9da3; --body:#3f4147;
      --border:#e7e7ea; --border-strong:#d4d4d8;
      --accent-hov:#4c56c8;
      --link:#111111;
      --link-ul:var(--accent); --link-ul-hov:var(--accent-hov);
      --cover-dir:#6b7280;
      --h1:#111111; --h2:#1f2937; --h3:#374151; --h4:#6b7280;
      --prose-body:#374151;
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
      --shadow-color:transparent;`

// appTokensCSS is the theme token block — drop it inside a <style> tag.
// Dark values live on bare :root (Linear's canvas is the default look); the
// light adaptation applies via an explicit data-theme="light" attribute on
// <html>, or automatically when the OS prefers light and no explicit choice
// has been made (mirrors the artifact-theming convention: dark-first, redefine
// only light under :root[data-theme="light"] and the guarded media query).
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
    }`
