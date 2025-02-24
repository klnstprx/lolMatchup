package models

type ChampionList struct {
	Type    string                     `json:"type"`
	Format  string                     `json:"format"`
	Version string                     `json:"version"`
	Data    map[string]ChampionSummary `json:"data"`
}

// ChampionSummary contains the minimal information about a champion needed for the map.
type ChampionSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
