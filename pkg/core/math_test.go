package core

import (
	"testing"
)

func TestDistWrapped(t *testing.T) {
	tests := []struct {
		name string
		a, b Vector2
		want float64
	}{
		{
			name: "Simple distance",
			a:    Vector2{0, 0},
			b:    Vector2{10, 0},
			want: 10,
		},
		{
			name: "Wrapped distance (X)",
			a:    Vector2{10, 0},
			b:    Vector2{ScreenWidth - 10, 0},
			want: 20, // 10 to 0 is 10, 0 to -10 (wrapped) is 10. Total 20.
		},
		{
			name: "Wrapped distance (Y)",
			a:    Vector2{0, 10},
			b:    Vector2{0, ScreenHeight - 10},
			want: 20,
		},
		{
			name: "Diagonal wrapped",
			a:    Vector2{10, 10},
			b:    Vector2{ScreenWidth - 10, ScreenHeight - 10},
			want: 28.2842712475, // sqrt(20^2 + 20^2)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DistWrapped(tt.a, tt.b); int(got*1000) != int(tt.want*1000) {
				t.Errorf("DistWrapped() = %v, want %v", got, tt.want)
			}
		})
	}
}
