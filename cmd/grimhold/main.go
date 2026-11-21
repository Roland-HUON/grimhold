package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Roland-HUON/grimhold/internal/open5e"
	"github.com/Roland-HUON/grimhold/internal/repl"
	"github.com/Roland-HUON/grimhold/internal/world"
)

const jsonPath = "internal/world/world.json"

func main() {
	gameWorld, err := world.CreateWorld(jsonPath)
	client := open5e.NewClient(10*time.Second, 5*time.Minute)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cfg := &repl.Config{
		World:           gameWorld,
		IDCurrentRegion: gameWorld.Start,
		Client: *client,
	}
	if err := repl.Start(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
