package d2

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// HTMLParsedUniqueItem represents a unique item extracted from HTML
type HTMLParsedUniqueItem struct {
	Name         string
	BaseName     string
	Quality      string // "Normal Unique", "Exceptional Unique", "Elite Unique"
	ReqLevel     int
	QualityLevel int
	Properties   []string // Raw property text lines
	ImagePath    string
}

// HTMLParsedSetItem represents a set item extracted from HTML
type HTMLParsedSetItem struct {
	Name         string
	BaseName     string
	Quality      string // "Normal Set", "Exceptional Set", "Elite Set"
	ReqLevel     int
	QualityLevel int
	Properties   []string // Raw property text lines
	SetBonuses   []HTMLSetBonus
	SetName      string
	ImagePath    string
}

// HTMLSetBonus represents a set item bonus from HTML
type HTMLSetBonus struct {
	Text     string
	ItemCount int // e.g., 2 for "(2 set items)"
}

// HTMLParsedFullSet represents a full set definition from HTML
type HTMLParsedFullSet struct {
	Name           string
	PartialBonuses []string // Text of partial set bonuses
	FullBonuses    []string // Text of full set bonuses
}

// HTMLVariantLink represents a link to a base item variant (normal/exceptional/elite)
type HTMLVariantLink struct {
	Name string
	Tier string // "Normal", "Exceptional", "Elite"
}

// HTMLParsedBaseItem represents a base item extracted from HTML
type HTMLParsedBaseItem struct {
	Name         string
	Quality      string // "Normal", "Exceptional", "Elite"
	TypeName     string // Primary type from hidden span, e.g., "Grimoires"
	TypeName2    string // Secondary type, e.g., "Shields"
	TypeTags     []string // All matched type tags from hidden span
	ImagePath    string
	URLSlug      string // From href, used for code generation
	VariantNames []HTMLVariantLink // Links to normal/exceptional/elite variants

	// Stats
	DefenseMin   int
	DefenseMax   int
	OneHMinDam   int
	OneHMaxDam   int
	TwoHMinDam   int
	TwoHMaxDam   int
	Speed        int
	Durability   int
	ReqStr       int
	ReqDex       int
	ReqLevel     int
	QualityLevel int
	MaxSockets   int
	RangeAdder   int
	InvWidth     int
	InvHeight    int
}

// HTMLParsedRune represents a rune item extracted from misc.html
type HTMLParsedRune struct {
	Name       string
	ImagePath  string
	Level      int
	RuneIndex  int
	WeaponMods []string // raw property text
	HelmMods   []string
	ShieldMods []string
}

// HTMLParsedGem represents a gem item extracted from misc.html
type HTMLParsedGem struct {
	Name       string
	ImagePath  string
	WeaponMods []string
	HelmMods   []string
	ShieldMods []string
}

// HTMLParsedMiscItem represents a miscellaneous item extracted from misc.html
type HTMLParsedMiscItem struct {
	Name         string
	ImagePath    string
	Description  string // e.g. "Terrorizes Act 2 when used"
	SubCategory  string // "Small Charm", "Large Charm", "Grand Charm", "Jewel", "Key", "Essence", etc.
}

// HTMLParsedRuneword represents a runeword extracted from HTML
type HTMLParsedRuneword struct {
	Name        string
	Runes       []string
	SocketCount int
	ReqLevel    int
	ValidTypes  []string
	Properties  []string // Raw property text lines
}

// HTMLItemParser parses detailed item data from diablo2.io HTML files
type HTMLItemParser struct{}

// NewHTMLItemParser creates a new HTML item parser
func NewHTMLItemParser() *HTMLItemParser {
	return &HTMLItemParser{}
}

// ParseUniquesFile parses uniques.html and returns all unique items
func (p *HTMLItemParser) ParseUniquesFile(filePath string) ([]HTMLParsedUniqueItem, error) {
	doc, err := p.openFile(filePath)
	if err != nil {
		return nil, err
	}

	var items []HTMLParsedUniqueItem
	doc.Find("article.element-item").Each(func(i int, s *goquery.Selection) {
		item := p.parseUniqueArticle(s)
		if item.Name != "" {
			items = append(items, item)
		}
	})

	return items, nil
}

// ParseSetsFile parses sets.html and returns set items and full sets
func (p *HTMLItemParser) ParseSetsFile(filePath string) ([]HTMLParsedSetItem, []HTMLParsedFullSet, error) {
	doc, err := p.openFile(filePath)
	if err != nil {
		return nil, nil, err
	}

	var setItems []HTMLParsedSetItem
	setMap := make(map[string]*HTMLParsedFullSet) // setName -> full set bonuses

	doc.Find("article.element-item").Each(func(i int, s *goquery.Selection) {
		// Check if this is a full set article (h4 contains "Full Set")
		h4 := s.Find("h4").First()
		if h4.Length() > 0 && strings.Contains(h4.Text(), "Full Set") {
			fs := p.parseFullSetArticle(s)
			if fs.Name != "" {
				setMap[fs.Name] = &fs
			}
			return
		}

		item := p.parseSetArticle(s)
		if item.Name != "" && item.SetName != "" {
			setItems = append(setItems, item)

			// Collect set bonuses for the full set
			if _, exists := setMap[item.SetName]; !exists {
				setMap[item.SetName] = &HTMLParsedFullSet{Name: item.SetName}
			}
		}
	})

	// Convert set map to slice
	var fullSets []HTMLParsedFullSet
	for _, fs := range setMap {
		fullSets = append(fullSets, *fs)
	}

	return setItems, fullSets, nil
}

// ParseBasesFile parses base.html and returns all base items
func (p *HTMLItemParser) ParseBasesFile(filePath string) ([]HTMLParsedBaseItem, error) {
	doc, err := p.openFile(filePath)
	if err != nil {
		return nil, err
	}

	var items []HTMLParsedBaseItem
	doc.Find("article.element-item").Each(func(i int, s *goquery.Selection) {
		item := p.parseBaseArticle(s)
		if item.Name != "" {
			items = append(items, item)
		}
	})

	return items, nil
}

// ParseRunewordsFile parses runewords.html and returns all runewords
func (p *HTMLItemParser) ParseRunewordsFile(filePath string) ([]HTMLParsedRuneword, error) {
	doc, err := p.openFile(filePath)
	if err != nil {
		return nil, err
	}

	var runewords []HTMLParsedRuneword
	doc.Find("article.element-item").Each(func(i int, s *goquery.Selection) {
		rw := p.parseRunewordArticle(s)
		if rw.Name != "" && len(rw.Runes) > 0 {
			runewords = append(runewords, rw)
		}
	})

	return runewords, nil
}

// ParseMiscFile parses misc.html and returns runes, gems, and misc items
func (p *HTMLItemParser) ParseMiscFile(filePath string) ([]HTMLParsedRune, []HTMLParsedGem, []HTMLParsedMiscItem, error) {
	doc, err := p.openFile(filePath)
	if err != nil {
		return nil, nil, nil, err
	}

	var runes []HTMLParsedRune
	var gems []HTMLParsedGem
	var miscItems []HTMLParsedMiscItem

	doc.Find("article.element-item").Each(func(i int, s *goquery.Selection) {
		// Read h4 text to determine item type
		h4 := s.Find("h4").First()
		if h4.Length() == 0 {
			return
		}
		h4Text := strings.TrimSpace(h4.Text())

		switch {
		case strings.Contains(h4Text, "Rune"):
			rn := p.parseRuneArticle(s, h4)
			if rn.Name != "" {
				runes = append(runes, rn)
			}
		case h4Text == "Gem":
			gem := p.parseGemArticle(s)
			if gem.Name != "" {
				gems = append(gems, gem)
			}
		case h4Text == "Miscellaneous Item",
			h4Text == "Small Charm", h4Text == "Large Charm", h4Text == "Grand Charm",
			h4Text == "Jewel", h4Text == "Key", h4Text == "Essence":
			item := p.parseMiscArticle(s)
			if item.Name != "" {
				// Set subcategory based on h4 text
				switch h4Text {
				case "Small Charm":
					item.SubCategory = "Small Charm"
				case "Large Charm":
					item.SubCategory = "Large Charm"
				case "Grand Charm":
					item.SubCategory = "Grand Charm"
				case "Jewel":
					item.SubCategory = "Jewel"
				case "Key":
					item.SubCategory = "Key"
				case "Essence":
					item.SubCategory = "Essence"
				default:
					item.SubCategory = "Miscellaneous"
				}
				miscItems = append(miscItems, item)
			}
		// Skip: Quest Item, Potion, Consumable, Crafted Item, etc.
		}
	})

	return runes, gems, miscItems, nil
}

// parseRuneArticle extracts rune data from an article element in misc.html
func (p *HTMLItemParser) parseRuneArticle(s *goquery.Selection, h4 *goquery.Selection) HTMLParsedRune {
	var rn HTMLParsedRune

	rn.Name = p.extractName(s)
	if rn.Name == "" {
		return rn
	}

	rn.ImagePath = p.extractImagePath(s)
	rn.Level = p.extractSpanInt(h4, "zso_runelevel")
	rn.RuneIndex = p.extractSpanInt(h4, "zso_runeindex")

	rn.WeaponMods, rn.HelmMods, rn.ShieldMods = p.parseMiscModSections(s)

	return rn
}

// parseGemArticle extracts gem data from an article element in misc.html
func (p *HTMLItemParser) parseGemArticle(s *goquery.Selection) HTMLParsedGem {
	var gem HTMLParsedGem

	gem.Name = p.extractName(s)
	if gem.Name == "" {
		return gem
	}

	gem.ImagePath = p.extractImagePath(s)
	gem.WeaponMods, gem.HelmMods, gem.ShieldMods = p.parseMiscModSections(s)

	return gem
}

// parseMiscArticle extracts miscellaneous item data from an article element in misc.html
func (p *HTMLItemParser) parseMiscArticle(s *goquery.Selection) HTMLParsedMiscItem {
	var item HTMLParsedMiscItem

	item.Name = p.extractName(s)
	if item.Name == "" {
		return item
	}

	item.ImagePath = p.extractImagePath(s)

	// Extract description from the p.z-smallstats content (cleaned)
	firstStats := s.Find("p.z-smallstats").First()
	if firstStats.Length() > 0 {
		lines := p.extractPropertiesFromStats(firstStats)
		if len(lines) > 0 {
			item.Description = strings.Join(lines, "; ")
		}
	}

	return item
}

// parseMiscModSections extracts weapon/helm(armor)/shield mod text from the rune/gem mods div.
// The div contains a z-hr child and sections headed by z-white spans ("Weapons", "Armor", "Shields").
func (p *HTMLItemParser) parseMiscModSections(s *goquery.Selection) (weapon, helm, shield []string) {
	// Find the div.z-vf-hide that contains a div.z-hr child (the rune/gem mods div)
	var modsDiv *goquery.Selection
	s.Find("div.z-vf-hide").Each(func(i int, div *goquery.Selection) {
		if modsDiv != nil {
			return
		}
		if div.Find("div.z-hr").Length() > 0 {
			modsDiv = div
		}
	})
	if modsDiv == nil {
		return
	}

	// Determine current section by iterating children
	currentSection := ""
	modsDiv.Children().Each(func(i int, child *goquery.Selection) {
		nodeName := goquery.NodeName(child)

		// Check for section headers (span.z-white containing "Weapons", "Armor", "Shields")
		if nodeName == "span" && child.HasClass("z-white") {
			headerText := strings.TrimSpace(child.Text())
			switch {
			case strings.Contains(headerText, "Weapons"):
				currentSection = "weapon"
			case strings.Contains(headerText, "Armor"):
				currentSection = "helm"
			case strings.Contains(headerText, "Shields"):
				currentSection = "shield"
			}
			return
		}

		// Collect mod text from span.z-smallstats
		if nodeName == "span" && child.HasClass("z-smallstats") {
			text := strings.TrimSpace(child.Text())
			if text == "" {
				return
			}
			switch currentSection {
			case "weapon":
				weapon = append(weapon, text)
			case "helm":
				helm = append(helm, text)
			case "shield":
				shield = append(shield, text)
			}
		}
	})

	return
}

func (p *HTMLItemParser) openFile(filePath string) (*goquery.Document, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return goquery.NewDocumentFromReader(file)
}

// parseUniqueArticle extracts unique item data from an article element
func (p *HTMLItemParser) parseUniqueArticle(s *goquery.Selection) HTMLParsedUniqueItem {
	var item HTMLParsedUniqueItem

	// Extract name
	item.Name = p.extractName(s)
	if item.Name == "" {
		return item
	}

	// Extract image path
	item.ImagePath = p.extractImagePath(s)

	// Extract quality and base name from h4
	item.Quality, item.BaseName = p.extractQualityAndBase(s)

	// Extract req level and quality level from first p.z-smallstats
	firstStats := s.Find("p.z-smallstats").First()
	item.ReqLevel = p.extractSpanInt(firstStats, "zso_rqlevel")
	item.QualityLevel = p.extractSpanInt(firstStats, "zso_qualitylvl")

	// Extract properties from the first p.z-smallstats
	item.Properties = p.extractPropertiesFromStats(firstStats)

	return item
}

// parseSetArticle extracts set item data from an article element
func (p *HTMLItemParser) parseSetArticle(s *goquery.Selection) HTMLParsedSetItem {
	var item HTMLParsedSetItem

	// Extract name
	item.Name = p.extractName(s)
	if item.Name == "" {
		return item
	}

	// Extract image path
	item.ImagePath = p.extractImagePath(s)

	// Extract quality and base name
	item.Quality, item.BaseName = p.extractQualityAndBase(s)

	// Extract req level and quality level
	firstStats := s.Find("p.z-smallstats").First()
	item.ReqLevel = p.extractSpanInt(firstStats, "zso_rqlevel")
	item.QualityLevel = p.extractSpanInt(firstStats, "zso_qualitylvl")

	// Extract properties from first p.z-smallstats
	item.Properties = p.extractPropertiesFromStats(firstStats)

	// Extract set bonuses from second p.z-smallstats
	secondStats := s.Find("p.z-smallstats").Eq(1)
	if secondStats.Length() > 0 {
		item.SetBonuses = p.extractSetBonuses(secondStats)
	}

	// Extract set name from "Part of set:" h4.
	// The h4 containing "Part of set:" has inline <span> children (no DOM restructuring),
	// so the a.ajax_link is still inside the h4.
	s.Find("h4").Each(func(i int, h4 *goquery.Selection) {
		if item.SetName != "" {
			return
		}
		if strings.Contains(h4.Text(), "Part of set:") {
			h4.Find("a.ajax_link").Each(func(j int, a *goquery.Selection) {
				if item.SetName == "" {
					item.SetName = strings.TrimSpace(a.Text())
				}
			})
		}
	})

	return item
}

// parseRunewordArticle extracts runeword data from an article element
func (p *HTMLItemParser) parseRunewordArticle(s *goquery.Selection) HTMLParsedRuneword {
	var rw HTMLParsedRuneword

	// Extract name
	rw.Name = p.extractName(s)
	if rw.Name == "" {
		return rw
	}

	// Extract runes
	s.Find("span.z-recipes").Each(func(i int, span *goquery.Selection) {
		runeName := strings.TrimSpace(span.Text())
		if runeName != "" {
			rw.Runes = append(rw.Runes, runeName)
		}
	})

	// Extract socket count
	rw.SocketCount = p.extractSpanInt(s, "zso_rwsock")

	// Extract req level
	rw.ReqLevel = p.extractSpanInt(s, "zso_rwlvlrq")

	// Extract valid types from filter links
	s.Find(`a[href*="#filter="]`).Each(func(i int, a *goquery.Selection) {
		typeName := strings.TrimSpace(a.Text())
		// Collapse internal whitespace/newlines (e.g., "4 socket\n              Weapons" -> "4 socket Weapons")
		typeName = strings.Join(strings.Fields(typeName), " ")
		if typeName != "" {
			rw.ValidTypes = append(rw.ValidTypes, typeName)
		}
	})

	// Extract properties from span.z-smallstats inside the runeword details section
	// The runeword properties are in the z-vf-hide div after the runes
	s.Find("div.z-vf-hide span.z-smallstats").Each(func(i int, span *goquery.Selection) {
		html, _ := span.Html()
		lines := p.cleanPropertyHTML(html)
		if len(lines) > 0 {
			rw.Properties = append(rw.Properties, lines...)
		}
	})

	return rw
}

// parseBaseArticle extracts base item data from an article element
func (p *HTMLItemParser) parseBaseArticle(s *goquery.Selection) HTMLParsedBaseItem {
	var item HTMLParsedBaseItem

	// Extract name
	item.Name = p.extractName(s)
	if item.Name == "" {
		return item
	}

	// Extract URL slug from the href for code generation
	s.Find("h3.z-sort-name a").Each(func(i int, a *goquery.Selection) {
		if href, exists := a.Attr("href"); exists {
			item.URLSlug = extractSlugFromHref(href)
		}
	})

	// Extract image path
	item.ImagePath = p.extractImagePath(s)

	// Extract quality tier from h4
	h4 := s.Find("h4").First()
	if h4.Length() > 0 {
		h4Text := strings.TrimSpace(h4.Text())
		if strings.Contains(h4Text, "Exceptional") {
			item.Quality = "Exceptional"
		} else if strings.Contains(h4Text, "Elite") {
			item.Quality = "Elite"
		} else {
			item.Quality = "Normal"
		}
	}

	// Extract item types from hidden span
	item.TypeName, item.TypeName2, item.TypeTags = p.extractItemTypesFromHidden(s)

	// Extract variant names from "Variants:" section
	item.VariantNames = p.extractVariantNames(s)

	// Extract stats from p.z-smallstats
	firstStats := s.Find("p.z-smallstats").First()

	item.ReqStr = p.extractSpanInt(firstStats, "zso_rqstr")
	item.ReqDex = p.extractSpanInt(firstStats, "zso_rqdex")
	item.ReqLevel = p.extractSpanInt(firstStats, "zso_rqlevel")
	item.QualityLevel = p.extractSpanInt(firstStats, "zso_qualitylvl")
	item.Durability = p.extractSpanInt(firstStats, "zso_durability")
	item.MaxSockets = p.extractSpanInt(firstStats, "zso_maxsock")

	// Parse defense range
	item.DefenseMin, item.DefenseMax = p.extractSpanRange(firstStats, "zso_defense")

	// Parse damage ranges
	item.OneHMinDam, item.OneHMaxDam = p.extractSpanRange(firstStats, "zso_onehdamage")
	item.TwoHMinDam, item.TwoHMaxDam = p.extractSpanRange(firstStats, "zso_twohdamage")

	// Parse base speed
	speedText := ""
	firstStats.Find("span.zso_basespeed").Each(func(i int, span *goquery.Selection) {
		speedText = strings.TrimSpace(span.Text())
	})
	speedText = strings.Trim(speedText, "[]") // Remove brackets like [0]
	item.Speed, _ = strconv.Atoi(speedText)

	// Parse "Adds range" from text content
	statsHTML, _ := firstStats.Html()
	rangeRegex := regexp.MustCompile(`Adds range:</span>\s*(\d+)`)
	if matches := rangeRegex.FindStringSubmatch(statsHTML); matches != nil {
		item.RangeAdder, _ = strconv.Atoi(matches[1])
	}

	// Extract inventory size from graphic div class
	item.InvWidth, item.InvHeight = p.extractInventorySize(s)

	return item
}

// extractSlugFromHref extracts the URL slug from an href like "/base/blasphemous-grimoire-t1673953.html"
func extractSlugFromHref(href string) string {
	// Remove /base/ prefix
	slug := strings.TrimPrefix(href, "/base/")
	// Remove -tNNNN.html suffix
	slugRegex := regexp.MustCompile(`-t\d+\.html$`)
	slug = slugRegex.ReplaceAllString(slug, "")
	return slug
}

// extractItemTypesFromHidden extracts item type names from the hidden span at the start of the article.
// Returns (primaryType, secondaryType, allTypeTags).
func (p *HTMLItemParser) extractItemTypesFromHidden(s *goquery.Selection) (string, string, []string) {
	// Get text from the first direct child hidden span
	hiddenText := ""
	s.ChildrenFiltered("span.z-hidden").First().Each(func(i int, span *goquery.Selection) {
		hiddenText = span.Text()
	})

	// Remove socket info
	socketRegex := regexp.MustCompile(`(?i)can have \d+ sockets?`)
	hiddenText = socketRegex.ReplaceAllString(hiddenText, "")
	hiddenText = strings.TrimSpace(hiddenText)

	if hiddenText == "" {
		return "", "", nil
	}

	// Known type names — all types we can match
	knownTypes := []string{
		"Melee Weapons", "Missile Weapons", "Body Armor", "Amazon Weapons",
		"Druid Pelts", "Barbarian Helms", "Necromancer Shields", "Shrunken Heads",
		"Paladin Shields",
		"Grimoires", "Swords", "Axes", "Maces", "Polearms", "Staves",
		"Scepters", "Wands", "Bows", "Crossbows", "Daggers", "Throwing",
		"Javelins", "Spears", "Claws", "Orbs", "Hammers", "Clubs",
		"Circlets", "Targes",
		"Shields", "Helms", "Gloves", "Boots", "Belts", "Weapons",
	}

	// Find all matching types with their position in the text
	type posMatch struct {
		name string
		pos  int
	}
	var matches []posMatch
	for _, typeName := range knownTypes {
		pos := strings.Index(hiddenText, typeName)
		if pos >= 0 {
			matches = append(matches, posMatch{typeName, pos})
		}
	}

	// Sort by position (earliest in text = most specific type)
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].pos < matches[i].pos {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Collect all unique type tags
	var allTags []string
	seen := make(map[string]bool)
	for _, m := range matches {
		if !seen[m.name] {
			allTags = append(allTags, m.name)
			seen[m.name] = true
		}
	}

	type1, type2 := "", ""
	if len(allTags) >= 1 {
		type1 = allTags[0]
	}
	if len(allTags) >= 2 {
		type2 = allTags[1]
	}
	return type1, type2, allTags
}

// classSpecificFromTypeTags detects class-specific items from type tags
func classSpecificFromTypeTags(tags []string) string {
	for _, tag := range tags {
		switch tag {
		case "Amazon Weapons":
			return "amazon"
		case "Paladin Shields":
			return "paladin"
		case "Necromancer Shields", "Shrunken Heads":
			return "necromancer"
		case "Barbarian Helms":
			return "barbarian"
		case "Druid Pelts":
			return "druid"
		case "Claws":
			return "assassin"
		case "Grimoires":
			return "warlock"
		case "Orbs":
			return "sorceress"
		}
	}
	return ""
}

// extractSpanRange gets a range value like "103-148" or "3-8" from a span
func (p *HTMLItemParser) extractSpanRange(s *goquery.Selection, class string) (int, int) {
	text := ""
	s.Find("span." + class).Each(func(i int, span *goquery.Selection) {
		text = strings.TrimSpace(span.Text())
	})
	if text == "" || text == "0" {
		return 0, 0
	}

	// Normalize en-dash (U+2013) and em-dash (U+2014) to regular hyphen
	text = strings.ReplaceAll(text, "\u2013", "-")
	text = strings.ReplaceAll(text, "\u2014", "-")

	// Handle range like "103-148"
	parts := strings.SplitN(text, "-", 2)
	if len(parts) == 2 {
		min, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		max, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		return min, max
	}

	// Handle formats like "8 to 20 (14 Avg)"
	toRegex := regexp.MustCompile(`(\d+)\s+to\s+(\d+)`)
	if matches := toRegex.FindStringSubmatch(text); matches != nil {
		min, _ := strconv.Atoi(matches[1])
		max, _ := strconv.Atoi(matches[2])
		return min, max
	}

	// Single value
	val, _ := strconv.Atoi(text)
	return val, val
}

// extractInventorySize gets width and height from graphic div CSS class
func (p *HTMLItemParser) extractInventorySize(s *goquery.Selection) (int, int) {
	width, height := 1, 1
	s.Find("div[data-background-image]").Each(func(i int, div *goquery.Selection) {
		classes, exists := div.Attr("class")
		if !exists {
			return
		}

		// Check for z-graphic-NxM pattern
		sizeRegex := regexp.MustCompile(`z-graphic-(\d+)x(\d+)`)
		if matches := sizeRegex.FindStringSubmatch(classes); matches != nil {
			width, _ = strconv.Atoi(matches[1])
			height, _ = strconv.Atoi(matches[2])
			return
		}

		// z-graphic-helm is 2x2
		if strings.Contains(classes, "z-graphic-helm") {
			width = 2
			height = 2
		}
	})
	return width, height
}

func containsStr(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// extractName gets the item name from h3.z-sort-name > a
func (p *HTMLItemParser) extractName(s *goquery.Selection) string {
	name := ""
	s.Find("h3.z-sort-name a").Each(func(i int, a *goquery.Selection) {
		name = strings.TrimSpace(a.Text())
	})
	return name
}

// extractImagePath gets the image path from data-background-image, skipping ticons
func (p *HTMLItemParser) extractImagePath(s *goquery.Selection) string {
	var imagePath string
	s.Find("div[data-background-image]").Each(func(i int, div *goquery.Selection) {
		if imagePath != "" {
			return
		}
		if bg, exists := div.Attr("data-background-image"); exists {
			if !strings.Contains(bg, "_ticon") {
				imagePath = normalizeImagePath(bg)
			}
		}
	})
	return imagePath
}

// extractQualityAndBase gets quality text and base item name from h4
func (p *HTMLItemParser) extractQualityAndBase(s *goquery.Selection) (string, string) {
	h4 := s.Find("h4").First()
	if h4.Length() == 0 {
		return "", ""
	}

	// Get quality from h4 text (first text node before any links)
	quality := ""
	h4.Contents().Each(func(i int, node *goquery.Selection) {
		if goquery.NodeName(node) == "#text" {
			text := strings.TrimSpace(node.Text())
			if text != "" && quality == "" {
				quality = text
			}
		}
	})

	// Get base name from a[href*="/base/"] within the article.
	// Note: HTML5 parsing moves a.ajax_link outside h4 when a <div> (base icon)
	// appears inside h4, so we search the article element directly.
	baseName := ""
	s.Find(`a[href*="/base/"]`).Each(func(i int, a *goquery.Selection) {
		if baseName == "" {
			baseName = strings.TrimSpace(a.Text())
		}
	})

	// Fallback: some items (charms, jewels, rings, amulets) don't have a base link.
	// The base name appears as <span class="z-white"> inside the h4.
	if baseName == "" {
		h4.Find(`span.z-white`).Each(func(i int, span *goquery.Selection) {
			if baseName == "" {
				baseName = strings.TrimSpace(span.Text())
			}
		})
	}

	return quality, baseName
}

// extractSpanInt gets an integer value from a span with the given class
func (p *HTMLItemParser) extractSpanInt(s *goquery.Selection, class string) int {
	text := ""
	s.Find("span." + class).Each(func(i int, span *goquery.Selection) {
		text = strings.TrimSpace(span.Text())
	})
	val, _ := strconv.Atoi(text)
	return val
}

// extractPropertiesFromStats extracts property text lines from a p.z-smallstats element
func (p *HTMLItemParser) extractPropertiesFromStats(stats *goquery.Selection) []string {
	if stats.Length() == 0 {
		return nil
	}

	html, _ := stats.Html()
	return p.cleanPropertyHTML(html)
}

// cleanPropertyHTML converts HTML property text to clean text lines
func (p *HTMLItemParser) cleanPropertyHTML(html string) []string {
	if html == "" {
		return nil
	}

	// Replace <code> tags with their text content (e.g., <code class="z-trusty...">25-35</code> -> 25-35)
	codeRegex := regexp.MustCompile(`<code[^>]*>([^<]*)</code>`)
	html = codeRegex.ReplaceAllString(html, "$1")

	// Remove all span tags with known stat classes (these are base item stats, not properties)
	statClasses := []string{
		"zso_defense", "zso_throwdamage", "zso_twohdamage", "zso_onehdamage",
		"zso_basespeed", "zso_durability", "zso_rqstr", "zso_rqdex",
		"zso_rqlevel", "zso_qualitylvl", "zso_trclass", "zso_maxsock",
		"zso_baseblock",
	}

	// Remove lines containing stat span classes and their labels
	for _, cls := range statClasses {
		// Remove the span itself and any associated label
		spanRegex := regexp.MustCompile(`(?s)<span[^>]*class="[^"]*` + cls + `[^"]*"[^>]*>.*?</span>`)
		html = spanRegex.ReplaceAllString(html, "")
	}

	// Remove z-white label spans (like "Defense:", "Req level:", etc.)
	whiteSpanRegex := regexp.MustCompile(`<span[^>]*class="[^"]*z-white[^"]*"[^>]*>.*?</span>`)
	html = whiteSpanRegex.ReplaceAllString(html, "")

	// Remove z-hidden spans
	hiddenSpanRegex := regexp.MustCompile(`(?s)<span[^>]*class="[^"]*z-hidden[^"]*"[^>]*>.*?</span>`)
	html = hiddenSpanRegex.ReplaceAllString(html, "")

	// Remove z-grey spans (item count labels like "(2 set items)")
	greySpanRegex := regexp.MustCompile(`<span[^>]*class="[^"]*z-grey[^"]*"[^>]*>.*?</span>`)
	html = greySpanRegex.ReplaceAllString(html, "")

	// Remove z-sets-title spans but keep their content
	setsTitleRegex := regexp.MustCompile(`<span[^>]*class="[^"]*z-sets-title[^"]*"[^>]*>(.*?)</span>`)
	html = setsTitleRegex.ReplaceAllString(html, "$1")

	// Remove div tags but keep content
	html = regexp.MustCompile(`<div[^>]*>`).ReplaceAllString(html, "")
	html = strings.ReplaceAll(html, "</div>", "")

	// Remove remaining span tags but keep content
	html = regexp.MustCompile(`<span[^>]*>`).ReplaceAllString(html, "")
	html = strings.ReplaceAll(html, "</span>", "")

	// Remove a tags but keep content
	html = regexp.MustCompile(`<a[^>]*>`).ReplaceAllString(html, "")
	html = strings.ReplaceAll(html, "</a>", "")

	// Remove h4 tags but keep content
	html = regexp.MustCompile(`<h4[^>]*>`).ReplaceAllString(html, "")
	html = strings.ReplaceAll(html, "</h4>", "")

	// Remove p tags
	html = regexp.MustCompile(`<p[^>]*>`).ReplaceAllString(html, "")
	html = strings.ReplaceAll(html, "</p>", "")

	// Remove i tags
	html = regexp.MustCompile(`<i[^>]*>`).ReplaceAllString(html, "")
	html = strings.ReplaceAll(html, "</i>", "")

	// Remove HTML comments
	html = regexp.MustCompile(`<!--.*?-->`).ReplaceAllString(html, "")

	// Split on <br>, <br/>, <br />
	html = regexp.MustCompile(`<br\s*/?\s*>`).ReplaceAllString(html, "\n")

	// Remove any remaining HTML tags
	html = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(html, "")

	// Decode HTML entities
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")
	html = strings.ReplaceAll(html, "&nbsp;", " ")
	html = strings.ReplaceAll(html, "&#8211;", "-") // en-dash
	html = strings.ReplaceAll(html, "\u2013", "-")   // en-dash unicode

	// Split into lines and clean
	var lines []string
	for _, line := range strings.Split(html, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip lines that are just numbers (leftover stat values)
		if isJustNumber(line) {
			continue
		}
		// Skip lines that are base stat labels
		if isBaseStatLabel(line) {
			continue
		}
		// Skip "Part of set:" lines
		if strings.HasPrefix(line, "Part of set:") {
			continue
		}
		// Skip patch version lines
		if strings.HasPrefix(line, "Patch ") {
			continue
		}
		// Skip class block lines (including residual values after label span removal)
		if strings.HasPrefix(line, "Class block:") || strings.HasPrefix(line, "Weight:") {
			continue
		}
		if matched, _ := regexp.MatchString(`^(Ama|Sor|Nec|Pal|Bar|Dru|Ass|War):\s*\d+%?$`, line); matched {
			continue
		}
		// Skip filter lines
		if strings.Contains(line, "equip only") {
			continue
		}
		lines = append(lines, line)
	}

	return lines
}

// extractSetBonuses gets set bonus entries from the second p.z-smallstats
func (p *HTMLItemParser) extractSetBonuses(stats *goquery.Selection) []HTMLSetBonus {
	var bonuses []HTMLSetBonus

	stats.Find("span.z-sets-title").Each(func(i int, span *goquery.Selection) {
		text := strings.TrimSpace(span.Text())
		if text == "" {
			return
		}

		// Look for the "(N set items)" text in the adjacent z-grey span
		itemCount := 0
		nextGrey := span.Next()
		if nextGrey.Length() > 0 && nextGrey.HasClass("z-grey") {
			greyText := strings.TrimSpace(nextGrey.Text())
			countRegex := regexp.MustCompile(`\((\d+) set items?\)`)
			if matches := countRegex.FindStringSubmatch(greyText); matches != nil {
				itemCount, _ = strconv.Atoi(matches[1])
			}
		}

		bonuses = append(bonuses, HTMLSetBonus{
			Text:      text,
			ItemCount: itemCount,
		})
	})

	return bonuses
}

// isJustNumber checks if a string is just a number (leftover from stat spans)
func isJustNumber(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	_, err := strconv.Atoi(s)
	return err == nil
}

// parseFullSetArticle extracts full set bonus data from a full set article
func (p *HTMLItemParser) parseFullSetArticle(s *goquery.Selection) HTMLParsedFullSet {
	var fs HTMLParsedFullSet

	fs.Name = p.extractName(s)
	if fs.Name == "" {
		return fs
	}

	// Extract bonuses from p.z-smallstats sections
	// Partial bonuses have "(N set items)" labels, full set bonuses come after "Full Set"
	inFullSet := false
	s.Find("p.z-smallstats").Each(func(i int, stats *goquery.Selection) {
		html, _ := stats.Html()
		lines := p.cleanPropertyHTML(html)

		// Check if this section has "Full Set" header
		statsText := stats.Text()
		if strings.Contains(statsText, "Full Set") {
			inFullSet = true
		}

		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			// Skip "Full Set" label itself
			if strings.TrimSpace(line) == "Full Set" {
				inFullSet = true
				continue
			}
			if inFullSet {
				fs.FullBonuses = append(fs.FullBonuses, line)
			} else {
				fs.PartialBonuses = append(fs.PartialBonuses, line)
			}
		}
	})

	return fs
}

// extractVariantNames extracts variant links (normal/exceptional/elite) from the base item article
func (p *HTMLItemParser) extractVariantNames(s *goquery.Selection) []HTMLVariantLink {
	var variants []HTMLVariantLink

	// Look for "Variants:" text in the article
	s.Find("h4").Each(func(i int, h4 *goquery.Selection) {
		if !strings.Contains(h4.Text(), "Variant") {
			return
		}
		// Find links within or after this h4
		h4.Find("a").Each(func(j int, a *goquery.Selection) {
			name := strings.TrimSpace(a.Text())
			if name == "" {
				return
			}
			// Determine tier from h4 context or link text
			tier := "Normal"
			h4Text := h4.Text()
			if strings.Contains(h4Text, "Exceptional") {
				tier = "Exceptional"
			} else if strings.Contains(h4Text, "Elite") {
				tier = "Elite"
			}
			variants = append(variants, HTMLVariantLink{Name: name, Tier: tier})
		})
	})

	// Also look for variant links in the article body
	s.Find(`a[href*="/base/"]`).Each(func(i int, a *goquery.Selection) {
		// Skip the first base link (that's the item's own base, handled elsewhere)
		if i == 0 {
			return
		}
		name := strings.TrimSpace(a.Text())
		if name != "" && !containsVariant(variants, name) {
			variants = append(variants, HTMLVariantLink{Name: name})
		}
	})

	return variants
}

func containsVariant(variants []HTMLVariantLink, name string) bool {
	for _, v := range variants {
		if v.Name == name {
			return true
		}
	}
	return false
}

// isBaseStatLabel checks if text is a base item stat label (not a property)
func isBaseStatLabel(s string) bool {
	baseLabels := []string{
		"Defense:", "1H damage:", "2H damage:", "Throw damage:",
		"Base speed:", "Durability:", "Req Strength:", "Req Dexterity:",
		"Req level:", "Quality level:", "Treasure class:", "Max sockets:",
		"Base block:", "Class block:",
	}
	for _, label := range baseLabels {
		if strings.Contains(s, label) {
			return true
		}
	}
	return false
}
