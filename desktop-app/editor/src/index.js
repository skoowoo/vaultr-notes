// Public entry point bundled to internal/server/static/editor.js and
// imported by drawer.js. Everything the live-preview editor needs — see
// src/cm-live/ for the actual implementation; this file is just the export
// surface drawer.js's dynamic import() pulls from.
export { EditorView, keymap } from '@codemirror/view';
export { EditorState, Compartment } from '@codemirror/state';
export { markdown } from '@codemirror/lang-markdown';
export { HighlightStyle, syntaxHighlighting } from '@codemirror/language';
export { tags } from '@lezer/highlight';
export { defaultKeymap, history, historyKeymap, undo as cmUndo, redo as cmRedo } from '@codemirror/commands';
export { search, openSearchPanel, closeSearchPanel, findNext, findPrevious, replaceNext, replaceAll as cmReplaceAll, SearchQuery, getSearchQuery, setSearchQuery } from '@codemirror/search';
export {
  livePreviewExtensions,
  wikiMarkdownLanguage,
  livePreviewPlugin,
  livePreviewAtomicRanges,
  horizontalRuleField,
  linkClickHandler,
  livePreviewTheme,
  codeHighlightStyle,
  listIndentExtension,
} from './cm-live/index.js';
