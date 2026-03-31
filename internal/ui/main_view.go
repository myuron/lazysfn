package ui

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/myuron/lazysfn/internal/aws"
)

const (
	// leftViewName is the gocui view name for the left (state machine list) panel.
	leftViewName = "left"
	// rightViewName is the gocui view name for the right (detail) panel.
	rightViewName = "right"
	// searchViewName is the gocui view name for the incremental search bar.
	searchViewName = "search"
	// searchBarHeight is the number of rows occupied by the search bar view (border + content).
	searchBarHeight = 3
)

// SetupMainView initializes the main screen layout (left + right panels) in the gocui GUI.
// Called after profile selection is complete. It stores the state machine list,
// creates both panels, and registers the main-view keybindings.
func (a *App) SetupMainView(g *gocui.Gui, machines []aws.StateMachine) error {
	a.machines = machines
	a.smCursor = 0

	// Switch to main view manager first. SetManagerFunc internally calls
	// SetManager which clears all views, keybindings, and currentView.
	// Everything must be re-created after this call.
	g.SetManagerFunc(a.mainViewManager)

	screenW, screenH := g.Size()
	leftW, _ := calcPanelWidths(screenW)

	// Left panel: x0=0, y0=0, x1=leftW-1, y1=screenH-1
	lv, err := g.SetView(leftViewName, 0, 0, leftW-1, screenH-1)
	if err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("creating left panel: %w", err)
	}
	lv.Title = "State Machine"

	// Right panel: x0=leftW, y0=0, x1=screenW-1, y1=screenH-1
	rv, err := g.SetView(rightViewName, leftW, 0, screenW-1, screenH-1)
	if err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("creating right panel: %w", err)
	}
	rv.Title = "history"

	if err := a.RenderLeftPanel(g); err != nil {
		return fmt.Errorf("rendering left panel: %w", err)
	}

	if err := a.setMainViewKeybindings(g); err != nil {
		return fmt.Errorf("setting main view keybindings: %w", err)
	}

	// Global quit keybindings (re-register after SetManagerFunc cleared them).
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return fmt.Errorf("binding Ctrl+C: %w", err)
	}

	if _, err := g.SetCurrentView(leftViewName); err != nil {
		return fmt.Errorf("setting current view to left panel: %w", err)
	}

	// Mark as main view mode.
	a.inMainView = true

	return nil
}

// mainViewManager is the gocui manager function used in the main view.
// It handles terminal resize by repositioning and re-rendering the left and right panels.
// In search mode the left panel is shortened and a search bar view is created below it.
func (a *App) mainViewManager(g *gocui.Gui) error {
	screenW, screenH := g.Size()
	leftW, _ := calcPanelWidths(screenW)

	// Determine left panel bottom edge based on search mode.
	leftBottom := screenH - 1
	if a.searchMode {
		leftBottom = screenH - searchBarHeight - 1
	}

	lv, err := g.SetView(leftViewName, 0, 0, leftW-1, leftBottom)
	if err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("resizing left panel: %w", err)
	}
	lv.Title = "State Machine"

	rv, err := g.SetView(rightViewName, leftW, 0, screenW-1, screenH-1)
	if err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("resizing right panel: %w", err)
	}
	rv.Title = "history"

	if a.searchMode {
		if err := a.ensureSearchView(g, screenW, screenH); err != nil {
			return fmt.Errorf("ensuring search view: %w", err)
		}
		// Read the current search query from the editable view buffer and update filter.
		if sv, svErr := g.View(searchViewName); svErr == nil {
			query := strings.TrimSpace(sv.Buffer())
			if query != a.searchQuery {
				a.searchQuery = query
				a.updateFilter()
			}
		}
	}

	if err := a.RenderLeftPanel(g); err != nil {
		return fmt.Errorf("rendering left panel on resize: %w", err)
	}
	if err := a.RenderRightPanel(g, a.executions); err != nil {
		return fmt.Errorf("rendering right panel on resize: %w", err)
	}

	// Keep the detail modal on top during resize.
	if _, err := g.View(detailModalName); err == nil {
		if _, err := g.SetViewOnTop(detailModalName); err != nil {
			return fmt.Errorf("setting detail modal on top: %w", err)
		}
	}

	return nil
}

// ensureSearchView creates (or resizes) the search bar view below the left panel.
func (a *App) ensureSearchView(g *gocui.Gui, screenW, screenH int) error {
	leftW, _ := calcPanelWidths(screenW)
	y0 := screenH - searchBarHeight
	y1 := screenH - 1

	sv, err := g.SetView(searchViewName, 0, y0, leftW-1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("setting search view: %w", err)
	}
	sv.Title = "Search"
	sv.Editable = true
	return nil
}

// RenderLeftPanel re-renders the left panel with the current state machine list.
// When a search filter is active, only the filtered machines are shown.
// Each row shows the machine name with a status bullet on the right.
// If the visible list is empty, displays "(0 state machines)".
// The cursor row is prefixed with "> ".
func (a *App) RenderLeftPanel(g *gocui.Gui) error {
	v, err := g.View(leftViewName)
	if err != nil {
		return fmt.Errorf("getting left panel view: %w", err)
	}

	v.Clear()
	panelW, panelH := v.Size()

	visible := a.visibleMachines()
	if len(visible) == 0 {
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
	if end > len(visible) {
		end = len(visible)
	}

	const bullet = " \u25cf"

	for localIdx, m := range visible[start:end] {
		absIdx := start + localIdx

		truncatedLine := formatSMLine(m.Name, m.LatestStatus, availableWidth)
		line := truncatedLine
		if a.searchQuery != "" {
			if strings.HasSuffix(truncatedLine, bullet) {
				namePart := strings.TrimSuffix(truncatedLine, bullet)
				line = HighlightMatch(namePart, a.searchQuery) + bullet
			} else {
				line = HighlightMatch(truncatedLine, a.searchQuery)
			}
		}

		prefix := "  "
		if absIdx == a.smCursor {
			if cv := g.CurrentView(); cv != nil && cv.Name() == leftViewName {
				prefix = "\033[1;36m> "
				line = line + "\033[0m"
			} else {
				prefix = "> "
			}
		}
		if _, err := fmt.Fprintln(v, prefix+line); err != nil {
			return fmt.Errorf("writing state machine row: %w", err)
		}
	}

	return nil
}

// setMainViewKeybindings registers keybindings for both the left and right panels.
// Left panel: j/k (navigation), q (quit), Tab/h/l/R (focus/refresh).
// Right panel: j/k (move selection), q (quit), Tab/h/l/R (focus/refresh).
func (a *App) setMainViewKeybindings(g *gocui.Gui) error {
	if err := a.bindPanelKeys(g, leftViewName); err != nil {
		return err
	}
	if err := a.bindPanelKeys(g, rightViewName); err != nil {
		return err
	}
	return nil
}

// bindPanelKeys registers the common set of keybindings (j/k/q/Tab/h/l/R/?) for
// the given panel view. j/k move the cursor for the respective panel; q quits;
// Tab/h/l change focus; R refreshes both panels; ? opens the keybinding help modal.
// For the left panel, / enters search mode.
func (a *App) bindPanelKeys(g *gocui.Gui, viewName string) error {
	// j/k navigate the cursor for the focused panel.
	cursorDown := a.smCursorDown
	cursorUp := a.smCursorUp
	if viewName == rightViewName {
		cursorDown = a.execCursorDown
		cursorUp = a.execCursorUp
	}
	if err := g.SetKeybinding(viewName, 'j', gocui.ModNone, cursorDown); err != nil {
		return fmt.Errorf("binding keys for %s: %w", viewName, err)
	}
	if err := g.SetKeybinding(viewName, 'k', gocui.ModNone, cursorUp); err != nil {
		return fmt.Errorf("binding keys for %s: %w", viewName, err)
	}
	if err := g.SetKeybinding(viewName, 'q', gocui.ModNone, quit); err != nil {
		return fmt.Errorf("binding keys for %s: %w", viewName, err)
	}
	if err := g.SetKeybinding(viewName, gocui.KeyTab, gocui.ModNone, a.tabFocus); err != nil {
		return fmt.Errorf("binding keys for %s: %w", viewName, err)
	}
	if err := g.SetKeybinding(viewName, 'R', gocui.ModNone, a.refresh); err != nil {
		return fmt.Errorf("binding keys for %s: %w", viewName, err)
	}
	if err := g.SetKeybinding(viewName, '?', gocui.ModNone, a.ShowHelpModal); err != nil {
		return fmt.Errorf("binding keys for %s: %w", viewName, err)
	}
	// The / key to enter search mode is only available from the left panel.
	if viewName == leftViewName {
		if err := g.SetKeybinding(viewName, '/', gocui.ModNone, a.enterSearchMode); err != nil {
			return fmt.Errorf("binding / for %s: %w", viewName, err)
		}
	}
	// Enter on the right panel opens the input parameter detail modal.
	if viewName == rightViewName {
		if err := g.SetKeybinding(viewName, gocui.KeyEnter, gocui.ModNone, a.ShowDetailModal); err != nil {
			return fmt.Errorf("binding Enter for %s: %w", viewName, err)
		}
	}
	return nil
}

// smCursorDown moves the state machine cursor down one row and re-renders the left panel.
// The bound check is against visibleMachines() so it respects the active search filter.
// If OnSMSelect is set, it is called with the newly selected machine's ARN.
func (a *App) smCursorDown(g *gocui.Gui, v *gocui.View) error {
	visible := a.visibleMachines()
	if a.smCursor < len(visible)-1 {
		a.smCursor++
		if a.OnSMSelect != nil {
			a.OnSMSelect(visible[a.smCursor].ARN)
		}
	}
	return a.RenderLeftPanel(g)
}

// smCursorUp moves the state machine cursor up one row and re-renders the left panel.
// The bound check is against visibleMachines() so it respects the active search filter.
// If OnSMSelect is set, it is called with the newly selected machine's ARN.
func (a *App) smCursorUp(g *gocui.Gui, v *gocui.View) error {
	visible := a.visibleMachines()
	if a.smCursor > 0 {
		a.smCursor--
		if a.OnSMSelect != nil {
			a.OnSMSelect(visible[a.smCursor].ARN)
		}
	}
	return a.RenderLeftPanel(g)
}

// execCursorDown moves the execution history cursor down one row and re-renders the right panel.
// When the cursor is at the last item and more pages are available, it triggers OnLoadMore.
func (a *App) execCursorDown(g *gocui.Gui, v *gocui.View) error {
	if a.execCursor < len(a.executions)-1 {
		a.execCursor++
	} else if a.GetExecNextToken() != nil && !a.loadingMore.Load() {
		if a.OnLoadMore != nil {
			a.OnLoadMore()
		}
	}
	return a.RenderRightPanel(g, a.executions)
}

// execCursorUp moves the execution history cursor up one row and re-renders the right panel.
func (a *App) execCursorUp(g *gocui.Gui, v *gocui.View) error {
	if a.execCursor > 0 {
		a.execCursor--
	}
	return a.RenderRightPanel(g, a.executions)
}

// enterSearchMode activates the incremental search bar below the left panel.
// It sets searchMode, creates the search view, registers search keybindings, and
// moves focus to the search view.
func (a *App) enterSearchMode(g *gocui.Gui, v *gocui.View) error {
	a.searchMode = true
	a.searchQuery = ""
	a.filteredMachines = nil

	screenW, screenH := g.Size()

	// Shrink the left panel to make room for the search bar.
	leftW, _ := calcPanelWidths(screenW)
	leftBottom := screenH - searchBarHeight - 1
	lv, err := g.SetView(leftViewName, 0, 0, leftW-1, leftBottom)
	if err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("resizing left panel for search: %w", err)
	}
	lv.Title = "State Machine"

	if err := a.ensureSearchView(g, screenW, screenH); err != nil {
		return fmt.Errorf("creating search view: %w", err)
	}

	// Register search view keybindings.
	if err := g.SetKeybinding(searchViewName, gocui.KeyEsc, gocui.ModNone, a.exitSearchMode); err != nil {
		return fmt.Errorf("binding Esc on search view: %w", err)
	}
	if err := g.SetKeybinding(searchViewName, gocui.KeyEnter, gocui.ModNone, a.confirmSearch); err != nil {
		return fmt.Errorf("binding Enter on search view: %w", err)
	}

	if _, err := g.SetCurrentView(searchViewName); err != nil {
		return fmt.Errorf("focusing search view: %w", err)
	}
	return a.RenderLeftPanel(g)
}

// exitSearchMode closes the search bar, clears the filter, resets the cursor, and
// returns focus to the left panel.
func (a *App) exitSearchMode(g *gocui.Gui, v *gocui.View) error {
	a.searchMode = false
	a.searchQuery = ""
	a.filteredMachines = nil
	a.smCursor = 0

	g.DeleteKeybindings(searchViewName)
	if err := g.DeleteView(searchViewName); err != nil {
		return fmt.Errorf("deleting search view: %w", err)
	}

	// Restore the left panel to full height.
	screenW, screenH := g.Size()
	leftW, _ := calcPanelWidths(screenW)
	lv, err := g.SetView(leftViewName, 0, 0, leftW-1, screenH-1)
	if err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("restoring left panel: %w", err)
	}
	lv.Title = "State Machine"

	if _, err := g.SetCurrentView(leftViewName); err != nil {
		return fmt.Errorf("focusing left panel: %w", err)
	}
	return a.RenderLeftPanel(g)
}

// confirmSearch keeps the current filter active and returns focus to the left panel.
// The cursor is reset to 0 so it is positioned at the first filtered result.
func (a *App) confirmSearch(g *gocui.Gui, v *gocui.View) error {
	a.searchMode = false
	a.smCursor = 0

	g.DeleteKeybindings(searchViewName)
	if err := g.DeleteView(searchViewName); err != nil {
		return fmt.Errorf("deleting search view: %w", err)
	}

	// Restore the left panel to full height.
	screenW, screenH := g.Size()
	leftW, _ := calcPanelWidths(screenW)
	lv, err := g.SetView(leftViewName, 0, 0, leftW-1, screenH-1)
	if err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("restoring left panel after search confirm: %w", err)
	}
	lv.Title = "State Machine"

	if _, err := g.SetCurrentView(leftViewName); err != nil {
		return fmt.Errorf("focusing left panel after search confirm: %w", err)
	}
	return a.RenderLeftPanel(g)
}
