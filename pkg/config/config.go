package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

const (
	defaultRedisAddr = "localhost:6379"
	defaultLanguage  = "en"
)

type Config struct {
	BotToken  string
	DBURL     string
	GptKey    string
	RedisAddr string
	Language  string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, relying on environment variables")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		slog.Error("BOT_TOKEN is required")
		os.Exit(1)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		slog.Error("DB_URL is required")
		os.Exit(1)
	}

	gptKey := os.Getenv("GROQ_API_KEY")
	redisAddr := getEnvOrDefault("REDIS_ADDR", defaultRedisAddr)
	language := getEnvOrDefault("LANGUAGE", defaultLanguage)

	return &Config{
		BotToken:  token,
		DBURL:     dbURL,
		GptKey:    gptKey,
		RedisAddr: redisAddr,
		Language:  language,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
