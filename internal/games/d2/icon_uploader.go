package d2

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/storage"
)

// iconVariantFiles maps item base codes to their variant icon filenames
var iconVariantFiles = map[string][]string{
	"cm1": {"charm_small.png", "charm_small2.png", "charm_small3.png"},
	"cm2": {"charm_medium.png", "charm_medium2.png", "charm_medium3.png"},
	"cm3": {"charm_large.png", "charm_large2.png", "charm_large3.png"},
	"jew": {"jewel02_graphic.png", "jewel04_graphic.png", "jewel05_graphic.png", "jewel06_graphic.png"},
}

// fallbackIconMappings maps item base codes to fallback icon filenames
// Used when findInHTMLMapping fails for these codes
var fallbackIconMappings = map[string]string{
	"cm1": "charm_small.png",
	"cm2": "charm_medium.png",
	"cm3": "charm_large.png",
	"jew": "jewel02_graphic.png",
	"tes": "essencesuffering_graphic.png",
	"ceh": "essencehatred_graphic.png",
	"bet": "essenceterror_graphic.png",
	"fed": "essencedestruction_graphic.png",
	"toa": "tokenofabsolution_graphic.png",
	"2hs": "2hsword_graphic.png",
}

// fallbackIconByName maps item names (normalized) to fallback icon filenames
// Used for unique/set items that aren't found in HTML mapping
var fallbackIconByName = map[string]string{
	"swordbackhold": "swordbackhold_graphic.png",
}

// UploadStats tracks upload statistics
type UploadStats struct {
	TotalDBItems   int
	Uploaded       int
	ReusedCache    int
	MatchedUnique  int
	MatchedSet     int
	MatchedBase    int
	MatchedRune    int
	MatchedGem     int
	NotInHTML      int
	MissingFiles   int
	Errors         int
	MissingImages  []string // Images referenced in HTML but not in icons folder
	NotInHTMLItems []string // Items not found in HTML files
}

// IconUploader handles uploading local images to Supabase
type IconUploader struct {
	repo       *Repository
	storage    storage.Storage
	dryRun     bool
	iconsPath  string
	pagesPath  string
	imageCache map[string]string // imagePath -> uploadedURL
}

// NewIconUploader creates a new icon uploader
func NewIconUploader(repo *Repository, stor storage.Storage, dryRun bool) *IconUploader {
	return &IconUploader{
		repo:       repo,
		storage:    stor,
		dryRun:     dryRun,
		imageCache: make(map[string]string),
	}
}

// Upload scans HTML files for item-image mappings and uploads images
func (u *IconUploader) Upload(ctx context.Context, catalogPath string) (*UploadStats, error) {
	stats := &UploadStats{}
	u.iconsPath = filepath.Join(catalogPath, "icons")
	u.pagesPath = filepath.Join(catalogPath, "pages")

	// 1. Parse all HTML files to build item name -> image path mapping
	fmt.Println("Parsing HTML files for item-image mappings...")
	parser := NewHTMLParser()

	htmlMappings := map[string]map[string]string{
		"unique": make(map[string]string),
		"set":    make(map[string]string),
		"base":   make(map[string]string),
		"misc":   make(map[string]string), // runes, gems, etc.
	}

	// Parse uniques.html
	if items, err := parser.ParseFile(filepath.Join(u.pagesPath, "uniques.html")); err == nil {
		for _, item := range items {
			htmlMappings["unique"][normalizeForMatch(item.Name)] = item.ImagePath
		}
		fmt.Printf("  Parsed %d items from uniques.html\n", len(items))
	} else {
		fmt.Printf("  Warning: Could not parse uniques.html: %v\n", err)
	}

	// Parse sets.html
	if items, err := parser.ParseFile(filepath.Join(u.pagesPath, "sets.html")); err == nil {
		for _, item := range items {
			htmlMappings["set"][normalizeForMatch(item.Name)] = item.ImagePath
		}
		fmt.Printf("  Parsed %d items from sets.html\n", len(items))
	} else {
		fmt.Printf("  Warning: Could not parse sets.html: %v\n", err)
	}

	// Parse base.html
	if items, err := parser.ParseFile(filepath.Join(u.pagesPath, "base.html")); err == nil {
		for _, item := range items {
			htmlMappings["base"][normalizeForMatch(item.Name)] = item.ImagePath
		}
		fmt.Printf("  Parsed %d items from base.html\n", len(items))
	} else {
		fmt.Printf("  Warning: Could not parse base.html: %v\n", err)
	}

	// Parse misc.html (runes, gems, etc.)
	if items, err := parser.ParseFile(filepath.Join(u.pagesPath, "misc.html")); err == nil {
		for _, item := range items {
			htmlMappings["misc"][normalizeForMatch(item.Name)] = item.ImagePath
		}
		fmt.Printf("  Parsed %d items from misc.html\n", len(items))
	} else {
		fmt.Printf("  Warning: Could not parse misc.html: %v\n", err)
	}

	// 2. Load all items from database
	fmt.Println("\nLoading items from database...")

	// Process unique items
	if err := u.processItemType(ctx, "unique", "d2/unique", htmlMappings["unique"], stats); err != nil {
		return nil, err
	}

	// Process set items
	if err := u.processItemType(ctx, "set", "d2/set", htmlMappings["set"], stats); err != nil {
		return nil, err
	}

	// Process base items
	if err := u.processItemType(ctx, "base", "d2/base", htmlMappings["base"], stats); err != nil {
		return nil, err
	}

	// Process runes (from misc)
	if err := u.processItemType(ctx, "rune", "d2/rune", htmlMappings["misc"], stats); err != nil {
		return nil, err
	}

	// Process gems (from misc)
	if err := u.processItemType(ctx, "gem", "d2/gem", htmlMappings["misc"], stats); err != nil {
		return nil, err
	}

	// Upload icon variants for charms and jewels
	if err := u.uploadIconVariants(ctx, stats); err != nil {
		fmt.Printf("  Warning: Icon variant upload encountered errors: %v\n", err)
	}

	return stats, nil
}

// processItemType processes all items of a specific type
func (u *IconUploader) processItemType(ctx context.Context, itemType, category string, htmlMapping map[string]string, stats *UploadStats) error {
	fmt.Printf("\nProcessing %s items...\n", itemType)

	var items []ItemWithoutImage
	var err error

	switch itemType {
	case "unique":
		items, err = u.loadAllUniques(ctx)
	case "set":
		items, err = u.loadAllSets(ctx)
	case "base":
		items, err = u.loadAllBases(ctx)
	case "rune":
		items, err = u.loadAllRunes(ctx)
	case "gem":
		items, err = u.loadAllGems(ctx)
	}

	if err != nil {
		return fmt.Errorf("failed to load %s items: %w", itemType, err)
	}

	stats.TotalDBItems += len(items)
	fmt.Printf("  Loaded %d %s items from database\n", len(items), itemType)

	for _, item := range items {
		normalizedName := normalizeForMatch(item.Name)

		// Look up image path from HTML mapping (try multiple variations)
		imagePath, found := u.findInHTMLMapping(item.Name, htmlMapping)
		if !found {
			// Try fallback icon mapping by code (for bases, runes, gems)
			if fallbackFile, hasFallback := fallbackIconMappings[item.Code]; hasFallback {
				imagePath = fallbackFile
				found = true
			}
		}
		if !found {
			// Try fallback icon mapping by name (for uniques, sets)
			if fallbackFile, hasFallback := fallbackIconByName[normalizedName]; hasFallback {
				imagePath = fallbackFile
				found = true
			}
		}
		if !found {
			stats.NotInHTML++
			if len(stats.NotInHTMLItems) < 50 {
				stats.NotInHTMLItems = append(stats.NotInHTMLItems, fmt.Sprintf("%s (%s) [key: %s]", item.Name, itemType, normalizedName))
			}
			continue
		}

		// Check if we've already uploaded this image
		if cachedURL, exists := u.imageCache[imagePath]; exists {
			// Reuse cached URL
			if !u.dryRun {
				if err := u.updateItemURL(ctx, item, cachedURL); err != nil {
					stats.Errors++
					continue
				}
			}
			stats.ReusedCache++
			u.incrementMatchCount(itemType, stats)
			continue
		}

		// Extract filename from image path
		imageFilename := filepath.Base(imagePath)

		// Try to find the image file (check for variations with (1), (2), etc.)
		imageData, foundPath := u.findImageFile(imageFilename)
		if imageData == nil {
			stats.MissingFiles++
			if len(stats.MissingImages) < 50 {
				stats.MissingImages = append(stats.MissingImages, fmt.Sprintf("%s (for %s)", imageFilename, item.Name))
			}
			continue
		}
		_ = foundPath // Used for debugging if needed

		// Upload image
		storagePath := storage.StoragePath(category, item.Name)
		contentType := "image/png"
		if strings.HasSuffix(strings.ToLower(imageFilename), ".jpg") || strings.HasSuffix(strings.ToLower(imageFilename), ".jpeg") {
			contentType = "image/jpeg"
		}

		if u.dryRun {
			fmt.Printf("  [DRY-RUN] Would upload %s -> %s\n", imageFilename, storagePath)
			u.imageCache[imagePath] = "dry-run-url"
			stats.Uploaded++
			u.incrementMatchCount(itemType, stats)
			continue
		}

		publicURL, err := u.storage.UploadImage(ctx, storagePath, imageData, contentType)
		if err != nil {
			fmt.Printf("  Error uploading %s: %v\n", imageFilename, err)
			stats.Errors++
			continue
		}

		// Cache the URL for reuse
		u.imageCache[imagePath] = publicURL

		// Update database
		if err := u.updateItemURL(ctx, item, publicURL); err != nil {
			fmt.Printf("  Error updating DB for %s: %v\n", item.Name, err)
			stats.Errors++
			continue
		}

		stats.Uploaded++
		u.incrementMatchCount(itemType, stats)
		fmt.Printf("  ✓ %s -> %s\n", item.Name, storagePath)
	}

	return nil
}

func (u *IconUploader) incrementMatchCount(itemType string, stats *UploadStats) {
	switch itemType {
	case "unique":
		stats.MatchedUnique++
	case "set":
		stats.MatchedSet++
	case "base":
		stats.MatchedBase++
	case "rune":
		stats.MatchedRune++
	case "gem":
		stats.MatchedGem++
	}
}

func (u *IconUploader) updateItemURL(ctx context.Context, item ItemWithoutImage, url string) error {
	switch item.Type {
	case "unique":
		return u.repo.UpdateUniqueItemImageURL(ctx, item.ID, url)
	case "set":
		return u.repo.UpdateSetItemImageURL(ctx, item.ID, url)
	case "base":
		return u.repo.UpdateItemBaseImageURL(ctx, item.Code, url)
	case "rune":
		return u.repo.UpdateRuneImageURL(ctx, item.ID, url)
	case "gem":
		return u.repo.UpdateGemImageURL(ctx, item.ID, url)
	}
	return nil
}

// Load functions for each item type
func (u *IconUploader) loadAllUniques(ctx context.Context) ([]ItemWithoutImage, error) {
	rows, err := u.repo.pool.Query(ctx, `SELECT id, name FROM d2.unique_items ORDER BY id`)
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

func (u *IconUploader) loadAllSets(ctx context.Context) ([]ItemWithoutImage, error) {
	rows, err := u.repo.pool.Query(ctx, `SELECT id, name FROM d2.set_items ORDER BY id`)
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

func (u *IconUploader) loadAllBases(ctx context.Context) ([]ItemWithoutImage, error) {
	rows, err := u.repo.pool.Query(ctx, `SELECT code, name FROM d2.item_bases ORDER BY code`)
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

func (u *IconUploader) loadAllRunes(ctx context.Context) ([]ItemWithoutImage, error) {
	rows, err := u.repo.pool.Query(ctx, `SELECT id, code, name FROM d2.runes ORDER BY id`)
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

func (u *IconUploader) loadAllGems(ctx context.Context) ([]ItemWithoutImage, error) {
	rows, err := u.repo.pool.Query(ctx, `SELECT id, code, name FROM d2.gems ORDER BY id`)
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

// findInHTMLMapping tries to find an item's image path in the HTML mapping
// using various normalization strategies
func (u *IconUploader) findInHTMLMapping(itemName string, htmlMapping map[string]string) (string, bool) {
	// Try exact normalized match
	key := normalizeForMatch(itemName)
	if path, ok := htmlMapping[key]; ok {
		return path, true
	}

	// Try without "the" prefix
	if strings.HasPrefix(key, "the") {
		if path, ok := htmlMapping[strings.TrimPrefix(key, "the")]; ok {
			return path, true
		}
	}

	// Try adding "the" prefix
	if path, ok := htmlMapping["the"+key]; ok {
		return path, true
	}

	// For runes: DB has "Ber Rune" but HTML has just "Ber"
	// Try removing "rune" suffix
	if strings.HasSuffix(key, "rune") {
		withoutRune := strings.TrimSuffix(key, "rune")
		if path, ok := htmlMapping[withoutRune]; ok {
			return path, true
		}
	}

	// For gems: DB might have different naming
	// Try removing gem quality prefixes/suffixes
	gemQualities := []string{"chipped", "flawed", "flawless", "perfect"}
	for _, quality := range gemQualities {
		if strings.HasPrefix(key, quality) {
			withoutQuality := strings.TrimPrefix(key, quality)
			if path, ok := htmlMapping[withoutQuality]; ok {
				return path, true
			}
			// Also try with quality at end
			if path, ok := htmlMapping[withoutQuality+quality]; ok {
				return path, true
			}
		}
	}

	// Try partial match (item name contains or is contained in HTML name)
	for htmlKey, path := range htmlMapping {
		if len(htmlKey) > 2 && len(key) > 2 {
			if strings.Contains(htmlKey, key) || strings.Contains(key, htmlKey) {
				return path, true
			}
		}
	}

	return "", false
}

// findImageFile looks for an image file with various naming patterns
// Returns the file data and the path where it was found, or nil if not found
func (u *IconUploader) findImageFile(filename string) ([]byte, string) {
	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	// List of patterns to try
	patterns := []string{
		filename,                          // exact match: foo.png
		baseName + " (1)" + ext,           // foo (1).png
		baseName + " (2)" + ext,           // foo (2).png
		baseName + " (3)" + ext,           // foo (3).png
		baseName + "(1)" + ext,            // foo(1).png (no space)
		baseName + "_1" + ext,             // foo_1.png
		strings.ToLower(filename),         // lowercase
		strings.ToLower(baseName) + " (1)" + ext,
	}

	for _, pattern := range patterns {
		path := filepath.Join(u.iconsPath, pattern)
		if data, err := os.ReadFile(path); err == nil {
			return data, path
		}
	}

	// Also try case-insensitive search by scanning the directory
	entries, err := os.ReadDir(u.iconsPath)
	if err != nil {
		return nil, ""
	}

	lowerFilename := strings.ToLower(baseName)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		entryName := entry.Name()
		entryBase := strings.TrimSuffix(entryName, filepath.Ext(entryName))

		// Remove (1), (2), etc. from entry name for comparison
		cleanBase := entryBase
		if idx := strings.LastIndex(cleanBase, " ("); idx != -1 {
			cleanBase = cleanBase[:idx]
		}
		if idx := strings.LastIndex(cleanBase, "("); idx != -1 {
			cleanBase = cleanBase[:idx]
		}

		if strings.ToLower(cleanBase) == lowerFilename {
			path := filepath.Join(u.iconsPath, entryName)
			if data, err := os.ReadFile(path); err == nil {
				return data, path
			}
		}
	}

	return nil, ""
}

// uploadIconVariants uploads icon variant files for charms and jewels
func (u *IconUploader) uploadIconVariants(ctx context.Context, stats *UploadStats) error {
	fmt.Println("\nProcessing icon variants for charms and jewels...")

	for code, files := range iconVariantFiles {
		var variantURLs []string

		for _, filename := range files {
			imageData, _ := u.findImageFile(filename)
			if imageData == nil {
				fmt.Printf("  Warning: Variant file %s not found for %s\n", filename, code)
				continue
			}

			storagePath := fmt.Sprintf("d2/base-variants/%s/%s", code, filename)
			contentType := "image/png"

			if u.dryRun {
				fmt.Printf("  [DRY-RUN] Would upload variant %s -> %s\n", filename, storagePath)
				variantURLs = append(variantURLs, "dry-run-url")
				continue
			}

			publicURL, err := u.storage.UploadImage(ctx, storagePath, imageData, contentType)
			if err != nil {
				fmt.Printf("  Error uploading variant %s: %v\n", filename, err)
				stats.Errors++
				continue
			}

			variantURLs = append(variantURLs, publicURL)
			fmt.Printf("  ✓ Variant %s -> %s\n", filename, storagePath)
		}

		if len(variantURLs) == 0 {
			continue
		}

		if !u.dryRun {
			// Save variant URLs to database
			if err := u.repo.UpdateItemBaseIconVariants(ctx, code, variantURLs); err != nil {
				fmt.Printf("  Error updating icon variants for %s: %v\n", code, err)
				stats.Errors++
				continue
			}

			// Set first variant as primary image_url
			if err := u.repo.UpdateItemBaseImageURL(ctx, code, variantURLs[0]); err != nil {
				fmt.Printf("  Error updating primary image for %s: %v\n", code, err)
				stats.Errors++
				continue
			}
		}

		fmt.Printf("  ✓ %s: %d variants uploaded\n", code, len(variantURLs))
	}

	return nil
}

// normalizeForMatch normalizes a string for matching (removes special chars, lowercase)
func normalizeForMatch(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, ":", "")
	return s
}
