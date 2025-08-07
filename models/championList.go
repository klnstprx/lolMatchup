package models

// ChampionSummary contains the minimal information about a champion needed for the map.
type ChampionList struct {
	Type    string                     `json:"type"`
	Format  string                     `json:"format"`
	Version string                     `json:"version"`
	Data    map[string]ChampionSummary `json:"data"`
}

// ChampionSummary is a reduced champion representation for listing.
type ChampionSummary struct {
   ID   string `json:"id"`
   Key  string `json:"key"`
   Name string `json:"name"`
}
