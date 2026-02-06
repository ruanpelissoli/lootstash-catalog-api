package d2

// CategoryInfo contains metadata about an item category
type CategoryInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// RarityInfo contains metadata about an item rarity
type RarityInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Color       string `json:"color"`       // Hex color for UI display
	Description string `json:"description"` // Brief description of this rarity type
}

// Categories returns all item categories for Diablo 2
func Categories() []CategoryInfo {
	return []CategoryInfo{
		{Code: "helm", Name: "Helms", Description: "Head armor including circlets, crowns, and helmets"},
		{Code: "armor", Name: "Body Armor", Description: "Chest armor including robes, plate, and leather"},
		{Code: "weapon", Name: "Weapons", Description: "All weapon types including swords, axes, bows, and staves"},
		{Code: "shield", Name: "Shields", Description: "Shields and paladin-specific shields"},
		{Code: "gloves", Name: "Gloves", Description: "Hand armor including gauntlets and bracers"},
		{Code: "boots", Name: "Boots", Description: "Foot armor including greaves and boots"},
		{Code: "belt", Name: "Belts", Description: "Waist armor including sashes and belts"},
		{Code: "amulet", Name: "Amulets", Description: "Neck jewelry"},
		{Code: "ring", Name: "Rings", Description: "Finger jewelry"},
		{Code: "charm", Name: "Charms", Description: "Inventory charms (small, large, grand)"},
		{Code: "jewel", Name: "Jewels", Description: "Socketable jewels with random magical properties"},
		{Code: "rune", Name: "Runes", Description: "Socketable runes used to create runewords"},
		{Code: "gem", Name: "Gems", Description: "Socketable gems from chipped to perfect quality"},
		{Code: "misc", Name: "Miscellaneous", Description: "Keys, organs, tokens, and other items"},
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
