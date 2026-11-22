package repl

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Roland-HUON/grimhold/internal/game"
	"github.com/Roland-HUON/grimhold/internal/open5e"
	"github.com/Roland-HUON/grimhold/internal/world"
)

func TestReadLineSurvivesTooLongLines(t *testing.T) {
	long := strings.Repeat("a", maxLineSize+10)
	reader := bufio.NewReaderSize(strings.NewReader("look\r\n"+long+"\nmap\nlast"), 16)

	want := []struct {
		line string
		err  error
	}{
		{"look", nil},
		{"", errLineTooLong},
		{"map", nil},
		{"last", nil},
		{"", io.EOF},
	}
	for i, w := range want {
		line, err := readLine(reader)
		if line != w.line || !errors.Is(err, w.err) {
			t.Fatalf("read %d: got (%q, %v), want (%q, %v)", i, line, err, w.line, w.err)
		}
	}
}

func TestWithArticle(t *testing.T) {
	tests := map[string]string{
		"Goblin":    "a Goblin",
		"Ogre":      "an Ogre",
		"orc":       "an orc",
		"Longsword": "a Longsword",
		"":          "",
	}
	for name, want := range tests {
		if got := withArticle(name); got != want {
			t.Errorf("withArticle(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestToSlug(t *testing.T) {
	tests := []struct {
		words []string
		want  string
	}{
		{[]string{"fire-bolt"}, "fire-bolt"},
		{[]string{"Fire", "Bolt"}, "fire-bolt"},
		{[]string{"ASHEN", "marsh"}, "ashen-marsh"},
	}
	for _, tt := range tests {
		if got := toSlug(tt.words); got != tt.want {
			t.Errorf("toSlug(%q) = %q, want %q", tt.words, got, tt.want)
		}
	}
}

func testWorld(t *testing.T) world.World {
	t.Helper()
	w, err := world.Default()
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestSaveRoundTrip(t *testing.T) {
	w := testWorld(t)
	path := filepath.Join(t.TempDir(), "save.json")

	player := game.NewPlayer()
	player.Gold = 42
	player.BestiaryKilled["goblin"] = 3
	want := saveData{IDCurrentRegion: "ashen-marsh", HasRested: true, Player: *player}

	if err := writeSave(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := readSave(path, w)
	if err != nil {
		t.Fatal(err)
	}
	if got.IDCurrentRegion != want.IDCurrentRegion || !got.HasRested || got.Player.Gold != 42 || got.Player.BestiaryKilled["goblin"] != 3 {
		t.Errorf("round trip mismatch: got %+v", got)
	}
	if _, err := os.Stat(path + ".tmp"); !errors.Is(err, os.ErrNotExist) {
		t.Error("the temporary save file should not remain")
	}
}

func TestReadSaveRejectsBrokenSaves(t *testing.T) {
	w := testWorld(t)
	tests := []struct {
		name    string
		content string
	}{
		{"broken JSON", `{"region": `},
		{"unknown region", `{"region": "nowhere", "player": {}}`},
		{"empty player", `{"region": "crossroads", "player": {}}`},
		{"null player", `{"region": "crossroads", "player": null}`},
		{"negative HP", `{"region": "crossroads", "player": {"hp": -50, "max_hp": -5, "level": 1, "armor_class": 12, "weapon": {"slug": "club", "damage_dice": "1d4"}}}`},
		{"huge spell dice", `{"region": "crossroads", "player": {"hp": 10, "max_hp": 20, "level": 1, "armor_class": 12, "weapon": {"slug": "club", "damage_dice": "1d4"}, "spells": [{"slug": "x", "dice": "9999999999999d6"}]}}`},
		{"boss lair below its level", `{"region": "manticore-spire", "player": {"hp": 20, "max_hp": 20, "level": 1, "armor_class": 12, "weapon": {"slug": "club", "damage_dice": "1d4"}}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "save.json")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := readSave(path, w); err == nil {
				t.Error("expected an error, got nil")
			}
		})
	}

	if _, err := readSave(filepath.Join(t.TempDir(), "missing.json"), w); err == nil {
		t.Error("a missing save should return an error")
	}
}

func TestReadSaveFillsNilBestiary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	content := `{"region": "crossroads", "player": {"hp": 20, "max_hp": 20, "level": 1, "armor_class": 12, "weapon": {"slug": "club", "damage_dice": "1d4"}, "bestiary_killed": null}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := readSave(path, testWorld(t))
	if err != nil {
		t.Fatal(err)
	}
	if data.Player.BestiaryKilled == nil {
		t.Error("a nil bestiary must be replaced by an empty map")
	}
}

func TestHandleLine(t *testing.T) {
	cfg := &Config{World: testWorld(t), IDCurrentRegion: "crossroads", Player: game.NewPlayer()}
	commands := getCommands()

	tests := []struct {
		line     string
		wantExit bool
	}{
		{"", false},
		{"   ", false},
		{"dance", false},
		{"help", false},
		{"LOOK", false},
		{"travel", false},
		{"travel nowhere", false},
		{"fight", false},
		{"loot", false},
		{"exit", true},
		{"EXIT", true},
	}
	for _, tt := range tests {
		if got := handleLine(cfg, commands, tt.line); got != tt.wantExit {
			t.Errorf("handleLine(%q) = %v, want %v", tt.line, got, tt.wantExit)
		}
	}
	if cfg.IDCurrentRegion != "crossroads" || cfg.CurrentFoe != nil {
		t.Error("invalid commands must not change the game state")
	}
}

func TestTryAmbushGuards(t *testing.T) {
	tests := []struct {
		name  string
		setup func(cfg *Config)
	}{
		{"already fighting", func(cfg *Config) { cfg.CurrentFoe = &open5e.Monster{Name: "Goblin", HP: 7} }},
		{"low HP", func(cfg *Config) { cfg.Player.HP = 1 }},
		{"boss region", func(cfg *Config) { cfg.IDCurrentRegion = "manticore-spire" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{World: testWorld(t), IDCurrentRegion: "crossroads", Player: game.NewPlayer()}
			tt.setup(cfg)
			foe := cfg.CurrentFoe
			for range 100 {
				if tryAmbush(cfg) || cfg.CurrentFoe != foe {
					t.Fatal("no ambush should happen")
				}
			}
		})
	}
}
