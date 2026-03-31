package ui

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/jroimartin/gocui"
)

const (
	// detailModalName is the gocui view name for the input parameter detail modal.
	detailModalName = "detailModal"
)

// ShowDetailModal displays the full input parameter of the currently selected
// execution in a centered, scrollable modal. If the input is valid JSON it is
// pretty-printed; otherwise it is shown as-is. The modal closes on Esc or q
// and supports j/k scrolling.
func (a *App) ShowDetailModal(g *gocui.Gui, v *gocui.View) error {
	if len(a.executions) == 0 || a.execCursor < 0 || a.execCursor >= len(a.executions) {
		return nil
	}

	// Guard against opening the modal when it is already displayed.
	if _, err := g.View(detailModalName); err == nil {
		return nil
	}

	exec := a.executions[a.execCursor]
	content := prettyPrintJSON(exec.InputParam)

	prevView := ""
	if cv := g.CurrentView(); cv != nil {
		prevView = cv.Name()
	}

	screenW, screenH := g.Size()
	modalW := int(float64(screenW) * 0.8)
	modalH := int(float64(screenH) * 0.8)
	if modalW < 20 {
		modalW = 20
	}
	if modalH < 5 {
		modalH = 5
	}
	x0, y0, x1, y1 := calcModalPosition(screenW, screenH, modalW, modalH)

	view, err := g.SetView(detailModalName, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return fmt.Errorf("setting detail modal view: %w", err)
	}
	view.Clear()
	view.Wrap = true

	// Title: truncated execution ID.
	title := truncateID(exec.ID, 30)
	view.Title = "Input: " + title

	if _, err := fmt.Fprint(view, content); err != nil {
		return fmt.Errorf("writing detail content: %w", err)
	}

	closeModal := func(g *gocui.Gui, v *gocui.View) error {
		return a.closeDetailModal(g, prevView)
	}

	if err := g.SetKeybinding(detailModalName, gocui.KeyEsc, gocui.ModNone, closeModal); err != nil {
		return fmt.Errorf("binding Esc on detail modal: %w", err)
	}
	if err := g.SetKeybinding(detailModalName, 'q', gocui.ModNone, closeModal); err != nil {
		return fmt.Errorf("binding q on detail modal: %w", err)
	}
	if err := g.SetKeybinding(detailModalName, 'j', gocui.ModNone, scrollDetailDown); err != nil {
		return fmt.Errorf("binding j on detail modal: %w", err)
	}
	if err := g.SetKeybinding(detailModalName, 'k', gocui.ModNone, scrollDetailUp); err != nil {
		return fmt.Errorf("binding k on detail modal: %w", err)
	}

	if _, err := g.SetCurrentView(detailModalName); err != nil {
		return fmt.Errorf("setting detail modal current view: %w", err)
	}
	return nil
}

// closeDetailModal removes the detail modal view, cleans up its keybindings,
// and restores focus to prevView.
func (a *App) closeDetailModal(g *gocui.Gui, prevView string) error {
	g.DeleteKeybindings(detailModalName)
	if err := g.DeleteView(detailModalName); err != nil {
		return fmt.Errorf("deleting detail modal: %w", err)
	}
	if prevView != "" {
		if _, err := g.SetCurrentView(prevView); err != nil {
			return fmt.Errorf("restoring focus from detail modal: %w", err)
		}
	}
	return nil
}

// scrollDetailDown scrolls the detail modal content down by one line.
// It stops at the last line of content so the view does not scroll into empty space.
func scrollDetailDown(g *gocui.Gui, v *gocui.View) error {
	_, oy := v.Origin()
	_, viewH := v.Size()
	lines := v.ViewBufferLines()
	// Do not scroll past the content.
	if oy+viewH >= len(lines) {
		return nil
	}
	if err := v.SetOrigin(0, oy+1); err != nil {
		return nil // at bottom, ignore
	}
	return nil
}

// scrollDetailUp scrolls the detail modal content up by one line.
func scrollDetailUp(g *gocui.Gui, v *gocui.View) error {
	ox, oy := v.Origin()
	if oy > 0 {
		if err := v.SetOrigin(ox, oy-1); err != nil {
			return nil // ignore
		}
	}
	return nil
}

// prettyPrintJSON attempts to format raw as indented JSON.
// If raw is not valid JSON, it is returned unchanged.
func prettyPrintJSON(raw string) string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(raw), "", "  "); err != nil {
		return raw
	}
	return buf.String()
}

// truncateID truncates an execution ID to maxLen characters, appending "…" if truncated.
func truncateID(id string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(id)
	if len(runes) <= maxLen {
		return id
	}
	return string(runes[:maxLen-1]) + "…"
}
