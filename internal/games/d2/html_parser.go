package d2

import (
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParsedItem represents an item extracted from HTML
type ParsedItem struct {
	Name          string
	ImagePath     string // e.g., /styles/zulu/theme/images/items/file.png
	NormalizedKey string // lowercase, trimmed for matching
}

// HTMLParser parses diablo2.io HTML files to extract item information
type HTMLParser struct{}

// NewHTMLParser creates a new HTML parser
func NewHTMLParser() *HTMLParser {
	return &HTMLParser{}
}

// ParseFile parses an HTML file and returns all items found
func (p *HTMLParser) ParseFile(filePath string) ([]ParsedItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return nil, err
	}

	return p.parseDocument(doc), nil
}

// parseDocument extracts items from a goquery document
func (p *HTMLParser) parseDocument(doc *goquery.Document) []ParsedItem {
	var items []ParsedItem

	doc.Find("article.element-item").Each(func(i int, s *goquery.Selection) {
		item := p.parseArticle(s)
		if item.Name != "" && item.ImagePath != "" {
			items = append(items, item)
		}
	})

	return items
}

// parseArticle extracts item data from an article element
func (p *HTMLParser) parseArticle(s *goquery.Selection) ParsedItem {
	var item ParsedItem

	// Extract image path from any div with data-background-image attribute
	// This handles z-graphic, z-graphic-helm, z-graphic-2x3, etc.
	s.Find("div[data-background-image]").Each(func(i int, graphic *goquery.Selection) {
		if item.ImagePath == "" { // Take the first one found
			if bgImage, exists := graphic.Attr("data-background-image"); exists {
				// Skip ticon (thumbnail) images, prefer graphic images
				if !strings.Contains(bgImage, "_ticon") {
					item.ImagePath = normalizeImagePath(bgImage)
				}
			}
		}
	})

	// If no graphic image found, try to get any image (including ticon as fallback)
	if item.ImagePath == "" {
		s.Find("div[data-background-image]").Each(func(i int, graphic *goquery.Selection) {
			if item.ImagePath == "" {
				if bgImage, exists := graphic.Attr("data-background-image"); exists {
					item.ImagePath = normalizeImagePath(bgImage)
				}
			}
		})
	}

	// Extract item name from h3.z-sort-name > a
	s.Find("h3.z-sort-name a").Each(func(i int, link *goquery.Selection) {
		item.Name = strings.TrimSpace(link.Text())
	})

	item.NormalizedKey = NormalizeItemName(item.Name)

	return item
}

// normalizeImagePath converts backslashes to forward slashes and cleans up the path
func normalizeImagePath(path string) string {
	// Convert backslashes to forward slashes
	path = strings.ReplaceAll(path, "\\", "/")
	// Ensure it starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

// NormalizeItemName normalizes an item name for matching
func NormalizeItemName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	// Convert curly quotes to straight quotes (using Unicode code points)
	name = strings.ReplaceAll(name, "\u2018", "'") // Left single quote
	name = strings.ReplaceAll(name, "\u2019", "'") // Right single quote
	name = strings.ReplaceAll(name, "\u201C", "\"") // Left double quote
	name = strings.ReplaceAll(name, "\u201D", "\"") // Right double quote
	return name
}
