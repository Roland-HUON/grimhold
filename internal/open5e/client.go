package open5e

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Roland-HUON/grimhold/internal/cache"
)

const baseURL = "https://api.open5e.com/v1/"

var ErrNotFound = errors.New("not found")

type Client struct {
	httpClient http.Client
	cache      *cache.Cache
}

func NewClient(timeout time.Duration, ttl time.Duration) *Client {
	return &Client{
		httpClient: http.Client{Timeout: timeout},
		cache:      cache.NewCache(ttl),
	}
}

func (c *Client) Close() {
	c.cache.Stop()
}

func (c *Client) fetch(url string) ([]byte, error) {
	cached, ok := c.cache.Get(url)
	if ok {
		return cached, nil
	}

	response, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("requesting %s: %w", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", response.StatusCode, url)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	c.cache.Add(url, data)
	return data, nil
}

func (c *Client) GetMonster(name string) (Monster, error) {
	data, err := c.fetch(baseURL + "monsters/" + url.PathEscape(name) + "/")
	if err != nil {
		return Monster{}, fmt.Errorf("fetching monster %q: %w", name, err)
	}

	var monster Monster
	if err := json.Unmarshal(data, &monster); err != nil {
		return Monster{}, fmt.Errorf("decoding monster %q: %w", name, err)
	}
	if monster.Slug == "" {
		return Monster{}, fmt.Errorf("fetching monster %q: %w", name, ErrNotFound)
	}

	return monster, nil
}

func (c *Client) GetPage(pageURL string) (MonstersPage, error) {
	currentURL := pageURL
	if pageURL == "" {
		currentURL = baseURL + "monsters/?limit=20&format=json"
	}

	data, err := c.fetch(currentURL)
	if err != nil {
		return MonstersPage{}, fmt.Errorf("fetching page %q: %w", currentURL, err)
	}

	var monsterPage MonstersPage
	if err := json.Unmarshal(data, &monsterPage); err != nil {
		return MonstersPage{}, fmt.Errorf("decoding page %q: %w", currentURL, err)
	}
	return monsterPage, nil
}

func (c *Client) GetWeapons() ([]Weapon, error) {
	var page WeaponsPage
	var equipment []Weapon
	data, err := c.fetch(baseURL + "weapons/?limit=100")
	if err != nil {
		return nil, fmt.Errorf("fetching weapons: %w", err)
	}
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, fmt.Errorf("decoding page %q: %w", baseURL+"weapons/?limit=100", err)
	}
	for _, weapon := range page.Results {
		if strings.Contains(weapon.DamageDice, "d") {
			equipment = append(equipment, weapon)
		}
	}
	return equipment, nil
}

func (c *Client) GetSpell(slug string) (Spell, error) {
	data, err := c.fetch(baseURL + "spells/" + url.PathEscape(slug) + "/")
	if err != nil {
		return Spell{}, fmt.Errorf("fetching spell %q: %w", slug, err)
	}

	var spell Spell
	if err := json.Unmarshal(data, &spell); err != nil {
		return Spell{}, fmt.Errorf("decoding spell %q: %w", slug, err)
	}
	if spell.Slug == "" {
		return Spell{}, fmt.Errorf("fetching spell %q: %w", slug, ErrNotFound)
	}

	return spell, nil
}
