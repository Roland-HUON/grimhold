package world

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

//go:embed world.json
var defaultWorld []byte

type Region struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Danger      int      `json:"danger"`
	Foes        []string `json:"foes"`
	Exits       []string `json:"exits"`
	MinLevel    int      `json:"min_level,omitempty"`
	Boss        string   `json:"boss,omitempty"`
	BossTitle   string   `json:"boss_title,omitempty"`
}

type World struct {
	Start       string            `json:"start"`
	Regions     []Region          `json:"regions"`
	IndexRegion map[string]Region `json:"-"`
}

func Default() (World, error) {
	return Parse(defaultWorld)
}

func CreateWorld(path string) (World, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return World{}, fmt.Errorf("reading world file: %w", err)
	}
	return Parse(data)
}

func Parse(data []byte) (World, error) {
	var configWorld World
	if err := json.Unmarshal(data, &configWorld); err != nil {
		return configWorld, fmt.Errorf("decoding world: %w", err)
	}

	configWorld.IndexRegion = make(map[string]Region, len(configWorld.Regions))
	for _, region := range configWorld.Regions {
		configWorld.IndexRegion[region.ID] = region
	}

	if _, ok := configWorld.IndexRegion[configWorld.Start]; !ok {
		return configWorld, fmt.Errorf("unknown starting region %q", configWorld.Start)
	}

	for _, region := range configWorld.Regions {
		for _, exit := range region.Exits {
			if _, ok := configWorld.IndexRegion[exit]; !ok {
				return configWorld, fmt.Errorf("region %q exits to an unknown region %q", region.ID, exit)
			}
		}
	}

	return configWorld, nil
}

func FindRegionByID(world World, id string) (Region, bool) {
	region, ok := world.IndexRegion[id]
	return region, ok
}
