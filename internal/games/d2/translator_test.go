package d2

import (
	"testing"
)

func TestTranslate(t *testing.T) {
	translator := NewPropertyTranslator()

	tests := []struct {
		name     string
		prop     Property
		expected string
	}{
		// Simple value (min == max)
		{"all skills", Property{Code: "allskills", Min: 2, Max: 2}, "+2 To All Skills"},
		{"strength", Property{Code: "str", Min: 15, Max: 15}, "+15 To Strength"},
		{"negative value removes plus", Property{Code: "str", Min: -5, Max: -5}, "-5 To Strength"},

		// Value ranges (min != max)
		{"fire resist range", Property{Code: "res-fire", Min: 25, Max: 35}, "+25-35% Fire Resist"},
		{"swapped min/max normalized", Property{Code: "res-ltng", Min: 35, Max: 25}, "+25-35% Lightning Resist"},
		{"both negative range", Property{Code: "res-cold", Min: -10, Max: -5}, "-(5-10)% Cold Resist"},

		// Min/Max placeholders
		{"fire damage", Property{Code: "dmg-fire", Min: 15, Max: 40}, "Adds 15-40 Fire Damage"},
		{"cold damage", Property{Code: "dmg-cold", Min: 3, Max: 14}, "Adds 3-14 Cold Damage"},
		{"normal damage", Property{Code: "dmg-norm", Min: 50, Max: 100}, "Adds 50-100 Damage"},

		// Param replacement
		{"skill with param", Property{Code: "skill", Min: 3, Max: 3, Param: "Teleport"}, "+3 To Teleport"},
		{"aura when equipped", Property{Code: "aura", Min: 12, Max: 12, Param: "Fanaticism"}, "Level 12 Fanaticism Aura When Equipped"},
		{"poison damage with seconds", Property{Code: "dmg-pois", Min: 100, Max: 100, Param: "3"}, "+100 Poison Damage Over 3 Seconds"},

		// Skilltab resolution
		{"skilltab 0 (bow)", Property{Code: "skilltab", Min: 3, Max: 3, Param: "0"}, "+3 To Bow and Crossbow Skills"},
		{"skilltab 3 (fire)", Property{Code: "skilltab", Min: 2, Max: 2, Param: "3"}, "+2 To Fire Skills"},
		{"skilltab 21 (warlock eldritch)", Property{Code: "skilltab", Min: 2, Max: 2, Param: "21"}, "+2 To Eldritch Skills"},
		{"skilltab unknown tab", Property{Code: "skilltab", Min: 1, Max: 1, Param: "99"}, "+1 To 99"},

		// Charged/proc skills
		{"charged skill", Property{Code: "charged", Min: 30, Max: 12, Param: "Hydra"}, "Level 30 Hydra (12 Charges)"},
		{"hit-skill proc", Property{Code: "hit-skill", Min: 5, Max: 10, Param: "Frost Nova"}, "5% Chance To Cast Level 10 Frost Nova On Striking"},
		{"gethit-skill proc", Property{Code: "gethit-skill", Min: 10, Max: 3, Param: "Nova"}, "10% Chance To Cast Level 3 Nova When Struck"},
		{"kill-skill proc", Property{Code: "kill-skill", Min: 100, Max: 44, Param: "Ice Blast"}, "100% Chance To Cast Level 44 Ice Blast On Kill"},

		// Fixed text properties
		{"cannot be frozen", Property{Code: "nofreeze"}, "Cannot Be Frozen"},
		{"indestructible", Property{Code: "indestruct"}, "Indestructible"},
		{"prevent monster heal", Property{Code: "noheal"}, "Prevent Monster Heal"},
		{"knockback", Property{Code: "knock"}, "Knockback"},
		{"teleport", Property{Code: "teleport"}, "+1 To Teleport"},

		// Miscellaneous
		{"socketed", Property{Code: "sock", Min: 4, Max: 4}, "Socketed (4)"},
		{"enhanced damage", Property{Code: "dmg%", Min: 200, Max: 200}, "+200% Enhanced Damage"},
		{"enhanced defense", Property{Code: "ac%", Min: 100, Max: 100}, "+100% Enhanced Defense"},
		{"magic find", Property{Code: "mag%", Min: 50, Max: 50}, "+50% Better Chance Of Getting Magic Items"},
		{"crushing blow", Property{Code: "crush", Min: 33, Max: 33}, "33% Chance Of Crushing Blow"},
		{"life stolen", Property{Code: "lifesteal", Min: 8, Max: 8}, "8% Life Stolen Per Hit"},

		// Class skills
		{"amazon skills", Property{Code: "ama", Min: 2, Max: 2}, "+2 To Amazon Skill Levels"},
		{"sorceress skills", Property{Code: "sor", Min: 3, Max: 3}, "+3 To Sorceress Skill Levels"},
		{"warlock skills", Property{Code: "war", Min: 1, Max: 1}, "+1 To Warlock Skill Levels"},

		// Unknown code fallback
		{"unknown equal values", Property{Code: "unknown_stat", Min: 10, Max: 10}, "unknown_stat: 10"},
		{"unknown range values", Property{Code: "unknown_stat", Min: 5, Max: 15}, "unknown_stat: 5-15"},
		{"unknown swapped range", Property{Code: "unknown_stat", Min: 20, Max: 8}, "unknown_stat: 8-20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translator.Translate(tt.prop)
			if got != tt.expected {
				t.Errorf("Translate(%+v) = %q, want %q", tt.prop, got, tt.expected)
			}
		})
	}
}

func TestTranslate_PerLevel(t *testing.T) {
	translator := NewPropertyTranslator()

	tests := []struct {
		name     string
		prop     Property
		expected string
	}{
		{
			"hp/lvl raw=12 (1.5 per level)",
			Property{Code: "hp/lvl", Min: 12, Max: 12},
			"(1.5 Per Character Level) 1-148 To Life (Based On Character Level)",
		},
		{
			"hp/lvl raw=8 (integer 1 per level)",
			Property{Code: "hp/lvl", Min: 8, Max: 8},
			"(1 Per Character Level) 1-99 To Life (Based On Character Level)",
		},
		{
			"ac/lvl raw=16 (2 per level)",
			Property{Code: "ac/lvl", Min: 16, Max: 16},
			"(2 Per Character Level) 2-198 Defense (Based On Character Level)",
		},
		{
			"str/lvl raw=4 (0.5 per level)",
			Property{Code: "str/lvl", Min: 4, Max: 4},
			"(0.5 Per Character Level) 1-49 To Strength (Based On Character Level)",
		},
		{
			"uses max of min/max for raw value",
			Property{Code: "mana/lvl", Min: 4, Max: 12},
			"(1.5 Per Character Level) 1-148 To Mana (Based On Character Level)",
		},
		{
			"dmg%/lvl raw=8",
			Property{Code: "dmg%/lvl", Min: 8, Max: 8},
			"(1 Per Character Level) 1-99% Enhanced Damage (Based On Character Level)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translator.Translate(tt.prop)
			if got != tt.expected {
				t.Errorf("Translate(%+v) =\n  %q\nwant:\n  %q", tt.prop, got, tt.expected)
			}
		})
	}
}

func TestTranslateProperties_Dedup(t *testing.T) {
	translator := NewPropertyTranslator()

	t.Run("duplicate properties deduplicated", func(t *testing.T) {
		props := []Property{
			{Code: "nofreeze"},
			{Code: "nofreeze"},
		}
		results := translator.TranslateProperties(props)
		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d: %v", len(results), results)
		}
		if results[0] != "Cannot Be Frozen" {
			t.Errorf("result = %q, want %q", results[0], "Cannot Be Frozen")
		}
	})

	t.Run("distinct properties preserved", func(t *testing.T) {
		props := []Property{
			{Code: "nofreeze"},
			{Code: "indestruct"},
		}
		results := translator.TranslateProperties(props)
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d: %v", len(results), results)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		results := translator.TranslateProperties(nil)
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})
}

func TestEnrichProperty(t *testing.T) {
	translator := NewPropertyTranslator()

	tests := []struct {
		name             string
		prop             Property
		expectedHasRange bool
		checkDisplayText string // substring to verify
	}{
		{"normal stat with range", Property{Code: "str", Min: 10, Max: 20}, true, "+10-20 To Strength"},
		{"normal stat no range", Property{Code: "str", Min: 15, Max: 15}, false, "+15 To Strength"},
		{"fixed code hit-skill always false", Property{Code: "hit-skill", Min: 5, Max: 10, Param: "Nova"}, false, "5% Chance To Cast Level 10 Nova On Striking"},
		{"fixed code dmg-fire always false", Property{Code: "dmg-fire", Min: 10, Max: 50}, false, "Adds 10-50 Fire Damage"},
		{"fixed code sunder always false", Property{Code: "pierce-immunity-cold"}, false, "Monster Cold Immunity is Sundered"},
		{"fixed code charged always false", Property{Code: "charged", Min: 20, Max: 5, Param: "Hydra"}, false, "Level 20 Hydra (5 Charges)"},
		{"per-level code always false", Property{Code: "hp/lvl", Min: 12, Max: 12}, false, "(1.5 Per Character Level) 1-148 To Life (Based On Character Level)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := tt.prop
			translator.EnrichProperty(&prop)

			if prop.HasRange != tt.expectedHasRange {
				t.Errorf("HasRange = %v, want %v", prop.HasRange, tt.expectedHasRange)
			}
			if prop.DisplayText == "" {
				t.Error("DisplayText is empty after enrichment")
			}
			if tt.checkDisplayText != "" && prop.DisplayText != tt.checkDisplayText {
				t.Errorf("DisplayText = %q, want %q", prop.DisplayText, tt.checkDisplayText)
			}
		})
	}
}

func TestEnrichProperties(t *testing.T) {
	translator := NewPropertyTranslator()

	original := []Property{
		{Code: "str", Min: 10, Max: 20},
		{Code: "nofreeze"},
	}

	enriched := translator.EnrichProperties(original)

	if len(enriched) != 2 {
		t.Fatalf("expected 2 enriched properties, got %d", len(enriched))
	}

	// Verify enriched properties have DisplayText
	if enriched[0].DisplayText == "" {
		t.Error("enriched[0].DisplayText is empty")
	}
	if enriched[1].DisplayText == "" {
		t.Error("enriched[1].DisplayText is empty")
	}

	// Verify original is not mutated
	if original[0].DisplayText != "" {
		t.Errorf("original[0].DisplayText was mutated: %q", original[0].DisplayText)
	}
	if original[1].DisplayText != "" {
		t.Errorf("original[1].DisplayText was mutated: %q", original[1].DisplayText)
	}
}

func TestGetDisplayName(t *testing.T) {
	translator := NewPropertyTranslator()

	tests := []struct {
		code     string
		expected string
	}{
		{"allskills", "All Skills"},
		{"hit-skill", "Chance To Cast On Striking"},
		{"gethit-skill", "Chance To Cast When Struck"},
		{"charged", "Charges"},
		{"skill", "Skill Bonus"},
		{"randclassskill", "Random Class Skills"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := translator.GetDisplayName(tt.code)
			if got != tt.expected {
				t.Errorf("GetDisplayName(%q) = %q, want %q", tt.code, got, tt.expected)
			}
		})
	}

	t.Run("unknown code passthrough", func(t *testing.T) {
		got := translator.GetDisplayName("totally_unknown")
		if got != "totally_unknown" {
			t.Errorf("GetDisplayName(%q) = %q, want passthrough", "totally_unknown", got)
		}
	})
}

func TestHasRange(t *testing.T) {
	translator := NewPropertyTranslator()

	tests := []struct {
		name     string
		prop     Property
		expected bool
	}{
		{"different min/max", Property{Min: 10, Max: 20}, true},
		{"equal min/max", Property{Min: 15, Max: 15}, false},
		{"both zero", Property{Min: 0, Max: 0}, false},
		{"negative range", Property{Min: -5, Max: 5}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translator.HasRange(tt.prop)
			if got != tt.expected {
				t.Errorf("HasRange(%+v) = %v, want %v", tt.prop, got, tt.expected)
			}
		})
	}
}
