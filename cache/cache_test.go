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
		{name: "fuzzy typo sah", input: "sah", limit: 10, expected: []string{"Ashe", "Ahri"}},
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
	champA := models.Champion{ID: "A", Key: "1", Name: "Ahri", Title: "TitleA"}
	champB := models.Champion{ID: "B", Key: "2", Name: "Ashe", Title: "TitleB"}
	orig.SetChampion(champA)
	orig.SetChampion(champB)

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
		got, ok := loaded.GetChampionByID(want.ID)
		if !ok {
			t.Errorf("Champion %s missing after Load", want.ID)
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Champion %s: got %+v, want %+v", want.ID, got, want)
		}
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
	initialChamp := models.Champion{ID: "X", Name: "X"}
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
	initialChamp := models.Champion{ID: "X", Name: "X"}
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

