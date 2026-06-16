import { $remark } from '@milkdown/utils';
import { KeymapReady, keymapCtx } from '@milkdown/core';
import { visit } from 'unist-util-visit';
import { splitBlock, liftEmptyBlock } from 'prosemirror-commands';

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

// ── Config ────────────────────────────────────────────────────────────────────
// Set once before editor creation via setBreaksConfig(). Read at keypress time
// from module-level vars — no per-keystroke localStorage/IPC overhead.
let _enterKey = 'newparagraph';    // 'hardbreak' | 'newparagraph'
let _shiftEnterKey = 'hardbreak';  // 'newparagraph' | 'hardbreak'

export function setBreaksConfig({ enterKey, shiftEnterKey } = {}) {
  if (enterKey) _enterKey = enterKey;
  if (shiftEnterKey) _shiftEnterKey = shiftEnterKey;
}

// ── Helpers ───────────────────────────────────────────────────────────────────
function isInParagraphOutsideList(state) {
  const { $from } = state.selection;
  if ($from.parent.type !== state.schema.nodes.paragraph) return false;
  for (let d = $from.depth - 1; d >= 0; d--) {
    if ($from.node(d).type.name === 'list_item') return false;
  }
  return true;
}

function insertHardBreak(state, dispatch) {
  const hardBreak = state.schema.nodes.hardbreak;
  if (!hardBreak) return false;
  if (dispatch) dispatch(state.tr.replaceSelectionWith(hardBreak.create()).scrollIntoView());
  return true;
}

// ── Commands (read module vars, not localStorage) ─────────────────────────────
function isInBlockquote(state) {
  const { $from } = state.selection;
  for (let d = $from.depth - 1; d >= 0; d--) {
    if ($from.node(d).type.name === 'blockquote') return true;
  }
  return false;
}

function enterCmd(state, dispatch, view) {
  // Empty paragraph inside blockquote → lift out (exit the blockquote)
  if (isInBlockquote(state) && state.selection.$from.parent.content.size === 0) {
    return liftEmptyBlock(state, dispatch);
  }
  if (!isInParagraphOutsideList(state)) return false;
  return _enterKey === 'hardbreak'
    ? insertHardBreak(state, dispatch)
    : splitBlock(state, dispatch, view);
}

function shiftEnterCmd(state, dispatch, view) {
  if (_shiftEnterKey === 'newparagraph') return splitBlock(state, dispatch, view);
  if (!isInParagraphOutsideList(state)) return false;
  return insertHardBreak(state, dispatch);
}

// ── Keymap (priority 100 > Milkdown default 50) ───────────────────────────────
const breaksKeymapPlugin = (ctx) => async () => {
  await ctx.wait(KeymapReady);
  const km = ctx.get(keymapCtx);
  km.add({ key: 'Enter',       onRun: () => enterCmd,       priority: 100 });
  km.add({ key: 'Shift-Enter', onRun: () => shiftEnterCmd,  priority: 100 });
};

export const breaksPlugin = [
  remarkHardBreaksPlugin,
  remarkBreakSerializePlugin,
  breaksKeymapPlugin,
];
