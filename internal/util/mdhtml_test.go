package util

import (
	"strings"
	"testing"
)

func TestMarkdownToHTMLFragment(t *testing.T) {
	out, err := MarkdownToHTMLFragment([]byte("# Hi\n\n**bold**"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, "<h1") || !strings.Contains(s, "bold") {
		t.Fatalf("unexpected HTML: %s", s)
	}
}

func TestMarkdownToHTMLFragmentExpandsWikilinks(t *testing.T) {
	out, err := MarkdownToHTMLFragment([]byte("[[My Note|read this]]"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, `href="/notes?name=My+Note.md"`) || !strings.Contains(s, "read this") {
		t.Fatalf("unexpected wikilink HTML: %s", s)
	}
}
