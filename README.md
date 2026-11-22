# Grimhold

Grimhold is a text-based dungeon crawler that runs in your terminal. You travel
through cursed regions, stumble into random encounters, fight monsters in
dice-based combat and record your kills in a bestiary. Monster, weapon and
spell data come live from the [Open5e API](https://open5e.com) (D&D 5e SRD),
and the world map is described in `internal/world/world.json` (embedded in the
binary).

Written in Go, standard library only.

## Build and run

Requires Go 1.27.4 or newer (the version in `go.mod`) and an internet connection
(Open5e). With an older Go (1.21+), the `go` command downloads the right
toolchain automatically. It works on Linux, macOS and Windows.

The simplest way, the same on every system:

```sh
go run ./cmd/grimhold
go run ./cmd/grimhold --continue    # continue a saved game
```

Or build a binary.

Linux / macOS:

```sh
go build -o grimhold ./cmd/grimhold
./grimhold
```

Windows (PowerShell or cmd):

```powershell
go build -o grimhold.exe ./cmd/grimhold
.\grimhold.exe
```

The game can be started from any directory. `save.json` is written to and read
from the current directory.

Colors are used when the output is a terminal that supports them (enabled
automatically on Windows 10+ consoles). They are turned off when the output is
redirected, or with the `NO_COLOR` environment variable.

## How to play

At start you pick 2 spells (or type `load` to continue your saved game, `exit`
to quit). Then the `grimhold >` prompt accepts these commands. `help` prints the
same list, generated from the command map.

| Command | Effect |
|---|---|
| `help` | Show the commands |
| `map` | List the regions, their danger and exits |
| `look` | Describe the current region |
| `travel <region>` | Move along an exit of the current region |
| `explore` | Risk an encounter with a foe of the region |
| `fight` | Fight in rounds (you strike, then the foe strikes back) until one of you falls |
| `flee` | Try to run away (d20 vs 10, failure gives the foe a free hit) |
| `cast <spell>` | Cast a spell (once per rest) |
| `rest` | Full heal and spells recharged, once per region visit |
| `character` | Character sheet: HP, XP, weapon, inventory, spells, bestiary |
| `bestiary [next\|prev]` | Browse the Open5e monsters page by page |
| `inspect <monster>` | Show a monster's stat block |
| `loot` | Search the body after a victory (may find a weapon) |
| `weapons` | List the weapons of the realm |
| `equip <weapon>` | Wield a weapon from your inventory |
| `save` / `load` | Save or restore the game |
| `exit` | Quit |

Commands are case-insensitive, and multi-word names are turned into slugs:
`cast Fire Bolt` is the same as `cast fire-bolt`, `travel Ashen Marsh` the same
as `travel ashen-marsh`.

**Combat**: `fight` plays rounds until you or your foe reaches 0 HP (it stops
after 50 rounds if nobody can hit, so you can `flee`). Before fighting, you can
also `cast` a spell or `flee`. The attack roll is `d20 + attack bonus` against the foe's armor class,
and a hit deals the weapon dice. Higher danger regions roll several foes and
keep the worst one. Victory gives XP based on the challenge rating (minimum 10),
gold and a bestiary entry. The next level needs `level² × 50` XP and gives
+10 max HP and +1 attack bonus. Defeat sends you back to the Crossroads with
1 HP.

**Save / load**: `save` writes the location, the character (HP, XP, level,
weapon, inventory, spells, gold, slain bestiary, bosses slain) and the rest
state to `save.json`. It writes a temporary file first, then renames it, so an
interrupted save never leaves a broken file. `load` restores it, or says so if
no save exists. A save that was edited by hand is checked before being used
(level, HP, XP, gold, dice, region and its level requirement) and refused if
invalid. The same check runs before writing, so `save` never replaces a good
save with a broken one.
Both commands are refused during a fight, so a death in combat can never
corrupt the save.

**Cache**: every Open5e response is cached in memory for 5 minutes. The cache is
a map guarded by a `sync.Mutex`, and a reaper goroutine purges expired entries.
The reaper stops when the client is closed.

**Robustness**: unknown commands, missing or extra arguments, empty lines and
lines that are too long (more than 4096 bytes) are reported without leaving the
game.

## Sample session

```text
Choose 2 spells (or type 'load' to continue your saved game, 'exit' to quit):
  - fire-bolt
  - ray-of-frost
  - thunderwave
  - burning-hands
  - guiding-bolt
  - cure-wounds
  - healing-word
spell 1/2 > fire-bolt
You learned Fire Bolt (1d10).
spell 2/2 > cure-wounds
You learned Cure Wounds (1d8).
grimhold > look
Name: The Crossroads
Description: A leaning gallows marks where four mud roads meet. Crows watch from the beam, patient as debt collectors.
Exits: ashen-marsh, whispering-woods, bone-quarry
Danger: The air feels calm
grimhold > travel ashen-marsh
You arrive at Ashen Marsh
grimhold > explore
A Ghoul emerges from the shadows! (HP 22, Armor class 12)
Type 'fight' to attack or 'flee' to run.
grimhold > cast fire bolt
You cast Fire Bolt: the Ghoul takes 5 damage.
Ghoul uses Bite (17 vs 12): hit for 11 damage.
You: [████░░░░░░] 9/20 HP | Ghoul: [███████░░░] 17/22 HP
grimhold > fight
You swing your Longsword (13 vs 12): hit for 10 damage.
Ghoul uses Bite (21 vs 12): hit for 7 damage.
You: [█░░░░░░░░░] 2/20 HP | Ghoul: [███░░░░░░░] 7/22 HP
You swing your Longsword (18 vs 12): hit for 10 damage.
The Ghoul falls!
You gained 100 XP and 15 gold.
Level up! You are now level 2.
grimhold > loot
You search the body... nothing useful.
grimhold > rest
You rest by the fire. HP: 30/30
Your spells are ready again.
grimhold > save
Game saved.
grimhold > exit
Bye!
```

An ambush, while idle in a region:

```text
grimhold >
Ambush! A Giant Rat leaps out of the shadows!
Giant Rat uses Bite (18 vs 12): hit for 6 damage.
You: [███████░░░] 14/20 HP | Giant Rat: [██████████] 7/7 HP
Type 'fight' to attack or 'flee' to run.
grimhold >
```

## Bonus features

### Ambush ticker

The REPL no longer blocks on reading stdin. A goroutine reads the lines and
sends them on a channel (closed at end of input), and the main loop `select`s
between that channel and a `time.Ticker`. When no command has been typed for
20 seconds, there is a 35% chance that a foe of the region ambushes you and
strikes first. There is no ambush during a fight, in a boss region, or when the
player is at 25% HP or less (so a freshly revived player is not killed again).

Only the main goroutine touches the game state: the reader goroutine only reads
lines, so there is no data race and no mutex is needed.
Code: `internal/repl/repl.go`, `internal/repl/ambush.go`.

### Loot and equipment

`weapons` lists the weapons from `/v1/weapons/` (those with damage dice).
After a victory, `loot` has a 1 in 2 chance to find a random weapon, which goes
to the inventory. `equip <weapon>` swaps it with the wielded one, and combat
damage uses the equipped weapon's dice.

### Boss region

The **Manticore Spire** is reachable from the Cinder Crypt but sealed until
level 6 (also checked when loading a save). Exploring it always summons its
unique boss, *Vexhara, Queen of the Spire*: an Open5e manticore whose actions
(Bite, Claw, Tail Spike) are listed in its banner, one of them being used each
turn. Once slain, the boss is remembered in the save, never comes back, and
`map` shows that its lair is empty.

### Spells

At start the player picks 2 spells among 7 starter spells, fetched from
`/v1/spells/`. The dice are extracted from the spell description, and a spell
whose description mentions "regain" heals. `cast <spell>` deals damage to the
current foe (who then strikes back) or heals the player, capped at max HP. Each
spell can be cast once per `rest`.

### Faster dungeon

When you start, `travel`, `load` or die, every foe of the region (and its boss)
is fetched concurrently: one goroutine per monster, a `sync.WaitGroup` for the
fan-in and a buffered error channel. The next `explore` is then served from the
cache. Failures are only printed as warnings.
Code: `internal/repl/prefetch.go`.

### Test suite

```sh
go test ./...
```

No test hits the network.

| File | What it checks |
|---|---|
| `internal/game/dice_test.go` | `Roll` always stays in range (1000 rolls per formula), invalid or too big formulas return an error |
| `internal/game/spell_test.go` | `NewSpell` extracts the dice, detects healing spells, rejects spells without dice |
| `internal/game/fight_test.go` | Attacks always hit AC 0 and miss AC 100, monsters without usable damage actions do nothing, `Win` gives the right XP and level-ups |
| `internal/game/validate_test.go` | Broken characters (level not matching XP, HP, armor class, gold, too strong dice, overflowing values) are rejected, levels stop at the maximum |
| `internal/cache/cache_test.go` | Cache hit, miss and expiry with an injected clock (before, at and after the TTL), the reaper removes only expired entries |
| `internal/open5e/client_test.go` | Decoding of a monsters page against an `httptest.Server`, second call served from the cache, 404 (`ErrNotFound`) / 500 / invalid JSON errors |
| `internal/world/world_test.go` | The embedded `world.json` is valid, broken files (bad JSON, unknown start or exit, missing file) are rejected |
| `internal/repl/repl_test.go` | Too long input lines are skipped without losing the next ones, commands are case-insensitive, invalid commands never change the state, save round trip, rejection of broken saves, `save` refusing an invalid game, `fight` going on until someone falls (and stopping if nobody can hit), ambush guards, articles and slugs |

All tests are table-driven where it makes sense.

### ANSI colors

Hand-rolled escape codes (no library): HP bars turn green, yellow then red, the danger of a
region is colored in `map` and `look`, and boss banners and ambushes stand out.
Code: `internal/repl/color.go`.

## Project layout

```text
cmd/grimhold/        main: flags, loads the world, creates the client and starts the REPL
internal/repl/       REPL loop, commands, save/load, ambush, boss, prefetch, colors
internal/game/       player, dice, combat, spells, validation
internal/open5e/     Open5e HTTP client and types
internal/cache/      TTL cache with a reaper goroutine
internal/world/      world.json (embedded) loading and validation
```
