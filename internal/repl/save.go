package repl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/Roland-HUON/grimhold/internal/game"
	"github.com/Roland-HUON/grimhold/internal/world"
)

const (
	saveFile = "save.json"
)

type saveData struct {
	IDCurrentRegion string      `json:"region"`
	HasRested       bool        `json:"hasrested"`
	Player          game.Player `json:"player"`
}

func commandSave(cfg *Config, args []string) error {
	if cfg.CurrentFoe != nil {
		return fmt.Errorf("you can't save while fighting %s", cfg.CurrentFoe.Name)
	}
	data := saveData{
		IDCurrentRegion: cfg.IDCurrentRegion,
		HasRested:       cfg.HasRested,
		Player:          *cfg.Player,
	}
	if err := writeSave(saveFile, data); err != nil {
		return err
	}
	fmt.Println("Game saved.")
	return nil
}

func writeSave(path string, data saveData) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding save: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, bytes, 0o644); err != nil {
		return fmt.Errorf("writing save: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("writing save: %w", err)
	}
	return nil
}

func commandLoad(cfg *Config, args []string) error {
	if cfg.CurrentFoe != nil {
		return fmt.Errorf("you can't load while fighting %s", cfg.CurrentFoe.Name)
	}
	data, err := readSave(saveFile, cfg.World)
	if err != nil {
		return err
	}
	region, _ := world.FindRegionByID(cfg.World, data.IDCurrentRegion)

	cfg.IDCurrentRegion = data.IDCurrentRegion
	cfg.HasRested = data.HasRested
	cfg.Player = &data.Player
	cfg.CanLoot = false
	cfg.CurrentFoe = nil
	fmt.Printf("Game loaded. You are at %s.\n", region.Name)
	warnPrefetch(prefetchFoes(cfg.Client, region))
	return nil
}

func readSave(path string, w world.World) (saveData, error) {
	var data saveData
	bytes, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return data, errors.New("no save found")
	}
	if err != nil {
		return data, fmt.Errorf("reading save: %w", err)
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return data, fmt.Errorf("decoding save: %w", err)
	}
	if err := validateSave(&data, w); err != nil {
		return data, fmt.Errorf("invalid save: %w", err)
	}
	return data, nil
}

func validateSave(data *saveData, w world.World) error {
	region, ok := world.FindRegionByID(w, data.IDCurrentRegion)
	if !ok {
		return fmt.Errorf("unknown region %q", data.IDCurrentRegion)
	}
	if data.Player.BestiaryKilled == nil {
		data.Player.BestiaryKilled = make(map[string]int)
	}
	if err := data.Player.Validate(); err != nil {
		return fmt.Errorf("player: %w", err)
	}
	if data.Player.Level < region.MinLevel {
		return fmt.Errorf("%s requires level %d, the player is level %d", region.Name, region.MinLevel, data.Player.Level)
	}
	return nil
}
