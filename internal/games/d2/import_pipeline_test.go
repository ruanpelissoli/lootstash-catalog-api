package d2

import (
	"strings"
	"testing"
)

// catalogPath returns the relative path to the HTML catalog pages directory.
func catalogPath(file string) string {
	return "../../../catalogs/d2/pages/" + file
}

// findProp searches a property slice for the first property with the given code.
func findProp(props []Property, code string) *Property {
	for i := range props {
		if props[i].Code == code {
			return &props[i]
		}
	}
	return nil
}

// findPropWithParam searches for a property matching both code and param.
func findPropWithParam(props []Property, code, param string) *Property {
	for i := range props {
		if props[i].Code == code && props[i].Param == param {
			return &props[i]
		}
	}
	return nil
}

// enrichAll reverse-translates lines and enriches properties (mirrors the import pipeline).
func enrichAll(lines []string) []Property {
	rt := NewReverseTranslator()
	translator := NewPropertyTranslator()

	props := rt.ReverseTranslateLines(lines)
	props = combineAllAttributes(props, translator)
	for i := range props {
		if props[i].Code != "raw" {
			translator.EnrichProperty(&props[i])
		}
	}
	return props
}

// --- Unique Items ---

func TestParseAndTranslateUniqueItem_HarlequinCrest(t *testing.T) {
	parser := NewHTMLItemParser()

	items, err := parser.ParseUniquesFile(catalogPath("uniques.html"))
	if err != nil {
		t.Fatalf("ParseUniquesFile failed: %v", err)
	}

	var target *HTMLParsedUniqueItem
	for i := range items {
		if items[i].Name == "Harlequin Crest" {
			target = &items[i]
			break
		}
	}
	if target == nil {
		t.Fatal("Harlequin Crest not found in uniques.html")
	}

	// Validate parsed metadata
	if target.BaseName != "Shako" {
		t.Errorf("BaseName = %q, want %q", target.BaseName, "Shako")
	}
	if target.Quality != "Elite Unique" {
		t.Errorf("Quality = %q, want %q", target.Quality, "Elite Unique")
	}
	if target.ReqLevel != 62 {
		t.Errorf("ReqLevel = %d, want %d", target.ReqLevel, 62)
	}
	if target.QualityLevel != 69 {
		t.Errorf("QualityLevel = %d, want %d", target.QualityLevel, 69)
	}
	if target.ImagePath == "" {
		t.Error("ImagePath should not be empty")
	}

	// Reverse-translate and enrich properties (same as import pipeline)
	props := enrichAll(target.Properties)

	if len(props) == 0 {
		t.Fatal("Expected properties after translation, got none")
	}

	// +2 To All Skills — fixed stat
	if p := findProp(props, "allskills"); p == nil {
		t.Error("missing property: allskills")
	} else {
		if p.Min != 2 || p.Max != 2 {
			t.Errorf("allskills: Min=%d, Max=%d, want 2/2", p.Min, p.Max)
		}
		if p.HasRange {
			t.Error("allskills should not have range (fixed stat)")
		}
	}

	// Per-level Life: (1.5 Per Character Level) 1-148 To Life
	// Reverse-translated: hp/lvl with raw value 12 (1.5 * 8 = 12)
	if p := findProp(props, "hp/lvl"); p == nil {
		t.Error("missing property: hp/lvl")
	} else {
		if p.Min != 12 || p.Max != 12 {
			t.Errorf("hp/lvl: Min=%d, Max=%d, want 12/12", p.Min, p.Max)
		}
		if p.HasRange {
			t.Error("hp/lvl should not have range (per-level codes are fixed)")
		}
		if !strings.Contains(p.DisplayText, "1-148") || !strings.Contains(p.DisplayText, "Life") {
			t.Errorf("hp/lvl DisplayText = %q, should contain '1-148' and 'Life'", p.DisplayText)
		}
	}

	// Per-level Mana: same pattern
	if p := findProp(props, "mana/lvl"); p == nil {
		t.Error("missing property: mana/lvl")
	} else {
		if p.Min != 12 || p.Max != 12 {
			t.Errorf("mana/lvl: Min=%d, Max=%d, want 12/12", p.Min, p.Max)
		}
		if !strings.Contains(p.DisplayText, "1-148") || !strings.Contains(p.DisplayText, "Mana") {
			t.Errorf("mana/lvl DisplayText = %q, should contain '1-148' and 'Mana'", p.DisplayText)
		}
	}

	// Damage Reduced By 10% — fixed stat
	if p := findProp(props, "red-dmg%"); p == nil {
		t.Error("missing property: red-dmg%")
	} else {
		if p.Min != 10 || p.Max != 10 {
			t.Errorf("red-dmg%%: Min=%d, Max=%d, want 10/10", p.Min, p.Max)
		}
		if p.HasRange {
			t.Error("red-dmg% should not have range")
		}
	}

	// 50% Better Chance of Getting Magic Items
	if p := findProp(props, "mag%"); p == nil {
		t.Error("missing property: mag%")
	} else {
		if p.Min != 50 || p.Max != 50 {
			t.Errorf("mag%%: Min=%d, Max=%d, want 50/50", p.Min, p.Max)
		}
	}

	// +2 To All Attributes (combined from str/dex/vit/enr via combineAllAttributes)
	if p := findProp(props, "all-stats"); p == nil {
		t.Error("missing property: all-stats (should be combined from str/dex/vit/enr)")
	} else {
		if p.Min != 2 || p.Max != 2 {
			t.Errorf("all-stats: Min=%d, Max=%d, want 2/2", p.Min, p.Max)
		}
	}

	// Should NOT have individual str/dex/vit/enr (they should be combined)
	for _, code := range []string{"str", "dex", "vit", "enr"} {
		if p := findProp(props, code); p != nil {
			t.Errorf("should not have individual %s property after combineAllAttributes", code)
		}
	}

	// No "raw" (unrecognized) properties
	for _, p := range props {
		if p.Code == "raw" {
			t.Errorf("unrecognized property: %q", p.DisplayText)
		}
	}
}

func TestParseAndTranslateUniqueItem_ColdRupture(t *testing.T) {
	parser := NewHTMLItemParser()

	items, err := parser.ParseUniquesFile(catalogPath("uniques.html"))
	if err != nil {
		t.Fatalf("ParseUniquesFile failed: %v", err)
	}

	var target *HTMLParsedUniqueItem
	for i := range items {
		if items[i].Name == "Cold Rupture" {
			target = &items[i]
			break
		}
	}
	if target == nil {
		t.Fatal("Cold Rupture not found in uniques.html")
	}

	// Validate parsed metadata
	if target.BaseName != "Grand Charm" {
		t.Errorf("BaseName = %q, want %q", target.BaseName, "Grand Charm")
	}
	if target.Quality != "Elite Unique" {
		t.Errorf("Quality = %q, want %q", target.Quality, "Elite Unique")
	}
	if target.ReqLevel != 75 {
		t.Errorf("ReqLevel = %d, want %d", target.ReqLevel, 75)
	}

	// Reverse-translate properties
	props := enrichAll(target.Properties)

	if len(props) == 0 {
		t.Fatal("Expected properties after translation, got none")
	}

	// Monster Cold Immunity is Sundered — fixed text, no variable values
	if p := findProp(props, "pierce-immunity-cold"); p == nil {
		t.Error("missing property: pierce-immunity-cold")
	} else {
		if p.HasRange {
			t.Error("pierce-immunity-cold should not have range (fixed text)")
		}
	}

	// Cold resist -(70-90)% — negative ranged stat
	if p := findProp(props, "res-cold"); p == nil {
		t.Error("missing property: res-cold")
	} else {
		// Negative range: -(70-90) means min=-90, max=-70
		if p.Min != -90 || p.Max != -70 {
			t.Errorf("res-cold: Min=%d, Max=%d, want -90/-70", p.Min, p.Max)
		}
		if !p.HasRange {
			t.Error("res-cold should have range (variable stat)")
		}
	}
}

// --- Set Items ---

func TestParseAndTranslateSetItem_TalRashaHelm(t *testing.T) {
	parser := NewHTMLItemParser()

	setItems, _, err := parser.ParseSetsFile(catalogPath("sets.html"))
	if err != nil {
		t.Fatalf("ParseSetsFile failed: %v", err)
	}

	var target *HTMLParsedSetItem
	for i := range setItems {
		if setItems[i].Name == "Tal Rasha's Horadric Crest" {
			target = &setItems[i]
			break
		}
	}
	if target == nil {
		// Try with curly apostrophe
		for i := range setItems {
			if strings.Contains(setItems[i].Name, "Tal Rasha") && strings.Contains(setItems[i].Name, "Horadric Crest") {
				target = &setItems[i]
				break
			}
		}
	}
	if target == nil {
		t.Fatal("Tal Rasha's Horadric Crest not found in sets.html")
	}

	// Validate parsed metadata
	if target.BaseName != "Death Mask" {
		t.Errorf("BaseName = %q, want %q", target.BaseName, "Death Mask")
	}
	if target.Quality != "Exceptional Set" {
		t.Errorf("Quality = %q, want %q", target.Quality, "Exceptional Set")
	}
	if target.ReqLevel != 66 {
		t.Errorf("ReqLevel = %d, want %d", target.ReqLevel, 66)
	}
	if !strings.Contains(target.SetName, "Tal Rasha") {
		t.Errorf("SetName = %q, should contain 'Tal Rasha'", target.SetName)
	}

	// Reverse-translate properties
	props := enrichAll(target.Properties)

	if len(props) == 0 {
		t.Fatal("Expected properties after translation, got none")
	}

	// 10% Life Stolen Per Hit
	if p := findProp(props, "lifesteal"); p == nil {
		t.Error("missing property: lifesteal")
	} else {
		if p.Min != 10 || p.Max != 10 {
			t.Errorf("lifesteal: Min=%d, Max=%d, want 10/10", p.Min, p.Max)
		}
	}

	// 10% Mana Stolen Per Hit
	if p := findProp(props, "manasteal"); p == nil {
		t.Error("missing property: manasteal")
	} else {
		if p.Min != 10 || p.Max != 10 {
			t.Errorf("manasteal: Min=%d, Max=%d, want 10/10", p.Min, p.Max)
		}
	}

	// All Resistances +15
	if p := findProp(props, "res-all"); p == nil {
		t.Error("missing property: res-all")
	} else {
		if p.Min != 15 || p.Max != 15 {
			t.Errorf("res-all: Min=%d, Max=%d, want 15/15", p.Min, p.Max)
		}
	}

	// +45 Defense
	if p := findProp(props, "ac"); p == nil {
		t.Error("missing property: ac")
	} else {
		if p.Min != 45 || p.Max != 45 {
			t.Errorf("ac: Min=%d, Max=%d, want 45/45", p.Min, p.Max)
		}
	}

	// +30 To Mana
	if p := findProp(props, "mana"); p == nil {
		t.Error("missing property: mana")
	} else {
		if p.Min != 30 || p.Max != 30 {
			t.Errorf("mana: Min=%d, Max=%d, want 30/30", p.Min, p.Max)
		}
	}

	// +60 To Life
	if p := findProp(props, "hp"); p == nil {
		t.Error("missing property: hp")
	} else {
		if p.Min != 60 || p.Max != 60 {
			t.Errorf("hp: Min=%d, Max=%d, want 60/60", p.Min, p.Max)
		}
	}

	// No raw (unrecognized) properties
	for _, p := range props {
		if p.Code == "raw" {
			t.Errorf("unrecognized property: %q", p.DisplayText)
		}
	}
}

// --- Runewords ---

func TestParseAndTranslateRuneword_Spirit(t *testing.T) {
	parser := NewHTMLItemParser()
	rt := NewReverseTranslator()
	translator := NewPropertyTranslator()

	runewords, err := parser.ParseRunewordsFile(catalogPath("runewords.html"))
	if err != nil {
		t.Fatalf("ParseRunewordsFile failed: %v", err)
	}

	var target *HTMLParsedRuneword
	for i := range runewords {
		if runewords[i].Name == "Spirit" {
			target = &runewords[i]
			break
		}
	}
	if target == nil {
		t.Fatal("Spirit not found in runewords.html")
	}

	// Validate parsed metadata
	if len(target.Runes) != 4 {
		t.Fatalf("Runes count = %d, want 4", len(target.Runes))
	}
	expectedRunes := []string{"Tal", "Thul", "Ort", "Amn"}
	for i, rune := range expectedRunes {
		if target.Runes[i] != rune {
			t.Errorf("Rune[%d] = %q, want %q", i, target.Runes[i], rune)
		}
	}
	if target.ReqLevel != 25 {
		t.Errorf("ReqLevel = %d, want %d", target.ReqLevel, 25)
	}

	// Spirit should have per-type properties (Swords vs Shields differ)
	if target.PropertiesByType == nil {
		t.Fatal("Spirit should have PropertiesByType (Swords and Shields have different stats)")
	}

	swordLines, hasSwords := target.PropertiesByType["4 socket Swords"]
	shieldLines, hasShields := target.PropertiesByType["4 socket Shields"]
	if !hasSwords {
		t.Fatal("Spirit should have '4 socket Swords' in PropertiesByType")
	}
	if !hasShields {
		t.Fatal("Spirit should have '4 socket Shields' in PropertiesByType")
	}

	// Translate Sword properties
	swordProps := rt.ReverseTranslateLines(swordLines)
	swordProps = combineAllAttributes(swordProps, translator)
	for i := range swordProps {
		if swordProps[i].Code != "raw" {
			translator.EnrichProperty(&swordProps[i])
		}
	}

	// Swords: +2 To All Skills
	if p := findProp(swordProps, "allskills"); p == nil {
		t.Error("Swords: missing allskills")
	} else if p.Min != 2 || p.Max != 2 {
		t.Errorf("Swords allskills: Min=%d, Max=%d, want 2/2", p.Min, p.Max)
	}

	// Swords: +25-35% Faster Cast Rate — ranged stat
	hasFCR := false
	for _, p := range swordProps {
		if (p.Code == "cast1" || p.Code == "cast2" || p.Code == "cast3") && p.Min == 25 && p.Max == 35 {
			hasFCR = true
			if !p.HasRange {
				t.Error("Swords FCR should have range (25-35)")
			}
			break
		}
	}
	if !hasFCR {
		t.Error("Swords: missing Faster Cast Rate 25-35")
	}

	// Swords: Adds 1-50 Lightning Damage — fixed damage range, HasRange=false
	if p := findProp(swordProps, "dmg-ltng"); p == nil {
		t.Error("Swords: missing dmg-ltng")
	} else {
		if p.Min != 1 || p.Max != 50 {
			t.Errorf("Swords dmg-ltng: Min=%d, Max=%d, want 1/50", p.Min, p.Max)
		}
		if p.HasRange {
			t.Error("Swords dmg-ltng should NOT have range (fixed damage per hit)")
		}
	}

	// Swords: 7% Life Stolen Per Hit
	if p := findProp(swordProps, "lifesteal"); p == nil {
		t.Error("Swords: missing lifesteal")
	} else if p.Min != 7 || p.Max != 7 {
		t.Errorf("Swords lifesteal: Min=%d, Max=%d, want 7/7", p.Min, p.Max)
	}

	// Translate Shield properties
	shieldProps := rt.ReverseTranslateLines(shieldLines)
	shieldProps = combineAllAttributes(shieldProps, translator)
	for i := range shieldProps {
		if shieldProps[i].Code != "raw" {
			translator.EnrichProperty(&shieldProps[i])
		}
	}

	// Shields: Cold Resist +35%
	if p := findProp(shieldProps, "res-cold"); p == nil {
		t.Error("Shields: missing res-cold")
	} else if p.Min != 35 || p.Max != 35 {
		t.Errorf("Shields res-cold: Min=%d, Max=%d, want 35/35", p.Min, p.Max)
	}

	// Shields: Lightning Resist +35%
	if p := findProp(shieldProps, "res-ltng"); p == nil {
		t.Error("Shields: missing res-ltng")
	} else if p.Min != 35 || p.Max != 35 {
		t.Errorf("Shields res-ltng: Min=%d, Max=%d, want 35/35", p.Min, p.Max)
	}

	// Shields: Attacker Takes Damage of 14
	if p := findProp(shieldProps, "thorns"); p == nil {
		t.Error("Shields: missing thorns")
	} else if p.Min != 14 || p.Max != 14 {
		t.Errorf("Shields thorns: Min=%d, Max=%d, want 14/14", p.Min, p.Max)
	}

	// Shields should NOT have Lightning Damage (only Swords have it)
	if p := findProp(shieldProps, "dmg-ltng"); p != nil {
		t.Error("Shields should NOT have dmg-ltng (only Swords)")
	}

	// Shields should NOT have Life Stolen Per Hit (only Swords have it)
	if p := findProp(shieldProps, "lifesteal"); p != nil {
		t.Error("Shields should NOT have lifesteal (only Swords)")
	}
}

// --- Base Items ---

func TestParseAndTranslateBaseItem_PhaseBlade(t *testing.T) {
	parser := NewHTMLItemParser()

	items, err := parser.ParseBasesFile(catalogPath("base.html"))
	if err != nil {
		t.Fatalf("ParseBasesFile failed: %v", err)
	}

	var target *HTMLParsedBaseItem
	for i := range items {
		if items[i].Name == "Phase Blade" {
			target = &items[i]
			break
		}
	}
	if target == nil {
		t.Fatal("Phase Blade not found in base.html")
	}

	// Quality tier
	if target.Quality != "Elite" {
		t.Errorf("Quality = %q, want %q", target.Quality, "Elite")
	}

	// Weapon stats
	if target.OneHMinDam != 31 || target.OneHMaxDam != 35 {
		t.Errorf("1H Damage = %d-%d, want 31-35", target.OneHMinDam, target.OneHMaxDam)
	}
	if target.TwoHMinDam != 0 || target.TwoHMaxDam != 0 {
		t.Errorf("2H Damage = %d-%d, want 0-0 (1H only weapon)", target.TwoHMinDam, target.TwoHMaxDam)
	}
	if target.Speed != -30 {
		t.Errorf("Speed = %d, want %d", target.Speed, -30)
	}
	if target.RangeAdder != 1 {
		t.Errorf("RangeAdder = %d, want %d", target.RangeAdder, 1)
	}

	// Requirements
	if target.ReqStr != 25 {
		t.Errorf("ReqStr = %d, want %d", target.ReqStr, 25)
	}
	if target.ReqDex != 136 {
		t.Errorf("ReqDex = %d, want %d", target.ReqDex, 136)
	}
	if target.ReqLevel != 54 {
		t.Errorf("ReqLevel = %d, want %d", target.ReqLevel, 54)
	}

	// Sockets
	if target.MaxSockets != 6 {
		t.Errorf("MaxSockets = %d, want %d", target.MaxSockets, 6)
	}

	// Durability (Phase Blade is indestructible, durability=0)
	if target.Durability != 0 {
		t.Errorf("Durability = %d, want %d (indestructible)", target.Durability, 0)
	}

	// Defense (should be 0 for a weapon)
	if target.DefenseMin != 0 || target.DefenseMax != 0 {
		t.Errorf("Defense = %d-%d, want 0-0 (weapon)", target.DefenseMin, target.DefenseMax)
	}

	// Type tags should contain Swords
	hasSwords := false
	hasMeleeWeapons := false
	hasWeapons := false
	for _, tag := range target.TypeTags {
		switch tag {
		case "Swords":
			hasSwords = true
		case "Melee Weapons":
			hasMeleeWeapons = true
		case "Weapons":
			hasWeapons = true
		}
	}
	if !hasSwords {
		t.Errorf("TypeTags %v should contain 'Swords'", target.TypeTags)
	}
	if !hasMeleeWeapons {
		t.Errorf("TypeTags %v should contain 'Melee Weapons'", target.TypeTags)
	}
	if !hasWeapons {
		t.Errorf("TypeTags %v should contain 'Weapons'", target.TypeTags)
	}

	// Inventory size
	if target.InvWidth != 2 || target.InvHeight != 3 {
		t.Errorf("InvSize = %dx%d, want 2x3", target.InvWidth, target.InvHeight)
	}

	// Variant links should include Crystal Sword and Dimensional Blade.
	// Variant names may include trailing whitespace/tier annotations from the HTML,
	// e.g., "Crystal Sword\n                       (Nor.)" — use Contains matching.
	hasCrystalSword := false
	hasDimensionalBlade := false
	for _, v := range target.VariantNames {
		if strings.Contains(v.Name, "Crystal Sword") {
			hasCrystalSword = true
		}
		if strings.Contains(v.Name, "Dimensional Blade") {
			hasDimensionalBlade = true
		}
	}
	if !hasCrystalSword {
		t.Error("Variants should include Crystal Sword")
	}
	if !hasDimensionalBlade {
		t.Error("Variants should include Dimensional Blade")
	}

	// URL slug for code generation
	if target.URLSlug != "phase-blade" {
		t.Errorf("URLSlug = %q, want %q", target.URLSlug, "phase-blade")
	}
}

// --- Runes ---

func TestParseAndTranslateRune_Ber(t *testing.T) {
	parser := NewHTMLItemParser()
	rt := NewReverseTranslator()
	translator := NewPropertyTranslator()

	runes, _, _, err := parser.ParseMiscFile(catalogPath("misc.html"))
	if err != nil {
		t.Fatalf("ParseMiscFile failed: %v", err)
	}

	var target *HTMLParsedRune
	for i := range runes {
		if runes[i].Name == "Ber" {
			target = &runes[i]
			break
		}
	}
	if target == nil {
		t.Fatal("Ber rune not found in misc.html")
	}

	// Validate parsed metadata
	if target.Level != 63 {
		t.Errorf("Level = %d, want %d", target.Level, 63)
	}
	if target.RuneIndex != 30 {
		t.Errorf("RuneIndex = %d, want %d", target.RuneIndex, 30)
	}
	if target.ImagePath == "" {
		t.Error("ImagePath should not be empty")
	}

	// Weapon mods: "20% crushing blow"
	// Note: Rune/gem mod text from the HTML uses abbreviated formats (e.g., "20% crushing blow"
	// instead of the standard "20% Chance Of Crushing Blow"). The reverse translator correctly
	// parses standard format but these abbreviated forms are stored as raw properties. We validate
	// that the parsing extracted the correct text and the number of mods.
	if len(target.WeaponMods) != 1 {
		t.Fatalf("WeaponMods count = %d, want 1", len(target.WeaponMods))
	}
	if !strings.Contains(strings.ToLower(target.WeaponMods[0]), "crushing blow") {
		t.Errorf("WeaponMods[0] = %q, should contain 'crushing blow'", target.WeaponMods[0])
	}
	weaponProps := rt.ReverseTranslateLines(target.WeaponMods)
	for i := range weaponProps {
		if weaponProps[i].Code != "raw" {
			translator.EnrichProperty(&weaponProps[i])
		}
	}
	if len(weaponProps) != 1 {
		t.Errorf("Expected 1 weapon property, got %d", len(weaponProps))
	}

	// Helm (armor) mods: "8% physical resist"
	if len(target.HelmMods) != 1 {
		t.Fatalf("HelmMods count = %d, want 1", len(target.HelmMods))
	}
	if !strings.Contains(strings.ToLower(target.HelmMods[0]), "physical resist") {
		t.Errorf("HelmMods[0] = %q, should contain 'physical resist'", target.HelmMods[0])
	}

	// Shield mods: same as helm ("8% physical resist")
	if len(target.ShieldMods) != 1 {
		t.Fatalf("ShieldMods count = %d, want 1", len(target.ShieldMods))
	}
	if !strings.Contains(strings.ToLower(target.ShieldMods[0]), "physical resist") {
		t.Errorf("ShieldMods[0] = %q, should contain 'physical resist'", target.ShieldMods[0])
	}
}

// --- Gems ---

func TestParseAndTranslateGem_PerfectRuby(t *testing.T) {
	parser := NewHTMLItemParser()
	rt := NewReverseTranslator()
	translator := NewPropertyTranslator()

	_, gems, _, err := parser.ParseMiscFile(catalogPath("misc.html"))
	if err != nil {
		t.Fatalf("ParseMiscFile failed: %v", err)
	}

	var target *HTMLParsedGem
	for i := range gems {
		if gems[i].Name == "Perfect Ruby" {
			target = &gems[i]
			break
		}
	}
	if target == nil {
		t.Fatal("Perfect Ruby not found in misc.html")
	}

	// Validate gem name parsing
	gemType, quality := parseGemNameParts(target.Name)
	if gemType != "ruby" {
		t.Errorf("GemType = %q, want %q", gemType, "ruby")
	}
	if quality != "perfect" {
		t.Errorf("Quality = %q, want %q", quality, "perfect")
	}

	if target.ImagePath == "" {
		t.Error("ImagePath should not be empty")
	}

	// Weapon mods: "15-20 Fire Damage"
	// Note: Gem mods use abbreviated text (e.g., "15-20 Fire Damage" instead of "Adds 15-20 Fire Damage").
	// We validate the raw text is correctly extracted from HTML.
	if len(target.WeaponMods) != 1 {
		t.Fatalf("WeaponMods count = %d, want 1", len(target.WeaponMods))
	}
	if !strings.Contains(target.WeaponMods[0], "Fire Damage") {
		t.Errorf("WeaponMods[0] = %q, should contain 'Fire Damage'", target.WeaponMods[0])
	}
	if !strings.Contains(target.WeaponMods[0], "15-20") {
		t.Errorf("WeaponMods[0] = %q, should contain '15-20'", target.WeaponMods[0])
	}

	// Helm (armor) mods: "+38 to Life"
	if len(target.HelmMods) != 1 {
		t.Fatalf("HelmMods count = %d, want 1", len(target.HelmMods))
	}
	// "+38 to Life" should reverse-translate to hp with the fixed sort order
	helmProps := rt.ReverseTranslateLines(target.HelmMods)
	for i := range helmProps {
		if helmProps[i].Code != "raw" {
			translator.EnrichProperty(&helmProps[i])
		}
	}
	if p := findProp(helmProps, "hp"); p == nil {
		t.Error("Helm: missing hp (expected '+38 to Life' to match hp pattern)")
	} else if p.Min != 38 || p.Max != 38 {
		t.Errorf("Helm hp: Min=%d, Max=%d, want 38/38", p.Min, p.Max)
	}

	// Shield mods: "40% Resist Fire"
	// This uses legacy format "Resist Fire" instead of "Fire Resist"
	if len(target.ShieldMods) != 1 {
		t.Fatalf("ShieldMods count = %d, want 1", len(target.ShieldMods))
	}
	if !strings.Contains(target.ShieldMods[0], "Resist Fire") {
		t.Errorf("ShieldMods[0] = %q, should contain 'Resist Fire'", target.ShieldMods[0])
	}
}

// --- combineAllAttributes ---

func TestCombineAllAttributes_SameValues(t *testing.T) {
	translator := NewPropertyTranslator()

	// All 4 attributes with same values → should combine into all-stats
	props := []Property{
		{Code: "str", Min: 10, Max: 10},
		{Code: "dmg%", Min: 200, Max: 200},
		{Code: "dex", Min: 10, Max: 10},
		{Code: "vit", Min: 10, Max: 10},
		{Code: "enr", Min: 10, Max: 10},
	}

	result := combineAllAttributes(props, translator)

	// Should have 2 properties: all-stats and dmg%
	if len(result) != 2 {
		t.Fatalf("Expected 2 properties after combining, got %d: %+v", len(result), result)
	}

	if p := findProp(result, "all-stats"); p == nil {
		t.Error("missing all-stats after combining")
	} else {
		if p.Min != 10 || p.Max != 10 {
			t.Errorf("all-stats: Min=%d, Max=%d, want 10/10", p.Min, p.Max)
		}
		if p.DisplayText == "" {
			t.Error("all-stats should have DisplayText set")
		}
	}

	if p := findProp(result, "dmg%"); p == nil {
		t.Error("missing dmg% (non-attribute property should be preserved)")
	}

	// Individual attributes should be gone
	for _, code := range []string{"str", "dex", "vit", "enr"} {
		if p := findProp(result, code); p != nil {
			t.Errorf("should not have individual %s after combining", code)
		}
	}
}

func TestCombineAllAttributes_DifferentValues(t *testing.T) {
	translator := NewPropertyTranslator()

	// Attributes with different values → should NOT combine
	props := []Property{
		{Code: "str", Min: 10, Max: 10},
		{Code: "dex", Min: 15, Max: 15}, // Different!
		{Code: "vit", Min: 10, Max: 10},
		{Code: "enr", Min: 10, Max: 10},
	}

	result := combineAllAttributes(props, translator)

	// Should remain 4 individual properties
	if len(result) != 4 {
		t.Fatalf("Expected 4 properties (not combined), got %d", len(result))
	}

	if p := findProp(result, "all-stats"); p != nil {
		t.Error("should NOT have all-stats when values differ")
	}
}

func TestCombineAllAttributes_MissingAttribute(t *testing.T) {
	translator := NewPropertyTranslator()

	// Only 3 of 4 attributes → should NOT combine
	props := []Property{
		{Code: "str", Min: 10, Max: 10},
		{Code: "dex", Min: 10, Max: 10},
		{Code: "vit", Min: 10, Max: 10},
		// Missing enr
	}

	result := combineAllAttributes(props, translator)

	if len(result) != 3 {
		t.Fatalf("Expected 3 properties (not combined), got %d", len(result))
	}

	if p := findProp(result, "all-stats"); p != nil {
		t.Error("should NOT have all-stats when an attribute is missing")
	}
}
