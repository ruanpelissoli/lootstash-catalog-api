package handlers

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/api/dto"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
)

// ItemHandler handles item-related API requests
type ItemHandler struct {
	repo       *d2.Repository
	translator *d2.PropertyTranslator
}

// slugifyParam lowercases and replaces spaces with hyphens for composite stat codes.
func slugifyParam(s string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", "-"))
}

// capitalize returns a string with the first letter uppercased
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// resolveItemTypeName looks up the item type by code and returns the parent
// type name if available, falling back to the type's own name. This ensures
// sub-types like "mcha" (Medium Charm) resolve to "Charm" instead of "Mcha".
func (h *ItemHandler) resolveItemTypeName(code string) string {
	if code == "" {
		return ""
	}
	it, err := h.repo.GetItemType(context.Background(), code)
	if err != nil {
		return capitalize(code)
	}
	return capitalize(it.Name)
}

// NewItemHandler creates a new item handler
func NewItemHandler(repo *d2.Repository) *ItemHandler {
	return &ItemHandler{
		repo:       repo,
		translator: d2.DefaultTranslator,
	}
}

// Search handles item search requests
// GET /api/d2/items/search?q=<query>&limit=<limit>
func (h *ItemHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Query parameter 'q' is required",
			Code:    400,
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	results, err := h.repo.SearchItems(c.Context(), query, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to search items",
			Code:    500,
		})
	}

	// Convert to DTOs
	items := make([]dto.ItemSearchResult, 0, len(results))
	for _, r := range results {
		category := capitalize(r.Category)
		baseName := capitalize(r.BaseName)
		// Omit baseName when it duplicates the category (e.g. jewel/Jewel, ring/Ring)
		if strings.EqualFold(baseName, category) {
			baseName = ""
		}
		items = append(items, dto.ItemSearchResult{
			ID:       strconv.Itoa(r.ID),
			Name:     r.Name,
			Type:     capitalize(r.Type),
			Category: category,
			ImageURL: r.ImageURL,
			BaseName: baseName,
		})
	}

	// Get total count
	totalCount, _ := h.repo.CountSearchResults(c.Context(), query)

	return c.JSON(dto.SearchResponse{
		Items:      items,
		TotalCount: totalCount,
		Query:      query,
	})
}

// GetUniqueItem handles unique item detail requests
// GET /api/d2/items/unique/:id
func (h *ItemHandler) GetUniqueItem(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid item ID",
			Code:    400,
		})
	}

	item, err := h.repo.GetUniqueItem(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Unique item not found",
			Code:    404,
		})
	}

	// Get base item info
	base, _ := h.repo.GetItemBaseByCode(c.Context(), item.BaseCode)

	detail := h.convertUniqueToDTO(item, base)

	return c.JSON(dto.UnifiedItemDetail{
		ItemType: "unique",
		Unique:   detail,
	})
}

// GetSetItem handles set item detail requests
// GET /api/d2/items/set/:id
func (h *ItemHandler) GetSetItem(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid item ID",
			Code:    400,
		})
	}

	item, err := h.repo.GetSetItem(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Set item not found",
			Code:    404,
		})
	}

	// Get base item info
	base, _ := h.repo.GetItemBaseByCode(c.Context(), item.BaseCode)

	detail := h.convertSetItemToDTO(item, base)

	return c.JSON(dto.UnifiedItemDetail{
		ItemType: "set",
		SetItem:  detail,
	})
}

// GetRuneword handles runeword detail requests
// GET /api/d2/items/runeword/:id
func (h *ItemHandler) GetRuneword(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid runeword ID",
			Code:    400,
		})
	}

	item, err := h.repo.GetRuneword(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Runeword not found",
			Code:    404,
		})
	}

	// Get valid base items for this runeword
	bases, _ := h.repo.GetBasesForRuneword(c.Context(), id)

	// Get rune info for display
	runeInfoMap, _ := h.repo.GetRunesByCodes(c.Context(), item.Runes)

	// Get item type names for display
	typeInfoMap, _ := h.repo.GetItemTypesByCodes(c.Context(), item.ValidItemTypes)

	detail := h.convertRunewordToDTO(item, bases, runeInfoMap, typeInfoMap)

	return c.JSON(dto.UnifiedItemDetail{
		ItemType: "runeword",
		Runeword: detail,
	})
}

// GetRunewordBases returns valid base items for a runeword
// GET /api/d2/items/runeword/:id/bases
func (h *ItemHandler) GetRunewordBases(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid runeword ID",
			Code:    400,
		})
	}

	bases, err := h.repo.GetBasesForRuneword(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get runeword bases",
			Code:    500,
		})
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

	return c.JSON(results)
}

// GetRune handles rune detail requests
// GET /api/d2/items/rune/:id
func (h *ItemHandler) GetRune(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid rune ID",
			Code:    400,
		})
	}

	item, err := h.repo.GetRune(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Rune not found",
			Code:    404,
		})
	}

	detail := h.convertRuneToDTO(item)

	return c.JSON(dto.UnifiedItemDetail{
		ItemType: "rune",
		Rune:     detail,
	})
}

// GetGem handles gem detail requests
// GET /api/d2/items/gem/:id
func (h *ItemHandler) GetGem(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid gem ID",
			Code:    400,
		})
	}

	item, err := h.repo.GetGem(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Gem not found",
			Code:    404,
		})
	}

	detail := h.convertGemToDTO(item)

	return c.JSON(dto.UnifiedItemDetail{
		ItemType: "gem",
		Gem:      detail,
	})
}

// GetBase handles base item detail requests
// GET /api/d2/items/base/:id
func (h *ItemHandler) GetBase(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid base item ID",
			Code:    400,
		})
	}

	item, err := h.repo.GetItemBase(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Base item not found",
			Code:    404,
		})
	}

	// Get item type info
	itemType, _ := h.repo.GetItemType(c.Context(), item.ItemType)

	detail := h.convertBaseToDTO(item, itemType)

	return c.JSON(dto.UnifiedItemDetail{
		ItemType: "base",
		Base:     detail,
	})
}

// GetItem handles generic item detail requests by type and ID
// GET /api/d2/items/:type/:id
func (h *ItemHandler) GetItem(c *fiber.Ctx) error {
	itemType := strings.ToLower(c.Params("type"))
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid item ID",
			Code:    400,
		})
	}

	switch itemType {
	case "unique":
		item, err := h.repo.GetUniqueItem(c.Context(), id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Item not found",
				Code:    404,
			})
		}
		base, _ := h.repo.GetItemBaseByCode(c.Context(), item.BaseCode)
		return c.JSON(dto.UnifiedItemDetail{
			ItemType: "unique",
			Unique:   h.convertUniqueToDTO(item, base),
		})

	case "set":
		item, err := h.repo.GetSetItem(c.Context(), id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Item not found",
				Code:    404,
			})
		}
		base, _ := h.repo.GetItemBaseByCode(c.Context(), item.BaseCode)
		return c.JSON(dto.UnifiedItemDetail{
			ItemType: "set",
			SetItem:  h.convertSetItemToDTO(item, base),
		})

	case "runeword":
		item, err := h.repo.GetRuneword(c.Context(), id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Item not found",
				Code:    404,
			})
		}
		bases, _ := h.repo.GetBasesForRuneword(c.Context(), id)
		runeInfoMap, _ := h.repo.GetRunesByCodes(c.Context(), item.Runes)
		typeInfoMap, _ := h.repo.GetItemTypesByCodes(c.Context(), item.ValidItemTypes)
		return c.JSON(dto.UnifiedItemDetail{
			ItemType: "runeword",
			Runeword: h.convertRunewordToDTO(item, bases, runeInfoMap, typeInfoMap),
		})

	case "rune":
		item, err := h.repo.GetRune(c.Context(), id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Item not found",
				Code:    404,
			})
		}
		return c.JSON(dto.UnifiedItemDetail{
			ItemType: "rune",
			Rune:     h.convertRuneToDTO(item),
		})

	case "gem":
		item, err := h.repo.GetGem(c.Context(), id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Item not found",
				Code:    404,
			})
		}
		return c.JSON(dto.UnifiedItemDetail{
			ItemType: "gem",
			Gem:      h.convertGemToDTO(item),
		})

	case "base":
		item, err := h.repo.GetItemBase(c.Context(), id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Item not found",
				Code:    404,
			})
		}
		itemTypeInfo, _ := h.repo.GetItemType(c.Context(), item.ItemType)
		return c.JSON(dto.UnifiedItemDetail{
			ItemType: "base",
			Base:     h.convertBaseToDTO(item, itemTypeInfo),
		})

	case "quest":
		item, err := h.repo.GetItemBase(c.Context(), id)
		if err != nil || !item.QuestItem {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Quest item not found",
				Code:    404,
			})
		}
		return c.JSON(dto.UnifiedItemDetail{
			ItemType: "quest",
			Quest:    h.convertQuestToDTO(item),
		})

	default:
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid item type. Must be one of: unique, set, runeword, rune, gem, base, quest",
			Code:    400,
		})
	}
}

// GetAllRunes returns all runes ordered by rune number
// GET /api/d2/runes
func (h *ItemHandler) GetAllRunes(c *fiber.Ctx) error {
	runes, err := h.repo.GetAllRunes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get runes",
			Code:    500,
		})
	}

	results := make([]*dto.RuneDetail, 0, len(runes))
	for _, r := range runes {
		results = append(results, h.convertRuneToDTO(&r))
	}

	return c.JSON(results)
}

// GetAllGems returns all gems ordered by quality and type
// GET /api/d2/gems
func (h *ItemHandler) GetAllGems(c *fiber.Ctx) error {
	gems, err := h.repo.GetAllGems(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get gems",
			Code:    500,
		})
	}

	results := make([]*dto.GemDetail, 0, len(gems))
	for _, g := range gems {
		results = append(results, h.convertGemToDTO(&g))
	}

	return c.JSON(results)
}

// GetAllBases returns all base items, optionally filtered by category or runeword
// GET /api/d2/bases?category=armor|weapon|misc&runeword=5
func (h *ItemHandler) GetAllBases(c *fiber.Ctx) error {
	category := c.Query("category")
	runewordIDStr := c.Query("runeword")

	// Validate category if provided
	if category != "" && category != "armor" && category != "weapon" && category != "misc" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid category. Must be one of: armor, weapon, misc",
			Code:    400,
		})
	}

	// If runeword filter is provided, return bases for that runeword
	if runewordIDStr != "" {
		runewordID, err := strconv.Atoi(runewordIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "bad_request",
				Message: "Invalid runeword ID",
				Code:    400,
			})
		}

		runewordBases, err := h.repo.GetBasesForRuneword(c.Context(), runewordID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to get base items for runeword",
				Code:    500,
			})
		}

		results := make([]*dto.BaseItemDetail, 0, len(runewordBases))
		for _, rb := range runewordBases {
			// Apply category filter if provided
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
		return c.JSON(results)
	}

	bases, err := h.repo.GetAllItemBases(c.Context(), category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get base items",
			Code:    500,
		})
	}

	results := make([]*dto.BaseItemDetail, 0, len(bases))
	for _, b := range bases {
		itemType, _ := h.repo.GetItemType(c.Context(), b.ItemType)
		results = append(results, h.convertBaseToDTO(&b, itemType))
	}

	return c.JSON(results)
}

// GetAllUniques returns all unique items
// GET /api/d2/uniques
func (h *ItemHandler) GetAllUniques(c *fiber.Ctx) error {
	items, err := h.repo.GetAllUniqueItems(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get unique items",
			Code:    500,
		})
	}

	results := make([]*dto.UniqueItemDetail, 0, len(items))
	for _, item := range items {
		base, _ := h.repo.GetItemBaseByCode(c.Context(), item.BaseCode)
		results = append(results, h.convertUniqueToDTO(&item, base))
	}

	return c.JSON(results)
}

// GetAllSets returns all set items
// GET /api/d2/sets
func (h *ItemHandler) GetAllSets(c *fiber.Ctx) error {
	items, err := h.repo.GetAllSetItems(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get set items",
			Code:    500,
		})
	}

	results := make([]*dto.SetItemDetail, 0, len(items))
	for _, item := range items {
		base, _ := h.repo.GetItemBaseByCode(c.Context(), item.BaseCode)
		results = append(results, h.convertSetItemToDTO(&item, base))
	}

	return c.JSON(results)
}

// GetAllRunewords returns all runewords
// GET /api/d2/runewords
func (h *ItemHandler) GetAllRunewords(c *fiber.Ctx) error {
	items, err := h.repo.GetAllRunewordsForList(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get runewords",
			Code:    500,
		})
	}

	// Collect all rune codes and type codes for batch lookup
	allRuneCodes := make([]string, 0)
	allTypeCodes := make([]string, 0)
	for _, item := range items {
		allRuneCodes = append(allRuneCodes, item.Runes...)
		allTypeCodes = append(allTypeCodes, item.ValidItemTypes...)
	}

	// Batch fetch rune and type info
	runeInfoMap, _ := h.repo.GetRunesByCodes(c.Context(), allRuneCodes)
	typeInfoMap, _ := h.repo.GetItemTypesByCodes(c.Context(), allTypeCodes)

	results := make([]*dto.RunewordDetail, 0, len(items))
	for _, item := range items {
		// Don't fetch bases for list view - use detail endpoint for full info
		results = append(results, h.convertRunewordToDTO(&item, nil, runeInfoMap, typeInfoMap))
	}

	return c.JSON(results)
}

// Helper methods for DTO conversion

func (h *ItemHandler) convertUniqueToDTO(item *d2.UniqueItem, base *d2.ItemBase) *dto.UniqueItemDetail {
	detail := &dto.UniqueItemDetail{
		ID:     item.ID,
		Name:   item.Name,
		Type:   "Unique",
		Rarity: "Unique",
		Requirements: dto.ItemRequirements{
			Level: item.LevelReq,
		},
		LadderOnly: item.LadderOnly,
		ImageURL:   item.ImageURL,
	}

	// Add base info if available
	if base != nil {
		detail.Base = dto.ItemBaseInfo{
			Code:     base.Code,
			Name:     base.Name,
			Category: capitalize(base.Category),
			ItemType: h.resolveItemTypeName(base.ItemType),
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

	// Convert properties to affixes
	detail.Affixes = h.convertPropertiesToAffixes(item.Properties)

	return detail
}

func (h *ItemHandler) convertSetItemToDTO(item *d2.SetItem, base *d2.ItemBase) *dto.SetItemDetail {
	detail := &dto.SetItemDetail{
		ID:      item.ID,
		Name:    item.Name,
		SetName: item.SetName,
		Type:    "Set",
		Rarity:  "Set",
		Requirements: dto.ItemRequirements{
			Level: item.LevelReq,
		},
		ImageURL: item.ImageURL,
	}

	// Add base info if available
	if base != nil {
		detail.Base = dto.ItemBaseInfo{
			Code:     base.Code,
			Name:     base.Name,
			Category: capitalize(base.Category),
			ItemType: h.resolveItemTypeName(base.ItemType),
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

	// Convert properties
	detail.Affixes = h.convertPropertiesToAffixes(item.Properties)
	detail.BonusAffixes = h.convertPropertiesToAffixes(item.BonusProperties)

	return detail
}

func (h *ItemHandler) convertRunewordToDTO(item *d2.Runeword, bases []d2.RunewordBase, runeInfoMap map[string]d2.RuneInfo, typeInfoMap map[string]d2.ItemTypeInfo) *dto.RunewordDetail {
	detail := &dto.RunewordDetail{
		ID:          item.ID,
		Name:        item.Name,
		DisplayName: item.DisplayName,
		Type:        "Runeword",
		Rarity:      "Runeword",
		LadderOnly:  item.LadderOnly,
		ImageURL:    item.ImageURL,
	}

	// Build runes with display info
	detail.Runes = make([]dto.RunewordRune, 0, len(item.Runes))
	for _, runeCode := range item.Runes {
		rune := dto.RunewordRune{Code: runeCode}
		if info, ok := runeInfoMap[runeCode]; ok {
			rune.ID = info.ID
			// Use short name (strip " Rune" suffix)
			shortName := strings.TrimSuffix(info.Name, " Rune")
			rune.Name = shortName
			rune.ImageURL = info.ImageURL
			detail.RuneOrder += shortName
		} else {
			detail.RuneOrder += runeCode
		}
		detail.Runes = append(detail.Runes, rune)
	}

	// Build valid types with names
	detail.ValidTypes = make([]dto.RunewordValidType, 0, len(item.ValidItemTypes))
	for _, typeCode := range item.ValidItemTypes {
		vt := dto.RunewordValidType{Code: typeCode}
		if info, ok := typeInfoMap[typeCode]; ok {
			vt.Name = info.Name
		} else {
			vt.Name = typeCode // fallback to code
		}
		detail.ValidTypes = append(detail.ValidTypes, vt)
	}

	// Convert properties
	detail.Affixes = h.convertPropertiesToAffixes(item.Properties)

	// Add valid base items
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

func (h *ItemHandler) convertRuneToDTO(item *d2.Rune) *dto.RuneDetail {
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

	// Convert mods
	detail.WeaponMods = h.convertPropertiesToAffixes(item.WeaponMods)
	detail.ArmorMods = h.convertPropertiesToAffixes(item.HelmMods)
	detail.ShieldMods = h.convertPropertiesToAffixes(item.ShieldMods)

	return detail
}

func (h *ItemHandler) convertGemToDTO(item *d2.Gem) *dto.GemDetail {
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

	// Convert mods
	detail.WeaponMods = h.convertPropertiesToAffixes(item.WeaponMods)
	detail.ArmorMods = h.convertPropertiesToAffixes(item.HelmMods)
	detail.ShieldMods = h.convertPropertiesToAffixes(item.ShieldMods)

	return detail
}

func (h *ItemHandler) convertBaseToDTO(item *d2.ItemBase, itemType *d2.ItemType) *dto.BaseItemDetail {
	detail := &dto.BaseItemDetail{
		ID:       item.ID,
		Code:     item.Code,
		Name:     item.Name,
		Type:     "Base",
		Rarity:   "Normal",
		Category: capitalize(item.Category),
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

	// Set item type name from lookup
	if itemType != nil {
		detail.ItemType = h.resolveItemTypeName(itemType.Code)
	} else {
		detail.ItemType = h.resolveItemTypeName(item.ItemType)
	}

	// Defense for armor
	if item.MinAC > 0 || item.MaxAC > 0 {
		detail.Defense = &dto.DefenseRange{
			Min: item.MinAC,
			Max: item.MaxAC,
		}
	}

	// Damage for weapons
	if item.MinDam > 0 || item.MaxDam > 0 || item.TwoHandMinDam > 0 || item.TwoHandMaxDam > 0 {
		detail.Damage = &dto.DamageRange{
			OneHandMin: item.MinDam,
			OneHandMax: item.MaxDam,
			TwoHandMin: item.TwoHandMinDam,
			TwoHandMax: item.TwoHandMaxDam,
		}
	}

	// Quality tiers
	if item.NormalCode != "" || item.ExceptionalCode != "" || item.EliteCode != "" {
		detail.QualityTiers = dto.QualityTiers{
			Normal:      item.NormalCode,
			Exceptional: item.ExceptionalCode,
			Elite:       item.EliteCode,
		}
	}

	return detail
}

func (h *ItemHandler) convertPropertiesToAffixes(props []d2.Property) []dto.ItemAffix {
	affixes := make([]dto.ItemAffix, 0, len(props))
	for _, prop := range props {
		name := prop.DisplayText
		hasRange := prop.HasRange

		// Fallback for old data without pre-computed values (remove after re-import)
		if name == "" {
			name = h.translator.Translate(prop)
			hasRange = prop.Min != prop.Max
		}

		// Generate composite code for parametric stats (e.g. "charged-hydra")
		code := prop.Code
		if prop.Param != "" {
			code = prop.Code + "-" + slugifyParam(prop.Param)
		}

		affix := dto.ItemAffix{
			Name:        name,
			DisplayName: h.translator.GetDisplayName(prop.Code),
			Code:        code,
			HasRange:    hasRange,
		}

		// Handle special affixes with selectable options
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

// GetAllStats returns all filterable stat codes for marketplace filtering
// GET /api/d2/stats
func (h *ItemHandler) GetAllStats(c *fiber.Ctx) error {
	stats, err := h.repo.GetAllStats(c.Context())
	if err != nil {
		// Fallback to hardcoded stats if DB query fails
		hardcoded := d2.FilterableStats()
		results := make([]dto.StatCode, 0, len(hardcoded))
		for _, s := range hardcoded {
			results = append(results, dto.StatCode{
				Code:        s.Code,
				Name:        s.Name,
				Description: s.Description,
				Category:    s.Category,
				Aliases:     s.Aliases,
				IsVariable:  s.IsVariable,
			})
		}
		return c.JSON(results)
	}

	results := make([]dto.StatCode, 0, len(stats))
	for _, s := range stats {
		results = append(results, dto.StatCode{
			Code:        s.Code,
			Name:        s.Name,
			Description: s.DisplayText,
			Category:    s.Category,
			Aliases:     s.Aliases,
			IsVariable:  s.IsVariable,
		})
	}

	return c.JSON(results)
}

// GetAllCategories returns all item categories for marketplace filtering
// GET /api/d2/categories
func (h *ItemHandler) GetAllCategories(c *fiber.Ctx) error {
	categories := d2.Categories()

	results := make([]dto.Category, 0, len(categories))
	for _, cat := range categories {
		results = append(results, dto.Category{
			Code:        cat.Code,
			Name:        cat.Name,
			Description: cat.Description,
		})
	}

	return c.JSON(results)
}

// GetAllRarities returns all item rarities for marketplace filtering
// GET /api/d2/rarities
func (h *ItemHandler) GetAllRarities(c *fiber.Ctx) error {
	rarities := d2.Rarities()

	results := make([]dto.Rarity, 0, len(rarities))
	for _, r := range rarities {
		results = append(results, dto.Rarity{
			Code:        r.Code,
			Name:        r.Name,
			Color:       r.Color,
			Description: r.Description,
		})
	}

	return c.JSON(results)
}

// GetQuestItem handles quest item detail requests
// GET /api/d2/items/quest/:id
func (h *ItemHandler) GetQuestItem(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid item ID",
			Code:    400,
		})
	}

	item, err := h.repo.GetItemBase(c.Context(), id)
	if err != nil || !item.QuestItem {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Quest item not found",
			Code:    404,
		})
	}

	return c.JSON(dto.UnifiedItemDetail{
		ItemType: "quest",
		Quest:    h.convertQuestToDTO(item),
	})
}

// GetAllQuestItems returns all quest items
// GET /api/d2/quests
func (h *ItemHandler) GetAllQuestItems(c *fiber.Ctx) error {
	items, err := h.repo.GetAllQuestItems(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get quest items",
			Code:    500,
		})
	}

	results := make([]*dto.QuestItemDetail, 0, len(items))
	for _, item := range items {
		results = append(results, h.convertQuestToDTO(&item))
	}

	return c.JSON(results)
}

// GetAllClasses returns all character classes
// GET /api/d2/classes
func (h *ItemHandler) GetAllClasses(c *fiber.Ctx) error {
	classes, err := h.repo.GetAllClasses(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get classes",
			Code:    500,
		})
	}

	results := make([]dto.ClassDetail, 0, len(classes))
	for _, cls := range classes {
		results = append(results, convertClassToDTO(&cls))
	}

	return c.JSON(results)
}

func (h *ItemHandler) convertQuestToDTO(item *d2.ItemBase) *dto.QuestItemDetail {
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

func convertClassToDTO(cls *d2.Class) dto.ClassDetail {
	trees := make([]dto.SkillTreeDTO, 0, len(cls.SkillTrees))
	for _, st := range cls.SkillTrees {
		trees = append(trees, dto.SkillTreeDTO{
			Name:   st.Name,
			Skills: st.Skills,
		})
	}
	return dto.ClassDetail{
		ID:          cls.ID,
		Name:        cls.Name,
		SkillSuffix: cls.SkillSuffix,
		SkillTrees:  trees,
	}
}
