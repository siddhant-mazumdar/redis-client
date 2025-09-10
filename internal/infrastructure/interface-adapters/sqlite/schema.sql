-- Redis Keys Table
CREATE TABLE IF NOT EXISTS redis_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key_name TEXT NOT NULL UNIQUE,
    key_type TEXT NOT NULL CHECK(key_type IN ('string', 'hash', 'list', 'set', 'zset', 'stream')),
    ttl INTEGER DEFAULT -1, -- -1 means no expiration
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Redis String Values
CREATE TABLE IF NOT EXISTS redis_strings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key_id INTEGER NOT NULL,
    value TEXT,
    FOREIGN KEY (key_id) REFERENCES redis_keys(id) ON DELETE CASCADE
);

-- Redis Hash Fields
CREATE TABLE IF NOT EXISTS redis_hashes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key_id INTEGER NOT NULL,
    field TEXT NOT NULL,
    value TEXT,
    UNIQUE(key_id, field),
    FOREIGN KEY (key_id) REFERENCES redis_keys(id) ON DELETE CASCADE
);

-- Redis List Elements
CREATE TABLE IF NOT EXISTS redis_lists (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key_id INTEGER NOT NULL,
    position INTEGER NOT NULL,
    value TEXT,
    FOREIGN KEY (key_id) REFERENCES redis_keys(id) ON DELETE CASCADE
);

-- Redis Set Members
CREATE TABLE IF NOT EXISTS redis_sets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key_id INTEGER NOT NULL,
    member TEXT NOT NULL,
    UNIQUE(key_id, member),
    FOREIGN KEY (key_id) REFERENCES redis_keys(id) ON DELETE CASCADE
);

-- Redis Sorted Set Members
CREATE TABLE IF NOT EXISTS redis_zsets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key_id INTEGER NOT NULL,
    member TEXT NOT NULL,
    score REAL NOT NULL,
    UNIQUE(key_id, member),
    FOREIGN KEY (key_id) REFERENCES redis_keys(id) ON DELETE CASCADE
);

-- Redis Commands Log
CREATE TABLE IF NOT EXISTS redis_commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command TEXT NOT NULL,
    key_name TEXT,
    arguments TEXT, -- JSON array of arguments
    result TEXT,
    executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_redis_keys_name ON redis_keys(key_name);
CREATE INDEX IF NOT EXISTS idx_redis_keys_type ON redis_keys(key_type);
CREATE INDEX IF NOT EXISTS idx_redis_hashes_key_field ON redis_hashes(key_id, field);
CREATE INDEX IF NOT EXISTS idx_redis_lists_key_pos ON redis_lists(key_id, position);
CREATE INDEX IF NOT EXISTS idx_redis_sets_key_member ON redis_sets(key_id, member);
CREATE INDEX IF NOT EXISTS idx_redis_zsets_key_score ON redis_zsets(key_id, score);
CREATE INDEX IF NOT EXISTS idx_redis_commands_key ON redis_commands(key_name);
CREATE INDEX IF NOT EXISTS idx_redis_commands_time ON redis_commands(executed_at);