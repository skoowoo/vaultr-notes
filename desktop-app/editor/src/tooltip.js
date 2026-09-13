import { $prose } from '@milkdown/utils';
import { Plugin, PluginKey, TextSelection } from '@milkdown/prose/state';
import { lift, toggleMark as pmToggleMark, setBlockType, wrapIn } from '@milkdown/prose/commands';
import { wrapInList } from '@milkdown/prose/schema-list';
import {
  Feather, Check, Copy, X, RemoveFormatting, List, ListOrdered, Quote,
  Bold, Italic, Strikethrough, Code2, Heading1, Heading2, Heading3, Heading4,
} from 'lucide';

const TOOLTIP_KEY = new PluginKey('vaultr-format-tooltip');

// ── Lucide SVG helper ─────────────────────────────────────────────────────────

function lucideSvg(iconData) {
  const children = iconData.map(([tag, attrs]) => {
    const attrStr = Object.entries(attrs).map(([k, v]) => `${k}="${v}"`).join(' ');
    return `<${tag} ${attrStr}/>`;
  }).join('');
  return (
    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ' +
    'stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">' +
    children + '</svg>'
  );
}

const SVG_FEATHER = lucideSvg(Feather);
const SVG_CHECK   = lucideSvg(Check);
const SVG_COPY    = lucideSvg(Copy);
const SVG_X       = lucideSvg(X);
const SVG_CLEAR   = lucideSvg(RemoveFormatting);
const SVG_LIST    = lucideSvg(List);
const SVG_OLIST   = lucideSvg(ListOrdered);
const SVG_QUOTE   = lucideSvg(Quote);
const SVG_BOLD    = lucideSvg(Bold);
const SVG_ITALIC  = lucideSvg(Italic);
const SVG_STRIKE  = lucideSvg(Strikethrough);
const SVG_CODE    = lucideSvg(Code2);
const SVG_HEADINGS = [Heading1, Heading2, Heading3, Heading4].map(lucideSvg);

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

// Returns true if cursor is inside the frontmatter node.
function selectionInFrontmatter(state) {
  const { $from } = state.selection;
  for (let d = $from.depth; d >= 0; d--) {
    if ($from.node(d).type.name === 'frontmatter') return true;
  }
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
  // Menus can run taller than the old single-row bar — clamp so a long one
  // (headings + lists + inline marks + actions) never runs off the bottom.
  top = Math.min(top, window.innerHeight - box.height - gap);
  top = Math.max(top, gap);

  // Horizontal: center on the head (cursor end) position, clamped to viewport.
  let left = headCoords.left - box.width / 2;
  left = Math.max(gap, Math.min(left, window.innerWidth - box.width - gap));

  el.style.top  = top  + 'px';
  el.style.left = left + 'px';
}

// ── Row builder ───────────────────────────────────────────────────────────────

function makeRow(iconSvg, label, opts) {
  const btn = document.createElement('button');
  btn.type = 'button';
  btn.className = 'mdt-row' + (opts && opts.active ? ' active' : '');
  if (opts && opts.title) btn.title = opts.title;
  const icon = document.createElement('span');
  icon.className = 'mdt-row-icon';
  icon.innerHTML = iconSvg;
  btn.appendChild(icon);
  const labelEl = document.createElement('span');
  labelEl.className = 'mdt-row-label';
  labelEl.textContent = label;
  btn.appendChild(labelEl);
  return { btn, icon, labelEl };
}

function sepEl() {
  const s = document.createElement('div');
  s.className = 'mdt-sep';
  return s;
}

// ── Plugin ────────────────────────────────────────────────────────────────────

const HEADING_LEVELS = [1, 2, 3, 4];

const INLINE_FMT_BTNS = [
  { mark: 'strong',         label: 'Bold',          icon: SVG_BOLD   },
  { mark: 'emphasis',       label: 'Italic',        icon: SVG_ITALIC },
  { mark: 'strike_through', label: 'Strikethrough', icon: SVG_STRIKE },
  { mark: 'inlineCode',     label: 'Inline code',   icon: SVG_CODE   },
];

export const tooltipPlugin = $prose(() => new Plugin({
  key: TOOLTIP_KEY,
  view() {
    const el = document.createElement('div');
    el.className = 'milkdown-tooltip';
    document.body.appendChild(el);

    let pmView  = null;
    let visible = false;

    // Making a selection no longer pops the tooltip by itself — that read as
    // noisy. It now only opens on an explicit ask: right-click (contextmenu)
    // on a non-empty selection, or the Mod-T shortcut (see __vaultrToggleFormatTooltip,
    // wired to a hotkey in drawer.js).

    function onMouseDown(e) {
      if (!pmView) return;
      const editArea = document.getElementById('drawer-edit-area');
      const inEditor = pmView.dom.contains(e.target) ||
                       (editArea && editArea.contains(e.target));
      if (visible && !inEditor && !el.contains(e.target)) {
        hide(false);
      }
    }

    function onContextMenu(e) {
      if (!pmView) return;
      const editArea = document.getElementById('drawer-edit-area');
      const inEditor = pmView.dom.contains(e.target) ||
                       (editArea && editArea.contains(e.target));
      if (!inEditor || pmView.state.selection.empty) return;
      e.preventDefault();
      // Always reposition from scratch for a fresh right-click.
      if (visible) { window.__vaultrEscPop?.('format-tooltip'); visible = false; }
      render(pmView);
    }

    function onScroll() {
      if (!visible || !pmView) return;
      positionTooltip(el, pmView);
    }

    document.addEventListener('mousedown',   onMouseDown,   true);
    document.addEventListener('contextmenu', onContextMenu, true);
    document.addEventListener('scroll',      onScroll,      true);

    // Keyboard-shortcut entry point (Mod-T, registered in drawer.js).
    window.__vaultrToggleFormatTooltip = function() {
      if (!pmView || pmView.state.selection.empty) return;
      if (visible) { hide(false); return; }
      render(pmView);
    };

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
      if (selectionInFrontmatter(state)) { hide(false); return; }

      const showClear = !hasToggleFormats(state);
      const clearTarget = showClear ? detectClearTarget(state) : null;

      const selText = state.doc.textBetween(state.selection.from, state.selection.to, ' ');
      const wc = countWords(selText);
      const wcText = wc > 0 ? (wc + (wc === 1 ? ' word' : ' words')) : '';

      el.innerHTML = '';

      // ── Header: context label + dismiss ─────────────────────────────────────
      const header = document.createElement('div');
      header.className = 'mdt-header';
      const headerLabel = document.createElement('span');
      headerLabel.className = 'mdt-header-label';
      headerLabel.textContent = clearTarget
        ? (clearTarget.label + (wcText ? ' · ' + wcText : ''))
        : (wcText || 'Format');
      header.appendChild(headerLabel);
      const xBtn = document.createElement('button');
      xBtn.type = 'button';
      xBtn.className = 'mdt-x';
      xBtn.setAttribute('aria-label', 'Dismiss');
      xBtn.innerHTML = SVG_X;
      xBtn.addEventListener('mousedown', e => { e.preventDefault(); hide(true); });
      header.appendChild(xBtn);
      el.appendChild(header);

      if (showClear && clearTarget) {
        // ── Clear group: a single row ────────────────────────────────────────
        const { btn } = makeRow(SVG_CLEAR, 'Clear ' + clearTarget.label.toLowerCase());
        btn.addEventListener('mousedown', e => {
          e.preventDefault();
          clearComplex(view, clearTarget.kind);
        });
        el.appendChild(btn);
        el.appendChild(sepEl());
      } else {
        // ── Toggle group ──────────────────────────────────────────────────────
        const showBlock = shouldShowBlockFormats(state) || selectionInBlockquote(state);

        if (showBlock) {
          const headingLevel = selectionHeadingLevel(state);
          HEADING_LEVELS.forEach((level, i) => {
            const { btn } = makeRow(SVG_HEADINGS[i], 'Heading ' + level, { active: headingLevel === level });
            btn.addEventListener('mousedown', e => {
              e.preventDefault();
              toggleHeading(view, level);
            });
            el.appendChild(btn);
          });

          const { btn: quoteBtn } = makeRow(SVG_QUOTE, 'Blockquote', { active: selectionInBlockquote(state) });
          quoteBtn.addEventListener('mousedown', e => {
            e.preventDefault();
            toggleBlockquote(view);
          });
          el.appendChild(quoteBtn);

          const listType = selectionListType(state);

          const { btn: bulletBtn } = makeRow(SVG_LIST, 'Bullet list', { active: listType === 'bullet_list' });
          bulletBtn.addEventListener('mousedown', e => {
            e.preventDefault();
            toggleList(view, 'bullet_list');
          });
          el.appendChild(bulletBtn);

          const { btn: orderedBtn } = makeRow(SVG_OLIST, 'Ordered list', { active: listType === 'ordered_list' });
          orderedBtn.addEventListener('mousedown', e => {
            e.preventDefault();
            toggleList(view, 'ordered_list');
          });
          el.appendChild(orderedBtn);

          el.appendChild(sepEl());
        }

        INLINE_FMT_BTNS.forEach(({ mark, label, icon }) => {
          const { btn } = makeRow(icon, label, { active: isMarkActive(state, mark) });
          btn.addEventListener('mousedown', e => {
            e.preventDefault();
            applyToggleMark(view, mark);
          });
          el.appendChild(btn);
        });

        el.appendChild(sepEl());
      }

      // ── Always: copy, save as short ──────────────────────────────────────────
      const { btn: copyBtn, icon: copyIcon, labelEl: copyLabel } = makeRow(SVG_COPY, 'Copy as Markdown');
      copyBtn.addEventListener('mousedown', e => {
        e.preventDefault();
        window.__vaultrCopySelectionAsMd?.();
        copyIcon.innerHTML = SVG_CHECK;
        copyLabel.textContent = 'Copied';
        copyBtn.classList.add('mdt-row-success');
        setTimeout(() => {
          copyIcon.innerHTML = SVG_COPY;
          copyLabel.textContent = 'Copy as Markdown';
          copyBtn.classList.remove('mdt-row-success');
        }, 1200);
      });
      el.appendChild(copyBtn);

      const { btn: shortBtn, icon: shortIcon, labelEl: shortLabel } = makeRow(SVG_FEATHER, 'Save as short note');
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
          shortIcon.innerHTML = SVG_CHECK;
          shortLabel.textContent = 'Saved';
          shortBtn.classList.add('mdt-row-success');
          if (window.__vaultrAfterVaultMutation) await window.__vaultrAfterVaultMutation();
          setTimeout(() => hide(true), 600);
        } catch (err) {
          if (window.showError) window.showError((err && err.message) || 'Save failed', 'Short note');
        }
      });
      el.appendChild(shortBtn);

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
        // Never initiate a new show from state updates — only contextmenu/Mod-T do that.
        if (!visible) return;
        const selSame = prevState && prevState.selection.eq(view.state.selection);
        const docSame = prevState && prevState.doc.eq(view.state.doc);
        if (selSame && docSame) return;
        render(view);
      },
      destroy() {
        hide(false);
        el.remove();
        document.removeEventListener('mousedown',   onMouseDown,   true);
        document.removeEventListener('contextmenu', onContextMenu, true);
        document.removeEventListener('scroll',      onScroll,      true);
        if (window.__vaultrToggleFormatTooltip) delete window.__vaultrToggleFormatTooltip;
      },
    };
  },
}));
