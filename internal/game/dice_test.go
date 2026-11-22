package game

import "testing"

func TestRollStaysInRange(t *testing.T) {
	tests := []struct {
		formula  string
		min, max int
	}{
		{"1d4", 1, 4},
		{"1d20", 1, 20},
		{"2d6", 2, 12},
		{"3d8", 3, 24},
	}

	for _, tt := range tests {
		t.Run(tt.formula, func(t *testing.T) {
			for range 1000 {
				got, err := Roll(tt.formula)
				if err != nil {
					t.Fatalf("Roll(%q) returned an error: %v", tt.formula, err)
				}
				if got < tt.min || got > tt.max {
					t.Fatalf("Roll(%q) = %d, want between %d and %d", tt.formula, got, tt.min, tt.max)
				}
			}
		})
	}
}

func TestRollRejectsInvalidFormulas(t *testing.T) {
	invalid := []string{"", "abc", "d6", "2d", "0d6", "2d0", "-1d6", "xd6"}

	for _, formula := range invalid {
		t.Run(formula, func(t *testing.T) {
			if _, err := Roll(formula); err == nil {
				t.Errorf("Roll(%q) should return an error", formula)
			}
		})
	}
}
