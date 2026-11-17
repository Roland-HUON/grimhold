# Final assignment — **Grimhold**

You will build **Grimhold**: a text-based dungeon crawler that runs in your
terminal. You explore cursed regions, stumble into random encounters, fight
foes in dice-based combat, and record your kills in a bestiary. Monster data
comes live from the [Open5e API](https://open5e.com) (D&D 5e SRD content, free,
no API key); the world map is a JSON file provided with this assignment.

Everything you need has been covered in the course: structs, interfaces, maps,
error handling, packages, `encoding/json`, HTTP, goroutines, mutexes, testing.

---

## 1. The REPL

The program is a loop: print a prompt, read a line, parse it into a command
and arguments, dispatch, repeat.

```text
grimhold > help
Grimhold — a dungeon crawler

Commands:
  help                     Show this message
  map                      List the known regions
  travel <region>          Travel to a connected region
  look                     Describe your surroundings
  explore                  Wander the region (risk an encounter)
  ...
grimhold > travel ashen-marsh
You trudge into the Ashen Marsh. The fog smells of iron.
```

**Requirements**

- Commands are registered in a `map[string]cliCommand`, where each entry is a
  struct `{ name, description string; callback func(...) error }`. `help` is
  generated from this map — never hand-maintained.
- Unknown commands, wrong argument counts and empty input are handled
  gracefully. The REPL never crashes on user input.

## 2. The world

A `world.json` file (starter version provided) describes 5–6 regions:

```json
{
  "start": "crossroads",
  "regions": [
    {
      "id": "ashen-marsh",
      "name": "Ashen Marsh",
      "description": "Fog, iron, and things that breathe under the water.",
      "danger": 1,
      "foes": ["goblin", "skeleton", "ghoul"],
      "exits": ["crossroads", "drowned-chapel"]
    }
  ]
}
```

- `map` lists all regions, their danger level, and their exits.
- `travel <region>` moves the player — **only** along an exit of the current
  region. Invalid travels are refused with a message.
- `look` prints the current region's description, exits, and a hint of its
  danger.
- You may extend `world.json` (more regions, grimmer prose) as long as the
  format stays compatible.

## 3. The Open5e client

All monster data comes from `https://api.open5e.com/v1/monsters/` — decoded
into typed structs (name, hit points, armor class, challenge rating, actions…).

- `bestiary` lists monsters page by page (`next` / `prev` follow the API's
  pagination URLs).
- `inspect <monster>` prints a readable stat block.
- **Caching**: API responses are cached in memory with a TTL
  (e.g. 5 minutes). The cache is a `map` guarded by a `sync.Mutex`, with a
  reaper **goroutine** that periodically purges expired entries.

## 4. Encounters and combat

- `explore` rolls a random foe from the current region's `foes` list (higher
  `danger` regions may roll extra times and keep the worst). The foe's stats
  are fetched from the API (through your cache).
- `fight` resolves combat in rounds:
  - attack roll: `d20 + attack bonus` vs the foe's armor class
  - damage: roll the weapon/attack dice on a hit
  - the foe strikes back the same way
  - repeat until someone reaches 0 HP — or the player `flee`s (one DEX-style
    roll, failure means a free hit)
- Victory grants XP based on the foe's challenge rating, an entry in your
  **slain bestiary** (with kill counts), and maybe a few gold pieces.
- Defeat is not the end: you wake up at the crossroads with 1 HP and your
  pride wounded. Death mid-`explore` must never corrupt the save.

## 5. The character

- The player has HP, max HP, level, XP, attack bonus, and a weapon.
- `character` prints the character sheet; `rest` restores HP (full heal, but
  you can only rest once per region visit).
- Level-up threshold is your choice (e.g. `level² × 50` XP); leveling raises
  max HP and attack bonus.

## 6. Save / load

- `save` writes the full game state (location, character, slain bestiary,
  gold, rest state) to `save.json`. `load` (or a `--continue` flag) restores
  it. If no save exists, say so — no crash.

## 7. Tests (optional)

Testing was not covered during the course, so tests are **not required** —
but they earn bonus points (see the bonus menu below). A few reference tests
are provided in `test-examples/` if you want to learn the pattern; a small
suite is also the easiest way to prove your combat and cache logic works
without playing through the game by hand.

## 8. Engineering rules

- Module named `github.com/<you>/grimhold`, Go 1.22+, **standard library only**
  for the core.
- Code split into packages, e.g. `cmd/grimhold/main.go` (thin) +
  `internal/repl`, `internal/game`, `internal/open5e`, `internal/world`.
- Errors are handled and wrapped (`fmt.Errorf("...: %w", err)`), never
  silently dropped.
- `gofmt`-clean and `go vet`-clean.
- A short `README.md`: what it is, how to build, run, save, and a sample
  session transcript.

---

## Bonus features (choose any, each documented in the README)

| Bonus | Pts | What it teaches |
|---|---|---|
| **Ambush ticker** — while exploring, a goroutine ticks every N seconds; a random ambush may interrupt the prompt (`select` on a channel) | +6 | goroutines, channels, `select` |
| **Loot & equipment** — `weapons` from `/v1/weapons/`, `loot`, `equip <weapon>`; damage uses the equipped weapon's dice | +4 | more JSON modeling, game state |
| **Boss region** — one region gated behind a level requirement; a unique boss with multiple actions | +3 | design, balancing |
| **Spells** — fetch `/v1/spells/`, pick 2 at character creation, `cast <spell>` once per rest | +4 | richer structs, resource management |
| **Faster dungeon** — prefetch the current region's foes concurrently when you `travel` (WaitGroup or errgroup-style fan-in) | +4 | concurrency patterns |
| **Test suite** — at least 3 meaningful tests (e.g. combat with a seeded `rand`, cache expiry with an injected clock, Open5e decoding against an `httptest.Server`); tests must not hit the network | +6 | `testing`, table-driven tests, `httptest` |
| **ANSI colors** — hand-rolled escape codes for HP bars and danger levels | +2 | polish |

Bonuses are capped at **+12** total.

---

## Grading sketch (/100)

| Criterion | Pts |
|---|---|
| REPL & commands work, robust to any input | 30 |
| Open5e client: typed decoding, pagination, TTL cache with mutex + reaper | 25 |
| Combat & game state (explore/fight/flee/rest, XP, death flow) | 25 |
| Idiomatic Go: package layout, error wrapping, `gofmt`/`go vet` | 15 |
| README with build instructions and a sample session | 5 |
| **Bonuses** | **up to +12** |
