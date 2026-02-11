package d2

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/cache"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/database"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/parser"
)

// forceSpawnableCodes are misc items that D2 marks as non-spawnable (can't drop randomly)
// but are legitimate tradeable items that drop from specific sources.
var forceSpawnableCodes = map[string]bool{
	"tes": true, // Twisted Essence of Suffering
	"ceh": true, // Charged Essence of Hatred
	"bet": true, // Burning Essence of Terror
	"fed": true, // Festering Essence of Destruction
	"toa": true, // Token of Absolution
}

// nameOverrides fixes known typos in D2 data files
var nameOverrides = map[string]string{
	"ceh": "Charged Essence of Hatred", // Data file has "Essense"
}

type Importer struct {
	repo       *Repository
	cache      *cache.RedisCache
	translator *PropertyTranslator
}

func NewImporter(db *database.DB, redisCache *cache.RedisCache) *Importer {
	return &Importer{
		repo:       NewRepository(db.Pool()),
		cache:      redisCache,
		translator: NewPropertyTranslator(),
	}
}

func (i *Importer) Import(ctx context.Context, catalogPath string) (*ImportResult, error) {
	result := &ImportResult{}

	// Import in dependency order
	fmt.Println("  Importing item types...")
	if stats, err := i.importItemTypes(ctx, catalogPath); err != nil {
		return nil, fmt.Errorf("item types import failed: %w", err)
	} else {
		result.ItemTypes = stats
	}

	fmt.Println("  Importing properties...")
	if stats, err := i.importProperties(ctx, catalogPath); err != nil {
		return nil, fmt.Errorf("properties import failed: %w", err)
	} else {
		result.Properties = stats
	}

	fmt.Println("  Importing armor bases...")
	armorStats, err := i.importArmorBases(ctx, catalogPath)
	if err != nil {
		return nil, fmt.Errorf("armor bases import failed: %w", err)
	}

	fmt.Println("  Importing weapon bases...")
	weaponStats, err := i.importWeaponBases(ctx, catalogPath)
	if err != nil {
		return nil, fmt.Errorf("weapon bases import failed: %w", err)
	}

	fmt.Println("  Importing misc items...")
	miscStats, err := i.importMiscItems(ctx, catalogPath)
	if err != nil {
		return nil, fmt.Errorf("misc items import failed: %w", err)
	}

	result.ItemBases = ImportStats{
		Imported: armorStats.Imported + weaponStats.Imported + miscStats.Imported,
		Skipped:  armorStats.Skipped + weaponStats.Skipped + miscStats.Skipped,
	}

	fmt.Println("  Importing unique items...")
	if stats, err := i.importUniqueItems(ctx, catalogPath); err != nil {
		return nil, fmt.Errorf("unique items import failed: %w", err)
	} else {
		result.UniqueItems = stats
	}

	fmt.Println("  Importing sets...")
	if stats, err := i.importSets(ctx, catalogPath); err != nil {
		return nil, fmt.Errorf("sets import failed: %w", err)
	} else {
		result.SetBonuses = stats
	}

	fmt.Println("  Importing set items...")
	if stats, err := i.importSetItems(ctx, catalogPath); err != nil {
		return nil, fmt.Errorf("set items import failed: %w", err)
	} else {
		result.SetItems = stats
	}

	fmt.Println("  Importing runes...")
	if stats, err := i.importRunes(ctx, catalogPath); err != nil {
		return nil, fmt.Errorf("runes import failed: %w", err)
	} else {
		result.Runes = stats
	}

	fmt.Println("  Importing runewords...")
	if stats, err := i.importRunewords(ctx, catalogPath); err != nil {
		return nil, fmt.Errorf("runewords import failed: %w", err)
	} else {
		result.Runewords = stats
	}

	fmt.Println("  Importing gems...")
	if stats, err := i.importGems(ctx, catalogPath); err != nil {
		return nil, fmt.Errorf("gems import failed: %w", err)
	} else {
		result.Gems = stats
	}

	fmt.Println("  Importing affixes...")
	prefixStats, err := i.importAffixes(ctx, catalogPath, "magicprefix.txt", "prefix")
	if err != nil {
		return nil, fmt.Errorf("prefix import failed: %w", err)
	}
	suffixStats, err := i.importAffixes(ctx, catalogPath, "magicsuffix.txt", "suffix")
	if err != nil {
		return nil, fmt.Errorf("suffix import failed: %w", err)
	}
	result.Affixes = ImportStats{
		Imported: prefixStats.Imported + suffixStats.Imported,
		Skipped:  prefixStats.Skipped + suffixStats.Skipped,
	}

	fmt.Println("  Importing treasure classes...")
	if stats, err := i.importTreasureClasses(ctx, catalogPath); err != nil {
		return nil, fmt.Errorf("treasure classes import failed: %w", err)
	} else {
		result.TreasureClasses = stats
	}

	fmt.Println("  Importing item ratios...")
	if stats, err := i.importItemRatios(ctx, catalogPath); err != nil {
		return nil, fmt.Errorf("item ratios import failed: %w", err)
	} else {
		result.ItemRatios = stats
	}

	fmt.Println("  Building runeword bases...")
	if stats, err := i.buildRunewordBases(ctx); err != nil {
		return nil, fmt.Errorf("runeword bases build failed: %w", err)
	} else {
		result.RunewordBases = stats
	}

	// Clear cache after import
	if i.cache != nil {
		i.cache.DeleteByPattern(ctx, "d2:*")
	}

	return result, nil
}

func (i *Importer) importItemTypes(ctx context.Context, catalogPath string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, "itemtypes.txt"))
	if err != nil {
		return stats, err
	}

	for _, row := range p.Rows() {
		r := parser.AsRow(row)
		code := r.GetString("Code", "")
		if code == "" || code == "none" {
			stats.Skipped++
			continue
		}

		itemType := &ItemType{
			Code:                 code,
			Name:                 r.GetString("ItemType", code),
			Equiv1:               r.GetString("Equiv1", ""),
			Equiv2:               r.GetString("Equiv2", ""),
			BodyLoc1:             r.GetString("BodyLoc1", ""),
			BodyLoc2:             r.GetString("BodyLoc2", ""),
			CanBeMagic:           r.GetBool("Magic"),
			CanBeRare:            r.GetBool("Rare"),
			MaxSocketsNormal:     r.GetInt("MaxSockets1", 0),
			MaxSocketsNightmare:  r.GetInt("MaxSockets2", 0),
			MaxSocketsHell:       r.GetInt("MaxSockets3", 0),
			StaffMods:            r.GetString("StaffMods", ""),
			ClassRestriction:     r.GetString("Class", ""),
			StorePage:            r.GetString("StorePage", ""),
		}

		if err := i.repo.UpsertItemType(ctx, itemType); err != nil {
			return stats, fmt.Errorf("failed to upsert item type %s: %w", code, err)
		}
		stats.Imported++
	}

	return stats, nil
}

func (i *Importer) importArmorBases(ctx context.Context, catalogPath string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, "armor.txt"))
	if err != nil {
		return stats, err
	}

	for _, row := range p.Rows() {
		r := parser.AsRow(row)
		code := r.GetString("code", "")
		name := r.GetString("name", "")
		if code == "" || name == "" {
			stats.Skipped++
			continue
		}

		// Skip expansion placeholder items
		if strings.HasPrefix(name, "Expansion") {
			stats.Skipped++
			continue
		}

		itemBase := &ItemBase{
			Code:            code,
			Name:            name,
			ItemType:        r.GetString("type", ""),
			ItemType2:       r.GetString("type2", ""),
			Category:        "armor",
			Level:           r.GetInt("level", 0),
			LevelReq:        r.GetInt("levelreq", 0),
			StrReq:          r.GetInt("reqstr", 0),
			DexReq:          r.GetInt("reqdex", 0),
			Durability:      r.GetInt("durability", 0),
			MinAC:           r.GetInt("minac", 0),
			MaxAC:           r.GetInt("maxac", 0),
			MaxSockets:      r.GetInt("gemsockets", 0),
			GemApplyType:    r.GetInt("gemapplytype", 0),
			NormalCode:      r.GetString("normcode", ""),
			ExceptionalCode: r.GetString("ubercode", ""),
			EliteCode:       r.GetString("ultracode", ""),
			InvWidth:        r.GetInt("invwidth", 1),
			InvHeight:       r.GetInt("invheight", 1),
			InvFile:         r.GetString("invfile", ""),
			FlippyFile:      r.GetString("flippyfile", ""),
			UniqueInvFile:   r.GetString("uniqueinvfile", ""),
			SetInvFile:      r.GetString("setinvfile", ""),
			Spawnable:       r.GetBool("spawnable"),
			Rarity:          r.GetInt("rarity", 1),
			Cost:            r.GetInt("cost", 0),
			Speed:           r.GetInt("speed", 0),
		}

		if err := i.repo.UpsertItemBase(ctx, itemBase); err != nil {
			return stats, fmt.Errorf("failed to upsert armor base %s: %w", code, err)
		}
		stats.Imported++
	}

	return stats, nil
}

func (i *Importer) importWeaponBases(ctx context.Context, catalogPath string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, "weapons.txt"))
	if err != nil {
		return stats, err
	}

	for _, row := range p.Rows() {
		r := parser.AsRow(row)
		code := r.GetString("code", "")
		name := r.GetString("name", "")
		if code == "" || name == "" {
			stats.Skipped++
			continue
		}

		// Skip expansion placeholder items
		if strings.HasPrefix(name, "Expansion") {
			stats.Skipped++
			continue
		}

		itemBase := &ItemBase{
			Code:            code,
			Name:            name,
			ItemType:        r.GetString("type", ""),
			ItemType2:       r.GetString("type2", ""),
			Category:        "weapon",
			Level:           r.GetInt("level", 0),
			LevelReq:        r.GetInt("levelreq", 0),
			StrReq:          r.GetInt("reqstr", 0),
			DexReq:          r.GetInt("reqdex", 0),
			Durability:      r.GetInt("durability", 0),
			MinDam:          r.GetInt("mindam", 0),
			MaxDam:          r.GetInt("maxdam", 0),
			TwoHandMinDam:   r.GetInt("2handmindam", 0),
			TwoHandMaxDam:   r.GetInt("2handmaxdam", 0),
			RangeAdder:      r.GetInt("rangeadder", 0),
			Speed:           r.GetInt("speed", 0),
			StrBonus:        r.GetInt("StrBonus", 0),
			DexBonus:        r.GetInt("DexBonus", 0),
			MaxSockets:      r.GetInt("gemsockets", 0),
			GemApplyType:    r.GetInt("gemapplytype", 0),
			NormalCode:      r.GetString("normcode", ""),
			ExceptionalCode: r.GetString("ubercode", ""),
			EliteCode:       r.GetString("ultracode", ""),
			InvWidth:        r.GetInt("invwidth", 1),
			InvHeight:       r.GetInt("invheight", 1),
			InvFile:         r.GetString("invfile", ""),
			FlippyFile:      r.GetString("flippyfile", ""),
			UniqueInvFile:   r.GetString("uniqueinvfile", ""),
			SetInvFile:      r.GetString("setinvfile", ""),
			Spawnable:       r.GetBool("spawnable"),
			Stackable:       r.GetBool("stackable"),
			Throwable:       r.GetBool("Throwable"),
			Rarity:          r.GetInt("rarity", 1),
			Cost:            r.GetInt("cost", 0),
		}

		if err := i.repo.UpsertItemBase(ctx, itemBase); err != nil {
			return stats, fmt.Errorf("failed to upsert weapon base %s: %w", code, err)
		}
		stats.Imported++
	}

	return stats, nil
}

func (i *Importer) importMiscItems(ctx context.Context, catalogPath string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, "misc.txt"))
	if err != nil {
		return stats, err
	}

	for _, row := range p.Rows() {
		r := parser.AsRow(row)
		code := r.GetString("code", "")
		name := r.GetString("name", "")
		if code == "" || name == "" {
			stats.Skipped++
			continue
		}

		// Skip expansion placeholder items
		if strings.HasPrefix(name, "Expansion") || name == "Not Used" {
			stats.Skipped++
			continue
		}

		// Apply name overrides for known typos
		if override, ok := nameOverrides[code]; ok {
			name = override
		}

		spawnable := r.GetBool("spawnable")
		// Force-enable spawnable for tradeable items that don't drop randomly
		if forceSpawnableCodes[code] {
			spawnable = true
		}

		itemBase := &ItemBase{
			Code:         code,
			Name:         name,
			ItemType:     r.GetString("type", ""),
			ItemType2:    r.GetString("type2", ""),
			Category:     "misc",
			Level:        r.GetInt("level", 0),
			LevelReq:     r.GetInt("levelreq", 0),
			StrReq:       r.GetInt("reqstr", 0),
			DexReq:       r.GetInt("reqdex", 0),
			InvWidth:     r.GetInt("invwidth", 1),
			InvHeight:    r.GetInt("invheight", 1),
			InvFile:      r.GetString("invfile", ""),
			FlippyFile:   r.GetString("flippyfile", ""),
			Spawnable:    spawnable,
			Stackable:    r.GetBool("stackable"),
			Useable:      r.GetBool("useable"),
			QuestItem:    r.GetInt("quest", 0) > 0,
			Rarity:       r.GetInt("rarity", 1),
			Cost:         r.GetInt("cost", 0),
		}

		if err := i.repo.UpsertItemBase(ctx, itemBase); err != nil {
			return stats, fmt.Errorf("failed to upsert misc item %s: %w", code, err)
		}
		stats.Imported++
	}

	return stats, nil
}

func (i *Importer) importUniqueItems(ctx context.Context, catalogPath string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, "uniqueitems.txt"))
	if err != nil {
		return stats, err
	}

	rows := p.Rows()
	for rowIdx, row := range rows {
		r := parser.AsRow(row)
		// Note: In D2 data files, "index" column contains the name, "*ID" contains numeric index
		name := r.GetString("index", "")
		indexID := r.GetInt("*ID", -1)
		if name == "" {
			stats.Skipped++
			continue
		}
		// Use row position if *ID is missing or invalid
		if indexID < 0 {
			indexID = rowIdx
		}

		// Parse properties
		properties := parseProperties(r, 12, i.translator)
		properties = combineAllAttributes(properties, i.translator)

		// Parse ladder seasons
		var firstSeason, lastSeason *int
		if fs := r.GetInt("firstLadderSeason", 0); fs > 0 {
			firstSeason = &fs
		}
		if ls := r.GetInt("lastLadderSeason", 0); ls > 0 {
			lastSeason = &ls
		}

		unique := &UniqueItem{
			IndexID:           indexID,
			Name:              name,
			BaseCode:          r.GetString("code", ""),
			BaseName:          r.GetString("*ItemName", ""),
			Level:             r.GetInt("lvl", 0),
			LevelReq:          r.GetInt("lvl req", 0),
			Rarity:            r.GetInt("rarity", 1),
			Enabled:           r.GetBool("enabled"),
			LadderOnly:        firstSeason != nil,
			FirstLadderSeason: firstSeason,
			LastLadderSeason:  lastSeason,
			Properties:        properties,
			InvTransform:      r.GetString("invtransform", ""),
			ChrTransform:      r.GetString("chrtransform", ""),
			InvFile:           r.GetString("invfile", ""),
			CostMult:          r.GetInt("cost mult", 0),
			CostAdd:           r.GetInt("cost add", 0),
		}

		// Disable quest items: base code references a non-spawnable or missing base item
		if unique.Enabled && unique.BaseCode != "" {
			base, err := i.repo.GetItemBaseByCode(ctx, unique.BaseCode)
			if err != nil || !base.Spawnable {
				unique.Enabled = false
			}
		}

		if err := i.repo.UpsertUniqueItem(ctx, unique); err != nil {
			return stats, fmt.Errorf("failed to upsert unique item %s: %w", name, err)
		}
		stats.Imported++
	}

	return stats, nil
}

func (i *Importer) importSets(ctx context.Context, catalogPath string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, "sets.txt"))
	if err != nil {
		return stats, err
	}

	for idx, row := range p.Rows() {
		r := parser.AsRow(row)
		name := r.GetString("name", "")
		if name == "" || name == "Expansion" {
			stats.Skipped++
			continue
		}

		// Parse partial bonuses (PCode2a-PCode5b)
		partialBonuses := parseSetPartialBonuses(r, i.translator)
		partialBonuses = combineAllAttributes(partialBonuses, i.translator)

		// Parse full set bonuses (FCode1-FCode8)
		fullBonuses := parseSetFullBonuses(r, i.translator)
		fullBonuses = combineAllAttributes(fullBonuses, i.translator)

		setBonus := &SetBonus{
			IndexID:        idx,
			Name:           name,
			Version:        r.GetInt("version", 0),
			PartialBonuses: partialBonuses,
			FullBonuses:    fullBonuses,
		}

		if err := i.repo.UpsertSetBonus(ctx, setBonus); err != nil {
			return stats, fmt.Errorf("failed to upsert set %s: %w", name, err)
		}
		stats.Imported++
	}

	return stats, nil
}

func (i *Importer) importSetItems(ctx context.Context, catalogPath string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, "setitems.txt"))
	if err != nil {
		return stats, err
	}

	rows := p.Rows()
	for rowIdx, row := range rows {
		r := parser.AsRow(row)
		// Note: In D2 data files, "index" column contains the name, "*ID" contains numeric index
		name := r.GetString("index", "")
		indexID := r.GetInt("*ID", -1)
		if name == "" {
			stats.Skipped++
			continue
		}
		// Use row position if *ID is missing or invalid
		if indexID < 0 {
			indexID = rowIdx
		}

		// Parse base properties
		properties := parseSetItemProperties(r, i.translator)
		properties = combineAllAttributes(properties, i.translator)

		// Parse bonus properties (activated by wearing more set items)
		bonusProperties := parseSetItemBonusProperties(r, i.translator)
		bonusProperties = combineAllAttributes(bonusProperties, i.translator)

		setItem := &SetItem{
			IndexID:         indexID,
			Name:            name,
			SetName:         r.GetString("set", ""),
			BaseCode:        r.GetString("item", ""),
			BaseName:        r.GetString("*ItemName", ""),
			Level:           r.GetInt("lvl", 0),
			LevelReq:        r.GetInt("lvl req", 0),
			Rarity:          r.GetInt("rarity", 1),
			Properties:      properties,
			BonusProperties: bonusProperties,
			InvTransform:    r.GetString("invtransform", ""),
			ChrTransform:    r.GetString("chrtransform", ""),
			InvFile:         r.GetString("invfile", ""),
			CostMult:        r.GetInt("cost mult", 0),
			CostAdd:         r.GetInt("cost add", 0),
		}

		if err := i.repo.UpsertSetItem(ctx, setItem); err != nil {
			return stats, fmt.Errorf("failed to upsert set item %s: %w", name, err)
		}
		stats.Imported++
	}

	return stats, nil
}

func (i *Importer) importRunewords(ctx context.Context, catalogPath string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, "runes.txt"))
	if err != nil {
		return stats, err
	}

	for _, row := range p.Rows() {
		r := parser.AsRow(row)
		name := r.GetString("Name", "")
		displayName := r.GetString("*Rune Name", "")
		complete := r.GetBool("complete")

		if name == "" || !complete {
			stats.Skipped++
			continue
		}

		// Parse valid item types
		validTypes := []string{}
		for j := 1; j <= 6; j++ {
			itype := r.GetString(fmt.Sprintf("itype%d", j), "")
			if itype != "" {
				validTypes = append(validTypes, itype)
			}
		}

		// Parse excluded item types
		excludedTypes := []string{}
		for j := 1; j <= 3; j++ {
			etype := r.GetString(fmt.Sprintf("etype%d", j), "")
			if etype != "" {
				excludedTypes = append(excludedTypes, etype)
			}
		}

		// Parse runes
		runes := []string{}
		for j := 1; j <= 6; j++ {
			rune := r.GetString(fmt.Sprintf("Rune%d", j), "")
			if rune != "" {
				runes = append(runes, rune)
			}
		}

		// Parse properties
		properties := parseRunewordProperties(r, i.translator)
		properties = combineAllAttributes(properties, i.translator)

		// Parse ladder seasons
		var firstSeason, lastSeason *int
		if fs := r.GetInt("firstLadderSeason", 0); fs > 0 {
			firstSeason = &fs
		}
		if ls := r.GetInt("lastLadderSeason", 0); ls > 0 {
			lastSeason = &ls
		}

		runeword := &Runeword{
			Name:              name,
			DisplayName:       displayName,
			Complete:          complete,
			LadderOnly:        firstSeason != nil,
			FirstLadderSeason: firstSeason,
			LastLadderSeason:  lastSeason,
			ValidItemTypes:    validTypes,
			ExcludedItemTypes: excludedTypes,
			Runes:             runes,
			Properties:        properties,
		}

		if err := i.repo.UpsertRuneword(ctx, runeword); err != nil {
			return stats, fmt.Errorf("failed to upsert runeword %s: %w", name, err)
		}
		stats.Imported++
	}

	return stats, nil
}

func (i *Importer) importRunes(ctx context.Context, catalogPath string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, "misc.txt"))
	if err != nil {
		return stats, err
	}

	// Also need gems.txt for rune socket mods
	gemsParser, err := parser.ParseFile(filepath.Join(catalogPath, "gems.txt"))
	if err != nil {
		return stats, err
	}

	// Build a map of rune codes to their mods from gems.txt
	runeMods := make(map[string]map[string][]Property)
	for _, row := range gemsParser.Rows() {
		r := parser.AsRow(row)
		name := r.GetString("name", "")
		code := r.GetString("code", "")

		if !strings.Contains(name, "Rune") || code == "" {
			continue
		}

		weaponMods := []Property{}
		helmMods := []Property{}
		shieldMods := []Property{}

		// Parse weapon mods
		for j := 1; j <= 3; j++ {
			modCode := r.GetString(fmt.Sprintf("weaponMod%dCode", j), "")
			if modCode != "" {
				prop := Property{
					Code:  modCode,
					Param: r.GetString(fmt.Sprintf("weaponMod%dParam", j), ""),
					Min:   r.GetInt(fmt.Sprintf("weaponMod%dMin", j), 0),
					Max:   r.GetInt(fmt.Sprintf("weaponMod%dMax", j), 0),
				}
				i.translator.EnrichProperty(&prop)
				weaponMods = append(weaponMods, prop)
			}
		}

		// Parse helm mods
		for j := 1; j <= 3; j++ {
			modCode := r.GetString(fmt.Sprintf("helmMod%dCode", j), "")
			if modCode != "" {
				prop := Property{
					Code:  modCode,
					Param: r.GetString(fmt.Sprintf("helmMod%dParam", j), ""),
					Min:   r.GetInt(fmt.Sprintf("helmMod%dMin", j), 0),
					Max:   r.GetInt(fmt.Sprintf("helmMod%dMax", j), 0),
				}
				i.translator.EnrichProperty(&prop)
				helmMods = append(helmMods, prop)
			}
		}

		// Parse shield mods
		for j := 1; j <= 3; j++ {
			modCode := r.GetString(fmt.Sprintf("shieldMod%dCode", j), "")
			if modCode != "" {
				prop := Property{
					Code:  modCode,
					Param: r.GetString(fmt.Sprintf("shieldMod%dParam", j), ""),
					Min:   r.GetInt(fmt.Sprintf("shieldMod%dMin", j), 0),
					Max:   r.GetInt(fmt.Sprintf("shieldMod%dMax", j), 0),
				}
				i.translator.EnrichProperty(&prop)
				shieldMods = append(shieldMods, prop)
			}
		}

		runeMods[code] = map[string][]Property{
			"weapon": weaponMods,
			"helm":   helmMods,
			"shield": shieldMods,
		}
	}

	// Now parse runes from misc.txt
	runeNumber := 0
	for _, row := range p.Rows() {
		r := parser.AsRow(row)
		name := r.GetString("name", "")
		code := r.GetString("code", "")
		itemType := r.GetString("type", "")

		// Only process runes
		if itemType != "rune" || !strings.Contains(name, "Rune") {
			continue
		}

		runeNumber++

		rune := &Rune{
			Code:       code,
			Name:       name,
			RuneNumber: runeNumber,
			Level:      r.GetInt("level", 0),
			LevelReq:   r.GetInt("levelreq", 0),
			InvFile:    r.GetString("invfile", ""),
			Cost:       r.GetInt("cost", 0),
		}

		// Add mods from gems.txt
		if mods, ok := runeMods[code]; ok {
			rune.WeaponMods = mods["weapon"]
			rune.HelmMods = mods["helm"]
			rune.ShieldMods = mods["shield"]
		}

		if err := i.repo.UpsertRune(ctx, rune); err != nil {
			return stats, fmt.Errorf("failed to upsert rune %s: %w", name, err)
		}
		stats.Imported++
	}

	return stats, nil
}

func (i *Importer) importGems(ctx context.Context, catalogPath string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, "gems.txt"))
	if err != nil {
		return stats, err
	}

	for _, row := range p.Rows() {
		r := parser.AsRow(row)
		name := r.GetString("name", "")
		code := r.GetString("code", "")

		// Skip runes (they're handled separately) and empty entries
		if name == "" || code == "" || strings.Contains(name, "Rune") {
			stats.Skipped++
			continue
		}

		// Determine gem type and quality from name
		gemType, quality := parseGemNameParts(name)

		// Parse weapon mods
		weaponMods := []Property{}
		for j := 1; j <= 3; j++ {
			modCode := r.GetString(fmt.Sprintf("weaponMod%dCode", j), "")
			if modCode != "" {
				prop := Property{
					Code:  modCode,
					Param: r.GetString(fmt.Sprintf("weaponMod%dParam", j), ""),
					Min:   r.GetInt(fmt.Sprintf("weaponMod%dMin", j), 0),
					Max:   r.GetInt(fmt.Sprintf("weaponMod%dMax", j), 0),
				}
				i.translator.EnrichProperty(&prop)
				weaponMods = append(weaponMods, prop)
			}
		}

		// Parse helm mods
		helmMods := []Property{}
		for j := 1; j <= 3; j++ {
			modCode := r.GetString(fmt.Sprintf("helmMod%dCode", j), "")
			if modCode != "" {
				prop := Property{
					Code:  modCode,
					Param: r.GetString(fmt.Sprintf("helmMod%dParam", j), ""),
					Min:   r.GetInt(fmt.Sprintf("helmMod%dMin", j), 0),
					Max:   r.GetInt(fmt.Sprintf("helmMod%dMax", j), 0),
				}
				i.translator.EnrichProperty(&prop)
				helmMods = append(helmMods, prop)
			}
		}

		// Parse shield mods
		shieldMods := []Property{}
		for j := 1; j <= 3; j++ {
			modCode := r.GetString(fmt.Sprintf("shieldMod%dCode", j), "")
			if modCode != "" {
				prop := Property{
					Code:  modCode,
					Param: r.GetString(fmt.Sprintf("shieldMod%dParam", j), ""),
					Min:   r.GetInt(fmt.Sprintf("shieldMod%dMin", j), 0),
					Max:   r.GetInt(fmt.Sprintf("shieldMod%dMax", j), 0),
				}
				i.translator.EnrichProperty(&prop)
				shieldMods = append(shieldMods, prop)
			}
		}

		gem := &Gem{
			Code:       code,
			Name:       name,
			GemType:    gemType,
			Quality:    quality,
			WeaponMods: weaponMods,
			HelmMods:   helmMods,
			ShieldMods: shieldMods,
			Transform:  r.GetInt("transform", 0),
		}

		if err := i.repo.UpsertGem(ctx, gem); err != nil {
			return stats, fmt.Errorf("failed to upsert gem %s: %w", name, err)
		}
		stats.Imported++
	}

	return stats, nil
}

func (i *Importer) importProperties(ctx context.Context, catalogPath string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, "properties.txt"))
	if err != nil {
		return stats, err
	}

	for _, row := range p.Rows() {
		r := parser.AsRow(row)
		code := r.GetString("code", "")
		if code == "" {
			stats.Skipped++
			continue
		}

		prop := &ItemProperty{
			Code:    code,
			Enabled: r.GetBool("*Enabled"),
			Stat1:   r.GetString("stat1", ""),
			Stat2:   r.GetString("stat2", ""),
			Stat3:   r.GetString("stat3", ""),
			Stat4:   r.GetString("stat4", ""),
			Stat5:   r.GetString("stat5", ""),
			Stat6:   r.GetString("stat6", ""),
			Stat7:   r.GetString("stat7", ""),
			Tooltip: r.GetString("*Tooltip", ""),
		}

		// Parse func values (only if non-zero)
		if f := r.GetInt("func1", 0); f != 0 {
			prop.Func1 = &f
		}
		if f := r.GetInt("func2", 0); f != 0 {
			prop.Func2 = &f
		}
		if f := r.GetInt("func3", 0); f != 0 {
			prop.Func3 = &f
		}
		if f := r.GetInt("func4", 0); f != 0 {
			prop.Func4 = &f
		}
		if f := r.GetInt("func5", 0); f != 0 {
			prop.Func5 = &f
		}
		if f := r.GetInt("func6", 0); f != 0 {
			prop.Func6 = &f
		}
		if f := r.GetInt("func7", 0); f != 0 {
			prop.Func7 = &f
		}

		if err := i.repo.UpsertProperty(ctx, prop); err != nil {
			return stats, fmt.Errorf("failed to upsert property %s: %w", code, err)
		}
		stats.Imported++
	}

	return stats, nil
}

func (i *Importer) importAffixes(ctx context.Context, catalogPath, filename, affixType string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, filename))
	if err != nil {
		return stats, err
	}

	for _, row := range p.Rows() {
		r := parser.AsRow(row)
		name := r.GetString("Name", "")
		if name == "" {
			stats.Skipped++
			continue
		}

		// Parse valid item types
		validTypes := []string{}
		for j := 1; j <= 7; j++ {
			itype := r.GetString(fmt.Sprintf("itype%d", j), "")
			if itype != "" {
				validTypes = append(validTypes, itype)
			}
		}

		// Parse excluded item types
		excludedTypes := []string{}
		for j := 1; j <= 5; j++ {
			etype := r.GetString(fmt.Sprintf("etype%d", j), "")
			if etype != "" {
				excludedTypes = append(excludedTypes, etype)
			}
		}

		var maxLevel *int
		if ml := r.GetInt("maxlevel", 0); ml > 0 {
			maxLevel = &ml
		}

		affix := &Affix{
			Name:              name,
			AffixType:         affixType,
			Version:           r.GetInt("version", 0),
			Spawnable:         r.GetBool("spawnable"),
			Rare:              r.GetBool("rare"),
			Level:             r.GetInt("level", 0),
			MaxLevel:          maxLevel,
			LevelReq:          r.GetInt("levelreq", 0),
			ClassSpecific:     r.GetString("classspecific", ""),
			ClassLevelReq:     r.GetInt("classlevelreq", 0),
			Frequency:         r.GetInt("frequency", 0),
			AffixGroup:        r.GetInt("group", 0),
			Mod1Code:          r.GetString("mod1code", ""),
			Mod1Param:         r.GetString("mod1param", ""),
			Mod1Min:           r.GetInt("mod1min", 0),
			Mod1Max:           r.GetInt("mod1max", 0),
			Mod2Code:          r.GetString("mod2code", ""),
			Mod2Param:         r.GetString("mod2param", ""),
			Mod2Min:           r.GetInt("mod2min", 0),
			Mod2Max:           r.GetInt("mod2max", 0),
			Mod3Code:          r.GetString("mod3code", ""),
			Mod3Param:         r.GetString("mod3param", ""),
			Mod3Min:           r.GetInt("mod3min", 0),
			Mod3Max:           r.GetInt("mod3max", 0),
			ValidItemTypes:    validTypes,
			ExcludedItemTypes: excludedTypes,
			TransformColor:    r.GetString("transformcolor", ""),
			Multiply:          r.GetInt("multiply", 0),
			AddCost:           r.GetInt("add", 0),
		}

		if err := i.repo.UpsertAffix(ctx, affix); err != nil {
			return stats, fmt.Errorf("failed to upsert affix %s: %w", name, err)
		}
		stats.Imported++
	}

	return stats, nil
}

// Helper functions

func parseProperties(r parser.Row, maxProps int, translator *PropertyTranslator) []Property {
	properties := []Property{}
	for j := 1; j <= maxProps; j++ {
		code := r.GetString(fmt.Sprintf("prop%d", j), "")
		if code == "" {
			continue
		}
		prop := Property{
			Code:  code,
			Param: r.GetString(fmt.Sprintf("par%d", j), ""),
			Min:   r.GetInt(fmt.Sprintf("min%d", j), 0),
			Max:   r.GetInt(fmt.Sprintf("max%d", j), 0),
		}
		translator.EnrichProperty(&prop)
		properties = append(properties, prop)
	}
	return properties
}

func parseRunewordProperties(r parser.Row, translator *PropertyTranslator) []Property {
	properties := []Property{}
	for j := 1; j <= 7; j++ {
		code := r.GetString(fmt.Sprintf("T1Code%d", j), "")
		if code == "" {
			continue
		}
		prop := Property{
			Code:  code,
			Param: r.GetString(fmt.Sprintf("T1Param%d", j), ""),
			Min:   r.GetInt(fmt.Sprintf("T1Min%d", j), 0),
			Max:   r.GetInt(fmt.Sprintf("T1Max%d", j), 0),
		}
		translator.EnrichProperty(&prop)
		properties = append(properties, prop)
	}
	return properties
}

func parseSetPartialBonuses(r parser.Row, translator *PropertyTranslator) []Property {
	properties := []Property{}
	// PCode2a-PCode5b (partial set bonuses)
	for j := 2; j <= 5; j++ {
		for _, suffix := range []string{"a", "b"} {
			code := r.GetString(fmt.Sprintf("PCode%d%s", j, suffix), "")
			if code == "" {
				continue
			}
			prop := Property{
				Code:  code,
				Param: r.GetString(fmt.Sprintf("PParam%d%s", j, suffix), ""),
				Min:   r.GetInt(fmt.Sprintf("PMin%d%s", j, suffix), 0),
				Max:   r.GetInt(fmt.Sprintf("PMax%d%s", j, suffix), 0),
			}
			translator.EnrichProperty(&prop)
			properties = append(properties, prop)
		}
	}
	return properties
}

func parseSetFullBonuses(r parser.Row, translator *PropertyTranslator) []Property {
	properties := []Property{}
	// FCode1-FCode8 (full set bonuses)
	for j := 1; j <= 8; j++ {
		code := r.GetString(fmt.Sprintf("FCode%d", j), "")
		if code == "" {
			continue
		}
		prop := Property{
			Code:  code,
			Param: r.GetString(fmt.Sprintf("FParam%d", j), ""),
			Min:   r.GetInt(fmt.Sprintf("FMin%d", j), 0),
			Max:   r.GetInt(fmt.Sprintf("FMax%d", j), 0),
		}
		translator.EnrichProperty(&prop)
		properties = append(properties, prop)
	}
	return properties
}

func parseSetItemProperties(r parser.Row, translator *PropertyTranslator) []Property {
	properties := []Property{}
	// prop1-prop9 (base properties)
	for j := 1; j <= 9; j++ {
		code := r.GetString(fmt.Sprintf("prop%d", j), "")
		if code == "" {
			continue
		}
		prop := Property{
			Code:  code,
			Param: r.GetString(fmt.Sprintf("par%d", j), ""),
			Min:   r.GetInt(fmt.Sprintf("min%d", j), 0),
			Max:   r.GetInt(fmt.Sprintf("max%d", j), 0),
		}
		translator.EnrichProperty(&prop)
		properties = append(properties, prop)
	}
	return properties
}

func parseSetItemBonusProperties(r parser.Row, translator *PropertyTranslator) []Property {
	properties := []Property{}
	// aprop1a-aprop5b (bonus properties when wearing more set items)
	for j := 1; j <= 5; j++ {
		for _, suffix := range []string{"a", "b"} {
			code := r.GetString(fmt.Sprintf("aprop%d%s", j, suffix), "")
			if code == "" {
				continue
			}
			prop := Property{
				Code:  code,
				Param: r.GetString(fmt.Sprintf("apar%d%s", j, suffix), ""),
				Min:   r.GetInt(fmt.Sprintf("amin%d%s", j, suffix), 0),
				Max:   r.GetInt(fmt.Sprintf("amax%d%s", j, suffix), 0),
			}
			translator.EnrichProperty(&prop)
			properties = append(properties, prop)
		}
	}
	return properties
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

func parseGemNameParts(name string) (gemType, quality string) {
	name = strings.ToLower(name)

	// Determine quality
	switch {
	case strings.HasPrefix(name, "chipped"):
		quality = "chipped"
	case strings.HasPrefix(name, "flawed"):
		quality = "flawed"
	case strings.HasPrefix(name, "flawless"):
		quality = "flawless"
	case strings.HasPrefix(name, "perfect"):
		quality = "perfect"
	default:
		quality = "normal"
	}

	// Determine gem type
	switch {
	case strings.Contains(name, "amethyst"):
		gemType = "amethyst"
	case strings.Contains(name, "sapphire"):
		gemType = "sapphire"
	case strings.Contains(name, "emerald"):
		gemType = "emerald"
	case strings.Contains(name, "ruby"):
		gemType = "ruby"
	case strings.Contains(name, "diamond"):
		gemType = "diamond"
	case strings.Contains(name, "topaz"):
		gemType = "topaz"
	case strings.Contains(name, "skull"):
		gemType = "skull"
	default:
		gemType = "unknown"
	}

	return
}

// buildRunewordBases pre-computes valid base items for each runeword
func (i *Importer) buildRunewordBases(ctx context.Context) (ImportStats, error) {
	stats := ImportStats{}

	// Clear existing mappings
	if err := i.repo.ClearRunewordBases(ctx); err != nil {
		return stats, fmt.Errorf("failed to clear runeword bases: %w", err)
	}

	// Build item type hierarchy map
	itemTypes, err := i.repo.GetAllItemTypesWithEquiv(ctx)
	if err != nil {
		return stats, fmt.Errorf("failed to get item types: %w", err)
	}

	// Build a map of type code -> all parent types (including self)
	typeHierarchy := buildTypeHierarchy(itemTypes)

	// Get all runewords with their requirements
	runewords, err := i.repo.GetAllRunewordsForMatching(ctx)
	if err != nil {
		return stats, fmt.Errorf("failed to get runewords: %w", err)
	}

	// Get all base items with socket info
	itemBases, err := i.repo.GetAllItemBasesForRunewordMatching(ctx)
	if err != nil {
		return stats, fmt.Errorf("failed to get item bases: %w", err)
	}

	// For each runeword, find matching bases
	for _, rw := range runewords {
		if len(rw.ValidItemTypes) == 0 {
			continue
		}

		// Build set of valid types (expanded with hierarchy)
		validTypesExpanded := make(map[string]bool)
		for _, vt := range rw.ValidItemTypes {
			validTypesExpanded[vt] = true
		}

		// Build set of excluded types (expanded with hierarchy)
		excludedTypesExpanded := make(map[string]bool)
		for _, et := range rw.ExcludedItemTypes {
			excludedTypesExpanded[et] = true
		}

		// Find matching bases
		for _, ib := range itemBases {
			// Check socket requirement
			if ib.MaxSockets < rw.RuneCount {
				continue
			}

			// Check if item type matches any valid type (through hierarchy)
			matches := false
			for _, typeCode := range []string{ib.ItemType, ib.ItemType2} {
				if typeCode == "" {
					continue
				}
				// Check if this type or any of its parents match valid types
				if parents, ok := typeHierarchy[typeCode]; ok {
					for _, parent := range parents {
						if validTypesExpanded[parent] {
							matches = true
							break
						}
					}
				}
				if matches {
					break
				}
			}

			if !matches {
				continue
			}

			// Check if item type is excluded
			excluded := false
			for _, typeCode := range []string{ib.ItemType, ib.ItemType2} {
				if typeCode == "" {
					continue
				}
				if parents, ok := typeHierarchy[typeCode]; ok {
					for _, parent := range parents {
						if excludedTypesExpanded[parent] {
							excluded = true
							break
						}
					}
				}
				if excluded {
					break
				}
			}

			if excluded {
				continue
			}

			// Insert the mapping
			rb := &RunewordBase{
				RunewordID:      rw.ID,
				ItemBaseID:      ib.ID,
				ItemBaseCode:    ib.Code,
				ItemBaseName:    ib.Name,
				Category:        ib.Category,
				MaxSockets:      ib.MaxSockets,
				RequiredSockets: rw.RuneCount,
			}

			if err := i.repo.InsertRunewordBase(ctx, rb); err != nil {
				return stats, fmt.Errorf("failed to insert runeword base for %s: %w", rw.Name, err)
			}
			stats.Imported++
		}
	}

	return stats, nil
}

// buildTypeHierarchy builds a map of type code -> all ancestor types (including self)
func buildTypeHierarchy(itemTypes []ItemTypeWithEquiv) map[string][]string {
	// First, build direct parent map
	directParents := make(map[string][]string)
	for _, it := range itemTypes {
		parents := []string{}
		if it.Equiv1 != "" {
			parents = append(parents, it.Equiv1)
		}
		if it.Equiv2 != "" {
			parents = append(parents, it.Equiv2)
		}
		directParents[it.Code] = parents
	}

	// Now recursively expand to get all ancestors
	hierarchy := make(map[string][]string)
	for _, it := range itemTypes {
		ancestors := getAllAncestors(it.Code, directParents, make(map[string]bool))
		hierarchy[it.Code] = ancestors
	}

	return hierarchy
}

// getAllAncestors recursively gets all ancestors of a type (including self)
func getAllAncestors(code string, directParents map[string][]string, visited map[string]bool) []string {
	if visited[code] {
		return nil
	}
	visited[code] = true

	result := []string{code}

	if parents, ok := directParents[code]; ok {
		for _, parent := range parents {
			ancestors := getAllAncestors(parent, directParents, visited)
			result = append(result, ancestors...)
		}
	}

	return result
}

// importTreasureClasses imports treasure class data from treasureclassex.txt
func (i *Importer) importTreasureClasses(ctx context.Context, catalogPath string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, "treasureclassex.txt"))
	if err != nil {
		return stats, err
	}

	rows := p.Rows()

	// First pass: collect all TC names to identify references vs item codes
	tcNames := make(map[string]bool)
	for _, row := range rows {
		r := parser.AsRow(row)
		name := r.GetString("Treasure Class", "")
		if name != "" {
			tcNames[name] = true
		}
	}

	// Second pass: import TCs with their items
	for _, row := range rows {
		r := parser.AsRow(row)
		name := r.GetString("Treasure Class", "")
		if name == "" {
			stats.Skipped++
			continue
		}

		// Parse group ID
		var groupID *int
		if g := r.GetInt("group", 0); g != 0 {
			groupID = &g
		}

		// Parse ladder seasons
		var firstSeason, lastSeason *int
		if fs := r.GetInt("firstLadderSeason", 0); fs > 0 {
			firstSeason = &fs
		}
		if ls := r.GetInt("lastLadderSeason", 0); ls > 0 {
			lastSeason = &ls
		}

		tc := &TreasureClass{
			Name:              name,
			GroupID:           groupID,
			Level:             r.GetInt("level", 0),
			Picks:             r.GetInt("Picks", 1),
			UniqueMod:         r.GetInt("Unique", 0),
			SetMod:            r.GetInt("Set", 0),
			RareMod:           r.GetInt("Rare", 0),
			MagicMod:          r.GetInt("Magic", 0),
			NoDrop:            r.GetInt("NoDrop", 0),
			FirstLadderSeason: firstSeason,
			LastLadderSeason:  lastSeason,
			NoAlwaysSpawn:     r.GetBool("noAlwaysSpawn"),
		}

		tcID, err := i.repo.UpsertTreasureClass(ctx, tc)
		if err != nil {
			return stats, fmt.Errorf("failed to upsert treasure class %s: %w", name, err)
		}

		// Import items (Item1-10 with Prob1-10)
		for slot := 1; slot <= 10; slot++ {
			itemCode := r.GetString(fmt.Sprintf("Item%d", slot), "")
			if itemCode == "" {
				continue
			}

			prob := r.GetInt(fmt.Sprintf("Prob%d", slot), 0)
			if prob == 0 {
				continue
			}

			// Check if this is a reference to another TC
			isTreasureClass := tcNames[itemCode]

			tci := &TreasureClassItem{
				TreasureClassID: tcID,
				Slot:            slot,
				ItemCode:        itemCode,
				IsTreasureClass: isTreasureClass,
				Probability:     prob,
			}

			if err := i.repo.UpsertTreasureClassItem(ctx, tci); err != nil {
				return stats, fmt.Errorf("failed to upsert treasure class item %s slot %d: %w", name, slot, err)
			}
		}

		stats.Imported++
	}

	return stats, nil
}

// importItemRatios imports item ratio data from itemratio.txt
func (i *Importer) importItemRatios(ctx context.Context, catalogPath string) (ImportStats, error) {
	stats := ImportStats{}

	p, err := parser.ParseFile(filepath.Join(catalogPath, "itemratio.txt"))
	if err != nil {
		return stats, err
	}

	for _, row := range p.Rows() {
		r := parser.AsRow(row)
		funcName := r.GetString("Function", "")
		if funcName == "" {
			stats.Skipped++
			continue
		}

		ir := &ItemRatio{
			FunctionName:     funcName,
			Version:          r.GetInt("Version", 0),
			IsUber:           r.GetInt("Uber", 0) == 1,
			IsClassSpecific:  r.GetInt("Class Specific", 0) == 1,
			UniqueRatio:      r.GetInt("Unique", 0),
			UniqueDivisor:    r.GetInt("UniqueDivisor", 1),
			UniqueMin:        r.GetInt("UniqueMin", 0),
			RareRatio:        r.GetInt("Rare", 0),
			RareDivisor:      r.GetInt("RareDivisor", 1),
			RareMin:          r.GetInt("RareMin", 0),
			SetRatio:         r.GetInt("Set", 0),
			SetDivisor:       r.GetInt("SetDivisor", 1),
			SetMin:           r.GetInt("SetMin", 0),
			MagicRatio:       r.GetInt("Magic", 0),
			MagicDivisor:     r.GetInt("MagicDivisor", 1),
			MagicMin:         r.GetInt("MagicMin", 0),
			HiQualityRatio:   r.GetInt("HiQuality", 0),
			HiQualityDivisor: r.GetInt("HiQualityDivisor", 1),
			NormalRatio:      r.GetInt("Normal", 0),
			NormalDivisor:    r.GetInt("NormalDivisor", 1),
		}

		if err := i.repo.UpsertItemRatio(ctx, ir); err != nil {
			return stats, fmt.Errorf("failed to upsert item ratio %s: %w", funcName, err)
		}
		stats.Imported++
	}

	return stats, nil
}
