package ui

import "testing"

func TestCalcModalHeight(t *testing.T) {
	tests := []struct {
		name         string
		profileCount int
		screenHeight int
		want         int
	}{
		{
			name:         "normal case",
			profileCount: 3,
			screenHeight: 30,
			want:         5,
		},
		{
			name:         "capped by 80% of screen height",
			profileCount: 25,
			screenHeight: 30,
			want:         24,
		},
		{
			name:         "zero profiles",
			profileCount: 0,
			screenHeight: 10,
			want:         2,
		},
		{
			name:         "exactly at 80% boundary",
			profileCount: 10,
			screenHeight: 10,
			want:         8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcModalHeight(tt.profileCount, tt.screenHeight)
			if got != tt.want {
				t.Errorf("calcModalHeight(%d, %d) = %d, want %d", tt.profileCount, tt.screenHeight, got, tt.want)
			}
		})
	}
}

func TestCalcModalPosition(t *testing.T) {
	tests := []struct {
		name    string
		screenW int
		screenH int
		modalW  int
		modalH  int
		wantX0  int
		wantY0  int
		wantX1  int
		wantY1  int
	}{
		{
			name:    "centered in large screen",
			screenW: 100,
			screenH: 40,
			modalW:  40,
			modalH:  10,
			wantX0:  30,
			wantY0:  15,
			wantX1:  70,
			wantY1:  25,
		},
		{
			name:    "centered square",
			screenW: 80,
			screenH: 24,
			modalW:  40,
			modalH:  8,
			wantX0:  20,
			wantY0:  8,
			wantX1:  60,
			wantY1:  16,
		},
		{
			name:    "odd dimensions",
			screenW: 81,
			screenH: 25,
			modalW:  40,
			modalH:  10,
			wantX0:  20,
			wantY0:  7,
			wantX1:  60,
			wantY1:  17,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x0, y0, x1, y1 := calcModalPosition(tt.screenW, tt.screenH, tt.modalW, tt.modalH)
			if x0 != tt.wantX0 || y0 != tt.wantY0 || x1 != tt.wantX1 || y1 != tt.wantY1 {
				t.Errorf("calcModalPosition(%d, %d, %d, %d) = (%d, %d, %d, %d), want (%d, %d, %d, %d)",
					tt.screenW, tt.screenH, tt.modalW, tt.modalH,
					x0, y0, x1, y1,
					tt.wantX0, tt.wantY0, tt.wantX1, tt.wantY1)
			}
		})
	}
}
