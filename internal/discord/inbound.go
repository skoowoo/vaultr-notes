package discord

import "strings"

// ExtractText returns the user-visible text from a Discord message.
func ExtractText(msg Message) string {
	return strings.TrimSpace(msg.Content)
}

// IsSessionCommand reports whether the text is a session-reset command (/new or /clear).
func IsSessionCommand(text string) bool {
	t := strings.ToLower(strings.TrimSpace(text))
	return t == "/new" || t == "/clear"
}

// PreviewMessage returns a short log preview of the message content.
func PreviewMessage(msg Message) string {
	text := ExtractText(msg)
	if text == "" {
		return "[empty]"
	}
	if len(text) > 50 {
		return text[:50] + "..."
	}
	return text
}
