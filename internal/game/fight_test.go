package game

import (
	"testing"

	"github.com/Roland-HUON/grimhold/internal/open5e"
)

func TestPlayerAttack(t *testing.T) {
	t.Run("always hits armor class 0", func(t *testing.T) {
		for range 200 {
			p := NewPlayer()
			m := &open5e.Monster{HP: 100, ArmorClass: 0}

			res, err := PlayerAttack(p, m)
			if err != nil {
				t.Fatal(err)
			}
			if !res.Hit {
				t.Fatal("attack against AC 0 should always hit")
			}
			if res.Damage < 3 || res.Damage > 10 {
				t.Fatalf("damage = %d, want between 3 and 10", res.Damage)
			}
			if m.HP != 100-res.Damage {
				t.Fatalf("monster HP = %d, want %d", m.HP, 100-res.Damage)
			}
		}
	})

	t.Run("never hits armor class 100", func(t *testing.T) {
		for range 200 {
			m := &open5e.Monster{HP: 100, ArmorClass: 100}
			res, err := PlayerAttack(NewPlayer(), m)
			if err != nil {
				t.Fatal(err)
			}
			if res.Hit || m.HP != 100 {
				t.Fatal("attack against AC 100 should always miss")
			}
		}
	})
}

func TestMonsterAttackWithoutDamageActions(t *testing.T) {
	p := NewPlayer()
	m := &open5e.Monster{Actions: []open5e.Action{{Name: "Multiattack"}}}

	res, err := MonsterAttack(m, p)
	if err != nil {
		t.Fatal(err)
	}
	if res.Hit || p.HP != p.MaxHP {
		t.Error("a monster without damage actions must not hurt the player")
	}
}

func TestWin(t *testing.T) {
	tests := []struct {
		name       string
		cr         float64
		wantXP     int
		wantLevel  int
		wantLevels int
	}{
		{"weak monster gives minimum XP", 0, 10, 1, 0},
		{"exactly enough for level 2", 0.5, 50, 2, 1},
		{"several levels at once", 5, 500, 4, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPlayer()
			startMaxHP := p.MaxHP
			m := &open5e.Monster{Slug: "goblin", CR: tt.cr}

			reward := p.Win(m)

			if reward.XP != tt.wantXP || p.XP != tt.wantXP {
				t.Errorf("XP = %d (reward %d), want %d", p.XP, reward.XP, tt.wantXP)
			}
			if p.Level != tt.wantLevel || reward.LevelsGained != tt.wantLevels {
				t.Errorf("level = %d (+%d), want %d (+%d)", p.Level, reward.LevelsGained, tt.wantLevel, tt.wantLevels)
			}
			if p.MaxHP != startMaxHP+10*tt.wantLevels {
				t.Errorf("MaxHP = %d, want %d", p.MaxHP, startMaxHP+10*tt.wantLevels)
			}
			if p.BestiaryKilled["goblin"] != 1 {
				t.Errorf("bestiary count = %d, want 1", p.BestiaryKilled["goblin"])
			}
			if reward.Gold <= 0 || p.Gold != reward.Gold {
				t.Errorf("gold = %d (reward %d), want a positive amount", p.Gold, reward.Gold)
			}
		})
	}
}
