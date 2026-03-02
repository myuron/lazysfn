package ui

import (
	"strings"
	"testing"

	"github.com/jroimartin/gocui"
	"github.com/myuron/lazysfn/internal/config"
)

func TestHelpContent(t *testing.T) {
	content := helpContent()

	keys := []string{"?", "Esc", "q", "R", "j", "k", "h", "l", "Tab", "Enter"}
	for _, key := range keys {
		if !strings.Contains(content, key) {
			t.Errorf("helpContent() missing key %q", key)
		}
	}

	sections := []string{"Global", "Main View", "Profile Modal"}
	for _, sec := range sections {
		if !strings.Contains(content, sec) {
			t.Errorf("helpContent() missing section %q", sec)
		}
	}
}

func TestShowHelpModal_ViewCreated(t *testing.T) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		t.Skipf("cannot create gocui in test environment: %v", err)
	}
	defer g.Close()

	app := NewApp([]config.Profile{{Name: "dev"}})

	// Create a base view so there is a current view before opening the modal.
	if _, err := g.SetView(leftViewName, 0, 0, 79, 23); err != nil && err != gocui.ErrUnknownView {
		t.Fatalf("SetView left: %v", err)
	}
	if _, err := g.SetCurrentView(leftViewName); err != nil {
		t.Fatalf("SetCurrentView left: %v", err)
	}

	if err := app.ShowHelpModal(g, nil); err != nil {
		t.Fatalf("ShowHelpModal() error: %v", err)
	}

	// The help modal view must exist after the call.
	if _, err := g.View(helpModalName); err != nil {
		t.Errorf("expected help modal view to exist, got error: %v", err)
	}

	// Focus must be on the help modal.
	cv := g.CurrentView()
	if cv == nil || cv.Name() != helpModalName {
		name := ""
		if cv != nil {
			name = cv.Name()
		}
		t.Errorf("expected current view to be %q, got %q", helpModalName, name)
	}
}

func TestShowHelpModal_CloseRestoresFocus(t *testing.T) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		t.Skipf("cannot create gocui in test environment: %v", err)
	}
	defer g.Close()

	app := NewApp([]config.Profile{{Name: "dev"}})

	if _, err := g.SetView(leftViewName, 0, 0, 79, 23); err != nil && err != gocui.ErrUnknownView {
		t.Fatalf("SetView left: %v", err)
	}
	if _, err := g.SetCurrentView(leftViewName); err != nil {
		t.Fatalf("SetCurrentView left: %v", err)
	}

	if err := app.ShowHelpModal(g, nil); err != nil {
		t.Fatalf("ShowHelpModal() error: %v", err)
	}

	// Invoke the real close handler (same code path as q/Esc/? keybindings).
	if err := app.closeHelpModal(g, leftViewName); err != nil {
		t.Fatalf("closeHelpModal() error: %v", err)
	}

	// Focus must be restored to the previous view.
	cv := g.CurrentView()
	if cv == nil || cv.Name() != leftViewName {
		name := ""
		if cv != nil {
			name = cv.Name()
		}
		t.Errorf("expected current view to be %q after close, got %q", leftViewName, name)
	}

	// The help modal view must no longer exist.
	if _, err := g.View(helpModalName); err == nil {
		t.Errorf("expected help modal view to be gone after close")
	}
}
