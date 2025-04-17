package cache

import (
   "reflect"
   "testing"
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
       {name: "fuzzy typo sah", input: "sah", limit: 10, expected: []string{"Ashe"}},
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