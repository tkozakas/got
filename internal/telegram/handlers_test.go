package telegram

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"got/internal/app"
	"got/pkg/i18n"
)

func newTestTranslator() *i18n.Translator {
	return i18n.NewWithTranslations("en", map[string]string{
		"help_header":       "*Available commands:*\n",
		"cmd_start":         "Start the bot",
		"cmd_help":          "Show available commands",
		"cmd_gpt":           "Chat with AI",
		"cmd_remind":        "Set a reminder",
		"cmd_meme":          "Get a random meme",
		"cmd_sticker":       "Get a random sticker",
		"cmd_fact":          "Get a random fact",
		"cmd_stats":         "Daily winner game",
		"welcome":           "Welcome! I am ready.",
		"gpt_usage":         "Usage: /gpt <prompt>",
		"gpt_no_key":        "GPT is not configured.",
		"gpt_cleared":       "Conversation history cleared.",
		"gpt_error":         "Failed to get AI response.",
		"gpt_models_header": "Available models:\n",
	})
}

func TestHandleHelp(t *testing.T) {
	tests := []struct {
		name         string
		wantContains []string
	}{
		{
			name: "Help message contains all commands",
			wantContains: []string{
				"/start",
				"/help",
				"/gpt",
				"/remind",
				"/meme",
				"/sticker",
				"/fact",
				"/stats",
				"Available commands",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sentMessage string
			server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
				payload := decodeJSONPayload(t, r)
				sentMessage = payload["text"].(string)
				w.WriteHeader(http.StatusOK)
			})

			client := newTestClient(server.URL)
			svc := newTestServiceForHandlers()
			handlers := newTestBotHandlers(client, svc)

			update := &Update{
				Message: &Message{
					Text: "/help",
					Chat: &Chat{ID: 123},
				},
			}

			err := handlers.HandleHelp(context.Background(), update)

			if err != nil {
				t.Fatalf("HandleHelp() error = %v", err)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(sentMessage, want) {
					t.Errorf("help message should contain %q, got: %s", want, sentMessage)
				}
			}
		})
	}
}

func TestHandleHelpAllCommandsListed(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	svc := newTestServiceForHandlers()
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/help",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleHelp(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleHelp() error = %v", err)
	}

	expectedCommands := []string{
		"/start",
		"/help",
		"/gpt",
		"/remind",
		"/meme",
		"/sticker",
		"/fact",
		"/stats",
	}

	for _, cmd := range expectedCommands {
		if !strings.Contains(sentMessage, cmd) {
			t.Errorf("help message missing command %q", cmd)
		}
	}

	lines := strings.Split(strings.TrimSpace(sentMessage), "\n")
	commandCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "/") {
			commandCount++
		}
	}

	if commandCount != len(expectedCommands) {
		t.Errorf("expected %d commands, got %d", len(expectedCommands), commandCount)
	}
}

func TestHandleHelpCommandDescriptions(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	svc := newTestServiceForHandlers()
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/help",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleHelp(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleHelp() error = %v", err)
	}

	expectedDescriptions := []string{
		"Start the bot",
		"Show available commands",
		"Chat with AI",
		"Set a reminder",
		"Get a random meme",
		"Get a random sticker",
		"Get a random fact",
		"Daily winner game",
	}

	for _, desc := range expectedDescriptions {
		if !strings.Contains(sentMessage, desc) {
			t.Errorf("help message missing description %q", desc)
		}
	}
}

func newTestServiceForHandlers() *app.Service {
	return app.NewService(
		&mockChatRepo{},
		&mockUserRepo{},
		&mockReminderRepo{},
		&mockFactRepo{},
		&mockStickerRepo{},
		&mockSubredditRepo{},
		&mockStatRepo{},
	)
}

func newTestBotHandlers(client *Client, svc *app.Service) *BotHandlers {
	return NewBotHandlers(client, svc, nil, nil, newTestTranslator())
}
