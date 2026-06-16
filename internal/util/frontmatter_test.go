package util

import (
	"strings"
	"testing"
	"time"
)

func TestFormatKnowledgeStyleFrontmatter(t *testing.T) {
	ts := time.Date(2026, 4, 12, 15, 30, 0, 0, time.UTC)
	got := FormatKnowledgeStyleFrontmatter([]string{"foo bar", "baz"}, ts)
	want := strings.TrimSpace(`
---
tags:
  - foo-bar
  - baz
compiled: 2026-04-12
---
`)
	if strings.TrimSpace(got) != want {
		t.Fatalf("unexpected output:\n%s", got)
	}
}

func TestFormatKnowledgeStyleFrontmatter_ParseRoundTrip(t *testing.T) {
	ts := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	block := FormatKnowledgeStyleFrontmatter([]string{"a", "b c"}, ts)
	fm, body := ParseFrontmatter([]byte(block + "# hi\n"))
	if len(fm.Tags) != 2 || fm.Tags[0] != "a" || fm.Tags[1] != "b-c" {
		t.Fatalf("tags: %#v", fm.Tags)
	}
	if fm.Compiled != "2026-01-02" {
		t.Fatalf("compiled: %q", fm.Compiled)
	}
	if strings.TrimSpace(string(body)) != "# hi" {
		t.Fatalf("body: %q", body)
	}
}
