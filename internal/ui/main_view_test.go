package ui

import (
	"testing"

	"github.com/myuron/lazysfn/internal/aws"
)

func TestVisibleMachines(t *testing.T) {
	machines := []aws.StateMachine{
		{Name: "alpha", ARN: "arn:alpha"},
		{Name: "beta", ARN: "arn:beta"},
		{Name: "gamma", ARN: "arn:gamma"},
	}

	t.Run("returns machines when filteredMachines is nil", func(t *testing.T) {
		app := &App{
			machines:         machines,
			filteredMachines: nil,
		}
		got := app.visibleMachines()
		if len(got) != len(machines) {
			t.Errorf("visibleMachines() with nil filter: got %d, want %d", len(got), len(machines))
		}
		if got[0].Name != "alpha" {
			t.Errorf("visibleMachines() first element: got %q, want %q", got[0].Name, "alpha")
		}
	})

	t.Run("returns filteredMachines when non-nil", func(t *testing.T) {
		filtered := []aws.StateMachine{
			{Name: "alpha", ARN: "arn:alpha"},
		}
		app := &App{
			machines:         machines,
			filteredMachines: filtered,
		}
		got := app.visibleMachines()
		if len(got) != 1 {
			t.Errorf("visibleMachines() with filter: got %d, want 1", len(got))
		}
		if got[0].Name != "alpha" {
			t.Errorf("visibleMachines() first element: got %q, want %q", got[0].Name, "alpha")
		}
	})

	t.Run("returns empty non-nil slice when filter is empty non-nil", func(t *testing.T) {
		app := &App{
			machines:         machines,
			filteredMachines: []aws.StateMachine{},
		}
		got := app.visibleMachines()
		if got == nil {
			t.Error("visibleMachines() with empty filter: got nil, want empty non-nil")
		}
		if len(got) != 0 {
			t.Errorf("visibleMachines() with empty filter: got %d results, want 0", len(got))
		}
	})
}

func TestCurrentSMARN(t *testing.T) {
	machines := []aws.StateMachine{
		{Name: "alpha", ARN: "arn:alpha"},
		{Name: "beta", ARN: "arn:beta"},
	}

	t.Run("returns ARN using visibleMachines (no filter)", func(t *testing.T) {
		app := &App{
			machines:         machines,
			filteredMachines: nil,
			smCursor:         1,
		}
		got := app.CurrentSMARN()
		if got != "arn:beta" {
			t.Errorf("CurrentSMARN() = %q, want %q", got, "arn:beta")
		}
	})

	t.Run("returns ARN using visibleMachines (with filter)", func(t *testing.T) {
		filtered := []aws.StateMachine{
			{Name: "beta", ARN: "arn:beta"},
		}
		app := &App{
			machines:         machines,
			filteredMachines: filtered,
			smCursor:         0,
		}
		got := app.CurrentSMARN()
		if got != "arn:beta" {
			t.Errorf("CurrentSMARN() = %q, want %q", got, "arn:beta")
		}
	})

	t.Run("returns empty string when cursor out of range", func(t *testing.T) {
		app := &App{
			machines:         machines,
			filteredMachines: []aws.StateMachine{},
			smCursor:         0,
		}
		got := app.CurrentSMARN()
		if got != "" {
			t.Errorf("CurrentSMARN() = %q, want %q", got, "")
		}
	})
}

func TestFilterMachinesOnSetMachines(t *testing.T) {
	t.Run("SetMachines recomputes filter when filteredMachines is non-nil (active search)", func(t *testing.T) {
		app := &App{
			searchMode:       true,
			searchQuery:      "foo",
			filteredMachines: []aws.StateMachine{{Name: "foo-old", ARN: "arn:foo-old"}},
		}
		newMachines := []aws.StateMachine{
			{Name: "foo-bar", ARN: "arn:foo-bar"},
			{Name: "baz-qux", ARN: "arn:baz-qux"},
		}
		app.SetMachines(newMachines)

		visible := app.visibleMachines()
		if len(visible) != 1 {
			t.Fatalf("After SetMachines with active filter %q: got %d machines, want 1", "foo", len(visible))
		}
		if visible[0].Name != "foo-bar" {
			t.Errorf("After SetMachines: got %q, want %q", visible[0].Name, "foo-bar")
		}
	})

	t.Run("SetMachines recomputes filter when filteredMachines is non-nil (confirmed search)", func(t *testing.T) {
		app := &App{
			searchMode:       false,
			searchQuery:      "foo",
			filteredMachines: []aws.StateMachine{{Name: "foo-old", ARN: "arn:foo-old"}},
		}
		newMachines := []aws.StateMachine{
			{Name: "foo-new", ARN: "arn:foo-new"},
			{Name: "baz-qux", ARN: "arn:baz-qux"},
		}
		app.SetMachines(newMachines)

		visible := app.visibleMachines()
		if len(visible) != 1 {
			t.Fatalf("After SetMachines with confirmed filter %q: got %d machines, want 1", "foo", len(visible))
		}
		if visible[0].Name != "foo-new" {
			t.Errorf("After SetMachines: got %q, want %q", visible[0].Name, "foo-new")
		}
	})

	t.Run("SetMachines does not filter when filteredMachines is nil", func(t *testing.T) {
		app := &App{
			searchMode:       false,
			filteredMachines: nil,
		}
		newMachines := []aws.StateMachine{
			{Name: "foo-bar", ARN: "arn:foo-bar"},
			{Name: "baz-qux", ARN: "arn:baz-qux"},
		}
		app.SetMachines(newMachines)

		if app.filteredMachines != nil {
			t.Errorf("SetMachines with nil filter: filteredMachines should remain nil, got %v", app.filteredMachines)
		}
	})
}
