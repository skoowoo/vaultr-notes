import { $nodeSchema, $remark, $view } from '@milkdown/utils';
import remarkFrontmatter from 'remark-frontmatter';
import yaml from 'js-yaml';

// ── Remark plugin ─────────────────────────────────────────────────────────────
// Passes ['yaml'] so remark parses --- blocks as yaml AST nodes instead of
// treating them as paragraph text or thematic breaks.
export const remarkFrontmatterPlugin = $remark(
  'remarkFrontmatter',
  () => remarkFrontmatter,
  ['yaml'],
);

// ── ProseMirror node schema ───────────────────────────────────────────────────
export const frontmatterSchema = $nodeSchema('frontmatter', () => ({
  group: 'block',
  atom: true,
  isolating: true,
  attrs: {
    yaml: { default: '' },
  },
  parseMarkdown: {
    match: node => node.type === 'yaml',
    runner: (state, node, type) => {
      state.addNode(type, { yaml: node.value });
    },
  },
  toMarkdown: {
    match: node => node.type.name === 'frontmatter',
    runner: (state, node) => {
      state.addNode('yaml', undefined, node.attrs.yaml);
    },
  },
  parseDOM: [{ tag: 'div[data-fm]' }],
  toDOM: () => ['div', { 'data-fm': '' }],
}));

function isURL(s) {
  return typeof s === 'string' && /^https?:\/\//i.test(s);
}

// js-yaml turns YAML timestamps into JS Date objects; those must render as plain
// body text like the reader (YYYY-MM-DD in UTC — same convention as distill),
// never as fm-pre dumps which read as bordered “cards”.
function dateToDistillDay(d) {
  if (!(d instanceof Date) || Number.isNaN(d.getTime())) return null;
  return d.toISOString().slice(0, 10);
}

// Display YAML date-time scalars as YYYY-MM-DD when they are plain strings (reader
// shows distill dates as calendar days in UTC).
function tryFormatIsoDateString(s) {
  if (typeof s !== 'string') return null;
  if (/^\d{4}-\d{2}-\d{2}$/.test(s)) return s;
  const m =
    /^(\d{4}-\d{2}-\d{2})[Tt ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?$/.exec(
      s,
    );
  return m ? m[1] : null;
}

function renderValueInto(parent, v) {
  if (v === null || v === undefined) return;

  if (v instanceof Date) {
    const day = dateToDistillDay(v);
    parent.appendChild(
      document.createTextNode(day ?? String(v)),
    );
    return;
  }

  if (Array.isArray(v)) {
    if (v.length === 0) return;
    const allScalar = v.every(
      item =>
        item === null ||
        item instanceof Date ||
        ['string', 'number', 'boolean'].includes(typeof item),
    );
    if (allScalar) {
      for (const item of v) {
        const tag = document.createElement('span');
        tag.className = 'fm-tag';
        if (item instanceof Date) {
          const day = dateToDistillDay(item);
          tag.textContent = day ?? String(item);
        } else {
          tag.textContent = item === null || item === undefined ? '' : String(item);
        }
        parent.appendChild(tag);
      }
      return;
    }
    const pre = document.createElement('pre');
    pre.className = 'fm-pre';
    pre.textContent = yaml.dump(v, { lineWidth: -1 }).trimEnd();
    parent.appendChild(pre);
    return;
  }

  if (typeof v === 'object') {
    const pre = document.createElement('pre');
    pre.className = 'fm-pre';
    pre.textContent = yaml.dump(v, { lineWidth: -1 }).trimEnd();
    parent.appendChild(pre);
    return;
  }

  const str = String(v);
  const isoDay = tryFormatIsoDateString(str);
  if (isoDay) {
    parent.appendChild(document.createTextNode(isoDay));
    return;
  }
  if (isURL(str)) {
    const wrap = document.createElement('span');
    wrap.className = 'fm-val-text';
    wrap.title = str;
    const a = document.createElement('a');
    a.href = str;
    a.target = '_blank';
    a.rel = 'noopener noreferrer';
    a.textContent = str;
    wrap.appendChild(a);
    parent.appendChild(wrap);
    return;
  }
  const wrap = document.createElement('span');
  wrap.className = 'fm-val-text';
  wrap.title = str;
  wrap.textContent = str;
  parent.appendChild(wrap);
}

// ── NodeView ──────────────────────────────────────────────────────────────────
// Renders frontmatter with the same fm-* structure as the server-side note view.
// Click "Edit" to switch to a raw-YAML textarea; Escape or Done saves.
export const frontmatterView = $view(frontmatterSchema.node, () => {
  return (node, view, getPos) => {
    const dom = document.createElement('div');
    dom.setAttribute('data-fm', '');

    let editing = false;

    function parseYaml(raw) {
      try {
        return yaml.load(raw);
      } catch {
        return null;
      }
    }

    function renderView() {
      dom.innerHTML = '';

      const card = document.createElement('div');
      card.className = 'fm-card';

      const hd = document.createElement('div');
      hd.className = 'fm-card-hd';

      const cardLabel = document.createElement('span');
      cardLabel.className = 'fm-card-label';
      cardLabel.textContent = 'metadata';
      hd.appendChild(cardLabel);

      const editBtn = document.createElement('button');
      editBtn.type = 'button';
      editBtn.className = 'fm-edit-btn';
      editBtn.title = 'Edit metadata';
      editBtn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h9"/><path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4Z"/></svg>`;
      editBtn.addEventListener('click', e => {
        e.preventDefault();
        e.stopPropagation();
        openEditor();
      });
      hd.appendChild(editBtn);
      card.appendChild(hd);

      const data = parseYaml(node.attrs.yaml);
      const dl = document.createElement('dl');
      dl.className = 'fm-grid';

      const rawTrim = (node.attrs.yaml || '').trim();

      if (data === null && rawTrim.length > 0) {
        const dt = document.createElement('dt');
        dt.className = 'fm-key';
        dt.textContent = 'YAML';
        const dd = document.createElement('dd');
        dd.className = 'fm-val';
        const hint = document.createElement('span');
        hint.className = 'fm-parse-hint';
        hint.textContent = 'Invalid YAML';
        dd.appendChild(hint);
        dl.appendChild(dt);
        dl.appendChild(dd);
        card.appendChild(dl);
      } else if (data != null && typeof data === 'object' && !Array.isArray(data)) {
        const entries = Object.entries(data);
        if (entries.length > 0) {
          for (const [k, v] of entries) {
            const dt = document.createElement('dt');
            dt.className = 'fm-key';
            dt.textContent = k;
            const dd = document.createElement('dd');
            dd.className = 'fm-val';
            renderValueInto(dd, v);
            dl.appendChild(dt);
            dl.appendChild(dd);
          }
          card.appendChild(dl);
        }
      } else if (rawTrim.length > 0) {
        const dt = document.createElement('dt');
        dt.className = 'fm-key';
        dt.textContent = ' ';
        const dd = document.createElement('dd');
        dd.className = 'fm-val';
        const pre = document.createElement('pre');
        pre.className = 'fm-pre';
        pre.textContent = yaml.dump(data, { lineWidth: -1 }).trimEnd();
        dd.appendChild(pre);
        dl.appendChild(dt);
        dl.appendChild(dd);
        card.appendChild(dl);
      }

      dom.appendChild(card);
    }

    function openEditor() {
      if (editing) return;
      editing = true;
      dom.innerHTML = '';

      const card = document.createElement('div');
      card.className = 'fm-card fm-card--editing';

      const hd = document.createElement('div');
      hd.className = 'fm-card-hd';

      const cardLabel = document.createElement('span');
      cardLabel.className = 'fm-card-label';
      cardLabel.textContent = 'metadata';
      hd.appendChild(cardLabel);

      const doneBtn = document.createElement('button');
      doneBtn.type = 'button';
      doneBtn.className = 'fm-done-btn';
      doneBtn.textContent = 'Done';
      doneBtn.addEventListener('click', e => {
        e.preventDefault();
        e.stopPropagation();
        commitEdit();
      });
      hd.appendChild(doneBtn);
      card.appendChild(hd);

      const ta = document.createElement('textarea');
      ta.className = 'fm-raw-yaml';
      ta.value = node.attrs.yaml;
      ta.rows = Math.max(3, (node.attrs.yaml.match(/\n/g) || []).length + 1);
      ta.spellcheck = false;
      ta.addEventListener('keydown', e => {
        if (e.key === 'Escape') {
          e.preventDefault();
          commitEdit();
        }
        e.stopPropagation();
      });
      ta.addEventListener('keyup', e => e.stopPropagation());
      ta.addEventListener('click', e => e.stopPropagation());
      ta.addEventListener('input', e => e.stopPropagation());

      card.appendChild(ta);
      dom.appendChild(card);
      ta.focus();

      function commitEdit() {
        const newYaml = ta.value.trim();
        editing = false;
        const pos = getPos();
        if (pos !== undefined && newYaml !== node.attrs.yaml) {
          view.dispatch(
            view.state.tr.setNodeMarkup(pos, undefined, { yaml: newYaml }),
          );
        } else {
          renderView();
        }
      }
    }

    dom.addEventListener('dragstart', e => e.preventDefault());
    renderView();

    return {
      dom,
      update(updatedNode) {
        if (updatedNode.type.name !== 'frontmatter') return false;
        node = updatedNode;
        if (!editing) renderView();
        return true;
      },
      stopEvent(e) {
        return editing || dom.contains(e.target);
      },
      ignoreMutation() {
        return true;
      },
    };
  };
});

export const frontmatterPlugin = [
  remarkFrontmatterPlugin,
  frontmatterSchema,
  frontmatterView,
];
