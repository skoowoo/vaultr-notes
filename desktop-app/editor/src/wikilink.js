import { $nodeSchema, $remark, $view } from '@milkdown/utils';
import remarkWikiLink from 'remark-wiki-link';

// ── Remark parse plugin ───────────────────────────────────────────────────────
// Use | as alias divider to match Obsidian format: [[Page|Alias]].
// hrefTemplate is unused in WYSIWYG mode but kept sensible for HTML export.
export const remarkWikiLinkPlugin = $remark(
  'remarkWikiLink',
  () => remarkWikiLink,
  { aliasDivider: '|', hrefTemplate: p => `/note/${p}` },
);

// ── Remark serialize override ─────────────────────────────────────────────────
// remark-wiki-link's default toMarkdown handler runs escape() on node.value and
// node.data.alias, which turns underscores into \_.  We register a second
// toMarkdownExtensions handler AFTER it; the last-registered handler for the
// same node type wins, so this replaces the default with a no-escape version.
function remarkWikiLinkSerializeOverride() {
  // Must push, not replace — this.data('key', value) replaces the whole array
  // and would wipe out handlers registered by earlier plugins (e.g. remark-frontmatter).
  const data = this.data();
  const ext = {
    handlers: {
      wikiLink(node) {
        const value = node.value || '';
        const alias = (node.data && node.data.alias) || value;
        return alias !== value ? `[[${value}|${alias}]]` : `[[${value}]]`;
      },
    },
  };
  if (data.toMarkdownExtensions) {
    data.toMarkdownExtensions.push(ext);
  } else {
    data.toMarkdownExtensions = [ext];
  }
}

export const remarkWikiLinkSerializePlugin = $remark(
  'remarkWikiLinkSerialize',
  () => remarkWikiLinkSerializeOverride,
);

// ── ProseMirror node schema ───────────────────────────────────────────────────
export const wikiLinkSchema = $nodeSchema('wikiLink', () => ({
  group: 'inline',
  inline: true,
  atom: true,
  attrs: {
    value: { default: '' },   // target page name, e.g. "My Page"
    alias: { default: '' },   // display text; empty string means same as value
  },
  parseMarkdown: {
    match: node => node.type === 'wikiLink',
    runner: (state, node, type) => {
      // data.alias equals value when there is no explicit alias
      const hasAlias = node.data.alias !== node.value;
      state.addNode(type, {
        value: node.value,
        alias: hasAlias ? node.data.alias : '',
      });
    },
  },
  toMarkdown: {
    match: node => node.type.name === 'wikiLink',
    runner: (state, node) => {
      // Props are spread into the AST node by SerializerState.#createMarkdownNode,
      // so { data: { alias } } becomes node.data.alias which mdast-util-wiki-link
      // reads when serializing back to [[value|alias]] or [[value]].
      const alias = node.attrs.alias || node.attrs.value;
      state.addNode('wikiLink', undefined, node.attrs.value, {
        data: { alias },
      });
    },
  },
  parseDOM: [{
    tag: 'span[data-wl]',
    getAttrs: dom => ({
      value: dom.getAttribute('data-wl-value') || '',
      alias: dom.getAttribute('data-wl-alias') || '',
    }),
  }],
  toDOM: node => ['span', {
    'data-wl': '',
    'data-wl-value': node.attrs.value,
    'data-wl-alias': node.attrs.alias,
  }],
}));

// ── NodeView ──────────────────────────────────────────────────────────────────
// Renders as a styled link chip.  The node is atomic so ProseMirror treats
// it as a single unit for cursor movement and deletion.
export const wikiLinkView = $view(wikiLinkSchema.node, () => {
  return (node) => {
    const dom = document.createElement('span');
    dom.setAttribute('data-wl', '');

    // Styling is entirely handled by .ProseMirror span[data-wl] in noteEditorCSS.
    function render(n) {
      const display = n.attrs.alias || n.attrs.value || '?';
      dom.setAttribute('data-wl-value', n.attrs.value);
      dom.setAttribute('data-wl-alias', n.attrs.alias);
      dom.textContent = display;
    }

    render(node);

    return {
      dom,
      update(updatedNode) {
        if (updatedNode.type.name !== 'wikiLink') return false;
        render(updatedNode);
        return true;
      },
      stopEvent: () => false,
      ignoreMutation: () => true,
    };
  };
});

export const wikiLinkPlugin = [
  remarkWikiLinkPlugin,
  remarkWikiLinkSerializePlugin,
  wikiLinkSchema,
  wikiLinkView,
];
