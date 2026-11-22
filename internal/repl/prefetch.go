package repl

import (
	"fmt"
	"slices"
	"sync"

	"github.com/Roland-HUON/grimhold/internal/open5e"
	"github.com/Roland-HUON/grimhold/internal/world"
)

func prefetchFoes(client *open5e.Client, region world.Region) []error {
	slugs := region.Foes
	if region.Boss != "" {
		slugs = append(slices.Clone(slugs), region.Boss)
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(slugs))

	for _, slug := range slugs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := client.GetMonster(slug); err != nil {
				errs <- fmt.Errorf("prefetching %q: %w", slug, err)
			}
		}()
	}

	wg.Wait()
	close(errs)

	var all []error
	for err := range errs {
		all = append(all, err)
	}
	return all
}

func warnPrefetch(errs []error) {
	for _, err := range errs {
		fmt.Println("Warning:", err)
	}
}
