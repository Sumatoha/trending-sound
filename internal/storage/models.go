package storage

import "time"

// Sound represents a TikTok sound/music track
type Sound struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	URL       string    `json:"url"`
	UsesCount int64     `json:"uses_count"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SoundHistory tracks historical uses_count for trend detection
type SoundHistory struct {
	ID         int64     `json:"id"`
	SoundID    int64     `json:"sound_id"`
	UsesCount  int64     `json:"uses_count"`
	RecordedAt time.Time `json:"recorded_at"`
}

// User represents a Telegram bot user
type User struct {
	ID         int64     `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Niches     string    `json:"niches"` // JSON array of selected niches
	IsPremium  bool      `json:"is_premium"`
	CreatedAt  time.Time `json:"created_at"`
}

// TrendingSound represents a sound with growth metrics
type TrendingSound struct {
	Sound
	GrowthPercent float64 `json:"growth_percent"`
	OldUsesCount  int64   `json:"old_uses_count"`
}
