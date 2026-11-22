package cache

import (
	"testing"
	"time"
)

func TestCacheAddGet(t *testing.T) {
	c := NewCache(time.Minute)
	c.Add("key", []byte("value"))

	got, ok := c.Get("key")
	if !ok || string(got) != "value" {
		t.Fatalf("Get(key) = %q, %v; want \"value\", true", got, ok)
	}

	if _, ok := c.Get("missing"); ok {
		t.Error("Get(missing) should report a miss")
	}
}

func TestCacheExpires(t *testing.T) {
	const ttl = 20 * time.Millisecond
	c := NewCache(ttl)
	c.Add("key", []byte("value"))

	time.Sleep(3 * ttl)

	if _, ok := c.Get("key"); ok {
		t.Error("entry should have expired after its TTL")
	}
}
