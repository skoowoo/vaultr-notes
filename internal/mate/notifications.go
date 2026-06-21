package mate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RunNotification is pushed when a trigger agent run starts or finishes.
type RunNotification struct {
	Type        string  `json:"type"`                  // "run_start" | "run_done"
	MateID      string  `json:"mateId"`
	MateName    string  `json:"mateName"`
	ConvID      string  `json:"convId,omitempty"`
	EventType   string  `json:"eventType,omitempty"`   // mate event type, e.g. "note_created"
	Success     bool    `json:"success,omitempty"`
	DurationSec float64 `json:"durationSec,omitempty"`
	LastMessage string  `json:"lastMessage,omitempty"`
}

// NotificationBus is a fan-out broadcaster for mate run notifications.
// Push is non-blocking and drops events for slow subscribers.
type NotificationBus struct {
	mu   sync.Mutex
	subs map[chan RunNotification]struct{}
}

// NewNotificationBus creates a new bus.
func NewNotificationBus() *NotificationBus {
	return &NotificationBus{subs: make(map[chan RunNotification]struct{})}
}

// Push broadcasts n to all current subscribers without blocking.
func (b *NotificationBus) Push(n RunNotification) {
	b.mu.Lock()
	subs := make([]chan RunNotification, 0, len(b.subs))
	for ch := range b.subs {
		subs = append(subs, ch)
	}
	b.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- n:
		default:
		}
	}
}

func (b *NotificationBus) subscribe() chan RunNotification {
	ch := make(chan RunNotification, 16)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *NotificationBus) unsubscribe(ch chan RunNotification) {
	b.mu.Lock()
	delete(b.subs, ch)
	b.mu.Unlock()
}

// StreamSSE serves GET /api/mate/run-notifications as a Server-Sent Events stream.
// Each event has a named type ("run_start" or "run_done") and a JSON data payload.
func (b *NotificationBus) StreamSSE(w http.ResponseWriter, r *http.Request) {
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
		case n := <-ch:
			data, err := json.Marshal(n)
			if err != nil {
				continue
			}
			id++
			if _, err := fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", id, n.Type, data); err != nil {
				return
			}
			fl.Flush()
		}
	}
}
