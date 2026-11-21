package open5e

type Action struct {
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	AttackBonus int    `json:"attack_bonus"`
	DamageDice  string `json:"damage_dice"`
	DamageBonus int    `json:"damage_bonus"`
}

type Monster struct {
	Slug            string   `json:"slug"`
	Name            string   `json:"name"`
	ArmorClass      int      `json:"armor_class"`
	HP              int      `json:"hit_points"`
	ChallengeRating string   `json:"challenge_rating"`
	CR              float64  `json:"cr"`
	Actions         []Action `json:"actions"`
}

type MonstersPage struct {
	Count    int       `json:"count"`
	NextURL  string    `json:"next"`
	PrevURL  string    `json:"previous"`
	Monsters []Monster `json:"results"`
}
