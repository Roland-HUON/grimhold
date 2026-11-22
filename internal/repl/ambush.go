package repl

import (
	"fmt"
	"math/rand/v2"

	"github.com/Roland-HUON/grimhold/internal/world"
)

const (
	ambushChance = 35
)

func tryAmbush(cfg *Config) bool {
	if cfg.CurrentFoe != nil {
		return false
	}
	currentRegion, ok := world.FindRegionByID(cfg.World, cfg.IDCurrentRegion)
	if !ok {
		fmt.Println("Error: not found the current region")
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
	fmt.Print(colorize(fmt.Sprintf("\nAmbush! A %s leaps out of the shadows!\n", cfg.CurrentFoe.Name), colorRed))

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
