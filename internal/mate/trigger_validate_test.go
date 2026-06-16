package mate

import (
	"testing"
	"time"
)

func TestValidateTriggerScheduled(t *testing.T) {
	if err := ValidateTrigger(MateTrigger{
		EventTypes: []string{string(MateEventScheduled)},
		Schedule:   "daily 09:00",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateTriggerScheduledMissingSchedule(t *testing.T) {
	if err := ValidateTrigger(MateTrigger{
		EventTypes: []string{string(MateEventScheduled)},
	}); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateTriggerMixedEvents(t *testing.T) {
	if err := ValidateTrigger(MateTrigger{
		EventTypes: []string{string(MateEventScheduled), string(MateEventNoteCreated)},
		Schedule:   "every 1h",
	}); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateTriggerScheduleWithoutEvent(t *testing.T) {
	if err := ValidateTrigger(MateTrigger{
		EventTypes: []string{string(MateEventNoteCreated)},
		Schedule:   "every 1h",
	}); err == nil {
		t.Fatal("expected error")
	}
}

func TestRenderPromptScheduled(t *testing.T) {
	fired := time.Date(2026, 5, 23, 9, 15, 0, 0, time.UTC)
	got := renderPrompt("Run digest at {{.Now}} on {{.Date}} {{.Time}}", MateEvent{
		Type:    MateEventScheduled,
		FiredAt: fired,
	})
	want := "Run digest at 2026-05-23T09:15:00Z on 2026-05-23 09:15"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
