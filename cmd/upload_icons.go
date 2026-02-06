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
	uploadDryRun    bool
	uploadCatalog   string
	s3Endpoint      string
	s3AccessKey     string
	s3SecretKey     string
	s3Region        string
)

var uploadIconsCmd = &cobra.Command{
	Use:   "upload-icons",
	Short: "Upload item icons to Supabase and update database",
	Long: `Parses HTML files from catalogs/d2/pages to find item-image mappings,
uploads images from catalogs/d2/icons to Supabase Storage,
and updates database records with image URLs.

Process:
1. Parse HTML files (uniques.html, sets.html, base.html, misc.html) to map item names to image paths
2. For each item in database, look up its image from the HTML mapping
3. Upload the image to Supabase Storage (or reuse if already uploaded)
4. Update the database with the public URL
5. Log any missing images that need to be downloaded manually

Examples:
  # Preview what would be uploaded
  lootstash-catalog upload-icons --dry-run

  # Upload images
  lootstash-catalog upload-icons`,
	RunE: runUploadIcons,
}

func init() {
	rootCmd.AddCommand(uploadIconsCmd)

	uploadIconsCmd.Flags().BoolVar(&uploadDryRun, "dry-run", false, "Preview without making changes")
	uploadIconsCmd.Flags().StringVar(&uploadCatalog, "catalog", "catalogs/d2", "Path to catalog folder (contains icons/ and pages/ subfolders)")

	// S3 configuration with local Supabase defaults
	uploadIconsCmd.Flags().StringVar(&s3Endpoint, "s3-endpoint", "http://127.0.0.1:54321/storage/v1/s3", "S3 endpoint URL")
	uploadIconsCmd.Flags().StringVar(&s3AccessKey, "s3-access-key", "625729a08b95bf1b7ff351a663f3a23c", "S3 access key")
	uploadIconsCmd.Flags().StringVar(&s3SecretKey, "s3-secret-key", "850181e4652dd023b7a98c58ae0d2d34bd487ee0cc3254aed6eda37307425907", "S3 secret key")
	uploadIconsCmd.Flags().StringVar(&s3Region, "s3-region", "local", "S3 region")
}

func runUploadIcons(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if uploadDryRun {
		PrintInfo("Running in DRY-RUN mode")
	}

	// Connect to database
	PrintInfo("Connecting to database...")
	db, err := database.NewConnection(ctx, GetDatabaseURL())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	PrintSuccess("Connected to database")

	// Initialize S3 storage
	PrintInfo("Connecting to S3 storage...")
	s3Storage, err := storage.NewS3Storage(
		s3Endpoint,
		s3AccessKey,
		s3SecretKey,
		s3Region,
		"d2-items",                    // bucket name
		"http://127.0.0.1:54321",      // public URL base
	)
	if err != nil {
		return fmt.Errorf("failed to create S3 storage: %w", err)
	}
	PrintSuccess("S3 storage initialized")

	// Create uploader
	repo := d2.NewRepository(db.Pool())
	uploader := d2.NewIconUploader(repo, s3Storage, uploadDryRun)

	// Run upload
	stats, err := uploader.Upload(ctx, uploadCatalog)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	// Print results
	fmt.Println()
	if uploadDryRun {
		PrintSuccess("Dry run completed!")
	} else {
		PrintSuccess("Upload completed!")
	}

	fmt.Println("\nStatistics:")
	fmt.Printf("  Total DB items:   %d\n", stats.TotalDBItems)
	fmt.Printf("  Uploaded (new):   %d\n", stats.Uploaded)
	fmt.Printf("  Reused (cache):   %d\n", stats.ReusedCache)
	fmt.Printf("  Matched unique:   %d\n", stats.MatchedUnique)
	fmt.Printf("  Matched set:      %d\n", stats.MatchedSet)
	fmt.Printf("  Matched base:     %d\n", stats.MatchedBase)
	fmt.Printf("  Matched rune:     %d\n", stats.MatchedRune)
	fmt.Printf("  Matched gem:      %d\n", stats.MatchedGem)
	fmt.Printf("  Not in HTML:      %d\n", stats.NotInHTML)
	fmt.Printf("  Missing files:    %d\n", stats.MissingFiles)
	fmt.Printf("  Errors:           %d\n", stats.Errors)

	if len(stats.NotInHTMLItems) > 0 {
		fmt.Printf("\nItems not found in HTML files (first %d):\n", len(stats.NotInHTMLItems))
		for _, item := range stats.NotInHTMLItems {
			fmt.Printf("  - %s\n", item)
		}
	}

	if len(stats.MissingImages) > 0 {
		fmt.Printf("\nMissing image files - add these to icons folder (first %d):\n", len(stats.MissingImages))
		for _, img := range stats.MissingImages {
			fmt.Printf("  - %s\n", img)
		}
	}

	return nil
}
