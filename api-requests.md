# LootStash Catalog API - Endpoint Reference

Base URL: `http://localhost:8080`

## Table of Contents

- [Health Check](#health-check)
- [Search](#search)
- [Collection Endpoints](#collection-endpoints)
  - [List All Runes](#list-all-runes)
  - [List All Gems](#list-all-gems)
  - [List All Base Items](#list-all-base-items)
  - [List All Unique Items](#list-all-unique-items)
  - [List All Set Items](#list-all-set-items)
  - [List All Runewords](#list-all-runewords)
- [Item Detail Endpoints](#item-detail-endpoints)
  - [Get Item by Type and ID](#get-item-by-type-and-id)
  - [Get Unique Item](#get-unique-item)
  - [Get Set Item](#get-set-item)
  - [Get Runeword](#get-runeword)
  - [Get Runeword Valid Bases](#get-runeword-valid-bases)
  - [Get Rune](#get-rune)
  - [Get Gem](#get-gem)
  - [Get Base Item](#get-base-item)
- [Reference Data](#reference-data)
  - [List All Stat Codes](#list-all-stat-codes)
  - [List All Categories](#list-all-categories)
  - [List All Rarities](#list-all-rarities)
- [Response Types](#response-types)
- [Error Handling](#error-handling)

---

## Health Check

Check if the API is running.

```
GET /health
```

### Response

```json
{
  "status": "ok",
  "service": "lootstash-catalog-api"
}
```

---

## Search

Search across all item types by name.

```
GET /api/v1/d2/items/search
```

### Query Parameters

| Parameter | Type   | Required | Default | Description                          |
|-----------|--------|----------|---------|--------------------------------------|
| `q`       | string | Yes      | -       | Search query (min 1 character)       |
| `limit`   | number | No       | 20      | Max results to return (1-100)        |

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/items/search?q=shako&limit=10"
```

### Response

```json
{
  "items": [
    {
      "id": "123",
      "name": "Harlequin Crest",
      "type": "unique",
      "category": "helm",
      "imageUrl": "https://...",
      "baseName": "Shako"
    }
  ],
  "totalCount": 1,
  "query": "shako"
}
```

### Response Fields

| Field        | Type   | Description                                      |
|--------------|--------|--------------------------------------------------|
| `items`      | array  | Array of matching items                          |
| `totalCount` | number | Total number of matches (for pagination)         |
| `query`      | string | The search query that was executed               |

### Item Search Result Object

| Field      | Type   | Description                                           |
|------------|--------|-------------------------------------------------------|
| `id`       | string | Item ID (use with detail endpoints)                   |
| `name`     | string | Item name                                             |
| `type`     | string | One of: `unique`, `set`, `runeword`, `rune`, `gem`, `base` |
| `category` | string | Item category (e.g., "helm", "armor", "weapon")       |
| `imageUrl` | string | URL to item image (optional)                          |
| `baseName` | string | Base item name for uniques/sets (optional)            |

---

## Collection Endpoints

### List All Runes

Get all 33 runes ordered by rune number (El through Zod).

```
GET /api/v1/d2/runes
```

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/runes"
```

### Response

```json
[
  {
    "id": 1,
    "code": "r01",
    "name": "El Rune",
    "runeNumber": 1,
    "type": "rune",
    "rarity": "rune",
    "requirements": {
      "level": 11
    },
    "weaponMods": [
      { "name": "+50 To Attack Rating", "code": "att", "hasRange": false }
    ],
    "armorMods": [
      { "name": "+15 Defense", "code": "ac", "hasRange": false }
    ],
    "shieldMods": [
      { "name": "+15 Defense", "code": "ac", "hasRange": false }
    ],
    "imageUrl": "https://..."
  }
]
```

---

### List All Gems

Get all gems ordered by quality (perfect to chipped) and type.

```
GET /api/v1/d2/gems
```

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/gems"
```

### Response

```json
[
  {
    "id": 1,
    "code": "gpw",
    "name": "Perfect Amethyst",
    "gemType": "amethyst",
    "quality": "perfect",
    "type": "gem",
    "rarity": "gem",
    "weaponMods": [
      { "name": "+150 To Attack Rating", "code": "att", "hasRange": false }
    ],
    "armorMods": [
      { "name": "+10 To Strength", "code": "str", "hasRange": false }
    ],
    "shieldMods": [
      { "name": "+10 To Strength", "code": "str", "hasRange": false }
    ],
    "imageUrl": "https://..."
  }
]
```

---

### List All Base Items

Get all base items (normal, exceptional, elite armors/weapons/misc).

```
GET /api/v1/d2/bases
```

### Query Parameters

| Parameter  | Type   | Required | Default | Description                              |
|------------|--------|----------|---------|------------------------------------------|
| `category` | string | No       | -       | Filter by category: `armor`, `weapon`, or `misc` |
| `runeword` | number | No       | -       | Filter by runeword ID to get only valid bases for that runeword |

### Example Requests

```bash
# Get all base items
curl "http://localhost:8080/api/v1/d2/bases"

# Get only armor bases
curl "http://localhost:8080/api/v1/d2/bases?category=armor"

# Get only weapon bases
curl "http://localhost:8080/api/v1/d2/bases?category=weapon"

# Get only misc bases
curl "http://localhost:8080/api/v1/d2/bases?category=misc"

# Get valid bases for a specific runeword (e.g., Spirit)
curl "http://localhost:8080/api/v1/d2/bases?runeword=5"

# Get only weapon bases valid for a runeword
curl "http://localhost:8080/api/v1/d2/bases?runeword=5&category=weapon"
```

### Response

```json
[
  {
    "id": 1,
    "code": "uap",
    "name": "Shako",
    "type": "base",
    "rarity": "normal",
    "category": "armor",
    "itemType": "Helm",
    "requirements": {
      "level": 43,
      "strength": 50,
      "dexterity": 0
    },
    "defense": {
      "min": 98,
      "max": 141
    },
    "maxSockets": 2,
    "durability": 12,
    "qualityTiers": {
      "normal": "cap",
      "exceptional": "skp",
      "elite": "uap"
    },
    "imageUrl": "https://..."
  }
]
```

---

### List All Unique Items

Get all unique items ordered alphabetically.

```
GET /api/v1/d2/uniques
```

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/uniques"
```

### Response

```json
[
  {
    "id": 1,
    "name": "Harlequin Crest",
    "type": "unique",
    "rarity": "unique",
    "base": {
      "code": "uap",
      "name": "Shako",
      "category": "armor",
      "itemType": "helm",
      "defense": 141,
      "maxSockets": 2
    },
    "requirements": {
      "level": 62,
      "strength": 50,
      "dexterity": 0
    },
    "affixes": [
      { "name": "+2 To All Skills", "code": "allskills", "hasRange": false },
      { "name": "+1-148 To Life (+1.5 Per Character Level)", "code": "hp/lvl", "hasRange": true, "minValue": 1, "maxValue": 148 }
    ],
    "ladderOnly": false,
    "imageUrl": "https://..."
  }
]
```

---

### List All Set Items

Get all set items ordered by set name.

```
GET /api/v1/d2/sets
```

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/sets"
```

### Response

```json
[
  {
    "id": 1,
    "name": "Tal Rasha's Horadric Crest",
    "setName": "Tal Rasha's Wrappings",
    "type": "set",
    "rarity": "set",
    "base": {
      "code": "urn",
      "name": "Death Mask",
      "category": "armor",
      "itemType": "helm"
    },
    "requirements": {
      "level": 66,
      "strength": 55,
      "dexterity": 0
    },
    "affixes": [
      { "name": "+10% Life Stolen Per Hit", "code": "lifesteal", "hasRange": false }
    ],
    "bonusAffixes": [
      { "name": "+65 To Life", "code": "hp", "hasRange": false }
    ],
    "imageUrl": "https://..."
  }
]
```

---

### List All Runewords

Get all runewords ordered by display name.

```
GET /api/v1/d2/runewords
```

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/runewords"
```

### Response

```json
[
  {
    "id": 1,
    "name": "Runeword33",
    "displayName": "Enigma",
    "type": "runeword",
    "rarity": "runeword",
    "runes": [
      { "id": 31, "code": "r31", "name": "Jah Rune", "imageUrl": "https://..." },
      { "id": 6, "code": "r06", "name": "Ith Rune", "imageUrl": "https://..." },
      { "id": 30, "code": "r30", "name": "Ber Rune", "imageUrl": "https://..." }
    ],
    "runeOrder": "Jah RuneIth RuneBer Rune",
    "validTypes": [
      { "code": "tors", "name": "Body Armor" }
    ],
    "requirements": {},
    "affixes": [
      { "name": "+2 To All Skills", "code": "allskills", "hasRange": false },
      { "name": "+45% Faster Run/Walk", "code": "frw", "hasRange": false },
      { "name": "+1 To Teleport", "code": "skill", "hasRange": false }
    ],
    "ladderOnly": false,
    "imageUrl": "https://..."
  }
]
```

---

## Item Detail Endpoints

### Get Item by Type and ID

Generic endpoint to get any item by its type and ID.

```
GET /api/v1/d2/items/:type/:id
```

### Path Parameters

| Parameter | Type   | Required | Description                                              |
|-----------|--------|----------|----------------------------------------------------------|
| `type`    | string | Yes      | One of: `unique`, `set`, `runeword`, `rune`, `gem`, `base` |
| `id`      | number | Yes      | Item ID                                                  |

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/items/unique/123"
```

### Response

Returns a `UnifiedItemDetail` object (see [Response Types](#response-types)).

---

### Get Unique Item

```
GET /api/v1/d2/items/unique/:id
```

### Path Parameters

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `id`      | number | Yes      | Unique item ID |

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/items/unique/123"
```

### Response

```json
{
  "itemType": "unique",
  "unique": {
    "id": 123,
    "name": "Harlequin Crest",
    "type": "unique",
    "rarity": "unique",
    "base": {
      "code": "uap",
      "name": "Shako",
      "category": "armor",
      "itemType": "helm",
      "defense": 141,
      "maxSockets": 2
    },
    "requirements": {
      "level": 62,
      "strength": 50,
      "dexterity": 0
    },
    "affixes": [
      { "name": "+2 To All Skills", "code": "allskills", "hasRange": false }
    ],
    "ladderOnly": false,
    "imageUrl": "https://..."
  }
}
```

---

### Get Set Item

```
GET /api/v1/d2/items/set/:id
```

### Path Parameters

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `id`      | number | Yes      | Set item ID |

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/items/set/45"
```

### Response

```json
{
  "itemType": "set",
  "setItem": {
    "id": 45,
    "name": "Tal Rasha's Horadric Crest",
    "setName": "Tal Rasha's Wrappings",
    "type": "set",
    "rarity": "set",
    "base": { ... },
    "requirements": { ... },
    "affixes": [ ... ],
    "bonusAffixes": [ ... ],
    "imageUrl": "https://..."
  }
}
```

---

### Get Runeword

```
GET /api/v1/d2/items/runeword/:id
```

### Path Parameters

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `id`      | number | Yes      | Runeword ID |

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/items/runeword/33"
```

### Response

```json
{
  "itemType": "runeword",
  "runeword": {
    "id": 33,
    "name": "Runeword33",
    "displayName": "Enigma",
    "type": "runeword",
    "rarity": "runeword",
    "runes": [
      { "id": 31, "code": "r31", "name": "Jah Rune", "imageUrl": "https://..." },
      { "id": 6, "code": "r06", "name": "Ith Rune", "imageUrl": "https://..." },
      { "id": 30, "code": "r30", "name": "Ber Rune", "imageUrl": "https://..." }
    ],
    "runeOrder": "Jah RuneIth RuneBer Rune",
    "validTypes": [
      { "code": "tors", "name": "Body Armor" }
    ],
    "validBaseItems": [
      { "id": 123, "code": "qui", "name": "Quilted Armor", "category": "armor", "maxSockets": 4 },
      { "id": 124, "code": "lea", "name": "Leather Armor", "category": "armor", "maxSockets": 4 },
      { "id": 125, "code": "hla", "name": "Hard Leather Armor", "category": "armor", "maxSockets": 4 }
    ],
    "requirements": {},
    "affixes": [ ... ],
    "ladderOnly": false,
    "imageUrl": "https://..."
  }
}
```

---

### Get Runeword Valid Bases

Get all valid base items for a specific runeword.

```
GET /api/v1/d2/items/runeword/:id/bases
```

### Path Parameters

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `id`      | number | Yes      | Runeword ID |

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/items/runeword/33/bases"
```

### Response

```json
[
  { "id": 123, "code": "qui", "name": "Quilted Armor", "category": "armor", "maxSockets": 4 },
  { "id": 124, "code": "lea", "name": "Leather Armor", "category": "armor", "maxSockets": 4 },
  { "id": 125, "code": "hla", "name": "Hard Leather Armor", "category": "armor", "maxSockets": 4 },
  { "id": 126, "code": "stu", "name": "Studded Leather", "category": "armor", "maxSockets": 4 }
]
```

### RunewordBaseItem Object

| Field       | Type   | Description                     |
|-------------|--------|---------------------------------|
| `id`        | number | Base item ID                    |
| `code`      | string | Base item code                  |
| `name`      | string | Base item name                  |
| `category`  | string | `armor`, `weapon`, or `misc`    |
| `maxSockets`| number | Maximum sockets for this base   |

---

### Get Rune

```
GET /api/v1/d2/items/rune/:id
```

### Path Parameters

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `id`      | number | Yes      | Rune ID     |

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/items/rune/30"
```

### Response

```json
{
  "itemType": "rune",
  "rune": {
    "id": 30,
    "code": "r30",
    "name": "Ber Rune",
    "runeNumber": 30,
    "type": "rune",
    "rarity": "rune",
    "requirements": {
      "level": 63
    },
    "weaponMods": [ ... ],
    "armorMods": [ ... ],
    "shieldMods": [ ... ],
    "imageUrl": "https://..."
  }
}
```

---

### Get Gem

```
GET /api/v1/d2/items/gem/:id
```

### Path Parameters

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `id`      | number | Yes      | Gem ID      |

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/items/gem/5"
```

### Response

```json
{
  "itemType": "gem",
  "gem": {
    "id": 5,
    "code": "gpw",
    "name": "Perfect Amethyst",
    "gemType": "amethyst",
    "quality": "perfect",
    "type": "gem",
    "rarity": "gem",
    "weaponMods": [ ... ],
    "armorMods": [ ... ],
    "shieldMods": [ ... ],
    "imageUrl": "https://..."
  }
}
```

---

### Get Base Item

```
GET /api/v1/d2/items/base/:id
```

### Path Parameters

| Parameter | Type   | Required | Description   |
|-----------|--------|----------|---------------|
| `id`      | number | Yes      | Base item ID  |

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/items/base/100"
```

### Response

```json
{
  "itemType": "base",
  "base": {
    "id": 100,
    "code": "uap",
    "name": "Shako",
    "type": "base",
    "rarity": "normal",
    "category": "armor",
    "itemType": "Helm",
    "requirements": {
      "level": 43,
      "strength": 50,
      "dexterity": 0
    },
    "defense": {
      "min": 98,
      "max": 141
    },
    "maxSockets": 2,
    "durability": 12,
    "qualityTiers": {
      "normal": "cap",
      "exceptional": "skp",
      "elite": "uap"
    },
    "imageUrl": "https://..."
  }
}
```

---

## Response Types

### UnifiedItemDetail

Wrapper object returned by all detail endpoints. Only one of the item-specific fields will be populated based on `itemType`.

```typescript
interface UnifiedItemDetail {
  itemType: "unique" | "set" | "runeword" | "rune" | "gem" | "base";
  unique?: UniqueItemDetail;
  setItem?: SetItemDetail;
  runeword?: RunewordDetail;
  rune?: RuneDetail;
  gem?: GemDetail;
  base?: BaseItemDetail;
}
```

### ItemAffix

Represents a single item property/affix.

```typescript
interface ItemAffix {
  name: string;        // Human-readable text: "+2 To All Skills"
  description?: string;
  code: string;        // Internal code for filtering
  hasRange: boolean;   // true if min != max
  minValue?: number;   // Only present if hasRange is true
  maxValue?: number;   // Only present if hasRange is true
}
```

### ItemRequirements

```typescript
interface ItemRequirements {
  level: number;
  strength?: number;
  dexterity?: number;
}
```

### ItemBaseInfo

```typescript
interface ItemBaseInfo {
  code: string;
  name: string;
  category: "armor" | "weapon" | "misc";
  itemType: string;
  defense?: number;
  minDamage?: number;
  maxDamage?: number;
  maxSockets?: number;
}
```

### DefenseRange

```typescript
interface DefenseRange {
  min: number;
  max: number;
}
```

### DamageRange

```typescript
interface DamageRange {
  oneHandMin?: number;
  oneHandMax?: number;
  twoHandMin?: number;
  twoHandMax?: number;
}
```

### QualityTiers

```typescript
interface QualityTiers {
  normal?: string;      // Item code for normal version
  exceptional?: string; // Item code for exceptional version
  elite?: string;       // Item code for elite version
}
```

### RunewordBaseItem

Represents a valid base item for a runeword.

```typescript
interface RunewordBaseItem {
  id: number;           // Base item ID
  code: string;         // Base item code
  name: string;         // Base item name
  category: string;     // "armor", "weapon", or "misc"
  maxSockets: number;   // Maximum sockets for this base
}
```

### RunewordRune

Represents a rune in a runeword with display info.

```typescript
interface RunewordRune {
  id: number;           // Rune ID
  code: string;         // Rune code (e.g., "r31")
  name: string;         // Rune name (e.g., "Jah Rune")
  imageUrl?: string;    // URL to rune icon
}
```

### RunewordValidType

Represents a valid item type for a runeword.

```typescript
interface RunewordValidType {
  code: string;         // Type code (e.g., "tors")
  name: string;         // Type name (e.g., "Body Armor")
}
```

---

## Error Handling

All errors return a consistent JSON structure:

```json
{
  "error": "error_type",
  "message": "Human-readable error message",
  "code": 400
}
```

### Error Types

| HTTP Code | Error Type       | Description                           |
|-----------|------------------|---------------------------------------|
| 400       | `bad_request`    | Invalid parameters or missing required fields |
| 404       | `not_found`      | Item not found                        |
| 500       | `internal_error` | Server error                          |

### Example Error Response

```json
{
  "error": "bad_request",
  "message": "Query parameter 'q' is required",
  "code": 400
}
```

---

## Headers

### Request Headers

| Header         | Required | Description                    |
|----------------|----------|--------------------------------|
| `Content-Type` | No       | Not required for GET requests  |

### Response Headers

| Header         | Value              |
|----------------|--------------------|
| `Content-Type` | `application/json` |

### CORS

The API supports CORS with the following configuration:

- **Allowed Origins**: `*` (configurable)
- **Allowed Methods**: `GET, POST, PUT, DELETE, OPTIONS`
- **Allowed Headers**: `Origin, Content-Type, Accept, Authorization`
- **Credentials**: Allowed

---

## Reference Data

These endpoints provide reference data for building marketplace filter UIs.

### List All Stat Codes

Get all filterable stat codes for marketplace item filtering. This endpoint returns the canonical stat codes with their display names, descriptions, categories, and any aliases.

```
GET /api/v1/d2/stats
```

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/stats"
```

### Response

```json
[
  {
    "code": "mf",
    "name": "Magic Find",
    "description": "+{value}% Better Chance Of Getting Magic Items",
    "category": "Magic Find",
    "aliases": ["mag%"],
    "isVariable": true
  },
  {
    "code": "fcr",
    "name": "Faster Cast Rate",
    "description": "+{value}% Faster Cast Rate",
    "category": "Speed",
    "aliases": ["cast1", "cast2", "cast3"],
    "isVariable": true
  },
  {
    "code": "allskills",
    "name": "All Skills",
    "description": "+{value} To All Skills",
    "category": "Skills",
    "isVariable": true
  },
  {
    "code": "nofreeze",
    "name": "Cannot Be Frozen",
    "description": "Cannot Be Frozen",
    "category": "Other",
    "isVariable": false
  }
]
```

### StatCode Object

| Field         | Type     | Description                                                      |
|---------------|----------|------------------------------------------------------------------|
| `code`        | string   | Primary code for filtering (use this in marketplace filters)    |
| `name`        | string   | Short display name for UI                                        |
| `description` | string   | Format template showing what the stat looks like on items        |
| `category`    | string   | Category for grouping in filter UI                               |
| `aliases`     | string[] | Alternative codes that map to this stat (from game data)         |
| `isVariable`  | boolean  | Whether this stat typically has variable rolls on items          |

### Categories

Stats are grouped into these categories for UI organization:

- **Skills** - Skill bonuses (+All Skills, +Class Skills)
- **Attributes** - Strength, Dexterity, Vitality, Energy
- **Life & Mana** - Life, Mana, regeneration
- **Speed** - FCR, IAS, FRW, FHR
- **Resistances** - Fire/Cold/Lightning/Poison/All resist
- **Absorb** - Elemental absorb
- **Damage** - Enhanced damage, elemental damage bonuses
- **Attack** - Attack rating, ignore defense
- **Defense** - Defense, damage reduction
- **Leech** - Life/mana steal, life/mana per kill
- **Combat** - Crushing blow, deadly strike, open wounds
- **Magic Find** - MF, GF
- **Pierce** - Enemy resistance reduction
- **Other** - Sockets, cannot be frozen, ethereal, etc.

### Aliases

Some stats have aliases because game data uses different code conventions than the user-friendly codes. For example:

- `mf` → `mag%` (game data uses `mag%`)
- `fcr` → `cast1`, `cast2`, `cast3` (game data uses numbered variants)
- `fire_res` → `res-fire` (game data uses hyphenated format)

The marketplace API's filter system accepts either the canonical code or any of its aliases, so clients can use whichever is convenient.

---

### List All Categories

Get all item categories for marketplace filtering.

```
GET /api/v1/d2/categories
```

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/categories"
```

### Response

```json
[
  {
    "code": "helm",
    "name": "Helms",
    "description": "Head armor including circlets, crowns, and helmets"
  },
  {
    "code": "armor",
    "name": "Body Armor",
    "description": "Chest armor including robes, plate, and leather"
  },
  {
    "code": "weapon",
    "name": "Weapons",
    "description": "All weapon types including swords, axes, bows, and staves"
  },
  {
    "code": "shield",
    "name": "Shields",
    "description": "Shields and paladin-specific shields"
  },
  {
    "code": "gloves",
    "name": "Gloves",
    "description": "Hand armor including gauntlets and bracers"
  },
  {
    "code": "boots",
    "name": "Boots",
    "description": "Foot armor including greaves and boots"
  },
  {
    "code": "belt",
    "name": "Belts",
    "description": "Waist armor including sashes and belts"
  },
  {
    "code": "amulet",
    "name": "Amulets",
    "description": "Neck jewelry"
  },
  {
    "code": "ring",
    "name": "Rings",
    "description": "Finger jewelry"
  },
  {
    "code": "charm",
    "name": "Charms",
    "description": "Inventory charms (small, large, grand)"
  },
  {
    "code": "jewel",
    "name": "Jewels",
    "description": "Socketable jewels with random magical properties"
  },
  {
    "code": "rune",
    "name": "Runes",
    "description": "Socketable runes used to create runewords"
  },
  {
    "code": "gem",
    "name": "Gems",
    "description": "Socketable gems from chipped to perfect quality"
  },
  {
    "code": "misc",
    "name": "Miscellaneous",
    "description": "Keys, organs, tokens, and other items"
  }
]
```

### Category Object

| Field         | Type   | Description                                      |
|---------------|--------|--------------------------------------------------|
| `code`        | string | Internal code for filtering                      |
| `name`        | string | Display name for UI                              |
| `description` | string | Brief description of items in this category      |

---

### List All Rarities

Get all item rarities for marketplace filtering.

```
GET /api/v1/d2/rarities
```

### Example Request

```bash
curl "http://localhost:8080/api/v1/d2/rarities"
```

### Response

```json
[
  {
    "code": "normal",
    "name": "Normal",
    "color": "#FFFFFF",
    "description": "White items with no magical properties"
  },
  {
    "code": "magic",
    "name": "Magic",
    "color": "#4169E1",
    "description": "Blue items with 1-2 magical affixes"
  },
  {
    "code": "rare",
    "name": "Rare",
    "color": "#FFFF00",
    "description": "Yellow items with 2-6 magical affixes"
  },
  {
    "code": "unique",
    "name": "Unique",
    "color": "#C4A000",
    "description": "Gold/tan items with fixed properties"
  },
  {
    "code": "set",
    "name": "Set",
    "color": "#00FF00",
    "description": "Green items that grant bonuses when worn together"
  },
  {
    "code": "runeword",
    "name": "Runeword",
    "color": "#C4A000",
    "description": "Items created by socketing specific runes in order"
  },
  {
    "code": "crafted",
    "name": "Crafted",
    "color": "#FFA500",
    "description": "Orange items created via Horadric Cube recipes"
  }
]
```

### Rarity Object

| Field         | Type   | Description                                      |
|---------------|--------|--------------------------------------------------|
| `code`        | string | Internal code for filtering                      |
| `name`        | string | Display name for UI                              |
| `color`       | string | Hex color code for UI display (matches in-game)  |
| `description` | string | Brief description of this rarity type            |

---

## Quick Reference

| Method | Endpoint                              | Description                          |
|--------|---------------------------------------|--------------------------------------|
| GET    | `/health`                             | Health check                         |
| GET    | `/api/v1/d2/items/search`             | Search all items                     |
| GET    | `/api/v1/d2/stats`                    | List all filterable stat codes       |
| GET    | `/api/v1/d2/categories`               | List all item categories             |
| GET    | `/api/v1/d2/rarities`                 | List all item rarities               |
| GET    | `/api/v1/d2/runes`                    | List all runes                       |
| GET    | `/api/v1/d2/gems`                     | List all gems                        |
| GET    | `/api/v1/d2/bases`                    | List all base items                  |
| GET    | `/api/v1/d2/bases?runeword=:id`       | List bases valid for a runeword      |
| GET    | `/api/v1/d2/uniques`                  | List all unique items                |
| GET    | `/api/v1/d2/sets`                     | List all set items                   |
| GET    | `/api/v1/d2/runewords`                | List all runewords                   |
| GET    | `/api/v1/d2/items/:type/:id`          | Get item by type and ID              |
| GET    | `/api/v1/d2/items/unique/:id`         | Get unique item detail               |
| GET    | `/api/v1/d2/items/set/:id`            | Get set item detail                  |
| GET    | `/api/v1/d2/items/runeword/:id`       | Get runeword detail                  |
| GET    | `/api/v1/d2/items/runeword/:id/bases` | Get valid bases for a runeword       |
| GET    | `/api/v1/d2/items/rune/:id`           | Get rune detail                      |
| GET    | `/api/v1/d2/items/gem/:id`            | Get gem detail                       |
| GET    | `/api/v1/d2/items/base/:id`           | Get base item detail                 |
