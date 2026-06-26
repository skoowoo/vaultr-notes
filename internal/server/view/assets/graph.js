// ── Knowledge Graph controller ───────────────────────────────────────────────
function graphCtrl() {
  return Object.assign(drawerCtrl(), {
    loading: true,
    empty: false,
    indexPath: '',
    cy: null,
    _tooltip: null,
    _focusedPath: '',
    nodePanel: null,

    init() {
      if (window.cytoscapeFcose && window.cytoscape) {
        try { cytoscape.use(cytoscapeFcose); } catch (_) { }
      }

      var saved = sessionStorage.getItem('vaultr-graph-index');
      if (saved) this.indexPath = saved;

      this._tooltip = document.getElementById('graph-tooltip');
      this.initDrawer();

      var self = this;

      // Override search hotkey: open in knowledge mode; no-op when already open.
      // Deferred so it runs after searchOverlay.init() (child component) which
      // also registers 'search' — last registration wins.
      setTimeout(function () {
        window.__vaultrHotkeys.register('search', 'k', function () {
          if (!window.__vaultrSearchOpen) {
            window.dispatchEvent(new CustomEvent('open-search', { detail: { mode: 'knowledge' } }));
          }
        });
      }, 0);

      // Intercept search result selection — focus node in graph instead of navigating.
      // Always return true to prevent navigation (ignore results not in the graph).
      window.handleSearchResultSelection = function (el) {
        var path = el && el.dataset.previewPath;
        if (!path) return true;
        if (!self.cy) return true;
        var node = self.cy.getElementById(path);
        if (!node || node.empty()) return true; // not in graph — ignore silently
        if (self._focusedPath === path) {
          self._clearFocus();
        } else {
          self._applyFocus(node);
        }
        return true;
      };

      this.loadGraph();
    },

    async loadGraph() {
      this.loading = true;
      this.empty = false;
      this._focusedPath = '';
      this.nodePanel = null;
      var url = '/api/graph/data';
      if (this.indexPath) url += '?index=' + encodeURIComponent(this.indexPath);
      try {
        var resp = await fetch(url);
        if (!resp.ok) { this.loading = false; return; }
        var data = await resp.json();
        this._renderGraph(data);
      } catch (e) {
        console.error('graph load:', e);
      }
      this.loading = false;
    },

    // ── Focus ─────────────────────────────────────────────────

    _applyFocus(node) {
      this._focusedPath = node.data('path');
      this.cy.stop();
      this.cy.elements().unselect();
      node.select();
      this.cy.elements().addClass('faded');
      node.removeClass('faded');
      var connected = node.connectedEdges();
      connected.removeClass('faded');
      connected.connectedNodes().removeClass('faded');
      this.cy.animate({
        center: { eles: node },
      }, { duration: 200 });

      // Build node panel data.
      var neighborNodes = connected.connectedNodes().filter(function (n) {
        return n.id() !== node.id();
      });
      var connectedData = [];
      neighborNodes.forEach(function (n) {
        connectedData.push({
          path: n.data('path'),
          label: n.data('label'),
          entityType: n.data('entityType') || '',
        });
      });
      this.nodePanel = {
        path: node.data('path'),
        label: node.data('label'),
        entityType: node.data('entityType') || '',
        edgeCount: connected.length,
        connected: connectedData,
      };
    },

    _clearFocus() {
      this._focusedPath = '';
      if (this.cy) {
        this.cy.stop();
        this.cy.elements().unselect();
        this.cy.elements().removeClass('faded');
      }
      this.nodePanel = null;
    },

    closeNodePanel() {
      this._clearFocus();
    },

    openNodeInDrawer(path, label, entityType) {
      var drawer = window.__vaultrDrawer;
      if (!drawer) return;
      drawer.openNoteInDrawer(path, label || path, true, false, false, false);
    },

    // ── Helpers ──────────────────────────────────────────────

    _hexToRgba(hex, alpha) {
      hex = hex.replace(/^#/, '');
      if (hex.length === 3) hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2];
      var r = parseInt(hex.slice(0, 2), 16);
      var g = parseInt(hex.slice(2, 4), 16);
      var b = parseInt(hex.slice(4, 6), 16);
      return 'rgba(' + r + ',' + g + ',' + b + ',' + alpha + ')';
    },

    _entityTypeColors: {
      'concept': '#facc15',
      'person': '#60a5fa',
      'product': '#34d399',
      'company': '#fb923c',
      'project': '#a78bfa',
      'topic': '#f472b6',
      'brand': '#f87171',
      'business-model': '#6366f1',
      'book': '#38bdf8',
      'tool': '#14b8a6',
      'framework': '#06b6d4',
      'technique': '#a855f7',
      'strategy': '#ef4444',
      'protocol': '#22d3ee',
      'product-platform': '#34d399',
      'startup': '#f59e0b',
      'role': '#84cc16',
      'market': '#fdba74',
      'opensource-project': '#22c55e',
      'service': '#2dd4bf',
      'event': '#f43f5e',
      'disease': '#dc2626',
      'community': '#60a5fa',
    },

    _tagPaletteColor(tag) {
      if (this._entityTypeColors[tag]) return this._entityTypeColors[tag];
      var palette = ['#60a5fa', '#f472b6', '#a78bfa', '#34d399', '#facc15', '#fb923c', '#22d3ee', '#f87171'];
      var h = 0;
      for (var i = 0; i < tag.length; i++) h = (Math.imul(31, h) + tag.charCodeAt(i)) | 0;
      return palette[Math.abs(h) % palette.length];
    },

    _tagColor(tag) {
      if (!tag) {
        var s = getComputedStyle(document.documentElement);
        return s.getPropertyValue('--card-bg').trim() || '#1a1a1a';
      }
      return this._hexToRgba(this._tagPaletteColor(tag), 0.30);
    },

    _tagBorder(tag) {
      if (!tag) {
        var s = getComputedStyle(document.documentElement);
        return s.getPropertyValue('--card-bd').trim() || 'rgba(244,244,245,0.09)';
      }
      return this._tagPaletteColor(tag);
    },

    // ── Render ───────────────────────────────────────────────

    _renderGraph(data) {
      var container = document.getElementById('graph-canvas');
      if (!container) return;

      var oldCy = this.cy; this.cy = null; if (oldCy) oldCy.destroy();
      this._focusedPath = '';

      if (!data.nodes || data.nodes.length === 0) {
        this.empty = true;
        return;
      }
      this.empty = false;

      var self = this;

      var css = getComputedStyle(document.documentElement);
      var nodeLabelColor = css.getPropertyValue('--fg').trim() || '#f4f4f5';
      var bgColor = css.getPropertyValue('--bg').trim() || '#0f0f0f';
      var accentColor = css.getPropertyValue('--accent').trim() || '#cc785c';
      var mutedHex = css.getPropertyValue('--muted').trim() || '#71717a';
      var edgeFallback = /^#[0-9a-f]{6}$/i.test(mutedHex)
        ? self._hexToRgba(mutedHex, 0.28)
        : 'rgba(113,113,122,0.28)';
      var selBorderColor = accentColor;

      var degreeMap = {};
      (data.nodes || []).forEach(function (n) { degreeMap[n.id] = 0; });
      (data.edges || []).forEach(function (e) {
        if (degreeMap[e.source] !== undefined) degreeMap[e.source]++;
        if (degreeMap[e.target] !== undefined) degreeMap[e.target]++;
      });
      function nodeSize(id) {
        var deg = degreeMap[id] || 0;
        return Math.round(Math.min(80, 22 + Math.log2(deg + 1) * 10));
      }

      var nodeEntityType = {};
      (data.nodes || []).forEach(function (n) { nodeEntityType[n.id] = n.entity_type || ''; });

      var elements = [];
      data.nodes.forEach(function (n) {
        var deg = degreeMap[n.id] || 0;
        elements.push({
          data: {
            id: n.id, label: n.label, path: n.path,
            entityType: n.entity_type || '', tags: n.tags || [],
            degree: deg, nodeSize: nodeSize(n.id),
          }
        });
      });
      (data.edges || []).forEach(function (e) {
        var et = nodeEntityType[e.source] || '';
        var borderHex = self._tagBorder(et);
        var edgeColor = /^#[0-9a-f]{6}$/i.test(borderHex)
          ? self._hexToRgba(borderHex, 0.20)
          : edgeFallback;
        elements.push({ data: { source: e.source, target: e.target, edgeColor: edgeColor } });
      });

      var nc = (data.nodes || []).length;
      // proof → default → draft as scale grows; draft avoids freezing at 1000+ nodes.
      var layoutQuality = nc <= 80 ? 'proof' : nc <= 400 ? 'default' : 'draft';
      var layoutNumIter = nc <= 80 ? 1500 : nc <= 300 ? 2000 : nc <= 600 ? 1200 : 800;
      // Repulsion capped at 28000: small graphs use lower floor, large graphs don't over-spread.
      var layoutRepulsion = Math.max(2500, Math.min(nc * 100, 28000));
      var layoutEdgeLen = Math.max(80, Math.min(350, 60 + nc * 2));
      var layoutGravity = Math.max(0.10, 0.30 - nc * 0.002);
      var layoutGravRange = Math.max(3.5, Math.min(8.0, 3.5 + nc * 0.03));
      var layoutTilePad = Math.max(20, Math.min(60, 10 + nc * 0.5));

      this.cy = cytoscape({
        container: container,
        elements: elements,
        style: [
          {
            selector: 'node',
            style: {
              'background-color': function (ele) { return self._tagColor(ele.data('entityType')); },
              'border-color': function (ele) { return self._tagBorder(ele.data('entityType')); },
              'border-width': 1.5,
              'label': 'data(label)',
              'color': nodeLabelColor,
              'font-size': function (ele) {
                return Math.max(10, Math.min(13, 10 + ele.data('degree') * 0.2)) + 'px';
              },
              'font-family': 'Inter,-apple-system,sans-serif',
              'text-valign': 'bottom',
              'text-halign': 'center',
              'text-margin-y': 5,
              'text-max-width': '120px',
              'text-wrap': 'ellipsis',
              'min-zoomed-font-size': 12,
              'text-background-opacity': 0,
              'text-shadow-blur': 6,
              'text-shadow-color': bgColor,
              'text-shadow-opacity': 0.9,
              'text-shadow-offset-x': 0,
              'text-shadow-offset-y': 0,
              'width': 'data(nodeSize)',
              'height': 'data(nodeSize)',
              'z-index': 10,
              'cursor': 'pointer',
            }
          },
          {
            selector: 'node:selected',
            style: {
              'border-color': '#000000',
              'border-width': 3.5,
              'z-index': 20,
            }
          },
          {
            selector: 'node.faded',
            style: { 'opacity': 0.18 }
          },
          {
            selector: 'edge',
            style: {
              'width': 1.2,
              'line-color': 'data(edgeColor)',
              'target-arrow-color': 'data(edgeColor)',
              'target-arrow-shape': 'triangle',
              'curve-style': 'bezier',
              'arrow-scale': 0.85,
              'opacity': 0.85,
              'z-index': 1,
            }
          },
          {
            selector: 'edge.faded',
            style: { 'opacity': 0.05 }
          },
        ],
        layout: {
          name: 'fcose',
          quality: layoutQuality,
          randomize: true,
          animate: true,
          animationDuration: 400,
          animationEasing: 'ease-out',
          fit: true,
          padding: 48,
          nodeDimensionsIncludeLabels: true,
          uniformNodeDimensions: false,
          packComponents: true,
          step: 'all',
          gravity: layoutGravity,
          gravityRange: layoutGravRange,
          initialEnergyOnIncremental: 0.5,
          nodeRepulsion: layoutRepulsion,
          idealEdgeLength: layoutEdgeLen,
          edgeElasticity: 0.45,
          nestingFactor: 0.1,
          numIter: layoutNumIter,
          tile: true,
          tilingPaddingVertical: layoutTilePad,
          tilingPaddingHorizontal: layoutTilePad,
          gravityCompound: 1.0,
          gravityRangeCompound: 1.5,
        },
        wheelSensitivity: 0.3,
        minZoom: 0.05,
        maxZoom: 4,
      });

      // ── Interactions ────────────────────────────────────────

      // Click node: toggle focus — no drawer.
      this.cy.on('tap', 'node', function (evt) {
        var node = evt.target;
        var path = node.data('path');
        if (!path) return;
        if (self._focusedPath === path) {
          self._clearFocus();
          // cytoscape re-selects the tapped node after tap handlers run,
          // overriding the unselect() inside _clearFocus. defer to win the race.
          setTimeout(function () { node.unselect(); }, 0);
        } else {
          self._applyFocus(node);
        }
      });

      // Click background: clear focus.
      this.cy.on('tap', function (evt) {
        if (evt.target === self.cy) {
          self._clearFocus();
        }
      });

      // Tooltip (hover label only — no fading).
      var tooltip = this._tooltip;
      if (tooltip) {
        this.cy.on('mouseover', 'node', function (evt) {
          tooltip.textContent = evt.target.data('label') || '';
          tooltip.classList.add('visible');
        });
        this.cy.on('mousemove', 'node', function (evt) {
          tooltip.style.left = (evt.originalEvent.clientX + 14) + 'px';
          tooltip.style.top = (evt.originalEvent.clientY - 8) + 'px';
        });
        this.cy.on('mouseout', 'node', function () {
          tooltip.classList.remove('visible');
        });
      }
    },

    // ── Index selection ──────────────────────────────────────

    selectIndex(path) {
      var next = (this.indexPath === path) ? '' : path;
      this.indexPath = next;
      sessionStorage.setItem('vaultr-graph-index', next);
      this.loadGraph();
    },

    refresh() {
      this.loadGraph();
    },

    // ── Zoom controls ────────────────────────────────────────
    zoomIn() {
      if (!this.cy) return;
      var cx = this.cy.width() / 2, cy = this.cy.height() / 2;
      this.cy.zoom({ level: this.cy.zoom() * 1.3, renderedPosition: { x: cx, y: cy } });
    },

    zoomOut() {
      if (!this.cy) return;
      var cx = this.cy.width() / 2, cy = this.cy.height() / 2;
      this.cy.zoom({ level: this.cy.zoom() / 1.3, renderedPosition: { x: cx, y: cy } });
    },

    zoomFit() {
      if (!this.cy) return;
      this.cy.fit(undefined, 48);
    },
  });
}
