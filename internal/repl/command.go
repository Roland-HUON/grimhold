package repl

// tomlouveau78@gmail.com
import (
	"errors"
	"fmt"
	"maps"
	"slices"

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
		"travel <region>": {
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
		fmt.Printf("%s     %s \n", name, cmds[name].Description)
	}
	return nil
}

func commandMap(cfg *Config, args []string) error { return nil }
func commandTravel(cfg *Config, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: travel <region>")
	}
	return nil
}
func commandLook(cfg *Config, args []string) error    { return nil }
func commandExplore(cfg *Config, args []string) error { return nil }
func commandExit(cfg *Config, args []string) error    { return errExit }
