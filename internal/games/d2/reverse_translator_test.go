package d2

import (
	"testing"
)

func TestParseValueStr(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedMin int
		expectedMax int
	}{
		{"simple integer", "25", 25, 25},
		{"positive range", "25-35", 25, 35},
		{"plus prefix single", "+25", 25, 25},
		{"plus prefix range", "+25-35", 25, 35},
		{"negative parenthesized range", "-(5-10)", -10, -5},
		{"simple negative", "-25", -25, -25},
		{"negative range dash separated", "-90-70", -90, -70},
		{"zero", "0", 0, 0},
		{"empty string", "", 0, 0},
		{"whitespace only", "  ", 0, 0},
		{"negative range reversed order", "-70-90", -90, -70},
		{"single digit", "5", 5, 5},
		{"large numbers", "1000-2000", 1000, 2000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := parseValueStr(tt.input)
			if min != tt.expectedMin || max != tt.expectedMax {
				t.Errorf("parseValueStr(%q) = (%d, %d), want (%d, %d)", tt.input, min, max, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestReverseTranslate(t *testing.T) {
	rt := NewReverseTranslator()

	tests := []struct {
		name          string
		displayText   string
		expectedCode  string
		expectedMin   int
		expectedMax   int
		expectedParam string
	}{
		// Simple value properties
		{"all skills", "+2 To All Skills", "allskills", 2, 2, ""},
		{"strength", "+15 To Strength", "str", 15, 15, ""},
		{"enhanced damage", "+200% Enhanced Damage", "dmg%", 200, 200, ""},
		{"fire resist current format", "+30% Fire Resist", "res-fire", 30, 30, ""},
		{"fire resist range", "+25-35% Fire Resist", "res-fire", 25, 35, ""},
		{"all resistances", "+60 All Resistances", "res-all", 60, 60, ""},
		{"magic find", "+50% Better Chance Of Getting Magic Items", "mag%", 50, 50, ""},
		{"life stolen", "8% Life Stolen Per Hit", "lifesteal", 8, 8, ""},
		{"crushing blow", "33% Chance Of Crushing Blow", "crush", 33, 33, ""},

		// Legacy formats
		{"legacy fire resist", "Fire Resist +30%", "res-fire", 30, 30, ""},
		{"legacy regen mana", "Regenerate Mana 15%", "regen-mana", 15, 15, ""},
		{"legacy all resistances", "All Resistances +60", "res-all", 60, 60, ""},
		{"legacy damage reduced", "Damage Reduced By 7", "red-dmg", 7, 7, ""},
		{"legacy requirements", "Requirements -15%", "ease", 15, 15, ""},

		// Fixed-text properties
		{"cannot be frozen", "Cannot Be Frozen", "nofreeze", 0, 0, ""},
		{"indestructible", "Indestructible", "indestruct", 0, 0, ""},
		{"prevent monster heal", "Prevent Monster Heal", "noheal", 0, 0, ""},
		{"knockback", "Knockback", "knock", 0, 0, ""},

		// Case insensitive
		{"case insensitive frozen", "cannot be frozen", "nofreeze", 0, 0, ""},
		{"case insensitive indestructible", "indestructible", "indestruct", 0, 0, ""},

		// Sunder charms
		{"sunder cold", "Monster Cold Immunity is Sundered", "pierce-immunity-cold", 0, 0, ""},
		{"sunder fire", "Monster Fire Immunity is Sundered", "pierce-immunity-fire", 0, 0, ""},
		{"sunder lightning", "Monster Lightning Immunity is Sundered", "pierce-immunity-light", 0, 0, ""},
		{"sunder poison", "Monster Poison Immunity is Sundered", "pierce-immunity-poison", 0, 0, ""},
		{"sunder physical", "Monster Physical Immunity is Sundered", "pierce-immunity-damage", 0, 0, ""},

		// Param-based properties
		// Note: "skill" and "oskill" share the same format template, so the reverse translator
		// may return either code. Both are valid skill bonus codes.
		{"hit-skill proc", "5% Chance To Cast Level 10 Frost Nova On Striking", "hit-skill", 5, 10, "Frost Nova"},
		{"gethit-skill proc", "10% Chance To Cast Level 3 Nova When Struck", "gethit-skill", 10, 3, "Nova"},
		{"charged skill", "Level 30 Hydra (12 Charges)", "charged", 30, 12, "Hydra"},

		// Min/Max based
		{"adds fire damage", "Adds 15-40 Fire Damage", "dmg-fire", 15, 40, ""},
		{"adds cold damage", "Adds 3-14 Cold Damage", "dmg-cold", 3, 14, ""},

		// Empty and unknown
		{"empty string", "", "raw", 0, 0, ""},
		{"unrecognized text", "Some Unknown Property", "raw", 0, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := rt.ReverseTranslate(tt.displayText)

			if prop.Code != tt.expectedCode {
				t.Errorf("Code = %q, want %q", prop.Code, tt.expectedCode)
			}
			if prop.Min != tt.expectedMin {
				t.Errorf("Min = %d, want %d", prop.Min, tt.expectedMin)
			}
			if prop.Max != tt.expectedMax {
				t.Errorf("Max = %d, want %d", prop.Max, tt.expectedMax)
			}
			if prop.Param != tt.expectedParam {
				t.Errorf("Param = %q, want %q", prop.Param, tt.expectedParam)
			}
		})
	}
}

func TestReverseTranslate_PerLevel(t *testing.T) {
	rt := NewReverseTranslator()

	tests := []struct {
		name        string
		displayText string
		expectedCode string
		expectedMin  int
		expectedMax  int
	}{
		{
			"parenthesized per-level hp",
			"(1.5 Per Character Level) 1-148 To Life (Based On Character Level)",
			"hp/lvl", 12, 12,
		},
		{
			"parenthesized per-level defense",
			"(2 Per Character Level) 2-198 Defense (Based On Character Level)",
			"ac/lvl", 16, 16,
		},
		{
			"parenthesized per-level mana",
			"(1 Per Character Level) 1-99 To Mana (Based On Character Level)",
			"mana/lvl", 8, 8,
		},
		{
			"simple per-level format",
			"+12 To Life (Based On Character Level)",
			"hp/lvl", 12, 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := rt.ReverseTranslate(tt.displayText)

			if prop.Code != tt.expectedCode {
				t.Errorf("Code = %q, want %q", prop.Code, tt.expectedCode)
			}
			if prop.Min != tt.expectedMin {
				t.Errorf("Min = %d, want %d", prop.Min, tt.expectedMin)
			}
			if prop.Max != tt.expectedMax {
				t.Errorf("Max = %d, want %d", prop.Max, tt.expectedMax)
			}
		})
	}
}

func TestReverseTranslate_SkillBonus(t *testing.T) {
	rt := NewReverseTranslator()

	// "skill" and "oskill" share the same display format "+{value} To {param}",
	// so the reverse translator returns whichever matches first. Both are valid.
	tests := []struct {
		name          string
		displayText   string
		expectedMin   int
		expectedMax   int
		expectedParam string
	}{
		{"simple skill bonus", "+3 To Teleport", 3, 3, "Teleport"},
		{"class suffix stripped", "+3 To Teleport (Warlock only)", 3, 3, "Teleport"},
		{"amazon class suffix", "+2 To Multiple Shot (Amazon Only)", 2, 2, "Multiple Shot"},
		{"multi-word skill", "+1 To Frozen Orb", 1, 1, "Frozen Orb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := rt.ReverseTranslate(tt.displayText)

			validCodes := map[string]bool{"skill": true, "oskill": true}
			if !validCodes[prop.Code] {
				t.Errorf("Code = %q, want skill or oskill", prop.Code)
			}
			if prop.Min != tt.expectedMin {
				t.Errorf("Min = %d, want %d", prop.Min, tt.expectedMin)
			}
			if prop.Max != tt.expectedMax {
				t.Errorf("Max = %d, want %d", prop.Max, tt.expectedMax)
			}
			if prop.Param != tt.expectedParam {
				t.Errorf("Param = %q, want %q", prop.Param, tt.expectedParam)
			}
		})
	}
}

func TestReverseTranslate_SkilltabResolution(t *testing.T) {
	rt := NewReverseTranslator()

	// The reverse translator may resolve skilltab patterns as either "skilltab" or "skill"/"oskill"
	// because the display formats overlap ("+{value} To {skilltab}" vs "+{value} To {param}").
	// What matters is the value is extracted correctly.
	tests := []struct {
		name          string
		displayText   string
		expectedMin   int
		expectedMax   int
	}{
		{"bow and crossbow", "+3 To Bow and Crossbow Skills", 3, 3},
		{"fire skills", "+2 To Fire Skills", 2, 2},
		{"cold skills", "+1 To Cold Skills", 1, 1},
		{"eldritch skills", "+1 To Eldritch Skills", 1, 1},
		{"traps", "+3 To Traps", 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := rt.ReverseTranslate(tt.displayText)

			// Code should be one of the skill-related codes
			validCodes := map[string]bool{"skilltab": true, "skill": true, "oskill": true}
			if !validCodes[prop.Code] {
				t.Errorf("Code = %q, want one of skilltab/skill/oskill", prop.Code)
			}
			if prop.Min != tt.expectedMin {
				t.Errorf("Min = %d, want %d", prop.Min, tt.expectedMin)
			}
			if prop.Max != tt.expectedMax {
				t.Errorf("Max = %d, want %d", prop.Max, tt.expectedMax)
			}
		})
	}
}

func TestReverseTranslateLines(t *testing.T) {
	rt := NewReverseTranslator()

	t.Run("mixed lines with empties", func(t *testing.T) {
		lines := []string{
			"+2 To All Skills",
			"",
			"Cannot Be Frozen",
			"  ",
		}
		props := rt.ReverseTranslateLines(lines)
		if len(props) != 2 {
			t.Fatalf("expected 2 properties, got %d", len(props))
		}
		if props[0].Code != "allskills" {
			t.Errorf("props[0].Code = %q, want %q", props[0].Code, "allskills")
		}
		if props[1].Code != "nofreeze" {
			t.Errorf("props[1].Code = %q, want %q", props[1].Code, "nofreeze")
		}
	})

	t.Run("nil input", func(t *testing.T) {
		props := rt.ReverseTranslateLines(nil)
		if len(props) != 0 {
			t.Errorf("expected 0 properties, got %d", len(props))
		}
	})

	t.Run("all empty lines", func(t *testing.T) {
		lines := []string{"", "  ", ""}
		props := rt.ReverseTranslateLines(lines)
		if len(props) != 0 {
			t.Errorf("expected 0 properties, got %d", len(props))
		}
	})
}

func TestRoundTrip_TranslateReverseTranslate(t *testing.T) {
	translator := NewPropertyTranslator()
	rt := NewReverseTranslator()

	tests := []struct {
		name         string
		original     Property
		expectedCode string
		expectedMin  int
		expectedMax  int
	}{
		{"all skills", Property{Code: "allskills", Min: 2, Max: 2}, "allskills", 2, 2},
		{"strength", Property{Code: "str", Min: 15, Max: 15}, "str", 15, 15},
		{"cannot be frozen", Property{Code: "nofreeze"}, "nofreeze", 0, 0},
		{"fire damage", Property{Code: "dmg-fire", Min: 10, Max: 50}, "dmg-fire", 10, 50},
		{"fire resist", Property{Code: "res-fire", Min: 30, Max: 30}, "res-fire", 30, 30},
		{"enhanced damage", Property{Code: "dmg%", Min: 200, Max: 200}, "dmg%", 200, 200},
		{"all resistances", Property{Code: "res-all", Min: 50, Max: 50}, "res-all", 50, 50},
		{"sunder cold", Property{Code: "pierce-immunity-cold"}, "pierce-immunity-cold", 0, 0},
		{"indestructible", Property{Code: "indestruct"}, "indestruct", 0, 0},
		{"magic find", Property{Code: "mag%", Min: 35, Max: 35}, "mag%", 35, 35},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Forward: translate property to display text
			displayText := translator.Translate(tt.original)
			if displayText == "" {
				t.Fatal("Translate returned empty string")
			}

			// Reverse: convert display text back to property
			recovered := rt.ReverseTranslate(displayText)

			if recovered.Code != tt.expectedCode {
				t.Errorf("round-trip Code: %q -> %q -> %q, want %q",
					tt.original.Code, displayText, recovered.Code, tt.expectedCode)
			}
			if recovered.Min != tt.expectedMin {
				t.Errorf("round-trip Min: %d -> %q -> %d, want %d",
					tt.original.Min, displayText, recovered.Min, tt.expectedMin)
			}
			if recovered.Max != tt.expectedMax {
				t.Errorf("round-trip Max: %d -> %q -> %d, want %d",
					tt.original.Max, displayText, recovered.Max, tt.expectedMax)
			}
		})
	}
}

func TestRoundTrip_SkillProperties(t *testing.T) {
	translator := NewPropertyTranslator()
	rt := NewReverseTranslator()

	tests := []struct {
		name          string
		original      Property
		expectedCode  string
		expectedParam string
		expectedMin   int
		expectedMax   int
	}{
		{
			"hit-skill proc",
			Property{Code: "hit-skill", Min: 5, Max: 10, Param: "Frost Nova"},
			"hit-skill", "Frost Nova", 5, 10,
		},
		{
			"gethit-skill proc",
			Property{Code: "gethit-skill", Min: 10, Max: 3, Param: "Nova"},
			"gethit-skill", "Nova", 10, 3,
		},
		{
			"skill bonus",
			Property{Code: "skill", Min: 3, Max: 3, Param: "Teleport"},
			// skill and oskill share the same format, reverse translator may return either
			"oskill", "Teleport", 3, 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			displayText := translator.Translate(tt.original)
			recovered := rt.ReverseTranslate(displayText)

			// For skill/oskill, accept either code since they share the same format
			if tt.expectedCode == "oskill" || tt.expectedCode == "skill" {
				validCodes := map[string]bool{"skill": true, "oskill": true}
				if !validCodes[recovered.Code] {
					t.Errorf("Code: %q, want skill or oskill (via %q)", recovered.Code, displayText)
				}
			} else if recovered.Code != tt.expectedCode {
				t.Errorf("Code: %q, want %q (via %q)", recovered.Code, tt.expectedCode, displayText)
			}
			if recovered.Param != tt.expectedParam {
				t.Errorf("Param: %q, want %q (via %q)", recovered.Param, tt.expectedParam, displayText)
			}
			if recovered.Min != tt.expectedMin {
				t.Errorf("Min: %d, want %d (via %q)", recovered.Min, tt.expectedMin, displayText)
			}
			if recovered.Max != tt.expectedMax {
				t.Errorf("Max: %d, want %d (via %q)", recovered.Max, tt.expectedMax, displayText)
			}
		})
	}
}
