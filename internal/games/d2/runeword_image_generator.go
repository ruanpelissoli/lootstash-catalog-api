package d2

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/storage"
)

// GenerateStats tracks runeword image generation statistics
type GenerateStats struct {
	TotalRunewords int
	Generated      int
	Skipped        int
	MissingRunes   int
	Failed         int
	Errors         []string
}

// RunewordImageGenerator handles generating composite runeword images
type RunewordImageGenerator struct {
	repo           *Repository
	storage        storage.Storage
	iconsPath      string
	dryRun         bool
	force          bool
	runeCodeToName map[string]string // "r30" -> "Ber"
}

// NewRunewordImageGenerator creates a new runeword image generator
func NewRunewordImageGenerator(repo *Repository, stor storage.Storage, iconsPath string, dryRun, force bool) *RunewordImageGenerator {
	return &RunewordImageGenerator{
		repo:           repo,
		storage:        stor,
		iconsPath:      iconsPath,
		dryRun:         dryRun,
		force:          force,
		runeCodeToName: make(map[string]string),
	}
}

// Generate creates composite images for runewords and uploads them
func (g *RunewordImageGenerator) Generate(ctx context.Context) (*GenerateStats, error) {
	stats := &GenerateStats{}

	// Build rune code to name mapping
	fmt.Println("Loading rune code mappings...")
	runeMap, err := g.repo.GetRuneCodeToNameMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rune mappings: %w", err)
	}
	g.runeCodeToName = runeMap
	fmt.Printf("  Loaded %d rune mappings\n", len(runeMap))

	// Get runewords to process
	fmt.Println("Loading runewords...")
	var runewords []RunewordWithRunes
	if g.force {
		runewords, err = g.repo.GetAllRunewords(ctx)
	} else {
		runewords, err = g.repo.GetRunewordsWithoutImages(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get runewords: %w", err)
	}
	stats.TotalRunewords = len(runewords)
	fmt.Printf("  Found %d runewords to process\n", len(runewords))

	if len(runewords) == 0 {
		fmt.Println("No runewords need image generation.")
		return stats, nil
	}

	// Process each runeword
	fmt.Println("\nGenerating runeword images...")
	for _, rw := range runewords {
		if err := g.processRuneword(ctx, rw, stats); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("%s: %v", rw.DisplayName, err))
			stats.Failed++
		}
	}

	return stats, nil
}

// processRuneword generates and uploads an image for a single runeword
func (g *RunewordImageGenerator) processRuneword(ctx context.Context, rw RunewordWithRunes, stats *GenerateStats) error {
	// Load rune images
	runeImages, missingRunes := g.loadRuneImages(rw.Runes)
	if len(missingRunes) > 0 {
		stats.MissingRunes++
		stats.Errors = append(stats.Errors, fmt.Sprintf("%s: missing runes %v", rw.DisplayName, missingRunes))
		return nil // Skip but don't fail
	}

	if len(runeImages) == 0 {
		stats.MissingRunes++
		stats.Errors = append(stats.Errors, fmt.Sprintf("%s: no valid rune images", rw.DisplayName))
		return nil
	}

	// Create composite image
	compositeImg := g.createCompositeImage(runeImages)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, compositeImg); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	// Generate storage path
	storagePath := storage.StoragePath("d2/runeword", rw.DisplayName)

	if g.dryRun {
		runeNames := make([]string, len(rw.Runes))
		for i, code := range rw.Runes {
			runeNames[i] = g.runeCodeToName[code]
		}
		fmt.Printf("  [DRY-RUN] Would generate %s (%v) -> %s\n", rw.DisplayName, runeNames, storagePath)
		stats.Generated++
		return nil
	}

	// Upload to storage
	publicURL, err := g.storage.UploadImage(ctx, storagePath, buf.Bytes(), "image/png")
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	// Update database
	if err := g.repo.UpdateRunewordImageURL(ctx, rw.ID, publicURL); err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}

	runeNames := make([]string, len(rw.Runes))
	for i, code := range rw.Runes {
		runeNames[i] = g.runeCodeToName[code]
	}
	fmt.Printf("  ✓ %s (%v) -> %s\n", rw.DisplayName, runeNames, storagePath)
	stats.Generated++
	return nil
}

// loadRuneImages loads individual rune images for the given rune codes
func (g *RunewordImageGenerator) loadRuneImages(runeCodes []string) ([]image.Image, []string) {
	var images []image.Image
	var missing []string

	for _, code := range runeCodes {
		runeName, ok := g.runeCodeToName[code]
		if !ok {
			missing = append(missing, code)
			continue
		}

		// Map rune name to file name
		fileName := g.runeNameToFileName(runeName)
		filePath := filepath.Join(g.iconsPath, fileName)

		img, err := g.loadPNG(filePath)
		if err != nil {
			missing = append(missing, fmt.Sprintf("%s (%s)", runeName, fileName))
			continue
		}
		images = append(images, img)
	}

	return images, missing
}

// runeNameToFileName maps a rune name to its corresponding image file name
func (g *RunewordImageGenerator) runeNameToFileName(runeName string) string {
	// Special cases for file naming discrepancies
	fileNameMap := map[string]string{
		"Jah":  "Jo",   // Jah Rune uses "Jo" in file name
		"Shael": "Shae", // Shael uses "Shae" in file name
	}

	displayName := runeName
	if mapped, ok := fileNameMap[runeName]; ok {
		displayName = mapped
	}

	return fmt.Sprintf("rune%s_graphic.png", displayName)
}

// loadPNG loads a PNG file and returns the decoded image
func (g *RunewordImageGenerator) loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// createCompositeImage combines rune images using adaptive layouts
func (g *RunewordImageGenerator) createCompositeImage(images []image.Image) image.Image {
	if len(images) == 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	padding := 2

	switch len(images) {
	case 1, 2, 3:
		return g.createVerticalStack(images, padding)
	case 4:
		return g.create2x2Grid(images, padding)
	case 5:
		return g.create2x2PlusCenter(images, padding)
	case 6:
		return g.create2x3Grid(images, padding)
	default:
		// For 7+ runes (unlikely), fall back to vertical stack
		return g.createVerticalStack(images, padding)
	}
}

// createVerticalStack stacks images vertically with padding
func (g *RunewordImageGenerator) createVerticalStack(images []image.Image, padding int) image.Image {
	if len(images) == 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	// Calculate dimensions
	maxWidth := 0
	totalHeight := 0
	for i, img := range images {
		bounds := img.Bounds()
		if bounds.Dx() > maxWidth {
			maxWidth = bounds.Dx()
		}
		totalHeight += bounds.Dy()
		if i > 0 {
			totalHeight += padding
		}
	}

	// Create destination image
	dest := image.NewRGBA(image.Rect(0, 0, maxWidth, totalHeight))

	// Draw each rune centered horizontally
	y := 0
	for i, img := range images {
		if i > 0 {
			y += padding
		}
		bounds := img.Bounds()
		x := (maxWidth - bounds.Dx()) / 2 // Center horizontally
		draw.Draw(dest, image.Rect(x, y, x+bounds.Dx(), y+bounds.Dy()), img, bounds.Min, draw.Over)
		y += bounds.Dy()
	}

	return dest
}

// create2x2Grid creates a 2x2 grid layout for 4 runes
func (g *RunewordImageGenerator) create2x2Grid(images []image.Image, padding int) image.Image {
	if len(images) < 4 {
		return g.createVerticalStack(images, padding)
	}

	// Get max dimensions
	maxW, maxH := g.getMaxDimensions(images[:4])

	// Create destination: 2 columns, 2 rows
	width := maxW*2 + padding
	height := maxH*2 + padding
	dest := image.NewRGBA(image.Rect(0, 0, width, height))

	// Position: [0][1]
	//           [2][3]
	positions := []image.Point{
		{0, 0},
		{maxW + padding, 0},
		{0, maxH + padding},
		{maxW + padding, maxH + padding},
	}

	for i := 0; i < 4; i++ {
		g.drawCentered(dest, images[i], positions[i], maxW, maxH)
	}

	return dest
}

// create2x2PlusCenter creates a 2x2 grid with 5th rune centered below
func (g *RunewordImageGenerator) create2x2PlusCenter(images []image.Image, padding int) image.Image {
	if len(images) < 5 {
		return g.create2x2Grid(images, padding)
	}

	// Get max dimensions
	maxW, maxH := g.getMaxDimensions(images[:5])

	// Create destination: 2 columns, 3 rows (but 5th centered in row 3)
	width := maxW*2 + padding
	height := maxH*3 + padding*2
	dest := image.NewRGBA(image.Rect(0, 0, width, height))

	// Position: [0][1]
	//           [2][3]
	//            [4]
	positions := []image.Point{
		{0, 0},
		{maxW + padding, 0},
		{0, maxH + padding},
		{maxW + padding, maxH + padding},
		{(width - maxW) / 2, maxH*2 + padding*2}, // Centered
	}

	for i := 0; i < 5; i++ {
		g.drawCentered(dest, images[i], positions[i], maxW, maxH)
	}

	return dest
}

// create2x3Grid creates a 2x3 grid layout for 6 runes
func (g *RunewordImageGenerator) create2x3Grid(images []image.Image, padding int) image.Image {
	if len(images) < 6 {
		return g.create2x2PlusCenter(images, padding)
	}

	// Get max dimensions
	maxW, maxH := g.getMaxDimensions(images[:6])

	// Create destination: 2 columns, 3 rows
	width := maxW*2 + padding
	height := maxH*3 + padding*2
	dest := image.NewRGBA(image.Rect(0, 0, width, height))

	// Position: [0][1]
	//           [2][3]
	//           [4][5]
	positions := []image.Point{
		{0, 0},
		{maxW + padding, 0},
		{0, maxH + padding},
		{maxW + padding, maxH + padding},
		{0, maxH*2 + padding*2},
		{maxW + padding, maxH*2 + padding*2},
	}

	for i := 0; i < 6; i++ {
		g.drawCentered(dest, images[i], positions[i], maxW, maxH)
	}

	return dest
}

// getMaxDimensions returns the maximum width and height from a slice of images
func (g *RunewordImageGenerator) getMaxDimensions(images []image.Image) (int, int) {
	maxW, maxH := 0, 0
	for _, img := range images {
		bounds := img.Bounds()
		if bounds.Dx() > maxW {
			maxW = bounds.Dx()
		}
		if bounds.Dy() > maxH {
			maxH = bounds.Dy()
		}
	}
	return maxW, maxH
}

// drawCentered draws an image centered within a cell
func (g *RunewordImageGenerator) drawCentered(dest *image.RGBA, src image.Image, pos image.Point, cellW, cellH int) {
	bounds := src.Bounds()
	x := pos.X + (cellW-bounds.Dx())/2
	y := pos.Y + (cellH-bounds.Dy())/2
	draw.Draw(dest, image.Rect(x, y, x+bounds.Dx(), y+bounds.Dy()), src, bounds.Min, draw.Over)
}
