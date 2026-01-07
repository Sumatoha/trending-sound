-- Sounds table
CREATE TABLE IF NOT EXISTS sounds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    author TEXT,
    url TEXT UNIQUE NOT NULL,
    uses_count INTEGER DEFAULT 0,
    category TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sounds_category ON sounds(category);
CREATE INDEX IF NOT EXISTS idx_sounds_updated ON sounds(updated_at);

-- Sound history table for trend detection
CREATE TABLE IF NOT EXISTS sound_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sound_id INTEGER NOT NULL,
    uses_count INTEGER DEFAULT 0,
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sound_id) REFERENCES sounds(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sound_history_recorded ON sound_history(sound_id, recorded_at);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    telegram_id INTEGER UNIQUE NOT NULL,
    niches TEXT, -- JSON array ["fitness", "beauty"]
    is_premium BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
