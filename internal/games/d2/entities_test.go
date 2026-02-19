package d2

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPropertyJSON(t *testing.T) {
	t.Run("full property round-trip", func(t *testing.T) {
		prop := Property{
			Code:        "allskills",
			Min:         2,
			Max:         2,
			DisplayText: "+2 To All Skills",
			HasRange:    false,
		}

		data, err := json.Marshal(prop)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}

		var got Property
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}

		if got.Code != prop.Code || got.Min != prop.Min || got.Max != prop.Max || got.DisplayText != prop.DisplayText {
			t.Errorf("round-trip mismatch: got %+v, want %+v", got, prop)
		}
	})

	t.Run("omitempty Param absent when empty", func(t *testing.T) {
		prop := Property{Code: "str", Min: 10, Max: 10}
		data, _ := json.Marshal(prop)
		if strings.Contains(string(data), `"param"`) {
			t.Errorf("expected Param to be omitted, got: %s", data)
		}
	})

	t.Run("omitempty DisplayText absent when empty", func(t *testing.T) {
		prop := Property{Code: "str", Min: 10, Max: 10}
		data, _ := json.Marshal(prop)
		if strings.Contains(string(data), `"displayText"`) {
			t.Errorf("expected DisplayText to be omitted, got: %s", data)
		}
	})

	t.Run("omitempty HasRange absent when false", func(t *testing.T) {
		prop := Property{Code: "str", Min: 10, Max: 10, HasRange: false}
		data, _ := json.Marshal(prop)
		if strings.Contains(string(data), `"hasRange"`) {
			t.Errorf("expected HasRange to be omitted when false, got: %s", data)
		}
	})

	t.Run("HasRange present when true", func(t *testing.T) {
		prop := Property{Code: "str", Min: 10, Max: 20, HasRange: true}
		data, _ := json.Marshal(prop)
		if !strings.Contains(string(data), `"hasRange":true`) {
			t.Errorf("expected HasRange to be present, got: %s", data)
		}
	})

	t.Run("Param present when set", func(t *testing.T) {
		prop := Property{Code: "skill", Param: "Teleport", Min: 3, Max: 3}
		data, _ := json.Marshal(prop)
		if !strings.Contains(string(data), `"param":"Teleport"`) {
			t.Errorf("expected Param to be present, got: %s", data)
		}
	})

	t.Run("Alternatives round-trip", func(t *testing.T) {
		prop := Property{
			Code: "or-group",
			Alternatives: []Property{
				{Code: "res-fire", Min: 25, Max: 25, DisplayText: "+25% Fire Resist"},
				{Code: "res-cold", Min: 30, Max: 30, DisplayText: "+30% Cold Resist"},
			},
		}
		data, err := json.Marshal(prop)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}

		var got Property
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}

		if len(got.Alternatives) != 2 {
			t.Fatalf("expected 2 alternatives, got %d", len(got.Alternatives))
		}
		if got.Alternatives[0].Code != "res-fire" {
			t.Errorf("alt[0].Code = %q, want %q", got.Alternatives[0].Code, "res-fire")
		}
		if got.Alternatives[1].Code != "res-cold" {
			t.Errorf("alt[1].Code = %q, want %q", got.Alternatives[1].Code, "res-cold")
		}
	})

	t.Run("Alternatives omitted when nil", func(t *testing.T) {
		prop := Property{Code: "str", Min: 10, Max: 10}
		data, _ := json.Marshal(prop)
		if strings.Contains(string(data), `"alternatives"`) {
			t.Errorf("expected alternatives to be omitted, got: %s", data)
		}
	})
}

func TestUniqueItemJSON(t *testing.T) {
	item := UniqueItem{
		ID:       1,
		IndexID:  100,
		Name:     "Shako",
		BaseCode: "uap",
		BaseName: "Shako",
		Level:    69,
		LevelReq: 62,
		Enabled:  true,
		Properties: []Property{
			{Code: "allskills", Min: 2, Max: 2, DisplayText: "+2 To All Skills"},
			{Code: "hp", Min: 99, Max: 141, DisplayText: "+99-141 To Life", HasRange: true},
		},
		ImageURL: "https://example.com/shako.png",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var got UniqueItem
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if got.Name != "Shako" {
		t.Errorf("Name = %q, want %q", got.Name, "Shako")
	}
	if len(got.Properties) != 2 {
		t.Errorf("len(Properties) = %d, want 2", len(got.Properties))
	}
	if got.ImageURL != item.ImageURL {
		t.Errorf("ImageURL = %q, want %q", got.ImageURL, item.ImageURL)
	}

	// Verify omitempty: FirstLadderSeason should be absent when nil
	if strings.Contains(string(data), `"first_ladder_season"`) {
		t.Error("expected first_ladder_season to be omitted when nil")
	}
}

func TestSetItemJSON(t *testing.T) {
	item := SetItem{
		ID:       1,
		Name:     "Tal Rasha's Horadric Crest",
		SetName:  "Tal Rasha's Wrappings",
		BaseCode: "msk",
		Properties: []Property{
			{Code: "lifesteal", Min: 10, Max: 10, DisplayText: "10% Life Stolen Per Hit"},
		},
		BonusProperties: []Property{
			{Code: "res-all", Min: 15, Max: 15, DisplayText: "+15 All Resistances"},
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var got SetItem
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if got.SetName != "Tal Rasha's Wrappings" {
		t.Errorf("SetName = %q, want %q", got.SetName, "Tal Rasha's Wrappings")
	}
	if len(got.Properties) != 1 {
		t.Errorf("len(Properties) = %d, want 1", len(got.Properties))
	}
	if len(got.BonusProperties) != 1 {
		t.Errorf("len(BonusProperties) = %d, want 1", len(got.BonusProperties))
	}

	// Verify JSON field names
	if !strings.Contains(string(data), `"set_name"`) {
		t.Error("expected set_name field in JSON")
	}
	if !strings.Contains(string(data), `"bonus_properties"`) {
		t.Error("expected bonus_properties field in JSON")
	}
}

func TestRunewordJSON(t *testing.T) {
	item := Runeword{
		ID:          1,
		Name:        "spirit",
		DisplayName: "Spirit",
		Complete:    true,
		ValidItemTypes: []string{"swor", "shie"},
		Runes:         []string{"r07", "r09", "r11", "r13"},
		Properties: []Property{
			{Code: "allskills", Min: 2, Max: 2, DisplayText: "+2 To All Skills"},
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var got Runeword
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if got.DisplayName != "Spirit" {
		t.Errorf("DisplayName = %q, want %q", got.DisplayName, "Spirit")
	}
	if len(got.ValidItemTypes) != 2 {
		t.Errorf("len(ValidItemTypes) = %d, want 2", len(got.ValidItemTypes))
	}
	if len(got.Runes) != 4 {
		t.Errorf("len(Runes) = %d, want 4", len(got.Runes))
	}

	// Verify JSON field names
	if !strings.Contains(string(data), `"display_name"`) {
		t.Error("expected display_name field in JSON")
	}
	if !strings.Contains(string(data), `"valid_item_types"`) {
		t.Error("expected valid_item_types field in JSON")
	}
}

func TestRuneJSON(t *testing.T) {
	item := Rune{
		ID:         1,
		Code:       "r01",
		Name:       "El Rune",
		RuneNumber: 1,
		LevelReq:   11,
		WeaponMods: []Property{
			{Code: "att", Min: 50, Max: 50, DisplayText: "+50 To Attack Rating"},
		},
		HelmMods: []Property{
			{Code: "light", Min: 1, Max: 1, DisplayText: "+1 To Light Radius"},
		},
		ShieldMods: []Property{
			{Code: "light", Min: 1, Max: 1, DisplayText: "+1 To Light Radius"},
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var got Rune
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if got.RuneNumber != 1 {
		t.Errorf("RuneNumber = %d, want 1", got.RuneNumber)
	}
	if len(got.WeaponMods) != 1 {
		t.Errorf("len(WeaponMods) = %d, want 1", len(got.WeaponMods))
	}
	if len(got.HelmMods) != 1 {
		t.Errorf("len(HelmMods) = %d, want 1", len(got.HelmMods))
	}
	if len(got.ShieldMods) != 1 {
		t.Errorf("len(ShieldMods) = %d, want 1", len(got.ShieldMods))
	}

	// Verify JSON field names
	if !strings.Contains(string(data), `"weapon_mods"`) {
		t.Error("expected weapon_mods field in JSON")
	}
	if !strings.Contains(string(data), `"helm_mods"`) {
		t.Error("expected helm_mods field in JSON")
	}
	if !strings.Contains(string(data), `"shield_mods"`) {
		t.Error("expected shield_mods field in JSON")
	}
}

func TestGemJSON(t *testing.T) {
	item := Gem{
		ID:      1,
		Code:    "gcv",
		Name:    "Chipped Amethyst",
		GemType: "amethyst",
		Quality: "chipped",
		WeaponMods: []Property{
			{Code: "att", Min: 40, Max: 40, DisplayText: "+40 To Attack Rating"},
		},
		HelmMods: []Property{
			{Code: "str", Min: 3, Max: 3, DisplayText: "+3 To Strength"},
		},
		ShieldMods: []Property{
			{Code: "ac", Min: 8, Max: 8, DisplayText: "+8 Defense"},
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var got Gem
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if got.GemType != "amethyst" {
		t.Errorf("GemType = %q, want %q", got.GemType, "amethyst")
	}
	if got.Quality != "chipped" {
		t.Errorf("Quality = %q, want %q", got.Quality, "chipped")
	}

	// Verify JSON field names
	if !strings.Contains(string(data), `"gem_type"`) {
		t.Error("expected gem_type field in JSON")
	}
	if !strings.Contains(string(data), `"quality"`) {
		t.Error("expected quality field in JSON")
	}
}
