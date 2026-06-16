package view

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
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
		OnlyOrigins: []storage.Origin{storage.OriginShort},
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

// Shorts handles GET /shorts — stream page.
func (vh *ViewHandler) Shorts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ?from=YYYY-MM jumps to a specific month's entries (loads from end of that month).
	before := time.Now().Add(time.Second)
	activeYM := time.Now().Format("2006-01")
	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.ParseInLocation("2006-01", from, time.Local); err == nil {
			before = t.AddDate(0, 1, 0) // first moment of next month = end of selected month
			activeYM = from
		}
	}

	groups, cursor, hasMore, err := vh.loadStreamGroups(before, 50)
	if err != nil {
		http.Error(w, "shorts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := shortsStreamPageData{
		Groups:  groups,
		Cursor:  cursor,
		HasMore: hasMore,
		Months:  vh.loadMonthsWithEntries(activeYM),
	}

	var buf bytes.Buffer
	if err := shortsPageTemplate.Execute(&buf, data); err != nil {
		http.Error(w, fmt.Sprintf("render: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

// ShortsStream handles GET /shorts/stream?before=RFC3339 — HTMX load-more fragment.
func (vh *ViewHandler) ShortsStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/shorts") {
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

// ─── legacy handlers (kept for backward compatibility) ────────────────────────

// shortsDatesForMonth returns all YYYY-MM-DD strings that have a short note
// file within the given calendar month.
func (vh *ViewHandler) shortsDatesForMonth(year, month int) ([]string, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	notes, err := vh.vault.ListAllNotes(storage.ListOptions{
		OnlyOrigins: []storage.Origin{storage.OriginShort},
		SortByTime: true,
		After:      start,
		Before:     end,
	})
	if err != nil {
		return nil, err
	}
	prefix := fmt.Sprintf("%04d-%02d-", year, month)
	var dates []string
	for _, n := range notes {
		date := strings.TrimSuffix(n.Name, ".md")
		if len(date) == 10 && strings.HasPrefix(date, prefix) {
			dates = append(dates, date)
		}
	}
	return dates, nil
}

func (vh *ViewHandler) loadDayEntries(date string) ([]shortRenderedEntry, error) {
	t, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return nil, err
	}

	p, err := storage.ShortDailyVaultPath("", t)
	if err != nil {
		return nil, err
	}
	rc, err := vh.vault.ReadNoteStream(p)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	raw, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	entries := storage.ParseShortEntries(raw, p.String())
	out := make([]shortRenderedEntry, 0, len(entries))
	for _, e := range entries {
		content := stripShortContent(e.Content)
		rendered, renderErr := util.MarkdownToHTMLFragment([]byte(content))
		if renderErr != nil {
			rendered = []byte(template.HTMLEscapeString(content))
		}
		out = append(out, shortRenderedEntry{
			Time: e.CreatedAt.Format("15:04"),
			HTML: template.HTML(rendered), //nolint:gosec // server-rendered markdown
		})
	}
	return out, nil
}

type shortsCalCell struct {
	Day        int
	Date       string
	HasEntries bool
	IsToday    bool
}

type shortsCalendarData struct {
	Today      string
	MonthLabel string
	PrevYear   int
	PrevMonth  int
	NextYear   int
	NextMonth  int
	Year       int
	Month      int
	Cells      []shortsCalCell
}

func (vh *ViewHandler) buildShortsCalendar(year, month int) (shortsCalendarData, error) {
	today := time.Now().Format("2006-01-02")
	dates, err := vh.shortsDatesForMonth(year, month)
	if err != nil {
		return shortsCalendarData{}, err
	}
	dateSet := make(map[string]bool, len(dates))
	for _, d := range dates {
		dateSet[d] = true
	}

	t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	prevT := t.AddDate(0, -1, 0)
	nextT := t.AddDate(0, 1, 0)

	offset := (int(t.Weekday()) + 6) % 7
	daysInMonth := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.Local).Day()

	cells := make([]shortsCalCell, 0, offset+daysInMonth)
	for i := 0; i < offset; i++ {
		cells = append(cells, shortsCalCell{})
	}
	for d := 1; d <= daysInMonth; d++ {
		date := fmt.Sprintf("%04d-%02d-%02d", year, month, d)
		cells = append(cells, shortsCalCell{
			Day:        d,
			Date:       date,
			HasEntries: dateSet[date],
			IsToday:    date == today,
		})
	}

	return shortsCalendarData{
		Today:      today,
		MonthLabel: strings.ToUpper(t.Format("Jan 2006")),
		PrevYear:   prevT.Year(),
		PrevMonth:  int(prevT.Month()),
		NextYear:   nextT.Year(),
		NextMonth:  int(nextT.Month()),
		Year:       year,
		Month:      month,
		Cells:      cells,
	}, nil
}

// ShortsDay handles GET /shorts/day?date=YYYY-MM-DD — legacy HTMX fragment.
func (vh *ViewHandler) ShortsDay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/shorts") {
		return
	}

	date := r.URL.Query().Get("date")
	if len(date) != 10 {
		http.Error(w, "invalid date", http.StatusBadRequest)
		return
	}

	entries, err := vh.loadDayEntries(date)
	if err != nil {
		http.Error(w, "shorts/day: "+err.Error(), http.StatusInternalServerError)
		return
	}

	today := time.Now().Format("2006-01-02")
	data := struct {
		DateLabel string
		Entries   []shortRenderedEntry
		IsToday   bool
	}{
		DateLabel: formatShortsDateLabel(date),
		Entries:   entries,
		IsToday:   date == today,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := shortsDayTemplate.Execute(w, data); err != nil {
		http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
	}
}

// ShortsCalendar handles GET /shorts/calendar?year=YYYY&month=M — legacy HTMX calendar fragment.
func (vh *ViewHandler) ShortsCalendar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !htmxOnly(w, r, "/shorts") {
		return
	}
	year, err1 := strconv.Atoi(r.URL.Query().Get("year"))
	month, err2 := strconv.Atoi(r.URL.Query().Get("month"))
	if err1 != nil || err2 != nil || month < 1 || month > 12 || year < 2000 || year > 2200 {
		http.Error(w, "invalid year/month", http.StatusBadRequest)
		return
	}
	data, err := vh.buildShortsCalendar(year, month)
	if err != nil {
		http.Error(w, "shorts/calendar: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := shortsCalendarTemplate.Execute(w, data); err != nil {
		http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
	}
}

// ─── templates ────────────────────────────────────────────────────────────────

var shortsDayTemplate = template.Must(template.New("shorts-day").Parse(shortsDayHTML))
var shortsCalendarTemplate = template.Must(template.New("shorts-calendar").Parse(shortsCalendarHTML))
var shortsStreamTemplate = template.Must(template.New("shorts-stream").Parse(shortsStreamHTML))

var shortsPageHTML = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
` + headHTML(headOpts{title: "Shorts — Vaultr", withFonts: true, withTW: true, withAlpine: true, withHTMX: true}) + `  <style>
    :root {` + appTokensDark + `}
    html[data-theme="light"] {` + appTokensLight + `}
` + infoDialogCSS + navCSS + pixelCSS + topbarCSS + shortsCSS + drawerCSS + noteSharedCSS + noteEditorCSS + searchOverlayStyles + shortDialogCSS + settingsModalCSS + `
  </style>
</head>
<body x-data="shortsCtrl()">
` + searchOnlyOverlayHTML + infoDialogHTML + shortDialogHTML + settingsModalHTML() + `

  <header class="lib-topbar">
    <div class="lib-topbar-left">
      <a href="/home" class="lib-back-btn" title="Back to Home">` + topbarIconBack + `<span class="lib-back-label">back</span></a>
    </div>
    <div class="lib-topbar-spacer"></div>
    <div class="lib-topbar-actions">
      ` + shortTriggerButton + `
      <button type="button" class="lib-action-btn" title="Refresh" @click="refresh()">` + topbarIconReload + `</button>
      <button type="button" class="lib-action-btn"
              :class="{ 'is-active': drawerOpen }"
              :title="/Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent) ? 'Reading drawer (⌘E)' : 'Reading drawer (Ctrl+E)'"
              @click="drawerOpen = !drawerOpen">` + topbarIconPanel + `</button>
      <button type="button" class="lib-action-btn"
              :title="/Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent) ? 'Search (⌘K)' : 'Search (Ctrl+K)'"
              @click="window.dispatchEvent(new CustomEvent('open-search'))">` + topbarIconSearch + `</button>
    </div>
  </header>

  <div class="shorts-body">
` + navHTML("shorts") + shortsMainHTML + `
  </div>

` + drawerHTML + `

  <script>
  document.addEventListener('alpine:init', () => {
` + themeStoreScript + `
  });

` + keysJS + drawerScript + searchOverlayScript + infoDialogJS + shortDialogJS + settingsCtrlJS + shortsJS + `
  </script>
` + noteSharedJS + `
</body>
</html>`

var shortsPageTemplate = template.Must(template.New("shorts").Parse(shortsPageHTML))
