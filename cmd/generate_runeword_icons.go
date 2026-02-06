package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/database"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/storage"
	"github.com/spf13/cobra"
)

var (
	generateDryRun  bool
	generateForce   bool
	generateCatalog string
	// S3 flags are reused from upload_icons.go
)

var generateRunewordIconsCmd = &cobra.Command{
	Use:   "generate-runeword-icons",
	Short: "Generate composite runeword icons by stacking individual rune images",
	Long: `Creates composite images for runewords by stacking individual rune icons.

For example, Enigma (Jah + Ith + Ber) will have an image with:
- Jah rune icon at top
- Ith rune icon in middle
- Ber rune icon at bottom

Layout adapts based on rune count:
- 1-3 runes: Vertical stack
- 4 runes: 2x2 grid
- 5 runes: 2x2 grid + center below
- 6 runes: 2x3 grid

Examples:
  # Preview what would be generated
  lootstash-catalog generate-runeword-icons --dry-run

  # Generate missing runeword icons
  lootstash-catalog generate-runeword-icons --catalog catalogs/d2

  # Force regenerate all (including existing)
  lootstash-catalog generate-runeword-icons --force`,
	RunE: runGenerateRunewordIcons,
}

func init() {
	rootCmd.AddCommand(generateRunewordIconsCmd)

	generateRunewordIconsCmd.Flags().BoolVar(&generateDryRun, "dry-run", false, "Preview without uploading")
	generateRunewordIconsCmd.Flags().BoolVar(&generateForce, "force", false, "Regenerate all runeword images (including existing)")
	generateRunewordIconsCmd.Flags().StringVar(&generateCatalog, "catalog", "catalogs/d2", "Path to catalog folder (contains icons/ subfolder)")

	// S3 configuration - reuse the same flag names as upload-icons
	generateRunewordIconsCmd.Flags().StringVar(&s3Endpoint, "s3-endpoint", "http://127.0.0.1:54321/storage/v1/s3", "S3 endpoint URL")
	generateRunewordIconsCmd.Flags().StringVar(&s3AccessKey, "s3-access-key", "625729a08b95bf1b7ff351a663f3a23c", "S3 access key")
	generateRunewordIconsCmd.Flags().StringVar(&s3SecretKey, "s3-secret-key", "850181e4652dd023b7a98c58ae0d2d34bd487ee0cc3254aed6eda37307425907", "S3 secret key")
	generateRunewordIconsCmd.Flags().StringVar(&s3Region, "s3-region", "local", "S3 region")
}

func runGenerateRunewordIcons(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if generateDryRun {
		PrintInfo("Running in DRY-RUN mode")
	}
	if generateForce {
		PrintInfo("Force mode enabled - will regenerate all runeword images")
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
		"d2-items",               // bucket name
		"http://127.0.0.1:54321", // public URL base
	)
	if err != nil {
		return fmt.Errorf("failed to create S3 storage: %w", err)
	}
	PrintSuccess("S3 storage initialized")

	// Create generator
	repo := d2.NewRepository(db.Pool())
	iconsPath := filepath.Join(generateCatalog, "icons")
	generator := d2.NewRunewordImageGenerator(repo, s3Storage, iconsPath, generateDryRun, generateForce)

	// Run generation
	stats, err := generator.Generate(ctx)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// Print results
	fmt.Println()
	if generateDryRun {
		PrintSuccess("Dry run completed!")
	} else {
		PrintSuccess("Generation completed!")
	}

	fmt.Println("\nStatistics:")
	fmt.Printf("  Total runewords:  %d\n", stats.TotalRunewords)
	fmt.Printf("  Generated:        %d\n", stats.Generated)
	fmt.Printf("  Skipped:          %d\n", stats.Skipped)
	fmt.Printf("  Missing runes:    %d\n", stats.MissingRunes)
	fmt.Printf("  Failed:           %d\n", stats.Failed)

	if len(stats.Errors) > 0 {
		maxErrors := 20
		if len(stats.Errors) < maxErrors {
			maxErrors = len(stats.Errors)
		}
		fmt.Printf("\nErrors (first %d):\n", maxErrors)
		for i := 0; i < maxErrors; i++ {
			fmt.Printf("  - %s\n", stats.Errors[i])
		}
		if len(stats.Errors) > maxErrors {
			fmt.Printf("  ... and %d more\n", len(stats.Errors)-maxErrors)
		}
	}

	return nil
}
