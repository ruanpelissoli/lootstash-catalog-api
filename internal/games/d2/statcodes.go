package d2

import "sort"

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
	"Defense",
	"Leech",
	"Combat",
	"Magic Find",
	"Pierce",
	"Per Level",
	"Sunder",
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
		{Code: "war", Name: "Warlock Skills", Description: "+{value} To Warlock Skill Levels", Category: "Skills", IsVariable: true},

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

		// Skill Trees - Warlock
		{Code: "war-psychic", Name: "Psychic Skills", Description: "+{value} To Psychic Skills (Warlock Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "war-demonic", Name: "Demonic Binding", Description: "+{value} To Demonic Binding Skills (Warlock Only)", Category: "Skill Trees", IsVariable: true},
		{Code: "war-chaos", Name: "Arts of Chaos", Description: "+{value} To Arts of Chaos Skills (Warlock Only)", Category: "Skill Trees", IsVariable: true},

		// Attributes
		{Code: "str", Name: "Strength", Description: "+{value} To Strength", Category: "Attributes", IsVariable: true},
		{Code: "dex", Name: "Dexterity", Description: "+{value} To Dexterity", Category: "Attributes", IsVariable: true},
		{Code: "vit", Name: "Vitality", Description: "+{value} To Vitality", Category: "Attributes", IsVariable: true},
		{Code: "enr", Name: "Energy", Description: "+{value} To Energy", Category: "Attributes", Aliases: []string{"*enr"}, IsVariable: true},
		{Code: "all-stats", Name: "All Attributes", Description: "+{value} To All Attributes", Category: "Attributes", IsVariable: true},

		// Life & Mana
		{Code: "hp", Name: "Life", Description: "+{value} To Life", Category: "Life & Mana", Aliases: []string{"*hp"}, IsVariable: true},
		{Code: "mana", Name: "Mana", Description: "+{value} To Mana", Category: "Life & Mana", IsVariable: true},
		{Code: "hp%", Name: "Life %", Description: "+{value}% To Life", Category: "Life & Mana", IsVariable: true},
		{Code: "mana%", Name: "Mana %", Description: "+{value}% To Mana", Category: "Life & Mana", IsVariable: true},
		{Code: "regen-mana", Name: "Mana Regen", Description: "Regenerate Mana {value}%", Category: "Life & Mana", IsVariable: true},
		{Code: "regen", Name: "Replenish Life", Description: "Replenish Life +{value}", Category: "Life & Mana", IsVariable: true},
		{Code: "stam", Name: "Stamina", Description: "+{value} To Stamina", Category: "Life & Mana", IsVariable: true},
		{Code: "regen-stam", Name: "Stamina Regen", Description: "Heal Stamina Plus {value}%", Category: "Life & Mana", IsVariable: true},
		{Code: "stamdrain", Name: "Slower Stamina Drain", Description: "+{value}% Slower Stamina Drain", Category: "Life & Mana", IsVariable: true},

		// Speed - these have aliases because game data uses numbered variants
		{Code: "fcr", Name: "Faster Cast Rate", Description: "+{value}% Faster Cast Rate", Category: "Speed", Aliases: []string{"cast1", "cast2", "cast3"}, IsVariable: true},
		{Code: "ias", Name: "Increased Attack Speed", Description: "+{value}% Increased Attack Speed", Category: "Speed", Aliases: []string{"swing1", "swing2", "swing3"}, IsVariable: true},
		{Code: "frw", Name: "Faster Run/Walk", Description: "+{value}% Faster Run/Walk", Category: "Speed", Aliases: []string{"move1", "move2", "move3"}, IsVariable: true},
		{Code: "fhr", Name: "Faster Hit Recovery", Description: "+{value}% Faster Hit Recovery", Category: "Speed", Aliases: []string{"balance1", "balance2", "balance3"}, IsVariable: true},
		{Code: "block", Name: "Faster Block Rate", Description: "+{value}% Faster Block Rate", Category: "Speed", Aliases: []string{"block1", "block2", "block3"}, IsVariable: true},

		// Resistances - these have aliases for different code conventions
		{Code: "fire_res", Name: "Fire Resistance", Description: "Fire Resist +{value}%", Category: "Resistances", Aliases: []string{"res-fire"}, IsVariable: true},
		{Code: "cold_res", Name: "Cold Resistance", Description: "Cold Resist +{value}%", Category: "Resistances", Aliases: []string{"res-cold"}, IsVariable: true},
		{Code: "light_res", Name: "Lightning Resistance", Description: "Lightning Resist +{value}%", Category: "Resistances", Aliases: []string{"res-ltng"}, IsVariable: true},
		{Code: "poison_res", Name: "Poison Resistance", Description: "Poison Resist +{value}%", Category: "Resistances", Aliases: []string{"res-pois"}, IsVariable: true},
		{Code: "all_res", Name: "All Resistances", Description: "All Resistances +{value}", Category: "Resistances", Aliases: []string{"res-all"}, IsVariable: true},
		{Code: "res-mag", Name: "Magic Resistance", Description: "Magic Resist +{value}%", Category: "Resistances", IsVariable: true},
		{Code: "res-fire-max", Name: "Maximum Fire Resistance", Description: "+{value}% To Maximum Fire Resist", Category: "Resistances", IsVariable: true},
		{Code: "res-cold-max", Name: "Maximum Cold Resistance", Description: "+{value}% To Maximum Cold Resist", Category: "Resistances", IsVariable: true},
		{Code: "res-ltng-max", Name: "Maximum Lightning Resistance", Description: "+{value}% To Maximum Lightning Resist", Category: "Resistances", IsVariable: true},
		{Code: "res-pois-max", Name: "Maximum Poison Resistance", Description: "+{value}% To Maximum Poison Resist", Category: "Resistances", IsVariable: true},
		{Code: "res-all-max", Name: "Maximum All Resistances", Description: "+{value}% To Maximum All Resistances", Category: "Resistances", IsVariable: true},
		{Code: "res-pois-len", Name: "Poison Length Reduced", Description: "Poison Length Reduced By {value}%", Category: "Resistances", IsVariable: true},

		// Absorb
		{Code: "abs-fire", Name: "Fire Absorb", Description: "+{value} Fire Absorb", Category: "Absorb", IsVariable: true},
		{Code: "abs-cold", Name: "Cold Absorb", Description: "+{value} Cold Absorb", Category: "Absorb", IsVariable: true},
		{Code: "abs-ltng", Name: "Lightning Absorb", Description: "+{value} Lightning Absorb", Category: "Absorb", IsVariable: true},
		{Code: "abs-fire%", Name: "Fire Absorb %", Description: "{value}% Fire Absorb", Category: "Absorb", IsVariable: true},
		{Code: "abs-cold%", Name: "Cold Absorb %", Description: "{value}% Cold Absorb", Category: "Absorb", IsVariable: true},
		{Code: "abs-ltng%", Name: "Lightning Absorb %", Description: "{value}% Lightning Absorb", Category: "Absorb", IsVariable: true},
		{Code: "abs-mag", Name: "Magic Absorb", Description: "+{value} Magic Absorb", Category: "Absorb", IsVariable: true},

		// Damage
		{Code: "ed", Name: "Enhanced Damage", Description: "+{value}% Enhanced Damage", Category: "Damage", Aliases: []string{"dmg%"}, IsVariable: true},
		{Code: "dmg-min", Name: "Minimum Damage", Description: "+{value} To Minimum Damage", Category: "Damage", IsVariable: true},
		{Code: "dmg-max", Name: "Maximum Damage", Description: "+{value} To Maximum Damage", Category: "Damage", IsVariable: true},
		{Code: "dmg", Name: "Damage", Description: "+{value} Damage", Category: "Damage", IsVariable: true},
		{Code: "dmg-norm", Name: "Adds Damage", Description: "Adds {min}-{max} Damage", Category: "Damage", IsVariable: true},
		{Code: "dmg-fire", Name: "Adds Fire Damage", Description: "Adds {min}-{max} Fire Damage", Category: "Damage", Aliases: []string{"fire-min", "fire-max"}, IsVariable: true},
		{Code: "dmg-cold", Name: "Adds Cold Damage", Description: "Adds {min}-{max} Cold Damage", Category: "Damage", Aliases: []string{"cold-min", "cold-max", "cold-len"}, IsVariable: true},
		{Code: "dmg-ltng", Name: "Adds Lightning Damage", Description: "Adds {min}-{max} Lightning Damage", Category: "Damage", Aliases: []string{"ltng-min", "ltng-max"}, IsVariable: true},
		{Code: "dmg-pois", Name: "Poison Damage", Description: "+{value} Poison Damage Over {param} Seconds", Category: "Damage", Aliases: []string{"pois-min", "pois-max", "pois-len"}, IsVariable: true},
		{Code: "dmg-mag", Name: "Adds Magic Damage", Description: "Adds {min}-{max} Magic Damage", Category: "Damage", IsVariable: true},
		{Code: "dmg-elem", Name: "Adds Elemental Damage", Description: "Adds Elemental Damage", Category: "Damage", IsVariable: true},
		{Code: "dmg-demon", Name: "Damage To Demons", Description: "+{value}% Damage To Demons", Category: "Damage", IsVariable: true},
		{Code: "dmg-undead", Name: "Damage To Undead", Description: "+{value}% Damage To Undead", Category: "Damage", IsVariable: true},
		{Code: "extra-fire", Name: "Fire Skill Damage", Description: "+{value}% To Fire Skill Damage", Category: "Damage", IsVariable: true},
		{Code: "extra-cold", Name: "Cold Skill Damage", Description: "+{value}% To Cold Skill Damage", Category: "Damage", IsVariable: true},
		{Code: "extra-ltng", Name: "Lightning Skill Damage", Description: "+{value}% To Lightning Skill Damage", Category: "Damage", IsVariable: true},
		{Code: "extra-pois", Name: "Poison Skill Damage", Description: "+{value}% To Poison Skill Damage", Category: "Damage", IsVariable: true},

		// Attack
		{Code: "ar", Name: "Attack Rating", Description: "+{value} To Attack Rating", Category: "Attack", Aliases: []string{"att", "att%"}, IsVariable: true},
		{Code: "att-demon", Name: "Attack Rating Against Demons", Description: "+{value} To Attack Rating Against Demons", Category: "Attack", IsVariable: true},
		{Code: "att-undead", Name: "Attack Rating Against Undead", Description: "+{value} To Attack Rating Against Undead", Category: "Attack", IsVariable: true},
		{Code: "ignore-ac", Name: "Ignore Target's Defense", Description: "Ignore Target's Defense", Category: "Attack", IsVariable: false},

		// Defense
		{Code: "ac", Name: "Defense", Description: "+{value} Defense", Category: "Defense", IsVariable: true},
		{Code: "ac%", Name: "Enhanced Defense", Description: "+{value}% Enhanced Defense", Category: "Defense", IsVariable: true},
		{Code: "ac-miss", Name: "Defense vs Missile", Description: "+{value} Defense vs. Missile", Category: "Defense", IsVariable: true},
		{Code: "ac-hth", Name: "Defense vs Melee", Description: "+{value} Defense vs. Melee", Category: "Defense", IsVariable: true},
		{Code: "red-dmg", Name: "Damage Reduced", Description: "Damage Reduced By {value}", Category: "Defense", IsVariable: true},
		{Code: "red-dmg%", Name: "Damage Reduced %", Description: "Damage Reduced By {value}%", Category: "Defense", IsVariable: true},
		{Code: "red-mag", Name: "Magic Damage Reduced", Description: "Magic Damage Reduced By {value}", Category: "Defense", IsVariable: true},

		// Leech
		{Code: "life_steal", Name: "Life Stolen Per Hit", Description: "{value}% Life Stolen Per Hit", Category: "Leech", Aliases: []string{"lifesteal"}, IsVariable: true},
		{Code: "mana_steal", Name: "Mana Stolen Per Hit", Description: "{value}% Mana Stolen Per Hit", Category: "Leech", Aliases: []string{"manasteal"}, IsVariable: true},
		{Code: "hp/kill", Name: "Life After Each Kill", Description: "+{value} Life After Each Kill", Category: "Leech", Aliases: []string{"heal-kill"}, IsVariable: true},
		{Code: "mana/kill", Name: "Mana After Each Kill", Description: "+{value} Mana After Each Kill", Category: "Leech", Aliases: []string{"mana-kill"}, IsVariable: true},
		{Code: "dmg-to-mana", Name: "Damage Taken Goes To Mana", Description: "{value}% Damage Taken Goes To Mana", Category: "Leech", Aliases: []string{"dmg-ac"}, IsVariable: true},

		// Combat
		{Code: "crushing_blow", Name: "Chance Of Crushing Blow", Description: "{value}% Chance Of Crushing Blow", Category: "Combat", Aliases: []string{"crush"}, IsVariable: true},
		{Code: "deadly_strike", Name: "Deadly Strike", Description: "{value}% Deadly Strike", Category: "Combat", Aliases: []string{"deadly"}, IsVariable: true},
		{Code: "open_wounds", Name: "Chance Of Open Wounds", Description: "{value}% Chance Of Open Wounds", Category: "Combat", Aliases: []string{"openwounds"}, IsVariable: true},
		{Code: "knock", Name: "Knockback", Description: "Knockback", Category: "Combat", IsVariable: false},
		{Code: "slow", Name: "Slows Target", Description: "Slows Target By {value}%", Category: "Combat", IsVariable: true},
		{Code: "noheal", Name: "Prevent Monster Heal", Description: "Prevent Monster Heal", Category: "Combat", IsVariable: false},
		{Code: "freeze", Name: "Freezes Target", Description: "Freezes Target +{value}", Category: "Combat", IsVariable: true},
		{Code: "howl", Name: "Hit Causes Monster To Flee", Description: "Hit Causes Monster To Flee {value}%", Category: "Combat", IsVariable: true},
		{Code: "stupidity", Name: "Hit Blinds Target", Description: "Hit Blinds Target +{value}", Category: "Combat", IsVariable: true},
		{Code: "reduce-ac", Name: "Target Defense", Description: "-{value}% Target Defense", Category: "Combat", IsVariable: true},
		{Code: "pierce", Name: "Piercing Attack", Description: "Piercing Attack {value}%", Category: "Combat", IsVariable: true},
		{Code: "bloody", Name: "Slain Monsters Rest In Peace", Description: "Slain Monsters Rest In Peace", Category: "Combat", Aliases: []string{"rip"}, IsVariable: false},
		{Code: "demon-heal", Name: "Damage To Demons Heals You", Description: "Damage Taken By Demons Heals You", Category: "Combat", IsVariable: false},
		{Code: "light-thorns", Name: "Attacker Takes Lightning Damage", Description: "Attacker Takes Lightning Damage Of {value}", Category: "Combat", IsVariable: true},

		// Magic Find & Gold
		{Code: "mf", Name: "Magic Find", Description: "+{value}% Better Chance Of Getting Magic Items", Category: "Magic Find", Aliases: []string{"mag%"}, IsVariable: true},
		{Code: "gf", Name: "Gold Find", Description: "+{value}% Extra Gold From Monsters", Category: "Magic Find", Aliases: []string{"gold%"}, IsVariable: true},

		// Pierce
		{Code: "pierce-fire", Name: "Enemy Fire Resistance", Description: "-{value}% To Enemy Fire Resistance", Category: "Pierce", IsVariable: true},
		{Code: "pierce-cold", Name: "Enemy Cold Resistance", Description: "-{value}% To Enemy Cold Resistance", Category: "Pierce", IsVariable: true},
		{Code: "pierce-ltng", Name: "Enemy Lightning Resistance", Description: "-{value}% To Enemy Lightning Resistance", Category: "Pierce", IsVariable: true},
		{Code: "pierce-pois", Name: "Enemy Poison Resistance", Description: "-{value}% To Enemy Poison Resistance", Category: "Pierce", IsVariable: true},
		{Code: "pierce-mag", Name: "Enemy Magic Resistance", Description: "-{value}% To Enemy Magic Resistance", Category: "Pierce", IsVariable: true},

		// Sunder Charms
		{Code: "pierce-immunity-fire", Name: "Fire Immunity Sundered", Description: "Monster Fire Immunity is Sundered", Category: "Sunder", IsVariable: false},
		{Code: "pierce-immunity-cold", Name: "Cold Immunity Sundered", Description: "Monster Cold Immunity is Sundered", Category: "Sunder", IsVariable: false},
		{Code: "pierce-immunity-light", Name: "Lightning Immunity Sundered", Description: "Monster Lightning Immunity is Sundered", Category: "Sunder", IsVariable: false},
		{Code: "pierce-immunity-poison", Name: "Poison Immunity Sundered", Description: "Monster Poison Immunity is Sundered", Category: "Sunder", IsVariable: false},
		{Code: "pierce-immunity-magic", Name: "Magic Immunity Sundered", Description: "Monster Magic Immunity is Sundered", Category: "Sunder", IsVariable: false},
		{Code: "pierce-immunity-damage", Name: "Physical Immunity Sundered", Description: "Monster Physical Immunity is Sundered", Category: "Sunder", IsVariable: false},

		// Per Level
		{Code: "str/lvl", Name: "Strength/Lvl", Description: "+{value} To Strength (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "dex/lvl", Name: "Dexterity/Lvl", Description: "+{value} To Dexterity (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "vit/lvl", Name: "Vitality/Lvl", Description: "+{value} To Vitality (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "hp/lvl", Name: "Life/Lvl", Description: "+{value} To Life (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "mana/lvl", Name: "Mana/Lvl", Description: "+{value} To Mana (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "stam/lvl", Name: "Stamina/Lvl", Description: "+{value} To Stamina (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "ac/lvl", Name: "Defense/Lvl", Description: "+{value} Defense (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "att/lvl", Name: "Attack Rating/Lvl", Description: "+{value} To Attack Rating (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "dmg/lvl", Name: "Max Damage/Lvl", Description: "+{value} To Maximum Damage (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "dmg%/lvl", Name: "Enh Damage/Lvl", Description: "+{value}% Enhanced Damage (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "thorns/lvl", Name: "Thorns/Lvl", Description: "+{value} Thorns (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "deadly/lvl", Name: "Deadly Strike/Lvl", Description: "+{value}% Deadly Strike (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "gold%/lvl", Name: "Gold Find/Lvl", Description: "+{value}% Extra Gold (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "mag%/lvl", Name: "Magic Find/Lvl", Description: "+{value}% Better Chance Of Magic Items (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "regen-stam/lvl", Name: "Stamina Regen/Lvl", Description: "Heal Stamina +{value}% (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "res-ltng/lvl", Name: "Lightning Res/Lvl", Description: "+{value}% Lightning Resist (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "abs-fire/lvl", Name: "Fire Absorb/Lvl", Description: "+{value} Fire Absorb (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "abs-cold/lvl", Name: "Cold Absorb/Lvl", Description: "+{value} Cold Absorb (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "dmg-fire/lvl", Name: "Fire Damage/Lvl", Description: "+{value} Fire Damage (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "dmg-cold/lvl", Name: "Cold Damage/Lvl", Description: "+{value} Cold Damage (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "dmg-ltng/lvl", Name: "Lightning Dmg/Lvl", Description: "+{value} Lightning Damage (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "dmg-dem/lvl", Name: "Dmg to Demons/Lvl", Description: "+{value}% Damage To Demons (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "dmg-und/lvl", Name: "Dmg to Undead/Lvl", Description: "+{value}% Damage To Undead (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "att-dem/lvl", Name: "AR vs Demons/Lvl", Description: "+{value} AR Against Demons (Based On Character Level)", Category: "Per Level", IsVariable: true},
		{Code: "att-und/lvl", Name: "AR vs Undead/Lvl", Description: "+{value} AR Against Undead (Based On Character Level)", Category: "Per Level", IsVariable: true},

		// Other
		{Code: "sock", Name: "Socketed", Description: "Socketed ({value})", Category: "Other", IsVariable: true},
		{Code: "nofreeze", Name: "Cannot Be Frozen", Description: "Cannot Be Frozen", Category: "Other", IsVariable: false},
		{Code: "half-freeze", Name: "Half Freeze Duration", Description: "Half Freeze Duration", Category: "Other", IsVariable: false},
		{Code: "indestruct", Name: "Indestructible", Description: "Indestructible", Category: "Other", IsVariable: false},
		{Code: "ethereal", Name: "Ethereal", Description: "Ethereal (Cannot Be Repaired)", Category: "Other", IsVariable: false},
		{Code: "light", Name: "Light Radius", Description: "+{value} To Light Radius", Category: "Other", IsVariable: true},
		{Code: "thorns", Name: "Attacker Takes Damage", Description: "Attacker Takes Damage Of {value}", Category: "Other", IsVariable: true},
		{Code: "ease", Name: "Requirements", Description: "Requirements -{value}%", Category: "Other", IsVariable: true},
		{Code: "exp", Name: "Experience Gained", Description: "+{value}% To Experience Gained", Category: "Other", Aliases: []string{"addxp"}, IsVariable: true},
		{Code: "dur", Name: "Maximum Durability", Description: "+{value} To Maximum Durability", Category: "Other", IsVariable: true},
		{Code: "rep-dur", Name: "Repairs Durability", Description: "Repairs 1 Durability In {value} Seconds", Category: "Other", IsVariable: true},
		{Code: "rep-quant", Name: "Replenishes Quantity", Description: "Replenishes Quantity", Category: "Other", IsVariable: false},
		{Code: "stack", Name: "Maximum Quantity", Description: "+{value} To Maximum Quantity", Category: "Other", IsVariable: true},
		{Code: "cheap", Name: "Reduces Vendor Prices", Description: "Reduces All Vendor Prices {value}%", Category: "Other", IsVariable: true},
		{Code: "teleport", Name: "Teleport", Description: "+1 To Teleport", Category: "Other", IsVariable: false},
	}
}

// parametricStatCodes are codes that are dynamic/parametric and don't belong
// in a fixed filter list (they use param to specify the actual skill/effect),
// or are internal component codes that combine into other display stats.
var parametricStatCodes = map[string]bool{
	"raw":              true,
	"skilltab":         true,
	"skill":            true,
	"oskill":           true,
	"aura":             true,
	"charged":          true,
	"*charged":         true,
	"hit-skill":        true,
	"gethit-skill":     true,
	"kill-skill":       true,
	"death-skill":      true,
	"levelup-skill":    true,
	"att-skill":        true,
	"randclassskill":   true,
	"skill-rand":       true,
	"reanimate":        true,
	"state":            true,
	"fade":             true,
	"charge-noconsume": true,
	"explosivearrow":   true,
	"magicarrow":       true,
	"fireskill":        true,
}

// ValidateStatCodes checks a list of property codes against the FilterableStats
// registry and returns any codes that are missing (not registered and not parametric).
func ValidateStatCodes(propertyCodes []string) []string {
	// Build lookup set from FilterableStats (codes + aliases)
	known := make(map[string]bool)
	for _, stat := range FilterableStats() {
		known[stat.Code] = true
		for _, alias := range stat.Aliases {
			known[alias] = true
		}
	}

	// Find missing codes (deduplicated)
	seen := make(map[string]bool)
	var missing []string
	for _, code := range propertyCodes {
		if seen[code] {
			continue
		}
		seen[code] = true
		if parametricStatCodes[code] || known[code] {
			continue
		}
		missing = append(missing, code)
	}

	sort.Strings(missing)
	return missing
}
