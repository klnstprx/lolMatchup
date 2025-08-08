package components

import (
	"fmt"
	"html/template"
	"regexp"
	"strconv"

	"github.com/klnstprx/lolMatchup/models"
)

// RenderSpellTooltip replaces template placeholders in spell.Tooltip with actual values.
func RenderSpellTooltip(spell models.Spell) template.HTML {
	text := spell.Tooltip
	// replace {{ eN }} with effectBurn[N]
	reE := regexp.MustCompile(`{{\s*e(\d+)\s*}}`)
	text = reE.ReplaceAllStringFunc(text, func(m string) string {
		parts := reE.FindStringSubmatch(m)
		if len(parts) != 2 {
			return m
		}
		idx, err := strconv.Atoi(parts[1])
		if err != nil || idx < 1 || idx > len(spell.EffectBurn)-1 {
			return m
		}
		return spell.EffectBurn[idx]
	})
	// replace {{ aN }} or {{ fN }} with vars coefficients
	reA := regexp.MustCompile(`{{\s*([af])(\d+)\s*}}`)
	text = reA.ReplaceAllStringFunc(text, func(m string) string {
		parts := reA.FindStringSubmatch(m)
		if len(parts) != 3 {
			return m
		}
		key := parts[1] + parts[2]
		for _, v := range spell.Vars {
			if v.Key == key {
				// take first coeff value
				switch c := v.Coeff.(type) {
				case float64:
					return fmt.Sprintf("%.2f", c)
				case []interface{}:
					if len(c) > 0 {
						if f, ok := c[0].(float64); ok {
							return fmt.Sprintf("%.2f", f)
						}
					}
				}
			}
		}
		return m
	})
	reUnknown := regexp.MustCompile(`{{[^}]*}}`)
	text = reUnknown.ReplaceAllString(text, "")
	return template.HTML(text)
}
