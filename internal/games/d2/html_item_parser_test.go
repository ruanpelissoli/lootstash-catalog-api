package d2

import (
	"strings"
	"testing"
)

func TestNormalizeItemName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase and trim", "  Harlequin Crest  ", "harlequin crest"},
		{"already lowercase", "shako", "shako"},
		{"left single curly quote", "Trang\u2018Oul", "trang'oul"},
		{"right single curly quote", "Trang\u2019Oul", "trang'oul"},
		{"left double curly quote", "\u201CHello\u201D", "\"hello\""},
		{"mixed case and quotes", "  Bartuc\u2019s Cut-Throat  ", "bartuc's cut-throat"},
		{"empty string", "", ""},
		{"only whitespace", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeItemName(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeItemName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNormalizeImagePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"already correct", "/styles/images/item.png", "/styles/images/item.png"},
		{"backslashes", "\\styles\\images\\item.png", "/styles/images/item.png"},
		{"no leading slash", "styles/images/item.png", "/styles/images/item.png"},
		{"mixed slashes", "styles\\images/item.png", "/styles/images/item.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeImagePath(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeImagePath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsBaseStatLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"defense label", "Defense: 100-150", true},
		{"req strength", "Req Strength: 50", true},
		{"max sockets", "Max sockets: 4", true},
		{"1H damage", "1H damage: 10-20", true},
		{"2H damage", "2H damage: 30-50", true},
		{"durability", "Durability: 40", true},
		{"req level", "Req level: 65", true},
		{"base speed", "Base speed: -10", true},
		{"not a label", "+25% Enhanced Defense", false},
		{"empty string", "", false},
		{"regular property", "+2 To All Skills", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBaseStatLabel(tt.input)
			if got != tt.expected {
				t.Errorf("isBaseStatLabel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsJustNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"positive integer", "123", true},
		{"zero", "0", true},
		{"negative integer", "-5", true},
		{"letters", "abc", false},
		{"empty string", "", false},
		{"decimal", "12.5", false},
		{"whitespace only", "  ", false},
		{"number with spaces", " 42 ", true}, // TrimSpace applied first
		{"mixed", "12a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isJustNumber(tt.input)
			if got != tt.expected {
				t.Errorf("isJustNumber(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractSlugFromHref(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"standard href", "/base/blasphemous-grimoire-t1673953.html", "blasphemous-grimoire"},
		{"short id", "/base/shako-t123.html", "shako"},
		{"no -tNNN suffix", "/base/item-no-id.html", "item-no-id.html"},
		{"no /base/ prefix", "some-item-t999.html", "some-item"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSlugFromHref(tt.input)
			if got != tt.expected {
				t.Errorf("extractSlugFromHref(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseRunewordPropertiesByType(t *testing.T) {
	parser := NewHTMLItemParser()
	runewords, err := parser.ParseRunewordsFile("../../../catalogs/d2/pages/runewords.html")
	if err != nil {
		t.Fatalf("ParseRunewordsFile failed: %v", err)
	}

	// Find Spirit
	var spirit *HTMLParsedRuneword
	for i := range runewords {
		if runewords[i].Name == "Spirit" {
			spirit = &runewords[i]
			break
		}
	}
	if spirit == nil {
		t.Fatal("Spirit runeword not found")
	}

	// Spirit should have per-type properties (Swords vs Shields have different stats)
	if spirit.PropertiesByType == nil {
		t.Fatal("Spirit should have PropertiesByType set (Swords and Shields have different stats)")
	}
	if len(spirit.Properties) > 0 {
		t.Error("Spirit.Properties should be empty when PropertiesByType is set")
	}

	// Should have exactly 2 types
	if len(spirit.PropertiesByType) != 2 {
		t.Fatalf("Spirit should have 2 type entries, got %d", len(spirit.PropertiesByType))
	}

	swordProps, hasSwords := spirit.PropertiesByType["4 socket Swords"]
	shieldProps, hasShields := spirit.PropertiesByType["4 socket Shields"]
	if !hasSwords {
		t.Error("Spirit should have '4 socket Swords' in PropertiesByType")
	}
	if !hasShields {
		t.Error("Spirit should have '4 socket Shields' in PropertiesByType")
	}

	// Sword and shield properties should differ
	if stringSlicesEqual(swordProps, shieldProps) {
		t.Error("Spirit Swords and Shields properties should be different")
	}

	// Swords should have Lightning Damage, Shields should not
	hasLightningDmg := false
	for _, line := range swordProps {
		if strings.Contains(line, "Lightning Damage") {
			hasLightningDmg = true
			break
		}
	}
	if !hasLightningDmg {
		t.Error("Spirit Swords should have Lightning Damage property")
	}

	// Shields should have Cold Resist, Swords should not
	hasColdResist := false
	for _, line := range shieldProps {
		if strings.Contains(line, "Cold Resist") {
			hasColdResist = true
			break
		}
	}
	if !hasColdResist {
		t.Error("Spirit Shields should have Cold Resist property")
	}

	// Find a runeword with same stats for all types (e.g., Rhyme - 2 socket Shields only)
	var rhyme *HTMLParsedRuneword
	for i := range runewords {
		if runewords[i].Name == "Rhyme" {
			rhyme = &runewords[i]
			break
		}
	}
	if rhyme != nil {
		// Single type runewords should use flat Properties, not PropertiesByType
		if rhyme.PropertiesByType != nil {
			t.Error("Rhyme should not have PropertiesByType (single base type)")
		}
		if len(rhyme.Properties) == 0 {
			t.Error("Rhyme should have Properties set")
		}
	}
}

func TestClassSpecificFromTypeTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected string
	}{
		{"amazon weapons", []string{"Amazon Weapons", "Javelins"}, "amazon"},
		{"paladin shields", []string{"Paladin Shields"}, "paladin"},
		{"necromancer shields", []string{"Necromancer Shields"}, "necromancer"},
		{"shrunken heads", []string{"Shrunken Heads"}, "necromancer"},
		{"barbarian helms", []string{"Barbarian Helms"}, "barbarian"},
		{"druid pelts", []string{"Druid Pelts"}, "druid"},
		{"claws", []string{"Claws"}, "assassin"},
		{"grimoires", []string{"Grimoires"}, "warlock"},
		{"orbs", []string{"Orbs"}, "sorceress"},
		{"no class specific", []string{"Swords", "Weapons"}, ""},
		{"empty slice", []string{}, ""},
		{"nil slice", nil, ""},
		{"class tag later in list", []string{"Weapons", "Melee Weapons", "Claws"}, "assassin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classSpecificFromTypeTags(tt.tags)
			if got != tt.expected {
				t.Errorf("classSpecificFromTypeTags(%v) = %q, want %q", tt.tags, got, tt.expected)
			}
		})
	}
}
