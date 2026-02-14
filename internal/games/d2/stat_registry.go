package d2

import (
	"context"
	"fmt"
	"sync"
)

// StatRegistry is an in-memory cache backed by the d2.stats table.
// It seeds from FilterableStats() on first run and dynamically discovers
// new stat codes during HTML import.
type StatRegistry struct {
	repo  *Repository
	known map[string]bool
	mu    sync.Mutex
}

// NewStatRegistry creates a new stat registry backed by the given repository.
func NewStatRegistry(repo *Repository) *StatRegistry {
	return &StatRegistry{
		repo:  repo,
		known: make(map[string]bool),
	}
}

// Load loads all existing stat codes from the database into memory.
func (sr *StatRegistry) Load(ctx context.Context) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	codes, err := sr.repo.GetAllStatCodes(ctx)
	if err != nil {
		return fmt.Errorf("load stat codes: %w", err)
	}
	sr.known = codes
	return nil
}

// SeedFromFilterableStats seeds the stats table from the hardcoded FilterableStats list.
// Returns the number of stats seeded.
func (sr *StatRegistry) SeedFromFilterableStats(ctx context.Context) (int, error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	categoryOrder := make(map[string]int)
	for i, cat := range StatCategories {
		categoryOrder[cat] = i * 100
	}

	stats := FilterableStats()
	seeded := 0
	for i, sc := range stats {
		sortOrder := categoryOrder[sc.Category] + i

		stat := &Stat{
			Code:         sc.Code,
			Name:         sc.Name,
			DisplayText:  sc.Description,
			Category:     sc.Category,
			IsVariable:   sc.IsVariable,
			IsParametric: false,
			Aliases:      sc.Aliases,
			SortOrder:    sortOrder,
		}

		if err := sr.repo.UpsertStat(ctx, stat); err != nil {
			return seeded, fmt.Errorf("seed stat %s: %w", sc.Code, err)
		}
		sr.known[sc.Code] = true
		for _, alias := range sc.Aliases {
			sr.known[alias] = true
		}
		seeded++
	}

	return seeded, nil
}

// SeedFromClasses seeds class skill and skill tree stat codes from the d2.classes table.
// Returns the number of new stats seeded.
func (sr *StatRegistry) SeedFromClasses(ctx context.Context) (int, error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	classes, err := sr.repo.GetAllClasses(ctx)
	if err != nil {
		return 0, fmt.Errorf("get classes: %w", err)
	}

	seeded := 0
	baseOrder := 100 // After Skills category

	for _, c := range classes {
		// Class skill code (e.g., "ama", "sor")
		classCode := c.ID
		if !sr.known[classCode] {
			stat := &Stat{
				Code:        classCode,
				Name:        c.Name + " Skills",
				DisplayText: fmt.Sprintf("+{value} To %s Skill Levels", c.Name),
				Category:    "Skills",
				IsVariable:  true,
				SortOrder:   baseOrder,
			}
			if err := sr.repo.UpsertStat(ctx, stat); err != nil {
				return seeded, fmt.Errorf("seed class stat %s: %w", classCode, err)
			}
			sr.known[classCode] = true
			seeded++
			baseOrder++
		}

		// Skill tree codes (e.g., "ama-bow", "sor-fire")
		for _, tree := range c.SkillTrees {
			treeCode := classCode + "-" + tree.Name
			if !sr.known[treeCode] {
				stat := &Stat{
					Code:        treeCode,
					Name:        tree.Name,
					DisplayText: fmt.Sprintf("+{value} To %s Skills (%s Only)", tree.Name, c.Name),
					Category:    "Skill Trees",
					IsVariable:  true,
					SortOrder:   baseOrder,
				}
				if err := sr.repo.UpsertStat(ctx, stat); err != nil {
					return seeded, fmt.Errorf("seed tree stat %s: %w", treeCode, err)
				}
				sr.known[treeCode] = true
				seeded++
				baseOrder++
			}
		}
	}

	return seeded, nil
}

// EnsureStat checks if a property's stat code exists in the registry.
// If not, it inserts a new stat with auto-derived name/category from the property.
func (sr *StatRegistry) EnsureStat(ctx context.Context, prop Property) error {
	if prop.Code == "" {
		return nil
	}

	// Skip parametric codes
	if parametricStatCodes[prop.Code] {
		return nil
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	if sr.known[prop.Code] {
		return nil
	}

	// Auto-derive name and category
	name := prop.Code
	displayText := prop.DisplayText
	if displayText == "" {
		displayText = prop.Code
	}
	category := "Other"

	stat := &Stat{
		Code:         prop.Code,
		Name:         name,
		DisplayText:  displayText,
		Category:     category,
		IsVariable:   true,
		IsParametric: false,
		SortOrder:    9999,
	}

	if err := sr.repo.UpsertStat(ctx, stat); err != nil {
		return fmt.Errorf("ensure stat %s: %w", prop.Code, err)
	}
	sr.known[prop.Code] = true
	return nil
}

// IsKnown returns whether a stat code is in the registry.
func (sr *StatRegistry) IsKnown(code string) bool {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.known[code]
}

// Count returns the number of known stat codes.
func (sr *StatRegistry) Count() int {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return len(sr.known)
}
