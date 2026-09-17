// Leading "---\n...\n---" as one Frontmatter block node.
// Without this: opening --- → HorizontalRule; closing --- → Setext H2.
const FRONTMATTER_NODE = 'Frontmatter';

function isDelimLine(line) {
  return line.text === '---';
}

export const frontmatterSyntax = {
  defineNodes: [{ name: FRONTMATTER_NODE, block: true }],
  parseBlock: [
    {
      name: FRONTMATTER_NODE,
      before: 'HorizontalRule', // must claim opening --- first
      parse(cx, line) {
        if (cx.lineStart !== 0 || !isDelimLine(line)) return false;
        const from = cx.lineStart;
        while (cx.nextLine()) {
          if (isDelimLine(line)) {
            cx.nextLine(); // consume closing ---
            break;
          }
        }
        // Unclosed → EOF, same as FencedCode.
        cx.addElement(cx.elt(FRONTMATTER_NODE, from, cx.prevLineEnd()));
        return true;
      },
    },
  ],
};
