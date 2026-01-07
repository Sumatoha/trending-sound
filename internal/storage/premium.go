package storage

import "time"

// SetPremium sets user premium status
func (s *SQLiteStorage) SetPremium(telegramID int64, isPremium bool) error {
	query := `
		UPDATE users
		SET is_premium = ?
		WHERE telegram_id = ?
	`
	_, err := s.db.Exec(query, isPremium, telegramID)
	return err
}

// SetPremiumExpiry sets when premium expires
func (s *SQLiteStorage) SetPremiumExpiry(telegramID int64, expiresAt time.Time) error {
	// Для этого нужно добавить колонку premium_expires_at в таблицу users
	// Пока просто возвращаем nil
	// TODO: добавить миграцию для premium_expires_at
	return nil
}

// CheckAndExpirePremium checks if premium has expired and removes it
func (s *SQLiteStorage) CheckAndExpirePremium() error {
	// TODO: реализовать когда добавим premium_expires_at колонку
	// UPDATE users SET is_premium = 0 WHERE premium_expires_at < NOW()
	return nil
}

// GetPremiumStats returns premium statistics
func (s *SQLiteStorage) GetPremiumStats() (total, premium int, err error) {
	// Total users
	err = s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return 0, 0, err
	}

	// Premium users
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_premium = 1").Scan(&premium)
	if err != nil {
		return 0, 0, err
	}

	return total, premium, nil
}
