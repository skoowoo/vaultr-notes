package view

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/hardhacker/vaultr/internal/storage"
	"github.com/hardhacker/vaultr/internal/util"
)

// ─── stream types ─────────────────────────────────────────────────────────────

type shortsStreamGroup struct {
	Date        string
	DateLabel   string
	CompactDate string // "Jun 16" or "Jun 16, 2025" — for per-entry timestamp display
	IsToday     bool
	Entries     []shortRenderedEntry
}

type shortsMonthItem struct {
	Abbr      string // "JUN"
	Year      string // "2026"
	YM        string // "2026-06"
	HasData   bool
	IsCurrent bool
}

type shortsStreamPageData struct {
	Groups  []shortsStreamGroup
	Cursor  string // RFC3339 of the oldest entry shown — load-more cursor
	HasMore bool
	Months  []shortsMonthItem
}

type shortRenderedEntry struct {
	Time string
	HTML template.HTML
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func stripShortContent(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "###### Short Note:") {
		if i := strings.Index(content, "\n"); i >= 0 {
			content = strings.TrimSpace(content[i+1:])
		} else {
			content = ""
		}
	}
	return content
}

func formatShortsDateLabel(date string) string {
	t, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return date
	}
	return t.Format("Monday, January 2")
}

// loadStreamGroups returns entries grouped by date, newest-first.
// It fetches limit+1 entries to determine whether more exist.
// Returns (groups, cursor, hasMore, err). cursor is the RFC3339 timestamp of
// the oldest entry returned, used as the next load-more `before` value.
func (vh *ViewHandler) loadStreamGroups(before time.Time, limit int) ([]shortsStreamGroup, string, bool, error) {
	entries, err := vh.vault.ListShortEntries(storage.ShortListOptions{
		Before: before,
		Limit:  limit + 1,
	})
	if err != nil {
		return nil, "", false, err
	}

	hasMore := len(entries) > limit
	if hasMore {
		entries = entries[:limit]
	}

	now := time.Now()
	today := now.Format("2006-01-02")
	currentYear := now.Year()
	var groups []shortsStreamGroup
	groupIdx := make(map[string]int)

	for _, e := range entries {
		content := stripShortContent(e.Content)
		rendered, renderErr := util.MarkdownToHTMLFragment([]byte(content))
		if renderErr != nil {
			rendered = []byte(template.HTMLEscapeString(content))
		}

		date := e.CreatedAt.Format("2006-01-02")
		idx, ok := groupIdx[date]
		if !ok {
			compactDate := e.CreatedAt.Format("Jan 2")
			if e.CreatedAt.Year() != currentYear {
				compactDate = e.CreatedAt.Format("Jan 2, 2006")
			}
			idx = len(groups)
			groupIdx[date] = idx
			groups = append(groups, shortsStreamGroup{
				Date:        date,
				DateLabel:   formatShortsDateLabel(date),
				CompactDate: compactDate,
				IsToday:     date == today,
			})
		}
		groups[idx].Entries = append(groups[idx].Entries, shortRenderedEntry{
			Time: e.CreatedAt.Format("15:04"),
			HTML: template.HTML(rendered), //nolint:gosec // server-rendered markdown
		})
	}

	var cursor string
	if len(entries) > 0 {
		cursor = entries[len(entries)-1].CreatedAt.Format(time.RFC3339)
	}

	return groups, cursor, hasMore, nil
}

// loadMonthsWithEntries returns the last 24 calendar months, marking which ones
// have at least one short note file. activeYM is the YYYY-MM that should be
// highlighted as selected (typically today's month or the ?from= param value).
func (vh *ViewHandler) loadMonthsWithEntries(activeYM string) []shortsMonthItem {
	notes, _ := vh.vault.ListAllNotes(storage.ListOptions{
		OnlyKinds: []storage.Kind{storage.KindShort},
	})

	monthSet := make(map[string]bool)
	for _, n := range notes {
		date := strings.TrimSuffix(n.Name, ".md")
		if len(date) == 10 {
			monthSet[date[:7]] = true
		}
	}

	base := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)

	items := make([]shortsMonthItem, 0, 24)
	for i := 0; i < 24; i++ {
		t := base.AddDate(0, -i, 0)
		ym := t.Format("2006-01")
		items = append(items, shortsMonthItem{
			Abbr:      strings.ToUpper(t.Format("Jan")),
			Year:      t.Format("2006"),
			YM:        ym,
			HasData:   monthSet[ym],
			IsCurrent: ym == activeYM,
		})
	}
	return items
}

// ─── handlers ─────────────────────────────────────────────────────────────────

// ShortsStream handles GET /shorts/stream?before=RFC3339 — HTMX load-more
// fragment, used by home's embedded Shorts section (see home.go).
func (vh *ViewHandler) ShortsStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/home") {
		return
	}

	beforeStr := r.URL.Query().Get("before")
	if beforeStr == "" {
		http.Error(w, "before required", http.StatusBadRequest)
		return
	}
	before, err := time.Parse(time.RFC3339, beforeStr)
	if err != nil {
		http.Error(w, "invalid before", http.StatusBadRequest)
		return
	}

	groups, cursor, hasMore, err := vh.loadStreamGroups(before, 30)
	if err != nil {
		http.Error(w, "shorts/stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Groups  []shortsStreamGroup
		Cursor  string
		HasMore bool
	}{groups, cursor, hasMore}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := shortsStreamTemplate.Execute(w, data); err != nil {
		http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
	}
}

// ─── templates ────────────────────────────────────────────────────────────────

var shortsStreamTemplate = template.Must(template.New("shorts-stream").Parse(shortsStreamHTML))
