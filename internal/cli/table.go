package cli

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
)

// Column defines a table column header and an optional display-width cap (0 = no cap).
type Column struct {
	Header   string
	MaxWidth int
}

// PrintTable renders a plain-text table that handles multi-byte (e.g. CJK) characters
// correctly by using rune widths instead of byte lengths.
//
// If a cell value exceeds the column's MaxWidth it is truncated and a "…" suffix is added.
// Column widths are computed from the widest visible cell (or MaxWidth when it is set).
func PrintTable(cols []Column, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = runewidth.StringWidth(c.Header)
		if c.MaxWidth > 0 && widths[i] > c.MaxWidth {
			widths[i] = c.MaxWidth
		}
	}

	for _, row := range rows {
		for i := range cols {
			if i >= len(row) {
				break
			}
			w := runewidth.StringWidth(row[i])
			if cols[i].MaxWidth > 0 && w > cols[i].MaxWidth {
				w = cols[i].MaxWidth
			}
			if w > widths[i] {
				widths[i] = w
			}
		}
	}

	gap := "  "

	// Header
	parts := make([]string, len(cols))
	for i, c := range cols {
		parts[i] = runewidth.FillRight(truncate(c.Header, widths[i]), widths[i])
	}
	fmt.Println(strings.Join(parts, gap))

	// Separator
	for i, w := range widths {
		parts[i] = strings.Repeat("-", w)
	}
	fmt.Println(strings.Join(parts, gap))

	// Rows
	for _, row := range rows {
		for i := range cols {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			parts[i] = runewidth.FillRight(truncate(cell, widths[i]), widths[i])
		}
		fmt.Println(strings.Join(parts, gap))
	}
}

// truncate shortens s so its display width is at most maxWidth, appending "…" if cut.
// maxWidth == 0 means no limit.
func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return s
	}
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	// Reserve 1 cell for the ellipsis rune.
	budget := maxWidth - 1
	out := []rune{}
	used := 0
	for _, r := range s {
		w := runewidth.RuneWidth(r)
		if used+w > budget {
			break
		}
		out = append(out, r)
		used += w
	}
	return string(out) + "…"
}
