package d2

import (
	"testing"
)

func TestGenerateBaseCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"two words", "Blasphemous Grimoire", "blagr"},
		{"single short word", "Cap", "cap"},
		{"single long word", "Shako", "shak"},
		{"single exactly 4 chars", "Helm", "helm"},
		{"apostrophe removed", "Trang'Oul's", "tran"},
		{"hyphen removed becomes single word", "Full-Plate", "full"},
		{"three words", "Ancient Armor Defense", "ancarde"},
		{"two words first word short", "An Axe", "anax"},
		{"many words truncated to 8", "The Ancient Evil Sword Of Doom", "theanevs"},
		{"single word exactly 4", "Ring", "ring"},
		{"single two char word", "Ax", "ax"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateBaseCode(tt.input)
			if got != tt.expected {
				t.Errorf("generateBaseCode(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSplitOrAlternatives(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"no marker", "+25% Fire Resist", nil},
		{"single OR marker", "A ||OR|| B", []string{"A", "B"}},
		{"multiple OR markers", "A ||OR|| B ||OR|| C", []string{"A", "B", "C"}},
		{"empty first part", " ||OR|| B", nil},
		{"empty second part", "A ||OR|| ", nil},
		{"only marker", " ||OR|| ", nil},
		{"sunder charm example", "Monster Cold Immunity is Sundered ||OR|| Monster Fire Immunity is Sundered",
			[]string{"Monster Cold Immunity is Sundered", "Monster Fire Immunity is Sundered"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitOrAlternatives(tt.input)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("splitOrAlternatives(%q) = %v, want nil", tt.input, got)
				}
				return
			}
			if len(got) != len(tt.expected) {
				t.Errorf("splitOrAlternatives(%q) returned %d parts, want %d", tt.input, len(got), len(tt.expected))
				return
			}
			for i := range tt.expected {
				if got[i] != tt.expected[i] {
					t.Errorf("splitOrAlternatives(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestSplitOrBonuses(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"no separator", "+25% To Life", []string{"+25% To Life"}},
		{"with or newline separator", "A or \nB", []string{"A", "B"}},
		{"empty string", "", []string{""}},
		{"multiple separators", "A or \nB or \nC", []string{"A", "B", "C"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitOrBonuses(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("splitOrBonuses(%q) returned %d parts, want %d: %v", tt.input, len(got), len(tt.expected), got)
				return
			}
			for i := range tt.expected {
				if got[i] != tt.expected[i] {
					t.Errorf("splitOrBonuses(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestHtmlTypeNameToCode(t *testing.T) {
	tests := []struct {
		typeName string
		expected string
	}{
		{"Body Armor", "tors"},
		{"Swords", "swor"},
		{"Grimoires", "grim"},
		{"Claws", "h2h"},
		{"Paladin Shields", "ashd"},
		{"Druid Pelts", "pelt"},
		{"Bows", "bow"},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			got := htmlTypeNameToCode[tt.typeName]
			if got != tt.expected {
				t.Errorf("htmlTypeNameToCode[%q] = %q, want %q", tt.typeName, got, tt.expected)
			}
		})
	}

	// Verify missing key returns empty string
	if got := htmlTypeNameToCode["NonExistent"]; got != "" {
		t.Errorf("htmlTypeNameToCode[NonExistent] = %q, want empty string", got)
	}
}
