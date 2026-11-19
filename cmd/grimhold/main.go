package main

import (
	"fmt"
	"os"
	
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

	cfg := &repl.Config{
		World:           gameWorld,
		IDCurrentRegion: gameWorld.Start,
	}
	if err := repl.Start(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
