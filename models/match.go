package models

// MatchDTO is the top-level response from match-v5.
type MatchDTO struct {
	Metadata MatchMetadata `json:"metadata"`
	Info     MatchInfo     `json:"info"`
}

// MatchMetadata contains match ID and participant PUUIDs.
type MatchMetadata struct {
	MatchID      string   `json:"matchId"`
	Participants []string `json:"participants"`
}

// MatchInfo contains game-level data and per-participant stats.
type MatchInfo struct {
	GameDuration       int64              `json:"gameDuration"`
	GameMode           string             `json:"gameMode"`
	GameStartTimestamp int64              `json:"gameStartTimestamp"`
	GameEndTimestamp   int64              `json:"gameEndTimestamp"`
	QueueID            int                `json:"queueId"`
	Participants       []MatchParticipant `json:"participants"`
}

// MatchParticipant holds per-player stats within a match.
type MatchParticipant struct {
	PUUID              string `json:"puuid"`
	RiotIDGameName     string `json:"riotIdGameName"`
	RiotIDTagline      string `json:"riotIdTagline"`
	TeamID             int    `json:"teamId"`
	ChampionName       string `json:"championName"`
	ChampionID         int    `json:"championId"`
	Win                bool   `json:"win"`
	Kills              int    `json:"kills"`
	Deaths             int    `json:"deaths"`
	Assists            int    `json:"assists"`
	ChampLevel         int    `json:"champLevel"`
	IndividualPosition string `json:"individualPosition"`

	// CS
	TotalMinionsKilled   int `json:"totalMinionsKilled"`
	NeutralMinionsKilled int `json:"neutralMinionsKilled"`

	// Damage
	TotalDamageDealtToChampions    int `json:"totalDamageDealtToChampions"`
	PhysicalDamageDealtToChampions int `json:"physicalDamageDealtToChampions"`
	MagicDamageDealtToChampions    int `json:"magicDamageDealtToChampions"`
	TrueDamageDealtToChampions     int `json:"trueDamageDealtToChampions"`
	TotalDamageTaken               int `json:"totalDamageTaken"`
	DamageSelfMitigated            int `json:"damageSelfMitigated"`
	DamageDealtToObjectives        int `json:"damageDealtToObjectives"`

	// Economy
	GoldEarned int `json:"goldEarned"`
	GoldSpent  int `json:"goldSpent"`

	// Vision
	VisionScore         int `json:"visionScore"`
	WardsPlaced         int `json:"wardsPlaced"`
	WardsKilled         int `json:"wardsKilled"`
	DetectorWardsPlaced int `json:"detectorWardsPlaced"`

	// Combat
	DoubleKills         int  `json:"doubleKills"`
	TripleKills         int  `json:"tripleKills"`
	QuadraKills         int  `json:"quadraKills"`
	PentaKills          int  `json:"pentaKills"`
	LargestKillingSpree int  `json:"largestKillingSpree"`
	LargestMultiKill    int  `json:"largestMultiKill"`
	TotalTimeSpentDead  int  `json:"totalTimeSpentDead"`
	FirstBloodKill      bool `json:"firstBloodKill"`

	// Objectives
	TurretTakedowns    int `json:"turretTakedowns"`
	InhibitorTakedowns int `json:"inhibitorTakedowns"`
	BaronKills         int `json:"baronKills"`
	DragonKills        int `json:"dragonKills"`
	ObjectivesStolen   int `json:"objectivesStolen"`

	// Healing/shielding
	TotalHeal                      int `json:"totalHeal"`
	TotalHealsOnTeammates          int `json:"totalHealsOnTeammates"`
	TotalDamageShieldedOnTeammates int `json:"totalDamageShieldedOnTeammates"`

	// Summoner spells
	Summoner1Id int `json:"summoner1Id"`
	Summoner2Id int `json:"summoner2Id"`

	// Items
	Item0 int `json:"item0"`
	Item1 int `json:"item1"`
	Item2 int `json:"item2"`
	Item3 int `json:"item3"`
	Item4 int `json:"item4"`
	Item5 int `json:"item5"`
	Item6 int `json:"item6"`
}

// PlayerStatsContext holds computed data for the player stats modal.
type PlayerStatsContext struct {
	Player            MatchParticipant
	MatchID           string
	GameDuration      int64
	TeamTotalDamage   int
	MaxGoldInGame     int
	MaxVisionInGame   int
	MaxDamageInGame   int
	KillParticipation float64
	PatchNumber       string
}

// MatchupRecord tracks win/loss for a specific champion vs champion lane pairing.
type MatchupRecord struct {
	PlayerChampion string
	EnemyChampion  string
	Wins           int
	Losses         int
	Games          int
}

// OpponentEnrichment holds data about a live game opponent derived from their recent matches.
type OpponentEnrichment struct {
	ChampionWins       int
	ChampionLosses     int
	ChampionGames      int
	WinStreak          int
	LossStreak         int
	MostPlayedPosition string
	PossiblyOffRole    bool    // true if current inferred role differs from most-played role
	TotalGames         int     // total matches analyzed
	RecentWinRate      float64 // overall win rate across recent matches (0.0-1.0)
	IsOTP              bool    // true if >= 60% of recent games on one champion
}

// MatchSummary is a condensed view of a player's performance in a match,
// extracted from MatchDTO for template rendering.
type MatchSummary struct {
	MatchID       string
	ChampionName  string
	ChampionID    int
	Win           bool
	Kills         int
	Deaths        int
	Assists       int
	CS            int // totalMinionsKilled + neutralMinionsKilled
	Position      string
	GameDuration  int64 // seconds
	GameStartTime int64 // epoch ms
	Damage        int
	Gold          int
	VisionScore   int
	Items         [7]int
	QueueID       int
}
