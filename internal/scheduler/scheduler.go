package scheduler

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/yourusername/trending-sound/internal/bot"
	"github.com/yourusername/trending-sound/internal/detector"
	"github.com/yourusername/trending-sound/internal/parser"
	"github.com/yourusername/trending-sound/internal/storage"
)

// Scheduler handles scheduled tasks for data collection and alerts
type Scheduler struct {
	cron     *cron.Cron
	parser   parser.Parser
	storage  storage.Storage
	detector *detector.TrendDetector
	bot      *bot.Bot
}

// New creates a new scheduler
func New(p parser.Parser, s storage.Storage, d *detector.TrendDetector, b *bot.Bot) *Scheduler {
	return &Scheduler{
		cron:     cron.New(),
		parser:   p,
		storage:  s,
		detector: d,
		bot:      b,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	// Collect sounds every 3 hours
	s.cron.AddFunc("0 */3 * * *", func() {
		log.Println("Starting scheduled sound collection...")
		s.CollectSounds()
	})

	// Send alerts every 6 hours
	s.cron.AddFunc("0 */6 * * *", func() {
		log.Println("Starting scheduled alert sending...")
		s.SendAlerts()
	})

	// Run initial collection on startup (after a short delay)
	go func() {
		time.Sleep(10 * time.Second)
		log.Println("Running initial sound collection...")
		s.CollectSounds()
	}()

	s.cron.Start()
	log.Println("Scheduler started")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("Scheduler stopped")
}

// CollectSounds collects sounds from all categories
func (s *Scheduler) CollectSounds() {
	log.Println("Collecting sounds from all categories...")

	for _, category := range parser.Categories {
		log.Printf("Collecting sounds for category: %s", category)

		sounds, err := s.parser.FetchTrendingSounds(category)
		if err != nil {
			log.Printf("Error fetching sounds for %s: %v", category, err)
			continue
		}

		log.Printf("Fetched %d sounds for category: %s", len(sounds), category)

		// Save each sound with history
		for _, sound := range sounds {
			err := storage.SaveSoundWithHistory(s.storage, &sound)
			if err != nil {
				log.Printf("Error saving sound %s: %v", sound.Title, err)
				continue
			}
		}

		log.Printf("Successfully saved %d sounds for category: %s", len(sounds), category)

		// Small delay between categories to avoid rate limiting
		time.Sleep(2 * time.Second)
	}

	log.Println("Sound collection completed")
}

// SendAlerts sends trending alerts to all users
func (s *Scheduler) SendAlerts() {
	log.Println("Sending trending alerts to users...")

	// Get all users
	users, err := s.storage.GetAllUsers()
	if err != nil {
		log.Printf("Error getting users: %v", err)
		return
	}

	log.Printf("Found %d users", len(users))

	alertsSent := 0

	for _, user := range users {
		niches := bot.GetUserNiches(&user)
		if len(niches) == 0 {
			continue
		}

		log.Printf("Sending alerts to user %d for niches: %v", user.TelegramID, niches)

		for _, niche := range niches {
			// Detect trending sounds for this niche
			trending, err := s.detector.DetectTrending(niche, 5)
			if err != nil {
				log.Printf("Error detecting trends for %s: %v", niche, err)
				continue
			}

			if len(trending) == 0 {
				log.Printf("No trending sounds found for niche: %s", niche)
				continue
			}

			// Send alert
			err = s.bot.SendTrendingAlert(user.TelegramID, niche, trending)
			if err != nil {
				log.Printf("Error sending alert to user %d: %v", user.TelegramID, err)
				continue
			}

			alertsSent++

			// Rate limiting: 1 message per second
			time.Sleep(1 * time.Second)
		}
	}

	log.Printf("Alert sending completed. Sent %d alerts", alertsSent)
}

// ManualCollect triggers a manual collection for a specific category
func (s *Scheduler) ManualCollect(category string) error {
	log.Printf("Manual collection triggered for category: %s", category)

	sounds, err := s.parser.FetchTrendingSounds(category)
	if err != nil {
		return err
	}

	for _, sound := range sounds {
		err := storage.SaveSoundWithHistory(s.storage, &sound)
		if err != nil {
			log.Printf("Error saving sound %s: %v", sound.Title, err)
		}
	}

	log.Printf("Manual collection completed for category: %s", category)
	return nil
}
