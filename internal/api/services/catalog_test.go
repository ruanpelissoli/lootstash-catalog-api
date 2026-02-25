package services

import (
	"testing"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
)

// newTestService creates a CatalogService with pre-populated skill maps for testing
// (no DB connection needed).
func newTestService() *CatalogService {
	s := &CatalogService{
		translator: d2.DefaultTranslator,
	}
	// Mark sync.Once as done so initSkillMaps won't try to hit the DB
	s.skillOnce.Do(func() {})
	// Pre-populate skill maps directly
	s.tabToCode = map[int]string{
		0: "amazon-Bow and Crossbow",
		1: "amazon-Passive and Magic",
		2: "amazon-Javelin and Spear",
		3: "sorceress-Fire",
		4: "sorceress-Lightning",
		5: "sorceress-Cold",
		6: "necromancer-Curses",
		7: "necromancer-Poison and Bone",
		8: "necromancer-Summoning",
		9: "paladin-Combat Skills",
	}
	s.skillToCode = map[string]string{
		"Fireball":    "sorceress-Fire-Fireball",
		"Frozen Orb":  "sorceress-Cold-Frozen Orb",
		"Zeal":        "paladin-Combat Skills-Zeal",
		"Teleport":    "sorceress-Lightning-Teleport",
	}
	return s
}

func TestResolveSkillCode_Skilltab(t *testing.T) {
	s := newTestService()

	tests := []struct {
		param string
		want  string
	}{
		{"3", "sorceress-Fire"},
		{"0", "amazon-Bow and Crossbow"},
		{"5", "sorceress-Cold"},
		{"99", ""},  // unknown tab
	}

	for _, tt := range tests {
		got := s.resolveSkillCode("skilltab", tt.param)
		if got != tt.want {
			t.Errorf("resolveSkillCode(skilltab, %q) = %q, want %q", tt.param, got, tt.want)
		}
	}
}

func TestResolveSkillCode_Skill(t *testing.T) {
	s := newTestService()

	tests := []struct {
		code  string
		param string
		want  string
	}{
		{"skill", "Fireball", "sorceress-Fire-Fireball"},
		{"oskill", "Fireball", "sorceress-Fire-Fireball"},
		{"skill", "Frozen Orb", "sorceress-Cold-Frozen Orb"},
		{"skill", "Zeal", "paladin-Combat Skills-Zeal"},
		{"skill", "Unknown Skill", ""},  // unknown skill
	}

	for _, tt := range tests {
		got := s.resolveSkillCode(tt.code, tt.param)
		if got != tt.want {
			t.Errorf("resolveSkillCode(%q, %q) = %q, want %q", tt.code, tt.param, got, tt.want)
		}
	}
}

func TestResolveSkillCode_NonSkill(t *testing.T) {
	s := newTestService()

	// Non-skill codes should return empty string
	if got := s.resolveSkillCode("dmg-fire", "10"); got != "" {
		t.Errorf("resolveSkillCode(dmg-fire, 10) = %q, want empty", got)
	}
	if got := s.resolveSkillCode("str", ""); got != "" {
		t.Errorf("resolveSkillCode(str, \"\") = %q, want empty", got)
	}
}

func TestConvertPropertiesToAffixes_SkillTab(t *testing.T) {
	s := newTestService()

	props := []d2.Property{
		{Code: "skilltab", Param: "3", Min: 2, Max: 2},
	}
	affixes := s.convertPropertiesToAffixes(props)

	if len(affixes) != 1 {
		t.Fatalf("expected 1 affix, got %d", len(affixes))
	}
	if affixes[0].Code != "sorceress-Fire" {
		t.Errorf("skilltab affix code = %q, want %q", affixes[0].Code, "sorceress-Fire")
	}
}

func TestConvertPropertiesToAffixes_Skill(t *testing.T) {
	s := newTestService()

	props := []d2.Property{
		{Code: "skill", Param: "Fireball", Min: 3, Max: 3},
	}
	affixes := s.convertPropertiesToAffixes(props)

	if len(affixes) != 1 {
		t.Fatalf("expected 1 affix, got %d", len(affixes))
	}
	if affixes[0].Code != "sorceress-Fire-Fireball" {
		t.Errorf("skill affix code = %q, want %q", affixes[0].Code, "sorceress-Fire-Fireball")
	}
}

func TestConvertPropertiesToAffixes_UnknownParam(t *testing.T) {
	s := newTestService()

	// Unknown skilltab param should fall back to slugified format
	props := []d2.Property{
		{Code: "skilltab", Param: "99", Min: 1, Max: 1},
	}
	affixes := s.convertPropertiesToAffixes(props)

	if len(affixes) != 1 {
		t.Fatalf("expected 1 affix, got %d", len(affixes))
	}
	if affixes[0].Code != "skilltab-99" {
		t.Errorf("unknown skilltab affix code = %q, want %q", affixes[0].Code, "skilltab-99")
	}
}

func TestConvertPropertiesToAffixes_NonSkillParam(t *testing.T) {
	s := newTestService()

	// Non-skill parametric codes should still use slugified format
	props := []d2.Property{
		{Code: "aura", Param: "Might", Min: 12, Max: 12},
	}
	affixes := s.convertPropertiesToAffixes(props)

	if len(affixes) != 1 {
		t.Fatalf("expected 1 affix, got %d", len(affixes))
	}
	if affixes[0].Code != "aura-might" {
		t.Errorf("aura affix code = %q, want %q", affixes[0].Code, "aura-might")
	}
}
