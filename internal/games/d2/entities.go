package d2

import (
	"time"
)

// Profile represents a user profile for admin access
type Profile struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
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

// ItemProperty represents a property definition
type ItemProperty struct {
	ID      int    `json:"id"`
	Code    string `json:"code"`
	Enabled bool   `json:"enabled"`

	Func1 *int   `json:"func1,omitempty"`
	Stat1 string `json:"stat1,omitempty"`
	Func2 *int   `json:"func2,omitempty"`
	Stat2 string `json:"stat2,omitempty"`
	Func3 *int   `json:"func3,omitempty"`
	Stat3 string `json:"stat3,omitempty"`
	Func4 *int   `json:"func4,omitempty"`
	Stat4 string `json:"stat4,omitempty"`
	Func5 *int   `json:"func5,omitempty"`
	Stat5 string `json:"stat5,omitempty"`
	Func6 *int   `json:"func6,omitempty"`
	Stat6 string `json:"stat6,omitempty"`
	Func7 *int   `json:"func7,omitempty"`
	Stat7 string `json:"stat7,omitempty"`

	Tooltip string `json:"tooltip,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Affix represents a magic prefix or suffix
type Affix struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	AffixType string `json:"affix_type"` // prefix or suffix

	Version   int  `json:"version"`
	Spawnable bool `json:"spawnable"`
	Rare      bool `json:"rare"`

	Level     int  `json:"level"`
	MaxLevel  *int `json:"max_level,omitempty"`
	LevelReq  int  `json:"level_req"`

	ClassSpecific string `json:"class_specific,omitempty"`
	ClassLevelReq int    `json:"class_level_req"`

	Frequency   int `json:"frequency"`
	AffixGroup  int `json:"affix_group"`

	Mod1Code  string `json:"mod1_code,omitempty"`
	Mod1Param string `json:"mod1_param,omitempty"`
	Mod1Min   int    `json:"mod1_min"`
	Mod1Max   int    `json:"mod1_max"`

	Mod2Code  string `json:"mod2_code,omitempty"`
	Mod2Param string `json:"mod2_param,omitempty"`
	Mod2Min   int    `json:"mod2_min"`
	Mod2Max   int    `json:"mod2_max"`

	Mod3Code  string `json:"mod3_code,omitempty"`
	Mod3Param string `json:"mod3_param,omitempty"`
	Mod3Min   int    `json:"mod3_min"`
	Mod3Max   int    `json:"mod3_max"`

	ValidItemTypes    []string `json:"valid_item_types"`
	ExcludedItemTypes []string `json:"excluded_item_types,omitempty"`

	TransformColor string `json:"transform_color,omitempty"`

	Multiply int `json:"multiply"`
	AddCost  int `json:"add_cost"`

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

// TreasureClass represents a drop table/loot pool
type TreasureClass struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	GroupID           *int      `json:"group_id,omitempty"`
	Level             int       `json:"level"`
	Picks             int       `json:"picks"`
	UniqueMod         int       `json:"unique_mod"`
	SetMod            int       `json:"set_mod"`
	RareMod           int       `json:"rare_mod"`
	MagicMod          int       `json:"magic_mod"`
	NoDrop            int       `json:"no_drop"`
	FirstLadderSeason *int      `json:"first_ladder_season,omitempty"`
	LastLadderSeason  *int      `json:"last_ladder_season,omitempty"`
	NoAlwaysSpawn     bool      `json:"no_always_spawn"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TreasureClassItem represents an item within a treasure class
type TreasureClassItem struct {
	ID              int       `json:"id"`
	TreasureClassID int       `json:"treasure_class_id"`
	Slot            int       `json:"slot"`
	ItemCode        string    `json:"item_code"`
	IsTreasureClass bool      `json:"is_treasure_class"`
	Probability     int       `json:"probability"`
	CreatedAt       time.Time `json:"created_at"`
}

// ItemRatio represents quality calculation ratios
type ItemRatio struct {
	ID               int       `json:"id"`
	FunctionName     string    `json:"function_name"`
	Version          int       `json:"version"`
	IsUber           bool      `json:"is_uber"`
	IsClassSpecific  bool      `json:"is_class_specific"`
	UniqueRatio      int       `json:"unique_ratio"`
	UniqueDivisor    int       `json:"unique_divisor"`
	UniqueMin        int       `json:"unique_min"`
	RareRatio        int       `json:"rare_ratio"`
	RareDivisor      int       `json:"rare_divisor"`
	RareMin          int       `json:"rare_min"`
	SetRatio         int       `json:"set_ratio"`
	SetDivisor       int       `json:"set_divisor"`
	SetMin           int       `json:"set_min"`
	MagicRatio       int       `json:"magic_ratio"`
	MagicDivisor     int       `json:"magic_divisor"`
	MagicMin         int       `json:"magic_min"`
	HiQualityRatio   int       `json:"hiquality_ratio"`
	HiQualityDivisor int       `json:"hiquality_divisor"`
	NormalRatio      int       `json:"normal_ratio"`
	NormalDivisor    int       `json:"normal_divisor"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ImportStats tracks import statistics
type ImportStats struct {
	Imported int
	Skipped  int
}

// ImportResult holds all import statistics
type ImportResult struct {
	ItemTypes       ImportStats
	ItemBases       ImportStats
	UniqueItems     ImportStats
	SetBonuses      ImportStats
	SetItems        ImportStats
	Runewords       ImportStats
	Runes           ImportStats
	Gems            ImportStats
	Properties      ImportStats
	Affixes         ImportStats
	RunewordBases   ImportStats
	TreasureClasses ImportStats
	ItemRatios      ImportStats
}
