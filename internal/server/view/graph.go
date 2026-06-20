package view

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/hardhacker/vaultr/internal/storage"
)

// ─── API data types ────────────────────────────────────────────────────────────

type graphAPINode struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	Path       string   `json:"path"`
	Tags       []string `json:"tags,omitempty"`
	EntityType string   `json:"entity_type,omitempty"`
}

type graphAPIEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type graphAPIData struct {
	Nodes []graphAPINode `json:"nodes"`
	Edges []graphAPIEdge `json:"edges"`
}

// ─── page template data ────────────────────────────────────────────────────────

type graphPageData struct {
	IndexNotes []noteItem
}

// ─── handlers ──────────────────────────────────────────────────────────────────

// KnowledgeGraph handles GET /graph — the full knowledge graph page.
func (vh *ViewHandler) KnowledgeGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	indexNotes := vh.listIndexItems()
	data := graphPageData{IndexNotes: indexNotes}

	var buf bytes.Buffer
	if err := graphPageTemplate.Execute(&buf, data); err != nil {
		http.Error(w, fmt.Sprintf("render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

// KnowledgeGraphRebuild handles POST /api/graph/rebuild
// Rebuilds the knowledge_links table from the current vault filesystem.
// Safe to call at any time — only touches knowledge_links.
func (vh *ViewHandler) KnowledgeGraphRebuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	knowledgeDir := ""
	if vh.cfg != nil {
		knowledgeDir = vh.cfg.Vault.KnowledgeDir
	}
	if err := vh.vault.BackfillKnowledgeLinks(knowledgeDir); err != nil {
		http.Error(w, "rebuild: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

// KnowledgeGraphData handles GET /api/graph/data?index=PATH
// Returns JSON {nodes, edges} for the knowledge graph.
// When index is provided, only nodes/edges in that index are returned.
func (vh *ViewHandler) KnowledgeGraphData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	indexParam := strings.TrimSpace(r.URL.Query().Get("index"))

	var notes []storage.Note
	var edges []storage.KnowledgeEdge

	if indexParam != "" {
		idxPath, ok := storage.ParsePath(indexParam)
		if !ok {
			http.Error(w, `index must be absolute (start with "/")`, http.StatusBadRequest)
			return
		}
		knowledgePaths, err := vh.vault.GetIndexDeps(idxPath)
		if err != nil {
			http.Error(w, "get index deps: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if len(knowledgePaths) > 0 {
			notes, err = vh.vault.GetNotesByPaths(knowledgePaths)
			if err != nil {
				http.Error(w, "get notes: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
		edges, err = vh.vault.GetKnowledgeLinksForIndex(idxPath)
		if err != nil {
			http.Error(w, "get edges: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		var err error
		notes, err = vh.vault.ListAllNotes(storage.ListOptions{
			OnlyKinds: []storage.Kind{storage.KindKnowledge},
		})
		if err != nil {
			http.Error(w, "list knowledge notes: "+err.Error(), http.StatusInternalServerError)
			return
		}
		edges, err = vh.vault.GetAllKnowledgeLinks()
		if err != nil {
			http.Error(w, "get edges: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	entityTypeByPath := make(map[string]string, len(edges))
	for _, e := range edges {
		if e.SourceEntityType != "" {
			entityTypeByPath[e.Source.String()] = e.SourceEntityType
		}
	}

	apiNodes := make([]graphAPINode, 0, len(notes))
	for _, n := range notes {
		label := n.Title
		if label == "" {
			label = strings.TrimSuffix(n.Name, ".md")
		}
		// Prefer entity_type from knowledge_links (set at build time from frontmatter).
		// Fall back to tags[0] which compile plugin also sets to entity_type.
		et := entityTypeByPath[n.PathString()]
		if et == "" && len(n.Tags) > 0 {
			et = n.Tags[0]
		}
		apiNodes = append(apiNodes, graphAPINode{
			ID:         n.PathString(),
			Label:      label,
			Path:       n.PathString(),
			Tags:       n.Tags,
			EntityType: et,
		})
	}

	apiEdges := make([]graphAPIEdge, 0, len(edges))
	for _, e := range edges {
		if e.Source == e.Target {
			continue // self-loops store entity_type for leaf nodes; exclude from graph edges
		}
		apiEdges = append(apiEdges, graphAPIEdge{
			Source: e.Source.String(),
			Target: e.Target.String(),
		})
	}

	result := graphAPIData{Nodes: apiNodes, Edges: apiEdges}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "encode: "+err.Error(), http.StatusInternalServerError)
	}
}

// ─── template ──────────────────────────────────────────────────────────────────

var graphPageHTML = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
` + headHTML(headOpts{title: "Knowledge Graph — Vaultr", withFonts: true, withTW: true, withAlpine: true, withHTMX: true}) + `
  <script src="/static/vendor/cytoscape.min.js"></script>
  <script src="/static/vendor/layout-base.js"></script>
  <script src="/static/vendor/cose-base.js"></script>
  <script src="/static/vendor/cytoscape-fcose.min.js"></script>
  <style>
    :root {` + appTokensDark + `
      --cnt-bg:rgba(244,244,245,0.07); --cnt-tx:rgba(244,244,245,0.38);
    }
    html[data-theme="light"] {` + appTokensLight + `
      --cnt-bg:rgba(17,17,17,0.07); --cnt-tx:rgba(17,17,17,0.42);
    }
` + navCSS + pixelCSS + topbarCSS + drawerCSS + noteSharedCSS + noteEditorCSS + searchOverlayStyles + confirmDialogCSS + shortDialogCSS + settingsModalCSS + graphCSS + `
  </style>
</head>
<body x-data="graphCtrl()" style="margin:0;padding:0;overflow:hidden">
` + searchOnlyOverlayHTML + confirmDialogHTML + shortDialogHTML + settingsModalHTML() + `

  <header class="lib-topbar">
    <div class="lib-topbar-left">
      <button type="button" class="lib-back-btn" title="Back" onclick="window.history.length > 1 ? history.back() : (location.href='/home')">` + topbarIconBack + `<span class="lib-back-label">back</span></button>
    </div>
    <div class="lib-topbar-spacer"></div>
` + topbarActionsHTML("refresh()", "Refresh graph", "loading && 'spinning'", "{ mode: 'knowledge' }") + `
  </header>

  <div class="graph-body">
` + navHTML("graph") + `

    <div class="graph-content">
      <!-- ── Columns — index sidebar + canvas ─────────────────── -->
      <div class="graph-columns">
        <div class="graph-index-col">
          <div class="graph-index-head">
            <span class="graph-index-head-title">Index</span>
            {{if .IndexNotes}}<span class="graph-index-head-count">{{len .IndexNotes}}</span>{{end}}
          </div>
          <div class="graph-index-items">
            {{range .IndexNotes}}
            <div class="graph-index-item"
                 :class="{active: indexPath === '{{.Path}}'}"
                 @click="selectIndex('{{.Path}}')">
              <span class="graph-index-item-name">{{if .Title}}{{.Title}}{{else}}{{.Name}}{{end}}</span>
              {{if .DepCount}}<span class="graph-index-item-count">{{.DepCount}}</span>{{end}}
            </div>
            {{end}}
            {{if not .IndexNotes}}
            <div class="graph-index-empty">No index notes</div>
            {{end}}
          </div>
        </div>

        <div class="graph-main">
          <div style="position:relative;flex:1;display:flex;flex-direction:column;overflow:hidden">
            <div id="graph-canvas" style="flex:1;width:100%"></div>

            <!-- Zoom controls -->
            <div class="graph-zoom-controls">
              <button type="button" class="graph-zoom-btn" title="Zoom in" @click="zoomIn()">
                <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" d="M12 5v14M5 12h14"/>
                </svg>
              </button>
              <div class="graph-zoom-divider"></div>
              <button type="button" class="graph-zoom-btn" title="Zoom out" @click="zoomOut()">
                <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" d="M5 12h14"/>
                </svg>
              </button>
              <div class="graph-zoom-divider"></div>
              <button type="button" class="graph-zoom-btn" title="Fit all nodes" @click="zoomFit()">
                <svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M8 3H5a2 2 0 0 0-2 2v3"/>
                  <path stroke-linecap="round" stroke-linejoin="round" d="M21 8V5a2 2 0 0 0-2-2h-3"/>
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3 16v3a2 2 0 0 0 2 2h3"/>
                  <path stroke-linecap="round" stroke-linejoin="round" d="M16 21h3a2 2 0 0 0 2-2v-3"/>
                </svg>
              </button>
            </div>

            <!-- Node info panel -->
            <div class="graph-node-panel" :class="{ open: !!nodePanel }">
              <div class="graph-node-panel-header">
                <div class="graph-node-panel-meta">
                  <span class="graph-node-panel-type"
                        x-show="nodePanel && nodePanel.entityType"
                        x-text="nodePanel ? nodePanel.entityType : ''"></span>
                  <span class="graph-node-panel-edges"
                        x-text="nodePanel ? (nodePanel.edgeCount + (nodePanel.edgeCount === 1 ? ' connection' : ' connections')) : ''"></span>
                </div>
                <button type="button" class="graph-node-panel-close" @click="closeNodePanel()" title="Close">
                  <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" d="M18 6 6 18"/>
                    <path stroke-linecap="round" d="m6 6 12 12"/>
                  </svg>
                </button>
              </div>
              <div class="graph-node-panel-title" x-text="nodePanel ? nodePanel.label : ''"></div>
              <div class="graph-node-panel-body">
                <template x-if="nodePanel">
                  <div class="graph-node-panel-cards">
                    <div class="graph-node-section-label">Knowledge Unit</div>
                    <button type="button" class="graph-node-card is-featured"
                            @click="openNodeInDrawer(nodePanel.path, nodePanel.label, nodePanel.entityType)">
                      <span class="graph-node-card-name" x-text="nodePanel.label"></span>
                      <span class="graph-node-card-path" x-text="nodePanel.path"></span>
                    </button>
                    <template x-if="nodePanel.connected.length > 0">
                      <div>
                        <div class="graph-node-section-label" style="margin-top:0.35rem">Connected</div>
                        <div style="display:flex;flex-direction:column;gap:0.35rem">
                          <template x-for="cn in nodePanel.connected" :key="cn.path">
                            <button type="button" class="graph-node-card"
                                    @click="openNodeInDrawer(cn.path, cn.label, cn.entityType)">
                              <span class="graph-node-card-name" x-text="cn.label"></span>
                              <span class="graph-node-card-path" x-text="cn.path"></span>
                            </button>
                          </template>
                        </div>
                      </div>
                    </template>
                  </div>
                </template>
              </div>
            </div>

            <div class="graph-loading" x-show="loading" x-cloak>
              <span>Loading graph…</span>
            </div>

            <div class="graph-zero" x-show="!loading && empty" x-cloak>
              <div class="graph-zero-icon">
                <svg fill="none" stroke="currentColor" stroke-width="1.25" viewBox="0 0 48 48">
                  <circle cx="24" cy="12" r="4.5"/>
                  <circle cx="10" cy="36" r="4.5"/>
                  <circle cx="38" cy="36" r="4.5"/>
                  <line x1="24" y1="16.5" x2="10.8" y2="31.7" stroke-linecap="round"/>
                  <line x1="24" y1="16.5" x2="37.2" y2="31.7" stroke-linecap="round"/>
                  <line x1="14.5" y1="36" x2="33.5" y2="36" stroke-linecap="round"/>
                </svg>
              </div>
              <div class="graph-zero-title">No knowledge notes yet</div>
              <div class="graph-zero-desc">Knowledge notes and their connections will appear here once the compile agent has run on your raw notes.</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

` + drawerHTML + `
  <div id="graph-tooltip" class="graph-tooltip"></div>

  <script>
  document.addEventListener('alpine:init', () => {
` + themeStoreScript + `
  });

` + keysJS + drawerScript + searchOverlayScript + confirmDialogJS + shortDialogJS + settingsCtrlJS + `

  ` + graphJS + `
  </script>
` + noteSharedJS + `
</body>
</html>
`

var graphPageTemplate = template.Must(template.New("graph").Parse(graphPageHTML))
