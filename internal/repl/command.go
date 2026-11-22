package repl

import (
	"errors"
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/Roland-HUON/grimhold/internal/game"
	"github.com/Roland-HUON/grimhold/internal/open5e"
	"github.com/Roland-HUON/grimhold/internal/world"
)

var errExit = errors.New("exit")

type Config struct {
	World           world.World
	IDCurrentRegion string
	HasRested       bool
	Client          *open5e.Client
	Player          *game.Player
	CurrentFoe      *open5e.Monster
	CurrentFoeMaxHP int
	CanLoot         bool
	BestiaryPrevURL string
	BestiaryNextURL string
	Continue        bool
}

type cliCommand struct {
	name        string
	description string
	takesArgs   bool
	callback    func(cfg *Config, args []string) error
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Show this message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "List the known regions",
			callback:    commandMap,
		},
		"bestiary": {
			name:        "bestiary [next|prev]",
			description: "Browse the monsters of the realm",
			takesArgs:   true,
			callback:    commandBestiary,
		},
		"character": {
			name:        "character",
			description: "Show your character sheet",
			callback:    commandCharacter,
		},
		"inspect": {
			name:        "inspect <monster>",
			description: "Show the stat block of a monster",
			takesArgs:   true,
			callback:    commandInspect,
		},
		"travel": {
			name:        "travel <region>",
			description: "Travel to a connected region",
			takesArgs:   true,
			callback:    commandTravel,
		},
		"rest": {
			name:        "rest",
			description: "Restore HP and spells (once per region visit)",
			callback:    commandRest,
		},
		"equip": {
			name:        "equip <weapon>",
			description: "Wield a weapon from your inventory",
			takesArgs:   true,
			callback:    commandEquip,
		},
		"loot": {
			name:        "loot",
			description: "Search the body after a victory",
			callback:    commandLoot,
		},
		"cast": {
			name:        "cast <spell>",
			description: "Cast a spell (once per rest)",
			takesArgs:   true,
			callback:    commandCast,
		},
		"fight": {
			name:        "fight",
			description: "Play one round of combat",
			callback:    commandFight,
		},
		"flee": {
			name:        "flee",
			description: "Try to run away from the fight",
			callback:    commandFlee,
		},
		"look": {
			name:        "look",
			description: "Describe your surroundings",
			callback:    commandLook,
		},
		"explore": {
			name:        "explore",
			description: "Wander the region (risk an encounter)",
			callback:    commandExplore,
		},
		"weapons": {
			name:        "weapons",
			description: "List the weapons of the realm",
			callback:    commandWeapons,
		},
		"save": {
			name:        "save",
			description: "Save your progress to save.json",
			callback:    commandSave,
		},
		"load": {
			name:        "load",
			description: "Load your progress from save.json",
			callback:    commandLoad,
		},
		"exit": {
			name:        "exit",
			description: "Quit the game",
			callback:    commandExit,
		},
	}
}

func commandHelp(cfg *Config, args []string) error {
	fmt.Print("Grimhold — a dungeon crawler\n\nCommands:\n")
	cmds := getCommands()
	for _, name := range slices.Sorted(maps.Keys(cmds)) {
		fmt.Printf("  %-22s %s\n", cmds[name].name, cmds[name].description)
	}
	return nil
}

func commandMap(cfg *Config, args []string) error {
	regions := cfg.World.Regions

	for _, region := range regions {
		marker, here := "   ", ""
		if cfg.IDCurrentRegion == region.ID {
			marker, here = "-> ", " (you are here)"
		}
		fmt.Printf("%sName: %s%s\n   Danger level: %s\n   Exits: %s\n", marker, region.Name, here, colorize(fmt.Sprint(region.Danger), dangerColor(region.Danger)), strings.Join(region.Exits, ", "))
		if region.Boss != "" {
			if slices.Contains(cfg.Player.BossesSlain, region.Boss) {
				fmt.Println("   Boss lair: its master is dead")
			} else {
				fmt.Printf("   Boss lair: requires level %d\n", region.MinLevel)
			}
		}
	}
	return nil
}

func commandBestiary(cfg *Config, args []string) error {
	var pageURL string
	switch {
	case len(args) == 0:
	case len(args) == 1 && args[0] == "next":
		if cfg.BestiaryNextURL == "" {
			return errors.New("no next page")
		}
		pageURL = cfg.BestiaryNextURL
	case len(args) == 1 && args[0] == "prev":
		if cfg.BestiaryPrevURL == "" {
			return errors.New("no previous page")
		}
		pageURL = cfg.BestiaryPrevURL
	default:
		return errors.New("usage: bestiary [next|prev]")
	}

	page, err := cfg.Client.GetPage(pageURL)
	if err != nil {
		return fmt.Errorf("fetching bestiary page: %w", err)
	}

	cfg.BestiaryNextURL = page.NextURL
	cfg.BestiaryPrevURL = page.PrevURL

	fmt.Printf("Bestiary - %d monsters known\n\n", page.Count)
	for _, monster := range page.Monsters {
		fmt.Printf("  %-32s %s\n", monster.Slug, monster.Name)
	}

	var moreText string
	var s []string
	if cfg.BestiaryPrevURL != "" {
		s = append(s, "bestiary prev")
	}
	if cfg.BestiaryNextURL != "" {
		s = append(s, "bestiary next")
	}

	if len(s) > 0 {
		moreText = "More: " + strings.Join(s, " | ")
	}
	fmt.Printf("\nType 'inspect <slug>' for details. %s\n", moreText)
	return nil
}

func commandCharacter(cfg *Config, args []string) error {
	fmt.Printf("%-20s %d\n", "Level:", cfg.Player.Level)
	fmt.Printf("%-20s %s\n", "HP:", hpBar(cfg.Player.HP, cfg.Player.MaxHP, 20))
	fmt.Printf("%-20s %d/%d\n", "XP:", cfg.Player.XP, cfg.Player.NextLevelXP())
	fmt.Printf("%-20s %d\n", "Attack Bonus:", cfg.Player.AttackBonus)
	fmt.Printf("%-20s %d\n", "Armor Class:", cfg.Player.ArmorClass)
	fmt.Printf("Weapon:\n")
	fmt.Printf("- %-18s %s\n", "Name:", cfg.Player.Weapon.Name)
	fmt.Printf("- %-18s %s\n", "Damage Dice:", cfg.Player.Weapon.DamageDice)
	fmt.Printf("- %-18s %d\n", "Damage Bonus:", cfg.Player.Weapon.DamageBonus)
	fmt.Printf("%-20s %d\n", "Gold:", cfg.Player.Gold)
	if len(cfg.Player.Inventory) == 0 {
		fmt.Printf("Inventory: empty\n")
	} else {
		fmt.Printf("Inventory:\n")
		for _, weapon := range cfg.Player.Inventory {
			label := fmt.Sprintf("%s (%s+%d)", weapon.Name, weapon.DamageDice, weapon.DamageBonus)
			fmt.Printf("- %-25s -> equip %s\n", label, weapon.Slug)
		}
	}
	if len(cfg.Player.Spells) == 0 {
		fmt.Printf("Spells: none\n")
	} else {
		fmt.Printf("Spells:\n")
		for _, spell := range cfg.Player.Spells {
			kind := "damage"
			if spell.Heal {
				kind = "heal"
			}
			state := "ready"
			if spell.Used {
				state = "spent"
			}
			label := fmt.Sprintf("%s (%s, %s)", spell.Name, spell.Dice, kind)
			fmt.Printf("- %-28s %-6s -> cast %s\n", label, state, spell.Slug)
		}
	}
	if len(cfg.Player.BestiaryKilled) == 0 {
		fmt.Printf("Bestiary killed: None yet\n")
	} else {
		fmt.Printf("Bestiary killed:\n")
		for _, monster := range slices.Sorted(maps.Keys(cfg.Player.BestiaryKilled)) {
			fmt.Printf("- %s x %d\n", monster, cfg.Player.BestiaryKilled[monster])
		}
	}
	return nil
}

func commandInspect(cfg *Config, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: inspect <monster>")
	}

	monsterName := toSlug(args)
	monster, err := cfg.Client.GetMonster(monsterName)
	if errors.Is(err, open5e.ErrNotFound) {
		return fmt.Errorf("no monster called %q, try 'bestiary' to find its slug", monsterName)
	}
	if err != nil {
		return fmt.Errorf("inspecting %q: %w", monsterName, err)
	}

	fmt.Printf("Name: %s\nHP: %d\nArmor Class: %d\nChallenge Rating: %s\nActions:\n", monster.Name, monster.HP, monster.ArmorClass, monster.ChallengeRating)
	for _, action := range monster.Actions {
		fmt.Printf("  - %s: %s\n", action.Name, action.Desc)
	}
	return nil
}

func commandTravel(cfg *Config, args []string) error {
	if cfg.CurrentFoe != nil {
		return fmt.Errorf("you can't leave while fighting %s, fight or flee", foeName(cfg))
	}
	if len(args) == 0 {
		return errors.New("usage: travel <region>")
	}

	destID := toSlug(args)
	region, ok := world.FindRegionByID(cfg.World, destID)
	if !ok {
		return fmt.Errorf("unknown region %q, type 'map' to see them", destID)
	}

	if destID == cfg.IDCurrentRegion {
		return errors.New("you are already in this region")
	}

	currentRegion, ok := world.FindRegionByID(cfg.World, cfg.IDCurrentRegion)
	if !ok {
		return fmt.Errorf("current region %q not found", cfg.IDCurrentRegion)
	}
	if !slices.Contains(currentRegion.Exits, destID) {
		return fmt.Errorf("you can't reach %s from %s", destID, currentRegion.Name)
	}
	if cfg.Player.Level < region.MinLevel {
		return fmt.Errorf("%s is sealed: you must be level %d to enter (you are level %d)", region.Name, region.MinLevel, cfg.Player.Level)
	}
	cfg.IDCurrentRegion = region.ID
	cfg.HasRested = false
	cfg.CanLoot = false
	fmt.Printf("You arrive at %s\n", region.Name)
	warnPrefetch(prefetchFoes(cfg.Client, region))
	return nil
}

func commandRest(cfg *Config, args []string) error {
	if cfg.CurrentFoe != nil {
		return fmt.Errorf("you can't rest while fighting %s, fight or flee", foeName(cfg))
	}
	if cfg.HasRested {
		return errors.New("you have already rested in this region")
	}
	cfg.Player.HP = cfg.Player.MaxHP
	for i := range cfg.Player.Spells {
		cfg.Player.Spells[i].Used = false
	}
	cfg.HasRested = true
	fmt.Printf("You rest by the fire. HP: %d/%d\n", cfg.Player.HP, cfg.Player.MaxHP)
	if len(cfg.Player.Spells) > 0 {
		fmt.Println("Your spells are ready again.")
	}
	return nil
}

func commandLook(cfg *Config, args []string) error {
	var dangerWarning string
	region, ok := world.FindRegionByID(cfg.World, cfg.IDCurrentRegion)
	if !ok {
		return fmt.Errorf("current region %q not found", cfg.IDCurrentRegion)
	}

	switch {
	case slices.Contains(cfg.Player.BossesSlain, region.Boss):
		dangerWarning = "The air is still, its master is dead"
	case region.Danger == 1:
		dangerWarning = "The air feels calm"
	case region.Danger == 2:
		dangerWarning = "Something is watching you"
	case region.Danger == 3:
		dangerWarning = "The smell of blood hangs in the air"
	case region.Danger == 4:
		dangerWarning = "Death lingers here"
	default:
		dangerWarning = "This place belongs to the void. No living being has ever come back from it"
	}
	fmt.Printf("Name: %s\nDescription: %s\nExits: %s\nDanger: %s\n", region.Name, region.Description, strings.Join(region.Exits, ", "), colorize(dangerWarning, dangerColor(region.Danger)))
	return nil
}

func commandExplore(cfg *Config, args []string) error {
	if cfg.CurrentFoe != nil {
		return fmt.Errorf("you are already fighting %s", foeName(cfg))
	}
	region, ok := world.FindRegionByID(cfg.World, cfg.IDCurrentRegion)
	if !ok {
		return fmt.Errorf("current region %q not found", cfg.IDCurrentRegion)
	}
	if region.Boss != "" {
		return encounterBoss(cfg, region)
	}
	if len(region.Foes) == 0 {
		return errors.New("nothing lurks here")
	}
	if err := spawnFoe(cfg, region); err != nil {
		return err
	}
	fmt.Printf("%s emerges from the shadows! (HP %d, Armor class %d)\nType 'fight' to attack or 'flee' to run.\n", capitalize(withArticle(cfg.CurrentFoe.Name)), cfg.CurrentFoe.HP, cfg.CurrentFoe.ArmorClass)
	return nil
}

func spawnFoe(cfg *Config, region world.Region) error {
	rolls := region.Danger
	if rolls < 1 {
		rolls = 1
	}
	var worstMonsterFind *open5e.Monster
	for i := 0; i < rolls; i++ {
		monsterIndex := rand.IntN(len(region.Foes))
		monsterSlug := region.Foes[monsterIndex]
		monster, err := cfg.Client.GetMonster(monsterSlug)
		if err != nil {
			return fmt.Errorf("exploring: %w", err)
		}
		if worstMonsterFind == nil || monster.CR > worstMonsterFind.CR {
			worstMonsterFind = &monster
		}
	}
	cfg.CanLoot = false
	cfg.CurrentFoe = worstMonsterFind
	cfg.CurrentFoeMaxHP = cfg.CurrentFoe.HP
	return nil
}

func commandEquip(cfg *Config, args []string) error {
	if cfg.CurrentFoe != nil {
		return fmt.Errorf("you can't equip while fighting %s, fight or flee", foeName(cfg))
	}
	if len(args) == 0 {
		return errors.New("usage: equip <weapon>")
	}

	p := cfg.Player
	slug := toSlug(args)
	if p.Weapon.Slug == slug {
		return fmt.Errorf("you already wield your %s", p.Weapon.Name)
	}

	for i, weapon := range p.Inventory {
		if weapon.Slug == slug {
			p.Weapon, p.Inventory[i] = p.Inventory[i], p.Weapon
			fmt.Printf("You now wield %s (%s).\n", withArticle(p.Weapon.Name), p.Weapon.DamageDice)
			return nil
		}
	}
	return fmt.Errorf("you don't have %q in your bag", slug)
}

func commandLoot(cfg *Config, args []string) error {
	if !cfg.CanLoot {
		return errors.New("nothing to loot: win a fight first, each body can only be searched once")
	}
	cfg.CanLoot = false
	if rand.IntN(2) == 0 {
		fmt.Printf("You search the body... nothing useful.\n")
		return nil
	}

	weapons, err := cfg.Client.GetWeapons()
	if err != nil {
		return fmt.Errorf("looting: %w", err)
	}
	if len(weapons) == 0 {
		return errors.New("no weapon to find")
	}

	found := weapons[rand.IntN(len(weapons))]
	loot := game.Weapon{
		Slug:        found.Slug,
		Name:        found.Name,
		DamageDice:  found.DamageDice,
		DamageBonus: 0,
	}
	cfg.Player.Inventory = append(cfg.Player.Inventory, loot)
	fmt.Printf("You found %s (%s)! Type 'equip %s' to wield it.\n", withArticle(loot.Name), loot.DamageDice, loot.Slug)
	return nil
}

func commandFight(cfg *Config, args []string) error {
	if cfg.CurrentFoe == nil {
		return errors.New("nothing to fight, try explore")
	}

	playerDamage, err := game.PlayerAttack(cfg.Player, cfg.CurrentFoe)
	if err != nil {
		return fmt.Errorf("player attack: %w", err)
	}
	if playerDamage.Hit {
		fmt.Printf("You swing your %s (%d vs %d): hit for %d damage.\n", cfg.Player.Weapon.Name, playerDamage.Roll, cfg.CurrentFoe.ArmorClass, playerDamage.Damage)
	} else {
		fmt.Printf("You swing your %s (%d vs %d): miss.\n", cfg.Player.Weapon.Name, playerDamage.Roll, cfg.CurrentFoe.ArmorClass)
	}
	if cfg.CurrentFoe.HP <= 0 {
		winFight(cfg)
		return nil
	}
	if err := monsterTurn(cfg); err != nil {
		return err
	}
	if cfg.CurrentFoe == nil {
		return nil
	}
	fmt.Printf("You: %s | %s: %s\n", hpBar(cfg.Player.HP, cfg.Player.MaxHP, 10), cfg.CurrentFoe.Name, hpBar(cfg.CurrentFoe.HP, cfg.CurrentFoeMaxHP, 10))
	return nil
}

func winFight(cfg *Config) {
	region, isBoss := currentBossRegion(cfg)

	reward := cfg.Player.Win(cfg.CurrentFoe)
	fmt.Printf("%s falls!\n", capitalize(foeName(cfg)))
	fmt.Printf("You gained %d XP and %d gold.\n", reward.XP, reward.Gold)
	if reward.LevelsGained >= 1 {
		fmt.Printf("Level up! You are now level %d.\n", cfg.Player.Level)
	}
	if isBoss {
		cfg.Player.BossesSlain = append(cfg.Player.BossesSlain, region.Boss)
		banner := fmt.Sprintf("*** %s is no more. %s falls silent forever. ***", cfg.CurrentFoe.Name, region.Name)
		fmt.Println(colorize(banner, colorBold+colorMagenta))
	}
	cfg.CanLoot = true
	cfg.CurrentFoe = nil
}

func monsterTurn(cfg *Config) error {
	monsterDamage, err := game.MonsterAttack(cfg.CurrentFoe, cfg.Player)
	if err != nil {
		return fmt.Errorf("monster attack: %w", err)
	}
	if monsterDamage.ActionName == "" {
		fmt.Printf("%s hesitates and does nothing.\n", capitalize(foeName(cfg)))
	} else if monsterDamage.Hit {
		fmt.Printf("%s uses %s (%d vs %d): hit for %d damage.\n", cfg.CurrentFoe.Name, monsterDamage.ActionName, monsterDamage.Roll, cfg.Player.ArmorClass, monsterDamage.Damage)
	} else {
		fmt.Printf("%s uses %s (%d vs %d): miss.\n", cfg.CurrentFoe.Name, monsterDamage.ActionName, monsterDamage.Roll, cfg.Player.ArmorClass)
	}
	if cfg.Player.HP <= 0 {
		cfg.Player.Defeat()
		cfg.CanLoot = false
		cfg.CurrentFoe = nil
		cfg.IDCurrentRegion = cfg.World.Start
		cfg.HasRested = false
		region, ok := world.FindRegionByID(cfg.World, cfg.IDCurrentRegion)
		if !ok {
			return errors.New("not found the starting region")
		}
		fmt.Printf("You have been slain...\nYou wake up at %s with 1 HP, your pride wounded.\n", region.Name)
		warnPrefetch(prefetchFoes(cfg.Client, region))
	}
	return nil
}

func commandCast(cfg *Config, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: cast <spell>")
	}
	slug := toSlug(args)
	index := slices.IndexFunc(cfg.Player.Spells, func(s game.Spell) bool {
		return s.Slug == slug
	})
	if index == -1 {
		return fmt.Errorf("you don't know %q", slug)
	}

	spell := &cfg.Player.Spells[index]
	if spell.Used {
		return fmt.Errorf("%s is spent, rest to recover it", spell.Name)
	}
	if !spell.Heal && cfg.CurrentFoe == nil {
		return errors.New("no target, try explore")
	}
	if spell.Heal && cfg.Player.HP >= cfg.Player.MaxHP {
		return errors.New("you are already at full health")
	}
	amount, err := game.Roll(spell.Dice)
	if err != nil {
		return fmt.Errorf("casting %s: %w", spell.Name, err)
	}
	spell.Used = true
	if spell.Heal {
		before := cfg.Player.HP
		cfg.Player.HP = min(cfg.Player.HP+amount, cfg.Player.MaxHP)
		fmt.Printf("You cast %s and recover %d HP. %s\n", spell.Name, cfg.Player.HP-before, hpBar(cfg.Player.HP, cfg.Player.MaxHP, 10))
	} else {
		cfg.CurrentFoe.HP -= amount
		fmt.Printf("You cast %s: %s takes %d damage.\n", spell.Name, foeName(cfg), amount)
	}

	if cfg.CurrentFoe == nil {
		return nil
	}
	if cfg.CurrentFoe.HP <= 0 {
		winFight(cfg)
		return nil
	}
	if err := monsterTurn(cfg); err != nil {
		return err
	}
	if cfg.CurrentFoe == nil {
		return nil
	}
	fmt.Printf("You: %s | %s: %s\n", hpBar(cfg.Player.HP, cfg.Player.MaxHP, 10), cfg.CurrentFoe.Name, hpBar(cfg.CurrentFoe.HP, cfg.CurrentFoeMaxHP, 10))
	return nil
}

func commandFlee(cfg *Config, args []string) error {
	if cfg.CurrentFoe == nil {
		return errors.New("nothing to flee from")
	}
	jet, success := game.TryFlee()
	if success {
		fmt.Printf("You escape from %s (%d vs %d)!\n", foeName(cfg), jet, game.FleeDC)
		cfg.CurrentFoe = nil
		return nil
	}
	fmt.Printf("You fail to escape (%d vs %d)!\n", jet, game.FleeDC)
	if err := monsterTurn(cfg); err != nil {
		return err
	}
	if cfg.CurrentFoe == nil {
		return nil
	}
	fmt.Printf("You: %s | %s: %s\n", hpBar(cfg.Player.HP, cfg.Player.MaxHP, 10), cfg.CurrentFoe.Name, hpBar(cfg.CurrentFoe.HP, cfg.CurrentFoeMaxHP, 10))
	return nil
}

func commandWeapons(cfg *Config, args []string) error {
	weapons, err := cfg.Client.GetWeapons()
	if err != nil {
		return fmt.Errorf("listing weapons: %w", err)
	}
	fmt.Printf("Armory - %d weapons\n", len(weapons))
	for _, weapon := range weapons {
		fmt.Printf("  %-28s %-28s %s\n", weapon.Slug, weapon.Name, weapon.DamageDice)
	}
	return nil
}

func commandExit(cfg *Config, args []string) error { return errExit }
