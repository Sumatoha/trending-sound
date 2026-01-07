package parser

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/yourusername/trending-sound/internal/storage"
)

// RodParser implements Parser using rod for browser automation
type RodParser struct {
	browser   *rod.Browser
	failCount int
	maxFails  int
}

// NewRodParser creates a new rod-based parser
func NewRodParser() (*RodParser, error) {
	// Launch browser
	u := launcher.New().
		Headless(true).
		Devtools(false).
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()

	return &RodParser{
		browser:   browser,
		failCount: 0,
		maxFails:  3,
	}, nil
}

// FetchTrendingSounds fetches trending sounds using browser automation
func (p *RodParser) FetchTrendingSounds(category string) ([]storage.Sound, error) {
	page := p.browser.MustPage()
	defer page.MustClose()

	// Set timeout
	page = page.Timeout(60 * time.Second)

	// Navigate to TikTok Creative Center
	// Note: This URL is a placeholder and needs to be adjusted based on actual TikTok Creative Center structure
	url := fmt.Sprintf("https://ads.tiktok.com/business/creativecenter/music/pc/en?from=001000")

	log.Printf("Navigating to %s for category: %s", url, category)

	err := page.Navigate(url)
	if err != nil {
		p.failCount++
		return nil, fmt.Errorf("failed to navigate: %w", err)
	}

	// Wait for page to load
	err = page.WaitLoad()
	if err != nil {
		p.failCount++
		return nil, fmt.Errorf("failed to wait for page load: %w", err)
	}

	// Additional wait for dynamic content
	time.Sleep(5 * time.Second)

	// Parse sounds from the page
	// Note: CSS selectors need to be adjusted based on actual TikTok Creative Center HTML structure
	sounds, err := p.parseSounds(page, category)
	if err != nil {
		p.failCount++
		return nil, err
	}

	// Reset fail count on success
	p.failCount = 0

	return sounds, nil
}

// parseSounds extracts sound data from the page
func (p *RodParser) parseSounds(page *rod.Page, category string) ([]storage.Sound, error) {
	var sounds []storage.Sound

	// NOTE: These selectors are placeholders and need to be updated based on actual TikTok Creative Center structure
	// You'll need to inspect the actual page to find the correct selectors

	// Try to find sound items
	// This is a generic approach - actual selectors will vary
	elements, err := page.Elements("div[class*='music-item'], div[class*='sound-item'], li[class*='music'], tr[class*='music']")
	if err != nil {
		return nil, fmt.Errorf("failed to find sound elements: %w", err)
	}

	log.Printf("Found %d potential sound elements", len(elements))

	// Limit to top 50 sounds
	limit := 50
	if len(elements) > limit {
		elements = elements[:limit]
	}

	for i, elem := range elements {
		sound, err := p.extractSoundFromElement(elem, category)
		if err != nil {
			log.Printf("Failed to extract sound from element %d: %v", i, err)
			continue
		}

		if sound != nil {
			sounds = append(sounds, *sound)
		}
	}

	if len(sounds) == 0 {
		return nil, fmt.Errorf("no sounds found - selectors may need updating")
	}

	log.Printf("Successfully parsed %d sounds for category: %s", len(sounds), category)

	return sounds, nil
}

// extractSoundFromElement extracts sound data from a single DOM element
func (p *RodParser) extractSoundFromElement(elem *rod.Element, category string) (*storage.Sound, error) {
	// NOTE: These selectors are placeholders and need to be updated
	// based on actual TikTok Creative Center HTML structure

	sound := &storage.Sound{
		Category: category,
	}

	// Try to extract title
	titleElem, err := elem.Element("*[class*='title'], *[class*='name'], *[class*='music-title']")
	if err == nil && titleElem != nil {
		if title, err := titleElem.Text(); err == nil {
			sound.Title = strings.TrimSpace(title)
		}
	}

	// Try to extract author
	authorElem, err := elem.Element("*[class*='author'], *[class*='artist'], *[class*='creator']")
	if err == nil && authorElem != nil {
		if author, err := authorElem.Text(); err == nil {
			sound.Author = strings.TrimSpace(author)
		}
	}

	// Try to extract uses count
	usesElem, err := elem.Element("*[class*='uses'], *[class*='count'], *[class*='post']")
	if err == nil && usesElem != nil {
		if usesText, err := usesElem.Text(); err == nil {
			sound.UsesCount = parseUsesCount(usesText)
		}
	}

	// Try to extract URL
	linkElem, err := elem.Element("a")
	if err == nil && linkElem != nil {
		if href, err := linkElem.Property("href"); err == nil {
			sound.URL = href.String()
		}
	}

	// Validate we have minimum required data
	if sound.Title == "" || sound.URL == "" {
		return nil, fmt.Errorf("missing required fields (title or url)")
	}

	// Generate URL from title if not found
	if sound.URL == "" {
		sound.URL = fmt.Sprintf("https://www.tiktok.com/music/%s", strings.ReplaceAll(sound.Title, " ", "-"))
	}

	return sound, nil
}

// parseUsesCount parses uses count from text like "15.2K" or "1.5M"
func parseUsesCount(text string) int64 {
	text = strings.TrimSpace(text)
	text = strings.ToUpper(text)

	// Remove "uses" or similar words
	text = strings.ReplaceAll(text, "USES", "")
	text = strings.ReplaceAll(text, "POSTS", "")
	text = strings.TrimSpace(text)

	multiplier := int64(1)

	if strings.HasSuffix(text, "K") {
		multiplier = 1000
		text = strings.TrimSuffix(text, "K")
	} else if strings.HasSuffix(text, "M") {
		multiplier = 1000000
		text = strings.TrimSuffix(text, "M")
	} else if strings.HasSuffix(text, "B") {
		multiplier = 1000000000
		text = strings.TrimSuffix(text, "B")
	}

	// Parse the number
	num, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0
	}

	return int64(num * float64(multiplier))
}

// ShouldFallback returns true if the parser has failed too many times
func (p *RodParser) ShouldFallback() bool {
	return p.failCount >= p.maxFails
}

// Close closes the browser
func (p *RodParser) Close() error {
	if p.browser != nil {
		return p.browser.Close()
	}
	return nil
}
