package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// slowFirstWriteRecorder delays the first Write so StreamSSE blocks on the HTTP body
// while Emit can fill the 128-slot subscriber channel and drop the terminal "end".
type slowFirstWriteRecorder struct {
	*httptest.ResponseRecorder
	mu     sync.Mutex
	writes int
	delay  time.Duration
}

func (s *slowFirstWriteRecorder) Write(p []byte) (int, error) {
	s.mu.Lock()
	s.writes++
	n := s.writes
	s.mu.Unlock()
	if n == 1 {
		time.Sleep(s.delay)
	}
	return s.ResponseRecorder.Write(p)
}

func (s *slowFirstWriteRecorder) Flush() {
	s.ResponseRecorder.Flush()
}

func TestStreamSSE_flushesEndAfterLiveDrop(t *testing.T) {
	h := NewHub()
	run := h.CreateRun(nil)

	rec := &slowFirstWriteRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		delay:            400 * time.Millisecond,
	}
	req := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(context.Background())

	done := make(chan struct{})
	go func() {
		h.StreamSSE(rec, req, run)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	// First event is read then blocked on Write; E2..E129 fill the 128-buffer; Finish's
	// "end" is dropped by Emit's non-blocking send.
	for i := 0; i < 129; i++ {
		h.Emit(run, "agent", map[string]any{"i": i})
	}
	h.Finish(run, "succeeded", map[string]any{"exitCode": 0})

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("StreamSSE did not finish")
	}

	body := rec.Body.String()
	if !strings.Contains(body, "event: end") {
		t.Fatalf("expected terminal end after flush from run.Events; got:\n%s", body)
	}
}
