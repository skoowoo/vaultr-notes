import { $nodeSchema, $remark, $view } from '@milkdown/utils';

// ── Micromark extension for ![[filename]] ─────────────────────────────────────
// Recognized BEFORE CommonMark image syntax so ![[x]] is never misread as ![alt](x).

const C_EXCLAMATION = 33;  // !
const C_OPEN        = 91;  // [
const C_CLOSE       = 93;  // ]

function tokenizeWikiImage(effects, ok, nok) {
  return start;

  function start(code) {
    effects.enter('wikiImage');
    effects.enter('wikiImageMarker');
    effects.consume(code); // !
    return firstBracket;
  }

  function firstBracket(code) {
    if (code !== C_OPEN) return nok(code);
    effects.consume(code); // [
    return secondBracket;
  }

  function secondBracket(code) {
    if (code !== C_OPEN) return nok(code);
    effects.consume(code); // [
    effects.exit('wikiImageMarker');
    effects.enter('wikiImageData');
    return data;
  }

  function data(code) {
    if (code === C_CLOSE) {
      effects.exit('wikiImageData');
      effects.enter('wikiImageMarkerClose');
      effects.consume(code); // first ]
      return closeSecond;
    }
    if (code === null || code < 0) return nok(code);
    effects.consume(code);
    return data;
  }

  function closeSecond(code) {
    if (code !== C_CLOSE) return nok(code);
    effects.consume(code); // second ]
    effects.exit('wikiImageMarkerClose');
    effects.exit('wikiImage');
    return ok;
  }
}

const wikiImageSyntax = {
  text: { [C_EXCLAMATION]: { tokenize: tokenizeWikiImage } },
};

// ── mdast fromMarkdown extension ──────────────────────────────────────────────

const wikiImageFromMarkdown = {
  enter: {
    wikiImage(token) {
      this.enter({ type: 'wikiImage', value: '' }, token);
    },
  },
  exit: {
    wikiImageData(token) {
      this.stack[this.stack.length - 1].value = this.sliceSerialize(token);
    },
    wikiImage(token) {
      this.exit(token);
    },
  },
};

// ── mdast toMarkdown extension ────────────────────────────────────────────────

const wikiImageToMarkdown = {
  handlers: {
    wikiImage(node) {
      return '![[' + (node.value || '') + ']]';
    },
  },
};

// ── Remark unified plugin ─────────────────────────────────────────────────────

function remarkWikiImagePlugin() {
  const data = this.data();

  // Prepend so our !-construct is tried before the built-in image construct.
  if (!data.micromarkExtensions) data.micromarkExtensions = [];
  data.micromarkExtensions.unshift(wikiImageSyntax);

  if (!data.fromMarkdownExtensions) data.fromMarkdownExtensions = [];
  data.fromMarkdownExtensions.push(wikiImageFromMarkdown);

  if (!data.toMarkdownExtensions) data.toMarkdownExtensions = [];
  data.toMarkdownExtensions.push(wikiImageToMarkdown);
}

export const remarkWikiImageRemark = $remark(
  'remarkWikiImage',
  () => remarkWikiImagePlugin,
);

// ── ProseMirror node schema ───────────────────────────────────────────────────

export const wikiImageSchema = $nodeSchema('wikiImage', () => ({
  group: 'inline',
  inline: true,
  atom: true,
  attrs: {
    value: { default: '' }, // filename, e.g. "photo.png"
  },
  parseMarkdown: {
    match: node => node.type === 'wikiImage',
    runner: (state, node, type) => {
      state.addNode(type, { value: node.value || '' });
    },
  },
  toMarkdown: {
    match: node => node.type.name === 'wikiImage',
    runner: (state, node) => {
      state.addNode('wikiImage', undefined, node.attrs.value, {});
    },
  },
  parseDOM: [{
    tag: 'span[data-wi]',
    getAttrs: dom => ({ value: dom.getAttribute('data-wi-value') || '' }),
  }],
  toDOM: node => ['span', { 'data-wi': '', 'data-wi-value': node.attrs.value }],
}));

// ── NodeView ──────────────────────────────────────────────────────────────────

export const wikiImageView = $view(wikiImageSchema.node, () => {
  return (node) => {
    const dom = document.createElement('span');
    dom.className = 'wiki-image-wrap';

    function render(n) {
      const filename = n.attrs.value || '';
      dom.setAttribute('data-wi', '');
      dom.setAttribute('data-wi-value', filename);
      dom.innerHTML = '';
      const img = document.createElement('img');
      img.src = '/api/images/serve?name=' + encodeURIComponent(filename);
      img.alt = filename;
      img.setAttribute('data-wiki-image', filename);
      dom.appendChild(img);
    }

    render(node);

    return {
      dom,
      update(updatedNode) {
        if (updatedNode.type.name !== 'wikiImage') return false;
        render(updatedNode);
        return true;
      },
      stopEvent: () => false,
      ignoreMutation: () => true,
    };
  };
});

export const wikiImagePlugin = [
  remarkWikiImageRemark,
  wikiImageSchema,
  wikiImageView,
];
