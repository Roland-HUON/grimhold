package repl

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/Roland-HUON/grimhold/internal/game"
	"github.com/Roland-HUON/grimhold/internal/world"
)

var starterSpells = []string{"fire-bolt", "ray-of-frost", "thunderwave", "burning-hands", "guiding-bolt", "cure-wounds", "healing-word"}

const (
	spellsToChoose = 2
	ambushEvery    = 20 * time.Second
)

func Start(cfg *Config) error {
	scanner := bufio.NewScanner(os.Stdin)
	commands := getCommands()

	if !chooseSpells(cfg, scanner) {
		fmt.Println("Bye!")
		return nil
	}
	if start, ok := world.FindRegionByID(cfg.World, cfg.IDCurrentRegion); ok {
		warnPrefetch(prefetchFoes(cfg.Client, start))
	}

	lines := make(chan string)
	go func() {
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()

	ticker := time.NewTicker(ambushEvery)
	defer ticker.Stop()

	fmt.Print("grimhold > ")
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				fmt.Println("Bye!")
				return nil
			}
			if handleLine(cfg, commands, line) {
				fmt.Println("Bye!")
				return nil
			}
			fmt.Print("grimhold > ")

		case <-ticker.C:
			if tryAmbush(cfg) {
				fmt.Print("grimhold > ")
			}
		}
	}
}

func handleLine(cfg *Config, commands map[string]cliCommand, line string) bool {
	args := strings.Fields(line)
	if len(args) == 0 {
		return false
	}

	cmd, ok := commands[args[0]]
	if !ok {
		fmt.Printf("Unknown command %q. Type 'help' for any help.\n", args[0])
		return false
	}
	if err := cmd.Callback(cfg, args[1:]); err != nil {
		if errors.Is(err, errExit) {
			return true
		}
		fmt.Println("Error:", err)
	}
	return false
}

func chooseSpells(cfg *Config, scanner *bufio.Scanner) bool {
	fmt.Printf("Choose %d spells:\n", spellsToChoose)
	for _, slug := range starterSpells {
		fmt.Printf("  - %s\n", slug)
	}

	for len(cfg.Player.Spells) < spellsToChoose {
		fmt.Printf("spell %d/%d > ", len(cfg.Player.Spells)+1, spellsToChoose)
		if !scanner.Scan() {
			return false
		}
		choice := strings.ToLower(strings.TrimSpace(scanner.Text()))

		if !slices.Contains(starterSpells, choice) {
			fmt.Printf("Unknown spell %q, pick one from the list.\n", choice)
			continue
		}
		known := slices.ContainsFunc(cfg.Player.Spells, func(s game.Spell) bool {
			return s.Slug == choice
		})
		if known {
			fmt.Printf("You already know %s, pick another one.\n", choice)
			continue
		}

		raw, err := cfg.Client.GetSpell(choice)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		spell, err := game.NewSpell(raw)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		cfg.Player.Spells = append(cfg.Player.Spells, spell)
		fmt.Printf("You learned %s (%s).\n", spell.Name, spell.Dice)
	}
	return true
}
