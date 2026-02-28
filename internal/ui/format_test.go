package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/myuron/lazysfn/internal/aws"
)

func TestStatusColor(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"SUCCEEDED", "green"},
		{"FAILED", "red"},
		{"RUNNING", "blue"},
		{"TIMED_OUT", "yellow"},
		{"ABORTED", "gray"},
		{"UNKNOWN_STATUS", ""},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := StatusColor(tt.status)
			if got != tt.want {
				t.Errorf("StatusColor(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestTruncateWithEllipsis(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "shorter than max",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "exceeds max",
			input:  "hello world",
			maxLen: 8,
			want:   "hello...",
		},
		{
			name:   "very short max",
			input:  "hello",
			maxLen: 3,
			want:   "...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateWithEllipsis(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("TruncateWithEllipsis(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{
			name: "seconds only",
			d:    45 * time.Second,
			want: "45s",
		},
		{
			name: "minutes and seconds",
			d:    5*time.Minute + 30*time.Second,
			want: "5m30s",
		},
		{
			name: "hours minutes seconds",
			d:    1*time.Hour + 23*time.Minute + 45*time.Second,
			want: "1h23m45s",
		},
		{
			name: "zero",
			d:    0,
			want: "0s",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.d)
			if got != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{
			name: "normal time",
			t:    time.Date(2024, 3, 15, 10, 30, 45, 0, time.UTC),
			want: "2024-03-15 10:30:45",
		},
		{
			name: "zero value",
			t:    time.Time{},
			want: "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTime(tt.t)
			if got != tt.want {
				t.Errorf("FormatTime(%v) = %q, want %q", tt.t, got, tt.want)
			}
		})
	}
}

func TestFormatExecutionRow(t *testing.T) {
	widths := ColumnWidths{
		ID:         30,
		Status:     10,
		FailState:  20,
		StartTime:  19,
		StopTime:   19,
		Duration:   10,
		InputParam: 20,
	}

	t.Run("RUNNING shows dash for stop time", func(t *testing.T) {
		exec := aws.Execution{
			ID:        "exec-running-123",
			Status:    "RUNNING",
			StartTime: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		}
		row := FormatExecutionRow(exec, widths)
		if !strings.Contains(row, "-") {
			t.Errorf("expected RUNNING row to contain '-' for stop time, got %q", row)
		}
	})

	t.Run("FAILED shows failed state name", func(t *testing.T) {
		exec := aws.Execution{
			ID:          "exec-failed-456",
			Status:      "FAILED",
			FailedState: "ProcessPayment",
			StartTime:   time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
			StopTime:    time.Date(2024, 3, 15, 10, 5, 0, 0, time.UTC),
		}
		row := FormatExecutionRow(exec, widths)
		if !strings.Contains(row, "ProcessPayment") {
			t.Errorf("expected FAILED row to contain 'ProcessPayment', got %q", row)
		}
	})
}
