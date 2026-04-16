package models

// CurrentGameInfo is the top-level response from Spectator v5.
type CurrentGameInfo struct {
	GameID            int64                    `json:"gameId"`
	GameType          string                   `json:"gameType"`
	GameStartTime     int64                    `json:"gameStartTime"`
	MapID             int64                    `json:"mapId"`
	GameLength        int64                    `json:"gameLength"`
	PlatformID        string                   `json:"platformId"`
	GameMode          string                   `json:"gameMode"`
	GameQueueConfigID int64                    `json:"gameQueueConfigId"`
	Participants      []CurrentGameParticipant `json:"participants"`
	BannedChampions   []BannedChampion         `json:"bannedChampions"`
}

// CurrentGameParticipant represents a player in a live game.
type CurrentGameParticipant struct {
	ChampionID    int64  `json:"championId"`
	TeamID        int64  `json:"teamId"`
	RiotID        string `json:"riotId"`
	PUUID         string `json:"puuid"`
	ProfileIconID int64  `json:"profileIconId"`
	Bot           bool   `json:"bot"`
	Spell1ID      int64  `json:"spell1Id"`
	Spell2ID      int64  `json:"spell2Id"`
}

// BannedChampion represents a banned champion in champion select.
type BannedChampion struct {
	PickTurn   int   `json:"pickTurn"`
	ChampionID int64 `json:"championId"`
	TeamID     int64 `json:"teamId"`
}
