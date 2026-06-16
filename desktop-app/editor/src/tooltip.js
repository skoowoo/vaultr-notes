import { $prose } from '@milkdown/utils';
import { Plugin, PluginKey, TextSelection } from '@milkdown/prose/state';
import { lift, toggleMark as pmToggleMark, setBlockType, wrapIn } from '@milkdown/prose/commands';
import { wrapInList } from '@milkdown/prose/schema-list';
import { Zap, Check, Copy, X, RemoveFormatting, List, ListOrdered, TextQuote } from 'lucide';

const TOOLTIP_KEY = new PluginKey('vaultr-format-tooltip');

// ── Lucide SVG helper ─────────────────────────────────────────────────────────

function lucideSvg(iconData) {
  const children = iconData.map(([tag, attrs]) => {
    const attrStr = Object.entries(attrs).map(([k, v]) => `${k}="${v}"`).join(' ');
    return `<${tag} ${attrStr}/>`;
  }).join('');
  return (
    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ' +
    'stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">' +
    children + '</svg>'
  );
}

const SVG_ZAP          = lucideSvg(Zap);
const SVG_CHECK        = lucideSvg(Check);
const SVG_COPY         = lucideSvg(Copy);
const SVG_X            = lucideSvg(X);
const SVG_CLEAR        = lucideSvg(RemoveFormatting);
const SVG_LIST         = lucideSvg(List);
const SVG_LIST_ORDERED = lucideSvg(ListOrdered);
const SVG_QUOTE        = lucideSvg(TextQuote);

// ── Detection helpers ─────────────────────────────────────────────────────────

// Returns heading level at cursor, or null.
function selectionHeadingLevel(state) {
  const parent = state.selection.$from.parent;
  if (parent.type === state.schema.nodes.heading) return parent.attrs.level;
  return null;
}

// Returns 'bullet_list' | 'ordered_list' | null based on cursor ancestry.
function selectionListType(state) {
  const { $from } = state.selection;
  for (let d = $from.depth; d >= 0; d--) {
    const name = $from.node(d).type.name;
    if (name === 'bullet_list' || name === 'ordered_list') return name;
  }
  return null;
}

function isMarkActive(state, markName) {
  const { from, to } = state.selection;
  const mt = state.schema.marks[markName];
  return !!(mt && state.doc.rangeHasMark(from, to, mt));
}

// Returns true when block-level format buttons are relevant to show.
// Lenient: any one of these signals is enough.
function shouldShowBlockFormats(state) {
  // Already inside a block format → always show for toggle-off.
  if (selectionHeadingLevel(state) !== null) return true;
  if (selectionListType(state) !== null) return true;
  // selectionInBlockquote checked below after its definition.

  const { $from, $to } = state.selection;
  // Selection starts at the very beginning of a textblock.
  if ($from.parentOffset === 0) return true;
  // Selection ends at the very end of a textblock.
  if ($to.parentOffset === $to.parent.content.size) return true;
  // Selection spans more than one textblock.
  if (!$from.sameParent($to)) return true;
  return false;
}

// Returns true if cursor is inside a blockquote.
function selectionInBlockquote(state) {
  const { $from } = state.selection;
  for (let d = $from.depth; d >= 0; d--) {
    if ($from.node(d).type.name === 'blockquote') return true;
  }
  return false;
}

// Returns true if the selection has any toggle-group format active.
function hasToggleFormats(state) {
  if (selectionHeadingLevel(state) !== null) return true;
  if (selectionListType(state) !== null) return true;
  if (selectionInBlockquote(state)) return true;
  for (const m of ['strong', 'emphasis', 'strike_through', 'inlineCode']) {
    if (isMarkActive(state, m)) return true;
  }
  return false;
}

// Returns { kind, label } for the first clear-group format found, or null.
// Only call when hasToggleFormats is false.
function detectClearTarget(state) {
  const { $from } = state.selection;
  for (let d = $from.depth; d >= 0; d--) {
    if ($from.node(d).type.name === 'code_block') return { kind: 'code_block', label: 'Code' };
  }
  const linkMt = state.schema.marks.link;
  if (linkMt && state.doc.rangeHasMark(state.selection.from, state.selection.to, linkMt)) {
    return { kind: 'link', label: 'Link' };
  }
  return null;
}

// ── Toggle actions ────────────────────────────────────────────────────────────

function toggleHeading(view, level) {
  const { state } = view;
  const { from, to } = state.selection;
  if (selectionHeadingLevel(state) === level) {
    view.dispatch(state.tr.setBlockType(from, to, state.schema.nodes.paragraph));
  } else {
    view.dispatch(state.tr.setBlockType(from, to, state.schema.nodes.heading, { level }));
  }
  view.focus();
}

function applyToggleMark(view, markName) {
  const markType = view.state.schema.marks[markName];
  if (!markType) return;
  pmToggleMark(markType)(view.state, view.dispatch);
  view.focus();
}

function clearListBlock(view) {
  const { state } = view;
  const { $from, from, to } = state.selection;
  let listDepth = -1;
  for (let d = $from.depth; d >= 0; d--) {
    const nm = $from.node(d).type.name;
    if (nm === 'bullet_list' || nm === 'ordered_list') { listDepth = d; break; }
  }
  if (listDepth < 0) return;
  const listNode  = $from.node(listDepth);
  const listStart = $from.before(listDepth);
  const listEnd   = listStart + listNode.nodeSize;
  const paragraphs = [];
  listNode.forEach(item => {
    if (item.type.name === 'list_item') item.forEach(child => paragraphs.push(child));
  });
  let tr = state.tr.replaceWith(listStart, listEnd, paragraphs);
  const newAnchor = tr.mapping.map(from, -1);
  const newHead   = tr.mapping.map(to,   1);
  tr = tr.setSelection(TextSelection.create(tr.doc, newAnchor, newHead));
  view.dispatch(tr);
}

function toggleBlockquote(view) {
  const { state } = view;
  if (selectionInBlockquote(state)) {
    lift(state, view.dispatch);
  } else {
    wrapIn(state.schema.nodes.blockquote)(state, view.dispatch);
  }
  view.focus();
}

function toggleList(view, listTypeName) {
  const { state } = view;
  const listType = state.schema.nodes[listTypeName];
  if (!listType) return;
  const currentType = selectionListType(state);
  if (currentType === listTypeName) {
    clearListBlock(view);
  } else if (currentType) {
    // Swap list node type in-place.
    const { $from } = state.selection;
    for (let d = $from.depth; d >= 0; d--) {
      const nm = $from.node(d).type.name;
      if (nm === 'bullet_list' || nm === 'ordered_list') {
        view.dispatch(state.tr.setNodeMarkup($from.before(d), listType));
        break;
      }
    }
  } else {
    wrapInList(listType)(state, view.dispatch);
  }
  view.focus();
}

// ── Clear actions (code_block / blockquote / link only) ───────────────────────

function clearComplex(view, kind) {
  const { state } = view;
  const { from, to } = state.selection;

  if (kind === 'code_block') {
    view.dispatch(state.tr.setBlockType(from, to, state.schema.nodes.paragraph));
  } else if (kind === 'link') {
    const linkMt = state.schema.marks.link;
    // Expand "text" → "text url" then strip the link mark.
    const expansions = [];
    state.doc.nodesBetween(from, to, (node, pos) => {
      if (!node.isText) return;
      const linkMark = node.marks.find(m => m.type === linkMt);
      if (!linkMark) return;
      const cf   = Math.max(pos, from);
      const ct   = Math.min(pos + node.nodeSize, to);
      const text = node.text.slice(cf - pos, ct - pos);
      const href = linkMark.attrs.href || '';
      expansions.push({ from: cf, to: ct, text: href ? text + ' ' + href : text });
    });
    let tr = state.tr;
    let shift = 0;
    for (let i = expansions.length - 1; i >= 0; i--) {
      const exp = expansions[i];
      tr = tr.replaceWith(exp.from, exp.to, state.schema.text(exp.text));
      shift += exp.text.length - (exp.to - exp.from);
    }
    tr = tr.removeMark(from, to + shift, linkMt);
    view.dispatch(tr);
  }

  view.focus();
}

// ── Word count ────────────────────────────────────────────────────────────────

function countWords(text) {
  const t = text.trim();
  if (!t) return 0;
  try {
    const seg = new Intl.Segmenter(undefined, { granularity: 'word' });
    return [...seg.segment(t)].filter(s => s.isWordLike).length;
  } catch {
    // Fallback: CJK chars each count as one word, remainder split on whitespace.
    const cjk = (t.match(/[一-鿿぀-ヿ가-힯]/g) || []).length;
    const latin = (t.replace(/[一-鿿぀-ヿ가-힯]/g, ' ')
                    .match(/\S+/g) || []).length;
    return cjk + latin;
  }
}

// ── Tooltip positioning ───────────────────────────────────────────────────────

function positionTooltip(el, view) {
  const { from, to, head } = view.state.selection;
  const topCoords  = view.coordsAtPos(from);
  const botCoords  = view.coordsAtPos(to, -1);
  const headCoords = view.coordsAtPos(head, -1);
  const box        = el.getBoundingClientRect();
  const gap        = 8;

  // Always prefer above the selection; fall back to below only if no room.
  let top = topCoords.top - box.height - gap;
  if (top < gap) top = botCoords.bottom + gap;

  // Horizontal: center on the head (cursor end) position, clamped to viewport.
  let left = headCoords.left - box.width / 2;
  left = Math.max(gap, Math.min(left, window.innerWidth - box.width - gap));

  el.style.top  = top  + 'px';
  el.style.left = left + 'px';
}

// ── Plugin ────────────────────────────────────────────────────────────────────

const HEADING_LEVELS = [1, 2, 3, 4];

const INLINE_FMT_BTNS = [
  { mark: 'strong',         label: 'B',   title: 'Bold'          },
  { mark: 'emphasis',       label: 'I',   title: 'Italic'        },
  { mark: 'strike_through', label: 'S',   title: 'Strikethrough' },
  { mark: 'inlineCode',     label: '</>', title: 'Inline code'   },
];

export const tooltipPlugin = $prose(() => new Plugin({
  key: TOOLTIP_KEY,
  view() {
    const el = document.createElement('div');
    el.className = 'milkdown-tooltip';
    document.body.appendChild(el);

    let pmView  = null;
    let visible = false;

    function onMouseUp(e) {
      setTimeout(() => {
        if (!pmView || pmView.state.selection.empty) return;
        // Respond to interactions inside the PM editor OR its scroll container
        // (#drawer-edit-area). When content is short and the user releases the
        // mouse in the empty space below the text, e.target is the container
        // element rather than a node inside pmView.dom, so we accept both.
        const editArea = document.getElementById('drawer-edit-area');
        const inEditor = pmView.dom.contains(e.target) ||
                         (editArea && editArea.contains(e.target));
        if (!inEditor) return;
        // New mouse selection: always reposition from scratch.
        if (visible) { window.__vaultrEscPop?.('format-tooltip'); visible = false; }
        render(pmView);
      }, 0);
    }

    function onKeyUp() {
      if (!pmView || pmView.state.selection.empty) return;
      // Only respond when focus is inside the PM editor.
      if (!pmView.dom.contains(document.activeElement)) return;
      if (!visible) render(pmView);
    }

    document.addEventListener('mouseup', onMouseUp, true);
    document.addEventListener('keyup',   onKeyUp,   true);

    function hide(collapseSelection) {
      if (!visible) return;
      visible = false;
      el.style.display = 'none';
      window.__vaultrEscPop?.('format-tooltip');
      if (collapseSelection && pmView && !pmView.state.selection.empty) {
        const from = pmView.state.selection.from;
        pmView.dispatch(
          pmView.state.tr.setSelection(TextSelection.create(pmView.state.doc, from))
        );
        pmView.focus();
      }
    }

    function render(view) {
      const { state } = view;
      if (state.selection.empty) { hide(false); return; }

      const showClear = !hasToggleFormats(state);
      const clearTarget = showClear ? detectClearTarget(state) : null;

      el.innerHTML = '';

      if (showClear && clearTarget) {
        // ── Clear group: chip + clear button ──────────────────────────────────
        const chip = document.createElement('span');
        chip.className = 'mdt-chip';
        chip.textContent = clearTarget.label;
        el.appendChild(chip);

        const clearBtn = document.createElement('button');
        clearBtn.className = 'mdt-clear';
        clearBtn.title = 'Clear ' + clearTarget.label.toLowerCase();
        clearBtn.innerHTML = SVG_CLEAR;
        clearBtn.addEventListener('mousedown', e => {
          e.preventDefault();
          clearComplex(view, clearTarget.kind);
        });
        el.appendChild(clearBtn);

        const sep = document.createElement('span');
        sep.className = 'mdt-sep';
        el.appendChild(sep);
      } else {
        // ── Toggle group ──────────────────────────────────────────────────────
        const showBlock = shouldShowBlockFormats(state) || selectionInBlockquote(state);

        if (showBlock) {
          const headingLevel = selectionHeadingLevel(state);
          HEADING_LEVELS.forEach(level => {
            const btn = document.createElement('button');
            btn.className = 'mdt-fmt' + (headingLevel === level ? ' mdt-fmt--active' : '');
            btn.textContent = 'H' + level;
            btn.title = 'Heading ' + level;
            btn.addEventListener('mousedown', e => {
              e.preventDefault();
              toggleHeading(view, level);
            });
            el.appendChild(btn);
          });

          const quoteBtn = document.createElement('button');
          quoteBtn.className = 'mdt-fmt mdt-fmt--icon' + (selectionInBlockquote(state) ? ' mdt-fmt--active' : '');
          quoteBtn.title = 'Blockquote';
          quoteBtn.innerHTML = SVG_QUOTE;
          quoteBtn.addEventListener('mousedown', e => {
            e.preventDefault();
            toggleBlockquote(view);
          });
          el.appendChild(quoteBtn);

          const listType = selectionListType(state);

          const bulletBtn = document.createElement('button');
          bulletBtn.className = 'mdt-fmt mdt-fmt--icon' + (listType === 'bullet_list' ? ' mdt-fmt--active' : '');
          bulletBtn.title = 'Bullet list';
          bulletBtn.innerHTML = SVG_LIST;
          bulletBtn.addEventListener('mousedown', e => {
            e.preventDefault();
            toggleList(view, 'bullet_list');
          });
          el.appendChild(bulletBtn);

          const orderedBtn = document.createElement('button');
          orderedBtn.className = 'mdt-fmt mdt-fmt--icon' + (listType === 'ordered_list' ? ' mdt-fmt--active' : '');
          orderedBtn.title = 'Ordered list';
          orderedBtn.innerHTML = SVG_LIST_ORDERED;
          orderedBtn.addEventListener('mousedown', e => {
            e.preventDefault();
            toggleList(view, 'ordered_list');
          });
          el.appendChild(orderedBtn);

          const sep1 = document.createElement('span');
          sep1.className = 'mdt-sep';
          el.appendChild(sep1);
        }

        INLINE_FMT_BTNS.forEach(({ mark, label, title }) => {
          const btn = document.createElement('button');
          btn.className = 'mdt-fmt' + (isMarkActive(state, mark) ? ' mdt-fmt--active' : '');
          btn.textContent = label;
          btn.title = title;
          btn.addEventListener('mousedown', e => {
            e.preventDefault();
            applyToggleMark(view, mark);
          });
          el.appendChild(btn);
        });

        const sep2 = document.createElement('span');
        sep2.className = 'mdt-sep';
        el.appendChild(sep2);
      }

      // ── Word count ───────────────────────────────────────────────────────────
      const selText = state.doc.textBetween(state.selection.from, state.selection.to, ' ');
      const wc = countWords(selText);
      if (wc > 0) {
        const countEl = document.createElement('span');
        countEl.className = 'mdt-count';
        countEl.textContent = wc + 'w';
        el.appendChild(countEl);

        const sepW = document.createElement('span');
        sepW.className = 'mdt-sep';
        el.appendChild(sepW);
      }

      // ── Always: copy, save as short, dismiss ─────────────────────────────────
      const copyBtn = document.createElement('button');
      copyBtn.className = 'mdt-copy';
      copyBtn.title = 'Copy as Markdown';
      copyBtn.innerHTML = SVG_COPY;
      copyBtn.addEventListener('mousedown', e => {
        e.preventDefault();
        window.__vaultrCopySelectionAsMd?.();
      });
      el.appendChild(copyBtn);

      const shortBtn = document.createElement('button');
      shortBtn.className = 'mdt-short';
      shortBtn.title = 'Save as short note';
      shortBtn.innerHTML = SVG_ZAP;
      shortBtn.addEventListener('mousedown', async e => {
        e.preventDefault();
        const md = window.__vaultrGetSelectionMd?.();
        if (!md) return;
        const de   = window.__vaultrDE;
        const path = de && de.currentPath;
        let content = md.replace(/^#{1,6}\s+/gm, '').replace(/^>\s?/gm, '');
        if (path) {
          const stem = path.split('/').pop().replace(/\.md$/i, '');
          content += '\n\nSource: [[' + stem + ']]';
        }
        try {
          const resp = await fetch('/api/vault/shorts', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ content }),
          });
          if (!resp.ok) throw new Error(await resp.text());
          shortBtn.innerHTML = SVG_CHECK;
          shortBtn.classList.add('saved');
          if (window.__vaultrAfterVaultMutation) await window.__vaultrAfterVaultMutation();
          setTimeout(() => hide(true), 600);
        } catch (err) {
          if (window.showError) window.showError((err && err.message) || 'Save failed', 'Short note');
        }
      });
      el.appendChild(shortBtn);

      const xBtn = document.createElement('button');
      xBtn.className = 'mdt-x';
      xBtn.setAttribute('aria-label', 'Dismiss');
      xBtn.innerHTML = SVG_X;
      xBtn.addEventListener('mousedown', e => {
        e.preventDefault();
        hide(true);
      });
      el.appendChild(xBtn);

      if (!visible) {
        el.style.display = 'flex';
        positionTooltip(el, view);
        visible = true;
        window.__vaultrEscPush?.('format-tooltip', () => hide(true));
      }
    }

    return {
      update(view, prevState) {
        pmView = view;
        if (view.state.selection.empty) { hide(false); return; }
        // Never initiate a new show from state updates — only mouseup/keyup do that.
        if (!visible) return;
        const selSame = prevState && prevState.selection.eq(view.state.selection);
        const docSame = prevState && prevState.doc.eq(view.state.doc);
        if (selSame && docSame) return;
        render(view);
      },
      destroy() {
        hide(false);
        el.remove();
        document.removeEventListener('mouseup', onMouseUp, true);
        document.removeEventListener('keyup',   onKeyUp,   true);
      },
    };
  },
}));
