package bot

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yourusername/trending-sound/internal/parser"
)

// handleStart handles the /start command
func (b *Bot) handleStart(message *tgbotapi.Message) {
	telegramID := message.From.ID

	// Check if user exists
	user, err := b.storage.GetUser(telegramID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "An error occurred. Please try again later.")
		b.api.Send(msg)
		return
	}

	// Create user if doesn't exist
	if user == nil {
		err := b.storage.CreateUser(telegramID)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "An error occurred. Please try again later.")
			b.api.Send(msg)
			return
		}
	}

	// Send welcome message
	welcomeText := `👋 Welcome to TikTok Trending Sounds Tracker!

I'll help you discover trending sounds on TikTok before they go viral.

🎯 *How it works:*
• Choose your niches
• Get automatic alerts when sounds start trending
• View current top trending sounds anytime

📱 *Commands:*
/niches - Select your niches
/trending - View current trending sounds

Let's get started! Choose your niches below:`

	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createNichesKeyboard([]string{})
	b.api.Send(msg)
}

// handleNiches handles the /niches command
func (b *Bot) handleNiches(message *tgbotapi.Message) {
	telegramID := message.From.ID

	user, err := b.storage.GetUser(telegramID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "An error occurred. Please try again later.")
		b.api.Send(msg)
		return
	}

	if user == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please use /start first to register.")
		b.api.Send(msg)
		return
	}

	currentNiches := GetUserNiches(user)

	text := "📊 *Your Niches*\n\nSelect the niches you want to track:"
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createNichesKeyboard(currentNiches)
	b.api.Send(msg)
}

// handleTrending handles the /trending command
func (b *Bot) handleTrending(message *tgbotapi.Message) {
	telegramID := message.From.ID

	user, err := b.storage.GetUser(telegramID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "An error occurred. Please try again later.")
		b.api.Send(msg)
		return
	}

	if user == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please use /start first to register.")
		b.api.Send(msg)
		return
	}

	niches := GetUserNiches(user)
	if len(niches) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "You haven't selected any niches yet. Use /niches to choose your interests.")
		b.api.Send(msg)
		return
	}

	// Send "loading" message
	loadingMsg := tgbotapi.NewMessage(message.Chat.ID, "🔍 Finding trending sounds...")
	b.api.Send(loadingMsg)

	// Get trending sounds for each niche
	for _, niche := range niches {
		trending, err := b.detector.DetectTrending(niche, 5)
		if err != nil {
			log.Printf("Error detecting trends for %s: %v", niche, err)
			continue
		}

		if len(trending) == 0 {
			msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("No trending sounds found for %s yet. Check back later!", parser.CategoryDisplayNames[niche]))
			b.api.Send(msg)
			continue
		}

		b.SendTrendingAlert(telegramID, niche, trending)
	}
}

// handleCallbackQuery handles callback queries from inline keyboards
func (b *Bot) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	telegramID := callback.From.ID

	// Answer callback to remove loading state
	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	b.api.Request(callbackConfig)

	// Parse callback data
	// Format: "niche:fitness" or "niche_done"
	parts := strings.Split(callback.Data, ":")

	if parts[0] == "niche_done" {
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "✅ Your niches have been saved! Use /trending to see current trending sounds.")
		b.api.Send(msg)
		return
	}

	// Handle premium activation
	if parts[0] == "premium" && len(parts) == 2 && parts[1] == "activate" {
		// Activate premium for MVP testing
		err := b.storage.SetPremium(telegramID, true)
		if err != nil {
			log.Printf("Error activating premium: %v", err)
			return
		}

		msg := tgbotapi.NewMessage(callback.Message.Chat.ID,
			"🎉 Premium activated!\n\n"+
			"You now have access to:\n"+
			"✅ All 7 niches\n"+
			"✅ Alerts every 3 hours\n"+
			"✅ Top 10 trending sounds\n\n"+
			"Use /niches to select more niches!")
		b.api.Send(msg)
		return
	}

	if parts[0] != "niche" || len(parts) != 2 {
		return
	}

	niche := parts[1]

	// Get user
	user, err := b.storage.GetUser(telegramID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return
	}

	if user == nil {
		return
	}

	// Toggle niche
	currentNiches := GetUserNiches(user)
	newNiches := toggleNiche(currentNiches, niche)

	// Update user niches
	nichesJSON := SetUserNiches(newNiches)
	err = b.storage.UpdateUserNiches(telegramID, nichesJSON)
	if err != nil {
		log.Printf("Error updating user niches: %v", err)
		return
	}

	// Update keyboard
	editMsg := tgbotapi.NewEditMessageReplyMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		createNichesKeyboard(newNiches),
	)
	b.api.Send(editMsg)
}

// createNichesKeyboard creates an inline keyboard for niche selection
func createNichesKeyboard(selectedNiches []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Create button for each niche (2 per row)
	var currentRow []tgbotapi.InlineKeyboardButton
	for i, category := range parser.Categories {
		displayName := parser.CategoryDisplayNames[category]
		if displayName == "" {
			displayName = category
		}

		// Add checkmark if selected
		if contains(selectedNiches, category) {
			displayName = "✅ " + displayName
		}

		button := tgbotapi.NewInlineKeyboardButtonData(displayName, "niche:"+category)
		currentRow = append(currentRow, button)

		// Create new row after 2 buttons or at the end
		if len(currentRow) == 2 || i == len(parser.Categories)-1 {
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}

	// Add "Done" button
	doneButton := tgbotapi.NewInlineKeyboardButtonData("✅ Done", "niche_done")
	rows = append(rows, []tgbotapi.InlineKeyboardButton{doneButton})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// toggleNiche adds or removes a niche from the list
func toggleNiche(niches []string, niche string) []string {
	// Check if niche exists
	for i, n := range niches {
		if n == niche {
			// Remove it
			return append(niches[:i], niches[i+1:]...)
		}
	}

	// Add it
	return append(niches, niche)
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// handlePremium handles the /premium command
func (b *Bot) handlePremium(message *tgbotapi.Message) {
	telegramID := message.From.ID

	user, err := b.storage.GetUser(telegramID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return
	}

	if user == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please use /start first.")
		b.api.Send(msg)
		return
	}

	if user.IsPremium {
		text := `✨ You already have Premium!

Premium features:
• All 7 niches
• Alerts every 3 hours
• Top 10 trending sounds
• Priority notifications

Thank you for your support! 💎`

		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		b.api.Send(msg)
		return
	}

	// Show upgrade options
	text := `🚀 Upgrade to Premium!

Get unlimited access:
✅ All 7 niches (Free: only 2)
✅ Alerts every 3 hours (Free: 12h)
✅ Top 10 sounds (Free: top 3)
✅ Priority notifications
✅ 30 days history

💰 Price: $4.99/month

For MVP testing, use /premium_activate to activate for free!`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎁 Activate (Free for MVP)", "premium:activate"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

// handleStats shows user statistics
func (b *Bot) handleStats(message *tgbotapi.Message) {
	telegramID := message.From.ID

	user, err := b.storage.GetUser(telegramID)
	if err != nil || user == nil {
		return
	}

	niches := GetUserNiches(user)
	nichesText := "None"
	if len(niches) > 0 {
		nichesText = fmt.Sprintf("%d selected", len(niches))
	}

	status := "Free"
	if user.IsPremium {
		status = "Premium 💎"
	}

	// Get total trending sounds count (example)
	totalTrending := 0
	for _, niche := range niches {
		trending, _ := b.detector.DetectTrending(niche, 10)
		totalTrending += len(trending)
	}

	text := fmt.Sprintf(`📊 Your Statistics

👤 Status: %s
🎯 Niches: %s
🔥 Trending now: %d sounds

Join date: %s

Use /premium to upgrade!`,
		status,
		nichesText,
		totalTrending,
		user.CreatedAt.Format("Jan 02, 2006"))

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)
}
