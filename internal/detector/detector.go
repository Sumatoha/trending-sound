package detector

import (
	"fmt"
	"log"
	"sort"

	"github.com/yourusername/trending-sound/internal/storage"
)

// TrendDetector detects trending sounds based on growth metrics
type TrendDetector struct {
	storage storage.Storage
}

// New creates a new trend detector
func New(s storage.Storage) *TrendDetector {
	return &TrendDetector{
		storage: s,
	}
}

// TrendCriteria defines the criteria for a sound to be considered trending
type TrendCriteria struct {
	MinUsesCount  int64   // Minimum uses count (default: 500)
	MaxUsesCount  int64   // Maximum uses count (default: 30000)
	MinGrowth     float64 // Minimum growth percentage (default: 150%)
	LookbackHours int     // Hours to look back for comparison (default: 24)
}

// DefaultCriteria returns default trend detection criteria
func DefaultCriteria() TrendCriteria {
	return TrendCriteria{
		MinUsesCount:  500,
		MaxUsesCount:  30000,
		MinGrowth:     150.0,
		LookbackHours: 24,
	}
}

// DetectTrending detects trending sounds for a specific category
func (d *TrendDetector) DetectTrending(category string, limit int) ([]storage.TrendingSound, error) {
	criteria := DefaultCriteria()
	return d.DetectTrendingWithCriteria(category, limit, criteria)
}

// DetectTrendingWithCriteria detects trending sounds with custom criteria
func (d *TrendDetector) DetectTrendingWithCriteria(category string, limit int, criteria TrendCriteria) ([]storage.TrendingSound, error) {
	// Get all sounds with their history
	sounds, historyMap, err := d.storage.GetAllSoundsWithHistory(category, criteria.LookbackHours)
	if err != nil {
		return nil, fmt.Errorf("failed to get sounds with history: %w", err)
	}

	log.Printf("Analyzing %d sounds for trends in category: %s", len(sounds), category)

	var trendingSounds []storage.TrendingSound

	for _, sound := range sounds {
		// Check if sound meets basic criteria
		if sound.UsesCount < criteria.MinUsesCount || sound.UsesCount > criteria.MaxUsesCount {
			continue
		}

		// Get historical data
		history, exists := historyMap[sound.ID]
		if !exists || history == nil {
			// No historical data - skip
			continue
		}

		// Calculate growth percentage
		oldCount := history.UsesCount
		if oldCount == 0 {
			// Avoid division by zero - if old count is 0, this is a new sound
			// We can consider it trending if it has enough uses
			if sound.UsesCount >= criteria.MinUsesCount {
				trendingSounds = append(trendingSounds, storage.TrendingSound{
					Sound:         sound,
					GrowthPercent: 999.9, // Special marker for new sounds
					OldUsesCount:  0,
				})
			}
			continue
		}

		growth := calculateGrowth(oldCount, sound.UsesCount)

		// Check if growth meets criteria
		if growth >= criteria.MinGrowth {
			trendingSounds = append(trendingSounds, storage.TrendingSound{
				Sound:         sound,
				GrowthPercent: growth,
				OldUsesCount:  oldCount,
			})
		}
	}

	// Sort by growth percentage (descending)
	sort.Slice(trendingSounds, func(i, j int) bool {
		return trendingSounds[i].GrowthPercent > trendingSounds[j].GrowthPercent
	})

	// Limit results
	if limit > 0 && len(trendingSounds) > limit {
		trendingSounds = trendingSounds[:limit]
	}

	log.Printf("Found %d trending sounds in category: %s", len(trendingSounds), category)

	return trendingSounds, nil
}

// calculateGrowth calculates growth percentage
func calculateGrowth(oldCount, newCount int64) float64 {
	if oldCount == 0 {
		return 0
	}
	return float64(newCount-oldCount) / float64(oldCount) * 100.0
}

// AnalyzeTrends provides detailed trend analysis for a category
func (d *TrendDetector) AnalyzeTrends(category string) (*TrendAnalysis, error) {
	trendingSounds, err := d.DetectTrending(category, 10)
	if err != nil {
		return nil, err
	}

	analysis := &TrendAnalysis{
		Category:       category,
		TrendingCount:  len(trendingSounds),
		TrendingSounds: trendingSounds,
	}

	if len(trendingSounds) > 0 {
		// Calculate average growth
		var totalGrowth float64
		for _, ts := range trendingSounds {
			totalGrowth += ts.GrowthPercent
		}
		analysis.AverageGrowth = totalGrowth / float64(len(trendingSounds))

		// Find top sound
		analysis.TopSound = &trendingSounds[0]
	}

	return analysis, nil
}

// TrendAnalysis contains trend analysis results
type TrendAnalysis struct {
	Category       string
	TrendingCount  int
	AverageGrowth  float64
	TopSound       *storage.TrendingSound
	TrendingSounds []storage.TrendingSound
}
