package d2

import (
	"fmt"
	"testing"
)

func TestSpiritValidTypes(t *testing.T) {
	parser := NewHTMLItemParser()

	// Check runeword valid types
	runewords, err := parser.ParseRunewordsFile("../../../catalogs/d2/pages/runewords.html")
	if err != nil {
		t.Fatalf("ParseRunewordsFile failed: %v", err)
	}

	for _, rw := range runewords {
		if rw.Name == "Spirit" {
			fmt.Printf("Spirit ValidTypes: %v\n", rw.ValidTypes)
			fmt.Printf("Spirit Runes: %v (count=%d)\n", rw.Runes, len(rw.Runes))
			break
		}
	}

	// Check base item type_tags for known Spirit bases
	bases, err := parser.ParseBasesFile("../../../catalogs/d2/pages/base.html")
	if err != nil {
		t.Fatalf("ParseBasesFile failed: %v", err)
	}

	spiritBases := []string{"Crystal Sword", "Broad Sword", "Long Sword", "Monarch", "Targe", "Rondache"}
	for _, name := range spiritBases {
		for _, b := range bases {
			if b.Name == name {
				fmt.Printf("Base %q: TypeTags=%v MaxSockets=%d\n", b.Name, b.TypeTags, b.MaxSockets)
				break
			}
		}
	}
}
