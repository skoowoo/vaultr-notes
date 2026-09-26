package util

import (
	"strings"

	gm "github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// PreviewMaxRunes is the default character budget for GeneratePreview's Text.
const PreviewMaxRunes = 240

var previewParser = gm.New(gm.WithExtensions(extension.GFM)).Parser()

// previewByteBudget bounds how much text GeneratePreview accumulates before
// it stops growing the excerpt — only a short prefix is ever needed, and 4
// bytes/rune is the worst case for UTF-8. It does not bound how far the walk
// itself goes: TodoTotal/TodoDone/HasImage/HasCode describe the whole
// document, not just the excerpt's prefix, so a checklist or image further
// down still gets counted even once the excerpt is "full".
const previewByteBudget = PreviewMaxRunes * 4

// PreviewSummary is what GeneratePreview extracts from a note's markdown
// body in a single parse+walk: a short excerpt plus a few structural facts
// that are essentially free to observe along the way.
type PreviewSummary struct {
	Text      string // short excerpt from the start of the body
	TodoTotal int    // GFM task-list checkboxes ("- [ ]" / "- [x]") anywhere in the note
	TodoDone  int    // how many of those are checked
	HasImage  bool   // saw an embedded image (markdown or Obsidian ![[...]]) anywhere in the note
	HasCode   bool   // saw a fenced or indented code block anywhere in the note
}

// GeneratePreview extracts a short, human-readable excerpt from the start of
// a note's markdown body, for display in note list rows, plus a few structural
// signals (checklist progress, whether it has images/code) read from the whole
// document. YAML frontmatter is stripped; Obsidian-style wikilinks/image-embeds
// are resolved to their display text the same way the live renderer does; code
// blocks, images, and raw HTML contribute no excerpt text (they carry no
// readable prose, though a code/image block still sets HasCode/HasImage); and
// a leading level-1 heading is skipped, since the note's filename already
// serves as its title in the UI and repeating it in the excerpt would waste
// the budget. Text is whitespace-collapsed and truncated to maxRunes runes
// (not bytes — this app's content is CJK-heavy, and a byte-length cut would
// slice a character in half); maxRunes <= 0 uses PreviewMaxRunes.
func GeneratePreview(content []byte, maxRunes int) PreviewSummary {
	if maxRunes <= 0 {
		maxRunes = PreviewMaxRunes
	}
	_, body := ParseFrontmatter(content)
	if len(strings.TrimSpace(string(body))) == 0 {
		return PreviewSummary{}
	}
	body = expandWikiImages(body)
	body = expandWikilinks(body)

	root := previewParser.Parse(text.NewReader(body))
	skip := firstLevel1Heading(root)

	var sb strings.Builder
	var sum PreviewSummary
	textDone := false
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n == skip {
			return ast.WalkSkipChildren, nil
		}
		switch n.Kind() {
		case ast.KindCodeBlock, ast.KindFencedCodeBlock:
			if entering {
				sum.HasCode = true
			}
			return ast.WalkSkipChildren, nil
		case ast.KindImage:
			if entering {
				sum.HasImage = true
			}
			return ast.WalkSkipChildren, nil
		case ast.KindHTMLBlock, ast.KindRawHTML:
			return ast.WalkSkipChildren, nil
		case extast.KindTaskCheckBox:
			if entering {
				sum.TodoTotal++
				if n.(*extast.TaskCheckBox).IsChecked {
					sum.TodoDone++
				}
			}
		case ast.KindText:
			if entering && !textDone {
				sb.Write(n.(*ast.Text).Value(body))
			}
		case ast.KindString:
			if entering && !textDone {
				sb.Write(n.(*ast.String).Value)
			}
		default:
			// Block-level elements (paragraphs, headings, list items, table
			// cells, ...) end without their own trailing space in the
			// source, so add a separator once their content has been
			// emitted — otherwise adjacent blocks would run together.
			if !entering && !textDone && n.Type() == ast.TypeBlock {
				sb.WriteByte(' ')
			}
		}
		if !textDone && sb.Len() >= previewByteBudget {
			textDone = true
		}
		return ast.WalkContinue, nil
	})

	sum.Text = truncateRunes(strings.Join(strings.Fields(sb.String()), " "), maxRunes)
	return sum
}

// firstLevel1Heading returns root's first child when it is a level-1
// heading, so callers can skip it as a duplicate of the note's title.
func firstLevel1Heading(root ast.Node) ast.Node {
	if h, ok := root.FirstChild().(*ast.Heading); ok && h.Level == 1 {
		return h
	}
	return nil
}

// truncateRunes cuts s to at most max runes, appending an ellipsis when it
// was actually cut short. Counting runes (not bytes) keeps multi-byte
// characters intact.
func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return strings.TrimRight(string(r[:max]), " ") + "…"
}
