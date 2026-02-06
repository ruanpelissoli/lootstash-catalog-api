package d2

import (
	"fmt"
	"strings"
)

// PropertyTranslator converts internal property codes to human-readable text
type PropertyTranslator struct {
	// Map of property code to display format
	// Format placeholders: {value}, {min}, {max}, {param}
	formats map[string]string

	// Skill tab names indexed by tab number
	skillTabs map[int]string
}

// NewPropertyTranslator creates a new property translator with D2 property formats
func NewPropertyTranslator() *PropertyTranslator {
	return &PropertyTranslator{
		formats: map[string]string{
			// Skills
			"allskills":       "+{value} To All Skills",
			"skill":           "+{value} To {param}",
			"skilltab":        "+{value} To {skilltab}",
			"aura":            "Level {value} {param} Aura When Equipped",
			"oskill":          "+{value} To {param}",
			"charged":         "Level {min} {param} ({max} Charges)",

			// Class skills
			"ama": "+{value} To Amazon Skill Levels",
			"sor": "+{value} To Sorceress Skill Levels",
			"nec": "+{value} To Necromancer Skill Levels",
			"pal": "+{value} To Paladin Skill Levels",
			"bar": "+{value} To Barbarian Skill Levels",
			"dru": "+{value} To Druid Skill Levels",
			"ass": "+{value} To Assassin Skill Levels",

			// Random class skill (Hellfire Torch, etc.)
			"randclassskill": "+{value} To Random Character Class Skills",

			// Attributes
			"str":        "+{value} To Strength",
			"dex":        "+{value} To Dexterity",
			"vit":        "+{value} To Vitality",
			"enr":        "+{value} To Energy",
			"all-stats":  "+{value} To All Attributes",

			// Life/Mana
			"hp":         "+{value} To Life",
			"mana":       "+{value} To Mana",
			"hp%":        "+{value}% To Life",
			"mana%":      "+{value}% To Mana",
			"regen-mana": "Regenerate Mana {value}%",
			"regen":      "Replenish Life +{value}",

			// Defense
			"ac":         "+{value} Defense",
			"ac%":        "+{value}% Enhanced Defense",
			"ac-miss":    "+{value} Defense vs. Missile",
			"red-dmg":    "Damage Reduced By {value}",
			"red-dmg%":   "Damage Reduced By {value}%",
			"red-mag":    "Magic Damage Reduced By {value}",

			// Attack
			"dmg%":           "+{value}% Enhanced Damage",
			"dmg":            "+{value} Damage",
			"dmg-min":        "+{value} To Minimum Damage",
			"dmg-max":        "+{value} To Maximum Damage",
			"ltng-min":       "+{value} To Minimum Lightning Damage",
			"ltng-max":       "+{value} To Maximum Lightning Damage",
			"fire-min":       "+{value} To Minimum Fire Damage",
			"fire-max":       "+{value} To Maximum Fire Damage",
			"cold-min":       "+{value} To Minimum Cold Damage",
			"cold-max":       "+{value} To Maximum Cold Damage",
			"pois-min":       "+{value} To Minimum Poison Damage",
			"pois-max":       "+{value} To Maximum Poison Damage",
			"mag-min":        "+{value} To Minimum Magic Damage",
			"mag-max":        "+{value} To Maximum Magic Damage",
			"dmg-norm":       "Adds {min}-{max} Damage",
			"dmg-fire":       "Adds {min}-{max} Fire Damage",
			"dmg-cold":       "Adds {min}-{max} Cold Damage",
			"dmg-ltng":       "Adds {min}-{max} Lightning Damage",
			"dmg-pois":       "+{value} Poison Damage Over {param} Seconds",
			"dmg-mag":        "Adds {min}-{max} Magic Damage",
			"extra-fire":     "+{value}% To Fire Skill Damage",
			"extra-cold":     "+{value}% To Cold Skill Damage",
			"extra-ltng":     "+{value}% To Lightning Skill Damage",
			"extra-pois":     "+{value}% To Poison Skill Damage",

			// Attack Rating
			"att":        "+{value} To Attack Rating",
			"att%":       "+{value}% To Attack Rating",
			"att-demon":  "+{value} To Attack Rating Against Demons",
			"att-undead": "+{value} To Attack Rating Against Undead",

			// Speed
			"swing1":         "+{value}% Increased Attack Speed",
			"swing2":         "+{value}% Increased Attack Speed",
			"swing3":         "+{value}% Increased Attack Speed",
			"cast1":          "+{value}% Faster Cast Rate",
			"cast2":          "+{value}% Faster Cast Rate",
			"cast3":          "+{value}% Faster Cast Rate",
			"move1":          "+{value}% Faster Run/Walk",
			"move2":          "+{value}% Faster Run/Walk",
			"move3":          "+{value}% Faster Run/Walk",
			"block":          "+{value}% Faster Block Rate",
			"block1":         "+{value}% Faster Block Rate",
			"block2":         "+{value}% Faster Block Rate",
			"block3":         "+{value}% Faster Block Rate",
			"balance1":       "+{value}% Faster Hit Recovery",
			"balance2":       "+{value}% Faster Hit Recovery",
			"balance3":       "+{value}% Faster Hit Recovery",

			// Resistances
			"res-fire":       "Fire Resist +{value}%",
			"res-cold":       "Cold Resist +{value}%",
			"res-ltng":       "Lightning Resist +{value}%",
			"res-pois":       "Poison Resist +{value}%",
			"res-all":        "All Resistances +{value}",
			"res-mag":        "Magic Resist +{value}%",
			"abs-fire":       "+{value} Fire Absorb",
			"abs-cold":       "+{value} Cold Absorb",
			"abs-ltng":       "+{value} Lightning Absorb",
			"abs-fire%":      "{value}% Fire Absorb",
			"abs-cold%":      "{value}% Cold Absorb",
			"abs-ltng%":      "{value}% Lightning Absorb",

			// Pierce
			"pierce-fire":    "-{value}% To Enemy Fire Resistance",
			"pierce-cold":    "-{value}% To Enemy Cold Resistance",
			"pierce-ltng":    "-{value}% To Enemy Lightning Resistance",
			"pierce-pois":    "-{value}% To Enemy Poison Resistance",

			// Sunder Charms (D2R Patch 2.5)
			"pierce-immunity-cold":   "Monster Cold Immunity is Sundered",
			"pierce-immunity-fire":   "Monster Fire Immunity is Sundered",
			"pierce-immunity-light":  "Monster Lightning Immunity is Sundered",
			"pierce-immunity-poison": "Monster Poison Immunity is Sundered",
			"pierce-immunity-damage": "Monster Physical Immunity is Sundered",
			"pierce-immunity-magic":  "Monster Magic Immunity is Sundered",

			// Leech
			"lifesteal":      "{value}% Life Stolen Per Hit",
			"manasteal":      "{value}% Mana Stolen Per Hit",

			// Kill bonuses
			"hp/kill":        "+{value} Life After Each Kill",
			"mana/kill":      "+{value} Mana After Each Kill",
			"heal-kill":      "+{value} Life After Each Kill",
			"mana-kill":      "+{value} Mana After Each Kill",
			"hp/lvl":         "+{value} To Life (Based On Character Level)",
			"mana/lvl":       "+{value} To Mana (Based On Character Level)",

			// Magic Find
			"mag%":           "+{value}% Better Chance Of Getting Magic Items",
			"gold%":          "+{value}% Extra Gold From Monsters",

			// Other
			"light":          "+{value} To Light Radius",
			"thorns":         "Attacker Takes Damage Of {value}",
			"nofreeze":       "Cannot Be Frozen",
			"half-freeze":    "Half Freeze Duration",
			"ignore-ac":      "Ignore Target's Defense",
			"knock":          "Knockback",
			"slow":           "Slows Target By {value}%",
			"howl":           "Hit Causes Monster To Flee {value}%",
			"stupidity":      "Hit Blinds Target +{value}",
			"crush":          "{value}% Chance Of Crushing Blow",
			"deadly":         "{value}% Deadly Strike",
			"openwounds":     "{value}% Chance Of Open Wounds",
			"dmg-demon":      "+{value}% Damage To Demons",
			"dmg-undead":     "+{value}% Damage To Undead",
			"indestruct":     "Indestructible",
			"ethereal":       "Ethereal (Cannot Be Repaired)",
			"sock":           "Socketed ({value})",
			"rep-dur":        "Repairs 1 Durability In {value} Seconds",
			"rep-quant":      "Replenishes Quantity",
			"stack":          "+{value} To Maximum Quantity",
			"bloody":         "Slain Monsters Rest In Peace",

			// Per level bonuses
			"str/lvl":        "+{value} To Strength (Based On Character Level)",
			"dex/lvl":        "+{value} To Dexterity (Based On Character Level)",
			"vit/lvl":        "+{value} To Vitality (Based On Character Level)",
			"enr/lvl":        "+{value} To Energy (Based On Character Level)",
			"ac/lvl":         "+{value} Defense (Based On Character Level)",
			"ac%/lvl":        "+{value}% Enhanced Defense (Based On Character Level)",
			"dmg%/lvl":       "+{value}% Enhanced Damage (Based On Character Level)",
			"dmg/lvl":        "+{value} To Maximum Damage (Based On Character Level)",
			"att/lvl":        "+{value} To Attack Rating (Based On Character Level)",
			"att%/lvl":       "+{value}% To Attack Rating (Based On Character Level)",

			// Teleport special
			"teleport": "+1 To Teleport",

			// Exp
			"exp":            "+{value}% To Experience Gained",

			// Requirements
			"ease":           "Requirements -{value}%",

			// Defense per time
			"dmg-ac":         "{value}% Damage Taken Goes To Mana",

			// Chance to cast (min = chance %, max = skill level, param = skill name)
			"hit-skill":      "{min}% Chance To Cast Level {max} {param} On Striking",
			"gethit-skill":   "{min}% Chance To Cast Level {max} {param} When Struck",
			"kill-skill":     "{min}% Chance To Cast Level {max} {param} On Kill",
			"death-skill":    "{min}% Chance To Cast Level {max} {param} On Death",
			"levelup-skill":  "{min}% Chance To Cast Level {max} {param} On Level Up",

			// Attack skill proc
			"att-skill":      "{min}% Chance To Cast Level {max} {param} On Attack",

			// Prevent monster heal
			"noheal":         "Prevent Monster Heal",

			// Durability
			"dur": "+{value} To Maximum Durability",

			// Stamina
			"stamdrain": "+{value}% Slower Stamina Drain",

			// Additional
			"addxp": "+{value}% To Experience Gained",
		},
		// Skill tab names - indexed by tab number from D2 data
		skillTabs: map[int]string{
			// Amazon (tabs 0-2)
			0: "Bow and Crossbow Skills",
			1: "Passive and Magic Skills",
			2: "Javelin and Spear Skills",
			// Sorceress (tabs 3-5)
			3: "Fire Skills",
			4: "Lightning Skills",
			5: "Cold Skills",
			// Necromancer (tabs 6-8)
			6:  "Curses",
			7:  "Poison and Bone Skills",
			8:  "Summoning Skills",
			// Paladin (tabs 9-11)
			9:  "Combat Skills",
			10: "Offensive Auras",
			11: "Defensive Auras",
			// Barbarian (tabs 12-14)
			12: "Combat Skills",
			13: "Masteries",
			14: "Warcries",
			// Druid (tabs 15-17)
			15: "Summoning Skills",
			16: "Shape Shifting Skills",
			17: "Elemental Skills",
			// Assassin (tabs 18-20)
			18: "Traps",
			19: "Shadow Disciplines",
			20: "Martial Arts",
		},
	}
}

// Translate converts a property to human-readable text
func (t *PropertyTranslator) Translate(prop Property) string {
	format, ok := t.formats[prop.Code]
	if !ok {
		// Fallback: return code with values
		if prop.Min == prop.Max {
			return fmt.Sprintf("%s: %d", prop.Code, prop.Min)
		}
		// Ensure smaller value comes first in range display
		minVal, maxVal := prop.Min, prop.Max
		if minVal > maxVal {
			minVal, maxVal = maxVal, minVal
		}
		return fmt.Sprintf("%s: %d-%d", prop.Code, minVal, maxVal)
	}

	// Replace placeholders
	result := format

	// Handle value placeholder
	if prop.Min == prop.Max {
		valueStr := fmt.Sprintf("%d", prop.Min)
		// If format has "+{value}" and value is negative, remove the "+"
		if prop.Min < 0 && strings.Contains(result, "+{value}") {
			result = strings.ReplaceAll(result, "+{value}", valueStr)
		} else {
			result = strings.ReplaceAll(result, "{value}", valueStr)
		}
	} else {
		minVal, maxVal := prop.Min, prop.Max
		if minVal > maxVal {
			minVal, maxVal = maxVal, minVal
		}

		var valueStr string
		if minVal < 0 && maxVal < 0 {
			// Both negative: show as -(absSmall-absLarge)
			absSmall := -maxVal // closer to zero = smaller absolute
			absLarge := -minVal // further from zero = larger absolute
			valueStr = fmt.Sprintf("-(%d-%d)", absSmall, absLarge)
		} else {
			valueStr = fmt.Sprintf("%d-%d", minVal, maxVal)
		}

		if minVal < 0 && strings.Contains(result, "+{value}") {
			result = strings.ReplaceAll(result, "+{value}", valueStr)
		} else {
			result = strings.ReplaceAll(result, "{value}", valueStr)
		}
	}

	result = strings.ReplaceAll(result, "{min}", fmt.Sprintf("%d", prop.Min))
	result = strings.ReplaceAll(result, "{max}", fmt.Sprintf("%d", prop.Max))

	// Handle skill tab placeholder
	if strings.Contains(result, "{skilltab}") && prop.Param != "" {
		tabNum := 0
		fmt.Sscanf(prop.Param, "%d", &tabNum)
		if tabName, ok := t.skillTabs[tabNum]; ok {
			result = strings.ReplaceAll(result, "{skilltab}", tabName)
		} else {
			result = strings.ReplaceAll(result, "{skilltab}", prop.Param)
		}
	}

	if prop.Param != "" {
		result = strings.ReplaceAll(result, "{param}", prop.Param)
	}

	return result
}

// TranslateProperties converts multiple properties to human-readable text
// Deduplicates properties that resolve to the same display text
func (t *PropertyTranslator) TranslateProperties(props []Property) []string {
	var results []string
	seen := make(map[string]bool)
	for _, prop := range props {
		text := t.Translate(prop)
		if !seen[text] {
			seen[text] = true
			results = append(results, text)
		}
	}
	return results
}

// HasRange returns true if the property has a range of values
func (t *PropertyTranslator) HasRange(prop Property) bool {
	return prop.Min != prop.Max
}

// GetDisplayName returns a formatted property name for filtering UI
func (t *PropertyTranslator) GetDisplayName(code string) string {
	// Map codes to simple display names for filter dropdowns
	names := map[string]string{
		// Stats
		"allskills":    "All Skills",
		"str":          "Strength",
		"dex":          "Dexterity",
		"vit":          "Vitality",
		"enr":          "Energy",
		"hp":           "Life",
		"mana":         "Mana",
		"all-stats":    "All Attributes",

		// Resistances
		"res-fire":     "Fire Resistance",
		"res-cold":     "Cold Resistance",
		"res-ltng":     "Lightning Resistance",
		"res-pois":     "Poison Resistance",
		"res-all":      "All Resistances",
		"res-mag":      "Magic Resistance",

		// Damage
		"dmg%":         "Enhanced Damage",
		"dmg":          "Damage",
		"dmg-min":      "Minimum Damage",
		"dmg-max":      "Maximum Damage",
		"ltng-min":     "Minimum Lightning Damage",
		"ltng-max":     "Maximum Lightning Damage",
		"fire-min":     "Minimum Fire Damage",
		"fire-max":     "Maximum Fire Damage",
		"cold-min":     "Minimum Cold Damage",
		"cold-max":     "Maximum Cold Damage",

		// Defense
		"ac%":          "Enhanced Defense",
		"ac":           "Defense",
		"red-dmg":      "Damage Reduced",
		"red-dmg%":     "Damage Reduced %",
		"red-mag":      "Magic Damage Reduced",

		// Speed
		"cast1":        "Faster Cast Rate",
		"cast2":        "Faster Cast Rate",
		"cast3":        "Faster Cast Rate",
		"swing1":       "Increased Attack Speed",
		"swing2":       "Increased Attack Speed",
		"swing3":       "Increased Attack Speed",
		"move1":        "Faster Run/Walk",
		"move2":        "Faster Run/Walk",
		"move3":        "Faster Run/Walk",
		"balance1":     "Faster Hit Recovery",
		"balance2":     "Faster Hit Recovery",
		"balance3":     "Faster Hit Recovery",
		"block1":       "Faster Block Rate",
		"block2":       "Faster Block Rate",
		"block3":       "Faster Block Rate",

		// MF/GF
		"mag%":         "Magic Find",
		"gold%":        "Extra Gold",

		// Leech
		"lifesteal":    "Life Steal",
		"manasteal":    "Mana Steal",

		// Physical bonuses
		"crush":        "Crushing Blow",
		"deadly":       "Deadly Strike",
		"openwounds":   "Open Wounds",

		// Pierce
		"pierce-fire":  "Fire Pierce",
		"pierce-cold":  "Cold Pierce",
		"pierce-ltng":  "Lightning Pierce",
		"pierce-pois":  "Poison Pierce",

		// Sunder charms
		"pierce-immunity-cold":   "Sundered Cold Immunity",
		"pierce-immunity-fire":   "Sundered Fire Immunity",
		"pierce-immunity-light":  "Sundered Lightning Immunity",
		"pierce-immunity-poison": "Sundered Poison Immunity",
		"pierce-immunity-damage": "Sundered Physical Immunity",
		"pierce-immunity-magic":  "Sundered Magic Immunity",

		// Absorb
		"abs-fire":     "Fire Absorb",
		"abs-cold":     "Cold Absorb",
		"abs-ltng":     "Lightning Absorb",
		"abs-fire%":    "Fire Absorb %",
		"abs-cold%":    "Cold Absorb %",
		"abs-ltng%":    "Lightning Absorb %",
	}

	if name, ok := names[code]; ok {
		return name
	}
	return code
}

// fixedValueCodes are properties where min/max are not item roll ranges
// - Skill procs: min = chance%, max = skill level
// - Damage adds: min-max is the damage range per hit, not a variable roll
var fixedValueCodes = map[string]bool{
	// Skill procs (chance% and skill level)
	"hit-skill":     true,
	"gethit-skill":  true,
	"kill-skill":    true,
	"death-skill":   true,
	"levelup-skill": true,
	"att-skill":     true,
	// Charged skills (min = skill level, max = charges count)
	"charged":       true,
	// Damage ranges (fixed damage per hit)
	"dmg-norm": true,
	"dmg-fire": true,
	"dmg-cold": true,
	"dmg-ltng": true,
	"dmg-mag":  true,
	"dmg-pois": true,
	// Sunder Charms (fixed text, no variable values)
	"pierce-immunity-cold":   true,
	"pierce-immunity-fire":   true,
	"pierce-immunity-light":  true,
	"pierce-immunity-poison": true,
	"pierce-immunity-damage": true,
	"pierce-immunity-magic":  true,
}

// EnrichProperty adds DisplayText and HasRange to a property
func (t *PropertyTranslator) EnrichProperty(prop *Property) {
	prop.DisplayText = t.Translate(*prop)
	// Fixed value codes use min/max for specific purposes, not as item roll ranges
	if fixedValueCodes[prop.Code] {
		prop.HasRange = false
	} else {
		prop.HasRange = prop.Min != prop.Max
	}
}

// EnrichProperties enriches all properties in a slice and returns a new slice
func (t *PropertyTranslator) EnrichProperties(props []Property) []Property {
	enriched := make([]Property, len(props))
	for i := range props {
		enriched[i] = props[i]
		t.EnrichProperty(&enriched[i])
	}
	return enriched
}

// DefaultTranslator is the global translator instance
var DefaultTranslator = NewPropertyTranslator()
