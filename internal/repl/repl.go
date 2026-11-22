package repl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
	maxLineSize    = 4096
)

var errLineTooLong = fmt.Errorf("line too long (max %d bytes), ignored", maxLineSize)

type input struct {
	line string
	err  error
}

func Start(cfg *Config) error {
	reader := bufio.NewReader(os.Stdin)
	commands := getCommands()

	if !startGame(cfg, reader) {
		fmt.Println("Bye!")
		return nil
	}

	lines := make(chan input)
	done := make(chan struct{})
	defer close(done)
	go readLines(reader, lines, done)

	ticker := time.NewTicker(ambushEvery)
	defer ticker.Stop()

	fmt.Print("grimhold > ")
	for {
		select {
		case in, ok := <-lines:
			if !ok {
				fmt.Println("Bye!")
				return nil
			}
			if in.err != nil {
				fmt.Println("Error:", in.err)
			} else if handleLine(cfg, commands, in.line) {
				fmt.Println("Bye!")
				return nil
			}
			ticker.Reset(ambushEvery)
			fmt.Print("grimhold > ")

		case <-ticker.C:
			if tryAmbush(cfg) {
				fmt.Print("grimhold > ")
			}
		}
	}
}

func readLines(reader *bufio.Reader, lines chan<- input, done <-chan struct{}) {
	defer close(lines)
	for {
		line, err := readLine(reader)
		var in input
		switch {
		case errors.Is(err, errLineTooLong):
			in = input{err: err}
		case err != nil:
			if !errors.Is(err, io.EOF) {
				fmt.Println("Error: reading input:", err)
			}
			return
		default:
			in = input{line: line}
		}
		select {
		case lines <- in:
		case <-done:
			return
		}
	}
}

func readLine(reader *bufio.Reader) (string, error) {
	var line []byte
	tooLong := false
	for {
		chunk, isPrefix, err := reader.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) && len(line) > 0 {
				break
			}
			return "", err
		}
		if len(line)+len(chunk) > maxLineSize {
			tooLong = true
		} else {
			line = append(line, chunk...)
		}
		if !isPrefix {
			break
		}
	}
	if tooLong {
		return "", errLineTooLong
	}
	return string(line), nil
}

func handleLine(cfg *Config, commands map[string]cliCommand, line string) bool {
	args := strings.Fields(strings.ToLower(line))
	if len(args) == 0 {
		return false
	}

	cmd, ok := commands[args[0]]
	if !ok {
		fmt.Printf("Unknown command %q. Type 'help' for any help.\n", shorten(args[0], 30))
		return false
	}
	if !cmd.takesArgs && len(args) > 1 {
		fmt.Printf("Error: '%s' takes no argument\n", args[0])
		return false
	}
	if err := cmd.callback(cfg, args[1:]); err != nil {
		if errors.Is(err, errExit) {
			return true
		}
		fmt.Println("Error:", err)
	}
	return false
}

func startGame(cfg *Config, reader *bufio.Reader) bool {
	if cfg.Continue {
		err := commandLoad(cfg, nil)
		if err == nil {
			return true
		}
		fmt.Println("Error:", err)
		fmt.Println("Starting a new game instead.")
	}
	return chooseSpells(cfg, reader)
}

func chooseSpells(cfg *Config, reader *bufio.Reader) bool {
	fmt.Printf("Choose %d spells (or type 'load' to continue your saved game, 'exit' to quit):\n", spellsToChoose)
	for _, slug := range starterSpells {
		fmt.Printf("  - %s\n", slug)
	}

	for len(cfg.Player.Spells) < spellsToChoose {
		fmt.Printf("spell %d/%d > ", len(cfg.Player.Spells)+1, spellsToChoose)
		line, err := readLine(reader)
		if errors.Is(err, errLineTooLong) {
			fmt.Println("Error:", err)
			continue
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Println("Error: reading input:", err)
			}
			return false
		}
		words := strings.Fields(strings.ToLower(line))
		if len(words) == 0 {
			continue
		}
		choice := toSlug(words)

		switch choice {
		case "exit":
			return false
		case "load":
			if err := commandLoad(cfg, nil); err != nil {
				fmt.Println("Error:", err)
				continue
			}
			return true
		}

		if !slices.Contains(starterSpells, choice) {
			fmt.Printf("Unknown spell %q, pick one from the list.\n", shorten(choice, 30))
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
	if start, ok := world.FindRegionByID(cfg.World, cfg.IDCurrentRegion); ok {
		warnPrefetch(prefetchFoes(cfg.Client, start))
	}
	return true
}

func shorten(text string, max int) string {
	runes := []rune(text)
	if len(runes) <= max {
		return text
	}
	return string(runes[:max]) + "..."
}

func toSlug(words []string) string {
	return strings.ToLower(strings.Join(words, "-"))
}
