package app

import (
	"context"
	"got/internal/app/model"
	"testing"
	"time"
)

func TestServiceRegisterChat(t *testing.T) {
	chatRepo := &MockChatRepository{}
	svc := NewService(chatRepo, &MockUserRepository{}, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, &MockSubredditRepository{}, &MockStatRepository{})

	chat := &model.Chat{ChatID: 1, ChatName: "test"}

	chatRepo.SaveFunc = func(ctx context.Context, c *model.Chat) error {
		if c.ChatID != chat.ChatID {
			t.Errorf("want chat ID %d, got %d", chat.ChatID, c.ChatID)
		}
		return nil
	}

	if err := svc.RegisterChat(context.Background(), chat); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	chatRepo.SaveFunc = func(ctx context.Context, c *model.Chat) error {
		return errMock
	}

	if err := svc.RegisterChat(context.Background(), chat); err != errMock {
		t.Errorf("want error %v, got %v", errMock, err)
	}
}

func TestServiceRegisterUser(t *testing.T) {
	userRepo := &MockUserRepository{}
	svc := NewService(&MockChatRepository{}, userRepo, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, &MockSubredditRepository{}, &MockStatRepository{})

	user := &model.User{UserID: 1, Username: "test"}

	userRepo.SaveFunc = func(ctx context.Context, u *model.User) error {
		if u.UserID != user.UserID {
			t.Errorf("want user ID %d, got %d", user.UserID, u.UserID)
		}
		return nil
	}

	if err := svc.RegisterUser(context.Background(), user, 1); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestServiceAddFact(t *testing.T) {
	chatRepo := &MockChatRepository{}
	factRepo := &MockFactRepository{}
	svc := NewService(chatRepo, &MockUserRepository{}, &MockReminderRepository{}, factRepo, &MockStickerRepository{}, &MockSubredditRepository{}, &MockStatRepository{})

	chat := &model.Chat{ChatID: 1}
	text := "interesting fact"

	chatRepo.GetFunc = func(ctx context.Context, id int64) (*model.Chat, error) {
		return chat, nil
	}

	factRepo.SaveFunc = func(ctx context.Context, f *model.Fact) error {
		if f.Comment != text {
			t.Errorf("want text %s, got %s", text, f.Comment)
		}
		if f.Chat.ChatID != chat.ChatID {
			t.Errorf("want chat ID %d, got %d", chat.ChatID, f.Chat.ChatID)
		}
		return nil
	}

	if err := svc.AddFact(context.Background(), text, 1); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestServiceAddReminder(t *testing.T) {
	chatRepo := &MockChatRepository{}
	userRepo := &MockUserRepository{}
	reminderRepo := &MockReminderRepository{}
	svc := NewService(chatRepo, userRepo, reminderRepo, &MockFactRepository{}, &MockStickerRepository{}, &MockSubredditRepository{}, &MockStatRepository{})

	chat := &model.Chat{ChatID: 1}
	user := &model.User{UserID: 1}
	msg := "remind me"
	dur := time.Second

	chatRepo.GetFunc = func(ctx context.Context, id int64) (*model.Chat, error) {
		return chat, nil
	}
	userRepo.GetFunc = func(ctx context.Context, id int64) (*model.User, error) {
		return user, nil
	}

	reminderRepo.SaveFunc = func(ctx context.Context, r *model.Reminder) error {
		if r.Message != msg {
			t.Errorf("want msg %s, got %s", msg, r.Message)
		}
		return nil
	}

	if err := svc.AddReminder(context.Background(), 1, 1, msg, dur); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test missing chat
	chatRepo.GetFunc = func(ctx context.Context, id int64) (*model.Chat, error) {
		return nil, nil // Chat not found
	}
	if err := svc.AddReminder(context.Background(), 1, 1, msg, dur); err == nil {
		t.Error("want error for missing chat, got nil")
	}
}

func TestServiceCheckReminders(t *testing.T) {
	reminderRepo := &MockReminderRepository{}
	svc := NewService(&MockChatRepository{}, &MockUserRepository{}, reminderRepo, &MockFactRepository{}, &MockStickerRepository{}, &MockSubredditRepository{}, &MockStatRepository{})

	reminders := []*model.Reminder{
		{ReminderID: 1},
		{ReminderID: 2},
	}

	reminderRepo.ListPendingFunc = func(ctx context.Context) ([]*model.Reminder, error) {
		return reminders, nil
	}

	sentIDs := make(map[int64]bool)
	reminderRepo.MarkSentFunc = func(ctx context.Context, id int64) error {
		sentIDs[id] = true
		return nil
	}

	got, err := svc.CheckReminders(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("want 2 reminders, got %d", len(got))
	}

	if !sentIDs[1] || !sentIDs[2] {
		t.Error("reminders were not marked as sent")
	}
}

func TestServiceAddSticker(t *testing.T) {
	tests := []struct {
		name      string
		fileID    string
		chatID    int64
		chatFound bool
		saveErr   error
		wantErr   bool
	}{
		{
			name:      "Success",
			fileID:    "sticker123",
			chatID:    1,
			chatFound: true,
			wantErr:   false,
		},
		{
			name:      "ChatNotFound",
			fileID:    "sticker123",
			chatID:    1,
			chatFound: false,
			wantErr:   true,
		},
		{
			name:      "SaveError",
			fileID:    "sticker123",
			chatID:    1,
			chatFound: true,
			saveErr:   errMock,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRepo := &MockChatRepository{}
			stickerRepo := &MockStickerRepository{}
			svc := NewService(chatRepo, &MockUserRepository{}, &MockReminderRepository{}, &MockFactRepository{}, stickerRepo, &MockSubredditRepository{}, &MockStatRepository{})

			chatRepo.GetFunc = func(ctx context.Context, id int64) (*model.Chat, error) {
				if tt.chatFound {
					return &model.Chat{ChatID: id}, nil
				}
				return nil, nil
			}

			stickerRepo.SaveFunc = func(ctx context.Context, s *model.Sticker) error {
				if s.FileID != tt.fileID {
					t.Errorf("want fileID %s, got %s", tt.fileID, s.FileID)
				}
				return tt.saveErr
			}

			err := svc.AddSticker(context.Background(), tt.fileID, tt.chatID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddSticker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServiceGetRandomSticker(t *testing.T) {
	stickerRepo := &MockStickerRepository{}
	svc := NewService(&MockChatRepository{}, &MockUserRepository{}, &MockReminderRepository{}, &MockFactRepository{}, stickerRepo, &MockSubredditRepository{}, &MockStatRepository{})

	expected := &model.Sticker{FileID: "random123"}
	stickerRepo.GetRandomByChatFunc = func(ctx context.Context, chatID int64) (*model.Sticker, error) {
		return expected, nil
	}

	got, err := svc.GetRandomSticker(context.Background(), 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got.FileID != expected.FileID {
		t.Errorf("want fileID %s, got %s", expected.FileID, got.FileID)
	}
}

func TestServiceListStickers(t *testing.T) {
	stickerRepo := &MockStickerRepository{}
	svc := NewService(&MockChatRepository{}, &MockUserRepository{}, &MockReminderRepository{}, &MockFactRepository{}, stickerRepo, &MockSubredditRepository{}, &MockStatRepository{})

	expected := []*model.Sticker{{FileID: "a"}, {FileID: "b"}}
	stickerRepo.ListByChatFunc = func(ctx context.Context, chatID int64) ([]*model.Sticker, error) {
		return expected, nil
	}

	got, err := svc.ListStickers(context.Background(), 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(got) != len(expected) {
		t.Errorf("want %d stickers, got %d", len(expected), len(got))
	}
}

func TestServiceSubredditOperations(t *testing.T) {
	t.Run("AddSubreddit", func(t *testing.T) {
		subRepo := &MockSubredditRepository{}
		svc := NewService(&MockChatRepository{}, &MockUserRepository{}, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, subRepo, &MockStatRepository{})

		subRepo.SaveFunc = func(ctx context.Context, s *model.Subreddit) error {
			if s.Name != "golang" {
				t.Errorf("want name golang, got %s", s.Name)
			}
			if s.ChatID != 1 {
				t.Errorf("want chatID 1, got %d", s.ChatID)
			}
			return nil
		}

		if err := svc.AddSubreddit(context.Background(), "golang", 1); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("GetRandomSubreddit", func(t *testing.T) {
		subRepo := &MockSubredditRepository{}
		svc := NewService(&MockChatRepository{}, &MockUserRepository{}, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, subRepo, &MockStatRepository{})

		expected := &model.Subreddit{Name: "programmerhumor"}
		subRepo.GetRandomByChatFunc = func(ctx context.Context, chatID int64) (*model.Subreddit, error) {
			return expected, nil
		}

		got, err := svc.GetRandomSubreddit(context.Background(), 1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if got.Name != expected.Name {
			t.Errorf("want name %s, got %s", expected.Name, got.Name)
		}
	})

	t.Run("ListSubreddits", func(t *testing.T) {
		subRepo := &MockSubredditRepository{}
		svc := NewService(&MockChatRepository{}, &MockUserRepository{}, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, subRepo, &MockStatRepository{})

		expected := []*model.Subreddit{{Name: "golang"}, {Name: "rust"}}
		subRepo.ListByChatFunc = func(ctx context.Context, chatID int64) ([]*model.Subreddit, error) {
			return expected, nil
		}

		got, err := svc.ListSubreddits(context.Background(), 1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(got) != len(expected) {
			t.Errorf("want %d subreddits, got %d", len(expected), len(got))
		}
	})

	t.Run("RemoveSubreddit", func(t *testing.T) {
		subRepo := &MockSubredditRepository{}
		svc := NewService(&MockChatRepository{}, &MockUserRepository{}, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, subRepo, &MockStatRepository{})

		deleteCalled := false
		subRepo.DeleteFunc = func(ctx context.Context, name string, chatID int64) error {
			deleteCalled = true
			if name != "golang" {
				t.Errorf("want name golang, got %s", name)
			}
			return nil
		}

		if err := svc.RemoveSubreddit(context.Background(), "golang", 1); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !deleteCalled {
			t.Error("Delete was not called")
		}
	})
}

func TestServiceGetOrCreateStat(t *testing.T) {
	tests := []struct {
		name         string
		existingStat *model.Stat
		userFound    bool
		chatFound    bool
		wantErr      bool
	}{
		{
			name:         "ExistingStatReturned",
			existingStat: &model.Stat{StatID: 1, Score: 5},
			wantErr:      false,
		},
		{
			name:         "NewStatCreated",
			existingStat: nil,
			userFound:    true,
			chatFound:    true,
			wantErr:      false,
		},
		{
			name:         "UserNotFound",
			existingStat: nil,
			userFound:    false,
			chatFound:    true,
			wantErr:      true,
		},
		{
			name:         "ChatNotFound",
			existingStat: nil,
			userFound:    true,
			chatFound:    false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRepo := &MockChatRepository{}
			userRepo := &MockUserRepository{}
			statRepo := &MockStatRepository{}
			svc := NewService(chatRepo, userRepo, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, &MockSubredditRepository{}, statRepo)

			statRepo.FindByUserChatYearFunc = func(ctx context.Context, userID, chatID int64, year int) (*model.Stat, error) {
				return tt.existingStat, nil
			}

			userRepo.GetFunc = func(ctx context.Context, userID int64) (*model.User, error) {
				if tt.userFound {
					return &model.User{UserID: userID, Username: "testuser"}, nil
				}
				return nil, nil
			}

			chatRepo.GetFunc = func(ctx context.Context, chatID int64) (*model.Chat, error) {
				if tt.chatFound {
					return &model.Chat{ChatID: chatID}, nil
				}
				return nil, nil
			}

			statRepo.SaveFunc = func(ctx context.Context, s *model.Stat) error {
				return nil
			}

			got, err := svc.GetOrCreateStat(context.Background(), 1, 1, 2025)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrCreateStat() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got == nil {
				t.Error("expected stat, got nil")
			}
		})
	}
}

func TestServiceGetTodayWinner(t *testing.T) {
	statRepo := &MockStatRepository{}
	svc := NewService(&MockChatRepository{}, &MockUserRepository{}, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, &MockSubredditRepository{}, statRepo)

	expected := &model.Stat{StatID: 1, IsWinner: true, User: &model.User{Username: "winner"}}
	statRepo.FindWinnerByChatFunc = func(ctx context.Context, chatID int64, year int) (*model.Stat, error) {
		return expected, nil
	}

	got, err := svc.GetTodayWinner(context.Background(), 1, 2025)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got == nil || !got.IsWinner {
		t.Error("expected winner stat")
	}
}

func TestServiceSelectRandomWinner(t *testing.T) {
	tests := []struct {
		name        string
		userFound   bool
		wantNil     bool
		wantErr     bool
		expectScore int64
	}{
		{
			name:        "WinnerSelected",
			userFound:   true,
			wantNil:     false,
			expectScore: 1,
		},
		{
			name:      "NoUsersInChat",
			userFound: false,
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRepo := &MockChatRepository{}
			userRepo := &MockUserRepository{}
			statRepo := &MockStatRepository{}
			svc := NewService(chatRepo, userRepo, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, &MockSubredditRepository{}, statRepo)

			userRepo.GetRandomByChatFunc = func(ctx context.Context, chatID int64) (*model.User, error) {
				if tt.userFound {
					return &model.User{UserID: 1, Username: "lucky"}, nil
				}
				return nil, nil
			}

			userRepo.GetFunc = func(ctx context.Context, userID int64) (*model.User, error) {
				return &model.User{UserID: userID, Username: "lucky"}, nil
			}

			chatRepo.GetFunc = func(ctx context.Context, chatID int64) (*model.Chat, error) {
				return &model.Chat{ChatID: chatID}, nil
			}

			statRepo.FindByUserChatYearFunc = func(ctx context.Context, userID, chatID int64, year int) (*model.Stat, error) {
				return nil, nil
			}

			var savedStat *model.Stat
			statRepo.SaveFunc = func(ctx context.Context, s *model.Stat) error {
				savedStat = s
				return nil
			}

			got, err := svc.SelectRandomWinner(context.Background(), 1, 2025)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectRandomWinner() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantNil && got != nil {
				t.Error("expected nil, got stat")
			}

			if !tt.wantNil && got == nil {
				t.Error("expected stat, got nil")
			}

			if !tt.wantNil && savedStat != nil {
				if !savedStat.IsWinner {
					t.Error("expected IsWinner to be true")
				}
				if savedStat.Score != tt.expectScore {
					t.Errorf("want score %d, got %d", tt.expectScore, savedStat.Score)
				}
			}
		})
	}
}

func TestServiceGetStatsByYear(t *testing.T) {
	statRepo := &MockStatRepository{}
	svc := NewService(&MockChatRepository{}, &MockUserRepository{}, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, &MockSubredditRepository{}, statRepo)

	expected := []*model.Stat{
		{StatID: 1, Score: 10, Year: 2025},
		{StatID: 2, Score: 5, Year: 2025},
	}

	statRepo.ListByChatAndYearFunc = func(ctx context.Context, chatID int64, year int) ([]*model.Stat, error) {
		if year != 2025 {
			t.Errorf("want year 2025, got %d", year)
		}
		return expected, nil
	}

	got, err := svc.GetStatsByYear(context.Background(), 1, 2025)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(got) != len(expected) {
		t.Errorf("want %d stats, got %d", len(expected), len(got))
	}
}

func TestServiceGetAllStats(t *testing.T) {
	statRepo := &MockStatRepository{}
	svc := NewService(&MockChatRepository{}, &MockUserRepository{}, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, &MockSubredditRepository{}, statRepo)

	expected := []*model.Stat{
		{StatID: 1, Score: 10, Year: 2024},
		{StatID: 2, Score: 5, Year: 2025},
	}

	statRepo.ListByChatFunc = func(ctx context.Context, chatID int64) ([]*model.Stat, error) {
		return expected, nil
	}

	got, err := svc.GetAllStats(context.Background(), 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(got) != len(expected) {
		t.Errorf("want %d stats, got %d", len(expected), len(got))
	}
}

func TestServiceResetDailyWinners(t *testing.T) {
	statRepo := &MockStatRepository{}
	svc := NewService(&MockChatRepository{}, &MockUserRepository{}, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, &MockSubredditRepository{}, statRepo)

	resetCalled := false
	statRepo.ResetDailyWinnersFunc = func(ctx context.Context) error {
		resetCalled = true
		return nil
	}

	if err := svc.ResetDailyWinners(context.Background()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !resetCalled {
		t.Error("ResetDailyWinners was not called")
	}
}

func TestServiceGetRandomFact(t *testing.T) {
	factRepo := &MockFactRepository{}
	svc := NewService(&MockChatRepository{}, &MockUserRepository{}, &MockReminderRepository{}, factRepo, &MockStickerRepository{}, &MockSubredditRepository{}, &MockStatRepository{})

	expected := &model.Fact{ID: 1, Comment: "interesting"}
	factRepo.GetRandomByChatFunc = func(ctx context.Context, chatID int64) (*model.Fact, error) {
		return expected, nil
	}

	got, err := svc.GetRandomFact(context.Background(), 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got.Comment != expected.Comment {
		t.Errorf("want comment %s, got %s", expected.Comment, got.Comment)
	}
}

func TestServiceListFacts(t *testing.T) {
	factRepo := &MockFactRepository{}
	svc := NewService(&MockChatRepository{}, &MockUserRepository{}, &MockReminderRepository{}, factRepo, &MockStickerRepository{}, &MockSubredditRepository{}, &MockStatRepository{})

	expected := []*model.Fact{{Comment: "fact1"}, {Comment: "fact2"}}
	factRepo.ListByChatFunc = func(ctx context.Context, chatID int64) ([]*model.Fact, error) {
		return expected, nil
	}

	got, err := svc.ListFacts(context.Background(), 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(got) != len(expected) {
		t.Errorf("want %d facts, got %d", len(expected), len(got))
	}
}

func TestServiceGetPendingReminders(t *testing.T) {
	reminderRepo := &MockReminderRepository{}
	svc := NewService(&MockChatRepository{}, &MockUserRepository{}, reminderRepo, &MockFactRepository{}, &MockStickerRepository{}, &MockSubredditRepository{}, &MockStatRepository{})

	expected := []*model.Reminder{{ReminderID: 1}, {ReminderID: 2}}
	reminderRepo.ListByChatFunc = func(ctx context.Context, chatID int64) ([]*model.Reminder, error) {
		return expected, nil
	}

	got, err := svc.GetPendingReminders(context.Background(), 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(got) != len(expected) {
		t.Errorf("want %d reminders, got %d", len(expected), len(got))
	}
}

func TestServiceRunAutoRoulette(t *testing.T) {
	tests := []struct {
		name            string
		chats           []*model.Chat
		existingWinners map[int64]bool
		usersInChat     map[int64]*model.User
		wantResults     int
		wantErr         bool
	}{
		{
			name:        "NoChats",
			chats:       []*model.Chat{},
			wantResults: 0,
		},
		{
			name:            "AllChatsHaveWinners",
			chats:           []*model.Chat{{ChatID: 1}, {ChatID: 2}},
			existingWinners: map[int64]bool{1: true, 2: true},
			wantResults:     0,
		},
		{
			name:            "OneNewWinner",
			chats:           []*model.Chat{{ChatID: 1}, {ChatID: 2}},
			existingWinners: map[int64]bool{1: true},
			usersInChat:     map[int64]*model.User{2: {UserID: 100, Username: "winner"}},
			wantResults:     1,
		},
		{
			name:            "NoUsersInChat",
			chats:           []*model.Chat{{ChatID: 1}},
			existingWinners: map[int64]bool{},
			usersInChat:     map[int64]*model.User{},
			wantResults:     0,
		},
		{
			name:            "MultipleNewWinners",
			chats:           []*model.Chat{{ChatID: 1}, {ChatID: 2}, {ChatID: 3}},
			existingWinners: map[int64]bool{},
			usersInChat: map[int64]*model.User{
				1: {UserID: 100, Username: "user1"},
				2: {UserID: 200, Username: "user2"},
				3: {UserID: 300, Username: "user3"},
			},
			wantResults: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRepo := &MockChatRepository{}
			userRepo := &MockUserRepository{}
			statRepo := &MockStatRepository{}
			svc := NewService(chatRepo, userRepo, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, &MockSubredditRepository{}, statRepo)

			chatRepo.ListAllFunc = func(ctx context.Context) ([]*model.Chat, error) {
				return tt.chats, nil
			}

			chatRepo.GetFunc = func(ctx context.Context, chatID int64) (*model.Chat, error) {
				return &model.Chat{ChatID: chatID}, nil
			}

			statRepo.FindWinnerByChatFunc = func(ctx context.Context, chatID int64, year int) (*model.Stat, error) {
				if tt.existingWinners[chatID] {
					return &model.Stat{IsWinner: true}, nil
				}
				return nil, nil
			}

			userRepo.GetRandomByChatFunc = func(ctx context.Context, chatID int64) (*model.User, error) {
				if user, ok := tt.usersInChat[chatID]; ok {
					return user, nil
				}
				return nil, nil
			}

			userRepo.GetFunc = func(ctx context.Context, userID int64) (*model.User, error) {
				for _, user := range tt.usersInChat {
					if user.UserID == userID {
						return user, nil
					}
				}
				return nil, nil
			}

			statRepo.FindByUserChatYearFunc = func(ctx context.Context, userID, chatID int64, year int) (*model.Stat, error) {
				return nil, nil
			}

			statRepo.SaveFunc = func(ctx context.Context, s *model.Stat) error {
				return nil
			}

			results, err := svc.RunAutoRoulette(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("RunAutoRoulette() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(results) != tt.wantResults {
				t.Errorf("want %d results, got %d", tt.wantResults, len(results))
			}
		})
	}
}

func TestServiceRunAutoRoulette_ListAllError(t *testing.T) {
	chatRepo := &MockChatRepository{}
	svc := NewService(chatRepo, &MockUserRepository{}, &MockReminderRepository{}, &MockFactRepository{}, &MockStickerRepository{}, &MockSubredditRepository{}, &MockStatRepository{})

	chatRepo.ListAllFunc = func(ctx context.Context) ([]*model.Chat, error) {
		return nil, errMock
	}

	_, err := svc.RunAutoRoulette(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}
