package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
	"github.com/spf13/cobra"
)

var extractURLsCmd = &cobra.Command{
	Use:   "extract-urls",
	Short: "Extract image URLs from HTML files",
	Long: `Extracts all image URLs from the diablo2.io HTML files and saves them to a text file.

You can then use this file with wget, curl, or a download manager to batch download the images.

Example workflow:
  1. Extract URLs:
     lootstash-catalog extract-urls

  2. Download with wget (if you have cookies):
     wget -i image-urls.txt -P ./images --header="Cookie: cf_clearance=xxx"

  3. Or import image-urls.txt into a download manager like Free Download Manager`,
	RunE: runExtractURLs,
}

func init() {
	rootCmd.AddCommand(extractURLsCmd)
}

func runExtractURLs(cmd *cobra.Command, args []string) error {
	parser := d2.NewHTMLParser()
	baseURL := "https://diablo2.io"

	htmlFiles := []string{
		"catalogs/d2/icons/uniques.html",
		"catalogs/d2/icons/sets.html",
		"catalogs/d2/icons/base.html",
		"catalogs/d2/icons/misc.html",
	}

	// Use a map to deduplicate URLs
	urlSet := make(map[string]bool)

	for _, htmlFile := range htmlFiles {
		PrintInfo(fmt.Sprintf("Parsing %s...", htmlFile))

		items, err := parser.ParseFile(htmlFile)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to parse %s: %v", htmlFile, err))
			continue
		}

		for _, item := range items {
			if item.ImagePath != "" {
				fullURL := baseURL + item.ImagePath
				urlSet[fullURL] = true
			}
		}

		PrintSuccess(fmt.Sprintf("Found %d items in %s", len(items), filepath.Base(htmlFile)))
	}

	// Convert to sorted slice
	var urls []string
	for url := range urlSet {
		urls = append(urls, url)
	}
	sort.Strings(urls)

	// Write to file
	outputFile := "image-urls.txt"
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	for _, url := range urls {
		fmt.Fprintln(f, url)
	}

	PrintSuccess(fmt.Sprintf("Extracted %d unique image URLs to %s", len(urls), outputFile))

	fmt.Println("\nNext steps:")
	fmt.Println("  1. Open diablo2.io in your browser")
	fmt.Println("  2. Open DevTools (F12) > Network tab")
	fmt.Println("  3. Copy a cookie value (especially cf_clearance)")
	fmt.Println("  4. Download with wget:")
	fmt.Println("     wget -i image-urls.txt -P ./images --header=\"Cookie: YOUR_COOKIE_HERE\"")
	fmt.Println("\n  Or use Free Download Manager (fdm) to import the URL list")

	return nil
}
