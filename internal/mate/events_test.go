package mate

import "testing"

func TestEventLabel(t *testing.T) {
	if EventLabel(MateEventScheduled) != "Scheduled" {
		t.Fatal("expected Scheduled label")
	}
	if EventLabel(MateEventType("unknown")) != "unknown" {
		t.Fatal("fallback to type string")
	}
}
