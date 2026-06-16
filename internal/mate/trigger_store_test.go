package mate

import (
	"path/filepath"
	"testing"
	"time"
)

func TestScheduledTriggerPersistence(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	m := &Mate{Name: "Digest", AgentID: "claude", Enabled: true}
	if err := store.CreateMate(m); err != nil {
		t.Fatal(err)
	}

	last := time.Date(2026, 5, 22, 9, 0, 0, 0, time.UTC)
	triggers := []MateTrigger{{
		ID:         "tr-1",
		EventTypes: []string{string(MateEventScheduled)},
		Schedule:   "daily 09:00",
		Prompt:     "digest",
		Enabled:    true,
	}}
	if err := store.ReplaceTriggersForMate(m.ID, triggers); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateTriggerLastFiredAt("tr-1", last); err != nil {
		t.Fatal(err)
	}

	got, err := store.ListScheduledTriggers()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Schedule != "daily 09:00" {
		t.Fatalf("list scheduled: %+v", got)
	}

	// Same schedule preserves last_fired_at on replace.
	triggers[0].Prompt = "updated"
	if err := store.ReplaceTriggersForMate(m.ID, triggers); err != nil {
		t.Fatal(err)
	}
	reloaded, err := store.ListTriggers(m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reloaded[0].LastFiredAt.Equal(last) {
		t.Fatalf("last_fired_at not preserved: got %v want %v", reloaded[0].LastFiredAt, last)
	}

	// Schedule change resets last_fired_at.
	triggers[0].Schedule = "every 1h"
	if err := store.ReplaceTriggersForMate(m.ID, triggers); err != nil {
		t.Fatal(err)
	}
	reloaded, err = store.ListTriggers(m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reloaded[0].LastFiredAt.IsZero() {
		t.Fatalf("expected reset last_fired_at, got %v", reloaded[0].LastFiredAt)
	}

	now := time.Now()
	if err := store.UpdateTriggerLastFiredAt("tr-1", now); err != nil {
		t.Fatal(err)
	}
	reloaded, err = store.ListTriggers(m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded[0].LastFiredAt.Unix() != now.Unix() {
		t.Fatalf("update last_fired_at: got %v want %v", reloaded[0].LastFiredAt, now)
	}
}
