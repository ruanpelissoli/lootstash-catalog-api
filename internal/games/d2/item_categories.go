package d2

import "strings"

// CategoryPair represents a resolved category + subcategory for an item.
type CategoryPair struct {
	Category    string
	Subcategory string
}

// TypeTagToCategoryMap maps HTML/type_tag display names to (category, subcategory).
// Keys match the type tag names already present in item_bases.type_tags from HTML parsing.
var TypeTagToCategoryMap = map[string]CategoryPair{
	// Armor — head
	"Helms":           {Category: "armor", Subcategory: "helms"},
	"Circlets":        {Category: "armor", Subcategory: "circlets"},
	"Barbarian Helms": {Category: "armor", Subcategory: "barbarian helms"},
	"Druid Pelts":     {Category: "armor", Subcategory: "druid pelts"},

	// Armor — body
	"Body Armor": {Category: "armor", Subcategory: "body armor"},
	"Armor":      {Category: "armor", Subcategory: "body armor"},

	// Armor — extremities
	"Gloves": {Category: "armor", Subcategory: "gloves"},
	"Boots":  {Category: "armor", Subcategory: "boots"},
	"Belts":  {Category: "armor", Subcategory: "belts"},

	// Weapons — shields (user decided shields are weapons)
	"Shields":             {Category: "weapons", Subcategory: "shields"},
	"Paladin Shields":     {Category: "weapons", Subcategory: "paladin shields"},
	"Targes":              {Category: "weapons", Subcategory: "paladin shields"},
	"Necromancer Shields": {Category: "weapons", Subcategory: "necromancer shields"},
	"Shrunken Heads":      {Category: "weapons", Subcategory: "necromancer shields"},

	// Weapons — melee
	"Swords":   {Category: "weapons", Subcategory: "swords"},
	"Axes":     {Category: "weapons", Subcategory: "axes"},
	"Maces":    {Category: "weapons", Subcategory: "maces"},
	"Hammers":  {Category: "weapons", Subcategory: "hammers"},
	"Clubs":    {Category: "weapons", Subcategory: "clubs"},
	"Polearms": {Category: "weapons", Subcategory: "polearms"},
	"Staves":   {Category: "weapons", Subcategory: "staves"},
	"Scepters": {Category: "weapons", Subcategory: "scepters"},
	"Wands":    {Category: "weapons", Subcategory: "wands"},
	"Wand":     {Category: "weapons", Subcategory: "wands"},
	"Daggers":  {Category: "weapons", Subcategory: "daggers"},
	"Spears":   {Category: "weapons", Subcategory: "spears"},

	// Weapons — ranged
	"Bows":       {Category: "weapons", Subcategory: "bows"},
	"Crossbows":  {Category: "weapons", Subcategory: "crossbows"},
	"Throwing":   {Category: "weapons", Subcategory: "throwing"},
	"Javelins":   {Category: "weapons", Subcategory: "javelins"},

	// Weapons — class-specific
	"Claws":          {Category: "weapons", Subcategory: "claws"},
	"Katars":         {Category: "weapons", Subcategory: "claws"},
	"Orbs":           {Category: "weapons", Subcategory: "orbs"},
	"Grimoires":      {Category: "weapons", Subcategory: "grimoires"},

	// Broad fallbacks (no subcategory — only used when no specific tag matches)
	"Weapons":        {Category: "weapons", Subcategory: ""},
	"All Weapons":    {Category: "weapons", Subcategory: ""},
	"Melee Weapons":  {Category: "weapons", Subcategory: ""},
	"Missile Weapons": {Category: "weapons", Subcategory: ""},
	"All Armor":      {Category: "armor", Subcategory: ""},
}

// socketPrefixes are prefixed type tag names like "2 socket Swords" that should
// resolve to the same category as the base type name.
var socketPrefixes = []string{
	"2 socket ", "3 socket ", "4 socket ", "5 socket ", "6 socket ",
}

// ResolveCategoryFromTypeTags examines the type_tags array on a base item
// and returns the best-matching (category, subcategory).
// Prefers the first tag with a non-empty subcategory; falls back to first
// broad match. Defaults to ("misc", "").
func ResolveCategoryFromTypeTags(typeTags []string) CategoryPair {
	var fallback *CategoryPair
	for _, tag := range typeTags {
		// Direct lookup
		if pair, ok := TypeTagToCategoryMap[tag]; ok {
			if pair.Subcategory != "" {
				return pair
			}
			if fallback == nil {
				fb := pair
				fallback = &fb
			}
			continue
		}
		// Strip socket prefixes: "3 socket Swords" → "Swords"
		for _, prefix := range socketPrefixes {
			if len(tag) > len(prefix) && tag[:len(prefix)] == prefix {
				base := tag[len(prefix):]
				if pair, ok := TypeTagToCategoryMap[base]; ok {
					if pair.Subcategory != "" {
						return pair
					}
					if fallback == nil {
						fb := pair
						fallback = &fb
					}
				}
				break
			}
		}
	}
	if fallback != nil {
		return *fallback
	}
	return CategoryPair{Category: "misc", Subcategory: ""}
}

// RunewordCategoriesFromValidTypes derives a category and subcategory list for
// a runeword from its valid_item_types array.
func RunewordCategoriesFromValidTypes(validTypes []string) (string, []string) {
	if len(validTypes) == 0 {
		return "weapons", nil
	}

	categoryCounts := make(map[string]int)
	var subcategories []string
	seen := make(map[string]bool)

	for _, t := range validTypes {
		// Try direct lookup
		pair, ok := TypeTagToCategoryMap[t]
		if !ok {
			// Try stripping socket prefix
			for _, prefix := range socketPrefixes {
				if len(t) > len(prefix) && t[:len(prefix)] == prefix {
					pair, ok = TypeTagToCategoryMap[t[len(prefix):]]
					break
				}
			}
		}
		if !ok {
			continue
		}
		categoryCounts[pair.Category]++
		if pair.Subcategory != "" && !seen[pair.Subcategory] {
			subcategories = append(subcategories, pair.Subcategory)
			seen[pair.Subcategory] = true
		}
	}

	// Find dominant category
	bestCat := "weapons"
	bestCount := 0
	for cat, count := range categoryCounts {
		if count > bestCount {
			bestCat = cat
			bestCount = count
		}
	}

	return bestCat, subcategories
}

// miscCategoryPair returns the appropriate category pair for a misc item based on its name.
func miscCategoryPair(name string) CategoryPair {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "small charm"):
		return CategoryPair{"charms", "small charm"}
	case strings.Contains(lower, "large charm"):
		return CategoryPair{"charms", "large charm"}
	case strings.Contains(lower, "grand charm"):
		return CategoryPair{"charms", "grand charm"}
	case strings.Contains(lower, "charm"):
		return CategoryPair{"charms", ""}
	case strings.Contains(lower, "jewel"):
		return CategoryPair{"jewelry", "jewels"}
	case strings.Contains(lower, "key"):
		return CategoryPair{"misc", "keys"}
	case strings.Contains(lower, "essence"):
		return CategoryPair{"misc", "essences"}
	case strings.Contains(lower, "token"):
		return CategoryPair{"misc", "tokens"}
	default:
		return CategoryPair{"misc", ""}
	}
}
