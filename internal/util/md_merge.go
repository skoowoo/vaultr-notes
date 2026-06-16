package util

import (
	"bytes"
	"strings"

	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// blockKind classifies a markdown block node for join decisions.
type blockKind int

const (
	blockOther       blockKind = iota
	blockBulletList            // unordered list: -, *, +
	blockOrderedList           // ordered list: 1., 2., …
)

// mdBlockParser is a lightweight goldmark parser for block-level AST inspection only.
var mdBlockParser = parser.NewParser(
	parser.WithBlockParsers(parser.DefaultBlockParsers()...),
)

// lastBlockKind parses src and returns the kind of the last top-level block node.
func lastBlockKind(src []byte) blockKind {
	if len(bytes.TrimSpace(src)) == 0 {
		return blockOther
	}
	doc := mdBlockParser.Parse(text.NewReader(src))
	var last gast.Node
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		last = child
	}
	return classifyNode(last)
}

// firstBlockKind parses src and returns the kind of the first top-level block node.
func firstBlockKind(src []byte) blockKind {
	if len(bytes.TrimSpace(src)) == 0 {
		return blockOther
	}
	doc := mdBlockParser.Parse(text.NewReader(src))
	return classifyNode(doc.FirstChild())
}

func classifyNode(node gast.Node) blockKind {
	if node == nil || node.Kind() != gast.KindList {
		return blockOther
	}
	if node.(*gast.List).IsOrdered() {
		return blockOrderedList
	}
	return blockBulletList
}

// MdAppend merges existing and incoming bytes following markdown block rules:
//
//  1. If existing is empty the incoming content is returned as-is (after
//     ensuring it ends with a newline).
//  2. If the last block of existing and the first block of incoming are both
//     lists of the same kind (bullet or ordered), AND incoming does not begin
//     with an explicit blank line, they are joined with a single newline —
//     continuing the list without a blank-line break.
//  3. In all other cases one blank line (two newlines) separates the blocks.
//
// The returned slice always ends with exactly one newline.
func MdAppend(existing, incoming []byte) []byte {
	incoming = ensureNewline(incoming)

	if len(bytes.TrimSpace(existing)) == 0 {
		return incoming
	}

	last := lastBlockKind(existing)
	// Trim leading newlines only for AST classification; the original bytes are
	// preserved in the output so the caller's explicit blank line is kept.
	first := firstBlockKind(bytes.TrimLeft(incoming, "\n\r"))

	var sep []byte
	if last != blockOther && last == first && !startsWithBlankLine(incoming) {
		// Continuing the same list type — only one newline needed.
		sep = neededNewlinesBetween(existing, incoming, 1)
	} else {
		// New block — ensure a blank line between the two.
		sep = neededNewlinesBetween(existing, incoming, 2)
	}

	out := make([]byte, 0, len(existing)+len(sep)+len(incoming))
	out = append(out, existing...)
	out = append(out, sep...)
	out = append(out, incoming...)
	return out
}

// startsWithBlankLine reports whether b begins with a newline character,
// indicating the caller intentionally separated the content with a blank line.
func startsWithBlankLine(b []byte) bool {
	return len(b) > 0 && b[0] == '\n'
}

// neededNewlinesBetween returns the separator bytes required so that the
// junction of existing+sep+incoming contains at least n newlines in a row.
// It counts both the trailing newlines of existing and the leading newlines of
// incoming, so it never double-adds separators.
func neededNewlinesBetween(existing, incoming []byte, n int) []byte {
	have := trailingNewlineCount(existing) + leadingNewlineCount(incoming)
	if need := n - have; need > 0 {
		return bytes.Repeat([]byte{'\n'}, need)
	}
	return nil
}

func leadingNewlineCount(b []byte) int {
	count := 0
	for i := 0; i < len(b) && b[i] == '\n'; i++ {
		count++
	}
	return count
}

func trailingNewlineCount(b []byte) int {
	count := 0
	for i := len(b) - 1; i >= 0 && b[i] == '\n'; i-- {
		count++
	}
	return count
}

// MdPrepend inserts incoming into existing immediately after the first H1
// heading (# …). This keeps the document title at the top while placing new
// content before any existing body. Two rules apply:
//
//   - If existing contains no H1, incoming is prepended at the very beginning
//     of the document (same as a plain prepend).
//   - Separator logic between the surrounding blocks follows the same
//     smart-join rules as MdJoin: same-kind list continuation uses a single
//     newline; everything else uses a blank line.
//
// The returned slice always ends with exactly one newline.
func MdPrepend(existing, incoming []byte) []byte {
	incoming = ensureNewline(incoming)

	if len(bytes.TrimSpace(existing)) == 0 {
		return incoming
	}

	insertAt := afterFirstH1(existing)
	head := existing[:insertAt] // up to and including the H1 line (may be empty)
	tail := existing[insertAt:] // everything after the H1 (may be empty)

	// Separator between head (H1) and incoming.
	// When head is empty there is no leading content to separate from.
	var sep1 []byte
	if len(head) > 0 {
		sep1 = neededNewlinesBetween(head, incoming, 2)
	}

	// Separator between incoming and the rest of the document.
	var sep2 []byte
	if len(bytes.TrimSpace(tail)) > 0 {
		lastOfIncoming := lastBlockKind(incoming)
		// Strip the leading whitespace that separated the original H1 from its
		// first body block; it is positional, not intentional separation.
		trimmedTail := bytes.TrimLeft(tail, "\n\r")
		firstOfTail := firstBlockKind(trimmedTail)
		if lastOfIncoming != blockOther && lastOfIncoming == firstOfTail {
			// Same list kind — collapse original spacing and continue the list.
			sep2 = neededNewlinesBetween(incoming, trimmedTail, 1)
			tail = trimmedTail
		} else {
			sep2 = neededNewlinesBetween(incoming, tail, 2)
		}
	}

	out := make([]byte, 0, len(head)+len(sep1)+len(incoming)+len(sep2)+len(tail))
	out = append(out, head...)
	out = append(out, sep1...)
	out = append(out, incoming...)
	out = append(out, sep2...)
	out = append(out, tail...)
	return out
}

// afterFirstH1 returns the byte offset in src immediately after the trailing
// newline of the first top-level H1 heading line.
//
// When YAML frontmatter is present it is skipped before searching for H1.
// When no H1 is found the offset returned is frontmatterEnd(src), so that
// callers insert after the frontmatter block (or at position 0 when there is
// neither frontmatter nor an H1).
//
// Note: goldmark's ATX heading Lines() segments cover only the inline content
// (e.g. "Title" for "# Title\n"), not the full raw line. We therefore scan
// forward from the segment's Stop to find and consume the trailing newline.
func afterFirstH1(src []byte) int {
	fm := frontmatterEnd(src)
	body := src[fm:]

	doc := mdBlockParser.Parse(text.NewReader(body))
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() != gast.KindHeading {
			continue
		}
		if child.(*gast.Heading).Level != 1 {
			continue
		}
		lines := child.Lines()
		if lines.Len() == 0 {
			continue
		}
		pos := lines.At(lines.Len() - 1).Stop
		// Advance past any non-newline characters (e.g. trailing spaces).
		for pos < len(body) && body[pos] != '\n' {
			pos++
		}
		// Consume the newline itself.
		if pos < len(body) {
			pos++
		}
		return fm + pos
	}
	// No H1: insert after the frontmatter block (or at 0 if none).
	return fm
}

// FrontmatterEnd is the exported form of frontmatterEnd.
func FrontmatterEnd(src []byte) int { return frontmatterEnd(src) }

// frontmatterEnd returns the byte offset immediately after the closing
// delimiter of a YAML frontmatter block (--- or ...). Returns 0 when src does
// not begin with a valid frontmatter block.
//
// A valid frontmatter block starts with "---\n" on the very first line and
// ends with a line that is exactly "---" or "...".
func frontmatterEnd(src []byte) int {
	if !bytes.HasPrefix(src, []byte("---\n")) {
		return 0
	}
	i := 4 // skip opening "---\n"
	for i < len(src) {
		nl := bytes.IndexByte(src[i:], '\n')
		var line []byte
		if nl < 0 {
			line = src[i:]
			i = len(src)
		} else {
			line = src[i : i+nl]
			i = i + nl + 1
		}
		if bytes.Equal(line, []byte("---")) || bytes.Equal(line, []byte("...")) {
			return i // i already points past the closing newline
		}
	}
	return 0 // no closing delimiter found
}

// ── Section-targeted operations ──────────────────────────────────────────────

// normalizeHeading strips any leading '#' characters and surrounding spaces
// from a heading argument, so that callers may pass either "Morning" or
// "## Morning" and get the same match result.
func normalizeHeading(h string) string {
	return strings.TrimSpace(strings.TrimLeft(h, "#"))
}

// MdAppendSection appends incoming after the content of the last section whose
// heading text matches heading (case-insensitive). A section ends at the next
// heading of equal or higher level (lower #-count). YAML frontmatter is
// transparent. Falls back to MdJoin when no matching heading is found.
//
// "Last" means the final occurrence in the document — this matches the
// expectation for reverse-chronological logs where the most recent section is
// at the bottom.
func MdAppendSection(existing, incoming []byte, heading string) []byte {
	incoming = ensureNewline(incoming)
	heading = normalizeHeading(heading)

	fm := frontmatterEnd(existing)
	body := existing[fm:]
	hs := extractHeadings(body)

	// Backwards search: find the last matching heading.
	idx := -1
	for i := len(hs) - 1; i >= 0; i-- {
		if strings.EqualFold(hs[i].Text, heading) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return MdAppend(existing, incoming)
	}

	end := bodySectionEnd(body, hs, idx)
	head := existing[:fm+end] // everything up to (not including) next sibling heading
	tail := existing[fm+end:] // next sibling heading onwards (may be empty)

	// sep1: join incoming to the end of the section content.
	last := lastBlockKind(head)
	first := firstBlockKind(bytes.TrimLeft(incoming, "\n\r"))
	var sep1 []byte
	if last != blockOther && last == first && !startsWithBlankLine(incoming) {
		sep1 = neededNewlinesBetween(head, incoming, 1)
	} else {
		sep1 = neededNewlinesBetween(head, incoming, 2)
	}

	// sep2: blank line before the next sibling heading (if any).
	var sep2 []byte
	if len(bytes.TrimSpace(tail)) > 0 {
		sep2 = neededNewlinesBetween(incoming, tail, 2)
	}

	out := make([]byte, 0, len(head)+len(sep1)+len(incoming)+len(sep2)+len(tail))
	out = append(out, head...)
	out = append(out, sep1...)
	out = append(out, incoming...)
	out = append(out, sep2...)
	out = append(out, tail...)
	return out
}

// MdPrependSection inserts incoming at the start of the content of the first
// section whose heading text matches heading (case-insensitive). YAML
// frontmatter is transparent. Falls back to MdPrepend when no matching
// heading is found.
//
// "First" means the earliest occurrence — natural for inserting a new entry
// at the top of a named section.
func MdPrependSection(existing, incoming []byte, heading string) []byte {
	incoming = ensureNewline(incoming)
	heading = normalizeHeading(heading)

	fm := frontmatterEnd(existing)
	body := existing[fm:]
	hs := extractHeadings(body)

	// Forward search: find the first matching heading.
	idx := -1
	for i := 0; i < len(hs); i++ {
		if strings.EqualFold(hs[i].Text, heading) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return MdPrepend(existing, incoming)
	}

	insertAt := fm + hs[idx].ContentStart
	head := existing[:insertAt] // up to and including the matched heading line
	tail := existing[insertAt:] // section content + rest of document

	// sep1: blank line between the heading and incoming.
	var sep1 []byte
	if len(head) > 0 {
		sep1 = neededNewlinesBetween(head, incoming, 2)
	}

	// sep2: join incoming to the existing section content.
	// The tail may begin with the blank line that separated the heading from
	// its content; strip that positional whitespace before smart-joining.
	var sep2 []byte
	if len(bytes.TrimSpace(tail)) > 0 {
		trimmedTail := bytes.TrimLeft(tail, "\n\r")
		lastOfIncoming := lastBlockKind(incoming)
		firstOfTail := firstBlockKind(trimmedTail)
		if lastOfIncoming != blockOther && lastOfIncoming == firstOfTail {
			sep2 = neededNewlinesBetween(incoming, trimmedTail, 1)
			tail = trimmedTail
		} else {
			sep2 = neededNewlinesBetween(incoming, tail, 2)
		}
	}

	out := make([]byte, 0, len(head)+len(sep1)+len(incoming)+len(sep2)+len(tail))
	out = append(out, head...)
	out = append(out, sep1...)
	out = append(out, incoming...)
	out = append(out, sep2...)
	out = append(out, tail...)
	return out
}

// ── Heading extraction ────────────────────────────────────────────────────────

// HeadingInfo records the position and content of an ATX heading line.
type HeadingInfo struct {
	Level        int    // 1–6
	Text         string // heading text without # markers or surrounding spaces
	LineStart    int    // byte offset of the first '#' character
	ContentStart int    // byte offset immediately after the heading line's '\n'
}

// headingInfo is a package-private alias so internal code keeps compiling
// without churn while callers in other packages use HeadingInfo.
type headingInfo = HeadingInfo

// ExtractHeadings scans src for ATX heading lines (# through ######) and
// returns them in document order. YAML frontmatter must be stripped before
// calling (use FrontmatterEnd if needed).
//
// Exported so that other packages (e.g. cli) can obtain reliable heading byte
// positions without re-implementing the scan.
func ExtractHeadings(src []byte) []HeadingInfo {
	return extractHeadings(src)
}

// extractHeadings is the internal implementation.
func extractHeadings(src []byte) []headingInfo {
	var out []headingInfo
	i := 0
	for i < len(src) {
		lineStart := i
		nl := bytes.IndexByte(src[i:], '\n')
		var line []byte
		var afterLine int
		if nl < 0 {
			line = src[i:]
			afterLine = len(src)
		} else {
			line = src[i : i+nl]
			afterLine = i + nl + 1
		}
		i = afterLine

		// Count leading '#' characters.
		level := 0
		for level < len(line) && line[level] == '#' {
			level++
		}
		if level == 0 || level > 6 {
			continue
		}
		// After the '#' run there must be a space/tab or end of line.
		rest := line[level:]
		if len(rest) > 0 && rest[0] != ' ' && rest[0] != '\t' {
			continue
		}

		// Normalise heading text: trim spaces, strip optional closing markers.
		text := bytes.TrimSpace(rest)
		if j := bytes.LastIndexFunc(text, func(r rune) bool { return r != '#' }); j >= 0 && j < len(text)-1 {
			if text[j] == ' ' || text[j] == '\t' {
				text = bytes.TrimSpace(text[:j])
			}
		}

		out = append(out, headingInfo{
			Level:        level,
			Text:         string(text),
			LineStart:    lineStart,
			ContentStart: afterLine,
		})
	}
	return out
}

// BodySectionEnd returns the byte offset within body where the section headed
// by headings[idx] ends: the LineStart of the next heading whose Level is ≤
// the target's Level, or len(body) if no such heading exists.
func BodySectionEnd(body []byte, headings []HeadingInfo, idx int) int {
	return bodySectionEnd(body, headings, idx)
}

// bodySectionEnd is the internal implementation.
func bodySectionEnd(body []byte, headings []headingInfo, idx int) int {
	target := headings[idx]
	for i := idx + 1; i < len(headings); i++ {
		if headings[i].Level <= target.Level {
			return headings[i].LineStart
		}
	}
	return len(body)
}

func ensureNewline(b []byte) []byte {
	if len(b) > 0 && b[len(b)-1] != '\n' {
		return append(b, '\n')
	}
	return b
}
