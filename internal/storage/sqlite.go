package storage

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage implements Storage interface using SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return &SQLiteStorage{db: db}, nil
}

// Init initializes the database schema
func (s *SQLiteStorage) Init() error {
	// Read migration file
	migrationSQL, err := os.ReadFile("migrations/init.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration
	if _, err := s.db.Exec(string(migrationSQL)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// SaveSound saves a new sound to the database
func (s *SQLiteStorage) SaveSound(sound *Sound) error {
	query := `
		INSERT INTO sounds (title, author, url, uses_count, category, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := s.db.Exec(query,
		sound.Title,
		sound.Author,
		sound.URL,
		sound.UsesCount,
		sound.Category,
		sound.CreatedAt,
		sound.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save sound: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	sound.ID = id

	return nil
}

// GetSoundByURL retrieves a sound by its URL
func (s *SQLiteStorage) GetSoundByURL(url string) (*Sound, error) {
	query := `
		SELECT id, title, author, url, uses_count, category, created_at, updated_at
		FROM sounds
		WHERE url = ?
	`
	sound := &Sound{}
	err := s.db.QueryRow(query, url).Scan(
		&sound.ID,
		&sound.Title,
		&sound.Author,
		&sound.URL,
		&sound.UsesCount,
		&sound.Category,
		&sound.CreatedAt,
		&sound.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sound: %w", err)
	}

	return sound, nil
}

// GetSoundsByCategory retrieves sounds by category with a limit
func (s *SQLiteStorage) GetSoundsByCategory(category string, limit int) ([]Sound, error) {
	query := `
		SELECT id, title, author, url, uses_count, category, created_at, updated_at
		FROM sounds
		WHERE category = ?
		ORDER BY updated_at DESC
		LIMIT ?
	`
	rows, err := s.db.Query(query, category, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get sounds by category: %w", err)
	}
	defer rows.Close()

	var sounds []Sound
	for rows.Next() {
		var sound Sound
		err := rows.Scan(
			&sound.ID,
			&sound.Title,
			&sound.Author,
			&sound.URL,
			&sound.UsesCount,
			&sound.Category,
			&sound.CreatedAt,
			&sound.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sound: %w", err)
		}
		sounds = append(sounds, sound)
	}

	return sounds, nil
}

// UpdateSound updates an existing sound
func (s *SQLiteStorage) UpdateSound(sound *Sound) error {
	query := `
		UPDATE sounds
		SET title = ?, author = ?, uses_count = ?, category = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := s.db.Exec(query,
		sound.Title,
		sound.Author,
		sound.UsesCount,
		sound.Category,
		sound.UpdatedAt,
		sound.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update sound: %w", err)
	}

	return nil
}

// SaveSoundHistory saves a sound history record
func (s *SQLiteStorage) SaveSoundHistory(soundID int64, usesCount int64) error {
	query := `
		INSERT INTO sound_history (sound_id, uses_count, recorded_at)
		VALUES (?, ?, ?)
	`
	_, err := s.db.Exec(query, soundID, usesCount, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save sound history: %w", err)
	}

	return nil
}

// GetSoundHistoryByTime retrieves sound history from N hours ago
func (s *SQLiteStorage) GetSoundHistoryByTime(soundID int64, hoursAgo int) (*SoundHistory, error) {
	cutoffTime := time.Now().Add(-time.Duration(hoursAgo) * time.Hour)

	query := `
		SELECT id, sound_id, uses_count, recorded_at
		FROM sound_history
		WHERE sound_id = ? AND recorded_at >= ?
		ORDER BY recorded_at ASC
		LIMIT 1
	`
	history := &SoundHistory{}
	err := s.db.QueryRow(query, soundID, cutoffTime).Scan(
		&history.ID,
		&history.SoundID,
		&history.UsesCount,
		&history.RecordedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sound history: %w", err)
	}

	return history, nil
}

// GetAllSoundsWithHistory retrieves all sounds and their history for trend detection
func (s *SQLiteStorage) GetAllSoundsWithHistory(category string, hoursAgo int) ([]Sound, map[int64]*SoundHistory, error) {
	// Get all sounds in category
	sounds, err := s.GetSoundsByCategory(category, 1000) // Get top 1000
	if err != nil {
		return nil, nil, err
	}

	// Get history for each sound
	historyMap := make(map[int64]*SoundHistory)
	for _, sound := range sounds {
		history, err := s.GetSoundHistoryByTime(sound.ID, hoursAgo)
		if err != nil {
			return nil, nil, err
		}
		if history != nil {
			historyMap[sound.ID] = history
		}
	}

	return sounds, historyMap, nil
}

// CreateUser creates a new user
func (s *SQLiteStorage) CreateUser(telegramID int64) error {
	query := `
		INSERT INTO users (telegram_id, niches, is_premium, created_at)
		VALUES (?, '[]', 0, ?)
	`
	_, err := s.db.Exec(query, telegramID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by Telegram ID
func (s *SQLiteStorage) GetUser(telegramID int64) (*User, error) {
	query := `
		SELECT id, telegram_id, niches, is_premium, created_at
		FROM users
		WHERE telegram_id = ?
	`
	user := &User{}
	err := s.db.QueryRow(query, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Niches,
		&user.IsPremium,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateUserNiches updates user's selected niches
func (s *SQLiteStorage) UpdateUserNiches(telegramID int64, niches string) error {
	query := `
		UPDATE users
		SET niches = ?
		WHERE telegram_id = ?
	`
	_, err := s.db.Exec(query, niches, telegramID)
	if err != nil {
		return fmt.Errorf("failed to update user niches: %w", err)
	}

	return nil
}

// GetAllUsers retrieves all users
func (s *SQLiteStorage) GetAllUsers() ([]User, error) {
	query := `
		SELECT id, telegram_id, niches, is_premium, created_at
		FROM users
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.TelegramID,
			&user.Niches,
			&user.IsPremium,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}
