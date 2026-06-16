import { $inputRule } from '@milkdown/utils';
import { InputRule } from 'prosemirror-inputrules';
import { linkSchema } from '@milkdown/preset-commonmark';

// Converts [text](url) to a link mark when the closing ) is typed.
export const linkInputRule = $inputRule((ctx) =>
  new InputRule(
    /\[([^\]]+)\]\(([^)]+)\)$/,
    (state, match, start, end) => {
      const [, text, href] = match;
      if (!href) return null;
      const linkType = linkSchema.type(ctx);
      const mark = linkType.create({ href, title: null });
      const textNode = state.schema.text(text, [mark]);
      return state.tr.replaceWith(start, end, textNode);
    }
  )
);
