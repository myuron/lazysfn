package ui

import (
	"fmt"
	"strings"

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
	colWidthID        = 30
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
	modalH := calcErrorModalHeight(msg)
	x0, y0, x1, y1 := calcModalPosition(screenW, screenH, errorModalWidth, modalH)

	v, err := g.SetView(errorModalName, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("setting error modal view: %w", err)
	}

	v.Clear()
	v.Title = "Error"
	if _, err := fmt.Fprint(v, msg); err != nil {
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
	a.executions = executions

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

	panelW, _ := v.Size()
	widths := defaultColumnWidths(panelW)

	for _, exec := range a.executions {
		row := FormatExecutionRow(exec, widths)
		if _, err := fmt.Fprintln(v, row); err != nil {
			return fmt.Errorf("writing execution row: %w", err)
		}
	}

	return nil
}

// defaultColumnWidths returns ColumnWidths for the right panel given its width.
// Fixed columns (ID=30, Status=10, FailState=20, StartTime=19, StopTime=19,
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
