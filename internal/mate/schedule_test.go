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
	if _, err := ParseSchedule("nonsense"); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseScheduleWeekly(t *testing.T) {
	s, err := ParseSchedule("weekly mon,wed,fri 09:30")
	if err != nil {
		t.Fatal(err)
	}
	if s.Kind != ScheduleWeekly || s.Hour != 9 || s.Minute != 30 {
		t.Fatalf("got %+v", s)
	}
	for _, wd := range []time.Weekday{time.Monday, time.Wednesday, time.Friday} {
		if !s.Weekdays[int(wd)] {
			t.Fatalf("expected %s to be set", wd)
		}
	}
	for _, wd := range []time.Weekday{time.Sunday, time.Tuesday, time.Thursday, time.Saturday} {
		if s.Weekdays[int(wd)] {
			t.Fatalf("expected %s to be unset", wd)
		}
	}
}

func TestParseScheduleWeeklyNoneSentinel(t *testing.T) {
	// "none" is the UI's placeholder for "weekly mode, zero days picked yet".
	if _, err := ParseSchedule("weekly none 09:00"); err == nil {
		t.Fatal("expected error for none sentinel")
	}
}

func TestParseScheduleWeeklyMissingTime(t *testing.T) {
	if _, err := ParseSchedule("weekly mon"); err == nil {
		t.Fatal("expected error for missing time")
	}
}

func TestParseScheduleWeeklyUnknownDay(t *testing.T) {
	if _, err := ParseSchedule("weekly funday 09:00"); err == nil {
		t.Fatal("expected error for unknown weekday")
	}
}

func TestScheduleWeeklyDue(t *testing.T) {
	loc := time.FixedZone("test", 0)
	var days [7]bool
	days[time.Monday] = true
	s := Schedule{Kind: ScheduleWeekly, Weekdays: days, Hour: 9, Minute: 0}

	// Monday before slot: not due.
	mondayBefore := time.Date(2026, 5, 25, 8, 30, 0, 0, loc) // 2026-05-25 is a Monday
	if s.Due(time.Time{}, mondayBefore) {
		t.Fatal("before slot should not be due")
	}

	// Monday after slot, first run: due.
	mondayAfter := time.Date(2026, 5, 25, 10, 0, 0, 0, loc)
	if !s.Due(time.Time{}, mondayAfter) {
		t.Fatal("first run after slot should be due")
	}

	// Already fired this Monday: not due again.
	lastFired := time.Date(2026, 5, 25, 9, 5, 0, 0, loc)
	if s.Due(lastFired, mondayAfter) {
		t.Fatal("already fired this monday")
	}

	// Missed last Monday: catches up today.
	lastMonday := time.Date(2026, 5, 18, 9, 5, 0, 0, loc)
	if !s.Due(lastMonday, mondayAfter) {
		t.Fatal("missed last monday should fire today")
	}

	// Non-matching weekday: never due regardless of time.
	tuesday := time.Date(2026, 5, 26, 10, 0, 0, 0, loc)
	if s.Due(time.Time{}, tuesday) {
		t.Fatal("non-matching weekday should not be due")
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
