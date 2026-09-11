package view

import _ "embed"

//go:embed assets/info_dialog.css
var infoDialogCSS string

//go:embed assets/info_dialog.html
var infoDialogHTML string

//go:embed assets/info_dialog.js
var infoDialogJS string

//go:embed assets/neo.css
var neoCSS string

// shortDialogCSS, shortDialogHTML, and shortDialogJS form a self-contained
// quick-capture dialog for short notes. Drop all three into any page shell to
// get a modal reachable via the nav rail's short-note button (see navHTML) or
// the keyboard shortcut Ctrl+Shift+Space (⌘+Shift+Space on macOS).
//
// The dialog POSTs to POST /api/vault/shorts and calls
// window.__vaultrAfterVaultMutation (if defined) on success.

//go:embed assets/short_dialog.css
var shortDialogCSS string

//go:embed assets/short_dialog.html
var shortDialogHTML string

//go:embed assets/short_dialog.js
var shortDialogJS string

//go:embed assets/confirm_dialog.css
var confirmDialogCSS string

//go:embed assets/confirm_dialog.html
var confirmDialogHTML string

//go:embed assets/confirm_dialog.js
var confirmDialogJS string

//go:embed assets/drawer.css
var drawerCSS string

//go:embed assets/drawer.html
var drawerHTML string

//go:embed assets/path_ac.js
var pathAcScript string

//go:embed assets/drawer.js
var drawerScript string

//go:embed assets/agent_chat.css
var agentChatCSS string

//go:embed assets/home.css
var homeCSS string

//go:embed assets/home.html
var homeMainHTML string

//go:embed assets/home.js
var homeJS string

//go:embed assets/images.css
var imagesCSS string

//go:embed assets/images_grid.html
var imagesGridHTML string

//go:embed assets/shorts.css
var shortsCSS string

//go:embed assets/shorts_stream.html
var shortsStreamHTML string

//go:embed assets/note_frontmatter.css
var noteFrontmatterCSS string

//go:embed assets/note_shared_prose.css
var noteSharedProseCSS string

//go:embed assets/note_editor_prose.css
var noteEditorProseCSS string

//go:embed assets/note_fonts.html
var noteFontsHTML string

//go:embed assets/note_shared.js
var noteSharedJSBody string

//go:embed assets/graph.css
var graphCSS string
