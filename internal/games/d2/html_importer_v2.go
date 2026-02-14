package d2

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/storage"
)

// HTMLImporterV2 is the HTML-only import pipeline.
// All operations are idempotent upserts. Re-runs produce the same result with no duplicates.
type HTMLImporterV2 struct {
	repo              *Repository
	parser            *HTMLItemParser
	reverseTranslator *ReverseTranslator
	translator        *PropertyTranslator
	statRegistry      *StatRegistry
	storage           storage.Storage
	dryRun            bool
	iconsPath         string

	// Caches loaded from DB
	baseNameToCode    map[string]string
	runeNameToCode    map[string]string
	existingImageURLs map[string]bool // normalized name -> has image
	imageCache        map[string]string // imagePath -> uploaded URL
}

// NewHTMLImporterV2 creates a new HTML-only importer
func NewHTMLImporterV2(repo *Repository, statRegistry *StatRegistry, stor storage.Storage, dryRun bool) *HTMLImporterV2 {
	return &HTMLImporterV2{
		repo:              repo,
		parser:            NewHTMLItemParser(),
		reverseTranslator: NewReverseTranslator(),
		translator:        NewPropertyTranslator(),
		statRegistry:      statRegistry,
		storage:           stor,
		dryRun:            dryRun,
		imageCache:        make(map[string]string),
	}
}

// ImportAll runs the full HTML import pipeline
func (h *HTMLImporterV2) ImportAll(ctx context.Context, catalogPath string) (*ImportResult, error) {
	result := &ImportResult{}

	h.iconsPath = filepath.Join(catalogPath, "icons")
	pagesPath := filepath.Join(catalogPath, "pages")

	// Load caches
	fmt.Println("  Loading lookup caches from database...")
	if err := h.loadCaches(ctx); err != nil {
		return nil, fmt.Errorf("failed to load caches: %w", err)
	}
	fmt.Printf("    Base names: %d, Rune names: %d, Items with images: %d\n",
		len(h.baseNameToCode), len(h.runeNameToCode), len(h.existingImageURLs))

	// 1. Import bases
	if err := h.importBases(ctx, pagesPath, result); err != nil {
		return result, err
	}

	// 2. Reload base cache after importing new bases
	h.reloadBaseCache(ctx)

	// 3. Import misc (runes, gems, charms, jewels, keys) - before runewords so rune names resolve
	if err := h.importMisc(ctx, pagesPath, result); err != nil {
		return result, err
	}

	// 4. Reload rune cache after importing runes
	h.reloadRuneCache(ctx)

	// 5. Import uniques
	if err := h.importUniques(ctx, pagesPath, result); err != nil {
		return result, err
	}

	// 6. Import sets
	if err := h.importSets(ctx, pagesPath, result); err != nil {
		return result, err
	}

	// 7. Import runewords (needs rune name→code cache from step 4)
	if err := h.importRunewords(ctx, pagesPath, result); err != nil {
		return result, err
	}

	// 8. Link variants
	if err := h.linkVariants(ctx, pagesPath); err != nil {
		fmt.Printf("    Warning: variant linking failed: %v\n", err)
	}

	// 9. Compute runeword bases
	if err := h.computeRunewordBases(ctx, result); err != nil {
		return result, err
	}

	return result, nil
}

func (h *HTMLImporterV2) loadCaches(ctx context.Context) error {
	var err error

	h.baseNameToCode, err = h.repo.GetAllItemBaseNameToCode(ctx)
	if err != nil {
		return fmt.Errorf("base name map: %w", err)
	}

	h.runeNameToCode, err = h.repo.GetRuneNameToCodeMap(ctx)
	if err != nil {
		return fmt.Errorf("rune name map: %w", err)
	}

	// Load all names that have images (across all tables)
	h.existingImageURLs = make(map[string]bool)
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
			h.existingImageURLs[name] = true
		}
	}

	return nil
}

func (h *HTMLImporterV2) reloadBaseCache(ctx context.Context) {
	h.baseNameToCode, _ = h.repo.GetAllItemBaseNameToCode(ctx)
}

func (h *HTMLImporterV2) reloadRuneCache(ctx context.Context) {
	h.runeNameToCode, _ = h.repo.GetRuneNameToCodeMap(ctx)
}

// importBases parses base.html and upserts item_bases with tier/type_tags
func (h *HTMLImporterV2) importBases(ctx context.Context, pagesPath string, result *ImportResult) error {
	basePath := filepath.Join(pagesPath, "base.html")
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		fmt.Println("\n  No base.html found, skipping base import")
		return nil
	}

	fmt.Println("\n  Parsing base.html...")
	items, err := h.parser.ParseBasesFile(basePath)
	if err != nil {
		return fmt.Errorf("parse base.html: %w", err)
	}
	fmt.Printf("    Found %d base items\n", len(items))

	usedCodes := make(map[string]bool)
	for _, code := range h.baseNameToCode {
		usedCodes[code] = true
	}

	ensuredTypes := make(map[string]bool)
	baseErrors := 0

	for _, item := range items {
		// Resolve or generate code
		code := ""
		if existing, ok := h.baseNameToCode[item.Name]; ok {
			code = existing
		} else {
			code = generateBaseCode(item.Name)
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
		}
		usedCodes[code] = true

		// Determine category
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

		// Ensure item types exist
		for _, tc := range []struct{ code, name string }{{itemType, item.TypeName}, {itemType2, item.TypeName2}} {
			if tc.code != "" && !ensuredTypes[tc.code] {
				if !h.dryRun {
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

		// Detect class-specific from type tags
		classSpecific := classSpecificFromTypeTags(item.TypeTags)

		// Upload image (only if no existing image)
		imageURL := h.maybeUploadImage(ctx, item.ImagePath, "d2/base", item.Name, result)

		base := &ItemBase{
			Code:          code,
			Name:          item.Name,
			ItemType:      itemType,
			ItemType2:     itemType2,
			Category:      category,
			Tier:          item.Quality,
			TypeTags:      item.TypeTags,
			ClassSpecific: classSpecific,
			Tradable:      true,
			Level:         item.QualityLevel,
			LevelReq:      item.ReqLevel,
			StrReq:        item.ReqStr,
			DexReq:        item.ReqDex,
			Durability:    item.Durability,
			MinAC:         item.DefenseMin,
			MaxAC:         item.DefenseMax,
			MinDam:        item.OneHMinDam,
			MaxDam:        item.OneHMaxDam,
			TwoHandMinDam: item.TwoHMinDam,
			TwoHandMaxDam: item.TwoHMaxDam,
			RangeAdder:    item.RangeAdder,
			Speed:         item.Speed,
			MaxSockets:    item.MaxSockets,
			InvWidth:      item.InvWidth,
			InvHeight:     item.InvHeight,
			Spawnable:     true,
			Rarity:        1,
			ImageURL:      imageURL,
		}

		if !h.dryRun {
			if err := h.repo.UpsertItemBase(ctx, base); err != nil {
				fmt.Printf("    ERROR: base '%s' (code=%s, category=%s): %v\n", item.Name, code, category, err)
				baseErrors++
				continue
			}
		}
		result.ItemBases.Imported++
	}

	fmt.Printf("    Bases: %d imported, %d errors\n", result.ItemBases.Imported, baseErrors)
	return nil
}

// importUniques parses uniques.html and upserts unique_items by name
func (h *HTMLImporterV2) importUniques(ctx context.Context, pagesPath string, result *ImportResult) error {
	fmt.Println("\n  Parsing uniques.html...")
	items, err := h.parser.ParseUniquesFile(filepath.Join(pagesPath, "uniques.html"))
	if err != nil {
		return fmt.Errorf("parse uniques.html: %w", err)
	}
	fmt.Printf("    Found %d unique items\n", len(items))

	// Get next index ID for items that don't exist yet
	maxID, _ := h.repo.GetMaxIndexID(ctx, "unique_items")
	nextID := maxID + 1

	skipped := 0
	for _, item := range items {
		// Resolve base code
		baseCode := ""
		if item.BaseName != "" {
			if code, ok := h.baseNameToCode[item.BaseName]; ok {
				baseCode = code
			} else {
				fmt.Printf("    Warning: unique '%s' has unresolved base '%s'\n", item.Name, item.BaseName)
			}
		}

		// Reverse-translate properties and register stats
		properties := h.reverseTranslator.ReverseTranslateLines(item.Properties)
		properties = combineAllAttributes(properties, h.translator)
		for i := range properties {
			if properties[i].Code != "raw" {
				h.translator.EnrichProperty(&properties[i])
			}
			h.statRegistry.EnsureStat(ctx, properties[i])
		}

		imageURL := h.maybeUploadImage(ctx, item.ImagePath, "d2/unique", item.Name, result)

		unique := &UniqueItem{
			IndexID:    nextID,
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
		nextID++

		if !h.dryRun {
			if err := h.repo.UpsertUniqueItemByName(ctx, unique); err != nil {
				fmt.Printf("    ERROR: unique '%s': %v\n", item.Name, err)
				skipped++
				continue
			}
		}
		result.UniqueItems.Imported++
	}

	fmt.Printf("    Uniques: %d imported, %d errors\n", result.UniqueItems.Imported, skipped)
	return nil
}

// importSets parses sets.html and upserts set_bonuses and set_items by name
func (h *HTMLImporterV2) importSets(ctx context.Context, pagesPath string, result *ImportResult) error {
	fmt.Println("\n  Parsing sets.html...")
	setItems, fullSets, err := h.parser.ParseSetsFile(filepath.Join(pagesPath, "sets.html"))
	if err != nil {
		return fmt.Errorf("parse sets.html: %w", err)
	}
	fmt.Printf("    Found %d set items, %d full sets\n", len(setItems), len(fullSets))

	// Build full set bonus lookup
	fullSetMap := make(map[string]HTMLParsedFullSet)
	for _, fs := range fullSets {
		fullSetMap[fs.Name] = fs
	}

	// First pass: upsert set bonuses
	setNames := make(map[string]bool)
	maxSetID, _ := h.repo.GetMaxIndexID(ctx, "set_bonuses")
	nextSetID := maxSetID + 1

	for _, item := range setItems {
		if item.SetName == "" || setNames[item.SetName] {
			continue
		}
		setNames[item.SetName] = true

		// Translate full set bonuses if available
		var partialBonuses, fullBonuses []Property
		if fs, ok := fullSetMap[item.SetName]; ok {
			for _, line := range fs.PartialBonuses {
				prop := h.reverseTranslator.ReverseTranslate(line)
				if prop.Code != "raw" {
					h.translator.EnrichProperty(&prop)
				}
				h.statRegistry.EnsureStat(ctx, prop)
				partialBonuses = append(partialBonuses, prop)
			}
			for _, line := range fs.FullBonuses {
				prop := h.reverseTranslator.ReverseTranslate(line)
				if prop.Code != "raw" {
					h.translator.EnrichProperty(&prop)
				}
				h.statRegistry.EnsureStat(ctx, prop)
				fullBonuses = append(fullBonuses, prop)
			}
		}

		setBonus := &SetBonus{
			IndexID:        nextSetID,
			Name:           item.SetName,
			PartialBonuses: partialBonuses,
			FullBonuses:    fullBonuses,
		}
		nextSetID++

		if !h.dryRun {
			if err := h.repo.UpsertSetBonus(ctx, setBonus); err != nil {
				fmt.Printf("    Error upserting set %s: %v\n", item.SetName, err)
				continue
			}
		}
		result.SetBonuses.Imported++
	}

	// Second pass: upsert set items
	maxItemID, _ := h.repo.GetMaxIndexID(ctx, "set_items")
	nextItemID := maxItemID + 1

	setItemErrors := 0
	for _, item := range setItems {
		baseCode := ""
		if item.BaseName != "" {
			if code, ok := h.baseNameToCode[item.BaseName]; ok {
				baseCode = code
			} else {
				fmt.Printf("    Warning: set item '%s' has unresolved base '%s'\n", item.Name, item.BaseName)
			}
		}

		// Reverse-translate properties
		properties := h.reverseTranslator.ReverseTranslateLines(item.Properties)
		properties = combineAllAttributes(properties, h.translator)
		for i := range properties {
			if properties[i].Code != "raw" {
				h.translator.EnrichProperty(&properties[i])
			}
			h.statRegistry.EnsureStat(ctx, properties[i])
		}

		// Reverse-translate set bonuses
		var bonusProperties []Property
		for _, bonus := range item.SetBonuses {
			bonusLines := splitOrBonuses(bonus.Text)
			for _, line := range bonusLines {
				prop := h.reverseTranslator.ReverseTranslate(line)
				if prop.Code != "raw" {
					h.translator.EnrichProperty(&prop)
				}
				h.statRegistry.EnsureStat(ctx, prop)
				bonusProperties = append(bonusProperties, prop)
			}
		}
		bonusProperties = combineAllAttributes(bonusProperties, h.translator)

		imageURL := h.maybeUploadImage(ctx, item.ImagePath, "d2/set", item.Name, result)

		setItem := &SetItem{
			IndexID:         nextItemID,
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
		nextItemID++

		if !h.dryRun {
			if err := h.repo.UpsertSetItemByName(ctx, setItem); err != nil {
				fmt.Printf("    ERROR: set item '%s': %v\n", item.Name, err)
				setItemErrors++
				continue
			}
		}
		result.SetItems.Imported++
	}

	fmt.Printf("    Sets: %d imported, Set items: %d imported, %d errors\n", result.SetBonuses.Imported, result.SetItems.Imported, setItemErrors)
	return nil
}

// importRunewords parses runewords.html and upserts runewords by name
func (h *HTMLImporterV2) importRunewords(ctx context.Context, pagesPath string, result *ImportResult) error {
	fmt.Println("\n  Parsing runewords.html...")
	runewords, err := h.parser.ParseRunewordsFile(filepath.Join(pagesPath, "runewords.html"))
	if err != nil {
		return fmt.Errorf("parse runewords.html: %w", err)
	}
	fmt.Printf("    Found %d runewords\n", len(runewords))

	skippedRW := 0
	for _, rw := range runewords {
		// Resolve rune names to codes
		var runeCodes []string
		var unresolvedRunes []string
		for _, runeName := range rw.Runes {
			if code, ok := h.runeNameToCode[runeName]; ok {
				runeCodes = append(runeCodes, code)
			} else {
				unresolvedRunes = append(unresolvedRunes, runeName)
			}
		}
		if len(unresolvedRunes) > 0 {
			fmt.Printf("    SKIP runeword '%s': unresolved runes %v (available: %d rune names in cache)\n", rw.Name, unresolvedRunes, len(h.runeNameToCode))
			skippedRW++
			continue
		}

		// Store valid types as tag names (not codes) for type_tags matching
		validTypes := rw.ValidTypes

		// Reverse-translate properties
		properties := h.reverseTranslator.ReverseTranslateLines(rw.Properties)
		properties = combineAllAttributes(properties, h.translator)
		for i := range properties {
			if properties[i].Code != "raw" {
				h.translator.EnrichProperty(&properties[i])
			}
			h.statRegistry.EnsureStat(ctx, properties[i])
		}

		internalName := fmt.Sprintf("HTMLRuneword_%s", strings.ReplaceAll(rw.Name, " ", ""))

		runeword := &Runeword{
			Name:           internalName,
			DisplayName:    rw.Name,
			Complete:       true,
			ValidItemTypes: validTypes,
			Runes:          runeCodes,
			Properties:     properties,
		}

		if !h.dryRun {
			if err := h.repo.UpsertRuneword(ctx, runeword); err != nil {
				fmt.Printf("    Error upserting runeword %s: %v\n", rw.Name, err)
				continue
			}
		}
		result.Runewords.Imported++
	}

	fmt.Printf("    Runewords: %d imported, %d skipped\n", result.Runewords.Imported, skippedRW)
	return nil
}

// importMisc parses misc.html and upserts runes, gems, and misc items
func (h *HTMLImporterV2) importMisc(ctx context.Context, pagesPath string, result *ImportResult) error {
	miscPath := filepath.Join(pagesPath, "misc.html")
	if _, err := os.Stat(miscPath); os.IsNotExist(err) {
		fmt.Println("\n  No misc.html found, skipping misc import")
		return nil
	}

	fmt.Println("\n  Parsing misc.html...")
	runes, gems, miscItems, err := h.parser.ParseMiscFile(miscPath)
	if err != nil {
		return fmt.Errorf("parse misc.html: %w", err)
	}
	fmt.Printf("    Found %d runes, %d gems, %d misc items\n", len(runes), len(gems), len(miscItems))

	// Import runes
	runeErrors := 0
	for _, rn := range runes {
		code := ""
		if c, ok := h.runeNameToCode[rn.Name]; ok {
			code = c
		} else {
			code = fmt.Sprintf("r%02d", rn.RuneIndex)
		}

		weaponMods := h.translateAndRegisterMods(ctx, rn.WeaponMods)
		helmMods := h.translateAndRegisterMods(ctx, rn.HelmMods)
		shieldMods := h.translateAndRegisterMods(ctx, rn.ShieldMods)

		imageURL := h.maybeUploadImage(ctx, rn.ImagePath, "d2/rune", rn.Name, result)

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
				fmt.Printf("    ERROR: rune '%s' (code=%s): %v\n", rn.Name, code, err)
				runeErrors++
				continue
			}
		}
		result.Runes.Imported++
	}
	fmt.Printf("    Runes: %d imported, %d errors\n", result.Runes.Imported, runeErrors)

	// Import gems
	gemErrors := 0
	for _, gem := range gems {
		gemType, quality := parseGemNameParts(gem.Name)
		code := generateBaseCode(gem.Name)

		weaponMods := h.translateAndRegisterMods(ctx, gem.WeaponMods)
		helmMods := h.translateAndRegisterMods(ctx, gem.HelmMods)
		shieldMods := h.translateAndRegisterMods(ctx, gem.ShieldMods)

		imageURL := h.maybeUploadImage(ctx, gem.ImagePath, "d2/gem", gem.Name, result)

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
				fmt.Printf("    ERROR: gem '%s' (code=%s, type=%s, quality=%s): %v\n", gem.Name, code, gemType, quality, err)
				gemErrors++
				continue
			}
		}
		result.Gems.Imported++
	}
	fmt.Printf("    Gems: %d imported, %d errors\n", result.Gems.Imported, gemErrors)

	// Import misc items as item_bases
	usedCodes := make(map[string]bool)
	for _, code := range h.baseNameToCode {
		usedCodes[code] = true
	}

	miscErrors := 0
	for _, item := range miscItems {
		code := ""
		if existing, ok := h.baseNameToCode[item.Name]; ok {
			code = existing
		} else {
			code = generateBaseCode(item.Name)
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
		}
		usedCodes[code] = true

		imageURL := h.maybeUploadImage(ctx, item.ImagePath, "d2/misc", item.Name, result)

		base := &ItemBase{
			Code:        code,
			Name:        item.Name,
			Category:    "misc",
			Tier:        "Normal",
			Tradable:    true,
			Spawnable:   true,
			Rarity:      1,
			Description: item.Description,
			ImageURL:    imageURL,
		}

		if !h.dryRun {
			if err := h.repo.UpsertItemBase(ctx, base); err != nil {
				fmt.Printf("    ERROR: misc '%s' (code=%s): %v\n", item.Name, code, err)
				miscErrors++
				continue
			}
		}
		result.ItemBases.Imported++
	}
	fmt.Printf("    Misc items: %d imported, %d errors\n", len(miscItems)-miscErrors, miscErrors)

	return nil
}

// linkVariants links normal/exceptional/elite base item variants
func (h *HTMLImporterV2) linkVariants(ctx context.Context, pagesPath string) error {
	basePath := filepath.Join(pagesPath, "base.html")
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return nil
	}

	items, err := h.parser.ParseBasesFile(basePath)
	if err != nil {
		return err
	}

	// Reload to get all codes
	h.reloadBaseCache(ctx)

	// Build name -> code map for variant linking
	for _, item := range items {
		if len(item.VariantNames) == 0 {
			continue
		}

		myCode, ok := h.baseNameToCode[item.Name]
		if !ok {
			continue
		}

		for _, variant := range item.VariantNames {
			variantCode, ok := h.baseNameToCode[variant.Name]
			if !ok {
				continue
			}

			// Set the variant link based on the variant's tier
			normalCode, exceptionalCode, eliteCode := "", "", ""
			switch variant.Tier {
			case "Normal":
				normalCode = variantCode
			case "Exceptional":
				exceptionalCode = variantCode
			case "Elite":
				eliteCode = variantCode
			}

			if !h.dryRun {
				h.repo.UpdateItemBaseVariants(ctx, myCode, normalCode, exceptionalCode, eliteCode)
			}
		}
	}

	return nil
}

// computeRunewordBases computes valid base items for each runeword using type_tags overlap
func (h *HTMLImporterV2) computeRunewordBases(ctx context.Context, result *ImportResult) error {
	fmt.Println("\n  Computing runeword bases...")

	runewords, err := h.repo.GetAllRunewordsForMatching(ctx)
	if err != nil {
		return fmt.Errorf("get runewords: %w", err)
	}

	if !h.dryRun {
		h.repo.ClearRunewordBases(ctx)
	}

	count := 0
	for _, rw := range runewords {
		if len(rw.ValidItemTypes) == 0 {
			continue
		}

		// Query bases that have matching type_tags and enough sockets
		bases, err := h.repo.GetBasesForRunewordByTypeTags(ctx, rw.ValidItemTypes, rw.RuneCount)
		if err != nil {
			fmt.Printf("    Warning: runeword base query failed for %s: %v\n", rw.Name, err)
			continue
		}

		for _, base := range bases {
			rb := &RunewordBase{
				RunewordID:      rw.ID,
				ItemBaseID:      base.ID,
				ItemBaseCode:    base.Code,
				ItemBaseName:    base.Name,
				Category:        base.Category,
				MaxSockets:      base.MaxSockets,
				RequiredSockets: rw.RuneCount,
			}

			if !h.dryRun {
				h.repo.InsertRunewordBase(ctx, rb)
			}
			count++
		}
	}

	result.RunewordBases.Imported = count
	fmt.Printf("    Runeword bases: %d computed\n", count)
	return nil
}

// translateAndRegisterMods reverse-translates mod text lines and registers stats
func (h *HTMLImporterV2) translateAndRegisterMods(ctx context.Context, lines []string) []Property {
	mods := h.reverseTranslator.ReverseTranslateLines(lines)
	for i := range mods {
		if mods[i].Code != "raw" {
			h.translator.EnrichProperty(&mods[i])
		}
		h.statRegistry.EnsureStat(ctx, mods[i])
	}
	return mods
}

// maybeUploadImage uploads an image only if the item doesn't already have one
func (h *HTMLImporterV2) maybeUploadImage(ctx context.Context, imagePath, category, itemName string, result *ImportResult) string {
	if imagePath == "" || h.storage == nil {
		return ""
	}

	normalized := NormalizeItemName(itemName)
	if h.existingImageURLs[normalized] {
		return "" // already has image
	}

	if url, ok := h.imageCache[imagePath]; ok {
		return url
	}

	imageFilename := filepath.Base(imagePath)
	localPath := filepath.Join(h.iconsPath, imageFilename)
	data, err := os.ReadFile(localPath)
	if err != nil {
		data = h.findImageFileCaseInsensitive(imageFilename)
		if data == nil {
			if result != nil {
				result.ImagesMissing++
			}
			return ""
		}
	}

	storagePath := storage.StoragePath(category, itemName)

	if h.dryRun {
		url := fmt.Sprintf("[dry-run] %s", storagePath)
		h.imageCache[imagePath] = url
		if result != nil {
			result.ImagesUploaded++
		}
		return url
	}

	publicURL, err := h.storage.UploadImage(ctx, storagePath, data, "image/png")
	if err != nil {
		fmt.Printf("    Error uploading image for %s: %v\n", itemName, err)
		return ""
	}

	h.imageCache[imagePath] = publicURL
	if result != nil {
		result.ImagesUploaded++
	}
	return publicURL
}

// combineAllAttributes detects when str, dex, vit, and enr all share the same
// min/max values and replaces them with a single "all-stats" property.
// If values differ or not all 4 are present, the properties are returned unchanged.
func combineAllAttributes(props []Property, translator *PropertyTranslator) []Property {
	attrCodes := map[string]int{"str": -1, "dex": -1, "vit": -1, "enr": -1}
	for i, p := range props {
		if _, ok := attrCodes[p.Code]; ok {
			attrCodes[p.Code] = i
		}
	}

	// Check all 4 are present
	for _, idx := range attrCodes {
		if idx == -1 {
			return props
		}
	}

	// Check all share the same min/max
	ref := props[attrCodes["str"]]
	for _, code := range []string{"dex", "vit", "enr"} {
		p := props[attrCodes[code]]
		if p.Min != ref.Min || p.Max != ref.Max {
			return props
		}
	}

	// Build replacement: keep all non-attribute props, insert all-stats at first attribute position
	firstIdx := len(props)
	for _, idx := range attrCodes {
		if idx < firstIdx {
			firstIdx = idx
		}
	}

	removeSet := map[int]bool{
		attrCodes["str"]: true,
		attrCodes["dex"]: true,
		attrCodes["vit"]: true,
		attrCodes["enr"]: true,
	}

	allStats := Property{
		Code: "all-stats",
		Min:  ref.Min,
		Max:  ref.Max,
	}
	translator.EnrichProperty(&allStats)

	result := make([]Property, 0, len(props)-3)
	inserted := false
	for i, p := range props {
		if removeSet[i] {
			if i == firstIdx {
				result = append(result, allStats)
				inserted = true
			}
			continue
		}
		result = append(result, p)
	}
	if !inserted {
		result = append(result, allStats)
	}

	return result
}

// parseGemNameParts extracts gem type and quality from a gem name
func parseGemNameParts(name string) (gemType, quality string) {
	nameLower := strings.ToLower(name)

	// Determine quality
	switch {
	case strings.HasPrefix(nameLower, "chipped"):
		quality = "chipped"
	case strings.HasPrefix(nameLower, "flawed"):
		quality = "flawed"
	case strings.HasPrefix(nameLower, "flawless"):
		quality = "flawless"
	case strings.HasPrefix(nameLower, "perfect"):
		quality = "perfect"
	default:
		quality = "normal"
	}

	// Determine gem type
	switch {
	case strings.Contains(nameLower, "amethyst"):
		gemType = "amethyst"
	case strings.Contains(nameLower, "sapphire"):
		gemType = "sapphire"
	case strings.Contains(nameLower, "emerald"):
		gemType = "emerald"
	case strings.Contains(nameLower, "ruby"):
		gemType = "ruby"
	case strings.Contains(nameLower, "diamond"):
		gemType = "diamond"
	case strings.Contains(nameLower, "topaz"):
		gemType = "topaz"
	case strings.Contains(nameLower, "skull"):
		gemType = "skull"
	default:
		gemType = "unknown"
	}

	return
}

func (h *HTMLImporterV2) findImageFileCaseInsensitive(filename string) []byte {
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
