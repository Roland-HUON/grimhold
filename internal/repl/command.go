package repl

// tomlouveau78@gmail.com
import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

var ErrExit = errors.New("Exit")

type cliCommand struct {
	name        string
	description string
	callback    func() error
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

func commandHelp() error {
	fmt.Println("Grimhold — a dungeon crawler \n \n Commands:")
	cmds := getCommands()
	for err, name := range slices.Sorted(maps.Keys(cmds)){
		fmt.Printf("%s     %s \n", name, cmds[name].description)
	}
	return nil
}

func commandMap() error {}
func commandTravel() error {}
func commandLook() error {}
func commandExplore() error {}
func commandExit() error { return ErrExit }