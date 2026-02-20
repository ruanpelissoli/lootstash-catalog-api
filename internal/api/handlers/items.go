package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/api/dto"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/api/services"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
)

// ItemHandler handles item-related API requests
type ItemHandler struct {
	service *services.CatalogService
}

// NewItemHandler creates a new item handler
func NewItemHandler(service *services.CatalogService) *ItemHandler {
	return &ItemHandler{service: service}
}

// capitalize returns a string with the first letter uppercased
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
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

	result, err := h.service.Search(c.Context(), query, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to search items",
			Code:    500,
		})
	}

	return c.JSON(result)
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

	result, err := h.service.GetUniqueItem(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Unique item not found",
			Code:    404,
		})
	}

	return c.JSON(result)
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

	result, err := h.service.GetSetItem(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Set item not found",
			Code:    404,
		})
	}

	return c.JSON(result)
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

	result, err := h.service.GetRuneword(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Runeword not found",
			Code:    404,
		})
	}

	return c.JSON(result)
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

	results, err := h.service.GetRunewordBases(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get runeword bases",
			Code:    500,
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

	result, err := h.service.GetRune(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Rune not found",
			Code:    404,
		})
	}

	return c.JSON(result)
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

	result, err := h.service.GetGem(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Gem not found",
			Code:    404,
		})
	}

	return c.JSON(result)
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

	result, err := h.service.GetBase(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Base item not found",
			Code:    404,
		})
	}

	return c.JSON(result)
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

	result, err := h.service.GetItem(c.Context(), itemType, id)
	if err != nil {
		if services.IsInvalidTypeError(err) {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "bad_request",
				Message: "Invalid item type. Must be one of: unique, set, runeword, rune, gem, base, quest",
				Code:    400,
			})
		}
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Item not found",
			Code:    404,
		})
	}

	return c.JSON(result)
}

// GetAllRunes returns all runes ordered by rune number
// GET /api/d2/runes
func (h *ItemHandler) GetAllRunes(c *fiber.Ctx) error {
	results, err := h.service.GetAllRunes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get runes",
			Code:    500,
		})
	}
	return c.JSON(results)
}

// GetAllGems returns all gems ordered by quality and type
// GET /api/d2/gems
func (h *ItemHandler) GetAllGems(c *fiber.Ctx) error {
	results, err := h.service.GetAllGems(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get gems",
			Code:    500,
		})
	}
	return c.JSON(results)
}

// GetAllBases returns all base items, optionally filtered by category or runeword
// GET /api/d2/bases?category=armor|weapon|misc&runeword=5
func (h *ItemHandler) GetAllBases(c *fiber.Ctx) error {
	category := c.Query("category")
	runewordIDStr := c.Query("runeword")

	validCategories := map[string]bool{
		"armor": true, "weapons": true, "jewelry": true,
		"charms": true, "runes": true, "gems": true, "misc": true,
	}
	if category != "" && !validCategories[category] {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid category. Must be one of: armor, weapons, jewelry, charms, runes, gems, misc",
			Code:    400,
		})
	}

	var runewordID int
	if runewordIDStr != "" {
		var err error
		runewordID, err = strconv.Atoi(runewordIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "bad_request",
				Message: "Invalid runeword ID",
				Code:    400,
			})
		}
	}

	results, err := h.service.GetAllBases(c.Context(), category, runewordID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get base items",
			Code:    500,
		})
	}

	return c.JSON(results)
}

// GetAllUniques returns all unique items
// GET /api/d2/uniques
func (h *ItemHandler) GetAllUniques(c *fiber.Ctx) error {
	results, err := h.service.GetAllUniques(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get unique items",
			Code:    500,
		})
	}
	return c.JSON(results)
}

// GetAllSets returns all set items
// GET /api/d2/sets
func (h *ItemHandler) GetAllSets(c *fiber.Ctx) error {
	results, err := h.service.GetAllSets(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get set items",
			Code:    500,
		})
	}
	return c.JSON(results)
}

// GetAllRunewords returns all runewords
// GET /api/d2/runewords
func (h *ItemHandler) GetAllRunewords(c *fiber.Ctx) error {
	results, err := h.service.GetAllRunewords(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get runewords",
			Code:    500,
		})
	}
	return c.JSON(results)
}

// GetAllStats returns all filterable stat codes for marketplace filtering
// GET /api/d2/stats
func (h *ItemHandler) GetAllStats(c *fiber.Ctx) error {
	results, err := h.service.GetAllStats(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get stats",
			Code:    500,
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
		subcats := make([]dto.SubcategoryDTO, 0, len(cat.Subcategories))
		for _, sc := range cat.Subcategories {
			subcats = append(subcats, dto.SubcategoryDTO{
				Code: sc.Code,
				Name: sc.Name,
			})
		}
		results = append(results, dto.Category{
			Code:          cat.Code,
			Name:          cat.Name,
			Description:   cat.Description,
			Subcategories: subcats,
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

	result, err := h.service.GetQuestItem(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Quest item not found",
			Code:    404,
		})
	}

	return c.JSON(result)
}

// GetAllQuestItems returns all quest items
// GET /api/d2/quests
func (h *ItemHandler) GetAllQuestItems(c *fiber.Ctx) error {
	results, err := h.service.GetAllQuestItems(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get quest items",
			Code:    500,
		})
	}
	return c.JSON(results)
}

// GetAllClasses returns all character classes
// GET /api/d2/classes
func (h *ItemHandler) GetAllClasses(c *fiber.Ctx) error {
	results, err := h.service.GetAllClasses(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get classes",
			Code:    500,
		})
	}
	return c.JSON(results)
}

// convertClassToDTO converts a class entity to a DTO (used by admin handler).
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
