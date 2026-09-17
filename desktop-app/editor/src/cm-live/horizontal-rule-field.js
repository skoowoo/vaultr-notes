// Thematic breaks need a StateField: CM6 rejects block:true decorations from
// a ViewPlugin (live-preview.js). HorizontalRuleWidget needs block:true for
// a full-width <hr>. Single-line only — multi-line block replace corrupts height-map.
import { StateField } from '@codemirror/state';
import { Decoration, EditorView } from '@codemirror/view';
import { syntaxTree } from '@codemirror/language';
import { selectionTouchesLine } from './selection.js';
import { HorizontalRuleWidget } from './widgets.js';

function buildDecorations(state) {
  const doc = state.doc;
  const ranges = [];
  syntaxTree(state).iterate({
    enter(node) {
      if (node.name !== 'HorizontalRule') return;
      const line = doc.lineAt(node.from);
      if (selectionTouchesLine(state, line)) return;
      const followedByBlank = line.number < doc.lines && doc.line(line.number + 1).text.trim() === '';
      // Include trailing newline or CM6 leaves a phantom empty .cm-line after <hr>.
      const to = line.to < doc.length ? line.to + 1 : line.to;
      ranges.push(
        Decoration.replace({
          widget: new HorizontalRuleWidget(followedByBlank),
          block: true,
          inclusiveEnd: false,
        }).range(node.from, to)
      );
    },
  });
  return Decoration.set(ranges, true);
}

export function horizontalRuleField() {
  return StateField.define({
    create(state) {
      return buildDecorations(state);
    },
    update(value, tr) {
      if (!tr.docChanged && !tr.selection) return value;
      return buildDecorations(tr.state);
    },
    provide: (f) => EditorView.decorations.from(f),
  });
}
