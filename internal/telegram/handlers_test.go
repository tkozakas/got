package telegram

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"got/internal/app"
	"got/internal/app/model"
	"got/internal/groq"
	"got/pkg/i18n"
)

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

func TestHandleStatsNoUsers(t *testing.T) {
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
			Text: "/stats",
			Chat: &Chat{ID: 123},
			From: &User{ID: 456, UserName: "testuser"},
		},
	}

	err := handlers.HandleStats(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleStats() error = %v", err)
	}

	if !strings.Contains(sentMessage, "No users") {
		t.Errorf("expected no users message, got: %s", sentMessage)
	}
}

func TestHandleStatsWithWinner(t *testing.T) {
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
			Text: "/stats",
			Chat: &Chat{ID: 123},
			From: &User{ID: 456, UserName: "testuser"},
		},
	}

	err := handlers.HandleStats(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleStats() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Winner") && !strings.Contains(sentMessage, "user1") {
		t.Errorf("expected winner message, got: %s", sentMessage)
	}
}

func TestHandleStatsYear(t *testing.T) {
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
			Text: "/stats stats",
			Chat: &Chat{ID: 123},
			From: &User{ID: 456, UserName: "testuser"},
		},
	}

	err := handlers.HandleStats(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleStats() error = %v", err)
	}

	if !strings.Contains(sentMessage, "Stats for") {
		t.Errorf("expected year stats header, got: %s", sentMessage)
	}
}

func TestHandleStatsAll(t *testing.T) {
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
			Text: "/stats all",
			Chat: &Chat{ID: 123},
			From: &User{ID: 456, UserName: "testuser"},
		},
	}

	err := handlers.HandleStats(context.Background(), update)
	if err != nil {
		t.Fatalf("HandleStats() error = %v", err)
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
			return []*model.Sticker{{FileID: "abc"}, {FileID: "def"}}, nil
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

	if !strings.Contains(sentMessage, "2") {
		t.Errorf("expected sticker count, got: %s", sentMessage)
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
