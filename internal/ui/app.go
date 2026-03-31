// Package ui provides TUI formatting and rendering utilities.
package ui

import (
	"fmt"
	"sync"
	"sync/atomic"

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
	execCursor      int
	executions      []aws.Execution
	loading         atomic.Bool
	loadingMore     atomic.Bool
	spinnerFrame    int
	paginationMu    sync.Mutex
	execNextToken   *string
	currentSMARN    string

	// searchMode indicates whether the incremental search bar is active.
	searchMode bool
	// filteredMachines holds the subset of machines matching the current search query.
	// nil means no filter is applied; non-nil (possibly empty) means a filter is active.
	filteredMachines []aws.StateMachine
	// searchQuery holds the current search text while searchMode is active.
	searchQuery string

	// OnProfileSelected is called when a profile is selected in the modal.
	// Set by main.go before calling Run. If nil, Run falls back to ErrQuit (old behavior).
	OnProfileSelected func(g *gocui.Gui) error

	// OnSMSelect is called when the state machine cursor changes.
	// Set by main.go to trigger execution history loading.
	OnSMSelect func(arn string)

	// OnLoadMore is called when the user scrolls past the last execution
	// and more pages are available.
	OnLoadMore func()

	// inMainView tracks whether the TUI is in main view mode (vs profile selection).
	inMainView bool
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
	g.Highlight = true
	g.SelFgColor = gocui.ColorCyan
	g.SetManagerFunc(a.layout)

	if err := a.setKeybindings(g); err != nil {
		return fmt.Errorf("setting keybindings: %w", err)
	}

	// Global Ctrl+C quit (fallback when no view-specific binding matches).
	// Note: this binding is cleared when SetupMainView calls SetManagerFunc,
	// so SetupMainView re-registers it.
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return fmt.Errorf("binding Ctrl+C: %w", err)
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

// selectProfile sets the selected profile and invokes OnProfileSelected if set.
// If OnProfileSelected is nil, it falls back to returning gocui.ErrQuit (old behavior).
func (a *App) selectProfile(g *gocui.Gui, v *gocui.View) error {
	if len(a.profiles) > 0 {
		a.selectedProfile = a.profiles[a.cursor]
	}
	if a.OnProfileSelected != nil {
		return a.OnProfileSelected(g)
	}
	return gocui.ErrQuit
}

// SetLoading sets the loading state of the application.
func (a *App) SetLoading(loading bool) { a.loading.Store(loading) }

// IsLoading returns the current loading state.
func (a *App) IsLoading() bool { return a.loading.Load() }

// SetMachines updates the state machine list without resetting the cursor.
// If a filter is active (searchMode or confirmed search), the filtered list is recomputed.
func (a *App) SetMachines(machines []aws.StateMachine) {
	a.machines = machines
	if a.filteredMachines != nil {
		a.updateFilter()
	}
}

// visibleMachines returns the list of state machines that should be displayed.
// When a search filter is active (filteredMachines != nil) it returns filteredMachines;
// otherwise it returns the full machines list.
func (a *App) visibleMachines() []aws.StateMachine {
	if a.filteredMachines != nil {
		return a.filteredMachines
	}
	return a.machines
}

// updateFilter recomputes filteredMachines from the current machines and searchQuery.
func (a *App) updateFilter() {
	a.filteredMachines = FilterMachines(a.machines, a.searchQuery)
}

// CurrentSMARN returns the ARN of the currently selected state machine.
// Returns "" if no machines are loaded or the cursor is out of range.
func (a *App) CurrentSMARN() string {
	visible := a.visibleMachines()
	if a.smCursor < len(visible) {
		return visible[a.smCursor].ARN
	}
	return ""
}

// SetExecNextToken sets the pagination token for execution history.
func (a *App) SetExecNextToken(token *string) {
	a.paginationMu.Lock()
	a.execNextToken = token
	a.paginationMu.Unlock()
}

// GetExecNextToken returns the current pagination token for execution history.
func (a *App) GetExecNextToken() *string {
	a.paginationMu.Lock()
	defer a.paginationMu.Unlock()
	return a.execNextToken
}

// SetCurrentSMARN sets the ARN of the currently selected state machine for pagination.
func (a *App) SetCurrentSMARN(arn string) {
	a.paginationMu.Lock()
	a.currentSMARN = arn
	a.paginationMu.Unlock()
}

// GetCurrentSMARN returns the ARN used for the current execution history pagination.
func (a *App) GetCurrentSMARN() string {
	a.paginationMu.Lock()
	defer a.paginationMu.Unlock()
	return a.currentSMARN
}

// SetLoadingMore sets the loading-more state for pagination.
func (a *App) SetLoadingMore(loading bool) { a.loadingMore.Store(loading) }

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
