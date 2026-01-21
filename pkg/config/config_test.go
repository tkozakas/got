package config

import (
	"os"
	"testing"
)

func TestIsEnvTrue(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{"true lowercase", "true", true},
		{"TRUE uppercase", "TRUE", true},
		{"True mixed", "True", true},
		{"1", "1", true},
		{"yes lowercase", "yes", true},
		{"YES uppercase", "YES", true},
		{"false", "false", false},
		{"0", "0", false},
		{"no", "no", false},
		{"empty", "", false},
		{"random", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEST_ENV_KEY", tt.envValue)
			defer os.Unsetenv("TEST_ENV_KEY")

			if got := isEnvTrue("TEST_ENV_KEY"); got != tt.want {
				t.Errorf("isEnvTrue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyDisabledCommands(t *testing.T) {
	cfg := &Config{}
	setDefaults(cfg)

	os.Setenv("DISABLE_CMD_MEME", "true")
	os.Setenv("DISABLE_CMD_GPT", "1")
	defer func() {
		os.Unsetenv("DISABLE_CMD_MEME")
		os.Unsetenv("DISABLE_CMD_GPT")
	}()

	applyDisabledCommands(cfg)

	if !cfg.IsDisabled("meme") {
		t.Error("meme command should be disabled")
	}
	if !cfg.IsDisabled("gpt") {
		t.Error("gpt command should be disabled")
	}
	if cfg.IsDisabled("start") {
		t.Error("start command should not be disabled")
	}
	if cfg.IsDisabled("help") {
		t.Error("help command should not be disabled")
	}
}

func TestIsDisabled(t *testing.T) {
	cfg := &Config{
		DisabledCommands: map[string]bool{
			"meme": true,
			"gpt":  true,
		},
	}

	if !cfg.IsDisabled("meme") {
		t.Error("meme should be disabled")
	}
	if !cfg.IsDisabled("gpt") {
		t.Error("gpt should be disabled")
	}
	if cfg.IsDisabled("start") {
		t.Error("start should not be disabled")
	}
	if cfg.IsDisabled("nonexistent") {
		t.Error("nonexistent should not be disabled")
	}
}

func TestIsDisabledWithNilMap(t *testing.T) {
	cfg := &Config{}

	if cfg.IsDisabled("meme") {
		t.Error("should return false for nil map")
	}
}

func TestApplyDisabledCommandsWithCustomCommandNames(t *testing.T) {
	cfg := &Config{}
	setDefaults(cfg)
	cfg.Commands.Meme = "custommeme"

	os.Setenv("DISABLE_CMD_MEME", "true")
	defer os.Unsetenv("DISABLE_CMD_MEME")

	applyDisabledCommands(cfg)

	if !cfg.IsDisabled("custommeme") {
		t.Error("custommeme command should be disabled")
	}
	if cfg.IsDisabled("meme") {
		t.Error("default meme command name should not be in disabled list")
	}
}
