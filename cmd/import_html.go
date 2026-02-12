package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/database"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/storage"
	"github.com/spf13/cobra"
)

var (
	htmlImportType    string
	htmlImportCatalog string
	htmlImportDryRun  bool
	htmlImportForce   bool
	htmlS3Endpoint    string
	htmlS3AccessKey   string
	htmlS3SecretKey   string
	htmlS3Region      string
	htmlS3PublicURL   string
)

var importHTMLCmd = &cobra.Command{
	Use:   "import-html [game]",
	Short: "Import items from HTML pages (for expansion data without TSV files)",
	Long: `Parses HTML files from diablo2.io to extract full item data (properties, set bonuses,
runeword details) and imports items that don't already exist in the database.
Also uploads item images from the local icons folder.

This is an alternative import path for expansion items where TSV data files
are not available. Properties are reverse-translated from display text back
to property codes where possible; unrecognized properties are stored with
code "raw" and the full display text preserved.

Examples:
  # Preview what would be imported
  lootstash-catalog import-html d2 --dry-run

  # Import only new base items
  lootstash-catalog import-html d2 --type=bases

  # Import only new unique items
  lootstash-catalog import-html d2 --type=uniques

  # Import all new items with images
  lootstash-catalog import-html d2 --type=all

  # Force re-import all items (overwrite existing)
  lootstash-catalog import-html d2 --type=all --force`,
	Args: cobra.ExactArgs(1),
	RunE: runImportHTML,
}

func init() {
	rootCmd.AddCommand(importHTMLCmd)

	importHTMLCmd.Flags().StringVar(&htmlImportType, "type", "all", "Item type to import: bases, uniques, sets, runewords, or all")
	importHTMLCmd.Flags().StringVar(&htmlImportCatalog, "catalog", "catalogs/d2", "Path to catalog folder (contains icons/ and pages/ subfolders)")
	importHTMLCmd.Flags().BoolVar(&htmlImportDryRun, "dry-run", false, "Preview without making changes")
	importHTMLCmd.Flags().BoolVar(&htmlImportForce, "force", false, "Re-import existing items (overwrite with fresh data)")

	// S3 configuration - same pattern as upload-icons
	supabaseDefault := getEnvOrDefault("SUPABASE_URL", "http://127.0.0.1:54321")
	importHTMLCmd.Flags().StringVar(&htmlS3Endpoint, "s3-endpoint", supabaseDefault+"/storage/v1/s3", "S3 endpoint URL")
	importHTMLCmd.Flags().StringVar(&htmlS3AccessKey, "s3-access-key", getEnvOrDefault("SUPABASE_S3_ACCESS_KEY", ""), "S3 access key")
	importHTMLCmd.Flags().StringVar(&htmlS3SecretKey, "s3-secret-key", getEnvOrDefault("SUPABASE_S3_SECRET_KEY", ""), "S3 secret key")
	importHTMLCmd.Flags().StringVar(&htmlS3Region, "s3-region", getEnvOrDefault("SUPABASE_S3_REGION", "local"), "S3 region")
	importHTMLCmd.Flags().StringVar(&htmlS3PublicURL, "s3-public-url", supabaseDefault, "S3 public URL base for generating public image URLs")
}

func runImportHTML(cmd *cobra.Command, args []string) error {
	game := args[0]
	if game != "d2" {
		return fmt.Errorf("unsupported game: %s (only 'd2' is supported)", game)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if htmlImportDryRun {
		PrintInfo("Running in DRY-RUN mode")
	}
	if htmlImportForce {
		PrintInfo("Running in FORCE mode (overwriting existing items)")
	}

	// Connect to database
	PrintInfo("Connecting to database...")
	db, err := database.NewConnection(ctx, GetDatabaseURL())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	PrintSuccess("Connected to database")

	// Optionally initialize S3 storage for image uploads
	var s3Storage storage.Storage
	if htmlS3AccessKey != "" && htmlS3SecretKey != "" {
		PrintInfo("Connecting to S3 storage...")
		s3Stor, err := storage.NewS3Storage(
			htmlS3Endpoint,
			htmlS3AccessKey,
			htmlS3SecretKey,
			htmlS3Region,
			"d2-items",
			htmlS3PublicURL,
		)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to create S3 storage: %v (images will be skipped)", err))
		} else {
			s3Storage = s3Stor
			PrintSuccess("S3 storage initialized")
		}
	} else {
		PrintInfo("No S3 credentials configured, skipping image uploads")
	}

	// Create importer and run
	repo := d2.NewRepository(db.Pool())
	importer := d2.NewHTMLImporter(repo, s3Storage, htmlImportDryRun, htmlImportForce)

	PrintInfo(fmt.Sprintf("Importing %s items from HTML...", htmlImportType))
	stats, err := importer.ImportFromHTML(ctx, htmlImportCatalog, htmlImportType)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	// Print results
	fmt.Println()
	if htmlImportDryRun {
		PrintSuccess("Dry run completed!")
	} else {
		PrintSuccess("Import completed!")
	}

	fmt.Println("\nStatistics:")
	fmt.Printf("  Bases imported:     %d\n", stats.BasesImported)
	fmt.Printf("  Bases skipped:      %d\n", stats.BasesSkipped)
	fmt.Printf("  Uniques imported:   %d\n", stats.UniquesImported)
	fmt.Printf("  Uniques skipped:    %d\n", stats.UniquesSkipped)
	fmt.Printf("  Sets imported:      %d\n", stats.SetsImported)
	fmt.Printf("  Set items imported: %d\n", stats.SetItemsImported)
	fmt.Printf("  Set items skipped:  %d\n", stats.SetItemsSkipped)
	fmt.Printf("  Runewords imported: %d\n", stats.RunewordsImported)
	fmt.Printf("  Runewords skipped:  %d\n", stats.RunewordsSkipped)
	fmt.Printf("  Images uploaded:    %d\n", stats.ImagesUploaded)
	fmt.Printf("  Raw properties:     %d\n", stats.RawProperties)
	fmt.Printf("  Errors:             %d\n", stats.Errors)

	if len(stats.ErrorMessages) > 0 {
		fmt.Println("\nErrors:")
		for _, msg := range stats.ErrorMessages {
			fmt.Printf("  - %s\n", msg)
		}
	}

	return nil
}
