# Frontend Migration Guide - V2 API Changes

## Overview

The backend import pipeline has been overhauled from a dual-source (TSV + HTML) approach to a single HTML-only import. This results in better data quality (more complete stats, correct variable ranges, proper icons) but introduces several breaking changes to the API response shapes.

**Impact level: MEDIUM** - Most changes are additive (new fields), but runeword responses have breaking field value changes.

---

## 1. BREAKING: Runeword `name` Field Changed

### What changed
The internal `name` field format changed from sequential numbering to a derived name.

### Before
```json
{ "name": "Runeword33", "displayName": "Enigma" }
```

### After
```json
{ "name": "HTMLRuneword_Enigma", "displayName": "Enigma" }
```

### Action required
- If you are using `name` for display, **switch to `displayName`** (this was always the correct field for display).
- If you are using `name` as a unique key/identifier, update your code to handle the new format. The `id` (integer) field is stable and should be preferred for linking/routing.
- If you are constructing URLs or routes using `name`, switch to `id` or `displayName`.

---

## 2. BREAKING: Runeword `validTypes` Values Changed

### What changed
The `validTypes` array items now contain human-readable type names instead of internal type codes. Both `code` and `name` contain the same human-readable value.

### Before
```json
"validTypes": [
  { "code": "tors", "name": "Body Armor" },
  { "code": "swor", "name": "Swords" }
]
```

### After
```json
"validTypes": [
  { "code": "Body Armor", "name": "Body Armor" },
  { "code": "Swords", "name": "Swords" }
]
```

### Action required
- If you display `validTypes[].name` to the user: **No change needed** - this still works.
- If you use `validTypes[].code` for filtering/matching/linking to other endpoints: **Update your logic**. The code is now a human-readable name, not an internal code. If you need to match against `itemType` codes, you'll need to use the name values directly or map them yourself.
- If you display the `code` field anywhere, it will now show readable names like "Body Armor" instead of "tors", which may actually be an improvement.

---

## 3. BREAKING: Runeword Rune Names Shortened

### What changed
Rune names in the `runes` array within runeword responses no longer include the " Rune" suffix.

### Before
```json
"runes": [
  { "id": 31, "code": "r31", "name": "Jah Rune" },
  { "id": 30, "code": "r30", "name": "Ber Rune" }
],
"runeOrder": "Jah RuneIth RuneBer Rune"
```

### After
```json
"runes": [
  { "id": 31, "code": "r31", "name": "Jah" },
  { "id": 30, "code": "r30", "name": "Ber" }
],
"runeOrder": "JahIthBer"
```

### Action required
- If you display rune names in the runeword context (e.g., "Jah + Ith + Ber"): **No change needed** if you just display `name`.
- If you append " Rune" yourself or expect the suffix in comparisons: **Remove that logic**.
- If you parse `runeOrder` for display: The format changed from "Jah RuneIth RuneBer Rune" (with spaces and suffix, hard to parse) to "JahIthBer" (concatenated short names). You may want to build the display string from the `runes` array instead: `runes.map(r => r.name).join(" + ")`.
- **Note**: The standalone rune endpoints (`/runes`, `/items/rune/:id`) still return full names with " Rune" suffix (e.g., "El Rune", "Ber Rune"). Only the `runes` array within runeword responses uses short names.

---

## 4. ADDITIVE: Base Item New Fields

### What changed
Base items now include `tier`, `typeTags`, `classSpecific`, and `iconVariants` fields.

### New fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `tier` | string? | Item quality tier | `"Normal"`, `"Exceptional"`, `"Elite"` |
| `typeTags` | string[]? | Full type hierarchy tags | `["Helms"]`, `["Swords", "Melee Weapons"]` |
| `classSpecific` | string? | Class restriction | `"amazon"`, `"paladin"`, `null` |
| `iconVariants` | string[]? | Alternate icon image URLs | `["https://.../charm_small.png", ...]` |

### Example response
```json
{
  "id": 1,
  "code": "uap",
  "name": "Shako",
  "type": "Base",
  "category": "Armor",
  "itemType": "Helm",
  "tier": "Elite",
  "typeTags": ["Helms"],
  "classSpecific": null,
  "iconVariants": [],
  ...
}
```

### Action required
- **No breaking changes** - all new fields are optional and nullable.
- **Recommended**: Use `tier` for filtering/display of Normal/Exceptional/Elite.
- **Recommended**: Use `typeTags` for category-based filtering if needed.
- **Recommended**: Use `classSpecific` to show class restriction badges/labels on items (e.g., "Amazon Only").
- **Recommended**: Use `iconVariants` for items like charms/jewels that have multiple visual appearances.

---

## 5. ADDITIVE: Stats Endpoint Now Dynamic

### What changed
`GET /api/v1/d2/stats` now returns stats from a database table instead of a hardcoded list. The response shape is identical, but the list may contain more stat codes than before.

### Before
~60 hardcoded stats returned.

### After
Dynamic discovery - any stat code found during item import is registered. May return 100+ stats.

### Action required
- **No breaking changes** - response shape is identical.
- If you hardcode stat categories or assume a fixed number of stats, **update to handle dynamic lists**.
- New stat categories may appear. Ensure your UI can handle unknown categories gracefully (e.g., display them in an "Other" group).

---

## 6. Data Quality Improvements (No Code Changes Needed)

These are improvements in data quality that don't require code changes but may affect what users see:

- **More complete item properties**: Items now have more accurate stat text extracted from HTML source (diablo2.io) instead of raw TSV codes. The `affixes[].name` field will contain properly formatted text like "+2 To All Skills" instead of raw code translations.
- **Better image coverage**: More items will have `imageUrl` populated.
- **More accurate variable stats**: `hasRange`, `minValue`, `maxValue` are now more accurate.
- **IDs may change**: Because the import pipeline was rebuilt, item IDs (auto-increment) will be different after a fresh seed. If you have bookmarked/cached item IDs, they will be invalid. Use item names for stable references.

---

## Quick Checklist

- [ ] **Runeword display**: Use `displayName` instead of `name` for user-facing display
- [ ] **Runeword routing**: Use `id` (integer) for URL routing, not `name`
- [ ] **Runeword valid types**: If using `validTypes[].code` for logic, update to handle human-readable names
- [ ] **Rune names in runewords**: Remove any " Rune" suffix handling, names are now short (e.g., "Jah")
- [ ] **`runeOrder` display**: Build from `runes` array instead of parsing `runeOrder` string
- [ ] **TypeScript interfaces**: Add optional `tier`, `typeTags`, `classSpecific`, `iconVariants` to BaseItemDetail
- [ ] **Stats filter UI**: Handle dynamic stat list (more codes, potentially new categories)
- [ ] **Cached IDs**: Clear any cached item IDs - they will change after re-seed

---

## TypeScript Interface Updates

Add these to your frontend types:

```typescript
// Updated BaseItemDetail
interface BaseItemDetail {
  id: number;
  code: string;
  name: string;
  type: string;
  rarity: string;
  category: string;
  itemType: string;
  tier?: string;                // NEW - "Normal" | "Exceptional" | "Elite"
  typeTags?: string[];          // NEW - e.g., ["Helms"], ["Swords", "Melee Weapons"]
  classSpecific?: string;       // NEW - e.g., "amazon", "paladin", or undefined
  requirements: ItemRequirements;
  defense?: DefenseRange;
  damage?: DamageRange;
  speed?: number;
  maxSockets: number;
  durability: number;
  qualityTiers?: QualityTiers;
  imageUrl?: string;
  iconVariants?: string[];      // NEW - alternate icon URLs
}

// RunewordValidType - code field changed meaning
interface RunewordValidType {
  code: string;   // NOW: human-readable name (e.g., "Body Armor"), NOT type code
  name: string;   // Human-readable name (same as code)
}
```
