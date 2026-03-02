package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/myuron/lazysfn/internal/aws"
)

const (
	// leftViewName is the gocui view name for the left (state machine list) panel.
	leftViewName = "left"
	// rightViewName is the gocui view name for the right (detail) panel.
	rightViewName = "right"
)

// SetupMainView initializes the main screen layout (left + right panels) in the gocui GUI.
// Called after profile selection is complete. It stores the state machine list,
// creates both panels, and registers the main-view keybindings.
func (a *App) SetupMainView(g *gocui.Gui, machines []aws.StateMachine) error {
	a.machines = machines
	a.smCursor = 0

	screenW, screenH := g.Size()
	leftW, _ := calcPanelWidths(screenW)

	// Left panel: x0=0, y0=0, x1=leftW-1, y1=screenH-1
	if _, err := g.SetView(leftViewName, 0, 0, leftW-1, screenH-1); err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("creating left panel: %w", err)
	}

	// Right panel: x0=leftW, y0=0, x1=screenW-1, y1=screenH-1
	if _, err := g.SetView(rightViewName, leftW, 0, screenW-1, screenH-1); err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("creating right panel: %w", err)
	}

	if err := a.RenderLeftPanel(g); err != nil {
		return fmt.Errorf("rendering left panel: %w", err)
	}

	if err := a.setMainViewKeybindings(g); err != nil {
		return fmt.Errorf("setting main view keybindings: %w", err)
	}

	if _, err := g.SetCurrentView(leftViewName); err != nil {
		return fmt.Errorf("setting current view to left panel: %w", err)
	}

	return nil
}

// RenderLeftPanel re-renders the left panel with the current state machine list.
// Each row shows the machine name with a status bullet on the right.
// If the list is empty, displays "(0 state machines)".
// The cursor row is prefixed with "> ".
func (a *App) RenderLeftPanel(g *gocui.Gui) error {
	v, err := g.View(leftViewName)
	if err != nil {
		return fmt.Errorf("getting left panel view: %w", err)
	}

	v.Clear()
	panelW, panelH := v.Size()

	if len(a.machines) == 0 {
		if _, err := fmt.Fprintln(v, "(0 state machines)"); err != nil {
			return fmt.Errorf("writing empty message: %w", err)
		}
		return nil
	}

	// availableWidth is the usable width after the 2-character prefix ("> " or "  ").
	availableWidth := panelW - 2

	// Scroll: compute the first visible index so the cursor is always in view.
	start := a.smCursor - (panelH - 1)
	if start < 0 {
		start = 0
	}
	end := start + panelH
	if end > len(a.machines) {
		end = len(a.machines)
	}

	for localIdx, m := range a.machines[start:end] {
		absIdx := start + localIdx

		line := formatSMLine(m.Name, m.LatestStatus, availableWidth)

		prefix := "  "
		if absIdx == a.smCursor {
			prefix = "> "
		}
		if _, err := fmt.Fprintln(v, prefix+line); err != nil {
			return fmt.Errorf("writing state machine row: %w", err)
		}
	}

	return nil
}

// setMainViewKeybindings registers keybindings for both the left and right panels.
// Left panel: j/k (navigation), q (quit), Tab/h/l/R (focus/refresh).
// Right panel: j/k (scroll), q (quit), Tab/h/l/R (focus/refresh).
func (a *App) setMainViewKeybindings(g *gocui.Gui) error {
	// --- left panel ---
	if err := g.SetKeybinding(leftViewName, 'j', gocui.ModNone, a.smCursorDown); err != nil {
		return fmt.Errorf("binding j in left panel: %w", err)
	}
	if err := g.SetKeybinding(leftViewName, 'k', gocui.ModNone, a.smCursorUp); err != nil {
		return fmt.Errorf("binding k in left panel: %w", err)
	}
	if err := g.SetKeybinding(leftViewName, 'q', gocui.ModNone, quit); err != nil {
		return fmt.Errorf("binding q in left panel: %w", err)
	}
	if err := g.SetKeybinding(leftViewName, gocui.KeyTab, gocui.ModNone, a.tabFocus); err != nil {
		return fmt.Errorf("binding Tab in left panel: %w", err)
	}
	if err := g.SetKeybinding(leftViewName, 'h', gocui.ModNone, a.focusLeft); err != nil {
		return fmt.Errorf("binding h in left panel: %w", err)
	}
	if err := g.SetKeybinding(leftViewName, 'l', gocui.ModNone, a.focusRight); err != nil {
		return fmt.Errorf("binding l in left panel: %w", err)
	}
	if err := g.SetKeybinding(leftViewName, 'R', gocui.ModNone, a.refresh); err != nil {
		return fmt.Errorf("binding R in left panel: %w", err)
	}

	// --- right panel ---
	if err := g.SetKeybinding(rightViewName, 'j', gocui.ModNone, a.smCursorDown); err != nil {
		return fmt.Errorf("binding j in right panel: %w", err)
	}
	if err := g.SetKeybinding(rightViewName, 'k', gocui.ModNone, a.smCursorUp); err != nil {
		return fmt.Errorf("binding k in right panel: %w", err)
	}
	if err := g.SetKeybinding(rightViewName, 'q', gocui.ModNone, quit); err != nil {
		return fmt.Errorf("binding q in right panel: %w", err)
	}
	if err := g.SetKeybinding(rightViewName, gocui.KeyTab, gocui.ModNone, a.tabFocus); err != nil {
		return fmt.Errorf("binding Tab in right panel: %w", err)
	}
	if err := g.SetKeybinding(rightViewName, 'h', gocui.ModNone, a.focusLeft); err != nil {
		return fmt.Errorf("binding h in right panel: %w", err)
	}
	if err := g.SetKeybinding(rightViewName, 'l', gocui.ModNone, a.focusRight); err != nil {
		return fmt.Errorf("binding l in right panel: %w", err)
	}
	if err := g.SetKeybinding(rightViewName, 'R', gocui.ModNone, a.refresh); err != nil {
		return fmt.Errorf("binding R in right panel: %w", err)
	}

	return nil
}

// smCursorDown moves the state machine cursor down one row and re-renders the left panel.
func (a *App) smCursorDown(g *gocui.Gui, v *gocui.View) error {
	if a.smCursor < len(a.machines)-1 {
		a.smCursor++
	}
	return a.RenderLeftPanel(g)
}

// smCursorUp moves the state machine cursor up one row and re-renders the left panel.
func (a *App) smCursorUp(g *gocui.Gui, v *gocui.View) error {
	if a.smCursor > 0 {
		a.smCursor--
	}
	return a.RenderLeftPanel(g)
}
