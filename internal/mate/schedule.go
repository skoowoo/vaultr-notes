package mate

import (
	"fmt"
	"strings"
	"time"
)

// MinScheduleInterval is the shortest allowed "every" interval.
const MinScheduleInterval = 15 * time.Minute

// ScheduleKind distinguishes interval, daily-at-time, and weekly-at-time schedules.
type ScheduleKind int

const (
	ScheduleInterval ScheduleKind = iota
	ScheduleDaily
	ScheduleWeekly
)

// Schedule is a parsed mate trigger schedule.
type Schedule struct {
	Kind     ScheduleKind
	Interval time.Duration
	Hour     int
	Minute   int
	Weekdays [7]bool // indexed by time.Weekday (Sunday=0); used when Kind == ScheduleWeekly
}

var weekdayAbbrs = map[string]time.Weekday{
	"sun": time.Sunday, "mon": time.Monday, "tue": time.Tuesday, "wed": time.Wednesday,
	"thu": time.Thursday, "fri": time.Friday, "sat": time.Saturday,
}

// ParseSchedule parses "every 1h", "daily 09:00", or "weekly mon,wed 09:00"
// (server local timezone for daily/weekly).
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
	if strings.HasPrefix(s, "weekly ") {
		rest := strings.Fields(strings.TrimSpace(s[len("weekly "):]))
		if len(rest) != 2 {
			return Schedule{}, fmt.Errorf("schedule: weekly format is \"weekly <days> HH:MM\", e.g. \"weekly mon,wed 09:00\"")
		}
		if rest[0] == "none" {
			return Schedule{}, fmt.Errorf("schedule: weekly requires at least one weekday")
		}
		var days [7]bool
		any := false
		for _, d := range strings.Split(rest[0], ",") {
			d = strings.TrimSpace(d)
			wd, ok := weekdayAbbrs[d]
			if !ok {
				return Schedule{}, fmt.Errorf("schedule: unknown weekday %q (use mon,tue,wed,thu,fri,sat,sun)", d)
			}
			days[wd] = true
			any = true
		}
		if !any {
			return Schedule{}, fmt.Errorf("schedule: weekly requires at least one weekday")
		}
		parsed, err := time.Parse("15:04", rest[1])
		if err != nil {
			return Schedule{}, fmt.Errorf("schedule: invalid weekly time %q (use HH:MM): %w", rest[1], err)
		}
		return Schedule{Kind: ScheduleWeekly, Weekdays: days, Hour: parsed.Hour(), Minute: parsed.Minute()}, nil
	}
	return Schedule{}, fmt.Errorf("schedule: unknown format %q (use \"every 1h\", \"daily 09:00\", or \"weekly mon,wed 09:00\")", raw)
}

// Due reports whether the schedule should fire given the last fire time and now.
// Daily and weekly schedules catch up once if the server was down past today's slot.
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
	case ScheduleWeekly:
		if !s.Weekdays[int(now.Weekday())] {
			return false
		}
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
