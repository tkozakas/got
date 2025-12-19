package telegram

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"got/internal/app"
	"got/internal/groq"
	"got/pkg/i18n"
)

type mockGroqClient struct {
	models       []string
	currentModel string
	chatResponse string
	chatError    error
}

func newMockGroqClient() *mockGroqClient {
	return &mockGroqClient{
		models: []string{
			"llama-3.3-70b-versatile",
			"llama-3.1-8b-instant",
			"mixtral-8x7b-32768",
			"gemma2-9b-it",
		},
		currentModel: "llama-3.3-70b-versatile",
	}
}

func (m *mockGroqClient) ListModels() []string {
	return m.models
}

func (m *mockGroqClient) SetModel(model string) error {
	for _, mod := range m.models {
		if mod == model {
			m.currentModel = model
			return nil
		}
	}
	return &modelError{model: model}
}

func (m *mockGroqClient) Chat(ctx context.Context, prompt string, history []groq.Message) (string, error) {
	if m.chatError != nil {
		return "", m.chatError
	}
	return m.chatResponse, nil
}

type modelError struct {
	model string
}

func (e *modelError) Error() string {
	return "invalid model: " + e.model
}

type mockRedisClient struct {
	history     map[int64][]groq.Message
	getError    error
	saveError   error
	clearError  error
	returnEmpty bool
}

func newMockRedisClient() *mockRedisClient {
	return &mockRedisClient{
		history: make(map[int64][]groq.Message),
	}
}

func (m *mockRedisClient) GetHistory(ctx context.Context, chatID int64) ([]groq.Message, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	if m.returnEmpty {
		return nil, nil
	}
	return m.history[chatID], nil
}

func (m *mockRedisClient) SaveHistory(ctx context.Context, chatID int64, history []groq.Message) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.history[chatID] = history
	return nil
}

func (m *mockRedisClient) ClearHistory(ctx context.Context, chatID int64) error {
	if m.clearError != nil {
		return m.clearError
	}
	delete(m.history, chatID)
	return nil
}

type GPTClient interface {
	ListModels() []string
	SetModel(model string) error
	Chat(ctx context.Context, prompt string, history []groq.Message) (string, error)
}

type RedisClient interface {
	GetHistory(ctx context.Context, chatID int64) ([]groq.Message, error)
	SaveHistory(ctx context.Context, chatID int64, history []groq.Message) error
	ClearHistory(ctx context.Context, chatID int64) error
}

func newTestTranslator() *i18n.Translator {
	return i18n.NewWithTranslations("en", map[string]string{
		"help_header":          "*Available commands:*\n",
		"cmd_start":            "Start the bot",
		"cmd_help":             "Show available commands",
		"cmd_gpt":              "Chat with AI",
		"cmd_remind":           "Set a reminder",
		"cmd_meme":             "Get a random meme",
		"cmd_sticker":          "Get a random sticker",
		"cmd_fact":             "Get a random fact",
		"cmd_stats":            "Daily winner game",
		"cmd_tts":              "Convert text to speech",
		"welcome":              "Welcome! I am ready.",
		"gpt_usage":            "Usage: /gpt <prompt>",
		"gpt_no_key":           "GPT is not configured.",
		"gpt_cleared":          "Conversation history cleared.",
		"gpt_error":            "Failed to get AI response.",
		"gpt_models_header":    "Available models:\n",
		"gpt_image_usage":      "Usage: /gpt image <prompt>",
		"gpt_model_set":        "Model set to: %s",
		"gpt_model_invalid":    "Invalid model. Available models:\n",
		"gpt_memory_header":    "Memory stats:\n",
		"gpt_memory_stats":     "Messages: %d, Characters: %d",
		"gpt_memory_empty":     "No conversation history.",
		"gpt_memory_no_redis":  "Memory feature is not available.",
		"tts_usage":            "Usage: /tts <text to speak>",
		"tts_error":            "Failed to generate speech.",
		"sticker_usage":        "Reply to a sticker with /sticker add",
		"sticker_added":        "Sticker added!",
		"sticker_error":        "Failed to process sticker.",
		"sticker_remove_usage": "Reply to a sticker with /sticker remove",
		"sticker_removed":      "Sticker removed!",
		"sticker_list_header":  "You have %d stickers saved.",
		"no_stickers":          "No stickers saved yet.",
		"fact_usage":           "Usage: /fact add <text>",
		"fact_added":           "Fact added!",
		"fact_error":           "Failed to process fact.",
		"fact_format":          "Fun fact: %s",
		"no_facts":             "No facts saved yet.",
		"meme_usage":           "Usage: /meme add <subreddit>",
		"meme_added":           "Subreddit r/%s added!",
		"meme_removed":         "Subreddit r/%s removed!",
		"meme_list_header":     "Saved subreddits:\n",
		"meme_error":           "Failed to fetch meme from ",
		"meme_count_invalid":   "Count must be between 1 and 5.",
		"subreddit_error":      "Failed to process subreddit.",
		"remind_usage":         "Usage: /remind <duration> <message>",
		"remind_invalid_time":  "Invalid time format.",
		"remind_success":       "Reminder set for %s.",
		"remind_no_pending":    "No pending reminders.",
		"remind_list_error":    "Failed to list reminders.",
		"remind_header":        "*Pending reminders:*\n",
		"remind_format":        "#%d: %s (at %s)\n",
		"remind_delete_usage":  "Usage: /remind delete <id>",
		"remind_deleted":       "Reminder deleted.",
		"remind_delete_error":  "Failed to delete reminder.",
		"stats_usage":          "Usage: /stats [year|all]",
		"stats_no_stats":       "No stats found.",
		"stats_no_users":       "No users registered.",
		"stats_alias":          "Winner",
		"stats_winner_exists":  "Today's %s: %s with %d points!",
		"stats_winner_new":     "New %s: %s!",
		"stats_header":         "Stats for %d",
		"stats_header_all":     "All-time stats",
		"stats_footer":         "Total: %d users",
		"stats_user":           "%d. %s: %d points",
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
	return NewBotHandlers(client, svc, nil, nil, newTestTranslator(), nil)
}

func newTestBotHandlersWithGPT(client *Client, svc *app.Service, gpt *groq.Client, _ *mockRedisClient) *BotHandlers {
	return &BotHandlers{
		client:  client,
		service: svc,
		gpt:     gpt,
		cache:   nil, // Cache tests require real Redis; use nil for unit tests
		t:       newTestTranslator(),
		tts:     nil,
	}
}

func newTestBotHandlersWithMockGPT(client *Client, svc *app.Service, gpt *mockGroqClient, _ *mockRedisClient) *BotHandlers {
	return &BotHandlers{
		client:  client,
		service: svc,
		gpt:     newMockGroqWrapper(gpt),
		cache:   nil, // Cache tests require real Redis; use nil for unit tests
		t:       newTestTranslator(),
		tts:     nil,
	}
}

// mockGroqWrapper wraps mockGroqClient to satisfy *groq.Client type requirement
// We use a real groq.Client with a test server for model operations
func newMockGroqWrapper(mock *mockGroqClient) *groq.Client {
	if mock == nil {
		return nil
	}
	// Create a real client that we'll use for testing
	return groq.NewClient("test-key")
}

// mockRedisWrapper converts mockRedisClient to satisfy the redis.Client interface
func newMockRedisWrapper(mock *mockRedisClient) *mockRedisClientWrapper {
	if mock == nil {
		return nil
	}
	return &mockRedisClientWrapper{mock: mock}
}

// mockRedisClientWrapper wraps our mock to be used in place of *redis.Client
type mockRedisClientWrapper struct {
	mock *mockRedisClient
}

func (w *mockRedisClientWrapper) GetHistory(ctx context.Context, chatID int64) ([]groq.Message, error) {
	return w.mock.GetHistory(ctx, chatID)
}

func (w *mockRedisClientWrapper) SaveHistory(ctx context.Context, chatID int64, history []groq.Message) error {
	return w.mock.SaveHistory(ctx, chatID, history)
}

func (w *mockRedisClientWrapper) ClearHistory(ctx context.Context, chatID int64) error {
	return w.mock.ClearHistory(ctx, chatID)
}

func TestHandleTTS_NoText(t *testing.T) {
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
			Text: "/tts",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleTTS(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleTTS() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Usage:") {
		t.Errorf("expected usage message, got: %s", sentMessage)
	}
}

func TestHandleTTS_NoTTSClient(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	svc := newTestServiceForHandlers()
	handlers := newTestBotHandlers(client, svc) // TTS client is nil

	update := &Update{
		Message: &Message{
			Text: "/tts Hello world",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleTTS(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleTTS() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Failed") {
		t.Errorf("expected error message when TTS client is nil, got: %s", sentMessage)
	}
}
