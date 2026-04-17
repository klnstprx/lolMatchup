package models

type Champion struct {
	ID        int                          `json:"id"`
	Key       string                       `json:"key"`
	Name      string                       `json:"name"`
	Title     string                       `json:"title"`
	Icon      string                       `json:"icon"`
	Positions []string                     `json:"positions"`
	Roles     []string                     `json:"roles"`
	Stats     ChampionStats                `json:"stats"`
	Abilities map[string][]ChampionAbility `json:"abilities"`
}

type ChampionStats struct {
	Health struct {
		Flat     float64 `json:"flat"`
		PerLevel float64 `json:"perLevel"`
	} `json:"health"`
	HealthRegen struct {
		Flat     float64 `json:"flat"`
		PerLevel float64 `json:"perLevel"`
	} `json:"healthRegen"`
	Mana struct {
		Flat     float64 `json:"flat"`
		PerLevel float64 `json:"perLevel"`
	} `json:"mana"`
	ManaRegen struct {
		Flat     float64 `json:"flat"`
		PerLevel float64 `json:"perLevel"`
	} `json:"manaRegen"`
	Armor struct {
		Flat     float64 `json:"flat"`
		PerLevel float64 `json:"perLevel"`
	} `json:"armor"`
	MagicResistance struct {
		Flat     float64 `json:"flat"`
		PerLevel float64 `json:"perLevel"`
	} `json:"magicResistance"`
	AttackDamage struct {
		Flat     float64 `json:"flat"`
		PerLevel float64 `json:"perLevel"`
	} `json:"attackDamage"`
	Movespeed struct {
		Flat float64 `json:"flat"`
	} `json:"movespeed"`
	AttackSpeed struct {
		Flat     float64 `json:"flat"`
		PerLevel float64 `json:"perLevel"`
	} `json:"attackSpeed"`
	AttackRange struct {
		Flat float64 `json:"flat"`
	} `json:"attackRange"`
}

type ChampionAbility struct {
	Name         string    `json:"name"`
	Icon         string    `json:"icon"`
	Description  string    `json:"blurb"`
	Effects      []Effect  `json:"effects"`
	Cost         *Cost     `json:"cost"`
	Cooldown     *Cooldown `json:"cooldown"`
	CastTime     *string   `json:"castTime"`
	TargetRange  *string   `json:"targetRange"`
	EffectRadius *string   `json:"effectRadius"`
	Speed        *string   `json:"speed"`
}

type Effect struct {
	Description string     `json:"description"`
	Leveling    []Leveling `json:"leveling"`
}

type Leveling struct {
	Attribute string     `json:"attribute"`
	Modifiers []Modifier `json:"modifiers"`
}

type Modifier struct {
	Values []float64 `json:"values"`
	Units  []string  `json:"units"`
}

type Cost struct {
	Modifiers []Modifier `json:"modifiers"`
}

type Cooldown struct {
	Modifiers []Modifier `json:"modifiers"`
}
