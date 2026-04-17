package models

// LeagueEntryDTO represents a ranked queue entry from League v4 API.
type LeagueEntryDTO struct {
	QueueType    string `json:"queueType"`
	Tier         string `json:"tier"`
	Rank         string `json:"rank"`
	LeaguePoints int    `json:"leaguePoints"`
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
	HotStreak    bool   `json:"hotStreak"`
}

// ChampionPoolEntry is a computed summary of a player's performance on a champion.
type ChampionPoolEntry struct {
	ChampionName string
	ChampionID   int
	Games        int
	Wins         int
	WinRate      float64 // 0.0 to 1.0
	AvgKills     float64
	AvgDeaths    float64
	AvgAssists   float64
}
