package d2

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/storage"
)

// HTMLImportStats tracks import statistics for HTML import
type HTMLImportStats struct {
	BasesImported    int
	BasesSkipped     int
	UniquesImported  int
	UniquesSkipped   int
	SetsImported     int
	SetsSkipped      int
	SetItemsImported int
	SetItemsSkipped  int
	RunewordsImported int
	RunewordsSkipped  int
	RunesImported    int
	RunesSkipped     int
	GemsImported     int
	GemsSkipped      int
	MiscImported     int
	MiscSkipped      int
	ImagesUploaded   int
	RawProperties    int
	Errors           int
	ErrorMessages    []string
	MissingStatCodes []string
}

// htmlTypeNameToCode maps HTML type display names to D2 item type codes
var htmlTypeNameToCode = map[string]string{
	"Body Armor":    "tors",
	"Helms":         "helm",
	"Shields":       "shie",
	"Swords":        "swor",
	"Axes":          "axe",
	"Maces":         "mace",
	"Polearms":      "pole",
	"Staves":        "staf",
	"Scepters":      "scep",
	"Wands":         "wand",
	"Bows":          "bow",
	"Crossbows":     "xbow",
	"Daggers":       "knif",
	"Throwing":      "tkni",
	"Javelins":      "jave",
	"Spears":        "spea",
	"Claws":         "h2h",
	"Orbs":          "orb",
	"Amazon Weapons": "amaz",
	"Hammers":       "hamm",
	"Clubs":         "club",
	"Weapons":       "weap",
	"Missile Weapons": "miss",
	"Melee Weapons": "mele",
	"Gloves":        "glov",
	"Boots":         "boot",
	"Belts":         "belt",
	"Circlets":      "circ",
	"Druid Pelts":   "pelt",
	"Barbarian Helms": "phlm",
	"Necromancer Shields": "head",
	"Shrunken Heads": "head",
	"Paladin Shields": "ashd",
	"Targes":         "ashd",
	"Grimoires":      "grim",
	"Katars":          "h2h",
	"Wand":            "wand",
	"Armor":           "tors",
	"All Weapons":    "weap",
	"All Armor":      "armo",
	// Socketed type variants from runeword HTML (e.g., "4 socket Weapons")
	"2 socket Weapons":    "weap",
	"3 socket Weapons":    "weap",
	"4 socket Weapons":    "weap",
	"5 socket Weapons":    "weap",
	"6 socket Weapons":    "weap",
	"2 socket Shields":    "shie",
	"3 socket Shields":    "shie",
	"4 socket Shields":    "shie",
	"2 socket Swords":     "swor",
	"3 socket Swords":     "swor",
	"4 socket Swords":     "swor",
	"5 socket Swords":     "swor",
	"6 socket Swords":     "swor",
	"2 socket Body Armor": "tors",
	"3 socket Body Armor": "tors",
	"4 socket Body Armor": "tors",
	"2 socket Armor":      "tors",
	"3 socket Armor":      "tors",
	"4 socket Armor":      "tors",
	"2 socket Helms":      "helm",
	"3 socket Helms":      "helm",
	"4 socket Helms":      "helm",
}

// HTMLImporter orchestrates importing items from HTML pages
type HTMLImporter struct {
	repo              *Repository
	parser            *HTMLItemParser
	reverseTranslator *ReverseTranslator
	translator        *PropertyTranslator
	storage           storage.Storage
	dryRun            bool
	force             bool
	iconsPath         string

	// Caches loaded from DB at startup
	baseNameToCode    map[string]string
	runeNameToCode    map[string]string
	existingBases     map[string]bool
	existingUniques   map[string]bool
	existingSetItems  map[string]bool
	existingSets      map[string]bool
	existingRunewords map[string]bool
	existingRunes     map[string]bool
	existingGems      map[string]bool
	existingImageURLs map[string]string // normalized name -> image_url (for items that already have images)
	imageCache        map[string]string // imagePath -> uploaded URL

	// Name -> index_id maps for force mode (reuse existing IDs)
	uniqueNameToID   map[string]int
	setItemNameToID  map[string]int
	setBonusNameToID map[string]int

	// Track all property codes seen during import for validation
	seenCodes map[string]bool
}

// NewHTMLImporter creates a new HTML importer
func NewHTMLImporter(repo *Repository, stor storage.Storage, dryRun bool, force bool) *HTMLImporter {
	return &HTMLImporter{
		repo:              repo,
		parser:            NewHTMLItemParser(),
		reverseTranslator: NewReverseTranslator(),
		translator:        NewPropertyTranslator(),
		storage:           stor,
		dryRun:            dryRun,
		force:             force,
		imageCache:        make(map[string]string),
		seenCodes:         make(map[string]bool),
	}
}

// ImportFromHTML imports items from HTML files
func (h *HTMLImporter) ImportFromHTML(ctx context.Context, catalogPath, itemType string) (*HTMLImportStats, error) {
	stats := &HTMLImportStats{}

	h.iconsPath = filepath.Join(catalogPath, "icons")
	pagesPath := filepath.Join(catalogPath, "pages")

	// Load caches from DB
	fmt.Println("  Loading lookup caches from database...")
	if err := h.loadCaches(ctx); err != nil {
		return nil, fmt.Errorf("failed to load caches: %w", err)
	}
	fmt.Printf("    Base names: %d, Rune names: %d\n", len(h.baseNameToCode), len(h.runeNameToCode))
	fmt.Printf("    Existing bases: %d, uniques: %d, sets: %d, set items: %d, runewords: %d, runes: %d, gems: %d\n",
		len(h.existingBases), len(h.existingUniques), len(h.existingSets), len(h.existingSetItems), len(h.existingRunewords),
		len(h.existingRunes), len(h.existingGems))

	// Import by type
	switch itemType {
	case "bases":
		if err := h.importBases(ctx, pagesPath, stats); err != nil {
			return stats, err
		}
	case "uniques":
		if err := h.importUniques(ctx, pagesPath, stats); err != nil {
			return stats, err
		}
	case "sets":
		if err := h.importSets(ctx, pagesPath, stats); err != nil {
			return stats, err
		}
	case "runewords":
		if err := h.importRunewords(ctx, pagesPath, stats); err != nil {
			return stats, err
		}
	case "runes":
		if err := h.importMiscRunes(ctx, pagesPath, stats); err != nil {
			return stats, err
		}
	case "gems":
		if err := h.importMiscGems(ctx, pagesPath, stats); err != nil {
			return stats, err
		}
	case "misc":
		if err := h.importMisc(ctx, pagesPath, stats); err != nil {
			return stats, err
		}
	case "all":
		// Import bases first so unique/set items can resolve base codes
		if err := h.importBases(ctx, pagesPath, stats); err != nil {
			return stats, err
		}
		// Reload baseNameToCode cache after importing new bases
		h.baseNameToCode, _ = h.repo.GetAllItemBaseNameToCode(ctx)
		if err := h.importUniques(ctx, pagesPath, stats); err != nil {
			return stats, err
		}
		if err := h.importSets(ctx, pagesPath, stats); err != nil {
			return stats, err
		}
		if err := h.importRunewords(ctx, pagesPath, stats); err != nil {
			return stats, err
		}
		if err := h.importMisc(ctx, pagesPath, stats); err != nil {
			return stats, err
		}
	default:
		return nil, fmt.Errorf("unknown item type: %s (use: bases, uniques, sets, runewords, runes, gems, misc, all)", itemType)
	}

	// Validate seen property codes against FilterableStats registry
	allCodes := make([]string, 0, len(h.seenCodes))
	for code := range h.seenCodes {
		allCodes = append(allCodes, code)
	}
	stats.MissingStatCodes = ValidateStatCodes(allCodes)

	return stats, nil
}

// collectPropertyCodes records all property codes from a slice into seenCodes
func (h *HTMLImporter) collectPropertyCodes(props []Property) {
	for _, p := range props {
		if p.Code != "" {
			h.seenCodes[p.Code] = true
		}
	}
}

// loadCaches loads lookup data from the database
func (h *HTMLImporter) loadCaches(ctx context.Context) error {
	var err error

	h.baseNameToCode, err = h.repo.GetAllItemBaseNameToCode(ctx)
	if err != nil {
		return fmt.Errorf("base name map: %w", err)
	}

	h.runeNameToCode, err = h.repo.GetRuneNameToCodeMap(ctx)
	if err != nil {
		return fmt.Errorf("rune name map: %w", err)
	}

	h.existingBases, err = h.repo.GetAllExistingNames(ctx, "item_bases", "name")
	if err != nil {
		return fmt.Errorf("existing bases: %w", err)
	}

	h.existingUniques, err = h.repo.GetAllExistingNames(ctx, "unique_items", "name")
	if err != nil {
		return fmt.Errorf("existing uniques: %w", err)
	}

	h.existingSetItems, err = h.repo.GetAllExistingNames(ctx, "set_items", "name")
	if err != nil {
		return fmt.Errorf("existing set items: %w", err)
	}

	h.existingSets, err = h.repo.GetAllExistingNames(ctx, "set_bonuses", "name")
	if err != nil {
		return fmt.Errorf("existing sets: %w", err)
	}

	h.existingRunewords, err = h.repo.GetAllExistingNames(ctx, "runewords", "display_name")
	if err != nil {
		return fmt.Errorf("existing runewords: %w", err)
	}

	h.existingRunes, err = h.repo.GetAllExistingNames(ctx, "runes", "name")
	if err != nil {
		return fmt.Errorf("existing runes: %w", err)
	}

	h.existingGems, err = h.repo.GetAllExistingNames(ctx, "gems", "name")
	if err != nil {
		return fmt.Errorf("existing gems: %w", err)
	}

	// Load names that already have images (to skip unnecessary uploads)
	h.existingImageURLs = make(map[string]string)
	for _, tbl := range []struct{ table, col string }{
		{"item_bases", "name"},
		{"unique_items", "name"},
		{"set_items", "name"},
		{"runewords", "display_name"},
		{"runes", "name"},
		{"gems", "name"},
	} {
		names, err := h.repo.GetNamesWithImages(ctx, tbl.table, tbl.col)
		if err != nil {
			return fmt.Errorf("images for %s: %w", tbl.table, err)
		}
		for name := range names {
			h.existingImageURLs[name] = tbl.table
		}
	}

	// Load name -> index_id maps for force mode
	if h.force {
		h.uniqueNameToID, err = h.repo.GetNameToIndexID(ctx, "unique_items", "name")
		if err != nil {
			return fmt.Errorf("unique name-to-id: %w", err)
		}
		h.setItemNameToID, err = h.repo.GetNameToIndexID(ctx, "set_items", "name")
		if err != nil {
			return fmt.Errorf("set item name-to-id: %w", err)
		}
		h.setBonusNameToID, err = h.repo.GetNameToIndexID(ctx, "set_bonuses", "name")
		if err != nil {
			return fmt.Errorf("set bonus name-to-id: %w", err)
		}
	}

	return nil
}

// importBases imports base items from HTML
func (h *HTMLImporter) importBases(ctx context.Context, pagesPath string, stats *HTMLImportStats) error {
	basePath := filepath.Join(pagesPath, "base.html")
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		fmt.Println("\n  No base.html found, skipping base import")
		return nil
	}

	fmt.Println("\n  Parsing base.html...")
	items, err := h.parser.ParseBasesFile(basePath)
	if err != nil {
		return fmt.Errorf("failed to parse base.html: %w", err)
	}
	fmt.Printf("    Found %d base items in HTML\n", len(items))

	// Track codes we generate to avoid collisions within this import
	usedCodes := make(map[string]bool)
	for code := range h.baseNameToCode {
		usedCodes[code] = true
	}

	// Ensure new item types exist in item_types table
	ensuredTypes := make(map[string]bool)

	for _, item := range items {
		normalized := NormalizeItemName(item.Name)

		// Check if already exists (skip unless --force)
		if h.existingBases[normalized] && !h.force {
			stats.BasesSkipped++
			continue
		}

		// Generate a short code (max 10 chars for DB varchar(10))
		code := generateBaseCode(item.Name)
		// Ensure code uniqueness
		if usedCodes[code] {
			for i := 2; ; i++ {
				candidate := fmt.Sprintf("%s%d", code, i)
				if len(candidate) > 10 {
					candidate = fmt.Sprintf("%s%d", code[:10-len(fmt.Sprintf("%d", i))], i)
				}
				if !usedCodes[candidate] {
					code = candidate
					break
				}
			}
		}
		usedCodes[code] = true

		// Determine category from stats
		category := "misc"
		if item.DefenseMax > 0 {
			category = "armor"
		} else if item.OneHMaxDam > 0 || item.TwoHMaxDam > 0 {
			category = "weapon"
		}

		// Map type names to codes
		itemType := ""
		if item.TypeName != "" {
			if tc, ok := htmlTypeNameToCode[item.TypeName]; ok {
				itemType = tc
			}
		}
		itemType2 := ""
		if item.TypeName2 != "" {
			if tc, ok := htmlTypeNameToCode[item.TypeName2]; ok {
				itemType2 = tc
			}
		}

		// Ensure item type exists in item_types table
		for _, tc := range []struct{ code, name string }{{itemType, item.TypeName}, {itemType2, item.TypeName2}} {
			if tc.code != "" && !ensuredTypes[tc.code] {
				exists, _ := h.repo.ItemTypeExists(ctx, tc.code)
				if !exists && !h.dryRun {
					fmt.Printf("    Creating new item type: %s (%s)\n", tc.name, tc.code)
					h.repo.UpsertItemType(ctx, &ItemType{
						Code:       tc.code,
						Name:       tc.name,
						CanBeMagic: true,
						CanBeRare:  true,
					})
				}
				ensuredTypes[tc.code] = true
			}
		}

		// Upload image
		imageURL := h.uploadItemImage(ctx, item.ImagePath, "d2/base", item.Name, stats)

		base := &ItemBase{
			Code:         code,
			Name:         item.Name,
			ItemType:     itemType,
			ItemType2:    itemType2,
			Category:     category,
			Level:        item.QualityLevel,
			LevelReq:     item.ReqLevel,
			StrReq:       item.ReqStr,
			DexReq:       item.ReqDex,
			Durability:   item.Durability,
			MinAC:        item.DefenseMin,
			MaxAC:        item.DefenseMax,
			MinDam:       item.OneHMinDam,
			MaxDam:       item.OneHMaxDam,
			TwoHandMinDam: item.TwoHMinDam,
			TwoHandMaxDam: item.TwoHMaxDam,
			RangeAdder:   item.RangeAdder,
			Speed:        item.Speed,
			MaxSockets:   item.MaxSockets,
			InvWidth:     item.InvWidth,
			InvHeight:    item.InvHeight,
			Spawnable:    true,
			Rarity:       1,
			ImageURL:     imageURL,
		}

		if !h.dryRun {
			if err := h.repo.UpsertItemBase(ctx, base); err != nil {
				stats.Errors++
				stats.ErrorMessages = append(stats.ErrorMessages, fmt.Sprintf("base %s: %v", item.Name, err))
				continue
			}
		} else {
			fmt.Printf("    [DRY-RUN] Would import base: %s (code: %s, type: %s, cat: %s)\n",
				item.Name, code, itemType, category)
		}

		stats.BasesImported++
		h.existingBases[normalized] = true
	}

	fmt.Printf("    Bases: %d imported, %d skipped\n", stats.BasesImported, stats.BasesSkipped)
	return nil
}

// generateBaseCode creates a short code (max 10 chars) from a base item name.
// Mimics the style of existing D2 item codes (3-6 lowercase chars).
func generateBaseCode(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "-", "")
	words := strings.Fields(name)
	if len(words) == 1 {
		w := words[0]
		if len(w) > 4 {
			return w[:4]
		}
		return w
	}
	// Multi-word: take 3 chars of first word + 2 chars of remaining words
	code := ""
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		take := 2
		if i == 0 {
			take = 3
			if len(w) < 3 {
				take = len(w)
			}
		} else if len(w) < 2 {
			take = len(w)
		}
		code += w[:take]
		if len(code) >= 8 {
			break
		}
	}
	if len(code) > 8 {
		code = code[:8]
	}
	return code
}

// importUniques imports unique items from HTML
func (h *HTMLImporter) importUniques(ctx context.Context, pagesPath string, stats *HTMLImportStats) error {
	fmt.Println("\n  Parsing uniques.html...")
	items, err := h.parser.ParseUniquesFile(filepath.Join(pagesPath, "uniques.html"))
	if err != nil {
		return fmt.Errorf("failed to parse uniques.html: %w", err)
	}
	fmt.Printf("    Found %d unique items in HTML\n", len(items))

	// Get next index ID
	maxID, err := h.repo.GetMaxIndexID(ctx, "unique_items")
	if err != nil {
		return fmt.Errorf("failed to get max unique index ID: %w", err)
	}
	nextID := maxID + 1

	for _, item := range items {
		normalized := NormalizeItemName(item.Name)

		// Check if already exists (skip unless --force)
		if h.existingUniques[normalized] && !h.force {
			stats.UniquesSkipped++
			continue
		}

		// In force mode, reuse existing index_id to avoid duplicates
		itemID := nextID
		if h.force {
			if existingID, ok := h.uniqueNameToID[normalized]; ok {
				itemID = existingID
			} else {
				nextID++
			}
		} else {
			nextID++
		}

		// Resolve base code
		baseCode := ""
		if item.BaseName != "" {
			if code, ok := h.baseNameToCode[item.BaseName]; ok {
				baseCode = code
			} else {
				fmt.Printf("    Warning: base '%s' not found for unique '%s'\n", item.BaseName, item.Name)
			}
		}

		// Reverse-translate properties
		properties := h.reverseTranslator.ReverseTranslateLines(item.Properties)
		for i := range properties {
			if properties[i].Code == "raw" {
				stats.RawProperties++
			}
		}
		properties = combineAllAttributes(properties, h.translator)

		// Enrich properties
		for i := range properties {
			if properties[i].Code != "raw" {
				h.translator.EnrichProperty(&properties[i])
			}
		}

		h.collectPropertyCodes(properties)

		// Upload image
		imageURL := h.uploadItemImage(ctx, item.ImagePath, "d2/unique", item.Name, stats)

		unique := &UniqueItem{
			IndexID:    itemID,
			Name:       item.Name,
			BaseCode:   baseCode,
			BaseName:   item.BaseName,
			Level:      item.QualityLevel,
			LevelReq:   item.ReqLevel,
			Rarity:     1,
			Enabled:    true,
			Properties: properties,
			ImageURL:   imageURL,
		}

		if !h.dryRun {
			if err := h.repo.UpsertUniqueItem(ctx, unique); err != nil {
				stats.Errors++
				stats.ErrorMessages = append(stats.ErrorMessages, fmt.Sprintf("unique %s: %v", item.Name, err))
				continue
			}
		} else {
			fmt.Printf("    [DRY-RUN] Would import unique: %s (base: %s, props: %d)\n", item.Name, item.BaseName, len(properties))
		}

		stats.UniquesImported++
	}

	fmt.Printf("    Uniques: %d imported, %d skipped\n", stats.UniquesImported, stats.UniquesSkipped)
	return nil
}

// importSets imports set bonuses and set items from HTML
func (h *HTMLImporter) importSets(ctx context.Context, pagesPath string, stats *HTMLImportStats) error {
	fmt.Println("\n  Parsing sets.html...")
	setItems, _, err := h.parser.ParseSetsFile(filepath.Join(pagesPath, "sets.html"))
	if err != nil {
		return fmt.Errorf("failed to parse sets.html: %w", err)
	}
	fmt.Printf("    Found %d set items in HTML\n", len(setItems))

	// First pass: collect unique set names and create SetBonus entries
	setNames := make(map[string]bool)
	for _, item := range setItems {
		if item.SetName == "" {
			continue
		}
		normalizedSet := NormalizeItemName(item.SetName)
		if (h.existingSets[normalizedSet] && !h.force) || setNames[normalizedSet] {
			continue
		}
		setNames[normalizedSet] = true

		// Determine index ID: reuse existing in force mode, otherwise get next
		setBonusID := 0
		if h.force {
			if existingID, ok := h.setBonusNameToID[normalizedSet]; ok {
				setBonusID = existingID
			}
		}
		if setBonusID == 0 {
			maxSetID, err := h.repo.GetMaxIndexID(ctx, "set_bonuses")
			if err != nil {
				return fmt.Errorf("failed to get max set bonus index ID: %w", err)
			}
			setBonusID = maxSetID + 1
		}

		setBonus := &SetBonus{
			IndexID: setBonusID,
			Name:    item.SetName,
		}

		if !h.dryRun {
			if err := h.repo.UpsertSetBonus(ctx, setBonus); err != nil {
				stats.Errors++
				stats.ErrorMessages = append(stats.ErrorMessages, fmt.Sprintf("set %s: %v", item.SetName, err))
				continue
			}
		} else {
			fmt.Printf("    [DRY-RUN] Would import set: %s\n", item.SetName)
		}

		stats.SetsImported++
		h.existingSets[normalizedSet] = true
	}

	// Second pass: import set items
	maxItemID, err := h.repo.GetMaxIndexID(ctx, "set_items")
	if err != nil {
		return fmt.Errorf("failed to get max set item index ID: %w", err)
	}
	nextItemID := maxItemID + 1

	for _, item := range setItems {
		normalized := NormalizeItemName(item.Name)

		// Check if already exists (skip unless --force)
		if h.existingSetItems[normalized] && !h.force {
			stats.SetItemsSkipped++
			continue
		}

		// In force mode, reuse existing index_id to avoid duplicates
		itemID := nextItemID
		if h.force {
			if existingID, ok := h.setItemNameToID[normalized]; ok {
				itemID = existingID
			} else {
				nextItemID++
			}
		} else {
			nextItemID++
		}

		// Resolve base code
		baseCode := ""
		if item.BaseName != "" {
			if code, ok := h.baseNameToCode[item.BaseName]; ok {
				baseCode = code
			} else {
				fmt.Printf("    Warning: base '%s' not found for set item '%s'\n", item.BaseName, item.Name)
			}
		}

		// Reverse-translate properties
		properties := h.reverseTranslator.ReverseTranslateLines(item.Properties)
		for i := range properties {
			if properties[i].Code == "raw" {
				stats.RawProperties++
			}
		}
		properties = combineAllAttributes(properties, h.translator)

		// Enrich properties
		for i := range properties {
			if properties[i].Code != "raw" {
				h.translator.EnrichProperty(&properties[i])
			}
		}

		// Reverse-translate set bonuses into bonus properties
		var bonusProperties []Property
		for _, bonus := range item.SetBonuses {
			// Split "or" bonuses (e.g., "+3 to Mana after each Kill or 30% Deadly Strike")
			bonusLines := splitOrBonuses(bonus.Text)
			for _, line := range bonusLines {
				prop := h.reverseTranslator.ReverseTranslate(line)
				if prop.Code == "raw" {
					stats.RawProperties++
				} else {
					h.translator.EnrichProperty(&prop)
				}
				bonusProperties = append(bonusProperties, prop)
			}
		}
		bonusProperties = combineAllAttributes(bonusProperties, h.translator)

		h.collectPropertyCodes(properties)
		h.collectPropertyCodes(bonusProperties)

		// Upload image
		imageURL := h.uploadItemImage(ctx, item.ImagePath, "d2/set", item.Name, stats)

		setItem := &SetItem{
			IndexID:         itemID,
			Name:            item.Name,
			SetName:         item.SetName,
			BaseCode:        baseCode,
			BaseName:        item.BaseName,
			Level:           item.QualityLevel,
			LevelReq:        item.ReqLevel,
			Rarity:          1,
			Properties:      properties,
			BonusProperties: bonusProperties,
			ImageURL:        imageURL,
		}

		if !h.dryRun {
			if err := h.repo.UpsertSetItem(ctx, setItem); err != nil {
				stats.Errors++
				stats.ErrorMessages = append(stats.ErrorMessages, fmt.Sprintf("set item %s: %v", item.Name, err))
				continue
			}
		} else {
			fmt.Printf("    [DRY-RUN] Would import set item: %s (set: %s, base: %s, props: %d, bonus: %d)\n",
				item.Name, item.SetName, item.BaseName, len(properties), len(bonusProperties))
		}

		stats.SetItemsImported++
	}

	fmt.Printf("    Sets: %d imported, Set items: %d imported, %d skipped\n",
		stats.SetsImported, stats.SetItemsImported, stats.SetItemsSkipped)
	return nil
}

// importRunewords imports runewords from HTML
func (h *HTMLImporter) importRunewords(ctx context.Context, pagesPath string, stats *HTMLImportStats) error {
	fmt.Println("\n  Parsing runewords.html...")
	runewords, err := h.parser.ParseRunewordsFile(filepath.Join(pagesPath, "runewords.html"))
	if err != nil {
		return fmt.Errorf("failed to parse runewords.html: %w", err)
	}
	fmt.Printf("    Found %d runewords in HTML\n", len(runewords))

	for _, rw := range runewords {
		normalized := NormalizeItemName(rw.Name)

		// Check if already exists (skip unless --force)
		if h.existingRunewords[normalized] && !h.force {
			stats.RunewordsSkipped++
			continue
		}

		// Resolve rune names to codes
		var runeCodes []string
		allResolved := true
		for _, runeName := range rw.Runes {
			if code, ok := h.runeNameToCode[runeName]; ok {
				runeCodes = append(runeCodes, code)
			} else {
				fmt.Printf("    Warning: rune '%s' not found for runeword '%s'\n", runeName, rw.Name)
				allResolved = false
			}
		}
		if !allResolved {
			stats.Errors++
			stats.ErrorMessages = append(stats.ErrorMessages, fmt.Sprintf("runeword %s: could not resolve all runes", rw.Name))
			continue
		}

		// Resolve valid type names to codes
		var validTypeCodes []string
		for _, typeName := range rw.ValidTypes {
			if code, ok := htmlTypeNameToCode[typeName]; ok {
				validTypeCodes = append(validTypeCodes, code)
			} else {
				fmt.Printf("    Warning: type '%s' not found for runeword '%s'\n", typeName, rw.Name)
			}
		}

		// Reverse-translate properties
		properties := h.reverseTranslator.ReverseTranslateLines(rw.Properties)
		for i := range properties {
			if properties[i].Code == "raw" {
				stats.RawProperties++
			}
		}
		properties = combineAllAttributes(properties, h.translator)

		// Enrich properties
		for i := range properties {
			if properties[i].Code != "raw" {
				h.translator.EnrichProperty(&properties[i])
			}
		}

		h.collectPropertyCodes(properties)

		// Generate an internal name for the runeword (like "Runeword123")
		internalName := fmt.Sprintf("HTMLRuneword_%s", strings.ReplaceAll(rw.Name, " ", ""))

		runeword := &Runeword{
			Name:           internalName,
			DisplayName:    rw.Name,
			Complete:       true,
			ValidItemTypes: validTypeCodes,
			Runes:          runeCodes,
			Properties:     properties,
		}

		if !h.dryRun {
			if err := h.repo.UpsertRuneword(ctx, runeword); err != nil {
				stats.Errors++
				stats.ErrorMessages = append(stats.ErrorMessages, fmt.Sprintf("runeword %s: %v", rw.Name, err))
				continue
			}
		} else {
			fmt.Printf("    [DRY-RUN] Would import runeword: %s (runes: %v, types: %v, props: %d)\n",
				rw.Name, rw.Runes, rw.ValidTypes, len(properties))
		}

		stats.RunewordsImported++
	}

	fmt.Printf("    Runewords: %d imported, %d skipped\n", stats.RunewordsImported, stats.RunewordsSkipped)
	return nil
}

// importMisc imports runes, gems, and misc items from misc.html
func (h *HTMLImporter) importMisc(ctx context.Context, pagesPath string, stats *HTMLImportStats) error {
	miscPath := filepath.Join(pagesPath, "misc.html")
	if _, err := os.Stat(miscPath); os.IsNotExist(err) {
		fmt.Println("\n  No misc.html found, skipping misc import")
		return nil
	}

	fmt.Println("\n  Parsing misc.html...")
	runes, gems, miscItems, err := h.parser.ParseMiscFile(miscPath)
	if err != nil {
		return fmt.Errorf("failed to parse misc.html: %w", err)
	}
	fmt.Printf("    Found %d runes, %d gems, %d misc items in HTML\n", len(runes), len(gems), len(miscItems))

	if err := h.importRuneItems(ctx, runes, stats); err != nil {
		return err
	}
	if err := h.importGemItems(ctx, gems, stats); err != nil {
		return err
	}
	if err := h.importMiscItems(ctx, miscItems, stats); err != nil {
		return err
	}

	return nil
}

// importMiscRunes imports only runes from misc.html
func (h *HTMLImporter) importMiscRunes(ctx context.Context, pagesPath string, stats *HTMLImportStats) error {
	miscPath := filepath.Join(pagesPath, "misc.html")
	if _, err := os.Stat(miscPath); os.IsNotExist(err) {
		fmt.Println("\n  No misc.html found, skipping rune import")
		return nil
	}

	fmt.Println("\n  Parsing misc.html for runes...")
	runes, _, _, err := h.parser.ParseMiscFile(miscPath)
	if err != nil {
		return fmt.Errorf("failed to parse misc.html: %w", err)
	}
	fmt.Printf("    Found %d runes in HTML\n", len(runes))

	return h.importRuneItems(ctx, runes, stats)
}

// importMiscGems imports only gems from misc.html
func (h *HTMLImporter) importMiscGems(ctx context.Context, pagesPath string, stats *HTMLImportStats) error {
	miscPath := filepath.Join(pagesPath, "misc.html")
	if _, err := os.Stat(miscPath); os.IsNotExist(err) {
		fmt.Println("\n  No misc.html found, skipping gem import")
		return nil
	}

	fmt.Println("\n  Parsing misc.html for gems...")
	_, gems, _, err := h.parser.ParseMiscFile(miscPath)
	if err != nil {
		return fmt.Errorf("failed to parse misc.html: %w", err)
	}
	fmt.Printf("    Found %d gems in HTML\n", len(gems))

	return h.importGemItems(ctx, gems, stats)
}

// importRuneItems imports rune items parsed from misc.html
func (h *HTMLImporter) importRuneItems(ctx context.Context, runes []HTMLParsedRune, stats *HTMLImportStats) error {
	for _, rn := range runes {
		normalized := NormalizeItemName(rn.Name)

		if h.existingRunes[normalized] && !h.force {
			stats.RunesSkipped++
			continue
		}

		// Look up existing code from runeNameToCode map, or generate from index
		code := ""
		if c, ok := h.runeNameToCode[rn.Name]; ok {
			code = c
		} else {
			// Generate code like "r27" from rune index
			code = fmt.Sprintf("r%02d", rn.RuneIndex)
		}

		// Reverse-translate weapon/helm/shield mod text -> []Property
		weaponMods := h.reverseTranslator.ReverseTranslateLines(rn.WeaponMods)
		for i := range weaponMods {
			if weaponMods[i].Code == "raw" {
				stats.RawProperties++
			} else {
				h.translator.EnrichProperty(&weaponMods[i])
			}
		}

		helmMods := h.reverseTranslator.ReverseTranslateLines(rn.HelmMods)
		for i := range helmMods {
			if helmMods[i].Code == "raw" {
				stats.RawProperties++
			} else {
				h.translator.EnrichProperty(&helmMods[i])
			}
		}

		shieldMods := h.reverseTranslator.ReverseTranslateLines(rn.ShieldMods)
		for i := range shieldMods {
			if shieldMods[i].Code == "raw" {
				stats.RawProperties++
			} else {
				h.translator.EnrichProperty(&shieldMods[i])
			}
		}

		h.collectPropertyCodes(weaponMods)
		h.collectPropertyCodes(helmMods)
		h.collectPropertyCodes(shieldMods)

		// Upload image
		imageURL := h.uploadItemImage(ctx, rn.ImagePath, "d2/rune", rn.Name, stats)

		runeItem := &Rune{
			Code:       code,
			Name:       rn.Name,
			RuneNumber: rn.RuneIndex,
			Level:      rn.Level,
			LevelReq:   rn.Level,
			WeaponMods: weaponMods,
			HelmMods:   helmMods,
			ShieldMods: shieldMods,
			ImageURL:   imageURL,
		}

		if !h.dryRun {
			if err := h.repo.UpsertRune(ctx, runeItem); err != nil {
				stats.Errors++
				stats.ErrorMessages = append(stats.ErrorMessages, fmt.Sprintf("rune %s: %v", rn.Name, err))
				continue
			}
		} else {
			fmt.Printf("    [DRY-RUN] Would import rune: %s (code: %s, index: %d, level: %d, weapon: %d, helm: %d, shield: %d)\n",
				rn.Name, code, rn.RuneIndex, rn.Level, len(weaponMods), len(helmMods), len(shieldMods))
		}

		stats.RunesImported++
		h.existingRunes[normalized] = true
	}

	fmt.Printf("    Runes: %d imported, %d skipped\n", stats.RunesImported, stats.RunesSkipped)
	return nil
}

// importGemItems imports gem items parsed from misc.html
func (h *HTMLImporter) importGemItems(ctx context.Context, gems []HTMLParsedGem, stats *HTMLImportStats) error {
	for _, gem := range gems {
		normalized := NormalizeItemName(gem.Name)

		if h.existingGems[normalized] && !h.force {
			stats.GemsSkipped++
			continue
		}

		// Use parseGemNameParts to get gemType/quality
		gemType, quality := parseGemNameParts(gem.Name)

		// Generate code from name
		code := generateBaseCode(gem.Name)

		// Reverse-translate weapon/helm/shield mod text -> []Property
		weaponMods := h.reverseTranslator.ReverseTranslateLines(gem.WeaponMods)
		for i := range weaponMods {
			if weaponMods[i].Code == "raw" {
				stats.RawProperties++
			} else {
				h.translator.EnrichProperty(&weaponMods[i])
			}
		}

		helmMods := h.reverseTranslator.ReverseTranslateLines(gem.HelmMods)
		for i := range helmMods {
			if helmMods[i].Code == "raw" {
				stats.RawProperties++
			} else {
				h.translator.EnrichProperty(&helmMods[i])
			}
		}

		shieldMods := h.reverseTranslator.ReverseTranslateLines(gem.ShieldMods)
		for i := range shieldMods {
			if shieldMods[i].Code == "raw" {
				stats.RawProperties++
			} else {
				h.translator.EnrichProperty(&shieldMods[i])
			}
		}

		h.collectPropertyCodes(weaponMods)
		h.collectPropertyCodes(helmMods)
		h.collectPropertyCodes(shieldMods)

		// Upload image
		imageURL := h.uploadItemImage(ctx, gem.ImagePath, "d2/gem", gem.Name, stats)

		gemItem := &Gem{
			Code:       code,
			Name:       gem.Name,
			GemType:    gemType,
			Quality:    quality,
			WeaponMods: weaponMods,
			HelmMods:   helmMods,
			ShieldMods: shieldMods,
			ImageURL:   imageURL,
		}

		if !h.dryRun {
			if err := h.repo.UpsertGem(ctx, gemItem); err != nil {
				stats.Errors++
				stats.ErrorMessages = append(stats.ErrorMessages, fmt.Sprintf("gem %s: %v", gem.Name, err))
				continue
			}
		} else {
			fmt.Printf("    [DRY-RUN] Would import gem: %s (code: %s, type: %s, quality: %s, weapon: %d, helm: %d, shield: %d)\n",
				gem.Name, code, gemType, quality, len(weaponMods), len(helmMods), len(shieldMods))
		}

		stats.GemsImported++
		h.existingGems[normalized] = true
	}

	fmt.Printf("    Gems: %d imported, %d skipped\n", stats.GemsImported, stats.GemsSkipped)
	return nil
}

// importMiscItems imports miscellaneous items (worldstone shards, essences, keys, etc.) as item_bases
func (h *HTMLImporter) importMiscItems(ctx context.Context, miscItems []HTMLParsedMiscItem, stats *HTMLImportStats) error {
	// Track codes we generate to avoid collisions within this import
	usedCodes := make(map[string]bool)
	for code := range h.baseNameToCode {
		usedCodes[code] = true
	}

	for _, item := range miscItems {
		normalized := NormalizeItemName(item.Name)

		if h.existingBases[normalized] && !h.force {
			stats.MiscSkipped++
			continue
		}

		// Generate code
		code := generateBaseCode(item.Name)
		if usedCodes[code] {
			for i := 2; ; i++ {
				candidate := fmt.Sprintf("%s%d", code, i)
				if len(candidate) > 10 {
					candidate = fmt.Sprintf("%s%d", code[:10-len(fmt.Sprintf("%d", i))], i)
				}
				if !usedCodes[candidate] {
					code = candidate
					break
				}
			}
		}
		usedCodes[code] = true

		// Upload image
		imageURL := h.uploadItemImage(ctx, item.ImagePath, "d2/misc", item.Name, stats)

		base := &ItemBase{
			Code:        code,
			Name:        item.Name,
			Category:    "misc",
			Spawnable:   true,
			Rarity:      1,
			Description: item.Description,
			ImageURL:    imageURL,
		}

		if !h.dryRun {
			if err := h.repo.UpsertItemBase(ctx, base); err != nil {
				stats.Errors++
				stats.ErrorMessages = append(stats.ErrorMessages, fmt.Sprintf("misc %s: %v", item.Name, err))
				continue
			}
		} else {
			fmt.Printf("    [DRY-RUN] Would import misc: %s (code: %s, desc: %s)\n",
				item.Name, code, item.Description)
		}

		stats.MiscImported++
		h.existingBases[normalized] = true
	}

	fmt.Printf("    Misc items: %d imported, %d skipped\n", stats.MiscImported, stats.MiscSkipped)
	return nil
}

// uploadItemImage handles image upload for an item.
// Skips upload if the item already has an image in the database.
func (h *HTMLImporter) uploadItemImage(ctx context.Context, imagePath, category, itemName string, stats *HTMLImportStats) string {
	if imagePath == "" || h.storage == nil {
		return ""
	}

	// Skip if item already has an image in the database
	normalized := NormalizeItemName(itemName)
	if _, hasImage := h.existingImageURLs[normalized]; hasImage {
		return ""
	}

	// Check image cache
	if url, ok := h.imageCache[imagePath]; ok {
		return url
	}

	// Extract filename from image path
	imageFilename := filepath.Base(imagePath)

	// Look for the file in icons folder
	localPath := filepath.Join(h.iconsPath, imageFilename)
	data, err := os.ReadFile(localPath)
	if err != nil {
		// Try case-insensitive match
		data = h.findImageFileCaseInsensitive(imageFilename)
		if data == nil {
			return ""
		}
	}

	storagePath := storage.StoragePath(category, itemName)

	if h.dryRun {
		url := fmt.Sprintf("[dry-run] %s", storagePath)
		h.imageCache[imagePath] = url
		stats.ImagesUploaded++
		return url
	}

	publicURL, err := h.storage.UploadImage(ctx, storagePath, data, "image/png")
	if err != nil {
		fmt.Printf("    Error uploading image for %s: %v\n", itemName, err)
		stats.Errors++
		return ""
	}

	h.imageCache[imagePath] = publicURL
	stats.ImagesUploaded++
	return publicURL
}

// findImageFileCaseInsensitive looks for an image file case-insensitively
func (h *HTMLImporter) findImageFileCaseInsensitive(filename string) []byte {
	lowerFilename := strings.ToLower(filename)
	entries, err := os.ReadDir(h.iconsPath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.ToLower(entry.Name()) == lowerFilename {
			data, err := os.ReadFile(filepath.Join(h.iconsPath, entry.Name()))
			if err == nil {
				return data
			}
		}
	}
	return nil
}

// splitOrBonuses splits bonus text that contains "or" into individual lines
// e.g., "+3 to Mana after each Kill or 30% Deadly Strike" -> ["+3 to Mana after each Kill", "30% Deadly Strike"]
func splitOrBonuses(text string) []string {
	text = strings.TrimSpace(text)
	// Split on " or " (with spaces to avoid splitting words like "Armor")
	// But only if "or" appears between stat descriptions
	parts := strings.Split(text, " or \n")
	if len(parts) > 1 {
		var result []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	}
	return []string{text}
}
