package services

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/api/dto"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/cache"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
)

const (
	listTTL   = 24 * time.Hour
	detailTTL = 6 * time.Hour
	searchTTL = 1 * time.Hour
)

// CatalogService provides cached access to catalog data.
// When cache is nil, all methods fall through to the database.
type CatalogService struct {
	repo       *d2.Repository
	cache      *cache.RedisCache
	translator *d2.PropertyTranslator
}

func NewCatalogService(repo *d2.Repository, c *cache.RedisCache) *CatalogService {
	return &CatalogService{
		repo:       repo,
		cache:      c,
		translator: d2.DefaultTranslator,
	}
}

// Repo exposes the repository for use by admin handlers.
func (s *CatalogService) Repo() *d2.Repository {
	return s.repo
}

// getOrCache is a generic cache-aside helper.
// On any cache error (miss, connection failure, etc.) it falls through to fetch.
// Cache write errors are silently ignored.
func getOrCache[T any](ctx context.Context, c *cache.RedisCache, key string, ttl time.Duration, fetch func() (T, error)) (T, error) {
	var zero T
	if c != nil {
		var result T
		if err := c.Get(ctx, key, &result); err == nil {
			return result, nil
		}
	}
	result, err := fetch()
	if err != nil {
		return zero, err
	}
	if c != nil {
		_ = c.SetWithTTL(ctx, key, result, ttl)
	}
	return result, nil
}

// --- Search ---

func (s *CatalogService) Search(ctx context.Context, query string, limit int) (dto.SearchResponse, error) {
	return getOrCache(ctx, s.cache, cache.D2SearchKey(query, limit), searchTTL, func() (dto.SearchResponse, error) {
		results, err := s.repo.SearchItems(ctx, query, limit)
		if err != nil {
			return dto.SearchResponse{}, err
		}

		items := make([]dto.ItemSearchResult, 0, len(results))
		for _, r := range results {
			category := capitalize(r.Category)
			baseName := capitalize(r.BaseName)
			if strings.EqualFold(baseName, category) {
				baseName = ""
			}
			items = append(items, dto.ItemSearchResult{
				ID:          strconv.Itoa(r.ID),
				Name:        r.Name,
				Type:        capitalize(r.Type),
				Category:    category,
				Subcategory: r.Subcategory,
				ImageURL:    r.ImageURL,
				BaseName:    baseName,
			})
		}

		totalCount, _ := s.repo.CountSearchResults(ctx, query)

		return dto.SearchResponse{
			Items:      items,
			TotalCount: totalCount,
			Query:      query,
		}, nil
	})
}

// --- Detail endpoints ---

func (s *CatalogService) GetUniqueItem(ctx context.Context, id int) (dto.UnifiedItemDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2UniqueItemKey(id), detailTTL, func() (dto.UnifiedItemDetail, error) {
		item, err := s.repo.GetUniqueItem(ctx, id)
		if err != nil {
			return dto.UnifiedItemDetail{}, err
		}
		base, _ := s.repo.GetItemBaseByCode(ctx, item.BaseCode)
		return dto.UnifiedItemDetail{
			ItemType: "unique",
			Unique:   s.convertUniqueToDTO(item, base),
		}, nil
	})
}

func (s *CatalogService) GetSetItem(ctx context.Context, id int) (dto.UnifiedItemDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2SetItemKey(id), detailTTL, func() (dto.UnifiedItemDetail, error) {
		item, err := s.repo.GetSetItem(ctx, id)
		if err != nil {
			return dto.UnifiedItemDetail{}, err
		}
		base, _ := s.repo.GetItemBaseByCode(ctx, item.BaseCode)
		return dto.UnifiedItemDetail{
			ItemType: "set",
			SetItem:  s.convertSetItemToDTO(item, base),
		}, nil
	})
}

func (s *CatalogService) GetRuneword(ctx context.Context, id int) (dto.UnifiedItemDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2RunewordKey(strconv.Itoa(id)), detailTTL, func() (dto.UnifiedItemDetail, error) {
		item, err := s.repo.GetRuneword(ctx, id)
		if err != nil {
			return dto.UnifiedItemDetail{}, err
		}
		bases, _ := s.repo.GetBasesForRuneword(ctx, id)
		runeInfoMap, _ := s.repo.GetRunesByCodes(ctx, item.Runes)
		typeInfoMap, _ := s.repo.GetItemTypesByCodes(ctx, item.ValidItemTypes)
		return dto.UnifiedItemDetail{
			ItemType: "runeword",
			Runeword: s.convertRunewordToDTO(item, bases, runeInfoMap, typeInfoMap),
		}, nil
	})
}

func (s *CatalogService) GetRunewordBases(ctx context.Context, id int) ([]dto.RunewordBaseItem, error) {
	key := cache.D2RunewordKey(strconv.Itoa(id)) + ":bases"
	return getOrCache(ctx, s.cache, key, detailTTL, func() ([]dto.RunewordBaseItem, error) {
		bases, err := s.repo.GetBasesForRuneword(ctx, id)
		if err != nil {
			return nil, err
		}
		results := make([]dto.RunewordBaseItem, 0, len(bases))
		for _, b := range bases {
			results = append(results, dto.RunewordBaseItem{
				ID:         b.ItemBaseID,
				Code:       b.ItemBaseCode,
				Name:       b.ItemBaseName,
				Category:   capitalize(b.Category),
				MaxSockets: b.MaxSockets,
			})
		}
		return results, nil
	})
}

func (s *CatalogService) GetRune(ctx context.Context, id int) (dto.UnifiedItemDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2RuneKey(strconv.Itoa(id)), detailTTL, func() (dto.UnifiedItemDetail, error) {
		item, err := s.repo.GetRune(ctx, id)
		if err != nil {
			return dto.UnifiedItemDetail{}, err
		}
		return dto.UnifiedItemDetail{
			ItemType: "rune",
			Rune:     s.convertRuneToDTO(item),
		}, nil
	})
}

func (s *CatalogService) GetGem(ctx context.Context, id int) (dto.UnifiedItemDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2GemKey(strconv.Itoa(id)), detailTTL, func() (dto.UnifiedItemDetail, error) {
		item, err := s.repo.GetGem(ctx, id)
		if err != nil {
			return dto.UnifiedItemDetail{}, err
		}
		return dto.UnifiedItemDetail{
			ItemType: "gem",
			Gem:      s.convertGemToDTO(item),
		}, nil
	})
}

func (s *CatalogService) GetBase(ctx context.Context, id int) (dto.UnifiedItemDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2ItemBaseKey(strconv.Itoa(id)), detailTTL, func() (dto.UnifiedItemDetail, error) {
		item, err := s.repo.GetItemBase(ctx, id)
		if err != nil {
			return dto.UnifiedItemDetail{}, err
		}
		itemType, _ := s.repo.GetItemType(ctx, item.ItemType)
		return dto.UnifiedItemDetail{
			ItemType: "base",
			Base:     s.convertBaseToDTO(item, itemType),
		}, nil
	})
}

func (s *CatalogService) GetQuestItem(ctx context.Context, id int) (dto.UnifiedItemDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2QuestItemKey(id), detailTTL, func() (dto.UnifiedItemDetail, error) {
		item, err := s.repo.GetItemBase(ctx, id)
		if err != nil || !item.QuestItem {
			return dto.UnifiedItemDetail{}, err
		}
		return dto.UnifiedItemDetail{
			ItemType: "quest",
			Quest:    s.convertQuestToDTO(item),
		}, nil
	})
}

// GetItem dispatches to the correct detail method by item type.
func (s *CatalogService) GetItem(ctx context.Context, itemType string, id int) (dto.UnifiedItemDetail, error) {
	switch itemType {
	case "unique":
		return s.GetUniqueItem(ctx, id)
	case "set":
		return s.GetSetItem(ctx, id)
	case "runeword":
		return s.GetRuneword(ctx, id)
	case "rune":
		return s.GetRune(ctx, id)
	case "gem":
		return s.GetGem(ctx, id)
	case "base":
		return s.GetBase(ctx, id)
	case "quest":
		return s.GetQuestItem(ctx, id)
	default:
		return dto.UnifiedItemDetail{}, errInvalidType
	}
}

var errInvalidType = &invalidTypeError{}

type invalidTypeError struct{}

func (e *invalidTypeError) Error() string { return "invalid item type" }

// IsInvalidTypeError checks if an error is an invalid type error.
func IsInvalidTypeError(err error) bool {
	_, ok := err.(*invalidTypeError)
	return ok
}

// --- List endpoints ---

func (s *CatalogService) GetAllRunes(ctx context.Context) ([]*dto.RuneDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2RunesKey(), listTTL, func() ([]*dto.RuneDetail, error) {
		runes, err := s.repo.GetAllRunes(ctx)
		if err != nil {
			return nil, err
		}
		results := make([]*dto.RuneDetail, 0, len(runes))
		for _, r := range runes {
			results = append(results, s.convertRuneToDTO(&r))
		}
		return results, nil
	})
}

func (s *CatalogService) GetAllGems(ctx context.Context) ([]*dto.GemDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2GemsKey(), listTTL, func() ([]*dto.GemDetail, error) {
		gems, err := s.repo.GetAllGems(ctx)
		if err != nil {
			return nil, err
		}
		results := make([]*dto.GemDetail, 0, len(gems))
		for _, g := range gems {
			results = append(results, s.convertGemToDTO(&g))
		}
		return results, nil
	})
}

func (s *CatalogService) GetAllBases(ctx context.Context, category string, runewordID int) ([]*dto.BaseItemDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2BasesFilteredKey(category, runewordID), listTTL, func() ([]*dto.BaseItemDetail, error) {
		if runewordID > 0 {
			runewordBases, err := s.repo.GetBasesForRuneword(ctx, runewordID)
			if err != nil {
				return nil, err
			}
			results := make([]*dto.BaseItemDetail, 0, len(runewordBases))
			for _, rb := range runewordBases {
				if category != "" && rb.Category != category {
					continue
				}
				results = append(results, &dto.BaseItemDetail{
					ID:         rb.ItemBaseID,
					Code:       rb.ItemBaseCode,
					Name:       rb.ItemBaseName,
					Type:       "Base",
					Rarity:     "Normal",
					Category:   capitalize(rb.Category),
					MaxSockets: rb.MaxSockets,
				})
			}
			return results, nil
		}

		bases, err := s.repo.GetAllItemBases(ctx, category)
		if err != nil {
			return nil, err
		}
		results := make([]*dto.BaseItemDetail, 0, len(bases))
		for _, b := range bases {
			itemType, _ := s.repo.GetItemType(ctx, b.ItemType)
			results = append(results, s.convertBaseToDTO(&b, itemType))
		}
		return results, nil
	})
}

func (s *CatalogService) GetAllUniques(ctx context.Context) ([]*dto.UniqueItemDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2UniqueItemsKey(), listTTL, func() ([]*dto.UniqueItemDetail, error) {
		items, err := s.repo.GetAllUniqueItems(ctx)
		if err != nil {
			return nil, err
		}
		results := make([]*dto.UniqueItemDetail, 0, len(items))
		for _, item := range items {
			base, _ := s.repo.GetItemBaseByCode(ctx, item.BaseCode)
			results = append(results, s.convertUniqueToDTO(&item, base))
		}
		return results, nil
	})
}

func (s *CatalogService) GetAllSets(ctx context.Context) ([]*dto.SetItemDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2SetItemsKey(), listTTL, func() ([]*dto.SetItemDetail, error) {
		items, err := s.repo.GetAllSetItems(ctx)
		if err != nil {
			return nil, err
		}
		results := make([]*dto.SetItemDetail, 0, len(items))
		for _, item := range items {
			base, _ := s.repo.GetItemBaseByCode(ctx, item.BaseCode)
			results = append(results, s.convertSetItemToDTO(&item, base))
		}
		return results, nil
	})
}

func (s *CatalogService) GetAllRunewords(ctx context.Context) ([]*dto.RunewordDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2RunewordsKey(), listTTL, func() ([]*dto.RunewordDetail, error) {
		items, err := s.repo.GetAllRunewordsForList(ctx)
		if err != nil {
			return nil, err
		}

		allRuneCodes := make([]string, 0)
		allTypeCodes := make([]string, 0)
		for _, item := range items {
			allRuneCodes = append(allRuneCodes, item.Runes...)
			allTypeCodes = append(allTypeCodes, item.ValidItemTypes...)
		}

		runeInfoMap, _ := s.repo.GetRunesByCodes(ctx, allRuneCodes)
		typeInfoMap, _ := s.repo.GetItemTypesByCodes(ctx, allTypeCodes)

		results := make([]*dto.RunewordDetail, 0, len(items))
		for _, item := range items {
			results = append(results, s.convertRunewordToDTO(&item, nil, runeInfoMap, typeInfoMap))
		}
		return results, nil
	})
}

func (s *CatalogService) GetAllQuestItems(ctx context.Context) ([]*dto.QuestItemDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2QuestItemsKey(), listTTL, func() ([]*dto.QuestItemDetail, error) {
		items, err := s.repo.GetAllQuestItems(ctx)
		if err != nil {
			return nil, err
		}
		results := make([]*dto.QuestItemDetail, 0, len(items))
		for _, item := range items {
			results = append(results, s.convertQuestToDTO(&item))
		}
		return results, nil
	})
}

func (s *CatalogService) GetAllClasses(ctx context.Context) ([]dto.ClassDetail, error) {
	return getOrCache(ctx, s.cache, cache.D2ClassesKey(), listTTL, func() ([]dto.ClassDetail, error) {
		classes, err := s.repo.GetAllClasses(ctx)
		if err != nil {
			return nil, err
		}
		results := make([]dto.ClassDetail, 0, len(classes))
		for _, cls := range classes {
			trees := make([]dto.SkillTreeDTO, 0, len(cls.SkillTrees))
			for _, st := range cls.SkillTrees {
				trees = append(trees, dto.SkillTreeDTO{
					Name:   st.Name,
					Skills: st.Skills,
				})
			}
			results = append(results, dto.ClassDetail{
				ID:          cls.ID,
				Name:        cls.Name,
				SkillSuffix: cls.SkillSuffix,
				SkillTrees:  trees,
			})
		}
		return results, nil
	})
}

func (s *CatalogService) GetAllStats(ctx context.Context) ([]dto.StatCode, error) {
	return getOrCache(ctx, s.cache, cache.D2StatsKey(), listTTL, func() ([]dto.StatCode, error) {
		stats, err := s.repo.GetAllStats(ctx)
		if err != nil {
			// Fallback to hardcoded stats if DB query fails
			hardcoded := d2.FilterableStats()
			results := make([]dto.StatCode, 0, len(hardcoded))
			for _, st := range hardcoded {
				results = append(results, dto.StatCode{
					Code:        st.Code,
					Name:        st.Name,
					DisplayName: st.Description,
					Category:    st.Category,
					Aliases:     st.Aliases,
					IsVariable:  st.IsVariable,
				})
			}
			return results, nil
		}

		results := make([]dto.StatCode, 0, len(stats))
		for _, st := range stats {
			results = append(results, dto.StatCode{
				Code:        st.Code,
				Name:        st.Name,
				DisplayName: st.DisplayText,
				Category:    st.Category,
				Aliases:     st.Aliases,
				IsVariable:  st.IsVariable,
			})
		}
		return results, nil
	})
}

// --- DTO conversion methods ---

func slugifyParam(s string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", "-"))
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (s *CatalogService) resolveItemTypeName(code string) string {
	if code == "" {
		return ""
	}
	it, err := s.repo.GetItemType(context.Background(), code)
	if err != nil {
		return capitalize(code)
	}
	return capitalize(it.Name)
}

func (s *CatalogService) convertUniqueToDTO(item *d2.UniqueItem, base *d2.ItemBase) *dto.UniqueItemDetail {
	detail := &dto.UniqueItemDetail{
		ID:          item.ID,
		Name:        item.Name,
		Type:        "Unique",
		Rarity:      "Unique",
		Category:    capitalize(item.Category),
		Subcategory: item.Subcategory,
		Requirements: dto.ItemRequirements{
			Level: item.LevelReq,
		},
		LadderOnly: item.LadderOnly,
		ImageURL:   item.ImageURL,
	}

	if base != nil {
		detail.Base = dto.ItemBaseInfo{
			Code:        base.Code,
			Name:        base.Name,
			Category:    capitalize(base.Category),
			Subcategory: base.Subcategory,
			ItemType:    s.resolveItemTypeName(base.ItemType),
		}
		if base.MaxAC > 0 {
			detail.Base.Defense = &dto.DefenseRange{
				Min: base.MinAC,
				Max: base.MaxAC,
			}
		}
		if base.MaxDam > 0 {
			detail.Base.MinDamage = &base.MinDam
			detail.Base.MaxDamage = &base.MaxDam
		}
		detail.Base.MaxSockets = base.MaxSockets
		detail.Base.Durability = base.Durability
		detail.Requirements.Strength = base.StrReq
		detail.Requirements.Dexterity = base.DexReq
	} else if item.BaseName != "" {
		detail.Base = dto.ItemBaseInfo{
			Name: item.BaseName,
		}
	}

	detail.Affixes = s.convertPropertiesToAffixes(item.Properties)
	return detail
}

func (s *CatalogService) convertSetItemToDTO(item *d2.SetItem, base *d2.ItemBase) *dto.SetItemDetail {
	detail := &dto.SetItemDetail{
		ID:          item.ID,
		Name:        item.Name,
		SetName:     item.SetName,
		Type:        "Set",
		Rarity:      "Set",
		Category:    capitalize(item.Category),
		Subcategory: item.Subcategory,
		Requirements: dto.ItemRequirements{
			Level: item.LevelReq,
		},
		ImageURL: item.ImageURL,
	}

	if base != nil {
		detail.Base = dto.ItemBaseInfo{
			Code:        base.Code,
			Name:        base.Name,
			Category:    capitalize(base.Category),
			Subcategory: base.Subcategory,
			ItemType:    s.resolveItemTypeName(base.ItemType),
		}
		if base.MaxAC > 0 {
			detail.Base.Defense = &dto.DefenseRange{
				Min: base.MinAC,
				Max: base.MaxAC,
			}
		}
		if base.MaxDam > 0 {
			detail.Base.MinDamage = &base.MinDam
			detail.Base.MaxDamage = &base.MaxDam
		}
		detail.Base.MaxSockets = base.MaxSockets
		detail.Base.Durability = base.Durability
		detail.Requirements.Strength = base.StrReq
		detail.Requirements.Dexterity = base.DexReq
	} else if item.BaseName != "" {
		detail.Base = dto.ItemBaseInfo{
			Name: item.BaseName,
		}
	}

	detail.Affixes = s.convertPropertiesToAffixes(item.Properties)
	detail.BonusAffixes = s.convertPropertiesToAffixes(item.BonusProperties)
	return detail
}

func (s *CatalogService) convertRunewordToDTO(item *d2.Runeword, bases []d2.RunewordBase, runeInfoMap map[string]d2.RuneInfo, typeInfoMap map[string]d2.ItemTypeInfo) *dto.RunewordDetail {
	detail := &dto.RunewordDetail{
		ID:          item.ID,
		Name:        item.Name,
		DisplayName: item.DisplayName,
		Type:        "Runeword",
		Rarity:      "Runeword",
		Category:    capitalize(item.Category),
		Subcategory: item.Subcategory,
		LadderOnly:  item.LadderOnly,
		ImageURL:    item.ImageURL,
	}

	detail.Runes = make([]dto.RunewordRune, 0, len(item.Runes))
	for _, runeCode := range item.Runes {
		r := dto.RunewordRune{Code: runeCode}
		if info, ok := runeInfoMap[runeCode]; ok {
			r.ID = info.ID
			shortName := strings.TrimSuffix(info.Name, " Rune")
			r.Name = shortName
			r.ImageURL = info.ImageURL
			detail.RuneOrder += shortName
		} else {
			detail.RuneOrder += runeCode
		}
		detail.Runes = append(detail.Runes, r)
	}

	detail.ValidTypes = make([]dto.RunewordValidType, 0, len(item.ValidItemTypes))
	for _, typeCode := range item.ValidItemTypes {
		vt := dto.RunewordValidType{Code: typeCode}
		if info, ok := typeInfoMap[typeCode]; ok {
			vt.Name = info.Name
		} else {
			vt.Name = typeCode
		}
		detail.ValidTypes = append(detail.ValidTypes, vt)
	}

	if len(item.PropertiesByType) > 0 {
		detail.AffixesByType = make(map[string][]dto.ItemAffix)
		for typeName, props := range item.PropertiesByType {
			detail.AffixesByType[typeName] = s.convertPropertiesToAffixes(props)
		}
		if len(item.ValidItemTypes) > 0 {
			if props, ok := item.PropertiesByType[item.ValidItemTypes[0]]; ok {
				detail.Affixes = s.convertPropertiesToAffixes(props)
			}
		}
	} else {
		detail.Affixes = s.convertPropertiesToAffixes(item.Properties)
	}

	if len(bases) > 0 {
		detail.ValidBaseItems = make([]dto.RunewordBaseItem, 0, len(bases))
		for _, b := range bases {
			detail.ValidBaseItems = append(detail.ValidBaseItems, dto.RunewordBaseItem{
				ID:         b.ItemBaseID,
				Code:       b.ItemBaseCode,
				Name:       b.ItemBaseName,
				Category:   capitalize(b.Category),
				MaxSockets: b.MaxSockets,
			})
		}
	}

	return detail
}

func (s *CatalogService) convertRuneToDTO(item *d2.Rune) *dto.RuneDetail {
	detail := &dto.RuneDetail{
		ID:         item.ID,
		Code:       item.Code,
		Name:       item.Name,
		RuneNumber: item.RuneNumber,
		Type:       "Rune",
		Rarity:     "Rune",
		Requirements: dto.ItemRequirements{
			Level: item.LevelReq,
		},
		ImageURL: item.ImageURL,
	}

	detail.WeaponMods = s.convertPropertiesToAffixes(item.WeaponMods)
	detail.ArmorMods = s.convertPropertiesToAffixes(item.HelmMods)
	detail.ShieldMods = s.convertPropertiesToAffixes(item.ShieldMods)
	return detail
}

func (s *CatalogService) convertGemToDTO(item *d2.Gem) *dto.GemDetail {
	detail := &dto.GemDetail{
		ID:       item.ID,
		Code:     item.Code,
		Name:     item.Name,
		GemType:  capitalize(item.GemType),
		Quality:  capitalize(item.Quality),
		Type:     "Gem",
		Rarity:   "Gem",
		ImageURL: item.ImageURL,
	}

	detail.WeaponMods = s.convertPropertiesToAffixes(item.WeaponMods)
	detail.ArmorMods = s.convertPropertiesToAffixes(item.HelmMods)
	detail.ShieldMods = s.convertPropertiesToAffixes(item.ShieldMods)
	return detail
}

func (s *CatalogService) convertBaseToDTO(item *d2.ItemBase, itemType *d2.ItemType) *dto.BaseItemDetail {
	detail := &dto.BaseItemDetail{
		ID:          item.ID,
		Code:        item.Code,
		Name:        item.Name,
		Type:        "Base",
		Rarity:      "Normal",
		Category:    capitalize(item.Category),
		Subcategory: item.Subcategory,
		Tier:          item.Tier,
		TypeTags:      item.TypeTags,
		ClassSpecific: item.ClassSpecific,
		Requirements: dto.ItemRequirements{
			Level:     item.LevelReq,
			Strength:  item.StrReq,
			Dexterity: item.DexReq,
		},
		MaxSockets: item.MaxSockets,
		Durability: item.Durability,
		Speed:      item.Speed,
		ImageURL:   item.ImageURL,
	}

	if len(item.IconVariants) > 0 {
		detail.IconVariants = item.IconVariants
	}

	if itemType != nil {
		detail.ItemType = s.resolveItemTypeName(itemType.Code)
	} else {
		detail.ItemType = s.resolveItemTypeName(item.ItemType)
	}

	if item.MinAC > 0 || item.MaxAC > 0 {
		detail.Defense = &dto.DefenseRange{
			Min: item.MinAC,
			Max: item.MaxAC,
		}
	}

	if item.MinDam > 0 || item.MaxDam > 0 || item.TwoHandMinDam > 0 || item.TwoHandMaxDam > 0 {
		detail.Damage = &dto.DamageRange{
			OneHandMin: item.MinDam,
			OneHandMax: item.MaxDam,
			TwoHandMin: item.TwoHandMinDam,
			TwoHandMax: item.TwoHandMaxDam,
		}
	}

	if item.NormalCode != "" || item.ExceptionalCode != "" || item.EliteCode != "" {
		detail.QualityTiers = dto.QualityTiers{
			Normal:      item.NormalCode,
			Exceptional: item.ExceptionalCode,
			Elite:       item.EliteCode,
		}
	}

	return detail
}

func (s *CatalogService) convertQuestToDTO(item *d2.ItemBase) *dto.QuestItemDetail {
	return &dto.QuestItemDetail{
		ID:          item.ID,
		Code:        item.Code,
		Name:        item.Name,
		Description: item.Description,
		Type:        "Quest",
		Rarity:      "Quest",
		ImageURL:    item.ImageURL,
	}
}

func (s *CatalogService) convertPropertiesToAffixes(props []d2.Property) []dto.ItemAffix {
	affixes := make([]dto.ItemAffix, 0, len(props))
	for _, prop := range props {
		if prop.Code == "or-group" && len(prop.Alternatives) > 0 {
			affix := dto.ItemAffix{
				Name: prop.DisplayText,
				Code: "or-group",
			}
			for _, alt := range prop.Alternatives {
				altName := alt.DisplayText
				if altName == "" {
					altName = s.translator.Translate(alt)
				}
				altCode := alt.Code
				if alt.Param != "" {
					altCode = alt.Code + "-" + slugifyParam(alt.Param)
				}
				opt := dto.AffixOption{
					Value: altCode,
					Label: altName,
				}
				if alt.HasRange || alt.Min != alt.Max {
					min, max := alt.Min, alt.Max
					opt.MinValue = &min
					opt.MaxValue = &max
					opt.HasRange = true
				}
				affix.Options = append(affix.Options, opt)
			}
			affixes = append(affixes, affix)
			continue
		}

		name := prop.DisplayText
		hasRange := prop.HasRange

		if name == "" {
			name = s.translator.Translate(prop)
			hasRange = prop.Min != prop.Max
		}

		code := prop.Code
		if prop.Param != "" {
			code = prop.Code + "-" + slugifyParam(prop.Param)
		}

		affix := dto.ItemAffix{
			Name:        name,
			DisplayName: s.translator.GetDisplayName(prop.Code),
			Code:        code,
			HasRange:    hasRange,
		}

		if prop.Code == "randclassskill" {
			affix.Options = dto.D2Classes
		}

		if affix.HasRange {
			min := prop.Min
			max := prop.Max
			affix.MinValue = &min
			affix.MaxValue = &max
		}
		affixes = append(affixes, affix)
	}
	return affixes
}
