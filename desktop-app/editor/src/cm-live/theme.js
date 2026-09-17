// Ported from note_editor_prose.css (+ shared prose/frontmatter rules).
// Values are old literals or deliberate, commented deviations.
//
// CM6 notes:
//   1. No <h1>/<li> — only .cm-line; DOM-adjacency rules live in decorators.js.
//   2. Blank lines are real — don't port inter-block margin (would double-gap).
// var() fallbacks: light-mode tokens so livepreview-demo renders without the app sheet.
import { EditorView } from '@codemirror/view';
import { HighlightStyle } from '@codemirror/language';
import { tags } from '@lezer/highlight';

export const livePreviewTheme = EditorView.theme({
  '.cm-content': {
    fontFamily: 'var(--font-sans, "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif)',
    fontSize: 'var(--text-body, 1rem)',
    color: 'var(--prose-body, #374151)',
    lineHeight: '1.75',
    // !important: drawer.css `#drawer-edit-area .cm-content { padding: 0 }` otherwise wins.
    padding: '24px 0 !important',
  },

  // ── Headings ────────────────────────────────────────────────────────────
  // Explicit lineHeight: large fonts without it leave hit-box taller than glyphs
  // → click lands on next line. h1/h2 bumped from old 1.2/1.3 (Cal Sans + CJK
  // measured taller). Padding not margin — CM6 ResizeObserver ignores margin.
  // !important: drawer.css zeros .cm-line padding via ID selector.
  '.cm-lp-heading': { fontWeight: '600' },
  '.cm-lp-h1': {
    fontFamily: '"Cal Sans", var(--font-sans, "Inter", sans-serif)',
    fontSize: '1.75em',
    lineHeight: '1.45', // old: 1.2
    letterSpacing: '-0.02em',
    color: 'var(--h1, #111111)',
    paddingTop: '1.5em !important',
    paddingBottom: '0.65em !important',
  },
  '.cm-lp-h2': {
    fontFamily: '"Cal Sans", var(--font-sans, "Inter", sans-serif)',
    fontSize: '1.375em',
    lineHeight: '1.4', // old: 1.3
    letterSpacing: '-0.014em',
    color: 'var(--h2, #1f2937)',
    paddingTop: '1.75em !important',
    paddingBottom: '0.65em !important',
  },
  '.cm-lp-h3': {
    fontFamily: '"Cal Sans", var(--font-sans, "Inter", sans-serif)',
    fontSize: '1.125em',
    lineHeight: '1.4',
    letterSpacing: '-0.008em',
    color: 'var(--h3, #374151)',
    paddingTop: '1.5em !important',
    paddingBottom: '0.5em !important',
  },
  // Editor-only (--lp-h4/--lp-h5): don't touch reader --h4.
  // h4 = body size + --h3 color (can't outrank body without outranking h3).
  // h5/h6: wider size gap — uppercase/tracking is inert for CJK.
  '.cm-lp-h4': {
    fontSize: '1em',
    lineHeight: '1.5',
    color: 'var(--lp-h4, #374151)',
    paddingTop: '1.25em !important',
    paddingBottom: '0.5em !important',
  },
  '.cm-lp-h5': {
    fontSize: '0.8125em', // was 0.875em
    lineHeight: '1.5',
    letterSpacing: '0.08em',
    textTransform: 'uppercase',
    fontWeight: '500',
    color: 'var(--lp-h5, #6b7280)',
    paddingTop: '1.25em !important',
    paddingBottom: '0.5em !important',
  },
  '.cm-lp-h6': {
    fontSize: '0.8125em',
    lineHeight: '1.5',
    letterSpacing: '0.08em',
    textTransform: 'uppercase',
    fontWeight: '500',
    color: 'var(--lp-h5, #6b7280)',
    paddingTop: '1.25em !important',
    paddingBottom: '0.5em !important',
  },

  '.cm-lp-strong': { fontWeight: '600', color: 'var(--prose-strong, #111111)' },
  '.cm-lp-em': { fontStyle: 'italic', color: 'var(--prose-em, #374151)' },
  '.cm-lp-strike': { textDecoration: 'line-through', color: 'var(--muted, #6d7080)' },
  '.cm-lp-code': {
    fontFamily: 'var(--font-mono, "JetBrains Mono", ui-monospace, monospace)',
    fontSize: '0.85em',
    color: 'var(--code-tx, #374151)',
    background: 'var(--th-bg, rgba(24,24,27,0.04))',
    borderRadius: 'var(--r-xs, 4px)',
    padding: '0.15em 0.42em',
  },

  // ── Blockquote ──────────────────────────────────────────────────────────
  // !important: drawer.css zeros .cm-line padding.
  // borderRadius:0 — global reset would round the inset bar into brackets.
  '.cm-lp-quote': {
    fontStyle: 'italic',
    paddingTop: '0.35rem !important',
    paddingBottom: '0.35rem !important',
    paddingLeft: '1.5rem !important',
    fontSize: '1em',
    lineHeight: '1.75',
    color: 'var(--bq-tx, rgba(55,65,81,0.88))',
    position: 'relative',
    borderRadius: '0',
  },
  // Inset box-shadow stands in for old ::before bar (no pseudo on line decos).
  '.cm-line.cm-lp-quote': {
    boxShadow: 'inset 2px 0 0 0 var(--bq-bd, rgba(94,106,210,0.55))',
  },
  // Nested: +1.5rem pad + extra inset bar per level (cap 4).
  '.cm-lp-quote-d2': { paddingLeft: '3rem !important' },
  '.cm-line.cm-lp-quote-d2': {
    boxShadow:
      'inset 2px 0 0 0 var(--bq-bd, rgba(94,106,210,0.55)), inset calc(2px + 1.5rem) 0 0 0 var(--bq-bd, rgba(94,106,210,0.55))',
  },
  '.cm-lp-quote-d3': { paddingLeft: '4.5rem !important' },
  '.cm-line.cm-lp-quote-d3': {
    boxShadow:
      'inset 2px 0 0 0 var(--bq-bd, rgba(94,106,210,0.55)), inset calc(2px + 1.5rem) 0 0 0 var(--bq-bd, rgba(94,106,210,0.55)), inset calc(2px + 3rem) 0 0 0 var(--bq-bd, rgba(94,106,210,0.55))',
  },
  '.cm-lp-quote-d4': { paddingLeft: '6rem !important' },
  '.cm-line.cm-lp-quote-d4': {
    boxShadow:
      'inset 2px 0 0 0 var(--bq-bd, rgba(94,106,210,0.55)), inset calc(2px + 1.5rem) 0 0 0 var(--bq-bd, rgba(94,106,210,0.55)), inset calc(2px + 3rem) 0 0 0 var(--bq-bd, rgba(94,106,210,0.55)), inset calc(2px + 4.5rem) 0 0 0 var(--bq-bd, rgba(94,106,210,0.55))',
  },

  // ── Links — pointer only when strictly inside (cm-lp-link-hit from link-click.js).
  '.cm-lp-link': {
    color: 'var(--link, #111111)',
    fontWeight: '400',
    textDecoration: 'underline',
    textDecorationColor: 'var(--link-ul, #5e6ad2)',
    textUnderlineOffset: '3px',
    textDecorationThickness: '1px',
  },
  '.cm-lp-link.cm-lp-link-hit': { cursor: 'pointer' },
  '.cm-lp-link:hover': {
    background: 'var(--tint-soft, rgba(94,106,210,0.06))',
    textDecorationColor: 'var(--link-ul-hov, #4c56c8)',
    borderRadius: '0', // defeat global 8px reset
  },

  // ── Lists ───────────────────────────────────────────────────────────────
  '.cm-lp-list-line': {
    // Was 1.7 — nearly identical to surrounding paragraph text's 1.75, so a
    // list never actually read as more compact than prose no matter how far
    // the padding above was cut; this is the value that actually needed to
    // move. 1.4 matches how mainstream editors (Notion, Typora) tighten
    // list-item line-height specifically, distinct from paragraph text.
    lineHeight: '1.4 !important',
    paddingTop: '0.1em !important',
    paddingBottom: '0.1em !important',
  },
  // A blank "loose list" line between two items — collapsed to a hairline
  // (same trick as the frontmatter/table hairline widgets) so cm-lp-list-line's
  // own padding is what controls the visible gap, not the blank line's
  // default full-text-line height.
  '.cm-lp-list-gap': { fontSize: '1px', lineHeight: '1px', padding: '0 !important' },
  '.cm-lp-list-mark-ol': { color: 'var(--ol-mk, #5e6ad2)', fontSize: '0.85em' },
  '.cm-lp-bullet': { color: 'var(--ul-mk, #5e6ad2)' },

  // ── Task checkboxes — real <input>, same look as old ::before SVG.
  '.cm-lp-task-checkbox': {
    appearance: 'none',
    WebkitAppearance: 'none',
    display: 'inline-block',
    width: '14px',
    height: '14px',
    verticalAlign: 'middle',
    marginRight: '0.5em',
    marginTop: '-0.15em',
    border: '1.5px solid var(--ul-mk, #5e6ad2)',
    borderRadius: 'var(--r-xs, 4px)',
    background: 'transparent',
    cursor: 'pointer',
  },
  '.cm-lp-task-checkbox:checked': {
    background: 'var(--ul-mk, #5e6ad2)',
    borderColor: 'transparent',
    backgroundImage:
      "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 12 12'%3E%3Cpolyline points='2,6 5,9 10,3' fill='none' stroke='%23fff' stroke-width='1.8' stroke-linecap='square' stroke-linejoin='miter'/%3E%3C/svg%3E\")",
    backgroundRepeat: 'no-repeat',
    backgroundSize: '100% 100%',
  },

  // ── HR — padding not margin (ResizeObserver); reduced vs old 2em (blank
  // line already adds ~1.75em). margin:0 clears <hr> UA default.
  '.cm-lp-hr': {
    border: 'none',
    margin: '0',
    borderRadius: '0',
    borderTop: '1px solid var(--border, #d2d2d9)',
    padding: '0.5em 0',
  },
  '.cm-lp-hr.cm-lp-hr-tight-bottom': {
    paddingBottom: '0',
  },

  // ── Fenced code — per-line chrome; borderRadius:0 so middle lines don't
  // pick up global 8px (would scallop the box edge).
  '.cm-lp-codeblock': {
    fontFamily: 'var(--font-mono, "JetBrains Mono", ui-monospace, monospace)',
    fontSize: 'var(--text-base, 0.875rem)',
    lineHeight: '1.65',
    color: 'var(--pre-tx, #111111)',
    background: 'var(--th-bg, rgba(24,24,27,0.04))',
    borderRadius: '0',
    paddingLeft: '1.5rem !important',
    paddingRight: '1.5rem !important',
  },
  '.cm-lp-codeblock-first': {
    position: 'relative', // anchors cm-lp-code-lang's absolute corner badge
    borderTopLeftRadius: 'var(--r-sm, 6px)',
    borderTopRightRadius: 'var(--r-sm, 6px)',
    paddingTop: '0.75rem !important',
  },
  '.cm-lp-codeblock-last': {
    borderBottomLeftRadius: 'var(--r-sm, 6px)',
    borderBottomRightRadius: 'var(--r-sm, 6px)',
    paddingBottom: '0.75rem !important',
  },
  // The opening/closing fence line itself, once its ``` marker is hidden, is
  // otherwise pure dead space — its own font line-height was stacking with
  // codeblock-first/last's padding above, which is what actually made the
  // empty top/bottom gap so big. Collapsed to a hairline (frontmatter/table/
  // list-gap's trick) so -first/-last's padding is the only thing left
  // controlling that gap — including when there's a language label: that's
  // pulled out of flow onto its own corner badge (below) instead of holding
  // the line open, so it no longer costs a whole extra line to show.
  '.cm-lp-codeblock-marker-line': { fontSize: '1px', lineHeight: '1px' },
  '.cm-lp-code-lang': {
    position: 'absolute',
    top: '0.7rem',
    right: '1.5rem',
    color: 'var(--muted, #6d7080)',
    // rem, not em, and lineHeight reset: this now lives inside
    // cm-lp-codeblock-marker-line, whose own font-size/line-height are
    // collapsed to 1px — em/inherited-px sizing here would inherit that and
    // round down to nothing (line-height's "1px" is a literal length, so it
    // inherits as-is regardless of this element's own font-size).
    fontSize: '0.75rem',
    lineHeight: 'normal',
    textTransform: 'uppercase',
    letterSpacing: '0.04em',
  },

  // ── Wikilink ────────────────────────────────────────────────────────────
  '.cm-lp-wikilink': {
    display: 'inline',
    color: 'var(--link, #111111)',
    fontWeight: '400',
    textDecoration: 'underline',
    textUnderlineOffset: '3px',
    textDecorationColor: 'var(--link-ul, #5e6ad2)',
    textDecorationThickness: '1px',
    cursor: 'pointer',
    userSelect: 'none',
  },
  '.cm-lp-wikilink::before': {
    content: '""',
    display: 'inline-block',
    width: '11px',
    height: '13px',
    marginRight: '0.2em',
    verticalAlign: 'middle',
    marginTop: '-0.1em',
    backgroundColor: 'var(--muted, #6d7080)',
    WebkitMaskImage:
      'url("data:image/svg+xml,%3Csvg xmlns=\'http://www.w3.org/2000/svg\' viewBox=\'0 0 24 24\' fill=\'none\' stroke=\'black\' stroke-width=\'2\' stroke-linecap=\'round\' stroke-linejoin=\'round\'%3E%3Cpath d=\'M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z\'/%3E%3Cpath d=\'M14 2v4a2 2 0 0 0 2 2h4\'/%3E%3C/svg%3E")',
    maskImage:
      'url("data:image/svg+xml,%3Csvg xmlns=\'http://www.w3.org/2000/svg\' viewBox=\'0 0 24 24\' fill=\'none\' stroke=\'black\' stroke-width=\'2\' stroke-linecap=\'round\' stroke-linejoin=\'round\'%3E%3Cpath d=\'M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z\'/%3E%3Cpath d=\'M14 2v4a2 2 0 0 0 2 2h4\'/%3E%3C/svg%3E")',
    WebkitMaskSize: '100% 100%',
    maskSize: '100% 100%',
    WebkitMaskRepeat: 'no-repeat',
    maskRepeat: 'no-repeat',
    borderRadius: '0',
  },
  '.cm-lp-wikilink:hover': {
    background: 'var(--tint-soft, rgba(94,106,210,0.06))',
    textDecorationColor: 'var(--link-ul-hov, #4c56c8)',
    borderRadius: '0',
  },
  '.cm-lp-wikilink-raw': { color: 'var(--link, #111111)' },

  // ── Images ──────────────────────────────────────────────────────────────
  '.cm-lp-wikiimage': { display: 'inline-block', verticalAlign: 'middle', maxWidth: '100%' },
  '.cm-lp-wikiimage img': {
    display: 'block',
    maxWidth: '100%',
    maxHeight: '360px',
    border: '1px solid var(--tbl-bd, rgba(24,24,27,0.12))',
    borderRadius: 'var(--r-sm, 6px)',
  },
  '.cm-lp-wikiimage-raw': { color: 'var(--link, #7c3aed)' },

  // ── Frontmatter — per-line card chrome (first/last edges only).
  '.cm-lp-fm-line': {
    display: 'block',
    background: 'var(--cnt-bg, rgba(24,24,27,0.055))',
    borderLeft: '1px solid var(--border, rgba(0,0,0,0.12))',
    borderRight: '1px solid var(--border, rgba(0,0,0,0.12))',
    borderRadius: '0',
    paddingLeft: '0.875rem !important',
    paddingRight: '0.875rem !important',
    paddingTop: '0.15rem !important',
    paddingBottom: '0.15rem !important',
    fontSize: 'var(--text-base, 0.875rem)',
    lineHeight: '1.6',
  },
  '.cm-lp-fm-first': {
    borderTop: '1px solid var(--border, rgba(0,0,0,0.12))',
    borderTopLeftRadius: 'var(--r-sm, 8px)',
    borderTopRightRadius: 'var(--r-sm, 8px)',
    paddingTop: '0.5rem !important',
  },
  '.cm-lp-fm-last': {
    borderBottom: '1px solid var(--border, rgba(0,0,0,0.12))',
    borderBottomLeftRadius: 'var(--r-sm, 8px)',
    borderBottomRightRadius: 'var(--r-sm, 8px)',
    paddingBottom: '0.5rem !important',
  },
  '.cm-lp-fm-label': {
    display: 'inline-block',
    fontSize: 'var(--text-2xs, 0.6875rem)',
    verticalAlign: 'middle',
    whiteSpace: 'nowrap', // overflow, don't wrap ":" if ch estimate is short
  },
  '.cm-lp-fm-key': {
    fontWeight: '500',
    letterSpacing: '0.07em',
    textTransform: 'uppercase',
    color: 'var(--muted, rgba(60,60,67,0.75))',
  },
  '.cm-lp-fm-colon': { color: 'var(--muted, rgba(60,60,67,0.5))' },
  '.cm-lp-fm-val': { color: 'var(--muted, rgba(60,60,67,0.75))' },
  '.cm-lp-fm-link': {
    color: 'var(--link, #2563eb)',
    textDecoration: 'underline',
    textDecorationColor: 'var(--link-ul, rgba(37,99,235,0.35))',
    textDecorationThickness: '1px',
    textUnderlineOffset: '2px',
    cursor: 'pointer',
  },
  '.cm-lp-fm-link:hover': {
    background: 'var(--tint-soft, rgba(94,106,210,0.06))',
    textDecorationColor: 'var(--link-ul-hov, #4c56c8)',
    borderRadius: '0',
  },
  '.cm-lp-fm-tag': {
    margin: '0 0.25em 0 0',
    padding: '0.05em 0.5em',
    borderRadius: 'var(--r-full, 999px)',
    background: 'var(--nav-act, rgba(0,0,0,0.06))',
    color: 'var(--muted, rgba(60,60,67,0.85))',
  },
  '.cm-lp-fm-more': {
    marginLeft: '0.6em',
    padding: '0.05em 0.5em',
    borderRadius: 'var(--r-full, 999px)',
    background: 'var(--nav-act, rgba(0,0,0,0.06))',
    color: 'var(--muted, rgba(60,60,67,0.75))',
    fontSize: '0.92em',
    cursor: 'pointer',
  },
  '.cm-lp-fm-more:hover': { color: 'var(--fg, #1a1a1a)' },
  // Shrink line font so CM6 widgetBuffers collapse with the hairline widget.
  '.cm-lp-fm-collapsed': { fontSize: '1px', lineHeight: '1px' },
  '.cm-lp-fm-collapsed-widget': { display: 'inline-block', height: '0' },

  // ── Tables — inline-block cells + % widths (not display:table per row).
  '.cm-lp-table-cell': {
    display: 'inline-block',
    boxSizing: 'border-box',
    color: 'var(--td-tx, #374151)',
    fontSize: 'var(--text-base, 0.875rem)',
    padding: '0.45rem 0.75rem',
    borderBottom: '1px solid var(--tbl-bd, rgba(24,24,27,0.12))',
    borderRight: '1px solid var(--tbl-bd, rgba(24,24,27,0.12))',
    borderRadius: '0',
    verticalAlign: 'top',
  },
  '.cm-lp-table-cell-first': { borderLeft: '1px solid var(--tbl-bd, rgba(24,24,27,0.12))' },
  '.cm-lp-table-row-header .cm-lp-table-cell': {
    background: 'var(--th-bg, rgba(24,24,27,0.04))',
    color: 'var(--th-tx, #9ca3af)',
    fontWeight: '500',
    borderTop: '1px solid var(--tbl-bd, rgba(24,24,27,0.12))',
  },
  '.cm-lp-table-row-header .cm-lp-table-cell-first': { borderTopLeftRadius: 'var(--r-sm, 6px)' },
  '.cm-lp-table-row-header .cm-lp-table-cell-last': { borderTopRightRadius: 'var(--r-sm, 6px)' },
  // Keep last-row bottom border — no outer <table> to close the box.
  '.cm-lp-table-row-last .cm-lp-table-cell-first': { borderBottomLeftRadius: 'var(--r-sm, 6px)' },
  '.cm-lp-table-row-last .cm-lp-table-cell-last': { borderBottomRightRadius: 'var(--r-sm, 6px)' },
  // Hairline: shrink line font + inline-block widget (not block — splits buffers).
  '.cm-lp-table-delim': { fontSize: '1px', lineHeight: '1px' },
  '.cm-lp-table-delim-widget': { display: 'inline-block', height: '0' },
});

// Code-fence tokens only — omit markdown structural tags (decorators.js owns those).
// Tokyo Night palette via --syn-* (shared_tokens.go): dark on bare :root, Tokyo
// Night Light under [data-theme="light"]. Fallbacks are the light values, per
// this file's var()-fallback convention above.
export const codeHighlightStyle = HighlightStyle.define([
  { tag: tags.keyword, color: 'var(--syn-keyword, #5a4a78)' },
  // No labelName — fence CodeInfo is styled as cm-lp-code-lang.
  { tag: [tags.atom, tags.bool], color: 'var(--syn-const, #965027)' },
  { tag: [tags.literal, tags.inserted], color: 'var(--syn-const, #965027)' },
  // Empty rule: block tags.url inheriting --syn-const from literal (cm-lp-link owns it).
  { tag: tags.url },
  { tag: [tags.string, tags.deleted], color: 'var(--syn-string, #385f0d)' },
  { tag: [tags.regexp, tags.escape, tags.special(tags.string)], color: 'var(--syn-escape, #0f4b6e)' },
  { tag: tags.definition(tags.variableName), color: 'var(--syn-def, #34548a)' },
  { tag: tags.local(tags.variableName), color: 'var(--syn-param, #8f5e15)' },
  { tag: [tags.typeName, tags.namespace], color: 'var(--syn-type, #166775)' },
  { tag: tags.className, color: 'var(--syn-class, #33635c)' },
  { tag: [tags.special(tags.variableName), tags.macroName], color: 'var(--syn-builtin, #8c4351)' },
  { tag: tags.definition(tags.propertyName), color: 'var(--syn-property, #33635c)' },
  { tag: tags.comment, color: 'var(--syn-comment, #848cb1)', fontStyle: 'italic' },
  { tag: tags.invalid, color: 'var(--syn-invalid, #c53b53)' },
]);
