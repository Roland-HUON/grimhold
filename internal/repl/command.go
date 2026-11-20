package repl

// tomlouveau78@gmail.com
import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/Roland-HUON/grimhold/internal/world"
)

var errExit = errors.New("exit")

type Config struct {
	World           world.World
	IDCurrentRegion string
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
		fmt.Printf("%s     %-20s\n", cmds[name].Name, cmds[name].Description)
	}
	return nil
}

func commandMap(cfg *Config, args []string) error { 
	regions := cfg.World.Regions

	for _, region := range regions {
		if cfg.IDCurrentRegion == region.ID {
			fmt.Printf("-> Name: %-20s\n Danger level: %-20d\n Exits: %-20s (you are here)\n", region.Name, region.Danger, strings.Join(region.Exits, ", "))
		}
		fmt.Printf("    Name: %-20s\n Danger level: %-20d\n Exits: %-20s\n", region.Name, region.Danger, strings.Join(region.Exits, ", "))
	}
	return nil 
}
func commandTravel(cfg *Config, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: travel <region>")
	}
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
	fmt.Printf("Name: %s\nDescription: %s\nExits: %s\nDanger: %s\n", region.Name, region.Description, strings.Join(region.Exits , ", "), dangerWarning)
	return nil
}

func commandExplore(cfg *Config, args []string) error { return nil }
func commandExit(cfg *Config, args []string) error    { return errExit }
