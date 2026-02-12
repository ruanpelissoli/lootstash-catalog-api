package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/api/dto"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
)

// AdminHandler handles admin CRUD API requests
type AdminHandler struct {
	repo       *d2.Repository
	translator *d2.PropertyTranslator
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(repo *d2.Repository) *AdminHandler {
	return &AdminHandler{
		repo:       repo,
		translator: d2.DefaultTranslator,
	}
}

// CreateItem handles creating items of any type
// POST /admin/d2/items/:type
func (h *AdminHandler) CreateItem(c *fiber.Ctx) error {
	itemType := c.Params("type")

	switch itemType {
	case "unique":
		return h.createUniqueItem(c)
	case "set":
		return h.createSetItem(c)
	case "runeword":
		return h.createRuneword(c)
	case "rune":
		return h.createRune(c)
	case "gem":
		return h.createGem(c)
	case "base":
		return h.createBaseItem(c)
	case "quest":
		return h.createQuestItem(c)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid item type. Must be one of: unique, set, runeword, rune, gem, base, quest",
			Code:    400,
		})
	}
}

// UpdateItem handles updating items of any type
// PUT /admin/d2/items/:type/:id
func (h *AdminHandler) UpdateItem(c *fiber.Ctx) error {
	itemType := c.Params("type")
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
		return h.updateUniqueItem(c, id)
	case "set":
		return h.updateSetItem(c, id)
	case "runeword":
		return h.updateRuneword(c, id)
	case "rune":
		return h.updateRune(c, id)
	case "gem":
		return h.updateGem(c, id)
	case "base":
		return h.updateBaseItem(c, id)
	case "quest":
		return h.updateQuestItem(c, id)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid item type. Must be one of: unique, set, runeword, rune, gem, base, quest",
			Code:    400,
		})
	}
}

// DeleteItem handles deleting items (only quest items supported)
// DELETE /admin/d2/items/:type/:id
func (h *AdminHandler) DeleteItem(c *fiber.Ctx) error {
	itemType := c.Params("type")
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid item ID",
			Code:    400,
		})
	}

	if itemType != "quest" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Delete is only supported for quest items",
			Code:    400,
		})
	}

	if err := h.repo.DeleteQuestItem(c.Context(), id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Quest item not found",
			Code:    404,
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// convertInputProperties converts PropertyInput DTOs to d2.Property entities
func convertInputProperties(inputs []dto.PropertyInput) []d2.Property {
	props := make([]d2.Property, 0, len(inputs))
	for _, p := range inputs {
		props = append(props, d2.Property{
			Code:  p.Code,
			Param: p.Param,
			Min:   p.Min,
			Max:   p.Max,
		})
	}
	return props
}

// Unique item CRUD

func (h *AdminHandler) createUniqueItem(c *fiber.Ctx) error {
	var req dto.CreateUniqueItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	if req.Name == "" || req.BaseCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Name and baseCode are required",
			Code:    400,
		})
	}

	// Get next index ID
	maxIndex, _ := h.repo.GetMaxIndexID(c.Context(), "unique_items")

	item := &d2.UniqueItem{
		IndexID:    maxIndex + 1,
		Name:       req.Name,
		BaseCode:   req.BaseCode,
		LevelReq:   req.LevelReq,
		LadderOnly: req.LadderOnly,
		Properties: convertInputProperties(req.Properties),
		ImageURL:   req.ImageURL,
		Enabled:    true,
	}

	// Translate properties
	for i := range item.Properties {
		item.Properties[i].DisplayText = h.translator.Translate(item.Properties[i])
		item.Properties[i].HasRange = item.Properties[i].Min != item.Properties[i].Max
	}

	if err := h.repo.UpsertUniqueItem(c.Context(), item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create unique item",
			Code:    500,
		})
	}

	// Fetch the created item to return
	created, err := h.repo.GetUniqueItemByName(c.Context(), req.Name)
	if err != nil {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Unique item created"})
	}

	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *AdminHandler) updateUniqueItem(c *fiber.Ctx, id int) error {
	var req dto.CreateUniqueItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	props := convertInputProperties(req.Properties)
	for i := range props {
		props[i].DisplayText = h.translator.Translate(props[i])
		props[i].HasRange = props[i].Min != props[i].Max
	}

	item := &d2.UniqueItem{
		Name:       req.Name,
		BaseCode:   req.BaseCode,
		LevelReq:   req.LevelReq,
		LadderOnly: req.LadderOnly,
		Properties: props,
		ImageURL:   req.ImageURL,
	}

	if err := h.repo.UpdateUniqueItemFields(c.Context(), id, item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update unique item",
			Code:    500,
		})
	}

	updated, err := h.repo.GetUniqueItem(c.Context(), id)
	if err != nil {
		return c.JSON(fiber.Map{"message": "Unique item updated"})
	}

	return c.JSON(updated)
}

// Set item CRUD

func (h *AdminHandler) createSetItem(c *fiber.Ctx) error {
	var req dto.CreateSetItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	if req.Name == "" || req.SetName == "" || req.BaseCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Name, setName, and baseCode are required",
			Code:    400,
		})
	}

	maxIndex, _ := h.repo.GetMaxIndexID(c.Context(), "set_items")

	props := convertInputProperties(req.Properties)
	for i := range props {
		props[i].DisplayText = h.translator.Translate(props[i])
		props[i].HasRange = props[i].Min != props[i].Max
	}

	bonusProps := convertInputProperties(req.BonusProperties)
	for i := range bonusProps {
		bonusProps[i].DisplayText = h.translator.Translate(bonusProps[i])
		bonusProps[i].HasRange = bonusProps[i].Min != bonusProps[i].Max
	}

	item := &d2.SetItem{
		IndexID:         maxIndex + 1,
		Name:            req.Name,
		SetName:         req.SetName,
		BaseCode:        req.BaseCode,
		LevelReq:        req.LevelReq,
		Properties:      props,
		BonusProperties: bonusProps,
		ImageURL:        req.ImageURL,
	}

	if err := h.repo.UpsertSetItem(c.Context(), item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create set item",
			Code:    500,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Set item created"})
}

func (h *AdminHandler) updateSetItem(c *fiber.Ctx, id int) error {
	var req dto.CreateSetItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	props := convertInputProperties(req.Properties)
	for i := range props {
		props[i].DisplayText = h.translator.Translate(props[i])
		props[i].HasRange = props[i].Min != props[i].Max
	}

	bonusProps := convertInputProperties(req.BonusProperties)
	for i := range bonusProps {
		bonusProps[i].DisplayText = h.translator.Translate(bonusProps[i])
		bonusProps[i].HasRange = bonusProps[i].Min != bonusProps[i].Max
	}

	item := &d2.SetItem{
		Name:            req.Name,
		SetName:         req.SetName,
		BaseCode:        req.BaseCode,
		LevelReq:        req.LevelReq,
		Properties:      props,
		BonusProperties: bonusProps,
		ImageURL:        req.ImageURL,
	}

	if err := h.repo.UpdateSetItemFields(c.Context(), id, item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update set item",
			Code:    500,
		})
	}

	updated, err := h.repo.GetSetItem(c.Context(), id)
	if err != nil {
		return c.JSON(fiber.Map{"message": "Set item updated"})
	}

	return c.JSON(updated)
}

// Runeword CRUD

func (h *AdminHandler) createRuneword(c *fiber.Ctx) error {
	var req dto.CreateRunewordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	if req.Name == "" || req.DisplayName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Name and displayName are required",
			Code:    400,
		})
	}

	props := convertInputProperties(req.Properties)
	for i := range props {
		props[i].DisplayText = h.translator.Translate(props[i])
		props[i].HasRange = props[i].Min != props[i].Max
	}

	item := &d2.Runeword{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Complete:       true,
		LadderOnly:     req.LadderOnly,
		ValidItemTypes: req.ValidItemTypes,
		Runes:          req.Runes,
		Properties:     props,
		ImageURL:       req.ImageURL,
	}

	if err := h.repo.UpsertRuneword(c.Context(), item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create runeword",
			Code:    500,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Runeword created"})
}

func (h *AdminHandler) updateRuneword(c *fiber.Ctx, id int) error {
	var req dto.CreateRunewordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	props := convertInputProperties(req.Properties)
	for i := range props {
		props[i].DisplayText = h.translator.Translate(props[i])
		props[i].HasRange = props[i].Min != props[i].Max
	}

	item := &d2.Runeword{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		LadderOnly:     req.LadderOnly,
		ValidItemTypes: req.ValidItemTypes,
		Runes:          req.Runes,
		Properties:     props,
		ImageURL:       req.ImageURL,
	}

	if err := h.repo.UpdateRunewordFields(c.Context(), id, item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update runeword",
			Code:    500,
		})
	}

	updated, err := h.repo.GetRuneword(c.Context(), id)
	if err != nil {
		return c.JSON(fiber.Map{"message": "Runeword updated"})
	}

	return c.JSON(updated)
}

// Rune CRUD

func (h *AdminHandler) createRune(c *fiber.Ctx) error {
	var req dto.CreateRuneRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	if req.Code == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Code and name are required",
			Code:    400,
		})
	}

	item := &d2.Rune{
		Code:       req.Code,
		Name:       req.Name,
		RuneNumber: req.RuneNumber,
		LevelReq:   req.LevelReq,
		WeaponMods: convertInputProperties(req.WeaponMods),
		HelmMods:   convertInputProperties(req.ArmorMods),
		ShieldMods: convertInputProperties(req.ShieldMods),
		ImageURL:   req.ImageURL,
	}

	if err := h.repo.UpsertRune(c.Context(), item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create rune",
			Code:    500,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Rune created"})
}

func (h *AdminHandler) updateRune(c *fiber.Ctx, id int) error {
	var req dto.CreateRuneRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	item := &d2.Rune{
		Code:       req.Code,
		Name:       req.Name,
		RuneNumber: req.RuneNumber,
		LevelReq:   req.LevelReq,
		WeaponMods: convertInputProperties(req.WeaponMods),
		HelmMods:   convertInputProperties(req.ArmorMods),
		ShieldMods: convertInputProperties(req.ShieldMods),
		ImageURL:   req.ImageURL,
	}

	if err := h.repo.UpdateRuneFields(c.Context(), id, item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update rune",
			Code:    500,
		})
	}

	updated, err := h.repo.GetRune(c.Context(), id)
	if err != nil {
		return c.JSON(fiber.Map{"message": "Rune updated"})
	}

	return c.JSON(updated)
}

// Gem CRUD

func (h *AdminHandler) createGem(c *fiber.Ctx) error {
	var req dto.CreateGemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	if req.Code == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Code and name are required",
			Code:    400,
		})
	}

	item := &d2.Gem{
		Code:       req.Code,
		Name:       req.Name,
		GemType:    req.GemType,
		Quality:    req.Quality,
		WeaponMods: convertInputProperties(req.WeaponMods),
		HelmMods:   convertInputProperties(req.ArmorMods),
		ShieldMods: convertInputProperties(req.ShieldMods),
		ImageURL:   req.ImageURL,
	}

	if err := h.repo.UpsertGem(c.Context(), item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create gem",
			Code:    500,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Gem created"})
}

func (h *AdminHandler) updateGem(c *fiber.Ctx, id int) error {
	var req dto.CreateGemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	item := &d2.Gem{
		Code:       req.Code,
		Name:       req.Name,
		GemType:    req.GemType,
		Quality:    req.Quality,
		WeaponMods: convertInputProperties(req.WeaponMods),
		HelmMods:   convertInputProperties(req.ArmorMods),
		ShieldMods: convertInputProperties(req.ShieldMods),
		ImageURL:   req.ImageURL,
	}

	if err := h.repo.UpdateGemFields(c.Context(), id, item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update gem",
			Code:    500,
		})
	}

	updated, err := h.repo.GetGem(c.Context(), id)
	if err != nil {
		return c.JSON(fiber.Map{"message": "Gem updated"})
	}

	return c.JSON(updated)
}

// Base item CRUD

func (h *AdminHandler) createBaseItem(c *fiber.Ctx) error {
	var req dto.CreateBaseItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	if req.Code == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Code and name are required",
			Code:    400,
		})
	}

	item := &d2.ItemBase{
		Code:          req.Code,
		Name:          req.Name,
		Category:      req.Category,
		ItemType:      req.ItemType,
		LevelReq:      req.LevelReq,
		StrReq:        req.StrReq,
		DexReq:        req.DexReq,
		MinAC:         req.MinAC,
		MaxAC:         req.MaxAC,
		MinDam:        req.MinDam,
		MaxDam:        req.MaxDam,
		TwoHandMinDam: req.TwoHandMinDam,
		TwoHandMaxDam: req.TwoHandMaxDam,
		MaxSockets:    req.MaxSockets,
		Durability:    req.Durability,
		Speed:         req.Speed,
		ImageURL:      req.ImageURL,
		Spawnable:     true,
	}

	if err := h.repo.UpsertItemBase(c.Context(), item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create base item",
			Code:    500,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Base item created"})
}

func (h *AdminHandler) updateBaseItem(c *fiber.Ctx, id int) error {
	var req dto.CreateBaseItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	item := &d2.ItemBase{
		Code:          req.Code,
		Name:          req.Name,
		Category:      req.Category,
		ItemType:      req.ItemType,
		LevelReq:      req.LevelReq,
		StrReq:        req.StrReq,
		DexReq:        req.DexReq,
		MinAC:         req.MinAC,
		MaxAC:         req.MaxAC,
		MinDam:        req.MinDam,
		MaxDam:        req.MaxDam,
		TwoHandMinDam: req.TwoHandMinDam,
		TwoHandMaxDam: req.TwoHandMaxDam,
		MaxSockets:    req.MaxSockets,
		Durability:    req.Durability,
		Speed:         req.Speed,
		ImageURL:      req.ImageURL,
	}

	if err := h.repo.UpdateItemBaseFields(c.Context(), id, item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update base item",
			Code:    500,
		})
	}

	updated, err := h.repo.GetItemBase(c.Context(), id)
	if err != nil {
		return c.JSON(fiber.Map{"message": "Base item updated"})
	}

	return c.JSON(updated)
}

// Quest item CRUD

func (h *AdminHandler) createQuestItem(c *fiber.Ctx) error {
	var req dto.CreateQuestItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	if req.Code == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Code and name are required",
			Code:    400,
		})
	}

	item := &d2.ItemBase{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		QuestItem:   true,
	}

	id, err := h.repo.CreateQuestItem(c.Context(), item)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create quest item",
			Code:    500,
		})
	}

	created, err := h.repo.GetItemBase(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Quest item created", "id": id})
	}

	return c.Status(fiber.StatusCreated).JSON(&dto.QuestItemDetail{
		ID:          created.ID,
		Code:        created.Code,
		Name:        created.Name,
		Description: created.Description,
		Type:        "Quest",
		Rarity:      "Quest",
		ImageURL:    created.ImageURL,
	})
}

func (h *AdminHandler) updateQuestItem(c *fiber.Ctx, id int) error {
	var req dto.CreateQuestItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	// Verify this is actually a quest item
	existing, err := h.repo.GetItemBase(c.Context(), id)
	if err != nil || !existing.QuestItem {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Quest item not found",
			Code:    404,
		})
	}

	item := &d2.ItemBase{
		Code:        req.Code,
		Name:        req.Name,
		Category:    "misc",
		ItemType:    "ques",
		Description: req.Description,
		ImageURL:    req.ImageURL,
	}

	if err := h.repo.UpdateItemBaseFields(c.Context(), id, item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update quest item",
			Code:    500,
		})
	}

	updated, err := h.repo.GetItemBase(c.Context(), id)
	if err != nil {
		return c.JSON(fiber.Map{"message": "Quest item updated"})
	}

	return c.JSON(&dto.QuestItemDetail{
		ID:          updated.ID,
		Code:        updated.Code,
		Name:        updated.Name,
		Description: updated.Description,
		Type:        "Quest",
		Rarity:      "Quest",
		ImageURL:    updated.ImageURL,
	})
}

// Class CRUD

// CreateClass handles creating a new class
// POST /admin/d2/classes
func (h *AdminHandler) CreateClass(c *fiber.Ctx) error {
	var req dto.CreateClassRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	if req.ID == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "ID and name are required",
			Code:    400,
		})
	}

	skillTrees := make([]d2.SkillTree, 0, len(req.SkillTrees))
	for _, st := range req.SkillTrees {
		skillTrees = append(skillTrees, d2.SkillTree{
			Name:   st.Name,
			Skills: st.Skills,
		})
	}

	cls := &d2.Class{
		ID:          req.ID,
		Name:        req.Name,
		SkillSuffix: req.SkillSuffix,
		SkillTrees:  skillTrees,
	}

	if err := h.repo.UpsertClass(c.Context(), cls); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create class",
			Code:    500,
		})
	}

	created, err := h.repo.GetClass(c.Context(), req.ID)
	if err != nil {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Class created"})
	}

	return c.Status(fiber.StatusCreated).JSON(convertClassToDTO(created))
}

// UpdateClass handles updating an existing class
// PUT /admin/d2/classes/:classId
func (h *AdminHandler) UpdateClass(c *fiber.Ctx) error {
	classID := c.Params("classId")

	var req dto.UpdateClassRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
			Code:    400,
		})
	}

	// Verify class exists
	_, err := h.repo.GetClass(c.Context(), classID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Class not found",
			Code:    404,
		})
	}

	skillTrees := make([]d2.SkillTree, 0, len(req.SkillTrees))
	for _, st := range req.SkillTrees {
		skillTrees = append(skillTrees, d2.SkillTree{
			Name:   st.Name,
			Skills: st.Skills,
		})
	}

	cls := &d2.Class{
		ID:          classID,
		Name:        req.Name,
		SkillSuffix: req.SkillSuffix,
		SkillTrees:  skillTrees,
	}

	if err := h.repo.UpsertClass(c.Context(), cls); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update class",
			Code:    500,
		})
	}

	updated, err := h.repo.GetClass(c.Context(), classID)
	if err != nil {
		return c.JSON(fiber.Map{"message": "Class updated"})
	}

	return c.JSON(convertClassToDTO(updated))
}
