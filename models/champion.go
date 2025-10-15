package models

type Champion struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	Title string `json:"title"`
	Icon  string `json:"icon"`
	Stats ChampionStats `json:"stats"`
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
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Description string `json:"blurb"`
	Effects     []struct {
		Description string `json:"description"`
	} `json:"effects"`
}
