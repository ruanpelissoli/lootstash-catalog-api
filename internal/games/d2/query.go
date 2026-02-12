package d2

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// SearchResult represents a unified search result from any item type
type SearchResult struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`     // "unique", "set", "runeword", "rune", "gem", "base"
	Category string `json:"category"` // Item category: "helm", "armor", etc.
	BaseName string `json:"baseName,omitempty"`
	ImageURL string `json:"imageUrl,omitempty"`
}

// SearchItems searches across all item types by name
func (r *Repository) SearchItems(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Prepare the search pattern for ILIKE
	pattern := "%" + strings.ToLower(query) + "%"

	// Union query across all item types
	sql := `
		WITH all_items AS (
			-- Unique items
			SELECT
				id,
				name,
				'unique' as type,
				COALESCE(
					(SELECT it.name
					 FROM d2.item_types it
					 JOIN d2.item_bases ib ON ib.item_type = it.code
					 WHERE ib.code = unique_items.base_code LIMIT 1),
					'Unknown'
				) as category,
				base_name,
				image_url
			FROM d2.unique_items
			WHERE enabled = true AND LOWER(name) LIKE $1

			UNION ALL

			-- Set items
			SELECT
				id,
				name,
				'set' as type,
				COALESCE(
					(SELECT it.name
					 FROM d2.item_types it
					 JOIN d2.item_bases ib ON ib.item_type = it.code
					 WHERE ib.code = set_items.base_code LIMIT 1),
					'Unknown'
				) as category,
				base_name,
				image_url
			FROM d2.set_items
			WHERE LOWER(name) LIKE $1

			UNION ALL

			-- Runewords
			SELECT
				id,
				display_name as name,
				'runeword' as type,
				'Runeword' as category,
				NULL as base_name,
				image_url
			FROM d2.runewords
			WHERE complete = true AND LOWER(display_name) LIKE $1

			UNION ALL

			-- Runes
			SELECT
				id,
				name,
				'rune' as type,
				'Rune' as category,
				NULL as base_name,
				image_url
			FROM d2.runes
			WHERE LOWER(name) LIKE $1

			UNION ALL

			-- Gems
			SELECT
				id,
				name,
				'gem' as type,
				'Gem' as category,
				NULL as base_name,
				image_url
			FROM d2.gems
			WHERE LOWER(name) LIKE $1

			UNION ALL

			-- Base items (only tradable items)
			SELECT
				id,
				name,
				'base' as type,
				COALESCE(
					(SELECT it.name
					 FROM d2.item_types it
					 WHERE it.code = item_bases.item_type LIMIT 1),
					category
				) as category,
				NULL as base_name,
				image_url
			FROM d2.item_bases
			WHERE spawnable = true AND tradable = true AND LOWER(name) LIKE $1
				AND NOT EXISTS (SELECT 1 FROM d2.gems g WHERE g.code = item_bases.code)
				AND NOT EXISTS (SELECT 1 FROM d2.runes r WHERE r.code = item_bases.code)

			UNION ALL

			-- Quest items
			SELECT
				id,
				name,
				'quest' as type,
				'Quest' as category,
				NULL as base_name,
				image_url
			FROM d2.item_bases
			WHERE quest_item = true AND LOWER(name) LIKE $1
		)
		SELECT id, name, type, category, base_name, image_url
		FROM all_items
		ORDER BY
			CASE
				WHEN LOWER(name) = LOWER($2) THEN 0  -- Exact match first
				WHEN LOWER(name) LIKE LOWER($2) || '%' THEN 1  -- Starts with
				ELSE 2
			END,
			type,
			name
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, sql, pattern, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search items query failed: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var sr SearchResult
		var baseName, imageURL *string
		err := rows.Scan(&sr.ID, &sr.Name, &sr.Type, &sr.Category, &baseName, &imageURL)
		if err != nil {
			return nil, fmt.Errorf("scan search result failed: %w", err)
		}
		if baseName != nil {
			sr.BaseName = *baseName
		}
		if imageURL != nil {
			sr.ImageURL = *imageURL
		}
		results = append(results, sr)
	}

	return results, rows.Err()
}

// GetUniqueItem retrieves a unique item by ID with all its properties
func (r *Repository) GetUniqueItem(ctx context.Context, id int) (*UniqueItem, error) {
	sql := `
		SELECT
			id, index_id, name, base_code, base_name, level, level_req, rarity,
			enabled, ladder_only, first_ladder_season, last_ladder_season,
			properties, inv_transform, chr_transform, inv_file, image_url,
			cost_mult, cost_add, created_at, updated_at
		FROM d2.unique_items
		WHERE id = $1
	`

	var ui UniqueItem
	var baseName, invTransform, chrTransform, invFile, imageURL *string
	var propsJSON []byte

	err := r.pool.QueryRow(ctx, sql, id).Scan(
		&ui.ID, &ui.IndexID, &ui.Name, &ui.BaseCode, &baseName, &ui.Level, &ui.LevelReq, &ui.Rarity,
		&ui.Enabled, &ui.LadderOnly, &ui.FirstLadderSeason, &ui.LastLadderSeason,
		&propsJSON, &invTransform, &chrTransform, &invFile, &imageURL,
		&ui.CostMult, &ui.CostAdd, &ui.CreatedAt, &ui.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get unique item failed: %w", err)
	}

	if baseName != nil {
		ui.BaseName = *baseName
	}
	if invTransform != nil {
		ui.InvTransform = *invTransform
	}
	if chrTransform != nil {
		ui.ChrTransform = *chrTransform
	}
	if invFile != nil {
		ui.InvFile = *invFile
	}
	if imageURL != nil {
		ui.ImageURL = *imageURL
	}

	if len(propsJSON) > 0 {
		if err := json.Unmarshal(propsJSON, &ui.Properties); err != nil {
			return nil, fmt.Errorf("unmarshal properties failed: %w", err)
		}
	}

	return &ui, nil
}

// GetUniqueItemByName retrieves a unique item by name
func (r *Repository) GetUniqueItemByName(ctx context.Context, name string) (*UniqueItem, error) {
	sql := `
		SELECT id FROM d2.unique_items WHERE LOWER(name) = LOWER($1) AND enabled = true LIMIT 1
	`
	var id int
	err := r.pool.QueryRow(ctx, sql, name).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetUniqueItem(ctx, id)
}

// GetSetItem retrieves a set item by ID with all its properties
func (r *Repository) GetSetItem(ctx context.Context, id int) (*SetItem, error) {
	sql := `
		SELECT
			id, index_id, name, set_name, base_code, base_name, level, level_req, rarity,
			properties, bonus_properties, inv_transform, chr_transform, inv_file, image_url,
			cost_mult, cost_add, created_at, updated_at
		FROM d2.set_items
		WHERE id = $1
	`

	var si SetItem
	var baseName, invTransform, chrTransform, invFile, imageURL *string
	var propsJSON, bonusPropsJSON []byte

	err := r.pool.QueryRow(ctx, sql, id).Scan(
		&si.ID, &si.IndexID, &si.Name, &si.SetName, &si.BaseCode, &baseName, &si.Level, &si.LevelReq, &si.Rarity,
		&propsJSON, &bonusPropsJSON, &invTransform, &chrTransform, &invFile, &imageURL,
		&si.CostMult, &si.CostAdd, &si.CreatedAt, &si.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get set item failed: %w", err)
	}

	if baseName != nil {
		si.BaseName = *baseName
	}
	if invTransform != nil {
		si.InvTransform = *invTransform
	}
	if chrTransform != nil {
		si.ChrTransform = *chrTransform
	}
	if invFile != nil {
		si.InvFile = *invFile
	}
	if imageURL != nil {
		si.ImageURL = *imageURL
	}

	if len(propsJSON) > 0 {
		if err := json.Unmarshal(propsJSON, &si.Properties); err != nil {
			return nil, fmt.Errorf("unmarshal properties failed: %w", err)
		}
	}
	if len(bonusPropsJSON) > 0 {
		if err := json.Unmarshal(bonusPropsJSON, &si.BonusProperties); err != nil {
			return nil, fmt.Errorf("unmarshal bonus properties failed: %w", err)
		}
	}

	return &si, nil
}

// GetRuneword retrieves a runeword by ID with all its properties
func (r *Repository) GetRuneword(ctx context.Context, id int) (*Runeword, error) {
	sql := `
		SELECT
			id, name, display_name, complete, ladder_only, first_ladder_season, last_ladder_season,
			valid_item_types, excluded_item_types, runes, properties, image_url,
			created_at, updated_at
		FROM d2.runewords
		WHERE id = $1
	`

	var rw Runeword
	var imageURL *string
	var validTypesJSON, excludedTypesJSON, runesJSON, propsJSON []byte

	err := r.pool.QueryRow(ctx, sql, id).Scan(
		&rw.ID, &rw.Name, &rw.DisplayName, &rw.Complete, &rw.LadderOnly, &rw.FirstLadderSeason, &rw.LastLadderSeason,
		&validTypesJSON, &excludedTypesJSON, &runesJSON, &propsJSON, &imageURL,
		&rw.CreatedAt, &rw.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get runeword failed: %w", err)
	}

	if imageURL != nil {
		rw.ImageURL = *imageURL
	}

	if len(validTypesJSON) > 0 {
		if err := json.Unmarshal(validTypesJSON, &rw.ValidItemTypes); err != nil {
			return nil, fmt.Errorf("unmarshal valid item types failed: %w", err)
		}
	}
	if len(excludedTypesJSON) > 0 {
		if err := json.Unmarshal(excludedTypesJSON, &rw.ExcludedItemTypes); err != nil {
			return nil, fmt.Errorf("unmarshal excluded item types failed: %w", err)
		}
	}
	if len(runesJSON) > 0 {
		if err := json.Unmarshal(runesJSON, &rw.Runes); err != nil {
			return nil, fmt.Errorf("unmarshal runes failed: %w", err)
		}
	}
	if len(propsJSON) > 0 {
		if err := json.Unmarshal(propsJSON, &rw.Properties); err != nil {
			return nil, fmt.Errorf("unmarshal properties failed: %w", err)
		}
	}

	return &rw, nil
}

// GetRunewordByName retrieves a runeword by name
func (r *Repository) GetRunewordByName(ctx context.Context, name string) (*Runeword, error) {
	sql := `
		SELECT id FROM d2.runewords WHERE LOWER(display_name) = LOWER($1) AND complete = true LIMIT 1
	`
	var id int
	err := r.pool.QueryRow(ctx, sql, name).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetRuneword(ctx, id)
}

// GetRune retrieves a rune by ID
func (r *Repository) GetRune(ctx context.Context, id int) (*Rune, error) {
	sql := `
		SELECT
			id, code, name, rune_number, level, level_req,
			weapon_mods, helm_mods, shield_mods,
			inv_file, image_url, cost, created_at, updated_at
		FROM d2.runes
		WHERE id = $1
	`

	var rn Rune
	var invFile, imageURL *string
	var weaponJSON, helmJSON, shieldJSON []byte

	err := r.pool.QueryRow(ctx, sql, id).Scan(
		&rn.ID, &rn.Code, &rn.Name, &rn.RuneNumber, &rn.Level, &rn.LevelReq,
		&weaponJSON, &helmJSON, &shieldJSON,
		&invFile, &imageURL, &rn.Cost, &rn.CreatedAt, &rn.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get rune failed: %w", err)
	}

	if invFile != nil {
		rn.InvFile = *invFile
	}
	if imageURL != nil {
		rn.ImageURL = *imageURL
	}

	if len(weaponJSON) > 0 {
		if err := json.Unmarshal(weaponJSON, &rn.WeaponMods); err != nil {
			return nil, fmt.Errorf("unmarshal weapon mods failed: %w", err)
		}
	}
	if len(helmJSON) > 0 {
		if err := json.Unmarshal(helmJSON, &rn.HelmMods); err != nil {
			return nil, fmt.Errorf("unmarshal helm mods failed: %w", err)
		}
	}
	if len(shieldJSON) > 0 {
		if err := json.Unmarshal(shieldJSON, &rn.ShieldMods); err != nil {
			return nil, fmt.Errorf("unmarshal shield mods failed: %w", err)
		}
	}

	return &rn, nil
}

// GetRuneByName retrieves a rune by name (e.g., "Ber")
func (r *Repository) GetRuneByName(ctx context.Context, name string) (*Rune, error) {
	sql := `SELECT id FROM d2.runes WHERE LOWER(name) = LOWER($1) LIMIT 1`
	var id int
	err := r.pool.QueryRow(ctx, sql, name).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetRune(ctx, id)
}

// GetGem retrieves a gem by ID
func (r *Repository) GetGem(ctx context.Context, id int) (*Gem, error) {
	sql := `
		SELECT
			id, code, name, gem_type, quality,
			weapon_mods, helm_mods, shield_mods,
			transform, inv_file, image_url, created_at, updated_at
		FROM d2.gems
		WHERE id = $1
	`

	var g Gem
	var invFile, imageURL *string
	var weaponJSON, helmJSON, shieldJSON []byte

	err := r.pool.QueryRow(ctx, sql, id).Scan(
		&g.ID, &g.Code, &g.Name, &g.GemType, &g.Quality,
		&weaponJSON, &helmJSON, &shieldJSON,
		&g.Transform, &invFile, &imageURL, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get gem failed: %w", err)
	}

	if invFile != nil {
		g.InvFile = *invFile
	}
	if imageURL != nil {
		g.ImageURL = *imageURL
	}

	if len(weaponJSON) > 0 {
		if err := json.Unmarshal(weaponJSON, &g.WeaponMods); err != nil {
			return nil, fmt.Errorf("unmarshal weapon mods failed: %w", err)
		}
	}
	if len(helmJSON) > 0 {
		if err := json.Unmarshal(helmJSON, &g.HelmMods); err != nil {
			return nil, fmt.Errorf("unmarshal helm mods failed: %w", err)
		}
	}
	if len(shieldJSON) > 0 {
		if err := json.Unmarshal(shieldJSON, &g.ShieldMods); err != nil {
			return nil, fmt.Errorf("unmarshal shield mods failed: %w", err)
		}
	}

	return &g, nil
}

// GetItemBase retrieves a base item by ID
func (r *Repository) GetItemBase(ctx context.Context, id int) (*ItemBase, error) {
	sql := `
		SELECT
			id, code, name, item_type, item_type2, category,
			level, level_req, str_req, dex_req, durability,
			min_ac, max_ac, min_dam, max_dam, two_hand_min_dam, two_hand_max_dam,
			range_adder, speed, str_bonus, dex_bonus,
			max_sockets, gem_apply_type,
			normal_code, exceptional_code, elite_code,
			inv_width, inv_height, inv_file, flippy_file, unique_inv_file, set_inv_file,
			image_url, icon_variants, spawnable, stackable, useable, throwable, quest_item,
			rarity, cost, description, created_at, updated_at
		FROM d2.item_bases
		WHERE id = $1
	`

	var ib ItemBase
	var itemType2, normalCode, exceptionalCode, eliteCode *string
	var invFile, flippyFile, uniqueInvFile, setInvFile, imageURL, description *string

	err := r.pool.QueryRow(ctx, sql, id).Scan(
		&ib.ID, &ib.Code, &ib.Name, &ib.ItemType, &itemType2, &ib.Category,
		&ib.Level, &ib.LevelReq, &ib.StrReq, &ib.DexReq, &ib.Durability,
		&ib.MinAC, &ib.MaxAC, &ib.MinDam, &ib.MaxDam, &ib.TwoHandMinDam, &ib.TwoHandMaxDam,
		&ib.RangeAdder, &ib.Speed, &ib.StrBonus, &ib.DexBonus,
		&ib.MaxSockets, &ib.GemApplyType,
		&normalCode, &exceptionalCode, &eliteCode,
		&ib.InvWidth, &ib.InvHeight, &invFile, &flippyFile, &uniqueInvFile, &setInvFile,
		&imageURL, &ib.IconVariants, &ib.Spawnable, &ib.Stackable, &ib.Useable, &ib.Throwable, &ib.QuestItem,
		&ib.Rarity, &ib.Cost, &description, &ib.CreatedAt, &ib.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get item base failed: %w", err)
	}

	if itemType2 != nil {
		ib.ItemType2 = *itemType2
	}
	if normalCode != nil {
		ib.NormalCode = *normalCode
	}
	if exceptionalCode != nil {
		ib.ExceptionalCode = *exceptionalCode
	}
	if eliteCode != nil {
		ib.EliteCode = *eliteCode
	}
	if invFile != nil {
		ib.InvFile = *invFile
	}
	if flippyFile != nil {
		ib.FlippyFile = *flippyFile
	}
	if uniqueInvFile != nil {
		ib.UniqueInvFile = *uniqueInvFile
	}
	if setInvFile != nil {
		ib.SetInvFile = *setInvFile
	}
	if imageURL != nil {
		ib.ImageURL = *imageURL
	}
	if description != nil {
		ib.Description = *description
	}

	return &ib, nil
}

// GetItemBaseByCode retrieves a base item by code
func (r *Repository) GetItemBaseByCode(ctx context.Context, code string) (*ItemBase, error) {
	sql := `SELECT id FROM d2.item_bases WHERE code = $1 LIMIT 1`
	var id int
	err := r.pool.QueryRow(ctx, sql, code).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetItemBase(ctx, id)
}

// GetItemType retrieves an item type by code
func (r *Repository) GetItemType(ctx context.Context, code string) (*ItemType, error) {
	sql := `
		SELECT
			id, code, name, equiv1, equiv2, body_loc1, body_loc2,
			can_be_magic, can_be_rare, max_sockets_normal, max_sockets_nightmare, max_sockets_hell,
			staff_mods, class_restriction, store_page, created_at, updated_at
		FROM d2.item_types
		WHERE code = $1
	`

	var it ItemType
	var equiv1, equiv2, bodyLoc1, bodyLoc2, staffMods, classRestriction, storePage *string

	err := r.pool.QueryRow(ctx, sql, code).Scan(
		&it.ID, &it.Code, &it.Name, &equiv1, &equiv2, &bodyLoc1, &bodyLoc2,
		&it.CanBeMagic, &it.CanBeRare, &it.MaxSocketsNormal, &it.MaxSocketsNightmare, &it.MaxSocketsHell,
		&staffMods, &classRestriction, &storePage, &it.CreatedAt, &it.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get item type failed: %w", err)
	}

	if equiv1 != nil {
		it.Equiv1 = *equiv1
	}
	if equiv2 != nil {
		it.Equiv2 = *equiv2
	}
	if bodyLoc1 != nil {
		it.BodyLoc1 = *bodyLoc1
	}
	if bodyLoc2 != nil {
		it.BodyLoc2 = *bodyLoc2
	}
	if staffMods != nil {
		it.StaffMods = *staffMods
	}
	if classRestriction != nil {
		it.ClassRestriction = *classRestriction
	}
	if storePage != nil {
		it.StorePage = *storePage
	}

	return &it, nil
}

// GetAllRunes retrieves all runes ordered by rune number
func (r *Repository) GetAllRunes(ctx context.Context) ([]Rune, error) {
	sql := `SELECT id FROM d2.runes ORDER BY rune_number`
	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runes []Rune
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		rn, err := r.GetRune(ctx, id)
		if err != nil {
			return nil, err
		}
		runes = append(runes, *rn)
	}
	return runes, rows.Err()
}

// GetAllGems retrieves all gems ordered by quality and type
func (r *Repository) GetAllGems(ctx context.Context) ([]Gem, error) {
	sql := `
		SELECT id FROM d2.gems
		ORDER BY
			CASE quality
				WHEN 'perfect' THEN 1
				WHEN 'flawless' THEN 2
				WHEN 'normal' THEN 3
				WHEN 'flawed' THEN 4
				WHEN 'chipped' THEN 5
				ELSE 6
			END,
			gem_type
	`
	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gems []Gem
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		g, err := r.GetGem(ctx, id)
		if err != nil {
			return nil, err
		}
		gems = append(gems, *g)
	}
	return gems, rows.Err()
}

// GetAllItemBases retrieves all base items with optional category filter
func (r *Repository) GetAllItemBases(ctx context.Context, category string) ([]ItemBase, error) {
	var rows pgx.Rows
	var err error

	if category != "" {
		rows, err = r.pool.Query(ctx, `SELECT id FROM d2.item_bases WHERE spawnable = true AND category = $1 ORDER BY name`, category)
	} else {
		rows, err = r.pool.Query(ctx, `SELECT id FROM d2.item_bases WHERE spawnable = true ORDER BY category, name`)
	}
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

// GetAllUniqueItems retrieves all unique items
func (r *Repository) GetAllUniqueItems(ctx context.Context) ([]UniqueItem, error) {
	sql := `SELECT id FROM d2.unique_items WHERE enabled = true ORDER BY name`
	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []UniqueItem
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		item, err := r.GetUniqueItem(ctx, id)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

// GetAllSetItems retrieves all set items
func (r *Repository) GetAllSetItems(ctx context.Context) ([]SetItem, error) {
	sql := `SELECT id FROM d2.set_items ORDER BY set_name, name`
	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SetItem
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		item, err := r.GetSetItem(ctx, id)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

// GetAllRunewordsForList retrieves all runewords for listing
func (r *Repository) GetAllRunewordsForList(ctx context.Context) ([]Runeword, error) {
	sql := `SELECT id FROM d2.runewords WHERE complete = true ORDER BY display_name`
	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Runeword
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		item, err := r.GetRuneword(ctx, id)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

// CountSearchResults counts total results for a search query
func (r *Repository) CountSearchResults(ctx context.Context, query string) (int, error) {
	pattern := "%" + strings.ToLower(query) + "%"

	sql := `
		SELECT COUNT(*) FROM (
			SELECT id FROM d2.unique_items WHERE enabled = true AND LOWER(name) LIKE $1
			UNION ALL
			SELECT id FROM d2.set_items WHERE LOWER(name) LIKE $1
			UNION ALL
			SELECT id FROM d2.runewords WHERE complete = true AND LOWER(display_name) LIKE $1
			UNION ALL
			SELECT id FROM d2.runes WHERE LOWER(name) LIKE $1
			UNION ALL
			SELECT id FROM d2.gems WHERE LOWER(name) LIKE $1
			UNION ALL
			SELECT id FROM d2.item_bases WHERE spawnable = true AND tradable = true AND LOWER(name) LIKE $1
				AND NOT EXISTS (SELECT 1 FROM d2.gems g WHERE g.code = item_bases.code)
				AND NOT EXISTS (SELECT 1 FROM d2.runes r WHERE r.code = item_bases.code)
			UNION ALL
			SELECT id FROM d2.item_bases WHERE quest_item = true AND LOWER(name) LIKE $1
		) AS all_items
	`

	var count int
	err := r.pool.QueryRow(ctx, sql, pattern).Scan(&count)
	return count, err
}
