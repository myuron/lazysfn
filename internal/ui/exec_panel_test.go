package ui

import (
	"strings"
	"testing"
)

func TestCalcErrorModalHeight(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want int
	}{
		{
			name: "single line message",
			msg:  "something went wrong",
			want: 3, // 1 line + 2 borders
		},
		{
			name: "three line message",
			msg:  "line one\nline two\nline three",
			want: 5, // 3 lines + 2 borders
		},
		{
			name: "empty string",
			msg:  "",
			want: 3, // minimum height: 0 lines + 2 = 2, but minimum is 3
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcErrorModalHeight(tt.msg)
			if got != tt.want {
				t.Errorf("calcErrorModalHeight(%q) = %d, want %d", tt.msg, got, tt.want)
			}
		})
	}
}

func TestDefaultColumnWidths(t *testing.T) {
	tests := []struct {
		name           string
		panelWidth     int
		wantID         int
		wantStatus     int
		wantFailState  int
		wantStartTime  int
		wantStopTime   int
		wantDuration   int
		wantInputParam int
	}{
		{
			name:       "panelWidth=120",
			panelWidth: 120,
			// Fixed widths: ID=30, Status=10, FailState=20, StartTime=19, StopTime=19, Duration=10
			// Separators (spaces between 7 columns = 6 spaces): 6
			// Total fixed: 30+10+20+19+19+10+6 = 114
			// InputParam = 120 - 114 = 6
			wantID:         30,
			wantStatus:     10,
			wantFailState:  20,
			wantStartTime:  19,
			wantStopTime:   19,
			wantDuration:   10,
			wantInputParam: 6,
		},
		{
			name:       "panelWidth=200",
			panelWidth: 200,
			// InputParam = 200 - 114 = 86
			wantID:         30,
			wantStatus:     10,
			wantFailState:  20,
			wantStartTime:  19,
			wantStopTime:   19,
			wantDuration:   10,
			wantInputParam: 86,
		},
		{
			name:       "panelWidth=114 (exactly fixed total, InputParam=0)",
			panelWidth: 114,
			// Fixed total: 30+10+20+19+19+10+6 = 114
			// InputParam = 114 - 114 = 0 (boundary: no space left)
			wantID:         30,
			wantStatus:     10,
			wantFailState:  20,
			wantStartTime:  19,
			wantStopTime:   19,
			wantDuration:   10,
			wantInputParam: 0,
		},
		{
			name:       "panelWidth=100 (narrower than fixed total, InputParam clamped to 0)",
			panelWidth: 100,
			// Fixed total: 114, panelWidth - fixedTotal = -14 → clamped to 0
			wantID:         30,
			wantStatus:     10,
			wantFailState:  20,
			wantStartTime:  19,
			wantStopTime:   19,
			wantDuration:   10,
			wantInputParam: 0,
		},
		{
			name:       "panelWidth=0 (extreme narrow, all clamped to 0)",
			panelWidth: 0,
			// panelWidth - fixedTotal = -114 → clamped to 0
			wantID:         30,
			wantStatus:     10,
			wantFailState:  20,
			wantStartTime:  19,
			wantStopTime:   19,
			wantDuration:   10,
			wantInputParam: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultColumnWidths(tt.panelWidth)
			if got.ID != tt.wantID {
				t.Errorf("ID: got %d, want %d", got.ID, tt.wantID)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("Status: got %d, want %d", got.Status, tt.wantStatus)
			}
			if got.FailState != tt.wantFailState {
				t.Errorf("FailState: got %d, want %d", got.FailState, tt.wantFailState)
			}
			if got.StartTime != tt.wantStartTime {
				t.Errorf("StartTime: got %d, want %d", got.StartTime, tt.wantStartTime)
			}
			if got.StopTime != tt.wantStopTime {
				t.Errorf("StopTime: got %d, want %d", got.StopTime, tt.wantStopTime)
			}
			if got.Duration != tt.wantDuration {
				t.Errorf("Duration: got %d, want %d", got.Duration, tt.wantDuration)
			}
			if got.InputParam != tt.wantInputParam {
				t.Errorf("InputParam: got %d, want %d", got.InputParam, tt.wantInputParam)
			}
		})
	}
}

// TestDefaultColumnWidthsInputParamGrowsWithWidth ensures InputParam increases
// as panelWidth increases.
func TestDefaultColumnWidthsInputParamGrowsWithWidth(t *testing.T) {
	w1 := defaultColumnWidths(120)
	w2 := defaultColumnWidths(200)
	if w2.InputParam <= w1.InputParam {
		t.Errorf("expected InputParam to grow with wider panel: panelWidth=120 gives %d, panelWidth=200 gives %d",
			w1.InputParam, w2.InputParam)
	}
}

// TestCalcErrorModalHeightConsistency verifies line-count logic via strings.Count.
func TestCalcErrorModalHeightConsistency(t *testing.T) {
	msg := "a\nb\nc\nd"
	lines := strings.Count(msg, "\n") + 1
	got := calcErrorModalHeight(msg)
	want := lines + 2
	if got != want {
		t.Errorf("calcErrorModalHeight(%q) = %d, want %d", msg, got, want)
	}
}
