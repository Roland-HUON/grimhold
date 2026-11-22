package game

import (
	"errors"
	"fmt"
)

const (
	MaxLevel       = 100
	maxXP          = 1_000_000_000
	maxGold        = 1_000_000_000
	maxHP          = 10_000
	maxKills       = 1_000_000
	maxArmorClass  = 50
	maxBonus       = 200
	maxDamageBonus = 50
	maxPlayerDice  = 10
	maxPlayerFaces = 20
)

func (p *Player) Validate() error {
	if p.Level < 1 || p.Level > MaxLevel {
		return fmt.Errorf("level %d is out of range (1-%d)", p.Level, MaxLevel)
	}
	if p.XP < 0 || p.XP > maxXP || p.XP < levelStartXP(p.Level) || (p.Level < MaxLevel && p.XP >= p.NextLevelXP()) {
		return fmt.Errorf("XP %d does not match level %d", p.XP, p.Level)
	}
	if p.MaxHP <= 0 || p.MaxHP > maxHP {
		return fmt.Errorf("max HP %d is out of range (1-%d)", p.MaxHP, maxHP)
	}
	if p.HP < 1 || p.HP > p.MaxHP {
		return fmt.Errorf("HP %d is out of range (1-%d)", p.HP, p.MaxHP)
	}
	if p.ArmorClass < 1 || p.ArmorClass > maxArmorClass {
		return fmt.Errorf("armor class %d is out of range (1-%d)", p.ArmorClass, maxArmorClass)
	}
	if p.AttackBonus < 0 || p.AttackBonus > maxBonus {
		return fmt.Errorf("attack bonus %d is out of range (0-%d)", p.AttackBonus, maxBonus)
	}
	if p.Gold < 0 || p.Gold > maxGold {
		return fmt.Errorf("gold %d is out of range (0-%d)", p.Gold, maxGold)
	}
	if err := p.Weapon.validate(); err != nil {
		return fmt.Errorf("weapon: %w", err)
	}
	for _, weapon := range p.Inventory {
		if err := weapon.validate(); err != nil {
			return fmt.Errorf("inventory: %w", err)
		}
	}
	for _, spell := range p.Spells {
		if err := validPlayerDice(spell.Dice); err != nil {
			return fmt.Errorf("spell %q: %w", spell.Slug, err)
		}
	}
	for monster, kills := range p.BestiaryKilled {
		if kills < 0 || kills > maxKills {
			return fmt.Errorf("bestiary: kill count %d for %q is out of range (0-%d)", kills, monster, maxKills)
		}
	}
	return nil
}

func (w Weapon) validate() error {
	if w.Slug == "" {
		return errors.New("missing slug")
	}
	if err := validPlayerDice(w.DamageDice); err != nil {
		return fmt.Errorf("%q: %w", w.Slug, err)
	}
	if w.DamageBonus < 0 || w.DamageBonus > maxDamageBonus {
		return fmt.Errorf("%q: damage bonus %d is out of range (0-%d)", w.Slug, w.DamageBonus, maxDamageBonus)
	}
	return nil
}

func levelStartXP(level int) int {
	if level <= 1 {
		return 0
	}
	previous := level - 1
	return previous * previous * 50
}

func validPlayerDice(formula string) error {
	dices, faces, err := parseDice(formula)
	if err != nil {
		return err
	}
	if dices > maxPlayerDice || faces > maxPlayerFaces {
		return fmt.Errorf("dice formula %q is too strong (max %dd%d)", formula, maxPlayerDice, maxPlayerFaces)
	}
	return nil
}
