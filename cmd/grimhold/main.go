package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Roland-HUON/grimhold/internal/game"
	"github.com/Roland-HUON/grimhold/internal/open5e"
	"github.com/Roland-HUON/grimhold/internal/repl"
	"github.com/Roland-HUON/grimhold/internal/world"
)

func main() {
	continueGame := flag.Bool("continue", false, "load save.json at start instead of creating a new character")
	flag.Parse()

	gameWorld, err := world.Default()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	client := open5e.NewClient(10*time.Second, 5*time.Minute)
	defer client.Close()

	cfg := &repl.Config{
		World:           gameWorld,
		IDCurrentRegion: gameWorld.Start,
		Client:          client,
		Player:          game.NewPlayer(),
		Continue:        *continueGame,
	}
	if err := repl.Start(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
