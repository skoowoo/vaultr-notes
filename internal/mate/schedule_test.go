package mate

import (
	"testing"
	"time"
)

func TestParseScheduleEvery(t *testing.T) {
	s, err := ParseSchedule("every 1h")
	if err != nil {
		t.Fatal(err)
	}
	if s.Kind != ScheduleInterval || s.Interval != time.Hour {
		t.Fatalf("got %+v", s)
	}
}

func TestParseScheduleEveryTooShort(t *testing.T) {
	_, err := ParseSchedule("every 5m")
	if err == nil {
		t.Fatal("expected error for short interval")
	}
}

func TestParseScheduleDaily(t *testing.T) {
	s, err := ParseSchedule("daily 09:30")
	if err != nil {
		t.Fatal(err)
	}
	if s.Kind != ScheduleDaily || s.Hour != 9 || s.Minute != 30 {
		t.Fatalf("got %+v", s)
	}
}

func TestParseScheduleInvalid(t *testing.T) {
	if _, err := ParseSchedule("weekly mon"); err == nil {
		t.Fatal("expected error")
	}
}

func TestScheduleIntervalDue(t *testing.T) {
	s := Schedule{Kind: ScheduleInterval, Interval: time.Hour}
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)
	if !s.Due(time.Time{}, now) {
		t.Fatal("first run should be due")
	}
	last := now.Add(-30 * time.Minute)
	if s.Due(last, now) {
		t.Fatal("should not be due yet")
	}
	last = now.Add(-time.Hour)
	if !s.Due(last, now) {
		t.Fatal("should be due after interval")
	}
}

func TestScheduleDailyDue(t *testing.T) {
	loc := time.FixedZone("test", 0)
	s := Schedule{Kind: ScheduleDaily, Hour: 9, Minute: 0}

	before := time.Date(2026, 5, 23, 8, 30, 0, 0, loc)
	if s.Due(time.Time{}, before) {
		t.Fatal("before slot should not be due")
	}

	after := time.Date(2026, 5, 23, 10, 0, 0, 0, loc)
	if !s.Due(time.Time{}, after) {
		t.Fatal("first run after slot should be due")
	}

	lastFired := time.Date(2026, 5, 23, 9, 5, 0, 0, loc)
	if s.Due(lastFired, after) {
		t.Fatal("already fired today")
	}

	yesterday := time.Date(2026, 5, 22, 9, 5, 0, 0, loc)
	if !s.Due(yesterday, after) {
		t.Fatal("missed yesterday should fire today")
	}
}
