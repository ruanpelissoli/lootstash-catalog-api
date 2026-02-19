package d2

// SubcategoryInfo contains metadata about an item subcategory
type SubcategoryInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CategoryInfo contains metadata about an item category
type CategoryInfo struct {
	Code          string            `json:"code"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Subcategories []SubcategoryInfo `json:"subcategories,omitempty"`
}

// RarityInfo contains metadata about an item rarity
type RarityInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Color       string `json:"color"`       // Hex color for UI display
	Description string `json:"description"` // Brief description of this rarity type
}

// Categories returns all item categories for Diablo 2 in a hierarchical structure
func Categories() []CategoryInfo {
	return []CategoryInfo{
		{
			Code:        "armor",
			Name:        "Armor",
			Description: "Protective gear including helms, body armor, gloves, boots, and belts",
			Subcategories: []SubcategoryInfo{
				{Code: "helms", Name: "Helms"},
				{Code: "circlets", Name: "Circlets"},
				{Code: "barbarian helms", Name: "Barbarian Helms"},
				{Code: "druid pelts", Name: "Druid Pelts"},
				{Code: "body armor", Name: "Body Armor"},
				{Code: "gloves", Name: "Gloves"},
				{Code: "boots", Name: "Boots"},
				{Code: "belts", Name: "Belts"},
			},
		},
		{
			Code:        "weapons",
			Name:        "Weapons",
			Description: "All weapon types including melee, ranged, caster weapons, and shields",
			Subcategories: []SubcategoryInfo{
				{Code: "swords", Name: "Swords"},
				{Code: "axes", Name: "Axes"},
				{Code: "maces", Name: "Maces"},
				{Code: "hammers", Name: "Hammers"},
				{Code: "clubs", Name: "Clubs"},
				{Code: "polearms", Name: "Polearms"},
				{Code: "staves", Name: "Staves"},
				{Code: "scepters", Name: "Scepters"},
				{Code: "wands", Name: "Wands"},
				{Code: "bows", Name: "Bows"},
				{Code: "crossbows", Name: "Crossbows"},
				{Code: "daggers", Name: "Daggers"},
				{Code: "throwing", Name: "Throwing"},
				{Code: "javelins", Name: "Javelins"},
				{Code: "spears", Name: "Spears"},
				{Code: "claws", Name: "Claws"},
				{Code: "orbs", Name: "Orbs"},
				{Code: "grimoires", Name: "Grimoires"},
				{Code: "shields", Name: "Shields"},
				{Code: "paladin shields", Name: "Paladin Shields"},
				{Code: "necromancer shields", Name: "Necromancer Shields"},
			},
		},
		{
			Code:        "jewelry",
			Name:        "Jewelry",
			Description: "Rings, amulets, and jewels",
			Subcategories: []SubcategoryInfo{
				{Code: "rings", Name: "Rings"},
				{Code: "amulets", Name: "Amulets"},
				{Code: "jewels", Name: "Jewels"},
			},
		},
		{
			Code:        "charms",
			Name:        "Charms",
			Description: "Inventory charms that provide passive bonuses",
			Subcategories: []SubcategoryInfo{
				{Code: "small charm", Name: "Small Charms"},
				{Code: "large charm", Name: "Large Charms"},
				{Code: "grand charm", Name: "Grand Charms"},
			},
		},
		{
			Code:        "runes",
			Name:        "Runes",
			Description: "Socketable runes used to create runewords",
		},
		{
			Code:        "gems",
			Name:        "Gems",
			Description: "Socketable gems from chipped to perfect quality",
		},
		{
			Code:        "misc",
			Name:        "Miscellaneous",
			Description: "Keys, essences, tokens, and other items",
			Subcategories: []SubcategoryInfo{
				{Code: "keys", Name: "Keys"},
				{Code: "essences", Name: "Essences"},
				{Code: "tokens", Name: "Tokens"},
				{Code: "quest items", Name: "Quest Items"},
			},
		},
	}
}

// Rarities returns all item rarities for Diablo 2
func Rarities() []RarityInfo {
	return []RarityInfo{
		{Code: "normal", Name: "Normal", Color: "#FFFFFF", Description: "White items with no magical properties"},
		{Code: "magic", Name: "Magic", Color: "#4169E1", Description: "Blue items with 1-2 magical affixes"},
		{Code: "rare", Name: "Rare", Color: "#FFFF00", Description: "Yellow items with 2-6 magical affixes"},
		{Code: "unique", Name: "Unique", Color: "#C4A000", Description: "Gold/tan items with fixed properties"},
		{Code: "set", Name: "Set", Color: "#00FF00", Description: "Green items that grant bonuses when worn together"},
		{Code: "runeword", Name: "Runeword", Color: "#C4A000", Description: "Items created by socketing specific runes in order"},
		{Code: "crafted", Name: "Crafted", Color: "#FFA500", Description: "Orange items created via Horadric Cube recipes"},
	}
}
