package storage

import "time"

// Storage defines the interface for data persistence
type Storage interface {
	// Init initializes the database schema
	Init() error

	// Close closes the database connection
	Close() error

	// Sound operations
	SaveSound(sound *Sound) error
	GetSoundByURL(url string) (*Sound, error)
	GetSoundsByCategory(category string, limit int) ([]Sound, error)
	UpdateSound(sound *Sound) error

	// Sound history operations
	SaveSoundHistory(soundID int64, usesCount int64) error
	GetSoundHistoryByTime(soundID int64, hoursAgo int) (*SoundHistory, error)
	GetAllSoundsWithHistory(category string, hoursAgo int) ([]Sound, map[int64]*SoundHistory, error)

	// User operations
	CreateUser(telegramID int64) error
	GetUser(telegramID int64) (*User, error)
	UpdateUserNiches(telegramID int64, niches string) error
	GetAllUsers() ([]User, error)
	SetPremium(telegramID int64, isPremium bool) error
}

// SaveSoundWithHistory is a helper to save sound and its history in one transaction
func SaveSoundWithHistory(s Storage, sound *Sound) error {
	// Try to get existing sound
	existing, err := s.GetSoundByURL(sound.URL)
	if err == nil && existing != nil {
		// Update existing sound
		sound.ID = existing.ID
		sound.CreatedAt = existing.CreatedAt
		sound.UpdatedAt = time.Now()
		if err := s.UpdateSound(sound); err != nil {
			return err
		}
	} else {
		// Create new sound
		sound.CreatedAt = time.Now()
		sound.UpdatedAt = time.Now()
		if err := s.SaveSound(sound); err != nil {
			return err
		}
		// Get the created sound to get its ID
		created, err := s.GetSoundByURL(sound.URL)
		if err != nil {
			return err
		}
		sound.ID = created.ID
	}

	// Save history record
	return s.SaveSoundHistory(sound.ID, sound.UsesCount)
}
