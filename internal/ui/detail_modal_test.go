package ui

import (
	"testing"

	"github.com/jroimartin/gocui"
	"github.com/myuron/lazysfn/internal/aws"
	"github.com/myuron/lazysfn/internal/config"
)

func TestPrettyPrintJSON_Valid(t *testing.T) {
	input := `{"key":"value","nested":{"a":1}}`
	want := "{\n  \"key\": \"value\",\n  \"nested\": {\n    \"a\": 1\n  }\n}"
	got := prettyPrintJSON(input)
	if got != want {
		t.Errorf("prettyPrintJSON(%q)\ngot:  %q\nwant: %q", input, got, want)
	}
}

func TestPrettyPrintJSON_Invalid(t *testing.T) {
	input := "not json at all"
	got := prettyPrintJSON(input)
	if got != input {
		t.Errorf("prettyPrintJSON(%q) = %q, want unchanged input", input, got)
	}
}

func TestShowDetailModal_NoExecutions(t *testing.T) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		t.Skipf("cannot create gocui in test environment: %v", err)
	}
	defer g.Close()

	app := NewApp([]config.Profile{{Name: "dev"}})
	// No executions set — should be a no-op returning nil.
	if err := app.ShowDetailModal(g, nil); err != nil {
		t.Fatalf("ShowDetailModal() with no executions returned error: %v", err)
	}

	// The detail modal view must NOT exist.
	if _, err := g.View(detailModalName); err == nil {
		t.Errorf("expected detail modal view to not exist when there are no executions")
	}
}

func TestShowDetailModal_ViewCreated(t *testing.T) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		t.Skipf("cannot create gocui in test environment: %v", err)
	}
	defer g.Close()

	app := NewApp([]config.Profile{{Name: "dev"}})
	app.executions = []aws.Execution{
		{ID: "arn:aws:states:us-east-1:123456789012:execution:MySM:exec-1", InputParam: `{"foo":"bar"}`},
	}

	// Create a base view for focus.
	if _, err := g.SetView(rightViewName, 0, 0, 79, 23); err != nil && err != gocui.ErrUnknownView {
		t.Fatalf("SetView right: %v", err)
	}
	if _, err := g.SetCurrentView(rightViewName); err != nil {
		t.Fatalf("SetCurrentView right: %v", err)
	}

	if err := app.ShowDetailModal(g, nil); err != nil {
		t.Fatalf("ShowDetailModal() error: %v", err)
	}

	// The detail modal view must exist.
	if _, err := g.View(detailModalName); err != nil {
		t.Errorf("expected detail modal view to exist, got error: %v", err)
	}

	// Focus must be on the detail modal.
	cv := g.CurrentView()
	if cv == nil || cv.Name() != detailModalName {
		name := ""
		if cv != nil {
			name = cv.Name()
		}
		t.Errorf("expected current view to be %q, got %q", detailModalName, name)
	}
}

func TestShowDetailModal_CloseRestoresFocus(t *testing.T) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		t.Skipf("cannot create gocui in test environment: %v", err)
	}
	defer g.Close()

	app := NewApp([]config.Profile{{Name: "dev"}})
	app.executions = []aws.Execution{
		{ID: "exec-1", InputParam: `{"foo":"bar"}`},
	}

	if _, err := g.SetView(rightViewName, 0, 0, 79, 23); err != nil && err != gocui.ErrUnknownView {
		t.Fatalf("SetView right: %v", err)
	}
	if _, err := g.SetCurrentView(rightViewName); err != nil {
		t.Fatalf("SetCurrentView right: %v", err)
	}

	if err := app.ShowDetailModal(g, nil); err != nil {
		t.Fatalf("ShowDetailModal() error: %v", err)
	}

	if err := app.closeDetailModal(g, rightViewName); err != nil {
		t.Fatalf("closeDetailModal() error: %v", err)
	}

	// Focus must be restored.
	cv := g.CurrentView()
	if cv == nil || cv.Name() != rightViewName {
		name := ""
		if cv != nil {
			name = cv.Name()
		}
		t.Errorf("expected current view to be %q after close, got %q", rightViewName, name)
	}

	// Modal must be gone.
	if _, err := g.View(detailModalName); err == nil {
		t.Errorf("expected detail modal view to be gone after close")
	}
}

func TestTruncateID(t *testing.T) {
	tests := []struct {
		id     string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly-ten", 11, "exactly-ten"},
		{"this-is-a-very-long-execution-id", 15, "this-is-a-very…"},
	}
	for _, tt := range tests {
		got := truncateID(tt.id, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncateID(%q, %d) = %q, want %q", tt.id, tt.maxLen, got, tt.want)
		}
	}
}
