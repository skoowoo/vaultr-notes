/**
 * Stage 3 — Post-Turndown Markdown cleanup.
 *
 * Turndown's GFM table plugin silently falls back to raw HTML when it cannot
 * satisfy its own requirements (e.g. inconsistent column counts, missing
 * thead). This pass converts any remaining <table> HTML fragments in the
 * Markdown string to proper GFM table syntax.
 */

/**
 * Replace remaining HTML <table> blocks in a Markdown string with GFM tables.
 *
 * @param markdown - Turndown output that may contain raw <table> HTML fragments.
 * @param doc      - Document used for parsing (document.createElement).
 */
export function convertRemainingHtmlTables(markdown: string, doc: Document): string {
  // Non-greedy match picks up the smallest (innermost) table first.
  return markdown.replace(/<table\b[^>]*>[\s\S]*?<\/table>/gi, (tableHtml) => {
    // If the fragment contains a nested <table> it is itself an inner table —
    // return unchanged; the outer table will be matched in a subsequent pass.
    const openingTags = tableHtml.match(/<table\b/gi)?.length ?? 0;
    if (openingTags > 1) return tableHtml;

    try {
      const wrapper = doc.createElement("div");
      wrapper.innerHTML = tableHtml;
      const table = wrapper.querySelector("table");
      if (!table) return tableHtml;

      const rows = Array.from(table.rows);
      if (rows.length === 0) return tableHtml;

      const maxCols = Math.max(...rows.map((r) => r.cells.length));
      if (maxCols === 0) return tableHtml;

      const mdRows = rows.map((row) => {
        const texts = Array.from(row.cells).map((cell) =>
          (cell.textContent ?? "")
            .trim()
            .replace(/\r?\n\s*/g, " ")
            .replace(/\|/g, "\\|"),
        );
        while (texts.length < maxCols) texts.push("");
        return "| " + texts.join(" | ") + " |";
      });

      const separator = "| " + Array(maxCols).fill("---").join(" | ") + " |";
      mdRows.splice(1, 0, separator);

      return "\n\n" + mdRows.join("\n") + "\n\n";
    } catch {
      return tableHtml;
    }
  });
}
