package game

type Weapon struct {
	Name        string
	DamageDice  string
	DamageBonus int
}

type Player struct {
	HP             int
	MaxHP          int
	XP             int
	Level          int
	AttackBonus    int
	ArmorClass     int
	Weapon         Weapon
	Gold           int
	BestiaryKilled map[string]int
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
			Name:        "Longsword",
			DamageDice:  "1d8",
			DamageBonus: 2,
		},
		Gold:           0,
		BestiaryKilled: make(map[string]int),
	}
}
