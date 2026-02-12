package d2

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// classSuffixRegex matches trailing class suffixes like "(Warlock only)", "(Amazon Only)", etc.
var classSuffixRegex = regexp.MustCompile(`\s*\((Amazon|Sorceress|Necromancer|Paladin|Barbarian|Druid|Assassin|Warlock)(\s+[Oo]nly)?\)\s*$`)

// ReverseTranslator converts display text back to Property structs
type ReverseTranslator struct {
	patterns         []reversePattern
	reverseSkillTabs map[string]int // skill tab name -> tab number
}

type reversePattern struct {
	code    string
	regex   *regexp.Regexp
	groups  []string // names of capture groups: "value", "min", "max", "param"
	isFixed bool     // true for fixed-text properties with no placeholders
}

// NewReverseTranslator builds a reverse translator from the existing PropertyTranslator formats
func NewReverseTranslator() *ReverseTranslator {
	translator := NewPropertyTranslator()
	var patterns []reversePattern

	// Build patterns from all format templates
	for code, template := range translator.formats {
		rp := buildReversePattern(code, template)
		if rp != nil {
			patterns = append(patterns, *rp)
		}
	}

	// Add per-level code patterns
	for code := range perLevelCodes {
		// Per-level display: "(X Per Character Level) Y-Z To Stat (Based On Character Level)"
		// We match the simpler translator.formats version: "+{value} To Stat (Based On Character Level)"
		// which is already in translator.formats, so skip here
		_ = code
	}

	// Sort patterns by regex length (longest first) for specificity
	sort.Slice(patterns, func(i, j int) bool {
		if patterns[i].isFixed != patterns[j].isFixed {
			return patterns[i].isFixed // fixed patterns first (exact matches)
		}
		return len(patterns[i].regex.String()) > len(patterns[j].regex.String())
	})

	// Build reverse skill tab lookup from the translator's skillTabs map
	reverseSkillTabs := make(map[string]int)
	for tabNum, tabName := range translator.skillTabs {
		reverseSkillTabs[strings.ToLower(tabName)] = tabNum
	}
	// Add variant mappings for inconsistent HTML names
	reverseSkillTabs["psychic skill tab"] = 21

	return &ReverseTranslator{patterns: patterns, reverseSkillTabs: reverseSkillTabs}
}

// buildReversePattern converts a template like "+{value}% Enhanced Damage" into a regex pattern
func buildReversePattern(code, template string) *reversePattern {
	// Check if this is a fixed-text property (no placeholders)
	if !strings.Contains(template, "{") {
		return &reversePattern{
			code:    code,
			regex:   regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(template) + `$`),
			isFixed: true,
		}
	}

	// Escape regex special chars in the template
	escaped := regexp.QuoteMeta(template)

	// Track which groups we have
	var groups []string

	// Replace escaped placeholders with capture groups
	// Note: QuoteMeta turns {value} into \{value\}
	if strings.Contains(escaped, `\{value\}`) {
		// Value can be a single number, range like "5-10", or negative range like "-(5-10)"
		escaped = strings.Replace(escaped, `\{value\}`, `([+-]?\d+(?:-\d+)?(?:\(\d+-\d+\))?)`, 1)
		groups = append(groups, "value")
	}
	// Handle +{value} pattern (the + is literal in template, but value might be negative)
	if strings.Contains(escaped, `\+`+`\{value\}`) {
		// Already handled above since we do global replacement
	}

	if strings.Contains(escaped, `\{min\}`) {
		escaped = strings.Replace(escaped, `\{min\}`, `(\d+)`, 1)
		groups = append(groups, "min")
	}
	if strings.Contains(escaped, `\{max\}`) {
		escaped = strings.Replace(escaped, `\{max\}`, `(\d+)`, 1)
		groups = append(groups, "max")
	}
	if strings.Contains(escaped, `\{param\}`) {
		escaped = strings.Replace(escaped, `\{param\}`, `(.+?)`, 1)
		groups = append(groups, "param")
	}
	if strings.Contains(escaped, `\{skilltab\}`) {
		escaped = strings.Replace(escaped, `\{skilltab\}`, `(.+?)`, 1)
		groups = append(groups, "skilltab")
	}

	// Handle the +{value} case where the + sign is part of the template
	// The value might already include the sign, so make the leading + optional
	escaped = strings.ReplaceAll(escaped, `\+([+-]`, `[+]?([+-]`)

	regex, err := regexp.Compile(`(?i)^` + escaped + `$`)
	if err != nil {
		return nil
	}

	return &reversePattern{
		code:   code,
		regex:  regex,
		groups: groups,
	}
}

// ReverseTranslate converts display text back to a Property
func (rt *ReverseTranslator) ReverseTranslate(displayText string) Property {
	displayText = strings.TrimSpace(displayText)
	if displayText == "" {
		return Property{Code: "raw", DisplayText: displayText}
	}

	// Try per-level pattern first: "(X Per Character Level) Y-Z To Stat (Based On Character Level)"
	if prop, ok := rt.tryPerLevelMatch(displayText); ok {
		return prop
	}

	// Try each pattern
	for _, p := range rt.patterns {
		matches := p.regex.FindStringSubmatch(displayText)
		if matches == nil {
			continue
		}

		prop := Property{Code: p.code}

		if p.isFixed {
			// Fixed-text property, no values to extract
			prop.DisplayText = displayText
			return prop
		}

		// Extract values from capture groups
		skilltabUnresolved := false
		for i, groupName := range p.groups {
			if i+1 >= len(matches) {
				break
			}
			val := matches[i+1]

			switch groupName {
			case "value":
				min, max := parseValueStr(val)
				prop.Min = min
				prop.Max = max
			case "min":
				prop.Min, _ = strconv.Atoi(val)
			case "max":
				prop.Max, _ = strconv.Atoi(val)
			case "param":
				// Strip class suffixes like "(Warlock only)" from skill params
				prop.Param = classSuffixRegex.ReplaceAllString(val, "")
			case "skilltab":
				// Strip class suffixes and resolve to tab number
				cleaned := classSuffixRegex.ReplaceAllString(val, "")
				if tabNum, ok := rt.reverseSkillTabs[strings.ToLower(cleaned)]; ok {
					prop.Param = fmt.Sprintf("%d", tabNum)
				} else {
					// Not a known skill tab — skip this match so other patterns
					// (like "skill") can try instead
					skilltabUnresolved = true
				}
			}
		}

		// If skilltab didn't resolve, this wasn't actually a skilltab property
		if skilltabUnresolved {
			continue
		}

		prop.DisplayText = displayText
		return prop
	}

	// No match found — return as raw property
	return Property{Code: "raw", DisplayText: displayText}
}

// tryPerLevelMatch handles the per-level display format:
// "(X Per Character Level) Y-Z To Stat (Based On Character Level)"
func (rt *ReverseTranslator) tryPerLevelMatch(text string) (Property, bool) {
	if !strings.Contains(text, "Based On Character Level") && !strings.Contains(text, "Based on Character Level") {
		return Property{}, false
	}

	// Match the per-level format with parenthesized per-level value
	// e.g., "(1.5 Per Character Level) 1-148 To Life (Based On Character Level)"
	perLevelRegex := regexp.MustCompile(`(?i)^\(([0-9.]+) Per Character Level\)\s+(\d+)-(\d+)\s+(.+?)\s+\(Based [Oo]n Character Level\)$`)
	matches := perLevelRegex.FindStringSubmatch(text)
	if matches != nil {
		statText := matches[4]
		// Find which per-level code matches this stat text
		for code, template := range perLevelCodes {
			// Extract the stat part from the template (between min/max and "Based On")
			// Template: "({perLevel} Per Character Level) {lvlMin}-{lvlMax} To Life (Based On Character Level)"
			templateStatRegex := regexp.MustCompile(`\{lvlMin\}-\{lvlMax\}\s+(.+?)\s+\(Based`)
			templateMatches := templateStatRegex.FindStringSubmatch(template)
			if templateMatches != nil {
				expectedStat := templateMatches[1]
				if strings.EqualFold(statText, expectedStat) {
					// Reverse-engineer the raw value from perLevel
					perLevel, _ := strconv.ParseFloat(matches[1], 64)
					raw := int(perLevel * 8)
					return Property{
						Code:        code,
						Min:         raw,
						Max:         raw,
						DisplayText: text,
					}, true
				}
			}
		}
	}

	// Also match the simpler format without parenthesized per-level value
	// e.g., "+1 To Maximum Damage (Based On Character Level)"
	simplePerLevelRegex := regexp.MustCompile(`(?i)^[+]?(-?\d+(?:-\d+)?)\s+(.+?)\s+\(Based [Oo]n Character Level\)$`)
	simpleMatches := simplePerLevelRegex.FindStringSubmatch(text)
	if simpleMatches != nil {
		valueStr := simpleMatches[1]
		statText := simpleMatches[2]

		for code, template := range perLevelCodes {
			templateStatRegex := regexp.MustCompile(`\{lvlMin\}-\{lvlMax\}\s+(.+?)\s+\(Based`)
			templateMatches := templateStatRegex.FindStringSubmatch(template)
			if templateMatches != nil {
				expectedStat := templateMatches[1]
				if strings.EqualFold(statText, expectedStat) {
					min, max := parseValueStr(valueStr)
					return Property{
						Code:        code,
						Min:         min,
						Max:         max,
						DisplayText: text,
					}, true
				}
			}
		}

		// Also check against simple formats from translator.formats
		translator := NewPropertyTranslator()
		for code, tmpl := range translator.formats {
			if !strings.HasSuffix(code, "/lvl") {
				continue
			}
			// Template: "+{value} To Life (Based On Character Level)"
			tmplRegex := regexp.MustCompile(`\{value\}[%]?\s+(.+?)\s+\(Based`)
			tmplMatches := tmplRegex.FindStringSubmatch(tmpl)
			if tmplMatches != nil {
				expectedStat := tmplMatches[1]
				if strings.EqualFold(statText, expectedStat) {
					min, max := parseValueStr(valueStr)
					return Property{
						Code:        code,
						Min:         min,
						Max:         max,
						DisplayText: text,
					}, true
				}
			}
		}
	}

	return Property{}, false
}

// parseValueStr parses a value string that can be:
// - "25" -> min=25, max=25
// - "25-35" -> min=25, max=35
// - "+25" -> min=25, max=25
// - "-(5-10)" -> min=-10, max=-5
func parseValueStr(s string) (int, int) {
	s = strings.TrimSpace(s)

	// Handle negative range: -(5-10)
	if strings.HasPrefix(s, "-(") && strings.HasSuffix(s, ")") {
		inner := s[2 : len(s)-1]
		parts := strings.SplitN(inner, "-", 2)
		if len(parts) == 2 {
			a, _ := strconv.Atoi(parts[0])
			b, _ := strconv.Atoi(parts[1])
			return -b, -a
		}
	}

	// Handle positive range: 25-35 or +25-35
	s = strings.TrimPrefix(s, "+")
	if strings.Contains(s, "-") && !strings.HasPrefix(s, "-") {
		parts := strings.SplitN(s, "-", 2)
		if len(parts) == 2 {
			a, _ := strconv.Atoi(parts[0])
			b, _ := strconv.Atoi(parts[1])
			return a, b
		}
	}

	// Handle negative number with range: -5-10 (meaning -5 to -10, but displayed as value)
	if strings.HasPrefix(s, "-") {
		// Could be just a negative number like "-25"
		val, err := strconv.Atoi(s)
		if err == nil {
			return val, val
		}
	}

	// Simple integer
	val, _ := strconv.Atoi(s)
	return val, val
}

// ReverseTranslateLines converts multiple display text lines to Properties
func (rt *ReverseTranslator) ReverseTranslateLines(lines []string) []Property {
	var props []Property
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		prop := rt.ReverseTranslate(line)
		props = append(props, prop)
	}
	return props
}

