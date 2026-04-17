package cache

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/klnstprx/lolMatchup/models"
)

// TestAutocomplete exercises the Autocomplete method under various scenarios.
func TestAutocomplete(t *testing.T) {
	// Initialize cache with a small champion map
	c := New("", 3)
	c.SetChampionMap(map[string]string{
		"Ashe":   "Ashe",
		"Azir":   "Azir",
		"Anivia": "Anivia",
		"Ahri":   "Ahri",
		"Braum":  "Braum",
	})

	tests := []struct {
		name     string
		input    string
		limit    int
		expected []string
	}{
		{name: "empty input", input: "", limit: 10, expected: nil},
		{name: "prefix a", input: "a", limit: 10, expected: []string{"Ahri", "Anivia", "Ashe", "Azir"}},
		{name: "prefix az", input: "az", limit: 10, expected: []string{"Azir"}},
		{name: "fuzzy typo brom", input: "brom", limit: 10, expected: []string{"Braum"}},
		{name: "fuzzy typo sah", input: "sah", limit: 10, expected: []string{"Ahri", "Ashe"}},
		{name: "limit results", input: "a", limit: 2, expected: []string{"Ahri", "Anivia"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := c.Autocomplete(tc.input, tc.limit)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("Autocomplete(%q, %d) = %v, want %v", tc.input, tc.limit, got, tc.expected)
			}
		})
	}
}

// TestSaveLoadCache verifies that Save and Load correctly persist and restore cache data.
func TestSaveLoadCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	// Prepare original cache with test data
	orig := New(path, 5)
	orig.Patch = "1.2.3"
	champMap := map[string]string{"Ahri": "A", "Ashe": "B"}
	orig.SetChampionMap(champMap)
	champA := models.Champion{ID: 1, Key: "Ahri", Name: "Ahri", Title: "the Nine-Tailed Fox"}
	champB := models.Champion{ID: 2, Key: "Ashe", Name: "Ashe", Title: "the Frost Archer"}
	orig.SetChampion(champA)
	orig.SetChampion(champB)
	spells := map[string]models.SummonerSpell{
		"4":  {Name: "Flash", Key: "4", ImageFull: "SummonerFlash.png", Cooldown: 300},
		"14": {Name: "Ignite", Key: "14", ImageFull: "SummonerDot.png", Cooldown: 180},
	}
	orig.SetSummonerSpells(spells)

	// Save to disk
	if err := orig.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load into a new cache instance
	loaded := New(path, 0)
	if err := loaded.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Check patch
	if loaded.Patch != orig.Patch {
		t.Errorf("Patch: got %q, want %q", loaded.Patch, orig.Patch)
	}
	// Check champion map
	if !reflect.DeepEqual(loaded.GetChampionMap(), champMap) {
		t.Errorf("ChampionMap: got %v, want %v", loaded.GetChampionMap(), champMap)
	}
	// Check champions
	for _, want := range []models.Champion{champA, champB} {
		got, ok := loaded.GetChampionByID(want.Key)
		if !ok {
			t.Errorf("Champion %s missing after Load", want.Key)
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Champion %s: got %+v, want %+v", want.Key, got, want)
		}
	}
	// Check summoner spells
	if !reflect.DeepEqual(loaded.GetSummonerSpells(), spells) {
		t.Errorf("SummonerSpells: got %v, want %v", loaded.GetSummonerSpells(), spells)
	}
}

// TestLoadNonexistentCache ensures loading a non-existent file leaves cache unchanged.
func TestLoadNonexistentCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nope.json")
	c := New(path, 0)
	// Set initial state
	c.Patch = "orig"
	initialMap := map[string]string{"X": "Y"}
	c.ChampionMap = initialMap
	initialChamp := models.Champion{ID: 1, Key: "X", Name: "X"}
	c.Champions = map[string]models.Champion{"X": initialChamp}

	// Load should not error and should not modify existing data
	if err := c.Load(); err != nil {
		t.Errorf("Load() error for non-existent file: %v", err)
	}
	if c.Patch != "orig" {
		t.Errorf("Patch changed: got %q, want %q", c.Patch, "orig")
	}
	if !reflect.DeepEqual(c.GetChampionMap(), initialMap) {
		t.Errorf("ChampionMap changed: got %v, want %v", c.GetChampionMap(), initialMap)
	}
	got, ok := c.GetChampionByID("X")
	if !ok || !reflect.DeepEqual(got, initialChamp) {
		t.Errorf("Champions changed: got %+v, ok=%v, want %+v", got, ok, initialChamp)
	}
}

// TestGetSetPatch verifies that GetPatch and SetPatch are thread-safe round-trips.
func TestGetSetPatch(t *testing.T) {
	c := New("", 3)
	if got := c.GetPatch(); got != "" {
		t.Errorf("initial GetPatch() = %q, want empty", got)
	}
	c.SetPatch("14.9.1")
	if got := c.GetPatch(); got != "14.9.1" {
		t.Errorf("GetPatch() = %q, want %q", got, "14.9.1")
	}
	c.SetPatch("15.1.1")
	if got := c.GetPatch(); got != "15.1.1" {
		t.Errorf("GetPatch() = %q, want %q", got, "15.1.1")
	}
}

// TestGetChampionMapLen verifies the champion map length accessor.
func TestGetChampionMapLen(t *testing.T) {
	c := New("", 3)
	if got := c.GetChampionMapLen(); got != 0 {
		t.Errorf("initial GetChampionMapLen() = %d, want 0", got)
	}
	c.SetChampionMap(map[string]string{"Ahri": "Ahri", "Ashe": "Ashe"})
	if got := c.GetChampionMapLen(); got != 2 {
		t.Errorf("GetChampionMapLen() = %d, want 2", got)
	}
}

// TestFuzzyScore tests the shared fuzzy scoring helper.
func TestFuzzyScore(t *testing.T) {
	tests := []struct {
		name      string
		typed     string
		candidate string
		threshold int
		wantOK    bool
	}{
		{"exact match", "ashe", "ashe", 3, true},
		{"prefix match", "ash", "ashe", 3, true},
		{"substring match", "sh", "ashe", 3, true},
		{"close typo", "ahse", "ashe", 3, true},
		{"too far", "zzzzz", "ashe", 3, false},
		{"empty typed", "", "ashe", 3, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := fuzzyScore(tc.typed, tc.candidate, tc.threshold)
			if ok != tc.wantOK {
				t.Errorf("fuzzyScore(%q, %q, %d): ok=%v, want %v", tc.typed, tc.candidate, tc.threshold, ok, tc.wantOK)
			}
		})
	}
}

// TestSummonerSpellsGetSet verifies summoner spell get/set and length methods.
func TestSummonerSpellsGetSet(t *testing.T) {
	c := New("", 3)
	if got := c.GetSummonerSpellsLen(); got != 0 {
		t.Errorf("initial GetSummonerSpellsLen() = %d, want 0", got)
	}
	spells := map[string]models.SummonerSpell{
		"4":  {Name: "Flash", Key: "4", ImageFull: "SummonerFlash.png", Cooldown: 300},
		"14": {Name: "Ignite", Key: "14", ImageFull: "SummonerDot.png", Cooldown: 180},
	}
	c.SetSummonerSpells(spells)
	if got := c.GetSummonerSpellsLen(); got != 2 {
		t.Errorf("GetSummonerSpellsLen() = %d, want 2", got)
	}
	if !reflect.DeepEqual(c.GetSummonerSpells(), spells) {
		t.Errorf("GetSummonerSpells() = %v, want %v", c.GetSummonerSpells(), spells)
	}
}

// TestInvalidateResetsSummonerSpells verifies Invalidate clears summoner spells.
func TestInvalidateResetsSummonerSpells(t *testing.T) {
	c := New("", 3)
	c.SetSummonerSpells(map[string]models.SummonerSpell{
		"4": {Name: "Flash", Key: "4", ImageFull: "SummonerFlash.png", Cooldown: 300},
	})
	if c.GetSummonerSpellsLen() == 0 {
		t.Fatal("expected non-empty spells before invalidate")
	}
	c.Invalidate()
	if got := c.GetSummonerSpellsLen(); got != 0 {
		t.Errorf("GetSummonerSpellsLen() after Invalidate() = %d, want 0", got)
	}
}

// TestLoadInvalidCache ensures invalid JSON is ignored and existing cache data is retained.
func TestLoadInvalidCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")
	// Write invalid content
	if err := os.WriteFile(path, []byte("not-json"), 0644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
	c := New(path, 0)
	c.Patch = "orig"
	initialMap := map[string]string{"X": "Y"}
	c.ChampionMap = initialMap
	initialChamp := models.Champion{ID: 1, Key: "X", Name: "X"}
	c.Champions = map[string]models.Champion{"X": initialChamp}

	if err := c.Load(); err != nil {
		t.Errorf("Load() error for invalid JSON: %v", err)
	}
	if c.Patch != "orig" {
		t.Errorf("Patch changed after invalid load: got %q, want %q", c.Patch, "orig")
	}
	if !reflect.DeepEqual(c.GetChampionMap(), initialMap) {
		t.Errorf("ChampionMap changed after invalid load: got %v, want %v", c.GetChampionMap(), initialMap)
	}
	got, ok := c.GetChampionByID("X")
	if !ok || !reflect.DeepEqual(got, initialChamp) {
		t.Errorf("Champions changed after invalid load: got %+v, ok=%v, want %+v", got, ok, initialChamp)
	}
}
