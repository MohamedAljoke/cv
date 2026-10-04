package resume

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var months = map[string][12]string{
	"en": {"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
	"pt": {"jan", "fev", "mar", "abr", "mai", "jun", "jul", "ago", "set", "out", "nov", "dez"},
}

// parseMonth turns "2025-11" into (2025, 11).
func parseMonth(s string) (year, month int) {
	y, _ := strconv.Atoi(s[:4])
	m, _ := strconv.Atoi(s[5:7])
	return y, m
}

// Month formats "2025-11" as "Nov 2025" (en) or "nov 2025" (pt).
func Month(lang, ym string) string {
	if ym == "" {
		return ""
	}
	y, m := parseMonth(ym)
	return fmt.Sprintf("%s %d", months[lang][m-1], y)
}

// Range formats a job's dates: "Nov 2025 – Present".
func Range(lang, start string, end *string, present string) string {
	to := present
	if end != nil && *end != "" {
		to = Month(lang, *end)
	}
	return Month(lang, start) + " – " + to
}

// Duration is the LinkedIn-style length of a job, counting both the first
// and last month: "1 yr 2 mos". asOf ("YYYY-MM") stands in for an open end,
// so the output only changes when the data changes.
func Duration(start string, end *string, asOf string, units [4]string) string {
	to := asOf
	if end != nil && *end != "" {
		to = *end
	}
	y1, m1 := parseMonth(start)
	y2, m2 := parseMonth(to)
	total := (y2*12 + m2) - (y1*12 + m1) + 1
	if total < 1 {
		total = 1
	}
	yrs, mos := total/12, total%12
	var parts []string
	if yrs > 0 {
		parts = append(parts, plural(yrs, units[0], units[1]))
	}
	if mos > 0 {
		parts = append(parts, plural(mos, units[2], units[3]))
	}
	return strings.Join(parts, " ")
}

func plural(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}

// MonthOf returns "YYYY-MM" for t.
func MonthOf(t time.Time) string { return t.Format("2006-01") }
