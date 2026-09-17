// Viewport-scoped tree walk → decorators.js. Publishes decorations + atomicRanges.
import { ViewPlugin, Decoration, EditorView } from '@codemirror/view';
import { syntaxTree } from '@codemirror/language';
import { nodeDecorators } from './decorators.js';

class LivePreviewPlugin {
  constructor(view, options) {
    this.options = options || {};
    this.decorations = Decoration.none;
    this.atomicRanges = Decoration.none;
    // Survives selection/viewport rebuilds; cleared on doc change.
    this.tableCache = new Map();
    this.rebuild(view);
  }

  update(update) {
    if (update.docChanged) this.tableCache.clear();
    if (update.docChanged || update.viewportChanged || update.selectionSet) {
      this.rebuild(update.view);
    }
  }

  rebuild(view) {
    const decos = [];
    const atomics = [];
    const tree = syntaxTree(view.state);
    for (const { from, to } of view.visibleRanges) {
      tree.iterate({
        from,
        to,
        enter: (node) => {
          const decorate = nodeDecorators[node.name];
          if (decorate) decorate(node, view, decos, atomics, this.options, this.tableCache);
        },
      });
    }
    this.decorations = Decoration.set(decos, true);
    this.atomicRanges = Decoration.set(atomics, true);
  }
}

export const livePreviewPlugin = ViewPlugin.fromClass(LivePreviewPlugin, {
  decorations: (v) => v.decorations,
});

export function livePreviewAtomicRanges() {
  return EditorView.atomicRanges.of((view) => {
    const plugin = view.plugin(livePreviewPlugin);
    return plugin ? plugin.atomicRanges : Decoration.none;
  });
}
