// Thematic breaks need a StateField: CM6 rejects block:true decorations from
// a ViewPlugin (live-preview.js). HorizontalRuleWidget needs block:true for
// a full-width <hr>. Single-line only — multi-line block replace corrupts height-map.
import { StateField } from '@codemirror/state';
import { Decoration, EditorView } from '@codemirror/view';
import { syntaxTree } from '@codemirror/language';
import { selectionTouchesLine } from './selection.js';
import { HorizontalRuleWidget } from './widgets.js';

// Unlike Frontmatter, a HorizontalRule can appear anywhere — finding them
// still needs a full-tree walk. But that walk only has to happen when the
// doc actually changes; a caret move (selectionSet, fired on every arrow
// key) just needs to recheck each already-found position's own line against
// the new selection, not re-walk the whole tree to rediscover them.
function findHorizontalRulePositions(state) {
  const positions = [];
  syntaxTree(state).iterate({
    enter(node) {
      if (node.name === 'HorizontalRule') positions.push(node.from);
    },
  });
  return positions;
}

function buildDecorations(state, positions) {
  const doc = state.doc;
  const ranges = [];
  for (const from of positions) {
    const line = doc.lineAt(from);
    if (selectionTouchesLine(state, line)) continue;
    const followedByBlank = line.number < doc.lines && doc.line(line.number + 1).text.trim() === '';
    // Include trailing newline or CM6 leaves a phantom empty .cm-line after <hr>.
    const to = line.to < doc.length ? line.to + 1 : line.to;
    ranges.push(
      Decoration.replace({
        widget: new HorizontalRuleWidget(followedByBlank),
        block: true,
        inclusiveEnd: false,
      }).range(from, to)
    );
  }
  return Decoration.set(ranges, true);
}

export function horizontalRuleField() {
  return StateField.define({
    create(state) {
      const positions = findHorizontalRulePositions(state);
      return { positions, decorations: buildDecorations(state, positions) };
    },
    update(value, tr) {
      if (!tr.docChanged && !tr.selection) return value;
      const positions = tr.docChanged ? findHorizontalRulePositions(tr.state) : value.positions;
      return { positions, decorations: buildDecorations(tr.state, positions) };
    },
    provide: (f) => EditorView.decorations.from(f, (v) => v.decorations),
  });
}
