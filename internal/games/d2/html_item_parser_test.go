package d2

import (
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
