package dto

// ItemSearchResult represents a single item in search autocomplete results
type ItemSearchResult struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`     // "unique", "set", "runeword", "rune", "base", "gem"
	Category string `json:"category"` // "Helms", "Armor", "Weapons", etc.
	ImageURL string `json:"imageUrl,omitempty"`
	BaseName string `json:"baseName,omitempty"` // For uniques/sets: "Shako", "Diadem", etc.
}

// SearchResponse wraps search results with pagination info
type SearchResponse struct {
	Items      []ItemSearchResult `json:"items"`
	TotalCount int                `json:"totalCount"`
	Query      string             `json:"query"`
}

// AffixOption represents a selectable option for an affix
type AffixOption struct {
	Value string `json:"value"` // Internal value (e.g., "amazon", "sorceress")
	Label string `json:"label"` // Display label (e.g., "Amazon", "Sorceress")
}

// D2Classes contains all Diablo 2 character classes
var D2Classes = []AffixOption{
	{Value: "amazon", Label: "Amazon"},
	{Value: "sorceress", Label: "Sorceress"},
	{Value: "necromancer", Label: "Necromancer"},
	{Value: "paladin", Label: "Paladin"},
	{Value: "barbarian", Label: "Barbarian"},
	{Value: "druid", Label: "Druid"},
	{Value: "assassin", Label: "Assassin"},
	{Value: "warlock", Label: "Warlock"},
}

// ItemAffix represents a human-readable item affix/property
type ItemAffix struct {
	Name        string        `json:"name"`              // Human readable name: "+2 To All Skills"
	DisplayName string        `json:"displayName"`       // Short name for UI inputs: "Cold Resist" (no value/%)
	Description string        `json:"description"`       // Additional context if needed
	MinValue    *int          `json:"minValue,omitempty"`
	MaxValue    *int          `json:"maxValue,omitempty"`
	HasRange    bool          `json:"hasRange"`          // true if min != max
	Code        string        `json:"code"`              // Internal code for filtering
	Options     []AffixOption `json:"options,omitempty"` // For special affixes like randclassskill
}

// ItemRequirements represents level and stat requirements
type ItemRequirements struct {
	Level    int `json:"level"`
	Strength int `json:"strength,omitempty"`
	Dexterity int `json:"dexterity,omitempty"`
}

// ItemBaseInfo represents the base item information
type ItemBaseInfo struct {
	Code       string        `json:"code"`
	Name       string        `json:"name"`
	Category   string        `json:"category"` // "armor", "weapon", "misc"
	ItemType   string        `json:"itemType"` // "helm", "body armor", etc.
	Defense    *DefenseRange `json:"defense,omitempty"`
	MinDamage  *int          `json:"minDamage,omitempty"`
	MaxDamage  *int          `json:"maxDamage,omitempty"`
	MaxSockets int           `json:"maxSockets,omitempty"`
	Durability int           `json:"durability,omitempty"`
}

// ItemQuality represents item quality flags
type ItemQuality struct {
	IsEthereal  bool `json:"isEthereal"`
	IsSuperior  bool `json:"isSuperior"`
	IsLadder    bool `json:"isLadder"`
}

// UniqueItemDetail represents a unique item with all its information
type UniqueItemDetail struct {
	ID           int              `json:"id"`
	Name         string           `json:"name"`
	Type         string           `json:"type"` // Always "unique"
	Rarity       string           `json:"rarity"` // "unique"
	Base         ItemBaseInfo     `json:"base"`
	Requirements ItemRequirements `json:"requirements"`
	Affixes      []ItemAffix      `json:"affixes"`
	LadderOnly   bool             `json:"ladderOnly"`
	ImageURL     string           `json:"imageUrl,omitempty"`
}

// SetItemDetail represents a set item with all its information
type SetItemDetail struct {
	ID              int              `json:"id"`
	Name            string           `json:"name"`
	SetName         string           `json:"setName"`
	Type            string           `json:"type"` // Always "set"
	Rarity          string           `json:"rarity"` // "set"
	Base            ItemBaseInfo     `json:"base"`
	Requirements    ItemRequirements `json:"requirements"`
	Affixes         []ItemAffix      `json:"affixes"`      // Always active
	BonusAffixes    []ItemAffix      `json:"bonusAffixes"` // Partial set bonuses
	ImageURL        string           `json:"imageUrl,omitempty"`
}

// SetBonusDetail represents a complete set with its bonuses
type SetBonusDetail struct {
	ID             int         `json:"id"`
	Name           string      `json:"name"`
	Items          []string    `json:"items"` // Names of items in the set
	PartialBonuses []ItemAffix `json:"partialBonuses"` // 2-4 items bonuses
	FullBonuses    []ItemAffix `json:"fullBonuses"`    // Complete set bonuses
}

// RunewordBaseItem represents a valid base item for a runeword
type RunewordBaseItem struct {
	ID         int    `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	MaxSockets int    `json:"maxSockets"`
}

// RunewordRune represents a rune in a runeword with display info
type RunewordRune struct {
	ID       int    `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	ImageURL string `json:"imageUrl,omitempty"`
}

// RunewordValidType represents a valid item type for a runeword
type RunewordValidType struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// RunewordDetail represents a runeword with all its information
type RunewordDetail struct {
	ID             int                 `json:"id"`
	Name           string              `json:"name"`
	DisplayName    string              `json:"displayName"` // Properly formatted name
	Type           string              `json:"type"`        // Always "runeword"
	Rarity         string              `json:"rarity"`      // "runeword"
	Runes          []RunewordRune      `json:"runes"`       // Runes with names and icons
	RuneOrder      string              `json:"runeOrder"`   // "JahIthBer"
	ValidTypes     []RunewordValidType `json:"validTypes"`  // Item types with names
	ValidBaseItems []RunewordBaseItem  `json:"validBaseItems,omitempty"` // Actual base items
	Requirements   ItemRequirements    `json:"requirements"`
	Affixes        []ItemAffix         `json:"affixes"`
	LadderOnly     bool                `json:"ladderOnly"`
	ImageURL       string              `json:"imageUrl,omitempty"`
}

// RuneDetail represents a rune with all its information
type RuneDetail struct {
	ID           int             `json:"id"`
	Code         string          `json:"code"`
	Name         string          `json:"name"`      // "Ber", "Jah", etc.
	RuneNumber   int             `json:"runeNumber"` // 1-33
	Type         string          `json:"type"`       // Always "rune"
	Rarity       string          `json:"rarity"`     // "rune"
	Requirements ItemRequirements `json:"requirements"`
	WeaponMods   []ItemAffix     `json:"weaponMods"`
	ArmorMods    []ItemAffix     `json:"armorMods"`  // Helm/Shield mods are same as armor
	ShieldMods   []ItemAffix     `json:"shieldMods"`
	ImageURL     string          `json:"imageUrl,omitempty"`
}

// GemDetail represents a gem with all its information
type GemDetail struct {
	ID         int         `json:"id"`
	Code       string      `json:"code"`
	Name       string      `json:"name"`      // "Perfect Ruby", "Flawless Sapphire"
	GemType    string      `json:"gemType"`   // "ruby", "sapphire", etc.
	Quality    string      `json:"quality"`   // "chipped", "flawed", "normal", "flawless", "perfect"
	Type       string      `json:"type"`      // Always "gem"
	Rarity     string      `json:"rarity"`    // "gem"
	WeaponMods []ItemAffix `json:"weaponMods"`
	ArmorMods  []ItemAffix `json:"armorMods"`
	ShieldMods []ItemAffix `json:"shieldMods"`
	ImageURL   string      `json:"imageUrl,omitempty"`
}

// BaseItemDetail represents a base item (armor, weapon, misc)
type BaseItemDetail struct {
	ID           int              `json:"id"`
	Code         string           `json:"code"`
	Name         string           `json:"name"`
	Type         string           `json:"type"`     // Always "base"
	Rarity       string           `json:"rarity"`   // "normal"
	Category     string           `json:"category"` // "armor", "weapon", "misc"
	ItemType     string           `json:"itemType"` // "helm", "body armor", etc.
	Requirements ItemRequirements `json:"requirements"`
	Defense      *DefenseRange    `json:"defense,omitempty"`
	Damage       *DamageRange     `json:"damage,omitempty"`
	Speed        int              `json:"speed,omitempty"`
	MaxSockets   int              `json:"maxSockets"`
	Durability   int              `json:"durability"`
	QualityTiers QualityTiers     `json:"qualityTiers,omitempty"`
	ImageURL     string           `json:"imageUrl,omitempty"`
	IconVariants []string         `json:"iconVariants,omitempty"`
}

// DefenseRange represents armor defense values
type DefenseRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// DamageRange represents weapon damage values
type DamageRange struct {
	OneHandMin int `json:"oneHandMin,omitempty"`
	OneHandMax int `json:"oneHandMax,omitempty"`
	TwoHandMin int `json:"twoHandMin,omitempty"`
	TwoHandMax int `json:"twoHandMax,omitempty"`
}

// QualityTiers represents normal/exceptional/elite versions
type QualityTiers struct {
	Normal      string `json:"normal,omitempty"`
	Exceptional string `json:"exceptional,omitempty"`
	Elite       string `json:"elite,omitempty"`
}

// UnifiedItemDetail is a wrapper that can contain any item type
// This is what the UI receives for a specific item
type UnifiedItemDetail struct {
	ItemType string `json:"itemType"` // "unique", "set", "runeword", "rune", "gem", "base"

	// Only one of these will be populated based on ItemType
	Unique   *UniqueItemDetail  `json:"unique,omitempty"`
	SetItem  *SetItemDetail     `json:"setItem,omitempty"`
	Runeword *RunewordDetail    `json:"runeword,omitempty"`
	Rune     *RuneDetail        `json:"rune,omitempty"`
	Gem      *GemDetail         `json:"gem,omitempty"`
	Base     *BaseItemDetail    `json:"base,omitempty"`
}

// AffixFilter represents a filter for affix values (for marketplace future use)
type AffixFilter struct {
	Code     string `json:"code"`
	MinValue *int   `json:"minValue,omitempty"`
	MaxValue *int   `json:"maxValue,omitempty"`
}

// MarketplaceFilters represents all available filters (placeholder for future)
type MarketplaceFilters struct {
	// Fixed filters
	Ladder   *bool  `json:"ladder,omitempty"`   // true = ladder only, false = non-ladder only, nil = both
	Hardcore *bool  `json:"hardcore,omitempty"` // true = hardcore, false = softcore, nil = both
	Platform string `json:"platform,omitempty"` // "pc", "xbox", "playstation", "switch", "" = all

	// Item type filters
	ItemTypes  []string `json:"itemTypes,omitempty"`  // ["unique", "set", "runeword"]
	Categories []string `json:"categories,omitempty"` // ["helm", "armor", "weapon"]

	// Affix value filters
	AffixFilters []AffixFilter `json:"affixFilters,omitempty"`

	// Price/asking filters (for marketplace)
	AskingForItems []string `json:"askingForItems,omitempty"` // ["Ist", "Ber"] - filter by what sellers want
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// StatCode represents a filterable stat code for marketplace filtering
type StatCode struct {
	Code        string   `json:"code"`                  // Internal code for filtering (e.g., "mf", "fcr", "res-fire")
	Name        string   `json:"name"`                  // Short display name (e.g., "Magic Find", "Faster Cast Rate")
	Description string   `json:"description"`           // Format template (e.g., "+{value}% Better Chance Of Getting Magic Items")
	Category    string   `json:"category"`              // Category for grouping in UI (e.g., "Speed", "Resistances", "Damage")
	Aliases     []string `json:"aliases,omitempty"`     // Alternative codes that map to this stat
	IsVariable  bool     `json:"isVariable"`            // Whether this stat typically has variable rolls on items
}

// Category represents an item category for filtering
type Category struct {
	Code        string `json:"code"`                  // Internal code for filtering (e.g., "helm", "armor", "weapon")
	Name        string `json:"name"`                  // Display name (e.g., "Helms", "Body Armor", "Weapons")
	Description string `json:"description,omitempty"` // Brief description of this category
}

// Rarity represents an item rarity for filtering
type Rarity struct {
	Code        string `json:"code"`        // Internal code for filtering (e.g., "unique", "set", "runeword")
	Name        string `json:"name"`        // Display name (e.g., "Unique", "Set", "Runeword")
	Color       string `json:"color"`       // Hex color for UI display (e.g., "#C4A000" for unique gold)
	Description string `json:"description"` // Brief description of this rarity type
}
