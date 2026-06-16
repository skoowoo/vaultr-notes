package util

import (
	"fmt"
	"time"
)

const timeLayout = "2006-01-02 15:04:05"

// FormatTime formats t in local time using a human-readable layout.
// Returns an empty string for zero values.
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format(timeLayout)
}

// FormatSize formats a byte count as a human-readable string (e.g. "1.2 KB").
func FormatSize(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
