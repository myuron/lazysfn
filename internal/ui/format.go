// Package ui provides TUI formatting and rendering utilities.
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/myuron/lazysfn/internal/aws"
)

// ColumnWidths defines the width of each column in the execution history table.
type ColumnWidths struct {
	ID         int
	Status     int
	FailState  int
	StartTime  int
	StopTime   int
	Duration   int
	InputParam int
}

// StatusColor returns a color code string for the given execution status.
func StatusColor(status string) string {
	switch status {
	case "SUCCEEDED":
		return "green"
	case "FAILED":
		return "red"
	case "RUNNING":
		return "blue"
	case "TIMED_OUT":
		return "yellow"
	case "ABORTED":
		return "gray"
	default:
		return ""
	}
}

// TruncateWithEllipsis truncates a string to maxLen characters, appending "..." if truncated.
func TruncateWithEllipsis(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// FormatDuration formats a duration in the "1h23m45s" style.
func FormatDuration(d time.Duration) string {
	d = d.Truncate(time.Second)
	if d < 0 {
		return "-"
	}
	if d == 0 {
		return "0s"
	}

	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	result := ""
	if h > 0 {
		result += fmt.Sprintf("%dh", h)
	}
	if m > 0 {
		result += fmt.Sprintf("%dm", m)
	}
	if s > 0 {
		result += fmt.Sprintf("%ds", s)
	}
	return result
}

// FormatTime formats a time.Time in "2006-01-02 15:04:05" format.
// Returns an empty string for zero-value time.
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// FormatHeaderRow formats the column header row string with the given column widths.
// Columns are separated by "│" box-drawing characters.
func FormatHeaderRow(widths ColumnWidths) string {
	return fmt.Sprintf("%-*s│%-*s│%-*s│%-*s│%-*s│%-*s│%s",
		widths.ID, "ID",
		widths.Status, "STATUS",
		widths.FailState, "FAIL STATE",
		widths.StartTime, "START TIME",
		widths.StopTime, "STOP TIME",
		widths.Duration, "DURATION",
		"INPUT PARAM",
	)
}

// FormatSeparatorRow formats a horizontal separator row using "─" and "┼" characters.
func FormatSeparatorRow(widths ColumnWidths) string {
	return strings.Repeat("─", widths.ID) + "┼" +
		strings.Repeat("─", widths.Status) + "┼" +
		strings.Repeat("─", widths.FailState) + "┼" +
		strings.Repeat("─", widths.StartTime) + "┼" +
		strings.Repeat("─", widths.StopTime) + "┼" +
		strings.Repeat("─", widths.Duration) + "┼" +
		strings.Repeat("─", widths.InputParam)
}

// HighlightMatch wraps the first case-insensitive occurrence of query in name
// with ANSI yellow foreground escape codes (\033[33m ... \033[0m).
// Returns name unchanged if query is empty or not found.
func HighlightMatch(name, query string) string {
	if query == "" {
		return name
	}
	lowerName := strings.ToLower(name)
	lowerQuery := strings.ToLower(query)
	idx := strings.Index(lowerName, lowerQuery)
	if idx < 0 {
		return name
	}
	before := name[:idx]
	match := name[idx : idx+len(lowerQuery)]
	after := name[idx+len(lowerQuery):]
	return before + "\033[33m" + match + "\033[0m" + after
}

// FilterMachines returns the subset of machines whose Name contains query as a
// case-insensitive substring. An empty query returns machines unchanged.
// The returned slice is never nil; an unmatched query returns an empty slice.
func FilterMachines(machines []aws.StateMachine, query string) []aws.StateMachine {
	if query == "" {
		return machines
	}
	lower := strings.ToLower(query)
	result := []aws.StateMachine{}
	for _, m := range machines {
		if strings.Contains(strings.ToLower(m.Name), lower) {
			result = append(result, m)
		}
	}
	return result
}

// FormatExecutionRow formats an execution as a single row string with the given column widths.
func FormatExecutionRow(exec aws.Execution, widths ColumnWidths) string {
	id := TruncateWithEllipsis(exec.ID, widths.ID)
	status := TruncateWithEllipsis(exec.Status, widths.Status)

	failState := ""
	if exec.Status == "FAILED" || exec.Status == "TIMED_OUT" || exec.Status == "ABORTED" {
		failState = TruncateWithEllipsis(exec.FailedState, widths.FailState)
	}

	startTime := FormatTime(exec.StartTime)

	var stopTime string
	if exec.Status == "RUNNING" {
		stopTime = "-"
	} else {
		stopTime = FormatTime(exec.StopTime)
	}

	var duration string
	if exec.Status == "RUNNING" {
		duration = FormatDuration(time.Since(exec.StartTime))
	} else if !exec.StopTime.IsZero() {
		duration = FormatDuration(exec.StopTime.Sub(exec.StartTime))
	}

	inputParam := TruncateWithEllipsis(exec.InputParam, widths.InputParam)

	return fmt.Sprintf("%-*s│%-*s│%-*s│%-*s│%-*s│%-*s│%s",
		widths.ID, id,
		widths.Status, status,
		widths.FailState, failState,
		widths.StartTime, startTime,
		widths.StopTime, stopTime,
		widths.Duration, duration,
		inputParam,
	)
}
