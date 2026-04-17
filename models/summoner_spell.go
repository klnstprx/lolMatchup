package models

// SummonerSpell holds resolved summoner spell info from DDragon.
type SummonerSpell struct {
	Name      string  `json:"name"`
	Key       string  `json:"key"`       // numeric ID as string, e.g. "4"
	ImageFull string  `json:"imageFull"` // e.g. "SummonerFlash.png"
	Cooldown  float64 `json:"cooldown"`
}

// DDragonSpellData is the top-level response from DDragon summoner.json.
type DDragonSpellData struct {
	Data map[string]DDragonSpell `json:"data"`
}

// DDragonSpell represents a single spell entry in DDragon summoner.json.
type DDragonSpell struct {
	Name     string    `json:"name"`
	Key      string    `json:"key"`
	Cooldown []float64 `json:"cooldown"`
	Image    struct {
		Full string `json:"full"`
	} `json:"image"`
}

// ParseSummonerSpells builds a map keyed by numeric spell ID from DDragon data.
func ParseSummonerSpells(data DDragonSpellData) map[string]SummonerSpell {
	m := make(map[string]SummonerSpell, len(data.Data))
	for _, spell := range data.Data {
		cd := 0.0
		if len(spell.Cooldown) > 0 {
			cd = spell.Cooldown[0]
		}
		m[spell.Key] = SummonerSpell{
			Name:      spell.Name,
			Key:       spell.Key,
			ImageFull: spell.Image.Full,
			Cooldown:  cd,
		}
	}
	return m
}
