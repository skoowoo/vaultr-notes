import { KeymapReady, keymapCtx } from '@milkdown/core';

// ── Fix: Tab / Shift-Tab defocus the editor ────────────────────────────────
// prosemirror-keymap only calls preventDefault() when a bound command
// returns true. Milkdown's list-item keymap binds Tab/Shift-Tab to
// sink/liftListItem, but those return false whenever they can't act —
// outside a list entirely, or already at the list's min/max depth. When
// that happens the keydown falls through to the browser's native Tab
// behavior, which moves focus off the editor to the next focusable element
// on the page. From the user's perspective: press Tab once too many times
// while indenting a list item, or press it anywhere outside a list, and the
// cursor vanishes — they have to click back into the editor to keep typing.
//
// Registered at a low priority (10, below the unspecified default of 50)
// so list indent/outdent — or any future Tab binding, e.g. a code-block
// indent — still gets first chance. This only swallows the key when
// nothing else handled it, keeping focus in the editor with no visible
// side effect.
function tabNoop() {
  return true;
}

export const tabGuardPlugin = (ctx) => async () => {
  await ctx.wait(KeymapReady);
  const km = ctx.get(keymapCtx);
  km.add({ key: 'Tab', onRun: () => tabNoop, priority: 10 });
  km.add({ key: 'Shift-Tab', onRun: () => tabNoop, priority: 10 });
};
