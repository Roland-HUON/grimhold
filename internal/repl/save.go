package repl

import "github.com/Roland-HUON/grimhold/internal/game"

const (
	saveFile = "save.json"
)

type saveData struct {
	IDCurrentRegion string      `json:"region"`
	HasRested       bool        `json:"hasrested"`
	Player          game.Player `json:"player"`
}
