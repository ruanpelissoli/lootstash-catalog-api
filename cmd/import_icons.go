package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/database"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/storage"
	"github.com/spf13/cobra"
)

var (
	iconDryRun       bool
	iconForce        bool
	supabaseURL      string
	supabaseKey      string
	cookieFile       string
	imagesPath       string
)

var importIconsCmd = &cobra.Command{
	Use:   "import-icons",
	Short: "Import item icons from scraped HTML files",
	Long: `Import parses HTML files scraped from diablo2.io, downloads item images,
uploads them to Supabase Storage, and updates database records with image URLs.

HTML files should be placed in catalogs/d2/icons/ directory:
  - uniques.html: Unique items
  - sets.html: Set items
  - base.html: Base items (armor, weapons)
  - misc.html: Miscellaneous items (runes, gems, etc.)

Downloading images:
  diablo2.io blocks automated downloads. Use DownThemAll browser extension to
  download all images locally, then use --images-path to import from local files.

  1. Install DownThemAll (Chrome/Firefox extension)
  2. Visit diablo2.io/uniques/, /sets/, /base/, /misc/
  3. Use DownThemAll to download all *.png files to a folder
  4. Run: import-icons --images-path /path/to/downloaded/images

Examples:
  # Import from locally downloaded images (recommended)
  lootstash-catalog import-icons --images-path ./images

  # Preview import (dry run)
  lootstash-catalog import-icons --images-path ./images --dry-run

  # Force re-import all images
  lootstash-catalog import-icons --images-path ./images --force`,
	RunE: runImportIcons,
}

func init() {
	rootCmd.AddCommand(importIconsCmd)

	importIconsCmd.Flags().BoolVar(&iconDryRun, "dry-run", false, "Preview import without making changes")
	importIconsCmd.Flags().BoolVar(&iconForce, "force", false, "Re-import all images, even if already set")
	importIconsCmd.Flags().StringVar(&supabaseURL, "supabase-url", getEnvOrDefault("SUPABASE_URL", ""), "Supabase project URL")
	importIconsCmd.Flags().StringVar(&supabaseKey, "supabase-key", getEnvOrDefault("SUPABASE_SERVICE_KEY", ""), "Supabase service key")
	importIconsCmd.Flags().StringVar(&cookieFile, "cookie-file", "cookies.txt", "File containing browser cookies for diablo2.io")
	importIconsCmd.Flags().StringVar(&imagesPath, "images-path", "", "Path to locally downloaded images (use DownThemAll to download)")
}

func runImportIcons(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if supabaseURL == "" || supabaseKey == "" {
		return fmt.Errorf("--supabase-url and --supabase-key flags (or SUPABASE_URL and SUPABASE_SERVICE_KEY env vars) are required")
	}

	if iconDryRun {
		PrintInfo("Running in DRY-RUN mode - no changes will be made")
	}
	if iconForce {
		PrintInfo("Running in FORCE mode - will re-import all images")
	}

	// Initialize database connection
	PrintInfo("Connecting to database...")
	db, err := database.NewConnection(ctx, GetDatabaseURL())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	PrintSuccess("Connected to database")

	// Initialize Supabase Storage
	PrintInfo("Initializing Supabase Storage...")
	supabaseStorage := storage.NewSupabaseStorage(supabaseURL, supabaseKey, "d2-items")
	PrintSuccess("Supabase Storage initialized")

	// Load cookies from file (optional)
	cookies := ""
	if cookieData, err := os.ReadFile(cookieFile); err == nil {
		cookies = strings.TrimSpace(string(cookieData))
		PrintSuccess(fmt.Sprintf("Loaded cookies from %s", cookieFile))
	}

	// Check for local images path
	if imagesPath != "" {
		PrintSuccess(fmt.Sprintf("Using local images from: %s", imagesPath))
	} else {
		PrintInfo("No --images-path specified. Will try to download from web (may fail without cookies).")
	}

	// Create repository and importer
	repo := d2.NewRepository(db.Pool())
	importer := d2.NewIconImporter(repo, supabaseStorage, iconDryRun, iconForce, cookies, imagesPath)

	// Run import
	PrintInfo("Starting icon import...")
	stats, err := importer.Import(ctx, "catalogs/d2")
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	// Print statistics
	fmt.Println()
	if iconDryRun {
		PrintSuccess("Dry run completed!")
	} else {
		PrintSuccess("Icon import completed!")
	}

	fmt.Println("\nImport Statistics:")
	fmt.Printf("  Total Parsed:    %d items from HTML\n", stats.TotalParsed)
	fmt.Printf("  Matched:         %d items matched to database\n", stats.Matched)
	fmt.Printf("  Downloaded:      %d images downloaded\n", stats.Downloaded)
	fmt.Printf("  Uploaded (new):  %d images uploaded\n", stats.UploadedNew)
	fmt.Printf("  Uploaded (cache):%d images reused from cache\n", stats.UploadedCached)
	fmt.Printf("  Failed:          %d items failed\n", stats.Failed)

	if len(stats.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range stats.Errors {
			PrintError(e)
		}
	}

	return nil
}
