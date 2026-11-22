package cache

import (
	"testing"
	"time"
)

type fakeClock struct {
	t time.Time
}

func (f *fakeClock) now() time.Time { return f.t }

func (f *fakeClock) advance(d time.Duration) { f.t = f.t.Add(d) }

func newTestCache(ttl time.Duration) (*Cache, *fakeClock) {
	clock := &fakeClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	c := NewCache(ttl)
	c.mu.Lock()
	c.now = clock.now
	c.mu.Unlock()
	return c, clock
}

func TestCacheAddGet(t *testing.T) {
	c, _ := newTestCache(time.Minute)
	defer c.Stop()
	c.Add("key", []byte("value"))

	got, ok := c.Get("key")
	if !ok || string(got) != "value" {
		t.Fatalf("Get(key) = %q, %v; want \"value\", true", got, ok)
	}

	if _, ok := c.Get("missing"); ok {
		t.Error("Get(missing) should report a miss")
	}
}

func TestCacheExpiry(t *testing.T) {
	tests := []struct {
		name    string
		elapsed time.Duration
		wantHit bool
	}{
		{"fresh entry", 0, true},
		{"just before the TTL", 5*time.Minute - time.Second, true},
		{"exactly at the TTL", 5 * time.Minute, true},
		{"after the TTL", 5*time.Minute + time.Second, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, clock := newTestCache(5 * time.Minute)
			defer c.Stop()
			c.Add("key", []byte("value"))

			clock.advance(tt.elapsed)

			if _, ok := c.Get("key"); ok != tt.wantHit {
				t.Errorf("after %v: hit = %v, want %v", tt.elapsed, ok, tt.wantHit)
			}
		})
	}
}

func TestCacheReapRemovesOnlyExpiredEntries(t *testing.T) {
	c, clock := newTestCache(time.Minute)
	defer c.Stop()

	c.Add("old", []byte("1"))
	clock.advance(45 * time.Second)
	c.Add("new", []byte("2"))
	clock.advance(30 * time.Second)

	c.reap()

	if n := c.Len(); n != 1 {
		t.Fatalf("Len() = %d after reap, want 1", n)
	}
	if _, ok := c.Get("new"); !ok {
		t.Error("the recent entry should survive the reap")
	}
}

func TestNewCacheWithInvalidTTL(t *testing.T) {
	c := NewCache(0)
	defer c.Stop()
	c.Add("key", []byte("value"))
	if _, ok := c.Get("key"); !ok {
		t.Error("a cache created with a zero TTL should fall back to a default TTL")
	}
}
