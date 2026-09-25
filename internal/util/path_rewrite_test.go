package util

import (
	"bytes"
	"testing"
)

func TestRewriteFrontmatterPathRef(t *testing.T) {
	raw := []byte("---\nkind: knowledge\nsource_notes:\n  - /journal/2026/april.md\n  - other.md\n---\n\nBody mentions /journal/2026/april.md too.\n")

	out, changed := RewriteFrontmatterPathRef(raw, "/journal/2026/april.md", "/archive/april.md")
	if !changed {
		t.Fatal("expected a change")
	}
	if bytes.Count(out, []byte("/journal/2026/april.md")) != 1 {
		t.Fatalf("body occurrence should survive untouched: %s", out)
	}
	if !bytes.Contains(out, []byte("- /archive/april.md")) {
		t.Fatalf("frontmatter occurrence should be rewritten: %s", out)
	}
	if !bytes.Contains(out, []byte("- other.md")) {
		t.Fatalf("unrelated entry should be untouched: %s", out)
	}
}

func TestRewriteFrontmatterPathRefNoMatch(t *testing.T) {
	raw := []byte("---\nkind: knowledge\nsource_notes:\n  - other.md\n---\n\nbody\n")
	out, changed := RewriteFrontmatterPathRef(raw, "/journal/2026/april.md", "/archive/april.md")
	if changed {
		t.Fatal("expected no change")
	}
	if !bytes.Equal(out, raw) {
		t.Fatalf("raw should be returned unmodified, got: %s", out)
	}
}

func TestRewriteFrontmatterPathRefNoFrontmatter(t *testing.T) {
	raw := []byte("# just a note\n\nmentions /journal/2026/april.md in prose\n")
	out, changed := RewriteFrontmatterPathRef(raw, "/journal/2026/april.md", "/archive/april.md")
	if changed {
		t.Fatal("expected no change without front matter")
	}
	if !bytes.Equal(out, raw) {
		t.Fatalf("raw should be returned unmodified, got: %s", out)
	}
}

func TestRewriteTableRowPathRef(t *testing.T) {
	raw := []byte("---\nkind: index\n---\n\n" +
		"See /_knowledge/Vertical Agent.md in the intro.\n\n" +
		"| Title          | Path                          |\n" +
		"| -------------- | ----------------------------- |\n" +
		"| Vertical Agent | /_knowledge/Vertical Agent.md |\n")

	out, changed := RewriteTableRowPathRef(raw, "/_knowledge/Vertical Agent.md", "/_knowledge/Agents/Vertical Agent.md")
	if !changed {
		t.Fatal("expected a change")
	}
	if bytes.Count(out, []byte("/_knowledge/Vertical Agent.md")) != 1 {
		t.Fatalf("prose mention should survive untouched: %s", out)
	}
	if !bytes.Contains(out, []byte("| /_knowledge/Agents/Vertical Agent.md |")) {
		t.Fatalf("table cell should be rewritten: %s", out)
	}
}

func TestRewriteTableRowPathRefNoMatch(t *testing.T) {
	raw := []byte("| Title | Path |\n| --- | --- |\n| A | /_knowledge/A.md |\n")
	out, changed := RewriteTableRowPathRef(raw, "/_knowledge/B.md", "/_knowledge/moved/B.md")
	if changed {
		t.Fatal("expected no change")
	}
	if !bytes.Equal(out, raw) {
		t.Fatalf("raw should be returned unmodified, got: %s", out)
	}
}
