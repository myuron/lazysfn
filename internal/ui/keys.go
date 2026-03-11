package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

// nextFocus returns the view that should receive focus after the current view,
// cycling through the provided views slice. If current is not found in views,
// the first element is returned. An empty views slice returns "".
func nextFocus(current string, views []string) string {
	if len(views) == 0 {
		return ""
	}
	for i, v := range views {
		if v == current {
			return views[(i+1)%len(views)]
		}
	}
	return views[0]
}

// nextSpinnerFrame returns the frame index that follows the given frame,
// wrapping around when the end of spinnerFrames is reached.
func nextSpinnerFrame(frame int) int {
	return (frame + 1) % len(spinnerFrames)
}

// focusView sets the current gocui view to viewName.
func (a *App) focusView(g *gocui.Gui, viewName string) error {
	if _, err := g.SetCurrentView(viewName); err != nil {
		return fmt.Errorf("focusing view %q: %w", viewName, err)
	}
	return nil
}

// tabFocus cycles focus to the next panel in the main view when Tab is pressed.
func (a *App) tabFocus(g *gocui.Gui, v *gocui.View) error {
	views := []string{leftViewName, rightViewName}
	current := ""
	if v != nil {
		current = v.Name()
	}
	next := nextFocus(current, views)
	return a.focusView(g, next)
}

// refresh re-renders both panels with the current application state.
func (a *App) refresh(g *gocui.Gui, v *gocui.View) error {
	if err := a.RenderLeftPanel(g); err != nil {
		return fmt.Errorf("refreshing left panel: %w", err)
	}
	if err := a.RenderRightPanel(g, a.executions); err != nil {
		return fmt.Errorf("refreshing right panel: %w", err)
	}
	return nil
}

// AdvanceSpinner increments the spinner frame and re-renders the right panel.
// It is intended to be called periodically from a ticker goroutine (via
// g.Update) while a.loading is true. The caller is responsible for starting
// and stopping the ticker around the loading period.
func (a *App) AdvanceSpinner(g *gocui.Gui) error {
	a.spinnerFrame = nextSpinnerFrame(a.spinnerFrame)
	if err := a.RenderRightPanel(g, a.executions); err != nil {
		return fmt.Errorf("advancing spinner: %w", err)
	}
	return nil
}
