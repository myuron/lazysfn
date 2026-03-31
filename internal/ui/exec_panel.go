package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jroimartin/gocui"
	"github.com/myuron/lazysfn/internal/aws"
)

const (
	// errorModalName is the gocui view name for the error modal.
	errorModalName = "errorModal"
	// errorModalWidth is the fixed display width (in columns) of the error modal.
	errorModalWidth = 60
	// minErrorModalHeight is the minimum height of the error modal (1 content row + 2 border rows).
	minErrorModalHeight = 3

	// spinnerFrames are the characters cycled through for the loading spinner.
	spinnerFrames = `|/-\`

	// Fixed column widths for the execution history table, matching SPEC.md.
	colWidthID        = 12
	colWidthStatus    = 10
	colWidthFailState = 20
	colWidthStartTime = 19
	colWidthStopTime  = 19
	colWidthDuration  = 10
)

// ShowErrorModal displays an error message in a centered modal.
// The modal shows the error text and closes on Enter or q, returning to profile selection.
func (a *App) ShowErrorModal(g *gocui.Gui, msg string) error {
	screenW, screenH := g.Size()
	innerW := errorModalWidth - 2 // subtract border columns
	wrapped := wrapText(msg, innerW)
	modalH := calcErrorModalHeight(wrapped)
	x0, y0, x1, y1 := calcModalPosition(screenW, screenH, errorModalWidth, modalH)

	v, err := g.SetView(errorModalName, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("setting error modal view: %w", err)
	}

	v.Clear()
	v.Title = "Error"
	if _, err := fmt.Fprint(v, wrapped); err != nil {
		return fmt.Errorf("writing error message: %w", err)
	}

	// closeModal dismisses the error modal and returns focus to the appropriate
	// view. In main view mode, focus returns to the left panel; otherwise it
	// returns to the profile selection modal.
	closeModal := func(g *gocui.Gui, v *gocui.View) error {
		// Delete keybindings first; gocui does not remove them automatically
		// when a view is deleted, which would cause a leak on repeated calls.
		g.DeleteKeybindings(errorModalName)
		if err := g.DeleteView(errorModalName); err != nil {
			return fmt.Errorf("deleting error modal: %w", err)
		}
		targetView := modalName
		if a.inMainView {
			targetView = leftViewName
		}
		if _, err := g.SetCurrentView(targetView); err != nil {
			return fmt.Errorf("setting current view after error: %w", err)
		}
		return nil
	}

	if err := g.SetKeybinding(errorModalName, gocui.KeyEnter, gocui.ModNone, closeModal); err != nil {
		return fmt.Errorf("binding Enter on error modal: %w", err)
	}
	if err := g.SetKeybinding(errorModalName, 'q', gocui.ModNone, closeModal); err != nil {
		return fmt.Errorf("binding q on error modal: %w", err)
	}

	if _, err := g.SetCurrentView(errorModalName); err != nil {
		return fmt.Errorf("setting current view to error modal: %w", err)
	}

	return nil
}

// wrapText wraps text so that each line is at most width runes wide,
// breaking at word boundaries. Existing newlines are preserved.
// Long words that exceed width are placed on their own line without splitting.
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	var result strings.Builder
	for i, line := range strings.Split(text, "\n") {
		if i > 0 {
			result.WriteByte('\n')
		}
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}
		lineLen := 0
		for j, word := range words {
			wordLen := utf8.RuneCountInString(word)
			if j == 0 {
				result.WriteString(word)
				lineLen = wordLen
			} else if lineLen+1+wordLen <= width {
				result.WriteByte(' ')
				result.WriteString(word)
				lineLen += 1 + wordLen
			} else {
				result.WriteByte('\n')
				result.WriteString(word)
				lineLen = wordLen
			}
		}
	}
	return result.String()
}

// calcErrorModalHeight returns the modal height for the given message.
// Height = number of lines in msg + 2 (borders).
// The result is at least minErrorModalHeight to keep the modal usable.
func calcErrorModalHeight(msg string) int {
	lines := strings.Count(msg, "\n") + 1
	if msg == "" {
		lines = 0
	}
	h := lines + 2
	if h < minErrorModalHeight {
		h = minErrorModalHeight
	}
	return h
}

// RenderRightPanel renders the execution history in the right panel.
// Shows a spinner while loading, or the execution list when loaded.
// The executions slice is stored in the App for use by the spinner goroutine.
func (a *App) RenderRightPanel(g *gocui.Gui, executions []aws.Execution) error {
	// Reset cursor when a new execution list is loaded.
	if len(executions) != len(a.executions) || (len(executions) > 0 && len(a.executions) > 0 && executions[0].ID != a.executions[0].ID) {
		a.execCursor = 0
	}
	a.executions = executions
	if a.execCursor >= len(a.executions) && len(a.executions) > 0 {
		a.execCursor = len(a.executions) - 1
	}

	v, err := g.View(rightViewName)
	if err != nil {
		return fmt.Errorf("getting right panel view: %w", err)
	}

	v.Clear()

	if a.loading.Load() {
		frame := string(spinnerFrames[a.spinnerFrame%len(spinnerFrames)])
		if _, err := fmt.Fprintf(v, "Loading %s", frame); err != nil {
			return fmt.Errorf("writing spinner: %w", err)
		}
		return nil
	}

	panelW, panelH := v.Size()
	widths := defaultColumnWidths(panelW)

	// Always render fixed header (2 lines: header row + separator).
	if _, err := fmt.Fprintln(v, FormatHeaderRow(widths)); err != nil {
		return fmt.Errorf("writing header row: %w", err)
	}

	separator := FormatSeparatorRow(widths)
	if _, err := fmt.Fprintln(v, separator); err != nil {
		return fmt.Errorf("writing separator row: %w", err)
	}

	// Snapshot shared state to avoid races with concurrent refreshes.
	execs := a.executions
	cursor := a.execCursor

	// Each execution takes 2 lines (data + separator), except the last (no trailing separator).
	// Available lines for execution rows = panelH - header - optional footer.
	headerLines := 2
	footerLines := 0
	if a.loadingMore.Load() {
		footerLines = 2 // spacer + indicator
	}
	availLines := panelH - headerLines - footerLines
	if availLines < 0 {
		availLines = 0
	}
	// Number of executions that fit in the viewport.
	// Each row uses 2 lines (data + separator); the last visible row needs only 1.
	visibleCount := (availLines + 1) / 2
	if visibleCount > len(execs) {
		visibleCount = len(execs)
	}

	// Compute start index so cursor is always within the visible window.
	start := cursor - (visibleCount - 1)
	if start < 0 {
		start = 0
	}
	end := start + visibleCount
	if end > len(execs) {
		end = len(execs)
	}

	for i := start; i < end; i++ {
		row := FormatExecutionRow(execs[i], widths)
		if i == cursor {
			if cv := g.CurrentView(); cv != nil && cv.Name() == rightViewName {
				// Apply bold cyan to the row, preserving status column colors.
				// After each \033[0m reset, re-apply bold cyan so non-status parts stay cyan.
				row = "\033[1;36m" + strings.ReplaceAll(row, "\033[0m", "\033[0m\033[1;36m") + "\033[0m"
			}
		}
		if _, err := fmt.Fprintln(v, row); err != nil {
			return fmt.Errorf("writing execution row: %w", err)
		}
		if i < end-1 {
			if _, err := fmt.Fprintln(v, separator); err != nil {
				return fmt.Errorf("writing separator row: %w", err)
			}
		}
	}

	if a.loadingMore.Load() {
		if _, err := fmt.Fprintln(v, ""); err != nil {
			return fmt.Errorf("writing loading more spacer: %w", err)
		}
		if _, err := fmt.Fprintln(v, "Loading more..."); err != nil {
			return fmt.Errorf("writing loading more indicator: %w", err)
		}
	}

	return nil
}

// AppendExecutions appends additional executions to the existing list and updates
// the pagination token. Used for incremental loading when scrolling past the last item.
// forARN is the ARN that was used to fetch these executions; if the current SM has
// changed since the fetch started, the stale results are silently discarded.
func (a *App) AppendExecutions(g *gocui.Gui, executions []aws.Execution, nextToken *string, forARN string) error {
	if forARN != a.GetCurrentSMARN() {
		return nil
	}
	a.executions = append(a.executions, executions...)
	a.SetExecNextToken(nextToken)
	return a.RenderRightPanel(g, a.executions)
}

// defaultColumnWidths returns ColumnWidths for the right panel given its width.
// Fixed columns (ID=12, Status=10, FailState=20, StartTime=19, StopTime=19,
// Duration=10) plus 6 separator spaces are subtracted from panelWidth to
// determine InputParam width.
func defaultColumnWidths(panelWidth int) ColumnWidths {
	// 6 fixed columns (ID, Status, FailState, StartTime, StopTime, Duration) plus
	// InputParam = 7 columns total, with 6 single-space separators between them.
	separators := 6
	fixedTotal := colWidthID + colWidthStatus + colWidthFailState +
		colWidthStartTime + colWidthStopTime + colWidthDuration + separators

	inputParam := panelWidth - fixedTotal
	if inputParam < 0 {
		inputParam = 0
	}

	return ColumnWidths{
		ID:         colWidthID,
		Status:     colWidthStatus,
		FailState:  colWidthFailState,
		StartTime:  colWidthStartTime,
		StopTime:   colWidthStopTime,
		Duration:   colWidthDuration,
		InputParam: inputParam,
	}
}
