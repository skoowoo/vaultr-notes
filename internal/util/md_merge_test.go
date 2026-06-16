package util

import (
	"testing"
)

func TestMdJoin(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		incoming string
		want     string
	}{
		// ── empty existing ────────────────────────────────────────────────────
		{
			name:     "empty file gets content as-is",
			existing: "",
			incoming: "# Hello\n",
			want:     "# Hello\n",
		},
		{
			name:     "empty file: trailing newline added to incoming",
			existing: "",
			incoming: "hello",
			want:     "hello\n",
		},
		{
			name:     "whitespace-only file treated as empty",
			existing: "   \n\n  \n",
			incoming: "# Title\n",
			want:     "# Title\n",
		},

		// ── paragraph → paragraph ────────────────────────────────────────────
		{
			name:     "paragraph appended with blank line",
			existing: "First paragraph.\n",
			incoming: "Second paragraph.\n",
			want:     "First paragraph.\n\nSecond paragraph.\n",
		},
		{
			name:     "existing already ends with one blank line",
			existing: "First paragraph.\n\n",
			incoming: "Second paragraph.\n",
			want:     "First paragraph.\n\nSecond paragraph.\n",
		},
		{
			name:     "existing ends with two blank lines — no extra added",
			existing: "First paragraph.\n\n\n",
			incoming: "Second paragraph.\n",
			want:     "First paragraph.\n\n\nSecond paragraph.\n",
		},
		{
			name:     "existing has no trailing newline",
			existing: "First paragraph.",
			incoming: "Second paragraph.\n",
			want:     "First paragraph.\n\nSecond paragraph.\n",
		},

		// ── heading transitions ───────────────────────────────────────────────
		{
			name:     "heading appended after paragraph",
			existing: "Some text.\n",
			incoming: "## Section\n",
			want:     "Some text.\n\n## Section\n",
		},
		{
			name:     "paragraph appended after heading",
			existing: "# Title\n",
			incoming: "Intro text.\n",
			want:     "# Title\n\nIntro text.\n",
		},

		// ── list continuation ─────────────────────────────────────────────────
		{
			name:     "unordered list item continues list (single newline)",
			existing: "- item 1\n",
			incoming: "- item 2\n",
			want:     "- item 1\n- item 2\n",
		},
		{
			name:     "asterisk list item continues list",
			existing: "* item 1\n",
			incoming: "* item 2\n",
			want:     "* item 1\n* item 2\n",
		},
		{
			name:     "ordered list item continues list",
			existing: "1. first\n",
			incoming: "2. second\n",
			want:     "1. first\n2. second\n",
		},
		{
			name:     "list already ends with newline — no extra added",
			existing: "- item 1\n",
			incoming: "- item 2\n",
			want:     "- item 1\n- item 2\n",
		},
		{
			name:     "mixed list markers still treated as continuation",
			existing: "- item 1\n",
			incoming: "* item 2\n",
			want:     "- item 1\n* item 2\n",
		},

		// ── list → non-list and vice versa ────────────────────────────────────
		{
			name:     "paragraph after list gets blank line",
			existing: "- item 1\n- item 2\n",
			incoming: "Closing remark.\n",
			want:     "- item 1\n- item 2\n\nClosing remark.\n",
		},
		{
			name:     "list after paragraph gets blank line",
			existing: "Intro.\n",
			incoming: "- item 1\n",
			want:     "Intro.\n\n- item 1\n",
		},

		// ── incoming without trailing newline ─────────────────────────────────
		{
			name:     "incoming without newline gets one appended",
			existing: "# Title\n",
			incoming: "body text",
			want:     "# Title\n\nbody text\n",
		},

		// ── multi-line incoming ───────────────────────────────────────────────
		{
			name:     "multi-line incoming block",
			existing: "# Title\n",
			incoming: "Line one.\nLine two.\n",
			want:     "# Title\n\nLine one.\nLine two.\n",
		},
		{
			name:     "incoming starting with blank line: first non-empty line drives decision",
			existing: "- item 1\n",
			incoming: "\n- item 2\n",
			want:     "- item 1\n\n- item 2\n",
		},

		// ── YAML frontmatter (append is unaffected — frontmatter is at the top) ──
		{
			name:     "frontmatter doc: append paragraph after body",
			existing: "---\ntitle: Note\n---\n\n# Title\n\nFirst paragraph.\n",
			incoming: "Second paragraph.\n",
			want:     "---\ntitle: Note\n---\n\n# Title\n\nFirst paragraph.\n\nSecond paragraph.\n",
		},
		{
			name:     "frontmatter doc: append list item continues list",
			existing: "---\ntitle: Note\n---\n\n# Title\n\n- item 1\n",
			incoming: "- item 2\n",
			want:     "---\ntitle: Note\n---\n\n# Title\n\n- item 1\n- item 2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(MdAppend([]byte(tt.existing), []byte(tt.incoming)))
			if got != tt.want {
				t.Errorf("\nexisting: %q\nincoming: %q\n    got:  %q\n   want:  %q",
					tt.existing, tt.incoming, got, tt.want)
			}
		})
	}
}

func TestMdPrepend(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		incoming string
		want     string
	}{
		// ── empty / no-content existing ───────────────────────────────────────
		{
			name:     "empty file returns incoming",
			existing: "",
			incoming: "new line\n",
			want:     "new line\n",
		},
		{
			name:     "whitespace-only file returns incoming",
			existing: "   \n\n",
			incoming: "new line\n",
			want:     "new line\n",
		},

		// ── no H1: prepend at the very beginning ─────────────────────────────
		{
			name:     "no H1: inserted before existing content",
			existing: "Some paragraph.\n",
			incoming: "Prepended.\n",
			want:     "Prepended.\n\nSome paragraph.\n",
		},
		{
			name:     "no H1: incoming without trailing newline",
			existing: "Body.\n",
			incoming: "Note",
			want:     "Note\n\nBody.\n",
		},

		// ── H1 present: insert after it ──────────────────────────────────────
		{
			name:     "H1 only: incoming appended after title",
			existing: "# Title\n",
			incoming: "First content.\n",
			want:     "# Title\n\nFirst content.\n",
		},
		{
			name:     "H1 + body: incoming inserted between title and body",
			existing: "# Title\n\nOld content.\n",
			incoming: "New content.\n",
			want:     "# Title\n\nNew content.\n\nOld content.\n",
		},
		{
			name:     "H1 + body: incoming without trailing newline",
			existing: "# Title\n\nOld content.\n",
			incoming: "Note",
			want:     "# Title\n\nNote\n\nOld content.\n",
		},

		// ── list continuation across H1 boundary ─────────────────────────────
		{
			name:     "existing list after H1: bullet list continues when incoming is same kind",
			existing: "# Title\n\n- old item\n",
			incoming: "- new item\n",
			want:     "# Title\n\n- new item\n- old item\n",
		},
		{
			name:     "incoming ordered list does not merge with existing bullet list",
			existing: "# Title\n\n- old item\n",
			incoming: "1. new item\n",
			want:     "# Title\n\n1. new item\n\n- old item\n",
		},

		// ── H2 ignored, only first H1 counts ─────────────────────────────────
		{
			name:     "H2 before H1 is not treated as insertion point",
			existing: "## Section\n\n# Title\n\nBody.\n",
			incoming: "Note.\n",
			want:     "## Section\n\n# Title\n\nNote.\n\nBody.\n",
		},
		{
			name:     "only first H1 is used when there are multiple H1s",
			existing: "# First\n\nMiddle.\n\n# Second\n\nEnd.\n",
			incoming: "Inserted.\n",
			want:     "# First\n\nInserted.\n\nMiddle.\n\n# Second\n\nEnd.\n",
		},

		// ── YAML frontmatter ──────────────────────────────────────────────────
		{
			name:     "frontmatter + H1: insert after H1",
			existing: "---\ntitle: My Note\ndate: 2024-01-01\n---\n\n# Title\n\nBody.\n",
			incoming: "New content.\n",
			want:     "---\ntitle: My Note\ndate: 2024-01-01\n---\n\n# Title\n\nNew content.\n\nBody.\n",
		},
		{
			name:     "frontmatter + no H1: insert after frontmatter",
			existing: "---\ntitle: My Note\n---\n\nBody.\n",
			incoming: "Prepended.\n",
			want:     "---\ntitle: My Note\n---\n\nPrepended.\n\nBody.\n",
		},
		{
			name:     "frontmatter with no body: insert after frontmatter",
			existing: "---\ntitle: My Note\n---\n",
			incoming: "First content.\n",
			want:     "---\ntitle: My Note\n---\n\nFirst content.\n",
		},
		{
			name:     "frontmatter + H1 + list body: bullet list continues",
			existing: "---\ntags: [go]\n---\n\n# Title\n\n- old item\n",
			incoming: "- new item\n",
			want:     "---\ntags: [go]\n---\n\n# Title\n\n- new item\n- old item\n",
		},
		{
			name:     "frontmatter closed with ...",
			existing: "---\ntitle: Note\n...\n\n# Title\n\nContent.\n",
			incoming: "Inserted.\n",
			want:     "---\ntitle: Note\n...\n\n# Title\n\nInserted.\n\nContent.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(MdPrepend([]byte(tt.existing), []byte(tt.incoming)))
			if got != tt.want {
				t.Errorf("\nexisting: %q\nincoming: %q\n    got:  %q\n   want:  %q",
					tt.existing, tt.incoming, got, tt.want)
			}
		})
	}
}

func TestMdJoinSection(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		heading  string
		incoming string
		want     string
	}{
		// ── no match: falls back to MdJoin ───────────────────────────────────
		{
			name:     "no matching heading: appends at end",
			existing: "# Title\n\nBody.\n",
			heading:  "Notes",
			incoming: "new line\n",
			want:     "# Title\n\nBody.\n\nnew line\n",
		},

		// ── section at end of document ────────────────────────────────────────
		{
			name:     "section at end: appends inside section",
			existing: "# Title\n\n## Notes\n\n- note 1\n",
			heading:  "Notes",
			incoming: "- note 2\n",
			want:     "# Title\n\n## Notes\n\n- note 1\n- note 2\n",
		},
		{
			name:     "empty section at end: appends with blank line after heading",
			existing: "# Title\n\n## Notes\n",
			heading:  "Notes",
			incoming: "first note\n",
			want:     "# Title\n\n## Notes\n\nfirst note\n",
		},
		{
			name:     "paragraph appended in section",
			existing: "# Title\n\n## Notes\n\nExisting.\n",
			heading:  "Notes",
			incoming: "New paragraph.\n",
			want:     "# Title\n\n## Notes\n\nExisting.\n\nNew paragraph.\n",
		},

		// ── section in middle: next sibling heading is preserved ──────────────
		{
			name:     "section in middle: appended before next sibling",
			existing: "# Title\n\n## Notes\n\nOld note.\n\n## Other\n\nOther content.\n",
			heading:  "Notes",
			incoming: "New note.\n",
			want:     "# Title\n\n## Notes\n\nOld note.\n\nNew note.\n\n## Other\n\nOther content.\n",
		},
		{
			name:     "subsections stay inside target section",
			existing: "## Notes\n\nText.\n\n### Sub\n\nSub text.\n\n## Other\n",
			heading:  "Notes",
			incoming: "Appended.\n",
			want:     "## Notes\n\nText.\n\n### Sub\n\nSub text.\n\nAppended.\n\n## Other\n",
		},

		// ── multiple matching headings: last one is used ──────────────────────
		{
			name:     "multiple matches: last heading is targeted",
			existing: "## Notes\n\nFirst.\n\n## Notes\n\nSecond.\n",
			heading:  "notes",
			incoming: "Third.\n",
			want:     "## Notes\n\nFirst.\n\n## Notes\n\nSecond.\n\nThird.\n",
		},

		// ── case-insensitive match ────────────────────────────────────────────
		{
			name:     "case-insensitive heading match",
			existing: "## NOTES\n\nItem.\n",
			heading:  "notes",
			incoming: "New item.\n",
			want:     "## NOTES\n\nItem.\n\nNew item.\n",
		},

		// ── YAML frontmatter ──────────────────────────────────────────────────
		{
			name:     "frontmatter transparent",
			existing: "---\ntitle: Doc\n---\n\n## Notes\n\nOld.\n",
			heading:  "Notes",
			incoming: "New.\n",
			want:     "---\ntitle: Doc\n---\n\n## Notes\n\nOld.\n\nNew.\n",
		},

		// ── list continuation inside section ──────────────────────────────────
		{
			name:     "bullet list continues inside section",
			existing: "## Log\n\n- entry 1\n",
			heading:  "Log",
			incoming: "- entry 2\n",
			want:     "## Log\n\n- entry 1\n- entry 2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(MdAppendSection([]byte(tt.existing), []byte(tt.incoming), tt.heading))
			if got != tt.want {
				t.Errorf("\nexisting: %q\nheading:  %q\nincoming: %q\n    got:  %q\n   want:  %q",
					tt.existing, tt.heading, tt.incoming, got, tt.want)
			}
		})
	}
}

func TestMdPrependSection(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		heading  string
		incoming string
		want     string
	}{
		// ── no match: falls back to MdPrepend ────────────────────────────────
		{
			name:     "no matching heading: falls back to MdPrepend (after H1)",
			existing: "# Title\n\nBody.\n",
			heading:  "Notes",
			incoming: "inserted\n",
			want:     "# Title\n\ninserted\n\nBody.\n",
		},

		// ── section at end ────────────────────────────────────────────────────
		{
			name:     "prepend to section at end",
			existing: "## Notes\n\n- old note\n",
			heading:  "Notes",
			incoming: "- new note\n",
			want:     "## Notes\n\n- new note\n- old note\n",
		},
		{
			name:     "prepend paragraph to section",
			existing: "## Notes\n\nExisting.\n",
			heading:  "Notes",
			incoming: "New.\n",
			want:     "## Notes\n\nNew.\n\nExisting.\n",
		},

		// ── section in middle: content after section is preserved ─────────────
		{
			name:     "prepend to section in middle; rest of doc untouched",
			existing: "## Notes\n\nOld.\n\n## Other\n\nOther.\n",
			heading:  "Notes",
			incoming: "New.\n",
			want:     "## Notes\n\nNew.\n\nOld.\n\n## Other\n\nOther.\n",
		},

		// ── multiple matches: first is used ──────────────────────────────────
		{
			name:     "multiple matches: first heading is targeted",
			existing: "## Notes\n\nFirst.\n\n## Notes\n\nSecond.\n",
			heading:  "Notes",
			incoming: "Zero.\n",
			want:     "## Notes\n\nZero.\n\nFirst.\n\n## Notes\n\nSecond.\n",
		},

		// ── case-insensitive ──────────────────────────────────────────────────
		{
			name:     "case-insensitive match",
			existing: "## NOTES\n\nOld.\n",
			heading:  "notes",
			incoming: "New.\n",
			want:     "## NOTES\n\nNew.\n\nOld.\n",
		},

		// ── YAML frontmatter ──────────────────────────────────────────────────
		{
			name:     "frontmatter transparent",
			existing: "---\ntitle: Doc\n---\n\n## Notes\n\nOld.\n",
			heading:  "Notes",
			incoming: "New.\n",
			want:     "---\ntitle: Doc\n---\n\n## Notes\n\nNew.\n\nOld.\n",
		},

		// ── list continuation ─────────────────────────────────────────────────
		{
			name:     "bullet list continues when prepending same list kind",
			existing: "## Log\n\n- old entry\n",
			heading:  "Log",
			incoming: "- new entry\n",
			want:     "## Log\n\n- new entry\n- old entry\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(MdPrependSection([]byte(tt.existing), []byte(tt.incoming), tt.heading))
			if got != tt.want {
				t.Errorf("\nexisting: %q\nheading:  %q\nincoming: %q\n    got:  %q\n   want:  %q",
					tt.existing, tt.heading, tt.incoming, got, tt.want)
			}
		})
	}
}
