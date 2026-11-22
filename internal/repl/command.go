package repl

// tomlouveau78@gmail.com
import (
	"errors"
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/Roland-HUON/grimhold/internal/game"
	"github.com/Roland-HUON/grimhold/internal/open5e"
	"github.com/Roland-HUON/grimhold/internal/world"
)

var errExit = errors.New("exit")

type Config struct {
	World           world.World
	IDCurrentRegion string
	Client          *open5e.Client
	Player          *game.Player
	CurrentFoe      *open5e.Monster
	BestiaryPrevURL string
	BestiaryNextURL string
}

type cliCommand struct {
	Name        string
	Description string
	Callback    func(cfg *Config, args []string) error
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			Name:        "help",
			Description: "Show this message",
			Callback:    commandHelp,
		},
		"map": {
			Name:        "map",
			Description: "List the known regions",
			Callback:    commandMap,
		},
		"bestiary": {
			Name:        "bestiary [next|prev]",
			Description: "Book of monsters",
			Callback:    commandBestiary,
		},
		"character": {
			Name:        "character",
			Description: "See the character sheet",
			Callback:    commandCharacter,
		},
		"inspect": {
			Name:        "inspect <monster>",
			Description: "Find information about a monster",
			Callback:    commandInspect,
		},
		"travel": {
			Name:        "travel <region>",
			Description: "Travel to a connected region",
			Callback:    commandTravel,
		},
		"fight": {
			Name:        "fight",
			Description: "Fight for your life",
			Callback:    commandFight,
		},
		"flee": {
			Name:        "flee",
			Description: "Run for your life",
			Callback:    commandFlee,
		},
		"look": {
			Name:        "look",
			Description: "Describe your surroundings",
			Callback:    commandLook,
		},
		"explore": {
			Name:        "explore",
			Description: "Wander the region (risk an encounter)",
			Callback:    commandExplore,
		},
		"exit": {
			Name:        "exit",
			Description: "You leave",
			Callback:    commandExit,
		},
	}
}

func commandHelp(cfg *Config, args []string) error {
	fmt.Println("Grimhold — a dungeon crawler \n \n Commands:")
	cmds := getCommands()
	for _, name := range slices.Sorted(maps.Keys(cmds)) {
		fmt.Printf("%-20s %s\n", cmds[name].Name, cmds[name].Description)
	}
	return nil
}

func commandMap(cfg *Config, args []string) error {
	regions := cfg.World.Regions

	for _, region := range regions {
		if cfg.IDCurrentRegion == region.ID {
			fmt.Printf("-> Name: %s\nDanger level: %d\nExits: %s (you are here)\n", region.Name, region.Danger, strings.Join(region.Exits, ", "))
		} else {
			fmt.Printf("   Name: %s\nDanger level: %d\nExits: %s\n", region.Name, region.Danger, strings.Join(region.Exits, ", "))
		}
	}
	return nil
}

func commandBestiary(cfg *Config, args []string) error {
	var pageURL string
	switch {
	case len(args) == 0:
	case len(args) == 1 && args[0] == "next":
		if cfg.BestiaryNextURL == "" {
			return errors.New("no next page")
		}
		pageURL = cfg.BestiaryNextURL
	case len(args) == 1 && args[0] == "prev":
		if cfg.BestiaryPrevURL == "" {
			return errors.New("no previous page")
		}
		pageURL = cfg.BestiaryPrevURL
	default:
		return errors.New("usage: bestiary [next|prev]")
	}

	page, err := cfg.Client.GetPage(pageURL)
	if err != nil {
		return fmt.Errorf("fetching bestiary page: %w", err)
	}

	cfg.BestiaryNextURL = page.NextURL
	cfg.BestiaryPrevURL = page.PrevURL

	fmt.Printf("Bestiary - %d monsters known\n\n", page.Count)
	for _, monster := range page.Monsters {
		fmt.Printf("  %-25s %s\n", monster.Slug, monster.Name)
	}

	var moreText string
	var s []string
	if cfg.BestiaryPrevURL != "" {
		s = append(s, "bestiary prev")
	}
	if cfg.BestiaryNextURL != "" {
		s = append(s, "bestiary next")
	}

	if len(s) > 0 {
		moreText = "More: " + strings.Join(s, " | ")
	}
	fmt.Printf("\nType 'inspect <slug>' for details. %s\n", moreText)
	return nil
}

func commandCharacter(cfg *Config, args []string) error {
	fmt.Printf("%-20s %d\n", "Level:", cfg.Player.Level)
	fmt.Printf("%-20s %d/%d\n", "HP:", cfg.Player.HP, cfg.Player.MaxHP)
	fmt.Printf("%-20s %d/%d\n", "XP:", cfg.Player.XP, cfg.Player.NextLevelXP())
	fmt.Printf("%-20s %d\n", "Attack Bonus:", cfg.Player.AttackBonus)
	fmt.Printf("%-20s %d\n", "Armor Class:", cfg.Player.ArmorClass)
	fmt.Printf("Weapon:\n")
	fmt.Printf("- %-18s %s\n", "Name:", cfg.Player.Weapon.Name)
	fmt.Printf("- %-18s %s\n", "Damage Dice:", cfg.Player.Weapon.DamageDice)
	fmt.Printf("- %-18s %d\n", "Damage Bonus:", cfg.Player.Weapon.DamageBonus)
	fmt.Printf("%-20s %d\n", "Gold:", cfg.Player.Gold)
	if len(cfg.Player.BestiaryKilled) == 0 {
		fmt.Printf("Bestiary killed : None yet\n")
	} else {
		fmt.Printf("Bestiary killed:\n")
		for _, monster := range slices.Sorted(maps.Keys(cfg.Player.BestiaryKilled)) {
			fmt.Printf("- %s x %d\n", monster, cfg.Player.BestiaryKilled[monster])
		}
	}
	return nil
}

func commandInspect(cfg *Config, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: inspect <monster>")
	}

	monsterName := args[0]
	monster, err := cfg.Client.GetMonster(monsterName)
	if err != nil {
		return fmt.Errorf("issue while finding the monster %q: %w", monsterName, err)
	}

	fmt.Printf("Name: %s\nHP: %d\nArmor Class: %d\nChallenge Rating: %s\nActions:\n", monster.Name, monster.HP, monster.ArmorClass, monster.ChallengeRating)
	for _, action := range monster.Actions {
		fmt.Printf("  - %s: %s\n", action.Name, action.Desc)
	}
	return nil
}

func commandTravel(cfg *Config, args []string) error {
	if cfg.CurrentFoe != nil {
		return fmt.Errorf("you can't leave while fighting %s, fight or flee", cfg.CurrentFoe.Name)
	}
	if len(args) != 1 {
		return errors.New("usage: travel <region>")
	}

	destID := args[0]
	region, ok := world.FindRegionByID(cfg.World, destID)
	if !ok {
		return fmt.Errorf("this region ID is not found: %q", destID)
	}

	if destID == cfg.IDCurrentRegion {
		return errors.New("you are already in this region")
	}

	currentRegion, _ := world.FindRegionByID(cfg.World, cfg.IDCurrentRegion)
	if !slices.Contains(currentRegion.Exits, destID) {
		return fmt.Errorf("you can't reach %s from %s", destID, currentRegion.Name)
	}
	cfg.IDCurrentRegion = region.ID
	fmt.Printf("you arrived to the region %s\n", region.Name)
	return nil
}

func commandLook(cfg *Config, args []string) error {
	var dangerWarning string
	region, ok := world.FindRegionByID(cfg.World, cfg.IDCurrentRegion)
	if !ok {
		return errors.New("not found the current region")
	}

	switch region.Danger {
	case 1:
		dangerWarning = "The air feels calm"
	case 2:
		dangerWarning = "Something watching me"
	case 3:
		dangerWarning = "The smell of blood keep floating in the air"
	case 4:
		dangerWarning = "Death lingers here"
	default:
		dangerWarning = "This place belong to the void, a place known to no one, yet it seems like I remember this place... A place where time and space belong to these creatures lurking in the void, where no living beings has ever set foot on this land."
	}
	fmt.Printf("Name: %s\nDescription: %s\nExits: %s\nDanger: %s\n", region.Name, region.Description, strings.Join(region.Exits, ", "), dangerWarning)
	return nil
}

func commandExplore(cfg *Config, args []string) error {
	if cfg.CurrentFoe != nil {
		return fmt.Errorf("you are already fighting %s", cfg.CurrentFoe.Name)
	}
	region, ok := world.FindRegionByID(cfg.World, cfg.IDCurrentRegion)
	if !ok {
		return errors.New("not found the current region")
	}
	if len(region.Foes) == 0 {
		return errors.New("nothing lurks here")
	}

	rolls := region.Danger
	if rolls < 1 {
		rolls = 1
	}

	var worstMonsterFind *open5e.Monster
	for i := 0; i < rolls; i++ {
		monsterIndex := rand.IntN(len(region.Foes))
		monsterSlug := region.Foes[monsterIndex]
		monster, err := cfg.Client.GetMonster(monsterSlug)
		if err != nil {
			return fmt.Errorf("exploring: %w", err)
		}
		if worstMonsterFind == nil || monster.CR > worstMonsterFind.CR {
			worstMonsterFind = &monster
		}
	}
	cfg.CurrentFoe = worstMonsterFind
	fmt.Printf("A %s emerges from the shadow! (HP %d, Armor class %d)\nType 'fight' to attack or 'flee' to run.\n", cfg.CurrentFoe.Name, cfg.CurrentFoe.HP, cfg.CurrentFoe.ArmorClass)
	return nil
}

func commandFight(cfg *Config, args []string) error {
	if cfg.CurrentFoe == nil {
		return errors.New("nothing to fight, try explore")
	}

	playerDamage, err := game.PlayerAttack(cfg.Player, cfg.CurrentFoe)
	if err != nil {
		return fmt.Errorf("player attack: %w", err)
	}
	if playerDamage.Hit {
		fmt.Printf("You swing your %s (%d vs %d): hit for %d damage.\n", cfg.Player.Weapon.Name, playerDamage.Roll, cfg.CurrentFoe.ArmorClass, playerDamage.Damage)
	} else {
		fmt.Printf("You swing your %s (%d vs %d): miss.\n", cfg.Player.Weapon.Name, playerDamage.Roll, cfg.CurrentFoe.ArmorClass)
	}
	if cfg.CurrentFoe.HP <= 0 {
		reward := cfg.Player.Win(cfg.CurrentFoe)
		fmt.Printf("The %s falls!\nYou gained %d XP and %d gold.\n", cfg.CurrentFoe.Name, reward.XP, reward.Gold)
		if reward.LevelsGained >= 1 {
			fmt.Printf("Level up! You are now level %d.\n", cfg.Player.Level)
		}
		cfg.CurrentFoe = nil
		return nil
	}
	if err := monsterTurn(cfg); err != nil {
		return err
	}
	if cfg.CurrentFoe == nil {
		return nil
	}
	fmt.Printf("You: %d/%d HP | %s: %d HP\n", cfg.Player.HP, cfg.Player.MaxHP, cfg.CurrentFoe.Name, cfg.CurrentFoe.HP)
	return nil
}

func monsterTurn(cfg *Config) error {
	monsterDamage, err := game.MonsterAttack(cfg.CurrentFoe, cfg.Player)
	if err != nil {
		return fmt.Errorf("monster attack: %w", err)
	}
	if monsterDamage.ActionName == "" {
		fmt.Printf("The %s hesitates and does nothing.\n", cfg.CurrentFoe.Name)
	} else if monsterDamage.Hit {
		fmt.Printf("%s uses %s (%d vs %d): hit for %d damage.\n", cfg.CurrentFoe.Name, monsterDamage.ActionName, monsterDamage.Roll, cfg.Player.ArmorClass, monsterDamage.Damage)
	} else {
		fmt.Printf("%s uses %s (%d vs %d): miss.\n", cfg.CurrentFoe.Name, monsterDamage.ActionName, monsterDamage.Roll, cfg.Player.ArmorClass)
	}
	if cfg.Player.HP <= 0 {
		cfg.Player.Defeat()
		cfg.CurrentFoe = nil
		cfg.IDCurrentRegion = cfg.World.Start
		region, ok := world.FindRegionByID(cfg.World, cfg.IDCurrentRegion)
		if !ok {
			return errors.New("not found the starting region")
		}
		fmt.Printf("You have been slain...\nYou wake up at %s with 1 HP, your pride wounded.\n", region.Name)
	}
	return nil
}

func commandFlee(cfg *Config, args []string) error {
	if cfg.CurrentFoe == nil {
		return errors.New("nothing to flee from")
	}
	jet, success := game.TryFlee()
	if success {
		fmt.Printf("You escape from the %s (%d vs %d)!\n", cfg.CurrentFoe.Name, jet, game.FleeDC)
		cfg.CurrentFoe = nil
		return nil
	}
	fmt.Printf("You fail to escape (%d vs %d)!\n", jet, game.FleeDC)
	if err := monsterTurn(cfg); err != nil {
		return err
	}
	if cfg.CurrentFoe == nil {
		return nil
	}
	fmt.Printf("You: %d/%d HP | %s: %d HP\n", cfg.Player.HP, cfg.Player.MaxHP, cfg.CurrentFoe.Name, cfg.CurrentFoe.HP)
	return nil
}
func commandExit(cfg *Config, args []string) error { return errExit }
