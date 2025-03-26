package models

// Root represents the top-level structure returned by champion data endpoints.
type Root struct {
	Data map[string]Champion `json:"data"`
}

// Champion represents an individual champion's full data structure.
type Champion struct {
	ID      string  `json:"id"`
	Key     string  `json:"key"`
	Name    string  `json:"name"`
	Title   string  `json:"title"`
	Partype string  `json:"partype"`
	Spells  []Spell `json:"spells"`
	Passive Passive `json:"passive"`
	Image   Image   `json:"image"`
	Stats   Stats   `json:"stats"`
}

// Image holds sprite information for visual assets.
type Image struct {
	Full   string `json:"full"`
	Sprite string `json:"sprite"`
	Group  string `json:"group"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	W      int    `json:"w"`
	H      int    `json:"h"`
}

// Stats represents a champion's base stats and growth per level.
type Stats struct {
	HP                   float64 `json:"hp"`
	HPPerLevel           float64 `json:"hpperlevel"`
	MP                   float64 `json:"mp"`
	MPPerLevel           float64 `json:"mpperlevel"`
	MoveSpeed            float64 `json:"movespeed"`
	Armor                float64 `json:"armor"`
	ArmorPerLevel        float64 `json:"armorperlevel"`
	SpellBlock           float64 `json:"spellblock"`
	SpellBlockPerLevel   float64 `json:"spellblockperlevel"`
	AttackRange          float64 `json:"attackrange"`
	HPRegen              float64 `json:"hpregen"`
	HPRegenPerLevel      float64 `json:"hpregenperlevel"`
	MPRegen              float64 `json:"mpregen"`
	MPRegenPerLevel      float64 `json:"mpregenperlevel"`
	Crit                 float64 `json:"crit"`
	CritPerLevel         float64 `json:"critperlevel"`
	AttackDamage         float64 `json:"attackdamage"`
	AttackDamagePerLevel float64 `json:"attackdamageperlevel"`
	AttackSpeedPerLevel  float64 `json:"attackspeedperlevel"`
	AttackSpeed          float64 `json:"attackspeed"`
}

// Spell represents details of an ability in a champion's kit.
type Spell struct {
	DataValues   DataValues    `json:"datavalues"`
	MaxAmmo      string        `json:"maxammo"`
	CostType     string        `json:"costType"`
	CostBurn     string        `json:"costBurn"`
	Name         string        `json:"name"`
	RangeBurn    string        `json:"rangeBurn"`
	ID           string        `json:"id"`
	CooldownBurn string        `json:"cooldownBurn"`
	Description  string        `json:"description"`
	Tooltip      string        `json:"tooltip"`
	Resource     string        `json:"resource"`
	LevelTip     LevelTip      `json:"leveltip"`
	EffectBurn   []string      `json:"effectBurn"`
	Effect       []interface{} `json:"effect"`
	Vars         []Var         `json:"vars"`
	Cost         []int         `json:"cost"`
	Cooldown     []float64     `json:"cooldown"`
	Range        []int         `json:"range"`
	Image        Image         `json:"image"`
	MaxRank      int           `json:"maxrank"`
}

// LevelTip provides textual hints for leveling up spells.
type LevelTip struct {
	Label  []string `json:"label"`
	Effect []string `json:"effect"`
}

// DataValues unused in typical champion JSON, placeholder for future expansions.
type DataValues struct {
	// Add fields here if data becomes available.
}

// Var describes dynamic scaling or coefficient data for spells.
type Var struct {
	Coeff interface{} `json:"coeff"`
	Key   string      `json:"key"`
	Link  string      `json:"link"`
}

// Passive is the champion's innate ability.
type Passive struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       Image  `json:"image"`
}
