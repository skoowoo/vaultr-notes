package view

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// AgentChat handles GET /agent — the mate chat page.
func (vh *ViewHandler) AgentChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var buf bytes.Buffer
	if err := agentChatTemplate.Execute(&buf, nil); err != nil {
		http.Error(w, fmt.Sprintf("render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

// ── template ──────────────────────────────────────────────────────────────────

var agentChatHTML = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
` + headHTML(headOpts{title: "Mates — Vaultr", withFonts: true, withTW: true, withAlpine: true, withHTMX: true}) + `  <style>
    :root{` + appTokensDark + `}
    html[data-theme="light"] {` + appTokensLight + `}
` + navCSS + pixelCSS + topbarCSS + agentChatCSS + searchOverlayStyles + confirmDialogCSS + shortDialogCSS + drawerCSS + noteSharedCSS + noteEditorCSS + settingsModalCSS + `
  </style>
  <script src="/static/vendor/marked.min.js"></script>
  <script src="/static/vendor/dompurify.min.js"></script>
  <script>
  if (typeof marked !== 'undefined') { marked.setOptions({ gfm: true, breaks: true }); }
  </script>
</head>
<body x-data="agentChatCtrl()" @vaultr:insert-path.window="insertPath($event)">
` + searchOnlyOverlayHTML + confirmDialogHTML + shortDialogHTML + settingsModalHTML() + `

  <header class="lib-topbar">
    <div class="lib-topbar-spacer"></div>
` + topbarActionsHTML("refreshPage()", "Refresh", "", "") + `
  </header>

  <div class="chat-body">
` + navHTML("agent") + agentChatMainHTML + `

` + drawerHTML + `

  <script>
  document.addEventListener('alpine:init', () => {
` + themeStoreScript + `
  });

` + keysJS + drawerScript + searchOverlayScript + confirmDialogJS + shortDialogJS + settingsCtrlJS + agentChatJS + `
  </script>
` + noteSharedJS + `
</body>
</html>
`

var agentChatTemplate = template.Must(template.New("agentchat").Parse(agentChatHTML))
