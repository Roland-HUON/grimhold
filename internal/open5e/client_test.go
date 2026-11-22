package open5e

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

const pageJSON = `{
	"count": 2,
	"next": "https://example.com/next",
	"previous": null,
	"results": [
		{"slug": "goblin", "name": "Goblin", "armor_class": 15, "hit_points": 7, "cr": 0.25,
		 "actions": [{"name": "Scimitar", "attack_bonus": 4, "damage_dice": "1d6", "damage_bonus": 2}]},
		{"slug": "kobold", "name": "Kobold", "armor_class": 12, "hit_points": 5, "cr": 0.125}
	]
}`

func TestGetPageDecodesAndCaches(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Write([]byte(pageJSON))
	}))
	defer server.Close()

	client := NewClient(time.Second, time.Minute)
	defer client.Close()

	page, err := client.GetPage(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if page.Count != 2 || len(page.Monsters) != 2 || page.NextURL != "https://example.com/next" || page.PrevURL != "" {
		t.Fatalf("unexpected page: %+v", page)
	}
	goblin := page.Monsters[0]
	if goblin.Name != "Goblin" || goblin.HP != 7 || goblin.ArmorClass != 15 || goblin.CR != 0.25 {
		t.Errorf("unexpected goblin: %+v", goblin)
	}
	if len(goblin.Actions) != 1 || goblin.Actions[0].DamageDice != "1d6" || goblin.Actions[0].DamageBonus != 2 {
		t.Errorf("unexpected goblin actions: %+v", goblin.Actions)
	}

	if _, err := client.GetPage(server.URL); err != nil {
		t.Fatal(err)
	}
	if n := hits.Load(); n != 1 {
		t.Errorf("server was called %d times, want 1 (second call must come from the cache)", n)
	}
}

func TestGetPageErrors(t *testing.T) {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		wantNotFound bool
	}{
		{"not found", func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) }, true},
		{"server error", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }, false},
		{"invalid JSON", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("{not json")) }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClient(time.Second, time.Minute)
			defer client.Close()

			_, err := client.GetPage(server.URL)
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if got := errors.Is(err, ErrNotFound); got != tt.wantNotFound {
				t.Errorf("errors.Is(err, ErrNotFound) = %v, want %v (err: %v)", got, tt.wantNotFound, err)
			}
		})
	}
}
