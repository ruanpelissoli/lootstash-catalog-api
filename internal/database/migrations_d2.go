package database

import (
	"context"
	"fmt"
)

const d2MigrationSQL = `
-- Create d2 schema for Diablo II catalog
CREATE SCHEMA IF NOT EXISTS d2;

-- Item Types (taxonomy/categories)
CREATE TABLE IF NOT EXISTS d2.item_types (
    id SERIAL PRIMARY KEY,
    code VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    equiv1 VARCHAR(10),
    equiv2 VARCHAR(10),
    body_loc1 VARCHAR(10),
    body_loc2 VARCHAR(10),
    can_be_magic BOOLEAN DEFAULT TRUE,
    can_be_rare BOOLEAN DEFAULT TRUE,
    max_sockets_normal INT DEFAULT 0,
    max_sockets_nightmare INT DEFAULT 0,
    max_sockets_hell INT DEFAULT 0,
    staff_mods VARCHAR(10),
    class_restriction VARCHAR(10),
    store_page VARCHAR(10),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Item Bases (armor, weapons, misc)
CREATE TABLE IF NOT EXISTS d2.item_bases (
    id SERIAL PRIMARY KEY,
    code VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    item_type VARCHAR(10) NOT NULL,
    item_type2 VARCHAR(10),
    category VARCHAR(20) NOT NULL, -- 'armor', 'weapon', 'misc'

    -- Stats
    level INT DEFAULT 0,
    level_req INT DEFAULT 0,
    str_req INT DEFAULT 0,
    dex_req INT DEFAULT 0,
    durability INT DEFAULT 0,

    -- Armor specific
    min_ac INT DEFAULT 0,
    max_ac INT DEFAULT 0,

    -- Weapon specific
    min_dam INT DEFAULT 0,
    max_dam INT DEFAULT 0,
    two_hand_min_dam INT DEFAULT 0,
    two_hand_max_dam INT DEFAULT 0,
    range_adder INT DEFAULT 0,
    speed INT DEFAULT 0,
    str_bonus INT DEFAULT 0,
    dex_bonus INT DEFAULT 0,

    -- Sockets
    max_sockets INT DEFAULT 0,
    gem_apply_type INT DEFAULT 0,

    -- Quality tiers
    normal_code VARCHAR(10),
    exceptional_code VARCHAR(10),
    elite_code VARCHAR(10),

    -- Inventory
    inv_width INT DEFAULT 1,
    inv_height INT DEFAULT 1,

    -- Graphics
    inv_file VARCHAR(50),
    flippy_file VARCHAR(50),
    unique_inv_file VARCHAR(50),
    set_inv_file VARCHAR(50),

    -- Image URL (for Supabase storage)
    image_url TEXT,

    -- Flags
    spawnable BOOLEAN DEFAULT TRUE,
    stackable BOOLEAN DEFAULT FALSE,
    useable BOOLEAN DEFAULT FALSE,
    throwable BOOLEAN DEFAULT FALSE,
    quest_item BOOLEAN DEFAULT FALSE,

    rarity INT DEFAULT 1,
    cost INT DEFAULT 0,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Unique Items
CREATE TABLE IF NOT EXISTS d2.unique_items (
    id SERIAL PRIMARY KEY,
    index_id INT UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    base_code VARCHAR(10) NOT NULL,
    base_name VARCHAR(100),

    level INT DEFAULT 0,
    level_req INT DEFAULT 0,
    rarity INT DEFAULT 1,

    enabled BOOLEAN DEFAULT TRUE,
    ladder_only BOOLEAN DEFAULT FALSE,
    first_ladder_season INT,
    last_ladder_season INT,

    -- Properties stored as JSONB for flexibility
    properties JSONB DEFAULT '[]'::jsonb,

    -- Graphics
    inv_transform VARCHAR(10),
    chr_transform VARCHAR(10),
    inv_file VARCHAR(50),

    -- Image URL
    image_url TEXT,

    cost_mult INT DEFAULT 0,
    cost_add INT DEFAULT 0,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Set Bonuses (Set definitions)
CREATE TABLE IF NOT EXISTS d2.set_bonuses (
    id SERIAL PRIMARY KEY,
    index_id INT UNIQUE NOT NULL,
    name VARCHAR(100) UNIQUE NOT NULL,
    version INT DEFAULT 0,

    -- Partial bonuses (2-4 items)
    partial_bonuses JSONB DEFAULT '[]'::jsonb,

    -- Full set bonuses
    full_bonuses JSONB DEFAULT '[]'::jsonb,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Set Items
CREATE TABLE IF NOT EXISTS d2.set_items (
    id SERIAL PRIMARY KEY,
    index_id INT UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    set_name VARCHAR(100) NOT NULL,
    base_code VARCHAR(10) NOT NULL,
    base_name VARCHAR(100),

    level INT DEFAULT 0,
    level_req INT DEFAULT 0,
    rarity INT DEFAULT 1,

    -- Base properties (always active)
    properties JSONB DEFAULT '[]'::jsonb,

    -- Bonus properties (activated by wearing more set items)
    bonus_properties JSONB DEFAULT '[]'::jsonb,

    -- Graphics
    inv_transform VARCHAR(10),
    chr_transform VARCHAR(10),
    inv_file VARCHAR(50),

    -- Image URL
    image_url TEXT,

    cost_mult INT DEFAULT 0,
    cost_add INT DEFAULT 0,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Runewords
CREATE TABLE IF NOT EXISTS d2.runewords (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,

    complete BOOLEAN DEFAULT FALSE,
    ladder_only BOOLEAN DEFAULT FALSE,
    first_ladder_season INT,
    last_ladder_season INT,

    -- Valid item types (can be applied to)
    valid_item_types JSONB DEFAULT '[]'::jsonb,

    -- Excluded item types
    excluded_item_types JSONB DEFAULT '[]'::jsonb,

    -- Runes required (in order)
    runes JSONB DEFAULT '[]'::jsonb,

    -- Properties granted
    properties JSONB DEFAULT '[]'::jsonb,

    -- Image URL
    image_url TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Runes (individual rune items)
CREATE TABLE IF NOT EXISTS d2.runes (
    id SERIAL PRIMARY KEY,
    code VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(50) NOT NULL,
    rune_number INT NOT NULL,

    level INT DEFAULT 0,
    level_req INT DEFAULT 0,

    -- Mods when socketed in different item types
    weapon_mods JSONB DEFAULT '[]'::jsonb,
    helm_mods JSONB DEFAULT '[]'::jsonb,
    shield_mods JSONB DEFAULT '[]'::jsonb,

    -- Graphics
    inv_file VARCHAR(50),

    -- Image URL
    image_url TEXT,

    cost INT DEFAULT 0,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Gems
CREATE TABLE IF NOT EXISTS d2.gems (
    id SERIAL PRIMARY KEY,
    code VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(50) NOT NULL,
    gem_type VARCHAR(20) NOT NULL, -- amethyst, sapphire, emerald, ruby, diamond, topaz, skull
    quality VARCHAR(20) NOT NULL,  -- chipped, flawed, normal, flawless, perfect

    -- Mods when socketed in different item types
    weapon_mods JSONB DEFAULT '[]'::jsonb,
    helm_mods JSONB DEFAULT '[]'::jsonb,
    shield_mods JSONB DEFAULT '[]'::jsonb,

    -- Graphics
    transform INT DEFAULT 0,
    inv_file VARCHAR(50),

    -- Image URL
    image_url TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Stats (dynamic stat code registry)
CREATE TABLE IF NOT EXISTS d2.stats (
    id SERIAL PRIMARY KEY,
    code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_text VARCHAR(500) NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'Other',
    is_variable BOOLEAN DEFAULT TRUE,
    is_parametric BOOLEAN DEFAULT FALSE,
    aliases TEXT[] DEFAULT '{}',
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_item_bases_type ON d2.item_bases(item_type);
CREATE INDEX IF NOT EXISTS idx_item_bases_category ON d2.item_bases(category);
CREATE INDEX IF NOT EXISTS idx_item_bases_code ON d2.item_bases(code);
CREATE INDEX IF NOT EXISTS idx_unique_items_base_code ON d2.unique_items(base_code);
CREATE INDEX IF NOT EXISTS idx_set_items_set_name ON d2.set_items(set_name);
CREATE INDEX IF NOT EXISTS idx_set_items_base_code ON d2.set_items(base_code);
-- GIN indexes for JSONB property searches
CREATE INDEX IF NOT EXISTS idx_unique_items_properties ON d2.unique_items USING GIN (properties);
CREATE INDEX IF NOT EXISTS idx_set_items_properties ON d2.set_items USING GIN (properties);
CREATE INDEX IF NOT EXISTS idx_runewords_properties ON d2.runewords USING GIN (properties);

-- Stats indexes
CREATE INDEX IF NOT EXISTS idx_stats_code ON d2.stats(code);
CREATE INDEX IF NOT EXISTS idx_stats_category ON d2.stats(category);

-- Profiles for auth (UUID matches Supabase auth.users.id)
CREATE TABLE IF NOT EXISTS d2.profiles (
    id UUID PRIMARY KEY,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Classes with skill trees
CREATE TABLE IF NOT EXISTS d2.classes (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    skill_suffix VARCHAR(100) NOT NULL DEFAULT '',
    skill_trees JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Description column for quest items
ALTER TABLE d2.item_bases ADD COLUMN IF NOT EXISTS description TEXT;

-- Index for quest item lookups
CREATE INDEX IF NOT EXISTS idx_item_bases_quest ON d2.item_bases(quest_item) WHERE quest_item = true;

-- V2: New columns on item_bases
ALTER TABLE d2.item_bases ADD COLUMN IF NOT EXISTS tier VARCHAR(20) DEFAULT 'Normal';
ALTER TABLE d2.item_bases ADD COLUMN IF NOT EXISTS type_tags TEXT[] DEFAULT '{}';
ALTER TABLE d2.item_bases ADD COLUMN IF NOT EXISTS class_specific VARCHAR(20);
ALTER TABLE d2.item_bases ADD COLUMN IF NOT EXISTS tradable BOOLEAN DEFAULT TRUE;
ALTER TABLE d2.item_bases ADD COLUMN IF NOT EXISTS icon_variants TEXT[] DEFAULT '{}';
CREATE INDEX IF NOT EXISTS idx_item_bases_tier ON d2.item_bases(tier);
CREATE INDEX IF NOT EXISTS idx_item_bases_type_tags ON d2.item_bases USING GIN (type_tags);

-- V2: UNIQUE constraints for name-based upserts
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_items_name ON d2.unique_items(name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_set_items_name ON d2.set_items(name);

-- V2: Drop legacy TSV-only tables
DROP TABLE IF EXISTS d2.treasure_class_items;
DROP TABLE IF EXISTS d2.treasure_classes;
DROP TABLE IF EXISTS d2.item_ratios;
DROP TABLE IF EXISTS d2.properties;
DROP TABLE IF EXISTS d2.affixes;
`

func (db *DB) MigrateD2(ctx context.Context) error {
	_, err := db.pool.Exec(ctx, d2MigrationSQL)
	if err != nil {
		return fmt.Errorf("failed to execute D2 migration: %w", err)
	}
	return nil
}
