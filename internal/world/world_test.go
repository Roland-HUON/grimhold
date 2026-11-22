package world

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultWorld(t *testing.T) {
	w, err := Default()
	if err != nil {
		t.Fatalf("the embedded world.json must be valid: %v", err)
	}
	if _, ok := FindRegionByID(w, w.Start); !ok {
		t.Errorf("start region %q not found", w.Start)
	}
	for _, region := range w.Regions {
		if len(region.Foes) == 0 && region.Boss == "" {
			t.Errorf("region %q has neither foes nor boss", region.ID)
		}
	}
}

func TestCreateWorldFromFile(t *testing.T) {
	if _, err := CreateWorld("world.json"); err != nil {
		t.Fatalf("CreateWorld(world.json) = %v", err)
	}
}

func TestCreateWorldRejectsInvalidFiles(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"broken JSON", `{"start": `},
		{"unknown start", `{"start": "nowhere", "regions": [{"id": "a", "exits": []}]}`},
		{"unknown exit", `{"start": "a", "regions": [{"id": "a", "exits": ["b"]}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "world.json")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := CreateWorld(path); err == nil {
				t.Error("expected an error, got nil")
			}
		})
	}

	if _, err := CreateWorld(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("a missing file should return an error")
	}
}
