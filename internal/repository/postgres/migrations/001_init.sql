-- +migrate Up

-- Chats table
CREATE TABLE IF NOT EXISTS chats (
    chat_id BIGINT PRIMARY KEY,
    chat_name VARCHAR(255) NOT NULL DEFAULT '',
    language VARCHAR(10) NOT NULL DEFAULT ''
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    user_id BIGINT PRIMARY KEY,
    username VARCHAR(255) NOT NULL DEFAULT ''
);

-- Chat-User association (for tracking users per chat)
CREATE TABLE IF NOT EXISTS chat_users (
    chat_id BIGINT NOT NULL REFERENCES chats(chat_id),
    user_id BIGINT NOT NULL REFERENCES users(user_id),
    PRIMARY KEY (chat_id, user_id)
);

-- Reminders table
CREATE TABLE IF NOT EXISTS reminders (
    reminder_id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL REFERENCES chats(chat_id),
    user_id BIGINT NOT NULL REFERENCES users(user_id),
    message TEXT NOT NULL,
    remind_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_reminders_pending ON reminders(remind_at) WHERE sent = FALSE;
CREATE INDEX IF NOT EXISTS idx_reminders_chat ON reminders(chat_id);

-- Facts table (chat_id for chat-specific facts)
CREATE TABLE IF NOT EXISTS facts (
    fact_id BIGSERIAL PRIMARY KEY,
    comment TEXT NOT NULL,
    chat_id BIGINT NOT NULL REFERENCES chats(chat_id)
);

CREATE INDEX IF NOT EXISTS idx_facts_chat ON facts(chat_id);

-- Stickers table (chat_id for chat-specific stickers)
CREATE TABLE IF NOT EXISTS stickers (
    sticker_id BIGSERIAL PRIMARY KEY,
    sticker_set_name VARCHAR(255) NOT NULL DEFAULT '',
    file_id VARCHAR(255) NOT NULL,
    chat_id BIGINT NOT NULL REFERENCES chats(chat_id)
);

CREATE INDEX IF NOT EXISTS idx_stickers_chat ON stickers(chat_id);

-- Subreddits table (chat_id for chat-specific subreddits)
CREATE TABLE IF NOT EXISTS subreddits (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    chat_id BIGINT NOT NULL REFERENCES chats(chat_id),
    UNIQUE(name, chat_id)
);

CREATE INDEX IF NOT EXISTS idx_subreddits_chat ON subreddits(chat_id);

-- Stats table (daily winner roulette per chat per year)
CREATE TABLE IF NOT EXISTS stats (
    stat_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(user_id),
    chat_id BIGINT NOT NULL REFERENCES chats(chat_id),
    score BIGINT NOT NULL DEFAULT 0,
    year INT NOT NULL,
    is_winner BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(user_id, chat_id, year)
);

CREATE INDEX IF NOT EXISTS idx_stats_chat_year ON stats(chat_id, year);
CREATE INDEX IF NOT EXISTS idx_stats_winner ON stats(chat_id, year) WHERE is_winner = TRUE;
