package components

import (
  "reflect"
  "testing"
  "github.com/klnstprx/lolMatchup/models"
)

// TestRenderSpellTooltip covers replacement of placeholders in spell.Tooltip
func TestRenderSpellTooltip(t *testing.T) {
   tests := []struct {
       name     string
       spell    models.Spell
       expected string
   }{
       {
           name: "simple eN replacement",
           spell: models.Spell{
               Tooltip:    "Hit deals {{e1}} magic damage.",
               EffectBurn: []string{"", "150"},
           },
           expected: "Hit deals 150 magic damage.",
       },
       {
           name: "vars replacement",
           spell: models.Spell{
               Tooltip:    "Heals for {{a1}} health.",
               EffectBurn: []string{""},
               Vars: []models.Var{{Key: "a1", Coeff: []interface{}{0.75}}},
           },
           expected: "Heals for 0.75 health.",
       },
       {
           name: "mixed replacement",
           spell: models.Spell{
               Tooltip:    "Damage {{e1}} (+{{a1}}).",
               EffectBurn: []string{"", "200"},
               Vars:       []models.Var{{Key: "a1", Coeff: []interface{}{1.2}}},
           },
           expected: "Damage 200 (+1.20).",
       },
       {
           name: "unknown placeholder removed",
           spell: models.Spell{
               Tooltip:    "{{foo}} text {{e1}} {{bar}}.",
               EffectBurn: []string{"", "300"},
               Vars:       []models.Var{},
           },
           expected: " text 300 .",
       },
       {
           name: "no placeholders",
           spell: models.Spell{
               Tooltip:    "Just text.",
               EffectBurn: []string{},
               Vars:       []models.Var{},
           },
           expected: "Just text.",
       },
   }
   for _, tt := range tests {
       t.Run(tt.name, func(t *testing.T) {
           out := RenderSpellTooltip(tt.spell)
           // compare as string
           got := string(out)
           if !reflect.DeepEqual(got, tt.expected) {
               t.Errorf("%s: expected %q, got %q", tt.name, tt.expected, got)
           }
       })
   }
}
