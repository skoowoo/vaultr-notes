package view

// noteSharedCSS = frontmatter CSS + prose typography.
// noteEditorCSS = frontmatter CSS + ProseMirror editor typography.
// Both are assembled from embedded asset files.
var noteSharedCSS = noteFrontmatterCSS + noteSharedProseCSS
var noteEditorCSS = noteFrontmatterCSS + noteEditorProseCSS

// noteSharedJS is a complete <script> block shared by all note-rendering pages.
var noteSharedJS = "<script>\n" + noteSharedJSBody + "</script>"
