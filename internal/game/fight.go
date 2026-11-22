package game

import (
	"fmt"
	"math/rand/v2"

	"github.com/Roland-HUON/grimhold/internal/open5e"
)

const (
	FleeDC int = 10
)

type AttackResult struct {
	ActionName string
	Roll       int
	Hit        bool
	Damage     int
}

type VictoryReward struct {
	XP           int
	Gold         int
	LevelsGained int
}

func PlayerAttack(p *Player, m *open5e.Monster) (AttackResult, error) {
	result := AttackResult{
		ActionName: p.Weapon.Name,
	}

	result.Roll = RollDie(20) + p.AttackBonus

	if result.Roll < m.ArmorClass {
		return result, nil
	}

	result.Hit = true
	damage, err := Roll(p.Weapon.DamageDice)
	if err != nil {
		return AttackResult{}, fmt.Errorf("rolling damage: %w", err)
	}

	result.Damage = damage + p.Weapon.DamageBonus
	m.HP -= result.Damage

	return result, nil
}

func MonsterAttack(m *open5e.Monster, p *Player) (AttackResult, error) {
	var playable []open5e.Action
	for _, action := range m.Actions {
		if action.DamageDice != "" {
			playable = append(playable, action)
		}
	}
	if len(playable) == 0 {
		return AttackResult{}, nil
	}

	chosen := playable[rand.IntN(len(playable))]
	result := AttackResult{
		ActionName: chosen.Name,
	}

	result.Roll = RollDie(20) + chosen.AttackBonus
	if result.Roll < p.ArmorClass {
		return result, nil
	}

	result.Hit = true
	damage, err := Roll(chosen.DamageDice)
	if err != nil {
		return AttackResult{}, fmt.Errorf("rolling damage: %w", err)
	}

	result.Damage = damage + chosen.DamageBonus
	p.HP -= result.Damage

	return result, nil
}

func TryFlee() (int, bool) {
	result := RollDie(20)
	return result, result >= FleeDC
}

func (p *Player) Win(m *open5e.Monster) VictoryReward {
	var levelGained int

	xp := int(m.CR * 100)
	if xp < 10 {
		xp = 10
	}
	gold := RollDie(6) + int(m.CR*10)*RollDie(6)
	p.XP += xp
	p.Gold += gold
	p.BestiaryKilled[m.Slug]++
	for p.XP >= p.NextLevelXP() && p.Level < MaxLevel {
		p.Level++
		p.MaxHP += 10
		p.HP = p.MaxHP
		p.AttackBonus += 1
		levelGained++
	}

	return VictoryReward{
		XP:           xp,
		Gold:         gold,
		LevelsGained: levelGained,
	}
}
func (p *Player) Defeat() {
	p.HP = 1
}
