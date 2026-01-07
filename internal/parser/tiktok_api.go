package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/yourusername/trending-sound/internal/storage"
)

// APIParser implements Parser using direct API calls
type APIParser struct {
	client *http.Client
}

// NewAPIParser creates a new API-based parser
func NewAPIParser() *APIParser {
	return &APIParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TikTokAPIResponse represents the API response structure
// Note: This is a placeholder and needs to be adjusted based on actual API response
type TikTokAPIResponse struct {
	Data struct {
		MusicList []struct {
			MusicID   string `json:"music_id"`
			Title     string `json:"title"`
			Author    string `json:"author"`
			UseCount  int64  `json:"use_count"`
			MusicURL  string `json:"music_url"`
		} `json:"music_list"`
	} `json:"data"`
}

// FetchTrendingSounds fetches trending sounds using TikTok API
func (p *APIParser) FetchTrendingSounds(category string) ([]storage.Sound, error) {
	// Note: This endpoint is a placeholder and needs to be adjusted
	// based on actual TikTok API structure. You may need to:
	// 1. Add authentication headers
	// 2. Adjust the endpoint URL
	// 3. Update the response parsing logic

	url := "https://m.tiktok.com/api/music/trending"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.tiktok.com/")

	// Add query parameters if needed
	q := req.URL.Query()
	q.Add("category", category)
	q.Add("count", "50")
	req.URL.RawQuery = q.Encode()

	log.Printf("Fetching sounds from API for category: %s", category)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp TikTokAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	// Convert to storage.Sound
	var sounds []storage.Sound
	for _, music := range apiResp.Data.MusicList {
		sound := storage.Sound{
			Title:     music.Title,
			Author:    music.Author,
			URL:       music.MusicURL,
			UsesCount: music.UseCount,
			Category:  category,
		}

		// Generate URL if not provided
		if sound.URL == "" {
			sound.URL = fmt.Sprintf("https://www.tiktok.com/music/%s-%s", music.Title, music.MusicID)
		}

		sounds = append(sounds, sound)
	}

	if len(sounds) == 0 {
		// Return mock data for testing purposes
		// This should be removed in production
		return p.getMockData(category), nil
	}

	log.Printf("Successfully fetched %d sounds from API for category: %s", len(sounds), category)

	return sounds, nil
}

// getMockData returns mock data for testing
// This provides realistic trending sounds data for MVP
func (p *APIParser) getMockData(category string) []storage.Sound {
	log.Printf("Using mock data for category: %s (MVP mode)", category)

	// Category-specific mock sounds
	mockSounds := map[string][]storage.Sound{
		"fitness": {
			{Title: "Workout Motivation Mix", Author: "DJ Fitness", URL: "https://www.tiktok.com/music/workout-1", UsesCount: 12500},
			{Title: "Gym Beast Mode", Author: "PowerHouse", URL: "https://www.tiktok.com/music/gym-1", UsesCount: 8900},
			{Title: "Running Energy", Author: "CardioKing", URL: "https://www.tiktok.com/music/run-1", UsesCount: 15200},
			{Title: "HIIT Workout Beat", Author: "FitBeats", URL: "https://www.tiktok.com/music/hiit-1", UsesCount: 6700},
			{Title: "Leg Day Anthem", Author: "GymFlow", URL: "https://www.tiktok.com/music/legs-1", UsesCount: 19800},
		},
		"beauty": {
			{Title: "Get Ready With Me", Author: "BeautyVibes", URL: "https://www.tiktok.com/music/grwm-1", UsesCount: 24500},
			{Title: "Makeup Tutorial Beats", Author: "GlamSquad", URL: "https://www.tiktok.com/music/makeup-1", UsesCount: 18300},
			{Title: "Skincare Routine", Author: "GlowUp", URL: "https://www.tiktok.com/music/skin-1", UsesCount: 11200},
			{Title: "Hair Transformation", Author: "HairGoals", URL: "https://www.tiktok.com/music/hair-1", UsesCount: 9400},
			{Title: "Baddie Energy", Author: "Confidence", URL: "https://www.tiktok.com/music/baddie-1", UsesCount: 27600},
		},
		"comedy": {
			{Title: "POV: You're Funny", Author: "ComedyGold", URL: "https://www.tiktok.com/music/pov-1", UsesCount: 31200},
			{Title: "Expectation vs Reality", Author: "Relatable", URL: "https://www.tiktok.com/music/expect-1", UsesCount: 22800},
			{Title: "When Mom Calls", Author: "FamilyHumor", URL: "https://www.tiktok.com/music/mom-1", UsesCount: 16500},
			{Title: "Monday Mood", Author: "WeekdayVibes", URL: "https://www.tiktok.com/music/monday-1", UsesCount: 14300},
			{Title: "Epic Fail Sound", Author: "Oops", URL: "https://www.tiktok.com/music/fail-1", UsesCount: 19700},
		},
		"business": {
			{Title: "CEO Mindset", Author: "Entrepreneur", URL: "https://www.tiktok.com/music/ceo-1", UsesCount: 8900},
			{Title: "Success Story", Author: "Motivation", URL: "https://www.tiktok.com/music/success-1", UsesCount: 12400},
			{Title: "Side Hustle Grind", Author: "Hustle", URL: "https://www.tiktok.com/music/hustle-1", UsesCount: 7600},
			{Title: "Financial Freedom", Author: "MoneyMoves", URL: "https://www.tiktok.com/music/money-1", UsesCount: 15800},
			{Title: "Boss Moves Only", Author: "Alpha", URL: "https://www.tiktok.com/music/boss-1", UsesCount: 11100},
		},
		"tech": {
			{Title: "AI Revolution", Author: "TechWave", URL: "https://www.tiktok.com/music/ai-1", UsesCount: 9200},
			{Title: "Coding Flow", Author: "DevLife", URL: "https://www.tiktok.com/music/code-1", UsesCount: 6800},
			{Title: "Tech Unboxing", Author: "Gadgets", URL: "https://www.tiktok.com/music/unbox-1", UsesCount: 13500},
			{Title: "Future Is Now", Author: "Innovation", URL: "https://www.tiktok.com/music/future-1", UsesCount: 10900},
			{Title: "Programmer Mood", Author: "DebugLife", URL: "https://www.tiktok.com/music/debug-1", UsesCount: 8400},
		},
		"lifestyle": {
			{Title: "That Girl Morning", Author: "Aesthetic", URL: "https://www.tiktok.com/music/morning-1", UsesCount: 21300},
			{Title: "Coffee Shop Vibes", Author: "CafeAmbience", URL: "https://www.tiktok.com/music/coffee-1", UsesCount: 17800},
			{Title: "Cozy Sunday", Author: "ChillVibes", URL: "https://www.tiktok.com/music/sunday-1", UsesCount: 14600},
			{Title: "Vlog Day", Author: "DailyLife", URL: "https://www.tiktok.com/music/vlog-1", UsesCount: 19200},
			{Title: "Self Care Time", Author: "Wellness", URL: "https://www.tiktok.com/music/selfcare-1", UsesCount: 16700},
		},
		"gaming": {
			{Title: "Victory Royale", Author: "GamerAnthem", URL: "https://www.tiktok.com/music/victory-1", UsesCount: 28900},
			{Title: "Rage Quit", Author: "GamingMoments", URL: "https://www.tiktok.com/music/rage-1", UsesCount: 23400},
			{Title: "Level Up Sound", Author: "Achievement", URL: "https://www.tiktok.com/music/levelup-1", UsesCount: 18600},
			{Title: "Gaming Setup Tour", Author: "PCMaster", URL: "https://www.tiktok.com/music/setup-1", UsesCount: 12300},
			{Title: "Clutch Moment", Author: "ProGamer", URL: "https://www.tiktok.com/music/clutch-1", UsesCount: 25700},
		},
	}

	// Get sounds for category or return generic if not found
	sounds, exists := mockSounds[category]
	if !exists {
		sounds = []storage.Sound{
			{Title: "Trending Sound 1", Author: "Artist", URL: "https://www.tiktok.com/music/1", UsesCount: 15000},
			{Title: "Trending Sound 2", Author: "Creator", URL: "https://www.tiktok.com/music/2", UsesCount: 10500},
			{Title: "Trending Sound 3", Author: "Producer", URL: "https://www.tiktok.com/music/3", UsesCount: 18200},
		}
	}

	// Set category for all sounds
	for i := range sounds {
		sounds[i].Category = category
	}

	return sounds
}

// Close closes the parser (no-op for API parser)
func (p *APIParser) Close() error {
	return nil
}
