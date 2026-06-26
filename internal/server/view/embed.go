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

//go:embed assets/neo_note.css
var neoNoteCSS string

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

//go:embed assets/drawer.js
var drawerScript string

//go:embed assets/agent_chat.css
var agentChatCSS string

//go:embed assets/agent_chat.html
var agentChatMainHTML string

//go:embed assets/agent_chat.js
var agentChatJS string

//go:embed assets/home.css
var homeCSS string

//go:embed assets/home.html
var homeMainHTML string

//go:embed assets/home.js
var homeJS string

//go:embed assets/images.css
var imagesCSS string

//go:embed assets/images.html
var imagesMainHTML string

//go:embed assets/images_grid.html
var imagesGridHTML string

//go:embed assets/images.js
var imagesJS string

//go:embed assets/library.css
var libraryCSS string

//go:embed assets/library.html
var libraryMainHTML string

//go:embed assets/library.js
var libraryJS string

//go:embed assets/library_notes_frag.html
var libraryNotesFragHTML string

//go:embed assets/library_tag_knowledge.html
var libraryTagKnowledgeHTML string

//go:embed assets/library_focus_oob.html
var libraryFocusOOBHTML string

//go:embed assets/library_unfocus_oob.html
var libraryUnfocusOOBHTML string

//go:embed assets/library_refresh.html
var libraryRefreshHTML string

//go:embed assets/dir.css
var dirCSS string

//go:embed assets/dir.html
var dirMainHTML string

//go:embed assets/dir.js
var dirJS string

//go:embed assets/folders.css
var foldersCSS string

//go:embed assets/folders.html
var foldersMainHTML string

//go:embed assets/folders.js
var foldersJS string

//go:embed assets/dir_notes_frag.html
var dirNotesFragHTML string

//go:embed assets/dir_refresh.html
var dirRefreshHTML string

//go:embed assets/shorts.css
var shortsCSS string

//go:embed assets/shorts.html
var shortsMainHTML string

//go:embed assets/shorts.js
var shortsJS string

//go:embed assets/shorts_day.html
var shortsDayHTML string

//go:embed assets/shorts_calendar.html
var shortsCalendarHTML string

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

//go:embed assets/graph.js
var graphJS string
