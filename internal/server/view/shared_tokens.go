package view

// Typography tokens (--text-*/--fw-*/--lh-* below) are the chrome-density
// subset of DESIGN.md's type scale actually wired into this app — dense UI
// chrome tops out at --text-title-lg (22px); there's no display tier.

// ── Token architecture ──────────────────────────────────────────────────────
//
// Visual language: DESIGN.md — near-black surface ladder, one lavender-blue
// accent, hairline borders over shadows, Inter throughout. Dark ships by
// default; light is this app's own from-scratch adaptation, kept for a
// comfortable reading surface in bright rooms.
//
// Four layers:
//
//  1. Primitives — raw tuples and theme-invariant literals (accent hue, its
//     on-color text, the true white/black inverse pair, modal scrim). Only
//     Layer 2 tokens reference these directly.
//
//  2. Semantic tokens — what components consume (--bg, --fg, --border,
//     --accent, etc.). Dark lives on bare :root; :root[data-theme="light"]
//     and the matching prefers-color-scheme block override just what
//     differs. Most rgba(var(--ink-rgb),…)/rgba(var(--accent-rgb),…) tokens
//     don't need restating per theme — overriding the base RGB triple
//     re-tints them automatically.
//
//  3. Visual-regime tokens — depth. Elevation levels 0–3 use no box-shadow;
//     depth comes from the surface ladder and hairline borders, plus a 2px
//     focus ring (level 4). A handful of floating layers (dialogs, settings
//     modal, search overlay, drawer popovers) opt into a shadow via
//     `box-shadow: var(--shadow-*) var(--shadow-color)`; everything else
//     stays flat. --r-* and --bd-w are generic structural scales, not
//     regime-specific.
//
//  4. Component-scoped tokens — narrow, single-purpose tokens (search
//     overlay, count badges, etc.) that rarely change with theme.
//
// ── State-class convention ──────────────────────────────────────────────
//
// New JS/Alpine-toggled state classes use an `is-*` prefix (is-open,
// is-active, is-selected, is-danger, is-visible, ...) — it reads
// unambiguously as "current state," not a structural/type modifier, and
// it's what most components already use. A handful of established shared
// primitives predate this and keep their bare-adjective names instead
// (.icon-btn.active, .seg-btn.active in base.css; .cselect-btn.open,
// .cselect-option.sel in cselect.css) — those aren't being renamed since
// each has many call sites and renaming buys nothing; don't add new bare
// state classes elsewhere, reach for `is-*` instead.
const appTokensShared = `
      /* ═══ Layer 1: primitives — raw tuples plus theme-invariant literals ═══ */
      --accent-rgb:94,106,210;
      --inverse-canvas:#ffffff; --inverse-ink:#000000;
      /* Brand accent — scarce by design: primary CTA, focus ring, link
         emphasis. Same hue both themes; only --accent-hov shifts per theme. */
      --accent:#5e6ad2; --accent-fg:#ffffff;
      /* Selected/active surface — surface-2 lift + full --fg text, not an
         accent fill (DESIGN.md reserves the accent for CTA/focus/links).
         The one treatment for "currently selected" app-wide (tabs, rows,
         toggles). Primary-action buttons use --accent/--accent-fg instead. */
      --control-active-bg:var(--surface-2); --control-active-fg:var(--fg);
      /* Modal scrim — fixed black veil regardless of theme. One recipe
         (this color + --glass-scrim-filter's blur) backs every floating
         layer that dims the app behind it: dialogs, the settings modal,
         search, and the drawer/inbox-sheet edge-docked panels. (The
         lightbox is the one deliberate exception — see
         --lightbox-overlay-bg below, tuned for arbitrary photo content
         instead of app chrome.) Drawer/inbox-sheet used to get their own
         lighter, unblurred tint back when their panel left most of the
         page peeking through beside it; now that the panel fills the
         whole content pane and only the sidebar shows through, there's
         no reason for them to carry a different recipe from every other
         floating layer — so they just consume this token directly. */
      --overlay-bg:rgba(0,0,0,0.25);
      /* Accent-derived interaction tints — --accent-rgb is theme-invariant,
         so no per-theme restating needed. */
      --card-hov:rgba(var(--accent-rgb),0.10);
      --icon-hov:rgba(var(--accent-rgb),0.10);
      --tint-soft:rgba(var(--accent-rgb),0.06);
      --tint-md:rgba(var(--accent-rgb),0.10);
      --tint-strong:rgba(var(--accent-rgb),0.14);
      --input-focus-ring:rgba(var(--accent-rgb),0.16);
      --cm-selection-bg:rgba(var(--accent-rgb),0.16);
      --bq-bd:rgba(var(--accent-rgb),0.55);
      /* List/checkbox marks use accent, not ink — the one place accent
         recurs per-item in prose (marks list structure, not content). */
      --ul-mk:var(--accent); --ol-mk:var(--accent);
      /* Code surfaces — ink-derived, re-tint automatically via --ink-rgb. */
      --code-bg:rgba(var(--ink-rgb),0.055); --code-bd:rgba(var(--ink-rgb),0.10);
      /* --hairline vs --border/--border-strong is a border-weight tier, not
         a "chrome vs content" one: use --hairline for a SECONDARY seam,
         where the two sides are already told apart by something else (a
         background-tint change, or both sides belonging to the same
         floating card) — the line is only a confirming detail, so it can
         all but disappear. Reach for --border/--border-strong instead
         whenever the line is the ONLY separator, or the element floats
         over an unpredictable backdrop and needs to read as a distinct
         object regardless of what's behind it.
           --hairline: app-shell seams (.home-side/.settings-sidebar's
             sidebar/content border, every "-head"/tab-bar/tool-bar chrome
             strip's border-bottom) — bg vs surface-soft already marks the
             boundary — AND a header-to-body seam inside one floating card
             (.cfg-section-head, .mdt-header) — the card's own outer
             --border-strong already carries the "distinct object" job, so
             the internal title/body seam is secondary too.
           --border/--border-strong: a card/dialog/dropdown's own outer
             edge, and any divider that's the sole boundary between two
             same-background regions (list-row separators, table cell
             lines, field rows) — remove the line there and the boundary
             is gone. */
      --hairline:rgba(var(--ink-rgb),0.13);
      /* Nav-item label — dimmer than --fg so sidebar/section nav reads as
         secondary chrome; hover/active states restore full --fg. */
      --nav-fg:rgba(var(--ink-rgb),0.82);
      --th-bg:rgba(var(--ink-rgb),0.04); --tbl-bd:rgba(var(--ink-rgb),0.12);
      --cm-active-line:rgba(var(--ink-rgb),0.04);
      --cnt-bg:rgba(var(--ink-rgb),0.07);
      --scrollbar-thumb:rgba(var(--ink-rgb),0.12); --scrollbar-thumb-hov:rgba(var(--ink-rgb),0.26);
      --nav-act:rgba(var(--ink-rgb),0.1);
      /* Content-identity palette (entity chips, avatar fallback) —
         theme-invariant: marks "which thing", not "which mode". */
      --p0:#22d3ee; --p1:#f472b6; --p2:#a78bfa; --p3:#34d399;
      --entity-type-bg:var(--p2);

      /* ═══ Layer 3: visual-regime — offset/blur shapes only; --shadow-color
         (set per theme below) supplies the one real shadow color, used only
         by the floating layers that opt in ═══ */
      --shadow-xs:0 1px 2px; --shadow-sm:0 2px 5px; --shadow-md:0 5px 14px; --shadow-lg:0 14px 34px;
      /* The one exception to "no shadow": 2px focus ring at 50% opacity
         (Elevation level 4), used by *:focus-visible below. */
      --accent-focus:#5e69d1;
      /* Radius scale, by visual weight: xs (chips) → sm (tags) → md
         (buttons/inputs, default) → lg (cards) → xl (modals/panels) → full
         (pills/avatars). Generic, not regime-specific. */
      --r-xs:4px; --r-sm:6px; --r-md:8px; --r-lg:12px; --r-xl:16px; --r-full:999px;
      /* Motion scale — three durations, by what's moving:
           fast (100ms)  hover/press feedback (bg/color/opacity)
           base (160ms)  small toggle transforms (chevrons)
           slow (220ms)  panel/scrim fades (drawer, lightbox)
         Bespoke choreography (drawer slide-in, zen-mode fades) keeps its
         own hand-tuned timing outside this scale. */
      --motion-fast:100ms; --motion-base:160ms; --motion-slow:220ms;
      /* Kept as a whole pixel — a fractional border width breaks
         anti-aliasing where a straight edge meets a radius arc. */
      --bd-w:1px;
      /* Floating-card scrim blur — shared by every layer that dims the app
         behind it, centered (search, dialogs, settings, lightbox) or
         edge-docked (drawer, inbox sheet) alike, so blur amounts can't
         drift apart. */
      --glass-scrim-filter:blur(3px);

      /* ═══ Layer 4: component-scoped tokens ═══ */
      /* Search overlay — icon/placeholder/timestamp/etc. point at --muted
         (or --muted-soft), matching the same roles elsewhere app-wide.
         --sr-dir is the one exception: a real directory label, tied to
         --cover-dir instead of the generic muted tone. */
      --srch-ic:var(--muted); --srch-ph:var(--muted-soft); --srch-av:var(--control-active-bg); --srch-backdrop:var(--overlay-bg);
      --sr-dir:var(--cover-dir); --sr-tm:var(--muted); --sr-ic:var(--muted); --sr-em:var(--muted);
      --srch-kbd-fg:var(--fg);
      --srch-panel-bd:var(--border-strong); --srch-row-bd:var(--border);

      /* Images grid & lightbox — chrome over arbitrary photo content, so
         these stay a fixed black/white veil independent of theme. */
      --lightbox-overlay-bg:rgba(0,0,0,0.4); --lightbox-viewer-bg:#0c0c0c;
      --img-chip-bg:rgba(17,17,17,0.72);
      --img-select-border:rgba(255,255,255,0.72); --img-select-bg:rgba(0,0,0,0.28);

      /* Inter carries both display and body — one continuous voice. */
      --font-ui:"Inter",-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;
      --font-sans:"Inter",-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;
      --font-mono:"JetBrains Mono",ui-monospace,monospace;
      --text-xs:0.75rem; --text-sm:0.8125rem; --text-base:0.875rem; --text-body:1rem;
      --text-title-sm:1rem; --text-title-lg:1.375rem;
      --text-2xs:0.6875rem; --text-3xs:0.625rem;
      --fw-regular:400; --fw-medium:500; --fw-semibold:600;
      --lh-tight:1.2; --lh-snug:1.3; --lh-normal:1.4; --lh-relaxed:1.5;
      --ls-cap:0.08em;
      /* Spacing — DESIGN.md's scale (4px base unit). No --space-section:
         nothing in this chrome-density UI operates at section scale. */
      --space-xxs:4px; --space-xs:8px; --space-sm:12px; --space-md:16px; --space-lg:24px;
      --space-xl:32px; --space-xxl:48px;
      --topbar-h:40px; --action-btn-sz:28px;
      /* Single source of truth for the app sidebar's width (home.css's
         .home-side) — the reading drawer and inbox sheet (drawer.css,
         home.css) size themselves off this too, so their edge-docked
         panels land flush against the sidebar's own right edge instead of
         drifting out of sync with it. */
      --home-side-w:248px;
      /* --btn-h (32px) matches DESIGN.md's button-primary spec and the
         app's form inputs. --btn-h-xs (28px) is the compact tier for
         secondary inline actions, never a primary CTA. --btn-h-sm is
         unrelated — a chrome-bar height (search overlay footers). */
      --btn-h:32px; --btn-h-sm:36px; --btn-h-xs:28px;
      --card-note-h:96px; --card-note-w-max:210px;`

// appTokensDark is the default palette (bare :root) — a near-black canvas,
// lifted off true black to avoid a flat void. Only tokens that differ from
// light live here; --ink-rgb/--accent-rgb-derived rgba() tokens re-tint on
// their own.
const appTokensDark = `
      --ink-rgb:238,240,244;
      --bg:#18191e; --fg:#eef0f4;
      --muted:#8d92a0; --surface-soft:#212229; --surface-2:#292a33;
      /* Selected-state lift for containers that already sit on
         --surface-soft (sidebars, dropdowns, floating toolbars) — stacking
         --surface-2 there only clears ~8/255, fainter than a plain hover
         tint. Such containers locally redefine --control-active-bg to this
         value instead (see .home-side, .settings-sidebar, .graph-index-col,
         .cselect-dropdown, .milkdown-tooltip). Light mode doesn't need this
         — see appTokensLight. */
      --control-active-bg-on-soft:#3d3e49;
      /* macOS Electron only (html.macos in home.css): the sidebar sits on
         the window's native vibrancy instead of a flat --surface-soft fill,
         and that material's own tint doesn't match this app's near-black
         palette — left untinted it reads lighter than --surface-soft, which
         throws off every colour tuned relative to it (active-row lift,
         hairline borders). This wash pulls it back toward --surface-soft's
         actual tone while staying translucent enough to still read as
         glass. The border pairs a white-alpha line instead of --border's
         flat hex, since a flat hex's contrast against a blurred, partly
         see-through backdrop can't be predicted the way it can against a
         solid fill. */
      --sidebar-glass-bg:rgba(33,34,41,0.6); --sidebar-glass-border:rgba(255,255,255,0.1);
      /* A translucent lift instead of --control-active-bg-on-soft's flat
         hex — a solid fill on top of the vibrancy read as a pasted-on chip
         rather than an integrated highlight. Only swapped in for the row's
         own background (home.css); --control-active-bg itself stays opaque
         since the active-row count badge also uses it as a text colour. */
      --sidebar-glass-active:rgba(255,255,255,0.08);
      /* The shared --scrollbar-thumb is tuned against a flat, predictable
         surface; over the blurred vibrancy it reads inconsistently (too
         bright here in dark, too dark in light — see the light-theme
         definition below) since the thumb's contrast now depends on
         whatever the vibrancy is blending with underneath. Toned down
         instead of matched to the flat value, and scoped to the glass
         sidebar only (home.css's html.macos .home-side-nav rule). */
      --sidebar-glass-scrollbar-thumb:rgba(255,255,255,0.10); --sidebar-glass-scrollbar-thumb-hov:rgba(255,255,255,0.20);
      --canvas:#ffffff; --muted-soft:#656a78; --body:#d0d6e0;
      --border:#35363f; --border-strong:#44454f;
      --accent-hov:#7b86e8;
      --link:#eef0f4;
      --link-ul:var(--accent); --link-ul-hov:var(--accent-hov);
      --cover-dir:#8d92a0;
      --h1:#eef0f4; --h2:#d0d6e0; --h3:#adb2bc; --h4:#8d92a0;
      --prose-body:#ccd1db;
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
      /* Text for solid fills of --s-ok/-err/-warn (a handful of status
         buttons/pills) — these are tuned bright for running text, too
         light for white text on top; dark ink reads correctly instead. */
      --s-ok-fg:var(--inverse-ink); --s-err-fg:var(--inverse-ink); --s-warn-fg:var(--inverse-ink);
      /* Lift-off shadow for floating layers that opt in (dialogs, settings
         modal, search overlay, drawer popovers). Everything else renders
         no shadow. */
      --shadow-color:rgba(0,0,0,0.55);`

// appTokensLight is vaultr's own light adaptation — same hue/radius/spacing
// scale as dark, but its own from-scratch palette (DESIGN.md documents no
// marketing-page light mode to inherit from). Cool near-neutral grays, not
// the old warm cream/pure-black pairing.
const appTokensLight = `
      --ink-rgb:24,24,27;
      --bg:#f9f9fb; --fg:#18181b;
      --muted:#6d7080; --surface-soft:#f5f5f7; --surface-2:#e3e3e9;
      /* Light's surface-soft/surface-2 are already far enough apart — no
         extra lift needed here (see the dark-theme definition above). */
      --control-active-bg-on-soft:var(--surface-2);
      /* Light-only override of the shared (Layer 1) scrollbar tokens — as
         --bg/--surface-soft climbed toward white, a thumb tuned against the
         old darker canvas started reading as barely-there; nudged up so it
         stays legible against the brighter light surfaces. Dark keeps the
         shared value (appTokensShared) since its canvas didn't change. */
      --scrollbar-thumb:rgba(var(--ink-rgb),0.16); --scrollbar-thumb-hov:rgba(var(--ink-rgb),0.30);
      /* See the dark-theme definition above for why these exist. */
      /* Brighter and a touch more opaque than --surface-soft's flat value —
         the native vibrancy underneath (main.js's "sidebar" material) reads
         dull/dark on its own, so this wash leans lighter to compensate
         while still letting some of the native blur show through. */
      --sidebar-glass-bg:rgba(250,250,251,0.72); --sidebar-glass-border:rgba(0,0,0,0.08);
      --sidebar-glass-active:rgba(0,0,0,0.06);
      /* See the dark-theme definition above for why this exists separately
         from --scrollbar-thumb: over the glass sidebar's vibrancy, the
         thumb reads darker/grayer than the flat --scrollbar-thumb value
         (0.16) does elsewhere, so this one stays lower. */
      --sidebar-glass-scrollbar-thumb:rgba(0,0,0,0.11); --sidebar-glass-scrollbar-thumb-hov:rgba(0,0,0,0.20);
      --canvas:#ffffff; --muted-soft:#9b9eac; --body:#3f4147;
      --border:#d2d2d9; --border-strong:#bcbcc5;
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
      --shadow-color:rgba(15,15,25,0.10);`

// appTokensCSS is the theme token block — drop it inside a <style> tag.
// Dark lives on bare :root; light applies via data-theme="light" or an
// unforced OS light preference.
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
    /* App-wide press-feedback floor — any component-scoped ":active" rule
       (higher specificity) still wins over this. */
    button:not(:disabled):active {
      opacity: 0.85;
    }`
