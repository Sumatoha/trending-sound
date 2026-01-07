package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/yourusername/trending-sound/internal/bot"
	"github.com/yourusername/trending-sound/internal/config"
	"github.com/yourusername/trending-sound/internal/detector"
	"github.com/yourusername/trending-sound/internal/parser"
	"github.com/yourusername/trending-sound/internal/scheduler"
	"github.com/yourusername/trending-sound/internal/storage"
)

func main() {
	log.Println("Starting TikTok Trending Sounds Bot...")

	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config loaded: DataDir=%s, LogLevel=%s", cfg.DataDir, cfg.LogLevel)

	// 2. Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// 3. Initialize database
	dbPath := filepath.Join(cfg.DataDir, "sounds.db")
	log.Printf("Initializing database at: %s", dbPath)

	db, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	log.Println("Database initialized successfully")

	// 4. Create parser (API-based for MVP)
	log.Println("Initializing API parser...")
	apiParser := parser.NewAPIParser()
	log.Println("API parser initialized (using mock data for MVP)")

	// 5. Create detector
	log.Println("Initializing trend detector...")
	trendDetector := detector.New(db)

	// 6. Create Telegram bot
	log.Println("Initializing Telegram bot...")
	telegramBot, err := bot.New(cfg.TelegramBotToken, db, trendDetector)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}

	// 7. Create and start scheduler
	log.Println("Initializing scheduler...")
	sched := scheduler.New(apiParser, db, trendDetector, telegramBot)
	sched.Start()
	defer sched.Stop()

	// 8. Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start bot in a goroutine
	go func() {
		log.Println("Starting Telegram bot...")
		if err := telegramBot.Start(); err != nil {
			log.Fatalf("Bot error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, cleaning up...")

	// Cleanup
	apiParser.Close()

	log.Println("Bot stopped successfully")
}
