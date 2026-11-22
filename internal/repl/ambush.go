package repl

import (
	"fmt"
	"math/rand/v2"

	"github.com/Roland-HUON/grimhold/internal/world"
)

const (
	ambushChance       = 35
	ambushMinHPPercent = 25
)

func tryAmbush(cfg *Config) bool {
	if cfg.CurrentFoe != nil {
		return false
	}
	if cfg.Player.HP*100 <= cfg.Player.MaxHP*ambushMinHPPercent {
		return false
	}
	currentRegion, ok := world.FindRegionByID(cfg.World, cfg.IDCurrentRegion)
	if !ok {
		fmt.Printf("Error: current region %q not found\n", cfg.IDCurrentRegion)
		return false
	}
	if len(currentRegion.Foes) == 0 || currentRegion.Boss != "" {
		return false
	}
	if rand.IntN(100) >= ambushChance {
		return false
	}
	if err := spawnFoe(cfg, currentRegion); err != nil {
		fmt.Println("Error:", err)
		return true
	}
	fmt.Print(colorize(fmt.Sprintf("\nAmbush! %s leaps out of the shadows!\n", capitalize(withArticle(cfg.CurrentFoe.Name))), colorRed))

	if err := monsterTurn(cfg); err != nil {
		fmt.Println("Error:", err)
		return true
	}
	if cfg.CurrentFoe != nil {
		fmt.Printf("You: %s | %s: %s\n", hpBar(cfg.Player.HP, cfg.Player.MaxHP, 10), cfg.CurrentFoe.Name, hpBar(cfg.CurrentFoe.HP, cfg.CurrentFoeMaxHP, 10))
		fmt.Println("Type 'fight' to attack or 'flee' to run.")
	}
	return true
}
