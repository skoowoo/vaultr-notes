package view

// noteSharedCSS = frontmatter CSS + prose typography + neo note overrides.
// noteEditorCSS = ProseMirror editor typography (always paired with noteSharedCSS).
// Both are assembled from embedded asset files.
var noteSharedCSS = noteFrontmatterCSS + noteSharedProseCSS + neoNoteCSS
var noteEditorCSS = noteEditorProseCSS

// noteSharedJS is a complete <script> block shared by all note-rendering pages.
var noteSharedJS = "<script>\n" + noteSharedJSBody + "</script>"
