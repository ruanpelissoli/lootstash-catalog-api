package d2

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ItemType operations
func (r *Repository) ItemTypeExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM d2.item_types WHERE code = $1)", code).Scan(&exists)
	return exists, err
}

func (r *Repository) UpsertItemType(ctx context.Context, it *ItemType) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.item_types (code, name, equiv1, equiv2, body_loc1, body_loc2, can_be_magic, can_be_rare,
			max_sockets_normal, max_sockets_nightmare, max_sockets_hell, staff_mods, class_restriction, store_page)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			equiv1 = EXCLUDED.equiv1,
			equiv2 = EXCLUDED.equiv2,
			body_loc1 = EXCLUDED.body_loc1,
			body_loc2 = EXCLUDED.body_loc2,
			can_be_magic = EXCLUDED.can_be_magic,
			can_be_rare = EXCLUDED.can_be_rare,
			max_sockets_normal = EXCLUDED.max_sockets_normal,
			max_sockets_nightmare = EXCLUDED.max_sockets_nightmare,
			max_sockets_hell = EXCLUDED.max_sockets_hell,
			staff_mods = EXCLUDED.staff_mods,
			class_restriction = EXCLUDED.class_restriction,
			store_page = EXCLUDED.store_page,
			updated_at = NOW()`,
		it.Code, it.Name, nullString(it.Equiv1), nullString(it.Equiv2), nullString(it.BodyLoc1), nullString(it.BodyLoc2),
		it.CanBeMagic, it.CanBeRare, it.MaxSocketsNormal, it.MaxSocketsNightmare, it.MaxSocketsHell,
		nullString(it.StaffMods), nullString(it.ClassRestriction), nullString(it.StorePage))
	return err
}

// ItemBase operations
func (r *Repository) ItemBaseExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM d2.item_bases WHERE code = $1)", code).Scan(&exists)
	return exists, err
}

func (r *Repository) UpsertItemBase(ctx context.Context, ib *ItemBase) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.item_bases (code, name, item_type, item_type2, category, level, level_req, str_req, dex_req,
			durability, min_ac, max_ac, min_dam, max_dam, two_hand_min_dam, two_hand_max_dam, range_adder, speed,
			str_bonus, dex_bonus, max_sockets, gem_apply_type, normal_code, exceptional_code, elite_code,
			inv_width, inv_height, inv_file, flippy_file, unique_inv_file, set_inv_file, image_url,
			spawnable, stackable, useable, throwable, quest_item, rarity, cost)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39)
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			item_type = EXCLUDED.item_type,
			item_type2 = EXCLUDED.item_type2,
			category = EXCLUDED.category,
			level = EXCLUDED.level,
			level_req = EXCLUDED.level_req,
			str_req = EXCLUDED.str_req,
			dex_req = EXCLUDED.dex_req,
			durability = EXCLUDED.durability,
			min_ac = EXCLUDED.min_ac,
			max_ac = EXCLUDED.max_ac,
			min_dam = EXCLUDED.min_dam,
			max_dam = EXCLUDED.max_dam,
			two_hand_min_dam = EXCLUDED.two_hand_min_dam,
			two_hand_max_dam = EXCLUDED.two_hand_max_dam,
			range_adder = EXCLUDED.range_adder,
			speed = EXCLUDED.speed,
			str_bonus = EXCLUDED.str_bonus,
			dex_bonus = EXCLUDED.dex_bonus,
			max_sockets = EXCLUDED.max_sockets,
			gem_apply_type = EXCLUDED.gem_apply_type,
			normal_code = EXCLUDED.normal_code,
			exceptional_code = EXCLUDED.exceptional_code,
			elite_code = EXCLUDED.elite_code,
			inv_width = EXCLUDED.inv_width,
			inv_height = EXCLUDED.inv_height,
			inv_file = EXCLUDED.inv_file,
			flippy_file = EXCLUDED.flippy_file,
			unique_inv_file = EXCLUDED.unique_inv_file,
			set_inv_file = EXCLUDED.set_inv_file,
			image_url = COALESCE(EXCLUDED.image_url, d2.item_bases.image_url),
			spawnable = EXCLUDED.spawnable,
			stackable = EXCLUDED.stackable,
			useable = EXCLUDED.useable,
			throwable = EXCLUDED.throwable,
			quest_item = EXCLUDED.quest_item,
			rarity = EXCLUDED.rarity,
			cost = EXCLUDED.cost,
			updated_at = NOW()`,
		ib.Code, ib.Name, ib.ItemType, nullString(ib.ItemType2), ib.Category, ib.Level, ib.LevelReq, ib.StrReq, ib.DexReq,
		ib.Durability, ib.MinAC, ib.MaxAC, ib.MinDam, ib.MaxDam, ib.TwoHandMinDam, ib.TwoHandMaxDam, ib.RangeAdder, ib.Speed,
		ib.StrBonus, ib.DexBonus, ib.MaxSockets, ib.GemApplyType, nullString(ib.NormalCode), nullString(ib.ExceptionalCode),
		nullString(ib.EliteCode), ib.InvWidth, ib.InvHeight, nullString(ib.InvFile), nullString(ib.FlippyFile),
		nullString(ib.UniqueInvFile), nullString(ib.SetInvFile), nullString(ib.ImageURL),
		ib.Spawnable, ib.Stackable, ib.Useable, ib.Throwable, ib.QuestItem, ib.Rarity, ib.Cost)
	return err
}

// UniqueItem operations
func (r *Repository) UniqueItemExists(ctx context.Context, indexID int) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM d2.unique_items WHERE index_id = $1)", indexID).Scan(&exists)
	return exists, err
}

func (r *Repository) UpsertUniqueItem(ctx context.Context, ui *UniqueItem) error {
	propsJSON, _ := json.Marshal(ui.Properties)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.unique_items (index_id, name, base_code, base_name, level, level_req, rarity, enabled,
			ladder_only, first_ladder_season, last_ladder_season, properties, inv_transform, chr_transform,
			inv_file, image_url, cost_mult, cost_add)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT (index_id) DO UPDATE SET
			name = EXCLUDED.name,
			base_code = EXCLUDED.base_code,
			base_name = EXCLUDED.base_name,
			level = EXCLUDED.level,
			level_req = EXCLUDED.level_req,
			rarity = EXCLUDED.rarity,
			enabled = EXCLUDED.enabled,
			ladder_only = EXCLUDED.ladder_only,
			first_ladder_season = EXCLUDED.first_ladder_season,
			last_ladder_season = EXCLUDED.last_ladder_season,
			properties = EXCLUDED.properties,
			inv_transform = EXCLUDED.inv_transform,
			chr_transform = EXCLUDED.chr_transform,
			inv_file = EXCLUDED.inv_file,
			image_url = COALESCE(EXCLUDED.image_url, d2.unique_items.image_url),
			cost_mult = EXCLUDED.cost_mult,
			cost_add = EXCLUDED.cost_add,
			updated_at = NOW()`,
		ui.IndexID, ui.Name, ui.BaseCode, nullString(ui.BaseName), ui.Level, ui.LevelReq, ui.Rarity, ui.Enabled,
		ui.LadderOnly, ui.FirstLadderSeason, ui.LastLadderSeason, propsJSON,
		nullString(ui.InvTransform), nullString(ui.ChrTransform), nullString(ui.InvFile), nullString(ui.ImageURL),
		ui.CostMult, ui.CostAdd)
	return err
}

// SetBonus operations
func (r *Repository) SetBonusExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM d2.set_bonuses WHERE name = $1)", name).Scan(&exists)
	return exists, err
}

func (r *Repository) UpsertSetBonus(ctx context.Context, sb *SetBonus) error {
	partialJSON, _ := json.Marshal(sb.PartialBonuses)
	fullJSON, _ := json.Marshal(sb.FullBonuses)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.set_bonuses (index_id, name, version, partial_bonuses, full_bonuses)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (name) DO UPDATE SET
			version = EXCLUDED.version,
			partial_bonuses = EXCLUDED.partial_bonuses,
			full_bonuses = EXCLUDED.full_bonuses,
			updated_at = NOW()`,
		sb.IndexID, sb.Name, sb.Version, partialJSON, fullJSON)
	return err
}

// SetItem operations
func (r *Repository) SetItemExists(ctx context.Context, indexID int) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM d2.set_items WHERE index_id = $1)", indexID).Scan(&exists)
	return exists, err
}

func (r *Repository) UpsertSetItem(ctx context.Context, si *SetItem) error {
	propsJSON, _ := json.Marshal(si.Properties)
	bonusJSON, _ := json.Marshal(si.BonusProperties)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.set_items (index_id, name, set_name, base_code, base_name, level, level_req, rarity,
			properties, bonus_properties, inv_transform, chr_transform, inv_file, image_url, cost_mult, cost_add)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (index_id) DO UPDATE SET
			name = EXCLUDED.name,
			set_name = EXCLUDED.set_name,
			base_code = EXCLUDED.base_code,
			base_name = EXCLUDED.base_name,
			level = EXCLUDED.level,
			level_req = EXCLUDED.level_req,
			rarity = EXCLUDED.rarity,
			properties = EXCLUDED.properties,
			bonus_properties = EXCLUDED.bonus_properties,
			inv_transform = EXCLUDED.inv_transform,
			chr_transform = EXCLUDED.chr_transform,
			inv_file = EXCLUDED.inv_file,
			image_url = COALESCE(EXCLUDED.image_url, d2.set_items.image_url),
			cost_mult = EXCLUDED.cost_mult,
			cost_add = EXCLUDED.cost_add,
			updated_at = NOW()`,
		si.IndexID, si.Name, si.SetName, si.BaseCode, nullString(si.BaseName), si.Level, si.LevelReq, si.Rarity,
		propsJSON, bonusJSON, nullString(si.InvTransform), nullString(si.ChrTransform),
		nullString(si.InvFile), nullString(si.ImageURL), si.CostMult, si.CostAdd)
	return err
}

// Runeword operations
func (r *Repository) RunewordExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM d2.runewords WHERE name = $1)", name).Scan(&exists)
	return exists, err
}

func (r *Repository) UpsertRuneword(ctx context.Context, rw *Runeword) error {
	validTypesJSON, _ := json.Marshal(rw.ValidItemTypes)
	excludedTypesJSON, _ := json.Marshal(rw.ExcludedItemTypes)
	runesJSON, _ := json.Marshal(rw.Runes)
	propsJSON, _ := json.Marshal(rw.Properties)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.runewords (name, display_name, complete, ladder_only, first_ladder_season, last_ladder_season,
			valid_item_types, excluded_item_types, runes, properties, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (name) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			complete = EXCLUDED.complete,
			ladder_only = EXCLUDED.ladder_only,
			first_ladder_season = EXCLUDED.first_ladder_season,
			last_ladder_season = EXCLUDED.last_ladder_season,
			valid_item_types = EXCLUDED.valid_item_types,
			excluded_item_types = EXCLUDED.excluded_item_types,
			runes = EXCLUDED.runes,
			properties = EXCLUDED.properties,
			image_url = COALESCE(EXCLUDED.image_url, d2.runewords.image_url),
			updated_at = NOW()`,
		rw.Name, rw.DisplayName, rw.Complete, rw.LadderOnly, rw.FirstLadderSeason, rw.LastLadderSeason,
		validTypesJSON, excludedTypesJSON, runesJSON, propsJSON, nullString(rw.ImageURL))
	return err
}

// Rune operations
func (r *Repository) RuneExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM d2.runes WHERE code = $1)", code).Scan(&exists)
	return exists, err
}

func (r *Repository) UpsertRune(ctx context.Context, rn *Rune) error {
	weaponJSON, _ := json.Marshal(rn.WeaponMods)
	helmJSON, _ := json.Marshal(rn.HelmMods)
	shieldJSON, _ := json.Marshal(rn.ShieldMods)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.runes (code, name, rune_number, level, level_req, weapon_mods, helm_mods, shield_mods, inv_file, image_url, cost)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			rune_number = EXCLUDED.rune_number,
			level = EXCLUDED.level,
			level_req = EXCLUDED.level_req,
			weapon_mods = EXCLUDED.weapon_mods,
			helm_mods = EXCLUDED.helm_mods,
			shield_mods = EXCLUDED.shield_mods,
			inv_file = EXCLUDED.inv_file,
			image_url = COALESCE(EXCLUDED.image_url, d2.runes.image_url),
			cost = EXCLUDED.cost,
			updated_at = NOW()`,
		rn.Code, rn.Name, rn.RuneNumber, rn.Level, rn.LevelReq, weaponJSON, helmJSON, shieldJSON,
		nullString(rn.InvFile), nullString(rn.ImageURL), rn.Cost)
	return err
}

// Gem operations
func (r *Repository) GemExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM d2.gems WHERE code = $1)", code).Scan(&exists)
	return exists, err
}

func (r *Repository) UpsertGem(ctx context.Context, g *Gem) error {
	weaponJSON, _ := json.Marshal(g.WeaponMods)
	helmJSON, _ := json.Marshal(g.HelmMods)
	shieldJSON, _ := json.Marshal(g.ShieldMods)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.gems (code, name, gem_type, quality, weapon_mods, helm_mods, shield_mods, transform, inv_file, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			gem_type = EXCLUDED.gem_type,
			quality = EXCLUDED.quality,
			weapon_mods = EXCLUDED.weapon_mods,
			helm_mods = EXCLUDED.helm_mods,
			shield_mods = EXCLUDED.shield_mods,
			transform = EXCLUDED.transform,
			inv_file = EXCLUDED.inv_file,
			image_url = COALESCE(EXCLUDED.image_url, d2.gems.image_url),
			updated_at = NOW()`,
		g.Code, g.Name, g.GemType, g.Quality, weaponJSON, helmJSON, shieldJSON,
		g.Transform, nullString(g.InvFile), nullString(g.ImageURL))
	return err
}

// Property operations
func (r *Repository) PropertyExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM d2.properties WHERE code = $1)", code).Scan(&exists)
	return exists, err
}

func (r *Repository) UpsertProperty(ctx context.Context, p *ItemProperty) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.properties (code, enabled, func1, stat1, func2, stat2, func3, stat3, func4, stat4,
			func5, stat5, func6, stat6, func7, stat7, tooltip)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (code) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			func1 = EXCLUDED.func1,
			stat1 = EXCLUDED.stat1,
			func2 = EXCLUDED.func2,
			stat2 = EXCLUDED.stat2,
			func3 = EXCLUDED.func3,
			stat3 = EXCLUDED.stat3,
			func4 = EXCLUDED.func4,
			stat4 = EXCLUDED.stat4,
			func5 = EXCLUDED.func5,
			stat5 = EXCLUDED.stat5,
			func6 = EXCLUDED.func6,
			stat6 = EXCLUDED.stat6,
			func7 = EXCLUDED.func7,
			stat7 = EXCLUDED.stat7,
			tooltip = EXCLUDED.tooltip,
			updated_at = NOW()`,
		p.Code, p.Enabled, p.Func1, nullString(p.Stat1), p.Func2, nullString(p.Stat2),
		p.Func3, nullString(p.Stat3), p.Func4, nullString(p.Stat4), p.Func5, nullString(p.Stat5),
		p.Func6, nullString(p.Stat6), p.Func7, nullString(p.Stat7), nullString(p.Tooltip))
	return err
}

// Affix operations
func (r *Repository) AffixExists(ctx context.Context, name, affixType string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM d2.affixes WHERE name = $1 AND affix_type = $2)", name, affixType).Scan(&exists)
	return exists, err
}

func (r *Repository) UpsertAffix(ctx context.Context, a *Affix) error {
	validTypesJSON, _ := json.Marshal(a.ValidItemTypes)
	excludedTypesJSON, _ := json.Marshal(a.ExcludedItemTypes)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.affixes (name, affix_type, version, spawnable, rare, level, max_level, level_req,
			class_specific, class_level_req, frequency, affix_group, mod1_code, mod1_param, mod1_min, mod1_max,
			mod2_code, mod2_param, mod2_min, mod2_max, mod3_code, mod3_param, mod3_min, mod3_max,
			valid_item_types, excluded_item_types, transform_color, multiply, add_cost)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27, $28, $29)
		ON CONFLICT (name, affix_type) DO UPDATE SET
			version = EXCLUDED.version,
			spawnable = EXCLUDED.spawnable,
			rare = EXCLUDED.rare,
			level = EXCLUDED.level,
			max_level = EXCLUDED.max_level,
			level_req = EXCLUDED.level_req,
			class_specific = EXCLUDED.class_specific,
			class_level_req = EXCLUDED.class_level_req,
			frequency = EXCLUDED.frequency,
			affix_group = EXCLUDED.affix_group,
			mod1_code = EXCLUDED.mod1_code,
			mod1_param = EXCLUDED.mod1_param,
			mod1_min = EXCLUDED.mod1_min,
			mod1_max = EXCLUDED.mod1_max,
			mod2_code = EXCLUDED.mod2_code,
			mod2_param = EXCLUDED.mod2_param,
			mod2_min = EXCLUDED.mod2_min,
			mod2_max = EXCLUDED.mod2_max,
			mod3_code = EXCLUDED.mod3_code,
			mod3_param = EXCLUDED.mod3_param,
			mod3_min = EXCLUDED.mod3_min,
			mod3_max = EXCLUDED.mod3_max,
			valid_item_types = EXCLUDED.valid_item_types,
			excluded_item_types = EXCLUDED.excluded_item_types,
			transform_color = EXCLUDED.transform_color,
			multiply = EXCLUDED.multiply,
			add_cost = EXCLUDED.add_cost,
			updated_at = NOW()`,
		a.Name, a.AffixType, a.Version, a.Spawnable, a.Rare, a.Level, a.MaxLevel, a.LevelReq,
		nullString(a.ClassSpecific), a.ClassLevelReq, a.Frequency, a.AffixGroup,
		nullString(a.Mod1Code), nullString(a.Mod1Param), a.Mod1Min, a.Mod1Max,
		nullString(a.Mod2Code), nullString(a.Mod2Param), a.Mod2Min, a.Mod2Max,
		nullString(a.Mod3Code), nullString(a.Mod3Param), a.Mod3Min, a.Mod3Max,
		validTypesJSON, excludedTypesJSON, nullString(a.TransformColor), a.Multiply, a.AddCost)
	return err
}

// Helper function to convert empty strings to nil
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// Utility function to print database error context
func (r *Repository) LogError(operation string, err error) error {
	return fmt.Errorf("%s failed: %w", operation, err)
}

// ItemWithoutImage represents an item that needs an image
type ItemWithoutImage struct {
	ID   int
	Code string // For item_bases, runes, gems
	Name string
	Type string // "unique", "set", "base", "rune", "gem"
}

// RunewordWithRunes represents a runeword with its rune codes for image generation
type RunewordWithRunes struct {
	ID          int
	Name        string   // Internal name (e.g., "Runeword33")
	DisplayName string   // Display name (e.g., "Enigma")
	Runes       []string // Rune codes ["r31", "r06", "r30"]
	ImageURL    string
}

// GetUniqueItemsWithoutImages returns unique items that don't have images
func (r *Repository) GetUniqueItemsWithoutImages(ctx context.Context) ([]ItemWithoutImage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name FROM d2.unique_items
		WHERE image_url IS NULL OR image_url = ''
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ItemWithoutImage
	for rows.Next() {
		var item ItemWithoutImage
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		item.Type = "unique"
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetSetItemsWithoutImages returns set items that don't have images
func (r *Repository) GetSetItemsWithoutImages(ctx context.Context) ([]ItemWithoutImage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name FROM d2.set_items
		WHERE image_url IS NULL OR image_url = ''
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ItemWithoutImage
	for rows.Next() {
		var item ItemWithoutImage
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		item.Type = "set"
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetItemBasesWithoutImages returns item bases that don't have images
func (r *Repository) GetItemBasesWithoutImages(ctx context.Context) ([]ItemWithoutImage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT code, name FROM d2.item_bases
		WHERE (image_url IS NULL OR image_url = '')
		ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ItemWithoutImage
	for rows.Next() {
		var item ItemWithoutImage
		if err := rows.Scan(&item.Code, &item.Name); err != nil {
			return nil, err
		}
		item.Type = "base"
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetRunesWithoutImages returns runes that don't have images
func (r *Repository) GetRunesWithoutImages(ctx context.Context) ([]ItemWithoutImage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name FROM d2.runes
		WHERE image_url IS NULL OR image_url = ''
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ItemWithoutImage
	for rows.Next() {
		var item ItemWithoutImage
		if err := rows.Scan(&item.ID, &item.Code, &item.Name); err != nil {
			return nil, err
		}
		item.Type = "rune"
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetGemsWithoutImages returns gems that don't have images
func (r *Repository) GetGemsWithoutImages(ctx context.Context) ([]ItemWithoutImage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name FROM d2.gems
		WHERE image_url IS NULL OR image_url = ''
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ItemWithoutImage
	for rows.Next() {
		var item ItemWithoutImage
		if err := rows.Scan(&item.ID, &item.Code, &item.Name); err != nil {
			return nil, err
		}
		item.Type = "gem"
		items = append(items, item)
	}
	return items, rows.Err()
}

// UpdateUniqueItemImageURL updates the image URL for a unique item
func (r *Repository) UpdateUniqueItemImageURL(ctx context.Context, id int, url string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE d2.unique_items SET image_url = $1, updated_at = NOW() WHERE id = $2`,
		url, id)
	return err
}

// UpdateSetItemImageURL updates the image URL for a set item
func (r *Repository) UpdateSetItemImageURL(ctx context.Context, id int, url string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE d2.set_items SET image_url = $1, updated_at = NOW() WHERE id = $2`,
		url, id)
	return err
}

// UpdateItemBaseImageURL updates the image URL for an item base
func (r *Repository) UpdateItemBaseImageURL(ctx context.Context, code string, url string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE d2.item_bases SET image_url = $1, updated_at = NOW() WHERE code = $2`,
		url, code)
	return err
}

// UpdateItemBaseIconVariants updates the icon variants for an item base
func (r *Repository) UpdateItemBaseIconVariants(ctx context.Context, code string, variants []string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE d2.item_bases SET icon_variants = $1, updated_at = NOW() WHERE code = $2`,
		variants, code)
	return err
}

// UpdateRuneImageURL updates the image URL for a rune
func (r *Repository) UpdateRuneImageURL(ctx context.Context, id int, url string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE d2.runes SET image_url = $1, updated_at = NOW() WHERE id = $2`,
		url, id)
	return err
}

// UpdateGemImageURL updates the image URL for a gem
func (r *Repository) UpdateGemImageURL(ctx context.Context, id int, url string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE d2.gems SET image_url = $1, updated_at = NOW() WHERE id = $2`,
		url, id)
	return err
}

// GetRuneCodeToNameMap returns a mapping of rune codes to rune names (e.g., "r30" -> "Ber")
func (r *Repository) GetRuneCodeToNameMap(ctx context.Context) (map[string]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT code, name FROM d2.runes ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var code, name string
		if err := rows.Scan(&code, &name); err != nil {
			return nil, err
		}
		// Store just the rune name without " Rune" suffix
		// DB has "Ber Rune", we want just "Ber"
		cleanName := name
		if len(name) > 5 && name[len(name)-5:] == " Rune" {
			cleanName = name[:len(name)-5]
		}
		result[code] = cleanName
	}
	return result, rows.Err()
}

// GetRunewordsWithoutImages returns runewords that don't have images yet
func (r *Repository) GetRunewordsWithoutImages(ctx context.Context) ([]RunewordWithRunes, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, display_name, runes, COALESCE(image_url, '')
		FROM d2.runewords
		WHERE image_url IS NULL OR image_url = ''
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRunewords(rows)
}

// GetAllRunewords returns all runewords (for force regeneration)
func (r *Repository) GetAllRunewords(ctx context.Context) ([]RunewordWithRunes, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, display_name, runes, COALESCE(image_url, '')
		FROM d2.runewords
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRunewords(rows)
}

// scanRunewords scans rows into RunewordWithRunes slice
func scanRunewords(rows interface{ Next() bool; Scan(dest ...interface{}) error; Err() error }) ([]RunewordWithRunes, error) {
	var runewords []RunewordWithRunes
	for rows.Next() {
		var rw RunewordWithRunes
		var runesJSON []byte
		if err := rows.Scan(&rw.ID, &rw.Name, &rw.DisplayName, &runesJSON, &rw.ImageURL); err != nil {
			return nil, err
		}
		// Parse runes JSON array
		if err := json.Unmarshal(runesJSON, &rw.Runes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal runes for %s: %w", rw.Name, err)
		}
		runewords = append(runewords, rw)
	}
	return runewords, rows.Err()
}

// UpdateRunewordImageURL updates the image URL for a runeword
func (r *Repository) UpdateRunewordImageURL(ctx context.Context, id int, url string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE d2.runewords SET image_url = $1, updated_at = NOW() WHERE id = $2`,
		url, id)
	return err
}

// RunewordBase operations

// ClearRunewordBases removes all runeword base mappings
func (r *Repository) ClearRunewordBases(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM d2.runeword_bases`)
	return err
}

// InsertRunewordBase inserts a runeword-base mapping
func (r *Repository) InsertRunewordBase(ctx context.Context, rb *RunewordBase) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.runeword_bases (runeword_id, item_base_id, item_base_code, item_base_name, category, max_sockets, required_sockets)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (runeword_id, item_base_id) DO NOTHING`,
		rb.RunewordID, rb.ItemBaseID, rb.ItemBaseCode, rb.ItemBaseName, rb.Category, rb.MaxSockets, rb.RequiredSockets)
	return err
}

// GetBasesForRuneword returns all valid base items for a runeword
func (r *Repository) GetBasesForRuneword(ctx context.Context, runewordID int) ([]RunewordBase, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, runeword_id, item_base_id, item_base_code, item_base_name, category, max_sockets, required_sockets, created_at
		FROM d2.runeword_bases
		WHERE runeword_id = $1
		ORDER BY category, item_base_name`, runewordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bases []RunewordBase
	for rows.Next() {
		var rb RunewordBase
		if err := rows.Scan(&rb.ID, &rb.RunewordID, &rb.ItemBaseID, &rb.ItemBaseCode, &rb.ItemBaseName, &rb.Category, &rb.MaxSockets, &rb.RequiredSockets, &rb.CreatedAt); err != nil {
			return nil, err
		}
		bases = append(bases, rb)
	}
	return bases, rows.Err()
}

// ItemTypeWithEquiv holds item type info with parent types for hierarchy building
type ItemTypeWithEquiv struct {
	Code   string
	Equiv1 string
	Equiv2 string
}

// GetAllItemTypesWithEquiv returns all item types with their equiv relationships
func (r *Repository) GetAllItemTypesWithEquiv(ctx context.Context) ([]ItemTypeWithEquiv, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT code, COALESCE(equiv1, ''), COALESCE(equiv2, '')
		FROM d2.item_types`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []ItemTypeWithEquiv
	for rows.Next() {
		var it ItemTypeWithEquiv
		if err := rows.Scan(&it.Code, &it.Equiv1, &it.Equiv2); err != nil {
			return nil, err
		}
		types = append(types, it)
	}
	return types, rows.Err()
}

// ItemBaseForRuneword holds base item info needed for runeword matching
type ItemBaseForRuneword struct {
	ID         int
	Code       string
	Name       string
	ItemType   string
	ItemType2  string
	Category   string
	MaxSockets int
}

// GetAllItemBasesForRunewordMatching returns all base items with socket info
func (r *Repository) GetAllItemBasesForRunewordMatching(ctx context.Context) ([]ItemBaseForRuneword, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name, item_type, COALESCE(item_type2, ''), category, max_sockets
		FROM d2.item_bases
		WHERE max_sockets > 0`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bases []ItemBaseForRuneword
	for rows.Next() {
		var ib ItemBaseForRuneword
		if err := rows.Scan(&ib.ID, &ib.Code, &ib.Name, &ib.ItemType, &ib.ItemType2, &ib.Category, &ib.MaxSockets); err != nil {
			return nil, err
		}
		bases = append(bases, ib)
	}
	return bases, rows.Err()
}

// RunewordForMatching holds runeword info needed for base matching
type RunewordForMatching struct {
	ID                int
	Name              string
	ValidItemTypes    []string
	ExcludedItemTypes []string
	RuneCount         int
}

// GetAllRunewordsForMatching returns all runewords with their type requirements
func (r *Repository) GetAllRunewordsForMatching(ctx context.Context) ([]RunewordForMatching, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, valid_item_types, excluded_item_types, runes
		FROM d2.runewords
		WHERE complete = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runewords []RunewordForMatching
	for rows.Next() {
		var rw RunewordForMatching
		var validTypesJSON, excludedTypesJSON, runesJSON []byte
		if err := rows.Scan(&rw.ID, &rw.Name, &validTypesJSON, &excludedTypesJSON, &runesJSON); err != nil {
			return nil, err
		}

		var validTypes, excludedTypes, runes []string
		json.Unmarshal(validTypesJSON, &validTypes)
		json.Unmarshal(excludedTypesJSON, &excludedTypes)
		json.Unmarshal(runesJSON, &runes)

		rw.ValidItemTypes = validTypes
		rw.ExcludedItemTypes = excludedTypes
		rw.RuneCount = len(runes)

		runewords = append(runewords, rw)
	}
	return runewords, rows.Err()
}

// TreasureClass operations

// UpsertTreasureClass inserts or updates a treasure class
func (r *Repository) UpsertTreasureClass(ctx context.Context, tc *TreasureClass) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, `
		INSERT INTO d2.treasure_classes (name, group_id, level, picks, unique_mod, set_mod, rare_mod, magic_mod,
			no_drop, first_ladder_season, last_ladder_season, no_always_spawn)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (name) DO UPDATE SET
			group_id = EXCLUDED.group_id,
			level = EXCLUDED.level,
			picks = EXCLUDED.picks,
			unique_mod = EXCLUDED.unique_mod,
			set_mod = EXCLUDED.set_mod,
			rare_mod = EXCLUDED.rare_mod,
			magic_mod = EXCLUDED.magic_mod,
			no_drop = EXCLUDED.no_drop,
			first_ladder_season = EXCLUDED.first_ladder_season,
			last_ladder_season = EXCLUDED.last_ladder_season,
			no_always_spawn = EXCLUDED.no_always_spawn,
			updated_at = NOW()
		RETURNING id`,
		tc.Name, tc.GroupID, tc.Level, tc.Picks, tc.UniqueMod, tc.SetMod, tc.RareMod, tc.MagicMod,
		tc.NoDrop, tc.FirstLadderSeason, tc.LastLadderSeason, tc.NoAlwaysSpawn).Scan(&id)
	return id, err
}

// UpsertTreasureClassItem inserts or updates a treasure class item
func (r *Repository) UpsertTreasureClassItem(ctx context.Context, tci *TreasureClassItem) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.treasure_class_items (treasure_class_id, slot, item_code, is_treasure_class, probability)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (treasure_class_id, slot) DO UPDATE SET
			item_code = EXCLUDED.item_code,
			is_treasure_class = EXCLUDED.is_treasure_class,
			probability = EXCLUDED.probability`,
		tci.TreasureClassID, tci.Slot, tci.ItemCode, tci.IsTreasureClass, tci.Probability)
	return err
}

// ItemRatio operations

// UpsertItemRatio inserts or updates an item ratio
func (r *Repository) UpsertItemRatio(ctx context.Context, ir *ItemRatio) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.item_ratios (function_name, version, is_uber, is_class_specific,
			unique_ratio, unique_divisor, unique_min, rare_ratio, rare_divisor, rare_min,
			set_ratio, set_divisor, set_min, magic_ratio, magic_divisor, magic_min,
			hiquality_ratio, hiquality_divisor, normal_ratio, normal_divisor)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		ON CONFLICT (version, is_uber, is_class_specific) DO UPDATE SET
			function_name = EXCLUDED.function_name,
			unique_ratio = EXCLUDED.unique_ratio,
			unique_divisor = EXCLUDED.unique_divisor,
			unique_min = EXCLUDED.unique_min,
			rare_ratio = EXCLUDED.rare_ratio,
			rare_divisor = EXCLUDED.rare_divisor,
			rare_min = EXCLUDED.rare_min,
			set_ratio = EXCLUDED.set_ratio,
			set_divisor = EXCLUDED.set_divisor,
			set_min = EXCLUDED.set_min,
			magic_ratio = EXCLUDED.magic_ratio,
			magic_divisor = EXCLUDED.magic_divisor,
			magic_min = EXCLUDED.magic_min,
			hiquality_ratio = EXCLUDED.hiquality_ratio,
			hiquality_divisor = EXCLUDED.hiquality_divisor,
			normal_ratio = EXCLUDED.normal_ratio,
			normal_divisor = EXCLUDED.normal_divisor,
			updated_at = NOW()`,
		ir.FunctionName, ir.Version, ir.IsUber, ir.IsClassSpecific,
		ir.UniqueRatio, ir.UniqueDivisor, ir.UniqueMin,
		ir.RareRatio, ir.RareDivisor, ir.RareMin,
		ir.SetRatio, ir.SetDivisor, ir.SetMin,
		ir.MagicRatio, ir.MagicDivisor, ir.MagicMin,
		ir.HiQualityRatio, ir.HiQualityDivisor,
		ir.NormalRatio, ir.NormalDivisor)
	return err
}

// RuneInfo holds basic rune display info
type RuneInfo struct {
	ID       int
	Code     string
	Name     string
	ImageURL string
}

// GetRunesByCodes returns rune info for the given codes
func (r *Repository) GetRunesByCodes(ctx context.Context, codes []string) (map[string]RuneInfo, error) {
	if len(codes) == 0 {
		return make(map[string]RuneInfo), nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name, COALESCE(image_url, '')
		FROM d2.runes
		WHERE code = ANY($1)`, codes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]RuneInfo)
	for rows.Next() {
		var ri RuneInfo
		if err := rows.Scan(&ri.ID, &ri.Code, &ri.Name, &ri.ImageURL); err != nil {
			return nil, err
		}
		result[ri.Code] = ri
	}
	return result, rows.Err()
}

// ItemTypeInfo holds basic item type display info
type ItemTypeInfo struct {
	Code string
	Name string
}

// GetItemTypesByCodes returns item type info for the given codes
func (r *Repository) GetItemTypesByCodes(ctx context.Context, codes []string) (map[string]ItemTypeInfo, error) {
	if len(codes) == 0 {
		return make(map[string]ItemTypeInfo), nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT code, name
		FROM d2.item_types
		WHERE code = ANY($1)`, codes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]ItemTypeInfo)
	for rows.Next() {
		var it ItemTypeInfo
		if err := rows.Scan(&it.Code, &it.Name); err != nil {
			return nil, err
		}
		result[it.Code] = it
	}
	return result, rows.Err()
}

// GetAllItemBaseNameToCode returns a mapping of base item names to codes (e.g., "Kris" -> "kri")
func (r *Repository) GetAllItemBaseNameToCode(ctx context.Context) (map[string]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT name, code FROM d2.item_bases`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var name, code string
		if err := rows.Scan(&name, &code); err != nil {
			return nil, err
		}
		result[name] = code
	}
	return result, rows.Err()
}

// GetRuneNameToCodeMap returns a mapping of rune names to codes (e.g., "Shael" -> "r13")
func (r *Repository) GetRuneNameToCodeMap(ctx context.Context) (map[string]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT name, code FROM d2.runes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var name, code string
		if err := rows.Scan(&name, &code); err != nil {
			return nil, err
		}
		// Store both full name ("Shael Rune") and short name ("Shael")
		result[name] = code
		cleanName := name
		if len(name) > 5 && name[len(name)-5:] == " Rune" {
			cleanName = name[:len(name)-5]
		}
		result[cleanName] = code
	}
	return result, rows.Err()
}

// GetAllExistingNames returns all names from a table column as a set for deduplication
func (r *Repository) GetAllExistingNames(ctx context.Context, table, column string) (map[string]bool, error) {
	// table/column are internal constants, not user input
	query := fmt.Sprintf("SELECT %s FROM d2.%s", column, table)
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result[NormalizeItemName(name)] = true
	}
	return result, rows.Err()
}

// GetNameToIndexID returns a map of normalized name -> index_id for a table
func (r *Repository) GetNameToIndexID(ctx context.Context, table, nameColumn string) (map[string]int, error) {
	query := fmt.Sprintf("SELECT %s, index_id FROM d2.%s", nameColumn, table)
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var name string
		var id int
		if err := rows.Scan(&name, &id); err != nil {
			return nil, err
		}
		result[NormalizeItemName(name)] = id
	}
	return result, rows.Err()
}

// GetMaxIndexID returns the maximum index_id from a table
func (r *Repository) GetMaxIndexID(ctx context.Context, table string) (int, error) {
	query := fmt.Sprintf("SELECT COALESCE(MAX(index_id), 0) FROM d2.%s", table)
	var maxID int
	err := r.pool.QueryRow(ctx, query).Scan(&maxID)
	return maxID, err
}

// Profile operations

// GetProfile retrieves a profile by ID
func (r *Repository) GetProfile(ctx context.Context, id string) (*Profile, error) {
	var p Profile
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, is_admin, created_at, updated_at
		FROM d2.profiles WHERE id = $1`, id).Scan(
		&p.ID, &p.Email, &p.IsAdmin, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get profile failed: %w", err)
	}
	return &p, nil
}

// IsAdmin checks if a user is an admin
func (r *Repository) IsAdmin(ctx context.Context, id string) (bool, error) {
	var isAdmin bool
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(is_admin, false) FROM d2.profiles WHERE id = $1`, id).Scan(&isAdmin)
	if err != nil {
		return false, err
	}
	return isAdmin, nil
}

// Class operations

// GetAllClasses retrieves all classes
func (r *Repository) GetAllClasses(ctx context.Context) ([]Class, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, skill_suffix, skill_trees, created_at, updated_at
		FROM d2.classes ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []Class
	for rows.Next() {
		var c Class
		var skillTreesJSON []byte
		if err := rows.Scan(&c.ID, &c.Name, &c.SkillSuffix, &skillTreesJSON, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		if len(skillTreesJSON) > 0 {
			json.Unmarshal(skillTreesJSON, &c.SkillTrees)
		}
		classes = append(classes, c)
	}
	return classes, rows.Err()
}

// GetClass retrieves a class by ID
func (r *Repository) GetClass(ctx context.Context, id string) (*Class, error) {
	var c Class
	var skillTreesJSON []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, skill_suffix, skill_trees, created_at, updated_at
		FROM d2.classes WHERE id = $1`, id).Scan(
		&c.ID, &c.Name, &c.SkillSuffix, &skillTreesJSON, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get class failed: %w", err)
	}
	if len(skillTreesJSON) > 0 {
		json.Unmarshal(skillTreesJSON, &c.SkillTrees)
	}
	return &c, nil
}

// UpsertClass inserts or updates a class
func (r *Repository) UpsertClass(ctx context.Context, c *Class) error {
	skillTreesJSON, _ := json.Marshal(c.SkillTrees)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO d2.classes (id, name, skill_suffix, skill_trees)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			skill_suffix = EXCLUDED.skill_suffix,
			skill_trees = EXCLUDED.skill_trees,
			updated_at = NOW()`,
		c.ID, c.Name, c.SkillSuffix, skillTreesJSON)
	return err
}

// Quest item operations

// GetAllQuestItems retrieves all quest items
func (r *Repository) GetAllQuestItems(ctx context.Context) ([]ItemBase, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id FROM d2.item_bases WHERE quest_item = true ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ItemBase
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		item, err := r.GetItemBase(ctx, id)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

// CreateQuestItem inserts a new quest item
func (r *Repository) CreateQuestItem(ctx context.Context, ib *ItemBase) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, `
		INSERT INTO d2.item_bases (code, name, item_type, category, quest_item, description, image_url)
		VALUES ($1, $2, 'ques', 'misc', true, $3, $4)
		RETURNING id`,
		ib.Code, ib.Name, nullString(ib.Description), nullString(ib.ImageURL)).Scan(&id)
	return id, err
}

// DeleteQuestItem deletes a quest item by ID (only if it is a quest item)
func (r *Repository) DeleteQuestItem(ctx context.Context, id int) error {
	result, err := r.pool.Exec(ctx, `
		DELETE FROM d2.item_bases WHERE id = $1 AND quest_item = true`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("quest item not found")
	}
	return nil
}

// Admin update operations

// UpdateUniqueItemFields updates specific fields on a unique item
func (r *Repository) UpdateUniqueItemFields(ctx context.Context, id int, item *UniqueItem) error {
	propsJSON, _ := json.Marshal(item.Properties)
	_, err := r.pool.Exec(ctx, `
		UPDATE d2.unique_items SET
			name = $2, base_code = $3, level_req = $4, ladder_only = $5,
			properties = $6, image_url = COALESCE($7, image_url),
			updated_at = NOW()
		WHERE id = $1`,
		id, item.Name, item.BaseCode, item.LevelReq, item.LadderOnly,
		propsJSON, nullString(item.ImageURL))
	return err
}

// UpdateSetItemFields updates specific fields on a set item
func (r *Repository) UpdateSetItemFields(ctx context.Context, id int, item *SetItem) error {
	propsJSON, _ := json.Marshal(item.Properties)
	bonusJSON, _ := json.Marshal(item.BonusProperties)
	_, err := r.pool.Exec(ctx, `
		UPDATE d2.set_items SET
			name = $2, set_name = $3, base_code = $4, level_req = $5,
			properties = $6, bonus_properties = $7,
			image_url = COALESCE($8, image_url),
			updated_at = NOW()
		WHERE id = $1`,
		id, item.Name, item.SetName, item.BaseCode, item.LevelReq,
		propsJSON, bonusJSON, nullString(item.ImageURL))
	return err
}

// UpdateRunewordFields updates specific fields on a runeword
func (r *Repository) UpdateRunewordFields(ctx context.Context, id int, item *Runeword) error {
	validTypesJSON, _ := json.Marshal(item.ValidItemTypes)
	runesJSON, _ := json.Marshal(item.Runes)
	propsJSON, _ := json.Marshal(item.Properties)
	_, err := r.pool.Exec(ctx, `
		UPDATE d2.runewords SET
			name = $2, display_name = $3, ladder_only = $4,
			valid_item_types = $5, runes = $6, properties = $7,
			image_url = COALESCE($8, image_url),
			updated_at = NOW()
		WHERE id = $1`,
		id, item.Name, item.DisplayName, item.LadderOnly,
		validTypesJSON, runesJSON, propsJSON, nullString(item.ImageURL))
	return err
}

// UpdateRuneFields updates specific fields on a rune
func (r *Repository) UpdateRuneFields(ctx context.Context, id int, item *Rune) error {
	weaponJSON, _ := json.Marshal(item.WeaponMods)
	helmJSON, _ := json.Marshal(item.HelmMods)
	shieldJSON, _ := json.Marshal(item.ShieldMods)
	_, err := r.pool.Exec(ctx, `
		UPDATE d2.runes SET
			code = $2, name = $3, rune_number = $4, level_req = $5,
			weapon_mods = $6, helm_mods = $7, shield_mods = $8,
			image_url = COALESCE($9, image_url),
			updated_at = NOW()
		WHERE id = $1`,
		id, item.Code, item.Name, item.RuneNumber, item.LevelReq,
		weaponJSON, helmJSON, shieldJSON, nullString(item.ImageURL))
	return err
}

// UpdateGemFields updates specific fields on a gem
func (r *Repository) UpdateGemFields(ctx context.Context, id int, item *Gem) error {
	weaponJSON, _ := json.Marshal(item.WeaponMods)
	helmJSON, _ := json.Marshal(item.HelmMods)
	shieldJSON, _ := json.Marshal(item.ShieldMods)
	_, err := r.pool.Exec(ctx, `
		UPDATE d2.gems SET
			code = $2, name = $3, gem_type = $4, quality = $5,
			weapon_mods = $6, helm_mods = $7, shield_mods = $8,
			image_url = COALESCE($9, image_url),
			updated_at = NOW()
		WHERE id = $1`,
		id, item.Code, item.Name, item.GemType, item.Quality,
		weaponJSON, helmJSON, shieldJSON, nullString(item.ImageURL))
	return err
}

// UpdateItemBaseFields updates specific fields on a base item
func (r *Repository) UpdateItemBaseFields(ctx context.Context, id int, item *ItemBase) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE d2.item_bases SET
			code = $2, name = $3, category = $4, item_type = $5,
			level_req = $6, str_req = $7, dex_req = $8,
			min_ac = $9, max_ac = $10, min_dam = $11, max_dam = $12,
			two_hand_min_dam = $13, two_hand_max_dam = $14,
			max_sockets = $15, durability = $16, speed = $17,
			description = $18, image_url = COALESCE($19, image_url),
			updated_at = NOW()
		WHERE id = $1`,
		id, item.Code, item.Name, item.Category, item.ItemType,
		item.LevelReq, item.StrReq, item.DexReq,
		item.MinAC, item.MaxAC, item.MinDam, item.MaxDam,
		item.TwoHandMinDam, item.TwoHandMaxDam,
		item.MaxSockets, item.Durability, item.Speed,
		nullString(item.Description), nullString(item.ImageURL))
	return err
}
