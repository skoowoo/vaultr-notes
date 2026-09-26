package util

import (
	"bytes"
	"testing"
)

func TestRewriteWikilinkTargetBasic(t *testing.T) {
	raw := []byte("see [[Old Name]] for details")
	out, changed := RewriteWikilinkTarget(raw, "Old Name.md", "New Name.md")
	if !changed {
		t.Fatal("expected a change")
	}
	if string(out) != "see [[New Name]] for details" {
		t.Fatalf("got %q", out)
	}
}

func TestRewriteWikilinkTargetIgnoresExtensionOnBothSides(t *testing.T) {
	cases := []struct{ raw, old, new_, want string }{
		{"[[Old]]", "Old.md", "New.md", "[[New]]"},
		{"[[Old.md]]", "Old", "New", "[[New]]"},
		{"[[Old.md]]", "Old.md", "New.md", "[[New]]"},
	}
	for _, c := range cases {
		out, changed := RewriteWikilinkTarget([]byte(c.raw), c.old, c.new_)
		if !changed || string(out) != c.want {
			t.Fatalf("RewriteWikilinkTarget(%q, %q, %q) = %q, %v; want %q", c.raw, c.old, c.new_, out, changed, c.want)
		}
	}
}

func TestRewriteWikilinkTargetPreservesDisplayText(t *testing.T) {
	raw := []byte("[[Old|shown text]]")
	out, changed := RewriteWikilinkTarget(raw, "Old.md", "New.md")
	if !changed {
		t.Fatal("expected a change")
	}
	if string(out) != "[[New|shown text]]" {
		t.Fatalf("got %q", out)
	}
}

func TestRewriteWikilinkTargetMultipleOccurrences(t *testing.T) {
	raw := []byte("[[Old]] and again [[Old|alias]] and [[Other]]")
	out, changed := RewriteWikilinkTarget(raw, "Old.md", "New.md")
	if !changed {
		t.Fatal("expected a change")
	}
	want := "[[New]] and again [[New|alias]] and [[Other]]"
	if string(out) != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestRewriteWikilinkTargetSkipsImageEmbeds(t *testing.T) {
	// ![[Old]] is an image embed in this app's markdown dialect, never a note
	// reference — it must survive untouched even though the inner target
	// text matches oldName.
	raw := []byte("![[Old]] and [[Old]]")
	out, changed := RewriteWikilinkTarget(raw, "Old.md", "New.md")
	if !changed {
		t.Fatal("expected a change (the non-embed occurrence)")
	}
	want := "![[Old]] and [[New]]"
	if string(out) != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestRewriteWikilinkTargetNoMatch(t *testing.T) {
	raw := []byte("[[Other]] unrelated")
	out, changed := RewriteWikilinkTarget(raw, "Old.md", "New.md")
	if changed {
		t.Fatal("expected no change")
	}
	if !bytes.Equal(out, raw) {
		t.Fatalf("raw should be returned unmodified, got: %s", out)
	}
}

func TestRewriteWikilinkTargetNoopWhenNamesEquivalent(t *testing.T) {
	raw := []byte("[[Old]]")
	out, changed := RewriteWikilinkTarget(raw, "Old.md", "Old")
	if changed {
		t.Fatal("expected no change when old/new stems are equal")
	}
	if !bytes.Equal(out, raw) {
		t.Fatalf("raw should be returned unmodified, got: %s", out)
	}
}

func TestRewriteFrontmatterBareNameRef(t *testing.T) {
	raw := []byte("---\nkind: knowledge\nsource_notes:\n  - old\n  - /journal/other.md\n  - unrelated.md\n---\n\nBody mentions old.md too.\n")
	out, changed := RewriteFrontmatterBareNameRef(raw, "old.md", "new.md")
	if !changed {
		t.Fatal("expected a change")
	}
	if !bytes.Contains(out, []byte("- new\n")) {
		t.Fatalf("bare entry should be rewritten (keeping bare style): %s", out)
	}
	if !bytes.Contains(out, []byte("- /journal/other.md")) {
		t.Fatalf("path-shaped entry should be left alone: %s", out)
	}
	if !bytes.Contains(out, []byte("- unrelated.md")) {
		t.Fatalf("unrelated entry should be left alone: %s", out)
	}
	if !bytes.Contains(out, []byte("Body mentions old.md too.")) {
		t.Fatalf("body text should never be touched by a frontmatter-only rewrite: %s", out)
	}
}

func TestRewriteFrontmatterBareNameRefNoMatch(t *testing.T) {
	raw := []byte("---\nkind: knowledge\nsource_notes:\n  - other.md\n---\n\nbody\n")
	out, changed := RewriteFrontmatterBareNameRef(raw, "old.md", "new.md")
	if changed {
		t.Fatal("expected no change")
	}
	if !bytes.Equal(out, raw) {
		t.Fatalf("raw should be returned unmodified, got: %s", out)
	}
}

func TestRewriteFrontmatterBareNameRefStopsAtListEnd(t *testing.T) {
	// A later, unrelated top-level key must not be mistaken for a further
	// list item, and must not itself be scanned as one.
	raw := []byte("---\nsource_notes:\n  - old\ntags: [old]\n---\n\nbody\n")
	out, changed := RewriteFrontmatterBareNameRef(raw, "old.md", "new.md")
	if !changed {
		t.Fatal("expected a change")
	}
	if !bytes.Contains(out, []byte("tags: [old]")) {
		t.Fatalf("flow-style fields after the list must be left untouched: %s", out)
	}
}
