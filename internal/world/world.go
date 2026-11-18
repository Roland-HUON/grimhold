package world

import (
	"fmt"
	"os"
	"encoding/json"
)

type region struct {
	Id          string      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Danger      int      `json:"danger"`
	Foes        []string `json:"foes"`
	Exits       []string `json:"exists"`
}

type world struct {
	Start   string   `json:"start"`
	Regions []region `json:"regions"`
}

func createWorld(path string) (world, error) {
	var configWorld world
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

func findRegionById(path string) (map[string]region, error) {
	var listRegion []region
	newWorld, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Issues with the file: %w", err)
		return nil, err
	}

	if err := json.Unmarshal(newWorld, &listRegion); err != nil {
		fmt.Println("Error JSON format: %w", err)
		return nil, err
	}

	indexRegion := make(map[string]region, len(listRegion))
	for _, regions := range listRegion {
		indexRegion[regions.Id] = regions
	}

	return indexRegion, nil
}