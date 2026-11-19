package world

import (
	"encoding/json"
	"fmt"
	"os"
)

type Region struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Danger      int      `json:"danger"`
	Foes        []string `json:"foes"`
	Exits       []string `json:"exits"`
}

type World struct {
	Start       string            `json:"start"`
	Regions     []Region          `json:"regions"`
	IndexRegion map[string]Region `json:"-"`
}

func CreateWorld(path string) (World, error) {
	var configWorld World
	newWorld, err := os.ReadFile(path)

	if err != nil {
		return configWorld, fmt.Errorf("issues with the file: %w", err)
	}

	if err := json.Unmarshal(newWorld, &configWorld); err != nil {
		return configWorld, fmt.Errorf("error JSON format: %w", err)
	}

	configWorld.IndexRegion = make(map[string]Region, len(configWorld.Regions))
	for _, region := range configWorld.Regions {
		configWorld.IndexRegion[region.ID] = region
	}

	if _, ok := configWorld.IndexRegion[configWorld.Start]; !ok {
		return configWorld, fmt.Errorf("error with the starting region : %q", configWorld.Start)
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
