package models

// Root represents the top-level structure containing the data.
type Root struct {
	Data map[string]Champion `json:"data"`
}

// Champion represents the data for an individual champion.
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

// Image represents image details for champions and spells.
type Image struct {
	Full   string `json:"full"`
	Sprite string `json:"sprite"`
	Group  string `json:"group"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	W      int    `json:"w"`
	H      int    `json:"h"`
}

// Stats represents the statistical data of a champion.
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

// Spell represents the details of a champion's spell.
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

// LevelTip represents the leveling information for a spell.
type LevelTip struct {
	Label  []string `json:"label"`
	Effect []string `json:"effect"`
}

// DataValues represents additional data for a spell (can be expanded as needed).
type DataValues struct {
	// Add fields here if data becomes available.
}

// Var represents variable coefficients in a spell's calculation.
type Var struct {
	Coeff interface{} `json:"coeff"`
	Key   string      `json:"key"`
	Link  string      `json:"link"`
}

// Passive represents a champion's passive ability.
type Passive struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       Image  `json:"image"`
}
