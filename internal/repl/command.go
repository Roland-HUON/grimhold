package repl

// tomlouveau78@gmail.com
import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/Roland-HUON/grimhold/internal/open5e"
	"github.com/Roland-HUON/grimhold/internal/world"
)

var errExit = errors.New("exit")

type Config struct {
	World           world.World
	IDCurrentRegion string
	Client          *open5e.Client
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

func commandExplore(cfg *Config, args []string) error { return nil }
func commandExit(cfg *Config, args []string) error    { return errExit }
