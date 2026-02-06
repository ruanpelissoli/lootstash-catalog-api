package d2

// StatCodeInfo contains metadata about a stat code for filtering
type StatCodeInfo struct {
	Code        string   // Primary code used for filtering
	Name        string   // Short display name
	Description string   // Format template showing what the stat looks like
	Category    string   // Category for grouping in UI
	Aliases     []string // Alternative codes that map to this stat
	IsVariable  bool     // Whether this stat typically has variable rolls
}

// StatCategories defines the ordering of stat categories in the UI
var StatCategories = []string{
	"Skills",
	"Skill Trees",
	"Attributes",
	"Life & Mana",
	"Speed",
	"Resistances",
	"Absorb",
	"Damage",
	"Attack",
	"Leech",
	"Combat",
	"Magic Find",
	"Other",
}

// FilterableStats returns all stat codes that are useful for marketplace filtering.
// These are the stats users typically search for when looking for items.
func FilterableStats() []StatCodeInfo {
	return []StatCodeInfo{
		// Skills
		{Code: "allskills", Name: "All Skills", Description: "+{value} To All Skills", Category: "Skills", IsVariable: true},
		{Code: "ama", Name: "Amazon Skills", Description: "+{value} To Amazon Skill Levels", Category: "Skills", IsVariable: true},
		{Code: "sor", Name: "Sorceress Skills", Description: "+{value} To Sorceress Skill Levels", Category: "Skills", IsVariable: true},
		{Code: "nec", Name: "Necromancer Skills", Description: "+{value} To Necromancer Skill Levels", Category: "Skills", IsVariable: true},
		{Code: "pal", Name: "Paladin Skills", Description: "+{value} To Paladin Skill Levels", Category: "Skills", IsVariable: true},
		{Code: "bar", Name: "Barbarian Skills", Description: "+{value} To Barbarian Skill Levels", Category: "Skills", IsVariable: true},
		{Code: "dru", Name: "Druid Skills", Description: "+{value} To Druid Skill Levels", Category: "Skills", IsVariable: true},
		{Code: "ass", Name: "Assassin Skills", Description: "+{value} To Assassin Skill Levels", Category: "Skills", IsVariable: true},

		// Skill Trees - Amazon
		{Code: "ama-bow", Name: "Bow and Crossbow", Description: "+{value} To Bow and Crossbow Skills (Amazon Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "ama-passive", Name: "Passive and Magic", Description: "+{value} To Passive and Magic Skills (Amazon Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "ama-javelin", Name: "Javelin and Spear", Description: "+{value} To Javelin and Spear Skills (Amazon Only)", Category: "Skill Trees", IsVariable: true},

		// Skill Trees - Sorceress
		{Code: "sor-fire", Name: "Fire Skills", Description: "+{value} To Fire Skills (Sorceress Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "sor-lightning", Name: "Lightning Skills", Description: "+{value} To Lightning Skills (Sorceress Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "sor-cold", Name: "Cold Skills", Description: "+{value} To Cold Skills (Sorceress Only)", Category: "Skill Trees", IsVariable: true},

		// Skill Trees - Necromancer
		{Code: "nec-curses", Name: "Curses", Description: "+{value} To Curses (Necromancer Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "nec-poisonbone", Name: "Poison and Bone", Description: "+{value} To Poison and Bone Skills (Necromancer Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "nec-summon", Name: "Summoning Skills", Description: "+{value} To Summoning Skills (Necromancer Only)", Category: "Skill Trees", IsVariable: true},

		// Skill Trees - Paladin
		{Code: "pal-combat", Name: "Combat Skills", Description: "+{value} To Combat Skills (Paladin Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "pal-offensive", Name: "Offensive Auras", Description: "+{value} To Offensive Auras (Paladin Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "pal-defensive", Name: "Defensive Auras", Description: "+{value} To Defensive Auras (Paladin Only)", Category: "Skill Trees", IsVariable: true},

		// Skill Trees - Barbarian
		{Code: "bar-combat", Name: "Combat Skills", Description: "+{value} To Combat Skills (Barbarian Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "bar-masteries", Name: "Masteries", Description: "+{value} To Masteries (Barbarian Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "bar-warcries", Name: "Warcries", Description: "+{value} To Warcries (Barbarian Only)", Category: "Skill Trees", IsVariable: true},

		// Skill Trees - Druid
		{Code: "dru-summon", Name: "Summoning Skills", Description: "+{value} To Summoning Skills (Druid Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "dru-shapeshifting", Name: "Shape Shifting", Description: "+{value} To Shape Shifting Skills (Druid Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "dru-elemental", Name: "Elemental Skills", Description: "+{value} To Elemental Skills (Druid Only)", Category: "Skill Trees", IsVariable: true},

		// Skill Trees - Assassin
		{Code: "ass-traps", Name: "Traps", Description: "+{value} To Traps (Assassin Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "ass-shadow", Name: "Shadow Disciplines", Description: "+{value} To Shadow Disciplines (Assassin Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "ass-martial", Name: "Martial Arts", Description: "+{value} To Martial Arts (Assassin Only)", Category: "Skill Trees", IsVariable: true},

		// Attributes
		{Code: "str", Name: "Strength", Description: "+{value} To Strength", Category: "Attributes", IsVariable: true},
		{Code: "dex", Name: "Dexterity", Description: "+{value} To Dexterity", Category: "Attributes", IsVariable: true},
		{Code: "vit", Name: "Vitality", Description: "+{value} To Vitality", Category: "Attributes", IsVariable: true},
		{Code: "enr", Name: "Energy", Description: "+{value} To Energy", Category: "Attributes", IsVariable: true},
		{Code: "all-stats", Name: "All Attributes", Description: "+{value} To All Attributes", Category: "Attributes", IsVariable: true},

		// Life & Mana
		{Code: "hp", Name: "Life", Description: "+{value} To Life", Category: "Life & Mana", IsVariable: true},
		{Code: "mana", Name: "Mana", Description: "+{value} To Mana", Category: "Life & Mana", IsVariable: true},
		{Code: "hp%", Name: "Life %", Description: "+{value}% To Life", Category: "Life & Mana", IsVariable: true},
		{Code: "mana%", Name: "Mana %", Description: "+{value}% To Mana", Category: "Life & Mana", IsVariable: true},
		{Code: "regen-mana", Name: "Mana Regen", Description: "Regenerate Mana {value}%", Category: "Life & Mana", IsVariable: true},
		{Code: "regen", Name: "Replenish Life", Description: "Replenish Life +{value}", Category: "Life & Mana", IsVariable: true},

		// Speed - these have aliases because game data uses numbered variants
		{Code: "fcr", Name: "Faster Cast Rate", Description: "+{value}% Faster Cast Rate", Category: "Speed", Aliases: []string{"cast1", "cast2", "cast3"}, IsVariable: true},
		{Code: "ias", Name: "Increased Attack Speed", Description: "+{value}% Increased Attack Speed", Category: "Speed", Aliases: []string{"swing1", "swing2", "swing3"}, IsVariable: true},
		{Code: "frw", Name: "Faster Run/Walk", Description: "+{value}% Faster Run/Walk", Category: "Speed", Aliases: []string{"move1", "move2", "move3"}, IsVariable: true},
		{Code: "fhr", Name: "Faster Hit Recovery", Description: "+{value}% Faster Hit Recovery", Category: "Speed", Aliases: []string{"balance1", "balance2", "balance3"}, IsVariable: true},
		{Code: "block", Name: "Faster Block Rate", Description: "+{value}% Faster Block Rate", Category: "Speed", Aliases: []string{"block1", "block2", "block3"}, IsVariable: true},

		// Resistances - these have aliases for different code conventions
		{Code: "fire_res", Name: "Fire Resist", Description: "Fire Resist +{value}%", Category: "Resistances", Aliases: []string{"res-fire"}, IsVariable: true},
		{Code: "cold_res", Name: "Cold Resist", Description: "Cold Resist +{value}%", Category: "Resistances", Aliases: []string{"res-cold"}, IsVariable: true},
		{Code: "light_res", Name: "Lightning Resist", Description: "Lightning Resist +{value}%", Category: "Resistances", Aliases: []string{"res-ltng"}, IsVariable: true},
		{Code: "poison_res", Name: "Poison Resist", Description: "Poison Resist +{value}%", Category: "Resistances", Aliases: []string{"res-pois"}, IsVariable: true},
		{Code: "all_res", Name: "All Resistances", Description: "All Resistances +{value}", Category: "Resistances", Aliases: []string{"res-all"}, IsVariable: true},
		{Code: "res-mag", Name: "Magic Resist", Description: "Magic Resist +{value}%", Category: "Resistances", IsVariable: true},

		// Absorb
		{Code: "abs-fire", Name: "Fire Absorb", Description: "+{value} Fire Absorb", Category: "Absorb", IsVariable: true},
		{Code: "abs-cold", Name: "Cold Absorb", Description: "+{value} Cold Absorb", Category: "Absorb", IsVariable: true},
		{Code: "abs-ltng", Name: "Lightning Absorb", Description: "+{value} Lightning Absorb", Category: "Absorb", IsVariable: true},
		{Code: "abs-fire%", Name: "Fire Absorb %", Description: "{value}% Fire Absorb", Category: "Absorb", IsVariable: true},
		{Code: "abs-cold%", Name: "Cold Absorb %", Description: "{value}% Cold Absorb", Category: "Absorb", IsVariable: true},
		{Code: "abs-ltng%", Name: "Lightning Absorb %", Description: "{value}% Lightning Absorb", Category: "Absorb", IsVariable: true},

		// Damage
		{Code: "ed", Name: "Enhanced Damage", Description: "+{value}% Enhanced Damage", Category: "Damage", Aliases: []string{"dmg%"}, IsVariable: true},
		{Code: "dmg-min", Name: "Minimum Damage", Description: "+{value} To Minimum Damage", Category: "Damage", IsVariable: true},
		{Code: "dmg-max", Name: "Maximum Damage", Description: "+{value} To Maximum Damage", Category: "Damage", IsVariable: true},
		{Code: "dmg-demon", Name: "Damage to Demons", Description: "+{value}% Damage To Demons", Category: "Damage", IsVariable: true},
		{Code: "dmg-undead", Name: "Damage to Undead", Description: "+{value}% Damage To Undead", Category: "Damage", IsVariable: true},
		{Code: "extra-fire", Name: "Fire Skill Damage", Description: "+{value}% To Fire Skill Damage", Category: "Damage", IsVariable: true},
		{Code: "extra-cold", Name: "Cold Skill Damage", Description: "+{value}% To Cold Skill Damage", Category: "Damage", IsVariable: true},
		{Code: "extra-ltng", Name: "Lightning Skill Damage", Description: "+{value}% To Lightning Skill Damage", Category: "Damage", IsVariable: true},
		{Code: "extra-pois", Name: "Poison Skill Damage", Description: "+{value}% To Poison Skill Damage", Category: "Damage", IsVariable: true},

		// Attack
		{Code: "ar", Name: "Attack Rating", Description: "+{value} To Attack Rating", Category: "Attack", Aliases: []string{"att", "att%"}, IsVariable: true},
		{Code: "att-demon", Name: "AR vs Demons", Description: "+{value} To Attack Rating Against Demons", Category: "Attack", IsVariable: true},
		{Code: "att-undead", Name: "AR vs Undead", Description: "+{value} To Attack Rating Against Undead", Category: "Attack", IsVariable: true},
		{Code: "ignore-ac", Name: "Ignore Defense", Description: "Ignore Target's Defense", Category: "Attack", IsVariable: false},

		// Defense
		{Code: "ac", Name: "Defense", Description: "+{value} Defense", Category: "Defense", IsVariable: true},
		{Code: "ac%", Name: "Enhanced Defense", Description: "+{value}% Enhanced Defense", Category: "Defense", IsVariable: true},
		{Code: "red-dmg", Name: "Damage Reduced", Description: "Damage Reduced By {value}", Category: "Defense", IsVariable: true},
		{Code: "red-dmg%", Name: "Damage Reduced %", Description: "Damage Reduced By {value}%", Category: "Defense", IsVariable: true},
		{Code: "red-mag", Name: "Magic Damage Reduced", Description: "Magic Damage Reduced By {value}", Category: "Defense", IsVariable: true},

		// Leech
		{Code: "life_steal", Name: "Life Steal", Description: "{value}% Life Stolen Per Hit", Category: "Leech", Aliases: []string{"lifesteal"}, IsVariable: true},
		{Code: "mana_steal", Name: "Mana Steal", Description: "{value}% Mana Stolen Per Hit", Category: "Leech", Aliases: []string{"manasteal"}, IsVariable: true},
		{Code: "hp/kill", Name: "Life per Kill", Description: "+{value} Life After Each Kill", Category: "Leech", IsVariable: true},
		{Code: "mana/kill", Name: "Mana per Kill", Description: "+{value} Mana After Each Kill", Category: "Leech", IsVariable: true},

		// Combat
		{Code: "crushing_blow", Name: "Crushing Blow", Description: "{value}% Chance Of Crushing Blow", Category: "Combat", Aliases: []string{"crush"}, IsVariable: true},
		{Code: "deadly_strike", Name: "Deadly Strike", Description: "{value}% Deadly Strike", Category: "Combat", Aliases: []string{"deadly"}, IsVariable: true},
		{Code: "open_wounds", Name: "Open Wounds", Description: "{value}% Chance Of Open Wounds", Category: "Combat", Aliases: []string{"openwounds"}, IsVariable: true},
		{Code: "knock", Name: "Knockback", Description: "Knockback", Category: "Combat", IsVariable: false},
		{Code: "slow", Name: "Slow Target", Description: "Slows Target By {value}%", Category: "Combat", IsVariable: true},
		{Code: "noheal", Name: "Prevent Monster Heal", Description: "Prevent Monster Heal", Category: "Combat", IsVariable: false},

		// Magic Find & Gold
		{Code: "mf", Name: "Magic Find", Description: "+{value}% Better Chance Of Getting Magic Items", Category: "Magic Find", Aliases: []string{"mag%"}, IsVariable: true},
		{Code: "gf", Name: "Gold Find", Description: "+{value}% Extra Gold From Monsters", Category: "Magic Find", Aliases: []string{"gold%"}, IsVariable: true},

		// Pierce
		{Code: "pierce-fire", Name: "Fire Pierce", Description: "-{value}% To Enemy Fire Resistance", Category: "Pierce", IsVariable: true},
		{Code: "pierce-cold", Name: "Cold Pierce", Description: "-{value}% To Enemy Cold Resistance", Category: "Pierce", IsVariable: true},
		{Code: "pierce-ltng", Name: "Lightning Pierce", Description: "-{value}% To Enemy Lightning Resistance", Category: "Pierce", IsVariable: true},
		{Code: "pierce-pois", Name: "Poison Pierce", Description: "-{value}% To Enemy Poison Resistance", Category: "Pierce", IsVariable: true},

		// Other
		{Code: "sock", Name: "Sockets", Description: "Socketed ({value})", Category: "Other", IsVariable: true},
		{Code: "nofreeze", Name: "Cannot Be Frozen", Description: "Cannot Be Frozen", Category: "Other", IsVariable: false},
		{Code: "half-freeze", Name: "Half Freeze Duration", Description: "Half Freeze Duration", Category: "Other", IsVariable: false},
		{Code: "indestruct", Name: "Indestructible", Description: "Indestructible", Category: "Other", IsVariable: false},
		{Code: "ethereal", Name: "Ethereal", Description: "Ethereal (Cannot Be Repaired)", Category: "Other", IsVariable: false},
		{Code: "light", Name: "Light Radius", Description: "+{value} To Light Radius", Category: "Other", IsVariable: true},
		{Code: "thorns", Name: "Thorns", Description: "Attacker Takes Damage Of {value}", Category: "Other", IsVariable: true},
		{Code: "ease", Name: "Requirements", Description: "Requirements -{value}%", Category: "Other", IsVariable: true},
		{Code: "exp", Name: "Experience", Description: "+{value}% To Experience Gained", Category: "Other", IsVariable: true},
	}
}
