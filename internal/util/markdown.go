// Package util provides shared utility functions used across internal packages.
package util

import (
	"bytes"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// markdownExts is the canonical set of accepted markdown file extensions.
var markdownExts = map[string]bool{
	".md":       true,
	".markdown": true,
}

// IsMarkdownPath reports whether path has a recognised markdown extension.
func IsMarkdownPath(path string) bool {
	return markdownExts[strings.ToLower(filepath.Ext(path))]
}

// IsValidText reports whether data is valid UTF-8 text (i.e. not binary).
func IsValidText(data []byte) bool {
	return utf8.Valid(data)
}

// StripMarkdownEscapes removes backslash-escape sequences from a string that
// was read verbatim from a markdown cell or other inline context.
// In CommonMark, a backslash before any ASCII punctuation is an escape and
// the backslash is dropped. Vault paths are Unix paths and never contain a
// literal backslash, so every '\' found in a path cell is a markdown escape.
func StripMarkdownEscapes(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	escaped := false
	for _, c := range s {
		switch {
		case escaped:
			b.WriteRune(c)
			escaped = false
		case c == '\\':
			escaped = true
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}

// CountLines returns the number of lines in data.
// A trailing newline does not count as an extra empty line.
func CountLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	n := bytes.Count(data, []byte{'\n'})
	// If the last byte is not a newline there is one more unterminated line.
	if data[len(data)-1] != '\n' {
		n++
	}
	return n
}
