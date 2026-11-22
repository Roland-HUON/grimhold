package game

type Weapon struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	DamageDice  string `json:"damage_dice"`
	DamageBonus int    `json:"damage_bonus"`
}

type Player struct {
	HP             int            `json:"hp"`
	MaxHP          int            `json:"max_hp"`
	XP             int            `json:"xp"`
	Level          int            `json:"level"`
	AttackBonus    int            `json:"attack_bonus"`
	ArmorClass     int            `json:"armor_class"`
	Weapon         Weapon         `json:"weapon"`
	Inventory      []Weapon       `json:"inventory"`
	Spells         []Spell        `json:"spells"`
	Gold           int            `json:"gold"`
	BestiaryKilled map[string]int `json:"bestiary_killed"`
	BossesSlain    []string       `json:"bosses_slain"`
}

func NewPlayer() *Player {
	return &Player{
		HP:          20,
		MaxHP:       20,
		XP:          0,
		Level:       1,
		AttackBonus: 3,
		ArmorClass:  12,
		Weapon: Weapon{
			Slug:        "longsword",
			Name:        "Longsword",
			DamageDice:  "1d8",
			DamageBonus: 2,
		},
		Gold:           0,
		BestiaryKilled: make(map[string]int),
	}
}

func (p *Player) NextLevelXP() int {
	return p.Level * p.Level * 50
}
