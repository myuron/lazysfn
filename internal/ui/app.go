// Package ui provides TUI formatting and rendering utilities.
package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/myuron/lazysfn/internal/aws"
	"github.com/myuron/lazysfn/internal/config"
)

const (
	// modalName is the gocui view name for the profile selection modal.
	modalName = "profileModal"
	// modalWidth is the fixed display width (in columns) of the profile selection modal.
	modalWidth = 40
	// minModalHeight is the minimum height of the profile modal (1 content row + 2 border rows).
	minModalHeight = 3
)

// App manages the overall application state and TUI lifecycle.
type App struct {
	profiles        []config.Profile
	selectedProfile config.Profile
	cursor          int
	gui             *gocui.Gui
	machines        []aws.StateMachine
	smCursor        int
}

// NewApp initializes and returns a new App with the given profiles.
func NewApp(profiles []config.Profile) *App {
	return &App{
		profiles: profiles,
	}
}

// GetSelectedProfile returns the profile chosen by the user after Run completes.
// It returns a zero-value config.Profile if no selection was made.
func (a *App) GetSelectedProfile() config.Profile {
	return a.selectedProfile
}

// Run starts the gocui main loop and displays the profile selection modal.
func (a *App) Run() error {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return fmt.Errorf("creating gui: %w", err)
	}
	defer g.Close()

	a.gui = g
	g.SetManagerFunc(a.layout)

	if err := a.setKeybindings(g); err != nil {
		return fmt.Errorf("setting keybindings: %w", err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return fmt.Errorf("main loop: %w", err)
	}
	return nil
}

// layout is the gocui manager function that renders the profile modal on each resize.
func (a *App) layout(g *gocui.Gui) error {
	screenW, screenH := g.Size()
	modalH := calcModalHeight(len(a.profiles), screenH)
	x0, y0, x1, y1 := calcModalPosition(screenW, screenH, modalWidth, modalH)

	v, err := g.SetView(modalName, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("setting view: %w", err)
	}

	v.Clear()
	v.Title = "Select Profile"

	for i, p := range a.profiles {
		prefix := "  "
		if i == a.cursor {
			prefix = "> "
		}
		if _, err := fmt.Fprintln(v, prefix+p.Name); err != nil {
			return fmt.Errorf("writing profile row: %w", err)
		}
	}

	if _, err := g.SetCurrentView(modalName); err != nil {
		return fmt.Errorf("setting current view: %w", err)
	}

	return nil
}

// setKeybindings registers all keybindings for the profile modal.
func (a *App) setKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding(modalName, 'j', gocui.ModNone, a.cursorDown); err != nil {
		return fmt.Errorf("binding j: %w", err)
	}
	if err := g.SetKeybinding(modalName, 'k', gocui.ModNone, a.cursorUp); err != nil {
		return fmt.Errorf("binding k: %w", err)
	}
	if err := g.SetKeybinding(modalName, gocui.KeyEnter, gocui.ModNone, a.selectProfile); err != nil {
		return fmt.Errorf("binding Enter: %w", err)
	}
	if err := g.SetKeybinding(modalName, 'q', gocui.ModNone, quit); err != nil {
		return fmt.Errorf("binding q: %w", err)
	}
	return nil
}

// cursorDown moves the cursor down one position (does not wrap at the end).
func (a *App) cursorDown(g *gocui.Gui, v *gocui.View) error {
	if a.cursor < len(a.profiles)-1 {
		a.cursor++
	}
	return nil
}

// cursorUp moves the cursor up one position (does not wrap at the beginning).
func (a *App) cursorUp(g *gocui.Gui, v *gocui.View) error {
	if a.cursor > 0 {
		a.cursor--
	}
	return nil
}

// selectProfile sets the selected profile and quits the application.
func (a *App) selectProfile(g *gocui.Gui, v *gocui.View) error {
	if len(a.profiles) > 0 {
		a.selectedProfile = a.profiles[a.cursor]
	}
	return gocui.ErrQuit
}

// quit exits the application by returning gocui.ErrQuit.
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// calcModalHeight calculates the height of the profile modal.
// It returns profileCount + 2 (for borders), capped at 80% of screenHeight.
// The result is always at least minModalHeight to keep the modal usable.
func calcModalHeight(profileCount, screenHeight int) int {
	h := profileCount + 2
	max := int(float64(screenHeight) * 0.8)
	if h > max {
		h = max
	}
	if h < minModalHeight {
		h = minModalHeight
	}
	return h
}

// calcModalPosition calculates the centered position of a modal within the screen.
// Returns (x0, y0, x1, y1) coordinates for gocui.SetView.
func calcModalPosition(screenW, screenH, modalW, modalH int) (x0, y0, x1, y1 int) {
	x0 = (screenW - modalW) / 2
	y0 = (screenH - modalH) / 2
	x1 = x0 + modalW
	y1 = y0 + modalH
	return
}
