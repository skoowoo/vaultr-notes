package view

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"

)

type folderItem struct {
	Dir      string
	Label    string
	Count    int
	IsSystem bool
}

type foldersPageData struct {
	Folders     []folderItem
	TotalCount  int
	SystemCount int
}

func (vh *ViewHandler) Folders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dirs, err := vh.vault.ListAllDirs()
	if err != nil {
		http.Error(w, "folders: list dirs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	items := make([]folderItem, 0, len(dirs))
	systemCount := 0
	for _, d := range dirs {
		isSystem := dirHasUnderscoreView(d.Dir)
		if isSystem {
			systemCount++
		}
		label := d.Dir
		if d.Dir != "/" {
			label = strings.TrimPrefix(d.Dir, "/")
		}
		items = append(items, folderItem{
			Dir:      d.Dir,
			Label:    label,
			Count:    d.Count,
			IsSystem: isSystem,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].IsSystem != items[j].IsSystem {
			return !items[i].IsSystem
		}
		return false
	})

	data := foldersPageData{
		Folders:     items,
		TotalCount:  len(items),
		SystemCount: systemCount,
	}

	var buf bytes.Buffer
	if err := foldersPageTemplate.Execute(&buf, data); err != nil {
		http.Error(w, fmt.Sprintf("render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

var foldersTemplateFuncs = template.FuncMap{
	"folderColorIdx": func(dir string) int {
		sum := 0
		for _, c := range dir {
			sum += int(c)
		}
		return sum % 4
	},
}

var foldersPageHTML = `<!DOCTYPE html>
<html lang="en" data-theme="neo">
` + headHTML(headOpts{title: "Folders — Vaultr", withFonts: true, withTW: true, withAlpine: true, withHTMX: true}) + `  <style>
` + appTokensCSS + `
` + navCSS + neoCSS + topbarCSS + foldersCSS + drawerCSS + noteSharedCSS + noteEditorCSS + searchOverlayStyles + confirmDialogCSS + shortDialogCSS + settingsModalCSS + `
  </style>
  <script>
  /* Apply hero background synchronously before first paint to avoid flash */
  (function(){
    var d=localStorage.getItem('vaultr-hero-bg');
    if(!d) return;
    var y=parseFloat(localStorage.getItem('vaultr-hero-bg-y'))||0;
    var s=document.createElement('style');
    s.id='hero-bg-preload';
    s.textContent='.home-hero-wrapper{background-image:url('+d+');background-size:100% auto;background-repeat:no-repeat;background-position:center '+y+'px}';
    document.head.appendChild(s);
  })();
  </script>
</head>
<body x-data="foldersCtrl()">
` + searchOnlyOverlayHTML + confirmDialogHTML + shortDialogHTML + settingsModalHTML() + `
  <header class="lib-topbar">
    <div class="lib-topbar-left">
      <a href="/home" class="lib-back-btn" title="Back to Home">` + topbarIconBack + `<span class="lib-back-label">back</span></a>
    </div>
    <div class="lib-topbar-spacer"></div>
` + topbarActionsHTML("window.location.reload()", "Refresh", "", "") + `
  </header>

  <div class="lib-body">
` + navHTML("folders") + foldersMainHTML + `
  </div>

` + drawerHTML + `

  <script>
  document.addEventListener('alpine:init', () => {
` + alpineStoresScript + `
  });

` + keysJS + drawerScript + searchOverlayScript + confirmDialogJS + shortDialogJS + settingsCtrlJS + foldersJS + `
  </script>
` + noteSharedJS + `
</body>
</html>`

var foldersPageTemplate = template.Must(template.New("folders").Funcs(foldersTemplateFuncs).Parse(foldersPageHTML))
