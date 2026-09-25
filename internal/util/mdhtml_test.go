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

func TestMarkdownToHTMLFragmentCheckedMarksBrokenWikilinks(t *testing.T) {
	exists := func(name string) bool { return name == "Alive.md" }

	out, err := MarkdownToHTMLFragmentChecked([]byte("[[Alive]] and [[Gone|missing]]"), exists)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, `href="/notes?name=Alive.md"`) {
		t.Fatalf("existing wikilink should still be a link: %s", s)
	}
	if strings.Contains(s, `href="/notes?name=Gone.md"`) {
		t.Fatalf("broken wikilink must not be a navigable link: %s", s)
	}
	if !strings.Contains(s, `<span class="wikilink-broken"`) || !strings.Contains(s, "missing</span>") {
		t.Fatalf("broken wikilink should render as a non-link span: %s", s)
	}
}

func TestMarkdownToHTMLFragmentCheckedNilExistsBehavesUnchecked(t *testing.T) {
	out, err := MarkdownToHTMLFragmentChecked([]byte("[[Anything]]"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `href="/notes?name=Anything.md"`) {
		t.Fatalf("nil exists func should treat every target as existing: %s", out)
	}
}
