package util

import (
	"strings"
	"testing"
)

func TestGeneratePreviewText(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "plain paragraph",
			content: "Hello world.",
			want:    "Hello world.",
		},
		{
			name:    "frontmatter is stripped",
			content: "---\ntitle: x\ntags: [a]\n---\nBody text.",
			want:    "Body text.",
		},
		{
			name:    "leading level-1 heading is skipped",
			content: "# My Title\n\nSecond paragraph here.",
			want:    "Second paragraph here.",
		},
		{
			name:    "a non-h1 leading heading is kept",
			content: "## Sub heading\n\nText.",
			want:    "Sub heading Text.",
		},
		{
			name:    "fenced code block is skipped",
			content: "Intro text.\n\n```go\nfunc foo() {}\n```\n\nAfter code.",
			want:    "Intro text. After code.",
		},
		{
			name:    "markdown image is skipped",
			content: "![alt text](img.png) Real text here.",
			want:    "Real text here.",
		},
		{
			name:    "obsidian image embed is skipped",
			content: "![[photo.png]] Caption text.",
			want:    "Caption text.",
		},
		{
			name:    "wikilink resolves to its display text",
			content: "See [[Some Note|Custom Label]] for details.",
			want:    "See Custom Label for details.",
		},
		{
			name:    "bare wikilink resolves to its target name",
			content: "See [[Some Note]] for details.",
			want:    "See Some Note for details.",
		},
		{
			name:    "blank lines and multiple paragraphs collapse to one line",
			content: "First paragraph.\n\n\nSecond paragraph.",
			want:    "First paragraph. Second paragraph.",
		},
		{
			name:    "a task list item's own text still contributes to the excerpt",
			content: "- [ ] buy milk\n- [x] walk the dog",
			want:    "buy milk walk the dog",
		},
		{
			name:    "only frontmatter yields an empty preview",
			content: "---\ntitle: x\n---\n",
			want:    "",
		},
		{
			name:    "empty content yields an empty preview",
			content: "",
			want:    "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := GeneratePreview([]byte(c.content), 0).Text
			if got != c.want {
				t.Errorf("GeneratePreview(%q).Text = %q, want %q", c.content, got, c.want)
			}
		})
	}
}

// TestGeneratePreviewTruncatesByRune checks the cutoff counts runes, not
// bytes — this app's content is CJK-heavy, and a byte-based cut would slice
// a multi-byte character in half.
func TestGeneratePreviewTruncatesByRune(t *testing.T) {
	body := strings.Repeat("测试", 100) // 200 runes, 600 bytes in UTF-8
	got := GeneratePreview([]byte(body), 10).Text

	runes := []rune(got)
	if runes[len(runes)-1] != '…' {
		t.Fatalf("expected truncated preview to end with an ellipsis, got %q", got)
	}
	if len(runes) != 11 { // 10 content runes + the ellipsis
		t.Fatalf("GeneratePreview truncated to %d runes, want 11: %q", len(runes), got)
	}
}

// TestGeneratePreviewNoTruncationBelowLimit ensures short content is
// returned verbatim with no trailing ellipsis.
func TestGeneratePreviewNoTruncationBelowLimit(t *testing.T) {
	got := GeneratePreview([]byte("short"), 120).Text
	if got != "short" {
		t.Fatalf("GeneratePreview(%q).Text = %q, want unchanged", "short", got)
	}
}

// TestGeneratePreviewTodoCounts checks the GFM task-checkbox tally: total
// count and how many are checked, independent of ordering or surrounding text.
func TestGeneratePreviewTodoCounts(t *testing.T) {
	content := "- [ ] one\n- [x] two\n- [X] three\n- [ ] four"
	sum := GeneratePreview([]byte(content), 0)
	if sum.TodoTotal != 4 {
		t.Errorf("TodoTotal = %d, want 4", sum.TodoTotal)
	}
	if sum.TodoDone != 2 {
		t.Errorf("TodoDone = %d, want 2", sum.TodoDone)
	}
}

// TestGeneratePreviewNoTodos checks a note with no task list at all reports
// zero, not some sentinel — plain list items aren't checkboxes.
func TestGeneratePreviewNoTodos(t *testing.T) {
	sum := GeneratePreview([]byte("- just a list\n- nothing to check off"), 0)
	if sum.TodoTotal != 0 || sum.TodoDone != 0 {
		t.Errorf("TodoTotal/TodoDone = %d/%d, want 0/0", sum.TodoTotal, sum.TodoDone)
	}
}

// TestGeneratePreviewHasImage checks both a standard markdown image and an
// Obsidian-style embed set HasImage — the two forms are normalised to the
// same AST node before the walk (see expandWikiImages).
func TestGeneratePreviewHasImage(t *testing.T) {
	cases := map[string]string{
		"markdown image": "![alt](photo.png)",
		"obsidian embed": "![[photo.png]]",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			if !GeneratePreview([]byte(content), 0).HasImage {
				t.Errorf("HasImage = false for %q, want true", content)
			}
		})
	}
	if GeneratePreview([]byte("just text, no pictures"), 0).HasImage {
		t.Error("HasImage = true for plain text, want false")
	}
}

// TestGeneratePreviewHasCode checks both fenced and indented code blocks set
// HasCode, while an inline code span (a single `word` in a sentence) does not
// — that's a much weaker signal than an actual code block.
func TestGeneratePreviewHasCode(t *testing.T) {
	if !GeneratePreview([]byte("```go\nfunc f() {}\n```"), 0).HasCode {
		t.Error("HasCode = false for a fenced code block, want true")
	}
	if !GeneratePreview([]byte("Some intro.\n\n    indented code line"), 0).HasCode {
		t.Error("HasCode = false for an indented code block, want true")
	}
	if GeneratePreview([]byte("Just call `foo()` inline."), 0).HasCode {
		t.Error("HasCode = true for an inline code span, want false")
	}
}

// TestGeneratePreviewCountsPastTheExcerptBudget is the key behavioral
// guarantee behind counting TodoTotal/HasImage/HasCode over the whole
// document instead of stopping wherever the excerpt text itself stops: a
// long lead paragraph that already fills the excerpt budget must not hide a
// checklist/image/code block that comes later in the same note.
func TestGeneratePreviewCountsPastTheExcerptBudget(t *testing.T) {
	longLead := strings.Repeat("填充正文用的长段落。", 60) // well past PreviewMaxRunes on its own
	content := longLead + "\n\n- [ ] a task after the excerpt cutoff\n\n![img](x.png)\n\n```\ncode\n```"

	sum := GeneratePreview([]byte(content), 0)
	if sum.TodoTotal != 1 {
		t.Errorf("TodoTotal = %d, want 1 (task list past the excerpt budget must still be counted)", sum.TodoTotal)
	}
	if !sum.HasImage {
		t.Error("HasImage = false, want true (image past the excerpt budget must still be counted)")
	}
	if !sum.HasCode {
		t.Error("HasCode = false, want true (code block past the excerpt budget must still be counted)")
	}
}
