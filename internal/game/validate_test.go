package game

import (
	"testing"

	"github.com/Roland-HUON/grimhold/internal/open5e"
)

func TestValidateAcceptsNewPlayer(t *testing.T) {
	if err := NewPlayer().Validate(); err != nil {
		t.Fatalf("a new player must be valid: %v", err)
	}
}

func TestValidateRejectsBrokenPlayers(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(p *Player)
	}{
		{"level 0", func(p *Player) { p.Level = 0 }},
		{"level too high", func(p *Player) { p.Level = MaxLevel + 1 }},
		{"negative XP", func(p *Player) { p.XP = -1 }},
		{"XP above the next level", func(p *Player) { p.XP = 9223372036854774807 }},
		{"zero max HP", func(p *Player) { p.MaxHP, p.HP = 0, 0 }},
		{"negative HP", func(p *Player) { p.HP = -50 }},
		{"HP above max", func(p *Player) { p.HP = p.MaxHP + 1 }},
		{"zero armor class", func(p *Player) { p.ArmorClass = 0 }},
		{"negative gold", func(p *Player) { p.Gold = -1 }},
		{"weapon without dice", func(p *Player) { p.Weapon.DamageDice = "" }},
		{"weapon with huge dice", func(p *Player) { p.Weapon.DamageDice = "9999999999999d6" }},
		{"inventory with invalid dice", func(p *Player) {
			p.Inventory = []Weapon{{Slug: "club", Name: "Club", DamageDice: "0d6"}}
		}},
		{"spell with huge dice", func(p *Player) {
			p.Spells = []Spell{{Slug: "x", Name: "X", Dice: "1d1000"}}
		}},
		{"level above its XP", func(p *Player) { p.Level, p.XP, p.MaxHP, p.HP = 6, 810, 70, 70 }},
		{"XP overflowing", func(p *Player) { p.Level, p.XP = MaxLevel, maxXP+1 }},
		{"huge max HP", func(p *Player) { p.MaxHP, p.HP = maxHP+1, 1 }},
		{"huge gold", func(p *Player) { p.Gold = maxGold + 1 }},
		{"weapon too strong", func(p *Player) { p.Weapon.DamageDice = "100d100" }},
		{"negative kills", func(p *Player) { p.BestiaryKilled["goblin"] = -1 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPlayer()
			tt.mutate(p)
			if err := p.Validate(); err == nil {
				t.Error("expected an error, got nil")
			}
		})
	}
}

func TestValidateAcceptsConsistentHighLevel(t *testing.T) {
	p := NewPlayer()
	p.Level, p.XP, p.MaxHP, p.HP, p.AttackBonus = 6, 1300, 70, 35, 8
	p.Weapon = Weapon{Slug: "greataxe", Name: "Greataxe", DamageDice: "1d12"}
	if err := p.Validate(); err != nil {
		t.Fatalf("a consistent level 6 player must be valid: %v", err)
	}
}

func TestWinStopsAtMaxLevel(t *testing.T) {
	p := NewPlayer()
	p.Level = MaxLevel
	p.XP = p.NextLevelXP()

	p.Win(&open5e.Monster{Slug: "dragon", CR: 30})

	if p.Level != MaxLevel {
		t.Errorf("level = %d, want it capped at %d", p.Level, MaxLevel)
	}
}
