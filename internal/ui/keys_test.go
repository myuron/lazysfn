package ui

import "testing"

func TestNextFocus(t *testing.T) {
	views := []string{leftViewName, rightViewName}

	tests := []struct {
		name    string
		current string
		views   []string
		want    string
	}{
		{
			name:    "left to right",
			current: leftViewName,
			views:   views,
			want:    rightViewName,
		},
		{
			name:    "right to left wraps",
			current: rightViewName,
			views:   views,
			want:    leftViewName,
		},
		{
			name:    "unknown current returns first",
			current: "unknown",
			views:   views,
			want:    leftViewName,
		},
		{
			name:    "empty views returns empty",
			current: leftViewName,
			views:   []string{},
			want:    "",
		},
		{
			name:    "single view returns itself",
			current: leftViewName,
			views:   []string{leftViewName},
			want:    leftViewName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextFocus(tt.current, tt.views)
			if got != tt.want {
				t.Errorf("nextFocus(%q, %v) = %q; want %q", tt.current, tt.views, got, tt.want)
			}
		})
	}
}

func TestNextSpinnerFrame(t *testing.T) {
	tests := []struct {
		frame int
		want  int
	}{
		{frame: 0, want: 1},
		{frame: 1, want: 2},
		{frame: 2, want: 3},
		{frame: 3, want: 0},
		{frame: 7, want: 0},
		{frame: 4, want: 1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := nextSpinnerFrame(tt.frame)
			if got != tt.want {
				t.Errorf("nextSpinnerFrame(%d) = %d; want %d", tt.frame, got, tt.want)
			}
		})
	}
}
