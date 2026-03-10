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

func TestFilterMachines(t *testing.T) {
	machines := []aws.StateMachine{
		{Name: "foo-machine", ARN: "arn:foo"},
		{Name: "bar-machine", ARN: "arn:bar"},
		{Name: "baz-machine", ARN: "arn:baz"},
	}

	t.Run("empty query returns all machines", func(t *testing.T) {
		got := FilterMachines(machines, "")
		if len(got) != len(machines) {
			t.Errorf("FilterMachines with empty query: got %d machines, want %d", len(got), len(machines))
		}
	})

	t.Run("query matches single machine", func(t *testing.T) {
		got := FilterMachines(machines, "foo")
		if len(got) != 1 {
			t.Fatalf("FilterMachines(%q): got %d results, want 1", "foo", len(got))
		}
		if got[0].Name != "foo-machine" {
			t.Errorf("FilterMachines(%q): got %q, want %q", "foo", got[0].Name, "foo-machine")
		}
	})

	t.Run("query matches multiple machines", func(t *testing.T) {
		got := FilterMachines(machines, "ba")
		if len(got) != 2 {
			t.Fatalf("FilterMachines(%q): got %d results, want 2", "ba", len(got))
		}
	})

	t.Run("case insensitive match", func(t *testing.T) {
		got := FilterMachines(machines, "FOO")
		if len(got) != 1 {
			t.Fatalf("FilterMachines(%q): got %d results, want 1", "FOO", len(got))
		}
		if got[0].Name != "foo-machine" {
			t.Errorf("FilterMachines(%q): got %q, want %q", "FOO", got[0].Name, "foo-machine")
		}
	})

	t.Run("no match returns empty non-nil slice", func(t *testing.T) {
		got := FilterMachines(machines, "zzz")
		if got == nil {
			t.Error("FilterMachines with no match: got nil, want empty non-nil slice")
		}
		if len(got) != 0 {
			t.Errorf("FilterMachines with no match: got %d results, want 0", len(got))
		}
	})
}

func TestHighlightMatch(t *testing.T) {
	cases := []struct {
		desc  string
		input string
		query string
		want  string
	}{
		{"empty query returns name unchanged", "foo", "", "foo"},
		{"no match returns name unchanged", "foo", "bar", "foo"},
		{"exact match wraps whole name", "foo", "foo", "\033[33mfoo\033[0m"},
		{"partial match wraps substring", "foobar", "oba", "fo\033[33moba\033[0mr"},
		{"case insensitive: query upper, name lower", "foobar", "FOO", "\033[33mfoo\033[0mbar"},
		{"case insensitive: query lower, name upper", "FOOBAR", "foo", "\033[33mFOO\033[0mBAR"},
		{"only first occurrence highlighted", "abab", "ab", "\033[33mab\033[0mab"},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := HighlightMatch(tc.input, tc.query)
			if got != tc.want {
				t.Errorf("HighlightMatch(%q, %q) = %q, want %q", tc.input, tc.query, got, tc.want)
			}
		})
	}
}

func TestFormatExecutionRow(t *testing.T) {
	widths := ColumnWidths{
		ID:         12,
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
