package ui

import (
	"testing"
)

func TestCalcPanelWidths(t *testing.T) {
	tests := []struct {
		totalWidth int
		wantLeft   int
		wantRight  int
	}{
		{120, 30, 90},
		{100, 25, 75},
		{7, 1, 6},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			gotLeft, gotRight := calcPanelWidths(tt.totalWidth)
			if gotLeft != tt.wantLeft || gotRight != tt.wantRight {
				t.Errorf("calcPanelWidths(%d) = (%d, %d), want (%d, %d)",
					tt.totalWidth, gotLeft, gotRight, tt.wantLeft, tt.wantRight)
			}
		})
	}
}

func TestFormatSMLine(t *testing.T) {
	tests := []struct {
		name       string
		smName     string
		status     string
		panelWidth int
		want       string
	}{
		{
			name:       "fits within panel with status",
			smName:     "my-state-machine",
			status:     "SUCCEEDED",
			panelWidth: 20,
			want:       "my-state-machine \u25cf",
		},
		{
			name:       "name truncated when too long with status",
			smName:     "very-long-state-machine-name",
			status:     "FAILED",
			panelWidth: 20,
			want:       "very-long-state-m \u25cf",
		},
		{
			name:       "empty status returns name only",
			smName:     "short",
			status:     "",
			panelWidth: 20,
			want:       "short",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSMLine(tt.smName, tt.status, tt.panelWidth)
			if got != tt.want {
				t.Errorf("formatSMLine(%q, %q, %d) = %q, want %q",
					tt.smName, tt.status, tt.panelWidth, got, tt.want)
			}
		})
	}
}
