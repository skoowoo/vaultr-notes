// Viewport-scoped tree walk → decorators.js. Publishes decorations + atomicRanges.
import { ViewPlugin, Decoration, EditorView } from '@codemirror/view';
import { syntaxTree } from '@codemirror/language';
import { nodeDecorators } from './decorators.js';
import { setFrontmatterCollapsed } from './frontmatter-collapse.js';

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
    // The "Metadata" header toggle dispatches a bare effect — no doc change,
    // no selection change, so it wouldn't otherwise trigger a rebuild and
    // decorateFrontmatter's collapsed-block branch would never run.
    const collapseToggled = update.transactions.some((tr) => tr.effects.some((e) => e.is(setFrontmatterCollapsed)));
    if (update.docChanged || update.viewportChanged || update.selectionSet || collapseToggled) {
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
