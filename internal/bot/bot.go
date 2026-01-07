package bot

import (
	"encoding/json"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yourusername/trending-sound/internal/detector"
	"github.com/yourusername/trending-sound/internal/parser"
	"github.com/yourusername/trending-sound/internal/storage"
)

// Bot represents the Telegram bot
type Bot struct {
	api      *tgbotapi.BotAPI
	storage  storage.Storage
	detector *detector.TrendDetector
}

// New creates a new Telegram bot instance
func New(token string, s storage.Storage, d *detector.TrendDetector) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:      api,
		storage:  s,
		detector: d,
	}, nil
}

// Start starts the bot and begins listening for updates
func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	log.Println("Bot started, listening for updates...")

	for update := range updates {
		if update.Message != nil {
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			b.handleCallbackQuery(update.CallbackQuery)
		}
	}

	return nil
}

// handleMessage handles incoming messages
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	if !message.IsCommand() {
		return
	}

	log.Printf("[%s] %s", message.From.UserName, message.Text)

	switch message.Command() {
	case "start":
		b.handleStart(message)
	case "niches":
		b.handleNiches(message)
	case "trending":
		b.handleTrending(message)
	case "premium":
		b.handlePremium(message)
	case "stats":
		b.handleStats(message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command. Available commands: /start, /niches, /trending, /premium")
		b.api.Send(msg)
	}
}

// SendTrendingAlert sends a trending alert to a user
func (b *Bot) SendTrendingAlert(telegramID int64, category string, sounds []storage.TrendingSound) error {
	if len(sounds) == 0 {
		return nil
	}

	message := formatTrendingMessage(category, sounds)

	msg := tgbotapi.NewMessage(telegramID, message)
	msg.ParseMode = "Markdown"

	_, err := b.api.Send(msg)
	return err
}

// formatTrendingMessage formats trending sounds into a message
func formatTrendingMessage(category string, sounds []storage.TrendingSound) string {
	categoryName := parser.CategoryDisplayNames[category]
	if categoryName == "" {
		categoryName = category
	}

	message := fmt.Sprintf("🔥 *Trending Sounds - %s*\n\n", categoryName)

	for i, ts := range sounds {
		message += fmt.Sprintf("*%d. \"%s\"*", i+1, ts.Title)
		if ts.Author != "" {
			message += fmt.Sprintf(" by %s", ts.Author)
		}
		message += "\n"
		message += fmt.Sprintf("   📊 Uses: %s", formatNumber(ts.UsesCount))
		if ts.GrowthPercent > 0 {
			message += fmt.Sprintf(" (+%.0f%%)", ts.GrowthPercent)
		}
		message += "\n"
		message += fmt.Sprintf("   🔗 [Listen](%s)\n\n", ts.URL)
	}

	return message
}

// formatNumber formats a number with K/M/B suffixes
func formatNumber(n int64) string {
	if n >= 1000000000 {
		return fmt.Sprintf("%.1fB", float64(n)/1000000000)
	}
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

// GetUserNiches returns the user's selected niches as a slice
func GetUserNiches(user *storage.User) []string {
	var niches []string
	if user.Niches != "" {
		json.Unmarshal([]byte(user.Niches), &niches)
	}
	return niches
}

// SetUserNiches sets the user's niches from a slice
func SetUserNiches(niches []string) string {
	if len(niches) == 0 {
		return "[]"
	}
	data, _ := json.Marshal(niches)
	return string(data)
}
