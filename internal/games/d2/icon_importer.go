package d2

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/storage"
	"golang.org/x/time/rate"
)

const (
	baseURL         = "https://diablo2.io"
	requestsPerSec  = 2 // Be polite to avoid rate limiting
	downloadTimeout = 30 * time.Second
)

// IconImportStats tracks import statistics
type IconImportStats struct {
	TotalParsed    int
	Matched        int
	Skipped        int // Already have image
	Downloaded     int
	UploadedNew    int
	UploadedCached int // Reused duplicate
	Failed         int
	Errors         []string
}

// IconImporter handles importing item icons from HTML files
type IconImporter struct {
	repo            *Repository
	storage         *storage.SupabaseStorage
	parser          *HTMLParser
	limiter         *rate.Limiter
	httpClient      *http.Client
	dryRun          bool
	force           bool
	cookies         string
	localImagesPath string // Path to locally downloaded images

	// Cache for duplicate image handling: imagePath -> uploadedURL
	imageCache map[string]string
}

// NewIconImporter creates a new icon importer
func NewIconImporter(repo *Repository, storage *storage.SupabaseStorage, dryRun, force bool, cookies, localImagesPath string) *IconImporter {
	return &IconImporter{
		repo:    repo,
		storage: storage,
		parser:  NewHTMLParser(),
		limiter: rate.NewLimiter(rate.Limit(requestsPerSec), 1),
		httpClient: &http.Client{
			Timeout: downloadTimeout,
		},
		dryRun:          dryRun,
		force:           force,
		cookies:         cookies,
		localImagesPath: localImagesPath,
		imageCache:      make(map[string]string),
	}
}

// Import runs the icon import process
func (i *IconImporter) Import(ctx context.Context, catalogPath string) (*IconImportStats, error) {
	stats := &IconImportStats{}

	// Define HTML file to table mappings
	fileMappings := []struct {
		htmlFile   string
		itemTypes  []string // "unique", "set", "base", "rune", "gem"
		categories []string // Storage path categories
	}{
		{
			htmlFile:   "uniques.html",
			itemTypes:  []string{"unique"},
			categories: []string{"d2/unique"},
		},
		{
			htmlFile:   "sets.html",
			itemTypes:  []string{"set"},
			categories: []string{"d2/set"},
		},
		{
			htmlFile:   "base.html",
			itemTypes:  []string{"base"},
			categories: []string{"d2/base"},
		},
		{
			htmlFile:   "misc.html",
			itemTypes:  []string{"base", "rune", "gem"},
			categories: []string{"d2/base", "d2/rune", "d2/gem"},
		},
	}

	// Load all items needing images from database
	dbItems, err := i.loadDBItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load database items: %w", err)
	}

	fmt.Printf("Loaded %d items from database needing images\n", len(dbItems))

	// Build lookup maps by normalized name for each type
	lookupMaps := i.buildLookupMaps(dbItems)

	// Process each HTML file
	for _, mapping := range fileMappings {
		htmlPath := filepath.Join(catalogPath, "icons", mapping.htmlFile)
		fmt.Printf("\nProcessing %s...\n", mapping.htmlFile)

		parsedItems, err := i.parser.ParseFile(htmlPath)
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to parse %s: %v", mapping.htmlFile, err))
			continue
		}

		stats.TotalParsed += len(parsedItems)
		fmt.Printf("  Parsed %d items from HTML\n", len(parsedItems))

		// Process each parsed item
		for _, parsed := range parsedItems {
			if err := i.processItem(ctx, parsed, mapping.itemTypes, mapping.categories, lookupMaps, stats); err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("Error processing %s: %v", parsed.Name, err))
			}
		}
	}

	return stats, nil
}

// loadDBItems loads all items from the database that need images
func (i *IconImporter) loadDBItems(ctx context.Context) ([]ItemWithoutImage, error) {
	var allItems []ItemWithoutImage

	if i.force {
		// When force is true, we need to load ALL items, not just those without images
		// This requires different queries
		return i.loadAllDBItems(ctx)
	}

	// Load unique items
	uniques, err := i.repo.GetUniqueItemsWithoutImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load unique items: %w", err)
	}
	allItems = append(allItems, uniques...)

	// Load set items
	sets, err := i.repo.GetSetItemsWithoutImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load set items: %w", err)
	}
	allItems = append(allItems, sets...)

	// Load item bases
	bases, err := i.repo.GetItemBasesWithoutImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load item bases: %w", err)
	}
	allItems = append(allItems, bases...)

	// Load runes
	runes, err := i.repo.GetRunesWithoutImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load runes: %w", err)
	}
	allItems = append(allItems, runes...)

	// Load gems
	gems, err := i.repo.GetGemsWithoutImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load gems: %w", err)
	}
	allItems = append(allItems, gems...)

	return allItems, nil
}

// loadAllDBItems loads ALL items regardless of image status (for --force mode)
func (i *IconImporter) loadAllDBItems(ctx context.Context) ([]ItemWithoutImage, error) {
	var allItems []ItemWithoutImage

	// Load all unique items
	rows, err := i.repo.pool.Query(ctx, `SELECT id, name FROM d2.unique_items ORDER BY id`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var item ItemWithoutImage
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			rows.Close()
			return nil, err
		}
		item.Type = "unique"
		allItems = append(allItems, item)
	}
	rows.Close()

	// Load all set items
	rows, err = i.repo.pool.Query(ctx, `SELECT id, name FROM d2.set_items ORDER BY id`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var item ItemWithoutImage
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			rows.Close()
			return nil, err
		}
		item.Type = "set"
		allItems = append(allItems, item)
	}
	rows.Close()

	// Load all item bases
	rows, err = i.repo.pool.Query(ctx, `SELECT code, name FROM d2.item_bases ORDER BY code`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var item ItemWithoutImage
		if err := rows.Scan(&item.Code, &item.Name); err != nil {
			rows.Close()
			return nil, err
		}
		item.Type = "base"
		allItems = append(allItems, item)
	}
	rows.Close()

	// Load all runes
	rows, err = i.repo.pool.Query(ctx, `SELECT id, code, name FROM d2.runes ORDER BY id`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var item ItemWithoutImage
		if err := rows.Scan(&item.ID, &item.Code, &item.Name); err != nil {
			rows.Close()
			return nil, err
		}
		item.Type = "rune"
		allItems = append(allItems, item)
	}
	rows.Close()

	// Load all gems
	rows, err = i.repo.pool.Query(ctx, `SELECT id, code, name FROM d2.gems ORDER BY id`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var item ItemWithoutImage
		if err := rows.Scan(&item.ID, &item.Code, &item.Name); err != nil {
			rows.Close()
			return nil, err
		}
		item.Type = "gem"
		allItems = append(allItems, item)
	}
	rows.Close()

	return allItems, nil
}

// buildLookupMaps creates lookup maps by normalized name for each item type
func (i *IconImporter) buildLookupMaps(items []ItemWithoutImage) map[string]map[string][]ItemWithoutImage {
	maps := make(map[string]map[string][]ItemWithoutImage)
	maps["unique"] = make(map[string][]ItemWithoutImage)
	maps["set"] = make(map[string][]ItemWithoutImage)
	maps["base"] = make(map[string][]ItemWithoutImage)
	maps["rune"] = make(map[string][]ItemWithoutImage)
	maps["gem"] = make(map[string][]ItemWithoutImage)

	for _, item := range items {
		key := NormalizeItemName(item.Name)
		maps[item.Type][key] = append(maps[item.Type][key], item)
	}

	return maps
}

// processItem processes a single parsed item
func (i *IconImporter) processItem(
	ctx context.Context,
	parsed ParsedItem,
	itemTypes []string,
	categories []string,
	lookupMaps map[string]map[string][]ItemWithoutImage,
	stats *IconImportStats,
) error {
	// Try to match against each item type
	for idx, itemType := range itemTypes {
		if matches, found := lookupMaps[itemType][parsed.NormalizedKey]; found {
			category := categories[idx]
			for _, match := range matches {
				stats.Matched++

				// Check if we've already uploaded this image (duplicate handling)
				if cachedURL, found := i.imageCache[parsed.ImagePath]; found {
					if i.dryRun {
						fmt.Printf("  [DRY-RUN] Would reuse cached image for %s (%s)\n", match.Name, itemType)
					} else {
						if err := i.updateImageURL(ctx, match, cachedURL); err != nil {
							stats.Failed++
							return fmt.Errorf("failed to update image URL: %w", err)
						}
						fmt.Printf("  Reused cached image for %s (%s)\n", match.Name, itemType)
					}
					stats.UploadedCached++
					continue
				}

				// Download and upload the image
				if i.dryRun {
					fmt.Printf("  [DRY-RUN] Would download and upload image for %s (%s): %s\n",
						match.Name, itemType, parsed.ImagePath)
					stats.Downloaded++
					stats.UploadedNew++
					// Cache it for dry-run duplicate detection
					i.imageCache[parsed.ImagePath] = "dry-run-url"
					continue
				}

				// Download image
				imageData, contentType, err := i.downloadImage(ctx, parsed.ImagePath)
				if err != nil {
					stats.Failed++
					return fmt.Errorf("failed to download image for %s: %w", match.Name, err)
				}
				stats.Downloaded++

				// Upload to Supabase Storage
				storagePath := storage.StoragePath(category, match.Name)
				publicURL, err := i.storage.UploadImage(ctx, storagePath, imageData, contentType)
				if err != nil {
					stats.Failed++
					return fmt.Errorf("failed to upload image for %s: %w", match.Name, err)
				}

				// Cache the uploaded URL
				i.imageCache[parsed.ImagePath] = publicURL

				// Update database
				if err := i.updateImageURL(ctx, match, publicURL); err != nil {
					stats.Failed++
					return fmt.Errorf("failed to update database for %s: %w", match.Name, err)
				}

				stats.UploadedNew++
				fmt.Printf("  Uploaded image for %s (%s): %s\n", match.Name, itemType, storagePath)
			}
			// Remove matched items from lookup to avoid double processing
			delete(lookupMaps[itemType], parsed.NormalizedKey)
		}
	}

	return nil
}

// downloadImage downloads an image from local files or diablo2.io
func (i *IconImporter) downloadImage(ctx context.Context, imagePath string) ([]byte, string, error) {
	// Extract filename from path (e.g., /styles/zulu/theme/images/items/stonecrusher_graphic.png -> stonecrusher_graphic.png)
	filename := filepath.Base(imagePath)

	// Try to load from local files first
	if i.localImagesPath != "" {
		localPath := filepath.Join(i.localImagesPath, filename)
		if data, err := os.ReadFile(localPath); err == nil {
			contentType := guessContentType(filename)
			return data, contentType, nil
		}
	}

	// Fall back to downloading from web
	// Rate limit
	if err := i.limiter.Wait(ctx); err != nil {
		return nil, "", err
	}

	url := baseURL + imagePath
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	// Use browser-like headers to avoid 403
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://diablo2.io/")
	req.Header.Set("Sec-Ch-Ua", "\"Not_A Brand\";v=\"8\", \"Chromium\";v=\"120\", \"Google Chrome\";v=\"120\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"Windows\"")
	req.Header.Set("Sec-Fetch-Dest", "image")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	// Add cookies if available (needed to bypass Cloudflare)
	if i.cookies != "" {
		req.Header.Set("Cookie", i.cookies)
	}

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP %d for %s (try downloading images locally with --images-path)", resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = guessContentType(imagePath)
	}

	return data, contentType, nil
}

// guessContentType guesses the content type from a filename
func guessContentType(filename string) string {
	switch {
	case strings.HasSuffix(filename, ".png"):
		return "image/png"
	case strings.HasSuffix(filename, ".jpg"), strings.HasSuffix(filename, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(filename, ".gif"):
		return "image/gif"
	case strings.HasSuffix(filename, ".webp"):
		return "image/webp"
	default:
		return "image/png"
	}
}

// updateImageURL updates the image URL in the database
func (i *IconImporter) updateImageURL(ctx context.Context, item ItemWithoutImage, url string) error {
	switch item.Type {
	case "unique":
		return i.repo.UpdateUniqueItemImageURL(ctx, item.ID, url)
	case "set":
		return i.repo.UpdateSetItemImageURL(ctx, item.ID, url)
	case "base":
		return i.repo.UpdateItemBaseImageURL(ctx, item.Code, url)
	case "rune":
		return i.repo.UpdateRuneImageURL(ctx, item.ID, url)
	case "gem":
		return i.repo.UpdateGemImageURL(ctx, item.ID, url)
	default:
		return fmt.Errorf("unknown item type: %s", item.Type)
	}
}
