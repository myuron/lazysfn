package ui

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

const (
	// helpModalName is the gocui view name for the keybinding help modal.
	helpModalName = "helpModal"
	// helpModalWidth is the fixed display width (in columns) of the help modal.
	helpModalWidth = 50
)

// ShowHelpModal displays the keybinding help in a centered modal.
// The modal closes on ?, Esc, or q, restoring focus to the previously active view.
func (a *App) ShowHelpModal(g *gocui.Gui, v *gocui.View) error {
	prevView := ""
	if cv := g.CurrentView(); cv != nil {
		prevView = cv.Name()
	}

	screenW, screenH := g.Size()
	content := helpContent()
	lines := strings.Split(content, "\n")
	modalH := len(lines) + 2 // +2 for borders
	x0, y0, x1, y1 := calcModalPosition(screenW, screenH, helpModalWidth, modalH)

	view, err := g.SetView(helpModalName, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("setting help modal view: %w", err)
	}
	view.Clear()
	view.Title = "Keybindings"
	if _, err := fmt.Fprint(view, content); err != nil {
		return fmt.Errorf("writing help content: %w", err)
	}

	closeModal := func(g *gocui.Gui, v *gocui.View) error {
		return a.closeHelpModal(g, prevView)
	}

	if err := g.SetKeybinding(helpModalName, '?', gocui.ModNone, closeModal); err != nil {
		return fmt.Errorf("binding ? on help modal: %w", err)
	}
	if err := g.SetKeybinding(helpModalName, gocui.KeyEsc, gocui.ModNone, closeModal); err != nil {
		return fmt.Errorf("binding Esc on help modal: %w", err)
	}
	if err := g.SetKeybinding(helpModalName, 'q', gocui.ModNone, closeModal); err != nil {
		return fmt.Errorf("binding q on help modal: %w", err)
	}

	if _, err := g.SetCurrentView(helpModalName); err != nil {
		return fmt.Errorf("setting help modal current view: %w", err)
	}
	return nil
}

// closeHelpModal removes the help modal view, cleans up its keybindings, and
// restores focus to prevView. It is called by the ?, Esc, and q keybindings
// registered in ShowHelpModal.
func (a *App) closeHelpModal(g *gocui.Gui, prevView string) error {
	g.DeleteKeybindings(helpModalName)
	if err := g.DeleteView(helpModalName); err != nil {
		return fmt.Errorf("deleting help modal: %w", err)
	}
	if prevView != "" {
		if _, err := g.SetCurrentView(prevView); err != nil {
			return fmt.Errorf("restoring focus: %w", err)
		}
	}
	return nil
}

// helpContent returns the keybinding reference text shown in the help modal.
func helpContent() string {
	return ` Global
   ?       Toggle help
   Esc     Close popup
   q       Quit / Close popup
   R       Refresh

 Main View
   j / k   Cursor down / up
   h / l   Focus left / right
   Tab     Switch panel
   /       Incremental search (left panel)

 Search Mode
   Esc     Cancel search (show all)
   Enter   Confirm search (keep filter)

 Profile Modal
   j / k   Cursor down / up
   Enter   Select profile
   q       Quit`
}
