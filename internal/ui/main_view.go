package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/myuron/lazysfn/internal/aws"
)

const (
	leftViewName  = "left"
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
	_, panelH := v.Size()
	panelW, _ := v.Size()

	if len(a.machines) == 0 {
		if _, err := fmt.Fprintln(v, "(0 state machines)"); err != nil {
			return fmt.Errorf("writing empty message: %w", err)
		}
		return nil
	}

	for i, m := range a.machines {
		// Scroll: only render rows visible in the panel.
		if i >= panelH {
			break
		}

		line := formatSMLine(m.Name, m.LatestStatus, panelW)

		prefix := "  "
		if i == a.smCursor {
			prefix = "> "
		}
		if _, err := fmt.Fprintln(v, prefix+line); err != nil {
			return fmt.Errorf("writing state machine row: %w", err)
		}
	}

	return nil
}

// setMainViewKeybindings registers j/k navigation for the left panel.
func (a *App) setMainViewKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding(leftViewName, 'j', gocui.ModNone, a.smCursorDown); err != nil {
		return fmt.Errorf("binding j in left panel: %w", err)
	}
	if err := g.SetKeybinding(leftViewName, 'k', gocui.ModNone, a.smCursorUp); err != nil {
		return fmt.Errorf("binding k in left panel: %w", err)
	}
	if err := g.SetKeybinding(leftViewName, 'q', gocui.ModNone, quit); err != nil {
		return fmt.Errorf("binding q in left panel: %w", err)
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
