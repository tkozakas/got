package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

const (
	defaultRedisAddr  = "localhost:6379"
	defaultLanguage   = "en"
	defaultConfigPath = "config.yaml"
)

type Config struct {
	BotToken  string
	DBURL     string
	GptKey    string
	RedisAddr string
	Bot       BotConfig      `yaml:"bot"`
	Schedule  ScheduleConfig `yaml:"schedule"`
}

type BotConfig struct {
	Language string `yaml:"language"`
}

type ScheduleConfig struct {
	WinnerReset string `yaml:"winner_reset"`
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

	cfg := &Config{
		BotToken:  token,
		DBURL:     dbURL,
		GptKey:    os.Getenv("GROQ_API_KEY"),
		RedisAddr: getEnvOrDefault("REDIS_ADDR", defaultRedisAddr),
	}

	loadYAMLConfig(cfg)

	if cfg.Bot.Language == "" {
		cfg.Bot.Language = getEnvOrDefault("LANGUAGE", defaultLanguage)
	}

	return cfg
}

func loadYAMLConfig(cfg *Config) {
	data, err := os.ReadFile(defaultConfigPath)
	if err != nil {
		slog.Warn("No config.yaml found, using defaults")
		setDefaults(cfg)
		return
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		slog.Error("Failed to parse config.yaml", "error", err)
		setDefaults(cfg)
	}
}

func setDefaults(cfg *Config) {
	cfg.Bot.Language = defaultLanguage
	cfg.Schedule.WinnerReset = "0 0 * * *"
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
