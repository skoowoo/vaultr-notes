// Whole-frontmatter collapse, toggled by the "Metadata" header row.
//
// The header needs block:true (full-width row above the "---" line) and CM6
// rejects block:true decorations from a ViewPlugin (see
// horizontal-rule-field.js's comment) — livePreviewPlugin (live-preview.js)
// is a ViewPlugin, so the header can't be produced by decorateFrontmatter.
// It gets its own StateField instead, same pattern as horizontalRuleField.
//
// The collapsed flag itself is a separate, plain StateField (no decorations
// of its own) so both this file's header field and decorateFrontmatter
// (decorators.js) read the same persisted value.
import { StateField, StateEffect } from '@codemirror/state';
import { Decoration, EditorView } from '@codemirror/view';
import { FrontmatterHeaderWidget } from './widgets.js';
import { findFrontmatterNode } from './frontmatter-syntax.js';
import { selectionTouchesRange } from './selection.js';

export const setFrontmatterCollapsed = StateEffect.define();

export const frontmatterCollapseField = StateField.define({
  create: () => false,
  update(value, tr) {
    for (const e of tr.effects) {
      if (e.is(setFrontmatterCollapsed)) value = e.value;
    }
    return value;
  },
});

function buildHeaderDecorations(state, options) {
  // Frontmatter, if present, is always the tree's first top-level node
  // (frontmatter-syntax.js) — an O(1) check, not a full-tree walk.
  const node = findFrontmatterNode(state);
  if (!node) return Decoration.none;
  const collapsed = state.field(frontmatterCollapseField, false);
  const onEditFrontmatter = options && options.onEditFrontmatter;
  const from = node.from;
  const to = node.to;
  // Edit pencil is hidden until the caret is actually inside the block —
  // it's the one escape hatch into frontmatterReadOnly()'s otherwise
  // uneditable block, so surfacing it only on approach (not as permanent
  // chrome) keeps the collapsed header from looking editable at a glance.
  const showEdit = selectionTouchesRange(state, from, to);
  return Decoration.set([
    Decoration.widget({
      widget: new FrontmatterHeaderWidget(
        collapsed,
        (view) => {
          view.dispatch({ effects: setFrontmatterCollapsed.of(!collapsed) });
        },
        onEditFrontmatter ? (view) => onEditFrontmatter(view, from, to) : null,
        showEdit
      ),
      block: true,
      side: -1,
    }).range(from),
  ]);
}

/**
 * @param {object} [options]
 * @param {(view: import('@codemirror/view').EditorView, from: number, to: number) => void} [options.onEditFrontmatter]
 *   Called with the whole Frontmatter node's range (including the "---"
 *   delimiters) when the header's pencil button is clicked. App-level
 *   concern (drawer.js) — this file doesn't know about dialogs.
 */
export function frontmatterHeaderField(options) {
  return StateField.define({
    create(state) {
      return buildHeaderDecorations(state, options);
    },
    update(value, tr) {
      // tr.selection: caret crossing the frontmatter boundary toggles the
      // edit pencil's visibility. Cheap to check on every selection change
      // now — buildHeaderDecorations locates the node in O(1)
      // (findFrontmatterNode), not a full tree walk.
      if (!tr.docChanged && !tr.selection && !tr.effects.some((e) => e.is(setFrontmatterCollapsed))) return value;
      return buildHeaderDecorations(tr.state, options);
    },
    provide: (f) => EditorView.decorations.from(f),
  });
}
