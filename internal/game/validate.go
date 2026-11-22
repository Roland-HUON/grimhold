package game

import (
	"errors"
	"fmt"
)

const (
	MaxLevel       = 100
	maxArmorClass  = 50
	maxBonus       = 200
	maxDamageBonus = 50
)

func (p *Player) Validate() error {
	if p.Level < 1 || p.Level > MaxLevel {
		return fmt.Errorf("level %d is out of range (1-%d)", p.Level, MaxLevel)
	}
	if p.XP < 0 || (p.Level < MaxLevel && p.XP >= p.NextLevelXP()) {
		return fmt.Errorf("XP %d does not match level %d", p.XP, p.Level)
	}
	if p.MaxHP <= 0 {
		return fmt.Errorf("max HP %d must be positive", p.MaxHP)
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
	if p.Gold < 0 {
		return errors.New("gold can't be negative")
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
		if err := ValidDice(spell.Dice); err != nil {
			return fmt.Errorf("spell %q: %w", spell.Slug, err)
		}
	}
	for monster, kills := range p.BestiaryKilled {
		if kills < 0 {
			return fmt.Errorf("bestiary: negative kill count for %q", monster)
		}
	}
	return nil
}

func (w Weapon) validate() error {
	if w.Slug == "" {
		return errors.New("missing slug")
	}
	if err := ValidDice(w.DamageDice); err != nil {
		return fmt.Errorf("%q: %w", w.Slug, err)
	}
	if w.DamageBonus < 0 || w.DamageBonus > maxDamageBonus {
		return fmt.Errorf("%q: damage bonus %d is out of range (0-%d)", w.Slug, w.DamageBonus, maxDamageBonus)
	}
	return nil
}
