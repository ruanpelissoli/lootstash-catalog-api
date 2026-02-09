package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/cache"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/database"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/storage"
	"github.com/spf13/cobra"
)

var (
	seedDryRun             bool
	seedSkipImport         bool
	seedSkipCleanup        bool
	seedSkipNames          bool
	seedSkipIcons          bool
	seedSkipRunewordIcons  bool
	seedSkipVerify         bool
	seedSupabaseURL        string
	seedSupabaseServiceKey string
	seedCatalogPath        string
)

var seedCmd = &cobra.Command{
	Use:   "seed [game]",
	Short: "Full catalog seed: import, cleanup, sync names, upload icons, and verify",
	Long: `Seed performs a complete catalog population for a game. This is the single
command to run after a database reset to restore all catalog data.

IMPORTANT: Run 'supabase db reset' first to apply all migrations (including game schemas).
Migrations are managed in supabase/migrations/ - each game has its own schema (d2, poe, etc.)

Steps performed (in order):
  1. Import         - Import catalog data from datatables
  2. Cleanup        - Remove duplicate runes/gems from item_bases
  3. Sync Names     - Update item names from HTML files (community names)
  4. Icons          - Upload icons to storage and update image URLs
  5. Runeword Icons - Generate composite runeword images from rune icons
  6. Verify         - Verify data integrity

Prerequisites:
  - Run 'supabase db reset' first to create schemas and tables
  - Database running and accessible
  - Catalog data files in catalogs/<game>/ directory
  - HTML files in catalogs/<game>/pages/ (for name sync)
  - Icon files in catalogs/<game>/icons/ (for icon upload)

Available games:
  d2 - Diablo II: Resurrected

Examples:
  # Full database reset and seed
  supabase db reset && lootstash-catalog seed d2

  # Just seed (if schema already exists)
  lootstash-catalog seed d2

  # Preview what would happen (dry run)
  lootstash-catalog seed d2 --dry-run

  # Skip specific steps
  lootstash-catalog seed d2 --skip-icons

  # Use custom S3 configuration
  lootstash-catalog seed d2 --s3-endpoint http://localhost:9000`,
	Args: cobra.ExactArgs(1),
	RunE: runSeed,
}

func init() {
	rootCmd.AddCommand(seedCmd)

	seedCmd.Flags().BoolVar(&seedDryRun, "dry-run", false, "Preview all steps without making changes")
	seedCmd.Flags().BoolVar(&seedSkipImport, "skip-import", false, "Skip catalog import step")
	seedCmd.Flags().BoolVar(&seedSkipCleanup, "skip-cleanup", false, "Skip duplicate cleanup step")
	seedCmd.Flags().BoolVar(&seedSkipNames, "skip-names", false, "Skip name sync step")
	seedCmd.Flags().BoolVar(&seedSkipIcons, "skip-icons", false, "Skip icon upload step")
	seedCmd.Flags().BoolVar(&seedSkipRunewordIcons, "skip-runeword-icons", false, "Skip runeword icon generation step")
	seedCmd.Flags().BoolVar(&seedSkipVerify, "skip-verify", false, "Skip verification step")

	seedCmd.Flags().StringVar(&seedCatalogPath, "catalog", "catalogs/d2", "Path to catalog folder")
	seedCmd.Flags().StringVar(&seedSupabaseURL, "supabase-url", getEnvOrDefault("SUPABASE_URL", "http://127.0.0.1:54321"), "Supabase URL (env: SUPABASE_URL)")
	seedCmd.Flags().StringVar(&seedSupabaseServiceKey, "supabase-service-key", getEnvOrDefault("SUPABASE_SERVICE_KEY", ""), "Supabase service role key (env: SUPABASE_SERVICE_KEY)")
}

func runSeed(cmd *cobra.Command, args []string) error {
	game := args[0]

	if game != "d2" {
		return fmt.Errorf("unknown game: %s. Available games: d2", game)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	fmt.Println("========================================")
	fmt.Printf("  LootStash Catalog Seed - %s\n", strings.ToUpper(game))
	fmt.Println("========================================")
	fmt.Println()
	PrintInfo("NOTE: Run 'supabase db reset' first if schema doesn't exist")

	if seedDryRun {
		PrintInfo("DRY-RUN MODE - No changes will be made")
	}
	fmt.Println()

	// Connect to database
	PrintInfo("Connecting to database...")
	db, err := database.NewConnection(ctx, GetDatabaseURL())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	PrintSuccess("Connected to database")

	// Verify schema exists
	var schemaExists bool
	err = db.Pool().QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = $1)", game).Scan(&schemaExists)
	if err != nil || !schemaExists {
		return fmt.Errorf("schema '%s' does not exist. Run 'supabase db reset' first to apply migrations", game)
	}
	PrintSuccess(fmt.Sprintf("Schema '%s' exists", game))
	fmt.Println()

	// Step 1: Import
	if !seedSkipImport {
		if err := seedStepImport(ctx, db, game); err != nil {
			return err
		}
	} else {
		PrintInfo("Skipping import (--skip-import)")
	}
	fmt.Println()

	// Step 2: Cleanup duplicates
	if !seedSkipCleanup {
		if err := seedStepCleanup(ctx, db); err != nil {
			return err
		}
	} else {
		PrintInfo("Skipping cleanup (--skip-cleanup)")
	}
	fmt.Println()

	// Step 3: Sync names
	if !seedSkipNames {
		if err := seedStepSyncNames(ctx, db); err != nil {
			return err
		}
	} else {
		PrintInfo("Skipping name sync (--skip-names)")
	}
	fmt.Println()

	// Step 4: Upload icons
	if !seedSkipIcons {
		if err := seedStepUploadIcons(ctx, db); err != nil {
			return err
		}
	} else {
		PrintInfo("Skipping icon upload (--skip-icons)")
	}
	fmt.Println()

	// Step 5: Generate runeword icons
	if !seedSkipRunewordIcons {
		if err := seedStepGenerateRunewordIcons(ctx, db); err != nil {
			return err
		}
	} else {
		PrintInfo("Skipping runeword icon generation (--skip-runeword-icons)")
	}
	fmt.Println()

	// Step 6: Verify
	if !seedSkipVerify {
		if err := seedStepVerify(ctx, db); err != nil {
			return err
		}
	} else {
		PrintInfo("Skipping verification (--skip-verify)")
	}

	fmt.Println()
	fmt.Println("========================================")
	if seedDryRun {
		PrintSuccess("DRY-RUN COMPLETED - No changes were made")
	} else {
		PrintSuccess("SEED COMPLETED SUCCESSFULLY")
	}
	fmt.Println("========================================")

	return nil
}

// Step 1: Import catalog data
func seedStepImport(ctx context.Context, db *database.DB, game string) error {
	fmt.Println("--- Step 1/6: Catalog Import ---")

	if seedDryRun {
		PrintInfo("Would import catalog data from " + seedCatalogPath)
		return nil
	}

	// Initialize Redis cache (optional)
	var redisCache *cache.RedisCache
	redisCache, err := cache.NewRedisCache(ctx, GetRedisURL())
	if err != nil {
		PrintInfo(fmt.Sprintf("Redis not available: %v (continuing without cache)", err))
		redisCache = nil
	} else {
		defer redisCache.Close()
	}

	PrintInfo("Importing catalog data...")
	importer := d2.NewImporter(db, redisCache)
	stats, err := importer.Import(ctx, seedCatalogPath)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	PrintSuccess("Catalog import completed!")
	fmt.Printf("  Item Types:       %d imported, %d skipped\n", stats.ItemTypes.Imported, stats.ItemTypes.Skipped)
	fmt.Printf("  Item Bases:       %d imported, %d skipped\n", stats.ItemBases.Imported, stats.ItemBases.Skipped)
	fmt.Printf("  Unique Items:     %d imported, %d skipped\n", stats.UniqueItems.Imported, stats.UniqueItems.Skipped)
	fmt.Printf("  Set Bonuses:      %d imported, %d skipped\n", stats.SetBonuses.Imported, stats.SetBonuses.Skipped)
	fmt.Printf("  Set Items:        %d imported, %d skipped\n", stats.SetItems.Imported, stats.SetItems.Skipped)
	fmt.Printf("  Runewords:        %d imported, %d skipped\n", stats.Runewords.Imported, stats.Runewords.Skipped)
	fmt.Printf("  Runes:            %d imported, %d skipped\n", stats.Runes.Imported, stats.Runes.Skipped)
	fmt.Printf("  Gems:             %d imported, %d skipped\n", stats.Gems.Imported, stats.Gems.Skipped)
	fmt.Printf("  Properties:       %d imported, %d skipped\n", stats.Properties.Imported, stats.Properties.Skipped)
	fmt.Printf("  Affixes:          %d imported, %d skipped\n", stats.Affixes.Imported, stats.Affixes.Skipped)
	fmt.Printf("  Treasure Classes: %d imported, %d skipped\n", stats.TreasureClasses.Imported, stats.TreasureClasses.Skipped)
	fmt.Printf("  Item Ratios:      %d imported, %d skipped\n", stats.ItemRatios.Imported, stats.ItemRatios.Skipped)
	fmt.Printf("  Runeword Bases:   %d imported, %d skipped\n", stats.RunewordBases.Imported, stats.RunewordBases.Skipped)

	return nil
}

// Step 2: Cleanup duplicates
func seedStepCleanup(ctx context.Context, db *database.DB) error {
	fmt.Println("--- Step 2/6: Duplicate Cleanup ---")

	pool := db.Pool()

	// Find runes in item_bases
	runeRows, err := pool.Query(ctx, `
		SELECT ib.code, ib.name
		FROM d2.item_bases ib
		INNER JOIN d2.runes r ON LOWER(ib.name) = LOWER(r.name)
	`)
	if err != nil {
		return fmt.Errorf("failed to query rune duplicates: %w", err)
	}

	var runeCodes []string
	for runeRows.Next() {
		var code, name string
		if err := runeRows.Scan(&code, &name); err != nil {
			runeRows.Close()
			return err
		}
		runeCodes = append(runeCodes, code)
	}
	runeRows.Close()

	// Find gems in item_bases
	gemRows, err := pool.Query(ctx, `
		SELECT ib.code, ib.name
		FROM d2.item_bases ib
		INNER JOIN d2.gems g ON LOWER(ib.name) = LOWER(g.name)
	`)
	if err != nil {
		return fmt.Errorf("failed to query gem duplicates: %w", err)
	}

	var gemCodes []string
	for gemRows.Next() {
		var code, name string
		if err := gemRows.Scan(&code, &name); err != nil {
			gemRows.Close()
			return err
		}
		gemCodes = append(gemCodes, code)
	}
	gemRows.Close()

	fmt.Printf("  Found %d duplicate runes and %d duplicate gems\n", len(runeCodes), len(gemCodes))

	if seedDryRun {
		PrintInfo("Would delete duplicates")
		return nil
	}

	// Delete duplicates
	deleted := 0
	for _, code := range runeCodes {
		_, err := pool.Exec(ctx, `DELETE FROM d2.item_bases WHERE code = $1`, code)
		if err == nil {
			deleted++
		}
	}
	for _, code := range gemCodes {
		_, err := pool.Exec(ctx, `DELETE FROM d2.item_bases WHERE code = $1`, code)
		if err == nil {
			deleted++
		}
	}

	PrintSuccess(fmt.Sprintf("Deleted %d duplicate items", deleted))
	return nil
}

// Step 3: Sync names from HTML
func seedStepSyncNames(ctx context.Context, db *database.DB) error {
	fmt.Println("--- Step 3/6: Name Sync ---")

	pool := db.Pool()
	parser := d2.NewHTMLParser()
	pagesPath := seedCatalogPath + "/pages"

	// Build HTML name lookup
	htmlNames := make(map[string]string)

	if items, err := parser.ParseFile(pagesPath + "/uniques.html"); err == nil {
		for _, item := range items {
			key := seedNormalizeForLookup(item.Name)
			htmlNames[key] = item.Name
		}
		fmt.Printf("  Parsed %d unique items from HTML\n", len(items))
	}

	if items, err := parser.ParseFile(pagesPath + "/sets.html"); err == nil {
		for _, item := range items {
			key := seedNormalizeForLookup(item.Name)
			htmlNames[key] = item.Name
		}
		fmt.Printf("  Parsed %d set items from HTML\n", len(items))
	}

	if items, err := parser.ParseFile(pagesPath + "/base.html"); err == nil {
		for _, item := range items {
			key := seedNormalizeForLookup(item.Name)
			htmlNames[key] = item.Name
		}
		fmt.Printf("  Parsed %d base items from HTML\n", len(items))
	}

	if items, err := parser.ParseFile(pagesPath + "/misc.html"); err == nil {
		for _, item := range items {
			key := seedNormalizeForLookup(item.Name)
			htmlNames[key] = item.Name
		}
		fmt.Printf("  Parsed %d misc items from HTML\n", len(items))
	}

	if len(htmlNames) == 0 {
		PrintInfo("No HTML files found, skipping name sync")
		return nil
	}

	updated := 0

	// Update unique items
	rows, err := pool.Query(ctx, `SELECT id, name FROM d2.unique_items ORDER BY id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var dbName string
		rows.Scan(&id, &dbName)

		htmlName := seedFindBestMatch(dbName, htmlNames)
		if htmlName != "" && htmlName != dbName {
			if !seedDryRun {
				pool.Exec(ctx, `UPDATE d2.unique_items SET name = $1 WHERE id = $2`, htmlName, id)
			}
			updated++
		}
	}
	rows.Close()

	// Update set items
	rows, err = pool.Query(ctx, `SELECT id, name FROM d2.set_items ORDER BY id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var dbName string
		rows.Scan(&id, &dbName)

		htmlName := seedFindBestMatch(dbName, htmlNames)
		if htmlName != "" && htmlName != dbName {
			if !seedDryRun {
				pool.Exec(ctx, `UPDATE d2.set_items SET name = $1 WHERE id = $2`, htmlName, id)
			}
			updated++
		}
	}
	rows.Close()

	if seedDryRun {
		fmt.Printf("  Would update %d item names\n", updated)
	} else {
		PrintSuccess(fmt.Sprintf("Updated %d item names", updated))
	}

	return nil
}

// Step 4: Upload icons
func seedStepUploadIcons(ctx context.Context, db *database.DB) error {
	fmt.Println("--- Step 4/6: Icon Upload ---")

	if seedDryRun {
		PrintInfo("Would upload icons to storage")
		return nil
	}

	// Initialize Supabase storage
	supabaseStorage := storage.NewSupabaseStorage(
		seedSupabaseURL,
		seedSupabaseServiceKey,
		"d2-items",
	)

	// Create uploader
	repo := d2.NewRepository(db.Pool())
	uploader := d2.NewIconUploader(repo, supabaseStorage, seedDryRun)

	// Run upload
	stats, err := uploader.Upload(ctx, seedCatalogPath)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	PrintSuccess("Icon upload completed!")
	fmt.Printf("  Total DB items:   %d\n", stats.TotalDBItems)
	fmt.Printf("  Uploaded (new):   %d\n", stats.Uploaded)
	fmt.Printf("  Reused (cache):   %d\n", stats.ReusedCache)
	fmt.Printf("  Missing files:    %d\n", stats.MissingFiles)

	return nil
}

// Step 5: Generate runeword icons
func seedStepGenerateRunewordIcons(ctx context.Context, db *database.DB) error {
	fmt.Println("--- Step 5/6: Runeword Icon Generation ---")

	if seedDryRun {
		PrintInfo("Would generate runeword composite icons")
		return nil
	}

	// Initialize Supabase storage
	supabaseStorage := storage.NewSupabaseStorage(
		seedSupabaseURL,
		seedSupabaseServiceKey,
		"d2-items",
	)

	// Create generator
	repo := d2.NewRepository(db.Pool())
	iconsPath := seedCatalogPath + "/icons"
	generator := d2.NewRunewordImageGenerator(repo, supabaseStorage, iconsPath, seedDryRun, false)

	// Run generation
	stats, err := generator.Generate(ctx)
	if err != nil {
		return fmt.Errorf("runeword icon generation failed: %w", err)
	}

	PrintSuccess("Runeword icon generation completed!")
	fmt.Printf("  Total runewords:  %d\n", stats.TotalRunewords)
	fmt.Printf("  Generated:        %d\n", stats.Generated)
	fmt.Printf("  Missing runes:    %d\n", stats.MissingRunes)
	fmt.Printf("  Failed:           %d\n", stats.Failed)

	return nil
}

// Step 6: Verify
func seedStepVerify(ctx context.Context, db *database.DB) error {
	fmt.Println("--- Step 6/6: Verification ---")

	pool := db.Pool()

	tables := []struct {
		name  string
		query string
	}{
		{"Item Types", "SELECT COUNT(*) FROM d2.item_types"},
		{"Item Bases", "SELECT COUNT(*) FROM d2.item_bases"},
		{"Unique Items", "SELECT COUNT(*) FROM d2.unique_items"},
		{"Set Bonuses", "SELECT COUNT(*) FROM d2.set_bonuses"},
		{"Set Items", "SELECT COUNT(*) FROM d2.set_items"},
		{"Runewords", "SELECT COUNT(*) FROM d2.runewords"},
		{"Runes", "SELECT COUNT(*) FROM d2.runes"},
		{"Gems", "SELECT COUNT(*) FROM d2.gems"},
		{"Properties", "SELECT COUNT(*) FROM d2.properties"},
		{"Affixes", "SELECT COUNT(*) FROM d2.affixes"},
		{"Treasure Classes", "SELECT COUNT(*) FROM d2.treasure_classes"},
		{"Item Ratios", "SELECT COUNT(*) FROM d2.item_ratios"},
		{"Runeword Bases", "SELECT COUNT(*) FROM d2.runeword_bases"},
	}

	fmt.Println("  Record Counts:")
	for _, t := range tables {
		var count int
		if err := pool.QueryRow(ctx, t.query).Scan(&count); err != nil {
			PrintError(fmt.Sprintf("  %s: ERROR", t.name))
		} else {
			fmt.Printf("    %-15s %d\n", t.name+":", count)
		}
	}

	// Quick duplicate check
	duplicateChecks := []struct {
		name  string
		query string
	}{
		{"Unique Items (index_id)", `SELECT COUNT(*) FROM (SELECT index_id FROM d2.unique_items GROUP BY index_id HAVING COUNT(*) > 1) x`},
		{"Set Items (index_id)", `SELECT COUNT(*) FROM (SELECT index_id FROM d2.set_items GROUP BY index_id HAVING COUNT(*) > 1) x`},
		{"Runewords (name)", `SELECT COUNT(*) FROM (SELECT name FROM d2.runewords GROUP BY name HAVING COUNT(*) > 1) x`},
	}

	fmt.Println("\n  Duplicate Checks:")
	allGood := true
	for _, check := range duplicateChecks {
		var count int
		if err := pool.QueryRow(ctx, check.query).Scan(&count); err != nil {
			PrintError(fmt.Sprintf("    %s: ERROR", check.name))
			allGood = false
		} else if count > 0 {
			PrintError(fmt.Sprintf("    %s: %d duplicates found", check.name, count))
			allGood = false
		} else {
			PrintSuccess(fmt.Sprintf("    %s: OK", check.name))
		}
	}

	// Check for items with images
	var withImages, total int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM d2.unique_items WHERE image_url IS NOT NULL AND image_url != ''`).Scan(&withImages)
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM d2.unique_items`).Scan(&total)
	fmt.Printf("\n  Unique items with images: %d/%d (%.1f%%)\n", withImages, total, float64(withImages)/float64(total)*100)

	pool.QueryRow(ctx, `SELECT COUNT(*) FROM d2.set_items WHERE image_url IS NOT NULL AND image_url != ''`).Scan(&withImages)
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM d2.set_items`).Scan(&total)
	fmt.Printf("  Set items with images: %d/%d (%.1f%%)\n", withImages, total, float64(withImages)/float64(total)*100)

	pool.QueryRow(ctx, `SELECT COUNT(*) FROM d2.runewords WHERE image_url IS NOT NULL AND image_url != ''`).Scan(&withImages)
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM d2.runewords`).Scan(&total)
	fmt.Printf("  Runewords with images: %d/%d (%.1f%%)\n", withImages, total, float64(withImages)/float64(total)*100)

	if allGood {
		PrintSuccess("Verification passed")
	} else {
		PrintError("Verification found issues")
	}

	return nil
}

func seedNormalizeForLookup(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

func seedFindBestMatch(dbName string, htmlNames map[string]string) string {
	key := seedNormalizeForLookup(dbName)
	if name, ok := htmlNames[key]; ok {
		return name
	}

	for htmlKey, htmlName := range htmlNames {
		if strings.Contains(htmlKey, key) || strings.Contains(key, htmlKey) {
			if len(key) > 3 && len(htmlKey) > 3 {
				return htmlName
			}
		}
	}

	return ""
}
