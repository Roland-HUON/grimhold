package game

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Roland-HUON/grimhold/internal/open5e"
)

type Spell struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Dice string `json:"dice"`
	Heal bool   `json:"heal"`
	Used bool   `json:"used"`
}

var diceRe = regexp.MustCompile(`\d+d\d+`)

func NewSpell(s open5e.Spell) (Spell, error) {
	dice := diceRe.FindString(s.Description)
	if dice == "" {
		return Spell{}, fmt.Errorf("spell %q has no dice", s.Slug)
	}
	return Spell{
		Slug: s.Slug,
		Name: s.Name,
		Dice: dice,
		Heal: strings.Contains(s.Description, "regain"),
		Used: false,
	}, nil
}
