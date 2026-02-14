package d2

import "strings"

// htmlTypeNameToCode maps HTML type display names to D2 item type codes
var htmlTypeNameToCode = map[string]string{
	"Body Armor":           "tors",
	"Helms":                "helm",
	"Shields":              "shie",
	"Swords":               "swor",
	"Axes":                 "axe",
	"Maces":                "mace",
	"Polearms":             "pole",
	"Staves":               "staf",
	"Scepters":             "scep",
	"Wands":                "wand",
	"Bows":                 "bow",
	"Crossbows":            "xbow",
	"Daggers":              "knif",
	"Throwing":             "tkni",
	"Javelins":             "jave",
	"Spears":               "spea",
	"Claws":                "h2h",
	"Orbs":                 "orb",
	"Amazon Weapons":       "amaz",
	"Hammers":              "hamm",
	"Clubs":                "club",
	"Weapons":              "weap",
	"Missile Weapons":      "miss",
	"Melee Weapons":        "mele",
	"Gloves":               "glov",
	"Boots":                "boot",
	"Belts":                "belt",
	"Circlets":             "circ",
	"Druid Pelts":          "pelt",
	"Barbarian Helms":      "phlm",
	"Necromancer Shields":  "head",
	"Shrunken Heads":       "head",
	"Paladin Shields":      "ashd",
	"Targes":               "ashd",
	"Grimoires":            "grim",
	"Katars":               "h2h",
	"Wand":                 "wand",
	"Armor":                "tors",
	"All Weapons":          "weap",
	"All Armor":            "armo",
	"2 socket Weapons":     "weap",
	"3 socket Weapons":     "weap",
	"4 socket Weapons":     "weap",
	"5 socket Weapons":     "weap",
	"6 socket Weapons":     "weap",
	"2 socket Shields":     "shie",
	"3 socket Shields":     "shie",
	"4 socket Shields":     "shie",
	"2 socket Swords":      "swor",
	"3 socket Swords":      "swor",
	"4 socket Swords":      "swor",
	"5 socket Swords":      "swor",
	"6 socket Swords":      "swor",
	"2 socket Body Armor":  "tors",
	"3 socket Body Armor":  "tors",
	"4 socket Body Armor":  "tors",
	"2 socket Armor":       "tors",
	"3 socket Armor":       "tors",
	"4 socket Armor":       "tors",
	"2 socket Helms":       "helm",
	"3 socket Helms":       "helm",
	"4 socket Helms":       "helm",
}

// generateBaseCode creates a short code from an item name for items without an explicit code.
func generateBaseCode(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "-", "")
	words := strings.Fields(name)
	if len(words) == 1 {
		w := words[0]
		if len(w) > 4 {
			return w[:4]
		}
		return w
	}
	code := ""
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		take := 2
		if i == 0 {
			take = 3
			if len(w) < 3 {
				take = len(w)
			}
		} else if len(w) < 2 {
			take = len(w)
		}
		code += w[:take]
		if len(code) >= 8 {
			break
		}
	}
	if len(code) > 8 {
		code = code[:8]
	}
	return code
}

// splitOrBonuses splits stat text that contains "or" alternatives into separate bonus lines.
func splitOrBonuses(text string) []string {
	text = strings.TrimSpace(text)
	parts := strings.Split(text, " or \n")
	if len(parts) > 1 {
		var result []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	}
	return []string{text}
}
