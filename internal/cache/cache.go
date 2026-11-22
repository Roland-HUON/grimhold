package cache

import (
	"sync"
	"time"
)

const defaultTTL = 5 * time.Minute

type cacheEntry struct {
	val       []byte
	createdAt time.Time
}

type Cache struct {
	mu      sync.Mutex
	entries map[string]cacheEntry
	ttl     time.Duration
	now     func() time.Time
	done    chan struct{}
	once    sync.Once
}

func NewCache(ttl time.Duration) *Cache {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	c := &Cache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
		now:     time.Now,
		done:    make(chan struct{}),
	}
	go c.reapLoop()
	return c
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{val: val, createdAt: c.now()}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if c.now().Sub(entry.createdAt) > c.ttl {
		delete(c.entries, key)
		return nil, false
	}
	return entry.val, true
}

func (c *Cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

func (c *Cache) Stop() {
	c.once.Do(func() { close(c.done) })
}

func (c *Cache) reap() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, entry := range c.entries {
		if c.now().Sub(entry.createdAt) > c.ttl {
			delete(c.entries, key)
		}
	}
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.reap()
		case <-c.done:
			return
		}
	}
}
