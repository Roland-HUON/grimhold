package world

import (
	"fmt"
	"os"
	"encoding/json"
)

type Region struct {
	Id          string      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Danger      int      `json:"danger"`
	Foes        []string `json:"foes"`
	Exits       []string `json:"exits"`
}

type World struct {
	Start   string   `json:"start"`
	Regions []Region `json:"regions"`
}

func CreateWorld(path string) (World, error) {
	var configWorld World
	newWorld, err := os.ReadFile(path)
	
	if err != nil {
		fmt.Println("Issues with the file: %w", err)
		return configWorld, err
	}

	if err := json.Unmarshal(newWorld, &configWorld); err != nil {
		fmt.Println("Error JSON format: %w", err)
		return configWorld, err
	}

	return configWorld, nil
}

func FindRegionById(path string) (map[string]Region, error) {
	var listRegion []Region
	newWorld, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Issues with the file: %w", err)
	}

	if err := json.Unmarshal(newWorld, &listRegion); err != nil {
		return nil, fmt.Errorf("Error JSON format: %w", err)
	}

	indexRegion := make(map[string]Region, len(listRegion))
	for _, regions := range listRegion {
		indexRegion[regions.Id] = regions
	}

	return indexRegion, nil
}