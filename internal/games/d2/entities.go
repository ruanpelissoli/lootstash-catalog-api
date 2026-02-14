package d2

import (
	"time"
)

// Profile represents a user profile for admin access
type Profile struct {
	ID        string    `json:"id"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Class represents a character class with skill trees
type Class struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	SkillSuffix string      `json:"skill_suffix"`
	SkillTrees  []SkillTree `json:"skill_trees"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// SkillTree represents a skill tree within a character class
type SkillTree struct {
	Name   string   `json:"name"`
	Skills []string `json:"skills"`
}

// Stat represents a stat code in the dynamic registry
type Stat struct {
	ID           int       `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	DisplayText  string    `json:"display_text"`
	Category     string    `json:"category"`
	IsVariable   bool      `json:"is_variable"`
	IsParametric bool      `json:"is_parametric"`
	Aliases      []string  `json:"aliases,omitempty"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Property represents a single item property/modifier
type Property struct {
	Code        string `json:"code"`
	Param       string `json:"param,omitempty"`
	Min         int    `json:"min"`
	Max         int    `json:"max"`
	DisplayText string `json:"displayText,omitempty"`
	HasRange    bool   `json:"hasRange,omitempty"`
}

// ItemType represents an item type/category
type ItemType struct {
	ID                   int       `json:"id"`
	Code                 string    `json:"code"`
	Name                 string    `json:"name"`
	Equiv1               string    `json:"equiv1,omitempty"`
	Equiv2               string    `json:"equiv2,omitempty"`
	BodyLoc1             string    `json:"body_loc1,omitempty"`
	BodyLoc2             string    `json:"body_loc2,omitempty"`
	CanBeMagic           bool      `json:"can_be_magic"`
	CanBeRare            bool      `json:"can_be_rare"`
	MaxSocketsNormal     int       `json:"max_sockets_normal"`
	MaxSocketsNightmare  int       `json:"max_sockets_nightmare"`
	MaxSocketsHell       int       `json:"max_sockets_hell"`
	StaffMods            string    `json:"staff_mods,omitempty"`
	ClassRestriction     string    `json:"class_restriction,omitempty"`
	StorePage            string    `json:"store_page,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// ItemBase represents a base item (armor, weapon, or miscellaneous)
type ItemBase struct {
	ID              int       `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	ItemType        string    `json:"item_type"`
	ItemType2       string    `json:"item_type2,omitempty"`
	Category        string    `json:"category"` // armor, weapon, misc
	Tier            string    `json:"tier,omitempty"`
	TypeTags        []string  `json:"type_tags,omitempty"`
	ClassSpecific   string    `json:"class_specific,omitempty"`
	Tradable        bool      `json:"tradable"`

	// Requirements and stats
	Level      int `json:"level"`
	LevelReq   int `json:"level_req"`
	StrReq     int `json:"str_req"`
	DexReq     int `json:"dex_req"`
	Durability int `json:"durability"`

	// Armor specific
	MinAC int `json:"min_ac"`
	MaxAC int `json:"max_ac"`

	// Weapon specific
	MinDam        int `json:"min_dam"`
	MaxDam        int `json:"max_dam"`
	TwoHandMinDam int `json:"two_hand_min_dam"`
	TwoHandMaxDam int `json:"two_hand_max_dam"`
	RangeAdder    int `json:"range_adder"`
	Speed         int `json:"speed"`
	StrBonus      int `json:"str_bonus"`
	DexBonus      int `json:"dex_bonus"`

	// Sockets
	MaxSockets   int `json:"max_sockets"`
	GemApplyType int `json:"gem_apply_type"`

	// Quality tiers
	NormalCode      string `json:"normal_code,omitempty"`
	ExceptionalCode string `json:"exceptional_code,omitempty"`
	EliteCode       string `json:"elite_code,omitempty"`

	// Inventory
	InvWidth  int `json:"inv_width"`
	InvHeight int `json:"inv_height"`

	// Graphics
	InvFile       string `json:"inv_file,omitempty"`
	FlippyFile    string `json:"flippy_file,omitempty"`
	UniqueInvFile string `json:"unique_inv_file,omitempty"`
	SetInvFile    string `json:"set_inv_file,omitempty"`

	// Image URL for Supabase storage
	ImageURL     string   `json:"image_url,omitempty"`
	IconVariants []string `json:"icon_variants,omitempty"`

	// Description for quest items
	Description string `json:"description,omitempty"`

	// Flags
	Spawnable bool `json:"spawnable"`
	Stackable bool `json:"stackable"`
	Useable   bool `json:"useable"`
	Throwable bool `json:"throwable"`
	QuestItem bool `json:"quest_item"`

	Rarity int `json:"rarity"`
	Cost   int `json:"cost"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UniqueItem represents a unique item
type UniqueItem struct {
	ID       int    `json:"id"`
	IndexID  int    `json:"index_id"`
	Name     string `json:"name"`
	BaseCode string `json:"base_code"`
	BaseName string `json:"base_name,omitempty"`

	Level    int `json:"level"`
	LevelReq int `json:"level_req"`
	Rarity   int `json:"rarity"`

	Enabled           bool `json:"enabled"`
	LadderOnly        bool `json:"ladder_only"`
	FirstLadderSeason *int `json:"first_ladder_season,omitempty"`
	LastLadderSeason  *int `json:"last_ladder_season,omitempty"`

	Properties []Property `json:"properties"`

	// Graphics
	InvTransform string `json:"inv_transform,omitempty"`
	ChrTransform string `json:"chr_transform,omitempty"`
	InvFile      string `json:"inv_file,omitempty"`

	ImageURL string `json:"image_url,omitempty"`

	CostMult int `json:"cost_mult"`
	CostAdd  int `json:"cost_add"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SetBonus represents a complete set definition with bonuses
type SetBonus struct {
	ID      int    `json:"id"`
	IndexID int    `json:"index_id"`
	Name    string `json:"name"`
	Version int    `json:"version"`

	PartialBonuses []Property `json:"partial_bonuses"` // 2-4 items bonuses
	FullBonuses    []Property `json:"full_bonuses"`    // Full set bonuses

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SetItem represents an individual set item
type SetItem struct {
	ID       int    `json:"id"`
	IndexID  int    `json:"index_id"`
	Name     string `json:"name"`
	SetName  string `json:"set_name"`
	BaseCode string `json:"base_code"`
	BaseName string `json:"base_name,omitempty"`

	Level    int `json:"level"`
	LevelReq int `json:"level_req"`
	Rarity   int `json:"rarity"`

	Properties      []Property `json:"properties"`       // Always active
	BonusProperties []Property `json:"bonus_properties"` // Partial set bonuses

	// Graphics
	InvTransform string `json:"inv_transform,omitempty"`
	ChrTransform string `json:"chr_transform,omitempty"`
	InvFile      string `json:"inv_file,omitempty"`

	ImageURL string `json:"image_url,omitempty"`

	CostMult int `json:"cost_mult"`
	CostAdd  int `json:"cost_add"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Runeword represents a runeword recipe
type Runeword struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`

	Complete          bool `json:"complete"`
	LadderOnly        bool `json:"ladder_only"`
	FirstLadderSeason *int `json:"first_ladder_season,omitempty"`
	LastLadderSeason  *int `json:"last_ladder_season,omitempty"`

	ValidItemTypes    []string `json:"valid_item_types"`
	ExcludedItemTypes []string `json:"excluded_item_types,omitempty"`

	Runes      []string   `json:"runes"`
	Properties []Property `json:"properties"`

	ImageURL string `json:"image_url,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Rune represents an individual rune item
type Rune struct {
	ID         int    `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	RuneNumber int    `json:"rune_number"`

	Level    int `json:"level"`
	LevelReq int `json:"level_req"`

	WeaponMods []Property `json:"weapon_mods"`
	HelmMods   []Property `json:"helm_mods"`
	ShieldMods []Property `json:"shield_mods"`

	InvFile string `json:"inv_file,omitempty"`

	ImageURL string `json:"image_url,omitempty"`

	Cost int `json:"cost"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Gem represents a gem item
type Gem struct {
	ID      int    `json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	GemType string `json:"gem_type"` // amethyst, sapphire, emerald, ruby, diamond, topaz, skull
	Quality string `json:"quality"`  // chipped, flawed, normal, flawless, perfect

	WeaponMods []Property `json:"weapon_mods"`
	HelmMods   []Property `json:"helm_mods"`
	ShieldMods []Property `json:"shield_mods"`

	Transform int    `json:"transform"`
	InvFile   string `json:"inv_file,omitempty"`

	ImageURL string `json:"image_url,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RunewordBase represents a pre-computed valid base item for a runeword
type RunewordBase struct {
	ID              int       `json:"id"`
	RunewordID      int       `json:"runeword_id"`
	ItemBaseID      int       `json:"item_base_id"`
	ItemBaseCode    string    `json:"item_base_code"`
	ItemBaseName    string    `json:"item_base_name"`
	Category        string    `json:"category"`
	MaxSockets      int       `json:"max_sockets"`
	RequiredSockets int       `json:"required_sockets"`
	CreatedAt       time.Time `json:"created_at"`
}

// ImportStats tracks import statistics
type ImportStats struct {
	Imported int
	Skipped  int
}

// ImportResult holds all import statistics
type ImportResult struct {
	ItemTypes      ImportStats
	ItemBases      ImportStats
	UniqueItems    ImportStats
	SetBonuses     ImportStats
	SetItems       ImportStats
	Runewords      ImportStats
	Runes          ImportStats
	Gems           ImportStats
	RunewordBases  ImportStats
	Stats          ImportStats
	ImagesUploaded int
	ImagesMissing  int
}
