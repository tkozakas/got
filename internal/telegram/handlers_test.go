package telegram

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"got/internal/app"
	"got/internal/app/model"
	"got/internal/groq"
	"got/internal/redis"
	"got/pkg/config"
	"got/pkg/i18n"
)

func newTestTranslator() *i18n.Translator {
	return i18n.NewWithTranslations("en", map[string]string{
		"help_header":            "*Available commands:*\n",
		"cmd_start":              "Start the bot",
		"cmd_help":               "Show available commands",
		"cmd_gpt":                "Chat with AI",
		"cmd_remind":             "Set a reminder",
		"cmd_meme":               "Get a random meme",
		"cmd_sticker":            "Get a random sticker",
		"cmd_fact":               "Get a random fact",
		"cmd_roulette":           "Daily winner roulette",
		"cmd_tts":                "Convert text to speech",
		"welcome":                "Welcome! I am ready.",
		"gpt_usage":              "Usage: /gpt <prompt>",
		"gpt_no_key":             "GPT is not configured.",
		"gpt_cleared":            "Conversation history cleared.",
		"gpt_error":              "Failed to get AI response.",
		"gpt_models_header":      "Available models:\n",
		"gpt_image_usage":        "Usage: /gpt image <prompt>",
		"gpt_model_set":          "Model set to: %s",
		"gpt_model_invalid":      "Invalid model. Available models:\n",
		"gpt_memory_header":      "Memory stats:\n",
		"gpt_memory_stats":       "Messages: %d, Characters: %d",
		"gpt_memory_empty":       "No conversation history.",
		"gpt_memory_no_redis":    "Memory feature is not available.",
		"tts_usage":              "Usage: /tts <text to speak>",
		"tts_error":              "Failed to generate speech.",
		"sticker_usage":          "Reply to a sticker with /sticker add",
		"sticker_added":          "Sticker added!",
		"sticker_error":          "Failed to process sticker.",
		"sticker_remove_usage":   "Reply to a sticker with /sticker remove",
		"sticker_removed":        "Sticker removed!",
		"sticker_list_header":    "*Available Sticker Sets:*\n\n",
		"no_stickers":            "No stickers saved yet.",
		"fact_usage":             "Usage: /fact add <text>",
		"fact_added":             "Fact added!",
		"fact_error":             "Failed to process fact.",
		"fact_format":            "Fun fact: %s",
		"no_facts":               "No facts saved yet.",
		"meme_usage":             "Usage: /meme add <subreddit>",
		"meme_added":             "Subreddit r/%s added!",
		"meme_removed":           "Subreddit r/%s removed!",
		"meme_list_header":       "Saved subreddits:\n",
		"meme_error":             "Failed to fetch meme from ",
		"meme_count_invalid":     "Count must be between 1 and 5.",
		"subreddit_error":        "Failed to process subreddit.",
		"remind_usage":           "Usage: /remind <duration> <message>",
		"remind_invalid_time":    "Invalid time format.",
		"remind_success":         "Reminder set for %s.",
		"remind_no_pending":      "No pending reminders.",
		"remind_list_error":      "Failed to list reminders.",
		"remind_header":          "*Pending reminders:*\n",
		"remind_format":          "#%d: %s (at %s)\n",
		"remind_delete_usage":    "Usage: /remind delete <id>",
		"remind_deleted":         "Reminder deleted.",
		"remind_delete_error":    "Failed to delete reminder.",
		"roulette_usage":         "Usage: /roulette [year|all]",
		"roulette_no_stats":      "No stats found.",
		"roulette_no_users":      "No users registered.",
		"roulette_alias":         "Winner",
		"roulette_winner_exists": "Today's %s: %s with %d points!",
		"roulette_winner_new":    "New %s: %s!",
		"roulette_header":        "Stats for %d",
		"roulette_header_all":    "All-time stats",
		"roulette_footer":        "Total: %d users",
		"roulette_user":          "%d. %s: %d points",
	})
}

func newTestCommandsConfig() *config.CommandsConfig {
	return &config.CommandsConfig{
		Start:    "start",
		Help:     "help",
		Gpt:      "gpt",
		Remind:   "remind",
		Meme:     "meme",
		Sticker:  "sticker",
		Fact:     "fact",
		Roulette: "roulette",
		Tts:      "tts",
	}
}

func TestHandleHelp(t *testing.T) {
	tests := []struct {
		name         string
		wantContains []string
	}{
		{
			name: "Help message contains all user-facing commands",
			wantContains: []string{
				"/gpt",
				"/remind",
				"/meme",
				"/sticker",
				"/fact",
				"/roulette",
				"/tts",
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

	// User-facing commands (start and help are excluded from help message)
	expectedCommands := []string{
		"/gpt",
		"/remind",
		"/meme",
		"/sticker",
		"/fact",
		"/roulette",
		"/tts",
	}

	for _, cmd := range expectedCommands {
		if !strings.Contains(sentMessage, cmd) {
			t.Errorf("help message missing command %q", cmd)
		}
	}

	// Count commands by looking for markdown-formatted command lines
	lines := strings.Split(strings.TrimSpace(sentMessage), "\n")
	commandCount := 0
	for _, line := range lines {
		if strings.Contains(line, "`/") {
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

	// Descriptions for user-facing commands (start and help are excluded)
	expectedDescriptions := []string{
		"Chat with AI",
		"Set a reminder",
		"Get a random meme",
		"Get a random sticker",
		"Get a random fact",
		"Daily winner roulette",
		"Convert text to speech",
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
	return NewBotHandlers(client, svc, nil, nil, newTestTranslator(), nil, newTestCommandsConfig(), "")
}

func newTestBotHandlersWithGPT(client *Client, svc *app.Service, gpt *groq.Client) *BotHandlers {
	return &BotHandlers{
		client:  client,
		service: svc,
		gpt:     gpt,
		cache:   nil,
		t:       newTestTranslator(),
		tts:     nil,
	}
}

func TestHandleTTSNoText(t *testing.T) {
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

func TestHandleTTSNoClient(t *testing.T) {
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

func TestHandleStartSuccess(t *testing.T) {
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
			Text: "/start",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleStart(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleStart() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Welcome") {
		t.Errorf("expected welcome message, got: %s", sentMessage)
	}
}

func TestHandleRemindNoArgs(t *testing.T) {
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
			Text: "/remind",
			Chat: &Chat{ID: 123},
			From: &User{ID: 456, UserName: "testuser"},
		},
	}

	err := handlers.HandleRemind(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleRemind() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Usage:") {
		t.Errorf("expected usage message, got: %s", sentMessage)
	}
}

func TestHandleRemindInvalidDuration(t *testing.T) {
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
			Text: "/remind invalid test message",
			Chat: &Chat{ID: 123},
			From: &User{ID: 456, UserName: "testuser"},
		},
	}

	err := handlers.HandleRemind(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleRemind() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Invalid") {
		t.Errorf("expected invalid time message, got: %s", sentMessage)
	}
}

func TestHandleRemindSuccess(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	mockChat := &mockChatRepo{
		getFunc: func(ctx context.Context, chatID int64) (*model.Chat, error) {
			return &model.Chat{ChatID: chatID}, nil
		},
	}
	mockUser := &mockUserRepo{
		getFunc: func(ctx context.Context, userID int64) (*model.User, error) {
			return &model.User{UserID: userID, Username: "testuser"}, nil
		},
	}
	mockReminder := &mockReminderRepo{
		saveFunc: func(ctx context.Context, r *model.Reminder) error {
			return nil
		},
	}
	svc := app.NewService(
		mockChat,
		mockUser,
		mockReminder,
		&mockFactRepo{},
		&mockStickerRepo{},
		&mockSubredditRepo{},
		&mockStatRepo{},
	)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/remind 1h test reminder",
			Chat: &Chat{ID: 123},
			From: &User{ID: 456, UserName: "testuser"},
		},
	}

	err := handlers.HandleRemind(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleRemind() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Reminder set") {
		t.Errorf("expected success message, got: %s", sentMessage)
	}
}

func TestHandleRemindListEmpty(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	mockReminder := &mockReminderRepo{
		listByChatFunc: func(ctx context.Context, chatID int64) ([]*model.Reminder, error) {
			return []*model.Reminder{}, nil
		},
	}
	svc := app.NewService(
		&mockChatRepo{},
		&mockUserRepo{},
		mockReminder,
		&mockFactRepo{},
		&mockStickerRepo{},
		&mockSubredditRepo{},
		&mockStatRepo{},
	)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/remind list",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleRemind(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleRemind() error = %v", err)
	}

	if !strings.Contains(sentMessage, "No pending") {
		t.Errorf("expected no pending reminders message, got: %s", sentMessage)
	}
}

func TestHandleRemindListWithReminders(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	mockReminder := &mockReminderRepo{
		listByChatFunc: func(ctx context.Context, chatID int64) ([]*model.Reminder, error) {
			chat := &model.Chat{ChatID: chatID}
			return []*model.Reminder{
				{ReminderID: 1, Chat: chat, Message: "Test reminder 1"},
				{ReminderID: 2, Chat: chat, Message: "Test reminder 2"},
			}, nil
		},
	}
	svc := app.NewService(
		&mockChatRepo{},
		&mockUserRepo{},
		mockReminder,
		&mockFactRepo{},
		&mockStickerRepo{},
		&mockSubredditRepo{},
		&mockStatRepo{},
	)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/remind list",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleRemind(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleRemind() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Pending reminders") {
		t.Errorf("expected reminders list header, got: %s", sentMessage)
	}
}

func TestHandleRemindDeleteNoID(t *testing.T) {
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
			Text: "/remind delete",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleRemind(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleRemind() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Usage:") {
		t.Errorf("expected delete usage message, got: %s", sentMessage)
	}
}

func TestHandleRemindDeleteSuccess(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	mockReminder := &mockReminderRepo{
		deleteFunc: func(ctx context.Context, reminderID int64, chatID int64) error {
			return nil
		},
	}
	svc := app.NewService(
		&mockChatRepo{},
		&mockUserRepo{},
		mockReminder,
		&mockFactRepo{},
		&mockStickerRepo{},
		&mockSubredditRepo{},
		&mockStatRepo{},
	)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/remind delete 1",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleRemind(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleRemind() error = %v", err)
	}

	if !strings.Contains(sentMessage, "deleted") {
		t.Errorf("expected reminder deleted message, got: %s", sentMessage)
	}
}

func TestHandleRouletteNoUsers(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	mockUser := &mockUserRepo{}
	mockStat := &mockStatRepo{
		findWinnerByChatFunc: func(ctx context.Context, chatID int64, year int) (*model.Stat, error) {
			return nil, nil
		},
	}
	svc := app.NewService(
		&mockChatRepo{},
		mockUser,
		&mockReminderRepo{},
		&mockFactRepo{},
		&mockStickerRepo{},
		&mockSubredditRepo{},
		mockStat,
	)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/roulette",
			Chat: &Chat{ID: 123},
			From: &User{ID: 456, UserName: "testuser"},
		},
	}

	err := handlers.HandleRoulette(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleRoulette() error = %v", err)
	}

	if !strings.Contains(sentMessage, "No users") {
		t.Errorf("expected no users message, got: %s", sentMessage)
	}
}

func TestHandleRouletteWithWinner(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	mockStat := &mockStatRepo{
		findByUserChatYearFunc: func(ctx context.Context, userID, chatID int64, year int) (*model.Stat, error) {
			return nil, nil
		},
		findWinnerByChatFunc: func(ctx context.Context, chatID int64, year int) (*model.Stat, error) {
			return &model.Stat{
				StatID:   1,
				Score:    5,
				IsWinner: true,
				User:     &model.User{UserID: 1, Username: "user1"},
			}, nil
		},
	}
	svc := app.NewService(
		&mockChatRepo{},
		&mockUserRepo{},
		&mockReminderRepo{},
		&mockFactRepo{},
		&mockStickerRepo{},
		&mockSubredditRepo{},
		mockStat,
	)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/roulette",
			Chat: &Chat{ID: 123},
			From: &User{ID: 456, UserName: "testuser"},
		},
	}

	err := handlers.HandleRoulette(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleRoulette() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Winner") && !strings.Contains(sentMessage, "user1") {
		t.Errorf("expected winner message, got: %s", sentMessage)
	}
}

func TestHandleRouletteYear(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	mockStat := &mockStatRepo{
		listByChatAndYearFunc: func(ctx context.Context, chatID int64, year int) ([]*model.Stat, error) {
			return []*model.Stat{
				{StatID: 1, Score: 10, User: &model.User{UserID: 1, Username: "user1"}},
				{StatID: 2, Score: 5, User: &model.User{UserID: 2, Username: "user2"}},
			}, nil
		},
	}
	svc := app.NewService(
		&mockChatRepo{},
		&mockUserRepo{},
		&mockReminderRepo{},
		&mockFactRepo{},
		&mockStickerRepo{},
		&mockSubredditRepo{},
		mockStat,
	)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/roulette stats",
			Chat: &Chat{ID: 123},
			From: &User{ID: 456, UserName: "testuser"},
		},
	}

	err := handlers.HandleRoulette(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleRoulette() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Stats for") {
		t.Errorf("expected year stats header, got: %s", sentMessage)
	}
}

func TestHandleRouletteAll(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	mockStat := &mockStatRepo{
		listByChatFunc: func(ctx context.Context, chatID int64) ([]*model.Stat, error) {
			return []*model.Stat{
				{StatID: 1, Score: 100, User: &model.User{UserID: 1, Username: "user1"}},
				{StatID: 2, Score: 50, User: &model.User{UserID: 2, Username: "user2"}},
			}, nil
		},
	}
	svc := app.NewService(
		&mockChatRepo{},
		&mockUserRepo{},
		&mockReminderRepo{},
		&mockFactRepo{},
		&mockStickerRepo{},
		&mockSubredditRepo{},
		mockStat,
	)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/roulette all",
			Chat: &Chat{ID: 123},
			From: &User{ID: 456, UserName: "testuser"},
		},
	}

	err := handlers.HandleRoulette(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleRoulette() error = %v", err)
	}

	if !strings.Contains(sentMessage, "All-time") {
		t.Errorf("expected all-time stats header, got: %s", sentMessage)
	}
}

func TestHandleGPTForget(t *testing.T) {
	tests := []struct {
		name         string
		command      string
		wantContains string
	}{
		{
			name:         "forgetClearsHistory",
			command:      "/gpt forget",
			wantContains: "cleared",
		},
		{
			name:         "clearClearsHistory",
			command:      "/gpt clear",
			wantContains: "cleared",
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
			gpt := groq.NewClient("test-key")
			handlers := newTestBotHandlersWithGPT(client, svc, gpt)

			update := &Update{
				Message: &Message{
					Text: tt.command,
					Chat: &Chat{ID: 123},
				},
			}

			err := handlers.HandleGPT(context.Background(), update)
			if err != nil {
				t.Fatalf("HandleGPT() error = %v", err)
			}

			if !strings.Contains(strings.ToLower(sentMessage), tt.wantContains) {
				t.Errorf("expected message containing %q, got: %s", tt.wantContains, sentMessage)
			}
		})
	}
}

func TestHandleGPTImageNoPrompt(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	svc := newTestServiceForHandlers()
	gpt := groq.NewClient("test-key")
	handlers := newTestBotHandlersWithGPT(client, svc, gpt)

	update := &Update{
		Message: &Message{
			Text: "/gpt image",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleGPT(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleGPT() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Usage:") {
		t.Errorf("expected usage message, got: %s", sentMessage)
	}
}

func TestHandleGPTImageSuccess(t *testing.T) {
	var photoURL string
	var caption string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		if url, ok := payload["photo"].(string); ok {
			photoURL = url
		}
		if cap, ok := payload["caption"].(string); ok {
			caption = cap
		}
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	svc := newTestServiceForHandlers()
	gpt := groq.NewClient("test-key")
	handlers := newTestBotHandlersWithGPT(client, svc, gpt)

	update := &Update{
		Message: &Message{
			Text: "/gpt image a cute cat",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleGPT(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleGPT() error = %v", err)
	}

	if !strings.Contains(photoURL, "image.pollinations.ai") {
		t.Errorf("expected pollinations URL, got: %s", photoURL)
	}
	if !strings.Contains(photoURL, "width=1024") {
		t.Errorf("expected width param in URL, got: %s", photoURL)
	}
	if !strings.Contains(photoURL, "nologo=true") {
		t.Errorf("expected nologo param in URL, got: %s", photoURL)
	}
	if caption != "a cute cat" {
		t.Errorf("expected caption 'a cute cat', got: %s", caption)
	}
}

func TestFormatPromptWithUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		prompt   string
		want     string
	}{
		{
			name:     "withUsername",
			username: "john",
			prompt:   "hello",
			want:     "john: hello",
		},
		{
			name:     "emptyUsername",
			username: "",
			prompt:   "hello",
			want:     "hello",
		},
		{
			name:     "whitespaceUsername",
			username: "   ",
			prompt:   "hello",
			want:     "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPromptWithUsername(tt.username, tt.prompt)
			if got != tt.want {
				t.Errorf("formatPromptWithUsername() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatHistoryAsText(t *testing.T) {
	tests := []struct {
		name    string
		history []groq.Message
		want    string
	}{
		{
			name:    "emptyHistory",
			history: []groq.Message{},
			want:    "",
		},
		{
			name: "singleMessage",
			history: []groq.Message{
				{Role: "user", Content: "Hello"},
			},
			want: "user: Hello\n",
		},
		{
			name: "multipleMessages",
			history: []groq.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
				{Role: "user", Content: "How are you?"},
			},
			want: "user: Hello\nassistant: Hi there!\nuser: How are you?\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatHistoryAsText(tt.history)
			if got != tt.want {
				t.Errorf("formatHistoryAsText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHandleFactNoFacts(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	mockFact := &mockFactRepo{
		getRandomByChatFunc: func(ctx context.Context, chatID int64) (*model.Fact, error) {
			return nil, nil
		},
	}
	svc := app.NewService(
		&mockChatRepo{},
		&mockUserRepo{},
		&mockReminderRepo{},
		mockFact,
		&mockStickerRepo{},
		&mockSubredditRepo{},
		&mockStatRepo{},
	)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/fact",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleFact(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleFact() error = %v", err)
	}

	if !strings.Contains(sentMessage, "No facts") {
		t.Errorf("expected no facts message, got: %s", sentMessage)
	}
}

func TestHandleFactWithFact(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	mockFact := &mockFactRepo{
		getRandomByChatFunc: func(ctx context.Context, chatID int64) (*model.Fact, error) {
			return &model.Fact{Comment: "The sky is blue"}, nil
		},
	}
	svc := app.NewService(
		&mockChatRepo{},
		&mockUserRepo{},
		&mockReminderRepo{},
		mockFact,
		&mockStickerRepo{},
		&mockSubredditRepo{},
		&mockStatRepo{},
	)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/fact",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleFact(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleFact() error = %v", err)
	}

	if !strings.Contains(sentMessage, "The sky is blue") {
		t.Errorf("expected fact content, got: %s", sentMessage)
	}
}

func TestHandleFactAddNoText(t *testing.T) {
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
			Text: "/fact add",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleFact(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleFact() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Usage") {
		t.Errorf("expected usage message, got: %s", sentMessage)
	}
}

func TestHandleStickerNoStickers(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	mockSticker := &mockStickerRepo{
		getRandomByChatFunc: func(ctx context.Context, chatID int64) (*model.Sticker, error) {
			return nil, nil
		},
	}
	svc := app.NewService(
		&mockChatRepo{},
		&mockUserRepo{},
		&mockReminderRepo{},
		&mockFactRepo{},
		mockSticker,
		&mockSubredditRepo{},
		&mockStatRepo{},
	)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/sticker",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleSticker(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleSticker() error = %v", err)
	}

	if !strings.Contains(sentMessage, "No stickers") {
		t.Errorf("expected no stickers message, got: %s", sentMessage)
	}
}

func TestHandleStickerList(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	mockSticker := &mockStickerRepo{
		listByChatFunc: func(ctx context.Context, chatID int64) ([]*model.Sticker, error) {
			return []*model.Sticker{
				{FileID: "abc", StickerSetName: "cool_stickers"},
				{FileID: "def", StickerSetName: "funny_pack"},
			}, nil
		},
	}
	svc := app.NewService(
		&mockChatRepo{},
		&mockUserRepo{},
		&mockReminderRepo{},
		&mockFactRepo{},
		mockSticker,
		&mockSubredditRepo{},
		&mockStatRepo{},
	)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/sticker list",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleSticker(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleSticker() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Available Sticker Sets") {
		t.Errorf("expected sticker list header, got: %s", sentMessage)
	}
	if !strings.Contains(sentMessage, "cool_stickers") || !strings.Contains(sentMessage, "funny_pack") {
		t.Errorf("expected sticker set names in list, got: %s", sentMessage)
	}
}

func TestHandleStickerAddNoReply(t *testing.T) {
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
			Text: "/sticker add",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleSticker(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleSticker() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Reply") {
		t.Errorf("expected reply instruction, got: %s", sentMessage)
	}
}

func TestHandleMemeListSubreddits(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	mockSub := &mockSubredditRepo{
		listByChatFunc: func(ctx context.Context, chatID int64) ([]*model.Subreddit, error) {
			return []*model.Subreddit{{Name: "funny"}, {Name: "memes"}}, nil
		},
	}
	svc := app.NewService(
		&mockChatRepo{},
		&mockUserRepo{},
		&mockReminderRepo{},
		&mockFactRepo{},
		&mockStickerRepo{},
		mockSub,
		&mockStatRepo{},
	)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/meme list",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleMeme(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleMeme() error = %v", err)
	}

	if !strings.Contains(sentMessage, "funny") || !strings.Contains(sentMessage, "memes") {
		t.Errorf("expected subreddit list, got: %s", sentMessage)
	}
}

func TestHandleMemeInvalidCount(t *testing.T) {
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
			Text: "/meme 10",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleMeme(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleMeme() error = %v", err)
	}

	if !strings.Contains(sentMessage, "1 and 5") {
		t.Errorf("expected count invalid message, got: %s", sentMessage)
	}
}

func TestHandleGPTNoKey(t *testing.T) {
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
			Text: "/gpt hello",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleGPT(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleGPT() error = %v", err)
	}

	if !strings.Contains(sentMessage, "not configured") {
		t.Errorf("expected not configured message, got: %s", sentMessage)
	}
}

func TestHandleGPTNoPrompt(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	svc := newTestServiceForHandlers()
	gpt := groq.NewClient("test-key")
	handlers := newTestBotHandlersWithGPT(client, svc, gpt)

	update := &Update{
		Message: &Message{
			Text: "/gpt",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleGPT(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleGPT() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Usage") {
		t.Errorf("expected usage message, got: %s", sentMessage)
	}
}

func TestIsAnimatedURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "gifExtension", url: "https://example.com/image.gif", want: true},
		{name: "gifWithQuery", url: "https://example.com/image.gif?v=1", want: true},
		{name: "mp4Extension", url: "https://example.com/video.mp4", want: true},
		{name: "mp4WithQuery", url: "https://example.com/video.mp4?v=1", want: true},
		{name: "jpgExtension", url: "https://example.com/image.jpg", want: false},
		{name: "pngExtension", url: "https://example.com/image.png", want: false},
		{name: "noExtension", url: "https://example.com/image", want: false},
		{name: "upperCaseGif", url: "https://example.com/image.GIF", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAnimatedURL(tt.url)
			if got != tt.want {
				t.Errorf("isAnimatedURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestBuildImageURL(t *testing.T) {
	tests := []struct {
		name         string
		prompt       string
		wantContains []string
	}{
		{
			name:   "simplePrompt",
			prompt: "a cat",
			wantContains: []string{
				"https://image.pollinations.ai/prompt/a+cat",
				"width=1024",
				"height=1024",
				"nologo=true",
				"enhance=true",
				"seed=",
			},
		},
		{
			name:   "promptWithSpecialChars",
			prompt: "a cat & dog",
			wantContains: []string{
				"a+cat+%26+dog",
				"width=1024",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildImageURL(tt.prompt)
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("buildImageURL() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func newTestBotHandlersWithGPTAndCache(client *Client, svc *app.Service, gpt *groq.Client, cache *redis.Client) *BotHandlers {
	return &BotHandlers{
		client:  client,
		service: svc,
		gpt:     gpt,
		cache:   cache,
		t:       newTestTranslator(),
		tts:     nil,
	}
}

func TestHandleGPTMemoryNoRedis(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	svc := newTestServiceForHandlers()
	gpt := groq.NewClient("test-key")
	handlers := newTestBotHandlersWithGPT(client, svc, gpt) // cache is nil

	update := &Update{
		Message: &Message{
			Text: "/gpt memory",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleGPT(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleGPT() error = %v", err)
	}

	if !strings.Contains(sentMessage, "not available") {
		t.Errorf("expected 'not available' message, got: %s", sentMessage)
	}
}

func TestHandleGPTMemoryEmpty(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	svc := newTestServiceForHandlers()
	gpt := groq.NewClient("test-key")
	// Use a redis client with invalid address - GetHistory will fail, triggering "No conversation history"
	cache := redis.NewClient("invalid:9999")
	handlers := newTestBotHandlersWithGPTAndCache(client, svc, gpt, cache)

	update := &Update{
		Message: &Message{
			Text: "/gpt memory",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleGPT(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleGPT() error = %v", err)
	}

	if !strings.Contains(sentMessage, "No conversation history") {
		t.Errorf("expected 'No conversation history' message, got: %s", sentMessage)
	}
}

func TestHandleGPTModels(t *testing.T) {
	var sentMessage string
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		sentMessage = payload["text"].(string)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	svc := newTestServiceForHandlers()
	gpt := groq.NewClient("test-key")
	handlers := newTestBotHandlersWithGPT(client, svc, gpt)

	update := &Update{
		Message: &Message{
			Text: "/gpt model",
			Chat: &Chat{ID: 123},
		},
	}

	err := handlers.HandleGPT(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleGPT() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Available models") {
		t.Errorf("expected 'Available models' header, got: %s", sentMessage)
	}

	expectedModels := []string{
		"llama-3.3-70b-versatile",
		"llama-3.1-8b-instant",
		"mixtral-8x7b-32768",
		"gemma2-9b-it",
	}
	for _, model := range expectedModels {
		if !strings.Contains(sentMessage, model) {
			t.Errorf("expected model %q in response, got: %s", model, sentMessage)
		}
	}
}

func TestRegisterChatUser(t *testing.T) {
	tests := []struct {
		name          string
		msg           *Message
		wantChatSaved bool
		wantUserSaved bool
	}{
		{
			name:          "nilChat",
			msg:           &Message{From: &User{ID: 1}},
			wantChatSaved: false,
			wantUserSaved: false,
		},
		{
			name:          "nilFrom",
			msg:           &Message{Chat: &Chat{ID: 1}},
			wantChatSaved: false,
			wantUserSaved: false,
		},
		{
			name:          "bothNil",
			msg:           &Message{},
			wantChatSaved: false,
			wantUserSaved: false,
		},
		{
			name: "validChatAndUser",
			msg: &Message{
				Chat: &Chat{ID: 123, Title: "Test Chat"},
				From: &User{ID: 456, UserName: "testuser"},
			},
			wantChatSaved: true,
			wantUserSaved: true,
		},
		{
			name: "userWithoutUsername",
			msg: &Message{
				Chat: &Chat{ID: 123, Title: "Test Chat"},
				From: &User{ID: 456, FirstName: "John"},
			},
			wantChatSaved: true,
			wantUserSaved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatSaved := false
			userSaved := false

			mockChat := &mockChatRepo{
				saveFunc: func(ctx context.Context, chat *model.Chat) error {
					chatSaved = true
					return nil
				},
			}
			mockUser := &mockUserRepo{
				saveFunc: func(ctx context.Context, user *model.User) error {
					userSaved = true
					return nil
				},
			}

			svc := app.NewService(
				mockChat,
				mockUser,
				&mockReminderRepo{},
				&mockFactRepo{},
				&mockStickerRepo{},
				&mockSubredditRepo{},
				&mockStatRepo{},
			)

			server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			client := newTestClient(server.URL)
			handlers := newTestBotHandlers(client, svc)

			handlers.registerChatUser(context.Background(), tt.msg)

			if chatSaved != tt.wantChatSaved {
				t.Errorf("chat saved = %v, want %v", chatSaved, tt.wantChatSaved)
			}
			if userSaved != tt.wantUserSaved {
				t.Errorf("user saved = %v, want %v", userSaved, tt.wantUserSaved)
			}
		})
	}
}

func TestRouletteRegistersUser(t *testing.T) {
	chatSaved := false
	userSaved := false

	mockChat := &mockChatRepo{
		saveFunc: func(ctx context.Context, chat *model.Chat) error {
			chatSaved = true
			return nil
		},
	}
	mockUser := &mockUserRepo{
		saveFunc: func(ctx context.Context, user *model.User) error {
			userSaved = true
			return nil
		},
	}

	svc := app.NewService(
		mockChat,
		mockUser,
		&mockReminderRepo{},
		&mockFactRepo{},
		&mockStickerRepo{},
		&mockSubredditRepo{},
		&mockStatRepo{},
	)

	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	client := newTestClient(server.URL)
	handlers := newTestBotHandlers(client, svc)

	update := &Update{
		Message: &Message{
			Text: "/roulette stats",
			Chat: &Chat{ID: 123, Title: "Test Chat"},
			From: &User{ID: 456, UserName: "testuser"},
		},
	}

	_ = handlers.HandleRoulette(context.Background(), update)

	if !chatSaved {
		t.Error("expected chat to be saved on /roulette")
	}
	if !userSaved {
		t.Error("expected user to be saved on /roulette")
	}
}
