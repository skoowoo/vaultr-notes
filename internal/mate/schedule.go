package mate

import (
	"fmt"
	"strings"
	"time"
)

// MinScheduleInterval is the shortest allowed "every" interval.
const MinScheduleInterval = 15 * time.Minute

// ScheduleKind distinguishes interval vs daily-at-time schedules.
type ScheduleKind int

const (
	ScheduleInterval ScheduleKind = iota
	ScheduleDaily
)

// Schedule is a parsed mate trigger schedule.
type Schedule struct {
	Kind     ScheduleKind
	Interval time.Duration
	Hour     int
	Minute   int
}

// ParseSchedule parses "every 1h" or "daily 09:00" (server local timezone for daily).
func ParseSchedule(raw string) (Schedule, error) {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return Schedule{}, fmt.Errorf("schedule: empty")
	}
	if strings.HasPrefix(s, "every ") {
		durStr := strings.TrimSpace(s[len("every "):])
		d, err := time.ParseDuration(durStr)
		if err != nil {
			return Schedule{}, fmt.Errorf("schedule: invalid interval %q: %w", durStr, err)
		}
		if d < MinScheduleInterval {
			return Schedule{}, fmt.Errorf("schedule: minimum interval is %s", MinScheduleInterval)
		}
		return Schedule{Kind: ScheduleInterval, Interval: d}, nil
	}
	if strings.HasPrefix(s, "daily ") {
		timeStr := strings.TrimSpace(s[len("daily "):])
		parsed, err := time.Parse("15:04", timeStr)
		if err != nil {
			return Schedule{}, fmt.Errorf("schedule: invalid daily time %q (use HH:MM): %w", timeStr, err)
		}
		return Schedule{Kind: ScheduleDaily, Hour: parsed.Hour(), Minute: parsed.Minute()}, nil
	}
	return Schedule{}, fmt.Errorf("schedule: unknown format %q (use \"every 1h\" or \"daily 09:00\")", raw)
}

// Due reports whether the schedule should fire given the last fire time and now.
// Daily schedules catch up once if the server was down past today's slot.
func (s Schedule) Due(lastFired, now time.Time) bool {
	if now.IsZero() {
		now = time.Now()
	}
	switch s.Kind {
	case ScheduleInterval:
		if lastFired.IsZero() {
			return true
		}
		return !now.Before(lastFired.Add(s.Interval))
	case ScheduleDaily:
		slot := dailySlot(now, s.Hour, s.Minute)
		if now.Before(slot) {
			return false
		}
		if lastFired.IsZero() {
			return true
		}
		return lastFired.Before(slot)
	default:
		return false
	}
}

func dailySlot(now time.Time, hour, minute int) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d, hour, minute, 0, 0, now.Location())
}
