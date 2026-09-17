// Public entry point for the CodeMirror-based live-preview editor (Phase 0
// foundation). Everything under cm-live/ is app-agnostic — no fetch calls,
// no Vaultr routing, no DOM ids from drawer.js — app wiring (image upload,
// wikilink navigation, autosave, tab state) stays in drawer.js and is
// injected here through `options`, the same seam __vaultrDE currently uses
// to configure Milkdown's plugins. That separation is what let this be
// built and demoed (see ../../livepreview-demo/) without touching the live
// app, and is what should let Phase 1-4 extend it without re-architecting.
import { markdown } from '@codemirror/lang-markdown';
import { GFM } from '@lezer/markdown';
import { languages } from '@codemirror/language-data';
import { syntaxHighlighting } from '@codemirror/language';
import { wikiSyntax } from './wiki-syntax.js';
import { frontmatterSyntax } from './frontmatter-syntax.js';
import { livePreviewPlugin, livePreviewAtomicRanges } from './live-preview.js';
import { horizontalRuleField } from './horizontal-rule-field.js';
import { linkClickHandler } from './link-click.js';
import { livePreviewTheme, codeHighlightStyle } from './theme.js';

export { wikiSyntax, wikiNodeInner, splitWikiLinkInner } from './wiki-syntax.js';
export { frontmatterSyntax } from './frontmatter-syntax.js';
export { nodeDecorators } from './decorators.js';
export { livePreviewPlugin, livePreviewAtomicRanges } from './live-preview.js';
export { horizontalRuleField } from './horizontal-rule-field.js';
export { linkClickHandler } from './link-click.js';
export { livePreviewTheme, codeHighlightStyle } from './theme.js';
export { listIndentExtension } from './list-indent.js';
export {
  WikiLinkWidget,
  WikiImageWidget,
  TaskCheckboxWidget,
  MarkdownImageWidget,
  BulletWidget,
  HorizontalRuleWidget,
} from './widgets.js';
export { selectionTouchesRange, selectionTouchesLine, selectionInsideRange } from './selection.js';

// Shared by live-preview mode and (Phase 4) plain source mode, so toggling
// between them reconfigures a Compartment around the *same* parse instead
// of swapping to a differently-configured language and re-parsing from
// scratch.
// codeLanguages: languages — @codemirror/lang-markdown's stock hook for
// fenced-code-content parsing. Matches a fence's info string ("```js")
// against @codemirror/language-data's registry and, on a hit, nests that
// language's own grammar inside the FencedCode node instead of leaving its
// content as opaque text — the standard, documented way to get real code
// tokens for syntaxHighlighting(codeHighlightStyle) (theme.js) to color.
export function wikiMarkdownLanguage() {
  return markdown({ extensions: [GFM, wikiSyntax, frontmatterSyntax], codeLanguages: languages });
}

/**
 * @param {object} [options]
 * @param {(filename: string) => string} [options.resolveImageSrc]
 *   Builds the <img src> for a ![[wikiimage]] widget. Defaults to using the
 *   filename verbatim, which only works for demo/test fixtures.
 * @param {(target: string, alias: string|null, event: MouseEvent) => void}
 *   [options.onWikiLinkClick] Called when a [[wikilink]] chip is clicked.
 */
export function livePreviewExtensions(options) {
  return [
    wikiMarkdownLanguage(),
    livePreviewPlugin.of(options),
    livePreviewAtomicRanges(),
    // Frontmatter used to need its own opt-in StateField here (a rendered
    // metadata CARD via a block:true widget — see git history for
    // frontmatter-field.js/frontmatter-widget.js) because collapsing
    // multi-line raw YAML into a totally different rendered structure hit a
    // real CM6 height-map crash under a realistic transaction sequence
    // (edit the card, then a full-document replace — exactly what every
    // mode toggle/note switch does). decorateFrontmatter (decorators.js) is
    // a from-scratch replacement that never collapses multiple lines into
    // one block: it prettifies each YAML line in place (key label, tag
    // chips, collapsing only single lines at a time via the same
    // widgetBuffer-safe technique the GFM table decorator uses), so it's
    // just another entry in nodeDecorators — no separate field, no
    // block:true, no exclusion needed here.
    horizontalRuleField(),
    linkClickHandler(),
    livePreviewTheme,
    syntaxHighlighting(codeHighlightStyle),
  ];
}
