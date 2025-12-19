package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

const (
	defaultRedisAddr    = "localhost:6379"
	defaultLanguage     = "en"
	defaultConfigPath   = "config.yaml"
	defaultWinnerReset  = "0 0 0 * * *"
	defaultAutoRoulette = "0 0 11 * * *"
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
	WinnerReset  string `yaml:"winner_reset"`
	AutoRoulette string `yaml:"auto_roulette"`
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
	applyEnvOverrides(cfg)

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

func applyEnvOverrides(cfg *Config) {
	if lang := os.Getenv("BOT_LANGUAGE"); lang != "" {
		cfg.Bot.Language = lang
	}
	if cfg.Bot.Language == "" {
		cfg.Bot.Language = defaultLanguage
	}

	if schedule := os.Getenv("SCHEDULE_WINNER_RESET"); schedule != "" {
		cfg.Schedule.WinnerReset = schedule
	}
	if cfg.Schedule.WinnerReset == "" {
		cfg.Schedule.WinnerReset = defaultWinnerReset
	}

	if schedule := os.Getenv("SCHEDULE_AUTO_ROULETTE"); schedule != "" {
		cfg.Schedule.AutoRoulette = schedule
	}
	if cfg.Schedule.AutoRoulette == "" {
		cfg.Schedule.AutoRoulette = defaultAutoRoulette
	}
}

func setDefaults(cfg *Config) {
	cfg.Bot.Language = defaultLanguage
	cfg.Schedule.WinnerReset = defaultWinnerReset
	cfg.Schedule.AutoRoulette = defaultAutoRoulette
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
