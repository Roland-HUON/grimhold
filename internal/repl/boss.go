package repl

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Roland-HUON/grimhold/internal/world"
)

func encounterBoss(cfg *Config, region world.Region) error {
	if slices.Contains(cfg.Player.BossesSlain, region.Boss) {
		return fmt.Errorf("%s is silent, its master is dead", region.Name)
	}

	boss, err := cfg.Client.GetMonster(region.Boss)
	if err != nil {
		return fmt.Errorf("summoning the boss: %w", err)
	}

	species := boss.Name
	if region.BossTitle != "" {
		boss.Name = region.BossTitle
	}

	var attacks []string
	for _, action := range boss.Actions {
		if action.DamageDice != "" {
			attacks = append(attacks, action.Name)
		}
	}

	cfg.CanLoot = false
	cfg.CurrentFoe = &boss
	cfg.CurrentFoeMaxHP = boss.HP

	banner := fmt.Sprintf("*** BOSS: %s, the %s ***", boss.Name, species)
	fmt.Println(colorize(banner, colorBold+colorMagenta))
	fmt.Printf("HP %d, Armor class %d. Attacks: %s.\n", boss.HP, boss.ArmorClass, strings.Join(attacks, ", "))
	fmt.Println("Type 'fight' to attack or 'flee' to run.")
	return nil
}
