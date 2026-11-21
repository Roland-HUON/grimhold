package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Roland-HUON/grimhold/internal/game"
	"github.com/Roland-HUON/grimhold/internal/open5e"
	"github.com/Roland-HUON/grimhold/internal/repl"
	"github.com/Roland-HUON/grimhold/internal/world"
)

const jsonPath = "internal/world/world.json"

func main() {
	gameWorld, err := world.CreateWorld(jsonPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	client := open5e.NewClient(10*time.Second, 5*time.Minute)
	player := game.NewPlayer()
	cfg := &repl.Config{
		World:           gameWorld,
		IDCurrentRegion: gameWorld.Start,
		Client:          client,
		Player:          player,
	}
	if err := repl.Start(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
