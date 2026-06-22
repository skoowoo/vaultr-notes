package discordplugin

import (
	"context"
	"time"
)

const typingRefreshInterval = 8 * time.Second

// startTypingIndicator sends a typing indicator to channelID immediately and
// refreshes every typingRefreshInterval until the returned cancel is called.
// Discord's typing indicator lasts ~10 seconds, so 8 s keeps it continuous.
func (p *Plugin) startTypingIndicator(channelID string) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		p.sendTyping(channelID)
		ticker := time.NewTicker(typingRefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.sendTyping(channelID)
			}
		}
	}()
	return cancel
}

func (p *Plugin) sendTyping(channelID string) {
	if p.session == nil {
		return
	}
	_ = p.session.ChannelTyping(channelID)
}
