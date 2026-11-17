package repl

// tomlouveau78@gmail.com
import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

var errExit = errors.New("Exit")

type cliCommand struct {
	name        string
	description string
	callback    func(args []string) error
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Show this message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "List the known regions",
			callback:    commandMap,
		},
		"travel <region>": {
			name:        "travel <region>",
			description: "Travel to a connected region",
			callback:    commandTravel,
		},
		"look": {
			name:        "look",
			description: "Describe your surroundings",
			callback:    commandLook,
		},
		"explore": {
			name:        "explore",
			description: "Wander the region (risk an encounter)",
			callback:    commandExplore,
		},
		"exit": {
			name:        "exit",
			description: "You leave",
			callback:    commandExit,
		},
	}
}

func commandHelp(args []string) error {
	fmt.Println("Grimhold — a dungeon crawler \n \n Commands:")
	cmds := getCommands()
	for _, name := range slices.Sorted(maps.Keys(cmds)) {
		fmt.Printf("%s     %s \n", name, cmds[name].description)
	}
	return nil
}

func commandMap(args []string) error { return nil }
func commandTravel(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: travel <region>")
	}
	return nil
}
func commandLook(args []string) error    { return nil }
func commandExplore(args []string) error { return nil }
func commandExit(args []string) error    { return errExit }
