// Package discord implements the Discord Bot Gateway bridge.
package discord

// TextChunkLimit is the maximum characters per Discord message.
const TextChunkLimit = 2000

// Message is a normalised inbound Discord message.
type Message struct {
	ChannelID string
	MessageID string
	AuthorID  string
	Content   string
	GuildID   string // empty for DMs
}
