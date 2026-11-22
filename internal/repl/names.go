package repl

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Roland-HUON/grimhold/internal/world"
)

func withArticle(name string) string {
	if name == "" {
		return name
	}
	first, _ := utf8.DecodeRuneInString(name)
	if strings.ContainsRune("AEIOUaeiou", first) {
		return "an " + name
	}
	return "a " + name
}

func capitalize(text string) string {
	first, size := utf8.DecodeRuneInString(text)
	if size == 0 {
		return text
	}
	return string(unicode.ToUpper(first)) + text[size:]
}

func currentBossRegion(cfg *Config) (world.Region, bool) {
	region, ok := world.FindRegionByID(cfg.World, cfg.IDCurrentRegion)
	isBoss := ok && cfg.CurrentFoe != nil && region.Boss != "" && region.Boss == cfg.CurrentFoe.Slug
	return region, isBoss
}

func foeName(cfg *Config) string {
	if cfg.CurrentFoe == nil {
		return ""
	}
	if _, isBoss := currentBossRegion(cfg); isBoss {
		return cfg.CurrentFoe.Name
	}
	return "the " + cfg.CurrentFoe.Name
}
