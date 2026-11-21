package game

import (
	"fmt"
	"math/rand/v2"

	"github.com/Roland-HUON/grimhold/internal/open5e"
)

type AttackResult struct {
	ActionName string
	Roll       int
	Hit        bool
	Damage     int
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
