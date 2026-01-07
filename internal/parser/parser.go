package parser

import "github.com/yourusername/trending-sound/internal/storage"

// Parser defines the interface for TikTok sound parsing
type Parser interface {
	// FetchTrendingSounds fetches trending sounds for a given category
	FetchTrendingSounds(category string) ([]storage.Sound, error)

	// Close closes any resources used by the parser
	Close() error
}

// Categories supported by the parser
var Categories = []string{
	"fitness",
	"beauty",
	"comedy",
	"business",
	"tech",
	"lifestyle",
	"gaming",
}

// CategoryDisplayNames maps category keys to display names
var CategoryDisplayNames = map[string]string{
	"fitness":   "Fitness",
	"beauty":    "Beauty",
	"comedy":    "Comedy",
	"business":  "Business",
	"tech":      "Tech",
	"lifestyle": "Lifestyle",
	"gaming":    "Gaming",
}
