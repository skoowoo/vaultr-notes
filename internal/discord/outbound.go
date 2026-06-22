package discord

import "strings"

// FormatForDiscord returns text suitable for Discord delivery.
// Discord renders Markdown natively, so only whitespace normalisation is needed.
func FormatForDiscord(text string) string {
	return strings.TrimSpace(text)
}

// SplitText splits text into segments of at most maxLen characters, preferring line breaks.
func SplitText(text string, maxLen int) []string {
	if maxLen <= 0 || len(text) <= maxLen {
		return []string{text}
	}
	var segments []string
	remaining := text
	for len(remaining) > 0 {
		if len(remaining) <= maxLen {
			segments = append(segments, remaining)
			break
		}
		breakAt := lastIndexByte(remaining, '\n', maxLen)
		if breakAt <= 0 {
			breakAt = maxLen
		}
		segments = append(segments, remaining[:breakAt])
		remaining = remaining[breakAt:]
		if len(remaining) > 0 && remaining[0] == '\n' {
			remaining = remaining[1:]
		}
	}
	return segments
}

func lastIndexByte(s string, b byte, limit int) int {
	if limit > len(s) {
		limit = len(s)
	}
	for i := limit - 1; i >= 0; i-- {
		if s[i] == b {
			return i
		}
	}
	return -1
}
