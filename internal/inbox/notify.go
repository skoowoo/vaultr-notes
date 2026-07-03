package inbox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Bus is a fan-out broadcaster that fires whenever a new Message is created,
// so subscribers (e.g. the desktop app) can show a real-time notification
// without polling. Push is non-blocking and drops events for slow subscribers.
type Bus struct {
	mu   sync.Mutex
	subs map[chan Message]struct{}
}

func newBus() *Bus {
	return &Bus{subs: make(map[chan Message]struct{})}
}

func (b *Bus) push(m Message) {
	b.mu.Lock()
	subs := make([]chan Message, 0, len(b.subs))
	for ch := range b.subs {
		subs = append(subs, ch)
	}
	b.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- m:
		default:
		}
	}
}

func (b *Bus) subscribe() chan Message {
	ch := make(chan Message, 16)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Bus) unsubscribe(ch chan Message) {
	b.mu.Lock()
	delete(b.subs, ch)
	b.mu.Unlock()
}

// StreamSSE serves GET /api/inbox/notifications as a Server-Sent Events
// stream: one "message" event per newly created inbox message, regardless
// of source.
func (b *Bus) StreamSSE(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	fl.Flush()

	ch := b.subscribe()
	defer b.unsubscribe(ch)

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	id := 0
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if _, err := fmt.Fprintf(w, ": ping\n\n"); err != nil {
				return
			}
			fl.Flush()
		case m := <-ch:
			data, err := json.Marshal(m)
			if err != nil {
				continue
			}
			id++
			if _, err := fmt.Fprintf(w, "id: %d\nevent: message\ndata: %s\n\n", id, data); err != nil {
				return
			}
			fl.Flush()
		}
	}
}
