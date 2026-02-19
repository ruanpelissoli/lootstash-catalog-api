package d2

import (
	"strings"
	"testing"
)

func TestCategories(t *testing.T) {
	cats := Categories()

	if len(cats) != 14 {
		t.Errorf("expected 14 categories, got %d", len(cats))
	}

	// Verify first and last entries
	if cats[0].Code != "helm" || cats[0].Name != "Helms" {
		t.Errorf("first category: got {%s, %s}, want {helm, Helms}", cats[0].Code, cats[0].Name)
	}
	if cats[len(cats)-1].Code != "misc" || cats[len(cats)-1].Name != "Miscellaneous" {
		t.Errorf("last category: got {%s, %s}, want {misc, Miscellaneous}", cats[len(cats)-1].Code, cats[len(cats)-1].Name)
	}

	// Check uniqueness and non-empty fields
	seen := make(map[string]bool)
	for _, c := range cats {
		if c.Code == "" {
			t.Error("found category with empty Code")
		}
		if c.Name == "" {
			t.Error("found category with empty Name")
		}
		if seen[c.Code] {
			t.Errorf("duplicate category code: %s", c.Code)
		}
		seen[c.Code] = true
	}
}

func TestRarities(t *testing.T) {
	rars := Rarities()

	if len(rars) != 7 {
		t.Errorf("expected 7 rarities, got %d", len(rars))
	}

	// Verify first entry
	first := rars[0]
	if first.Code != "normal" || first.Name != "Normal" || first.Color != "#FFFFFF" {
		t.Errorf("first rarity: got {%s, %s, %s}, want {normal, Normal, #FFFFFF}", first.Code, first.Name, first.Color)
	}

	// Check uniqueness and format
	seen := make(map[string]bool)
	for _, r := range rars {
		if r.Code == "" {
			t.Error("found rarity with empty Code")
		}
		if r.Name == "" {
			t.Error("found rarity with empty Name")
		}
		if !strings.HasPrefix(r.Color, "#") || len(r.Color) != 7 {
			t.Errorf("rarity %s has invalid color format: %s (expected #RRGGBB)", r.Code, r.Color)
		}
		if seen[r.Code] {
			t.Errorf("duplicate rarity code: %s", r.Code)
		}
		seen[r.Code] = true
	}
}

func TestFilterableStats(t *testing.T) {
	stats := FilterableStats()

	if len(stats) == 0 {
		t.Fatal("FilterableStats returned empty slice")
	}

	seen := make(map[string]bool)
	for _, s := range stats {
		if s.Code == "" {
			t.Error("found stat with empty Code")
		}
		if s.Name == "" {
			t.Errorf("stat %s has empty Name", s.Code)
		}
		if s.Category == "" {
			t.Errorf("stat %s has empty Category", s.Code)
		}
		if seen[s.Code] {
			t.Errorf("duplicate stat code: %s", s.Code)
		}
		seen[s.Code] = true
	}
}

func TestStatCategories(t *testing.T) {
	if len(StatCategories) != 17 {
		t.Errorf("expected 17 stat categories, got %d", len(StatCategories))
	}

	required := []string{"Skills", "Resistances", "Damage", "Other"}
	catSet := make(map[string]bool)
	for _, c := range StatCategories {
		catSet[c] = true
	}
	for _, r := range required {
		if !catSet[r] {
			t.Errorf("StatCategories missing required category: %s", r)
		}
	}
}
