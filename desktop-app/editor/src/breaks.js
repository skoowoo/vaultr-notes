import { $remark } from '@milkdown/utils';
import { KeymapReady, keymapCtx } from '@milkdown/core';
import { visit } from 'unist-util-visit';
import { Fragment, Slice } from 'prosemirror-model';
import { Selection } from 'prosemirror-state';

// ── Remark transformer ────────────────────────────────────────────────────────
// Milkdown's built-in remarkLineBreak converts single \n to {type:"break",
// data:{isInline:true}}, which renders as a space. Flip isInline to false so
// all soft-break newlines render as <br> instead.
function remarkSoftBreakToHard() {
  return (tree) => {
    visit(tree, 'break', (node) => {
      if (node.data) node.data.isInline = false;
      else node.data = { isInline: false };
    });
  };
}
const remarkHardBreaksPlugin = $remark('remarkHardBreaks', () => remarkSoftBreakToHard);

// Serializer: non-inline break → plain \n (not CommonMark "  \n")
function remarkBreakSerialize() {
  const data = this.data();
  const exts = data.toMarkdownExtensions || (data.toMarkdownExtensions = []);
  exts.push({ handlers: { break: () => '\n' } });
}
const remarkBreakSerializePlugin = $remark('remarkBreakSerialize', () => remarkBreakSerialize);

// ── Fix: nested task list Enter ───────────────────────────────────────────────
// prosemirror-schema-list's splitListItem uses createAndFill() (no attrs) when
// lifting an empty nested item, so checked defaults to null and the new outer
// item renders as a plain list item. This handler runs first (priority 150) and
// replicates that branch with { checked: false } to preserve task-item type.
function enterInNestedTaskList(state, dispatch) {
  if (!state.selection.empty) return false;
  const { $from } = state.selection;

  if ($from.parent.content.size !== 0) return false;
  if ($from.depth < 4) return false;

  const listItem = $from.node(-1);
  if (listItem.type.name !== 'list_item') return false;
  if (listItem.attrs.checked === null || listItem.attrs.checked === undefined) return false;
  if ($from.node(-1).childCount !== $from.indexAfter(-1)) return false;

  // Mirrors splitListItem's nested-empty guard exactly
  if ($from.depth === 3) return false;
  if ($from.node(-3).type.name !== 'list_item') return false;
  if ($from.index(-2) !== $from.node(-2).childCount - 1) return false;

  if (dispatch) {
    const listItemType = listItem.type;
    const depthBefore = $from.index(-1) ? 1 : $from.index(-2) ? 2 : 3;
    let wrap = Fragment.empty;
    for (let d = $from.depth - depthBefore; d >= $from.depth - 3; d--) {
      wrap = Fragment.from($from.node(d).copy(wrap));
    }
    const depthAfter = $from.indexAfter(-1) < $from.node(-2).childCount ? 1
      : $from.indexAfter(-2) < $from.node(-3).childCount ? 2 : 3;

    const newItem = listItemType.createAndFill({ checked: false });
    if (!newItem) return false;
    wrap = wrap.append(Fragment.from(newItem));

    const start = $from.before($from.depth - (depthBefore - 1));
    let tr = state.tr.replace(start, $from.after(-depthAfter), new Slice(wrap, 4 - depthBefore, 0));

    let sel = -1;
    tr.doc.nodesBetween(start, tr.doc.content.size, (node, pos) => {
      if (sel > -1) return false;
      if (node.isTextblock && node.content.size === 0) sel = pos + 1;
    });
    if (sel > -1) tr.setSelection(Selection.near(tr.doc.resolve(sel)));
    dispatch(tr.scrollIntoView());
  }
  return true;
}

// ── Keymap ────────────────────────────────────────────────────────────────────
const breaksKeymapPlugin = (ctx) => async () => {
  await ctx.wait(KeymapReady);
  const km = ctx.get(keymapCtx);
  km.add({ key: 'Enter', onRun: () => enterInNestedTaskList, priority: 150 });
};

export const breaksPlugin = [
  remarkHardBreaksPlugin,
  remarkBreakSerializePlugin,
  breaksKeymapPlugin,
];
