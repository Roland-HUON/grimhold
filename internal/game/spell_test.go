package game

import (
	"testing"

	"github.com/Roland-HUON/grimhold/internal/open5e"
)

func TestNewSpell(t *testing.T) {
	tests := []struct {
		name     string
		desc     string
		wantDice string
		wantHeal bool
		wantErr  bool
	}{
		{"damage spell", "A ray of cold deals 1d8 cold damage.", "1d8", false, false},
		{"heal spell", "A creature you touch regains a number of hit points equal to 1d8.", "1d8", true, false},
		{"first dice wins", "Deals 2d6 fire damage, or 4d6 on a crit.", "2d6", false, false},
		{"no dice", "You create a harmless sensory effect.", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSpell(open5e.Spell{Slug: "test-spell", Name: "Test Spell", Description: tt.desc})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Dice != tt.wantDice {
				t.Errorf("Dice = %q, want %q", got.Dice, tt.wantDice)
			}
			if got.Heal != tt.wantHeal {
				t.Errorf("Heal = %v, want %v", got.Heal, tt.wantHeal)
			}
			if got.Used {
				t.Error("a new spell must not be used")
			}
		})
	}
}
