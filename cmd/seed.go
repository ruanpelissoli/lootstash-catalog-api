package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/database"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/storage"
	"github.com/spf13/cobra"
)

var (
	seedDryRun            bool
	seedSkipIcons         bool
	seedSkipRunewordIcons bool
	seedSkipVerify        bool
	seedCatalogPath       string
)

var seedCmd = &cobra.Command{
	Use:   "seed [game]",
	Short: "Full catalog seed: migrate, seed stats, import from HTML, upload icons, and verify",
	Long: `Seed performs a complete catalog population for a game using HTML data files.

IMPORTANT: Run 'supabase db reset' first to apply all migrations (including game schemas).

Steps performed (in order):
  1. Migrate       - Apply V2 schema changes (stats table, item_bases columns)
  2. Seed Stats    - Seed stat codes from FilterableStats + class data
  3. HTML Import   - Import all items from HTML pages (bases, uniques, sets, runewords, misc)
  4. Upload Icons  - Upload icons to storage for items without images
  5. Runeword Icons - Generate composite runeword images from rune icons
  6. Verify        - Verify data integrity

Prerequisites:
  - Run 'supabase db reset' first to create schemas and tables
  - Database running and accessible
  - HTML files in catalogs/<game>/pages/
  - Icon files in catalogs/<game>/icons/

Available games:
  d2 - Diablo II: Resurrected

Examples:
  supabase db reset && lootstash-catalog seed d2
  lootstash-catalog seed d2 --dry-run
  lootstash-catalog seed d2 --skip-icons`,
	Args: cobra.ExactArgs(1),
	RunE: runSeed,
}

func init() {
	rootCmd.AddCommand(seedCmd)

	seedCmd.Flags().BoolVar(&seedDryRun, "dry-run", false, "Preview all steps without making changes")
	seedCmd.Flags().BoolVar(&seedSkipIcons, "skip-icons", false, "Skip icon upload step")
	seedCmd.Flags().BoolVar(&seedSkipRunewordIcons, "skip-runeword-icons", false, "Skip runeword icon generation step")
	seedCmd.Flags().BoolVar(&seedSkipVerify, "skip-verify", false, "Skip verification step")
	seedCmd.Flags().StringVar(&seedCatalogPath, "catalog", "catalogs/d2", "Path to catalog folder")
}

func runSeed(cmd *cobra.Command, args []string) error {
	game := args[0]

	if game != "d2" {
		return fmt.Errorf("unknown game: %s. Available games: d2", game)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Minute)
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

	// Step 1: Migrate schema
	if err := seedStepMigrate(ctx, db); err != nil {
		return err
	}
	fmt.Println()

	// Step 2: Seed stats
	repo := d2.NewRepository(db.Pool())
	if err := seedStepSeedStats(ctx, repo); err != nil {
		return err
	}
	fmt.Println()

	// Step 3: HTML Import
	if err := seedStepHTMLImportV2(ctx, repo); err != nil {
		return err
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

// Step 1: Migrate schema
func seedStepMigrate(ctx context.Context, db *database.DB) error {
	fmt.Println("--- Step 1/6: Schema Migration ---")

	if seedDryRun {
		PrintInfo("Would apply V2 schema migrations")
		return nil
	}

	PrintInfo("Applying schema migrations...")
	if err := db.MigrateD2(ctx); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	PrintSuccess("Schema migration completed")
	return nil
}

// Step 2: Seed stats from FilterableStats + classes
func seedStepSeedStats(ctx context.Context, repo *d2.Repository) error {
	fmt.Println("--- Step 2/6: Seed Stats ---")

	if seedDryRun {
		PrintInfo("Would seed stat codes from FilterableStats + classes")
		return nil
	}

	statRegistry := d2.NewStatRegistry(repo)

	// Load existing stats
	if err := statRegistry.Load(ctx); err != nil {
		return fmt.Errorf("load stats: %w", err)
	}
	fmt.Printf("  Existing stats: %d\n", statRegistry.Count())

	// Seed from FilterableStats
	seeded, err := statRegistry.SeedFromFilterableStats(ctx)
	if err != nil {
		return fmt.Errorf("seed filterable stats: %w", err)
	}
	fmt.Printf("  Seeded from FilterableStats: %d\n", seeded)

	// Seed from classes
	classSeeded, err := statRegistry.SeedFromClasses(ctx)
	if err != nil {
		return fmt.Errorf("seed class stats: %w", err)
	}
	fmt.Printf("  Seeded from classes: %d\n", classSeeded)

	PrintSuccess(fmt.Sprintf("Stats seeded: %d total known", statRegistry.Count()))
	return nil
}

// Step 3: HTML Import (V2 pipeline)
func seedStepHTMLImportV2(ctx context.Context, repo *d2.Repository) error {
	fmt.Println("--- Step 3/6: HTML Import ---")

	// Initialize S3 storage for image uploads
	var stor storage.Storage
	if !seedDryRun {
		s3Stor, err := seedCreateS3Storage()
		if err != nil {
			PrintInfo(fmt.Sprintf("S3 storage not available: %v (images will be skipped)", err))
		} else {
			stor = s3Stor
			PrintSuccess("S3 storage initialized")
		}
	}

	// Create stat registry
	statRegistry := d2.NewStatRegistry(repo)
	if err := statRegistry.Load(ctx); err != nil {
		return fmt.Errorf("load stat registry: %w", err)
	}

	// Create and run V2 importer
	importer := d2.NewHTMLImporterV2(repo, statRegistry, stor, seedDryRun)

	PrintInfo("Importing all items from HTML...")
	result, err := importer.ImportAll(ctx, seedCatalogPath)
	if err != nil {
		return fmt.Errorf("HTML import failed: %w", err)
	}

	PrintSuccess("HTML import completed!")
	fmt.Printf("  Item Bases:       %d imported\n", result.ItemBases.Imported)
	fmt.Printf("  Unique Items:     %d imported\n", result.UniqueItems.Imported)
	fmt.Printf("  Set Bonuses:      %d imported\n", result.SetBonuses.Imported)
	fmt.Printf("  Set Items:        %d imported\n", result.SetItems.Imported)
	fmt.Printf("  Runewords:        %d imported\n", result.Runewords.Imported)
	fmt.Printf("  Runes:            %d imported\n", result.Runes.Imported)
	fmt.Printf("  Gems:             %d imported\n", result.Gems.Imported)
	fmt.Printf("  Runeword Bases:   %d computed\n", result.RunewordBases.Imported)
	fmt.Printf("  Images uploaded:  %d\n", result.ImagesUploaded)
	fmt.Printf("  Images missing:   %d\n", result.ImagesMissing)
	fmt.Printf("  Stats discovered: %d total\n", statRegistry.Count())

	return nil
}

// Step 4: Upload icons
func seedStepUploadIcons(ctx context.Context, db *database.DB) error {
	fmt.Println("--- Step 4/6: Icon Upload ---")

	if seedDryRun {
		PrintInfo("Would upload icons to storage")
		return nil
	}

	// Initialize S3 storage from env vars
	s3Stor, err := seedCreateS3Storage()
	if err != nil {
		return fmt.Errorf("S3 storage required for icon upload: %w", err)
	}

	// Create uploader
	repo := d2.NewRepository(db.Pool())
	uploader := d2.NewIconUploader(repo, s3Stor, seedDryRun, true)

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

	// Initialize S3 storage from env vars
	s3Stor, err := seedCreateS3Storage()
	if err != nil {
		return fmt.Errorf("S3 storage required for runeword icon generation: %w", err)
	}

	// Create generator
	repo := d2.NewRepository(db.Pool())
	iconsPath := seedCatalogPath + "/icons"
	generator := d2.NewRunewordImageGenerator(repo, s3Stor, iconsPath, seedDryRun, false)

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
		{"Stats", "SELECT COUNT(*) FROM d2.stats"},
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
		{"Unique Items (name)", `SELECT COUNT(*) FROM (SELECT name FROM d2.unique_items GROUP BY name HAVING COUNT(*) > 1) x`},
		{"Set Items (name)", `SELECT COUNT(*) FROM (SELECT name FROM d2.set_items GROUP BY name HAVING COUNT(*) > 1) x`},
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
	if total > 0 {
		fmt.Printf("\n  Unique items with images: %d/%d (%.1f%%)\n", withImages, total, float64(withImages)/float64(total)*100)
	}

	pool.QueryRow(ctx, `SELECT COUNT(*) FROM d2.set_items WHERE image_url IS NOT NULL AND image_url != ''`).Scan(&withImages)
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM d2.set_items`).Scan(&total)
	if total > 0 {
		fmt.Printf("  Set items with images: %d/%d (%.1f%%)\n", withImages, total, float64(withImages)/float64(total)*100)
	}

	pool.QueryRow(ctx, `SELECT COUNT(*) FROM d2.runewords WHERE image_url IS NOT NULL AND image_url != ''`).Scan(&withImages)
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM d2.runewords`).Scan(&total)
	if total > 0 {
		fmt.Printf("  Runewords with images: %d/%d (%.1f%%)\n", withImages, total, float64(withImages)/float64(total)*100)
	}

	// Check stats
	var statCount int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM d2.stats`).Scan(&statCount)
	fmt.Printf("\n  Stat codes registered: %d\n", statCount)

	if allGood {
		PrintSuccess("Verification passed")
	} else {
		PrintError("Verification found issues")
	}

	return nil
}

// seedCreateS3Storage creates an S3 storage client from environment variables
func seedCreateS3Storage() (storage.Storage, error) {
	s3AccessKey := getEnvOrDefault("SUPABASE_S3_ACCESS_KEY", "")
	s3SecretKey := getEnvOrDefault("SUPABASE_S3_SECRET_KEY", "")
	if s3AccessKey == "" || s3SecretKey == "" {
		return nil, fmt.Errorf("SUPABASE_S3_ACCESS_KEY and SUPABASE_S3_SECRET_KEY must be set")
	}
	supabaseURL := getEnvOrDefault("SUPABASE_URL", "http://127.0.0.1:54321")
	return storage.NewS3Storage(
		supabaseURL+"/storage/v1/s3",
		s3AccessKey,
		s3SecretKey,
		getEnvOrDefault("SUPABASE_S3_REGION", "local"),
		"d2-items",
		supabaseURL,
	)
}
