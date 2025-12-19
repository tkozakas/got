package telegram

import (
	"context"
	"errors"
	"got/internal/app"
	"got/internal/app/model"
	"testing"
)

type mockChatRepo struct {
	saveFunc func(ctx context.Context, chat *model.Chat) error
	getFunc  func(ctx context.Context, chatID int64) (*model.Chat, error)
}

func (m *mockChatRepo) Save(ctx context.Context, chat *model.Chat) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, chat)
	}
	return nil
}

func (m *mockChatRepo) Get(ctx context.Context, chatID int64) (*model.Chat, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, chatID)
	}
	return nil, nil
}

func (m *mockChatRepo) ListAll(ctx context.Context) ([]*model.Chat, error) {
	return nil, nil
}

type mockUserRepo struct {
	saveFunc      func(ctx context.Context, user *model.User) error
	addToChatFunc func(ctx context.Context, userID, chatID int64) error
	getFunc       func(ctx context.Context, userID int64) (*model.User, error)
}

func (m *mockUserRepo) Save(ctx context.Context, user *model.User) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) Get(ctx context.Context, userID int64) (*model.User, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockUserRepo) AddToChat(ctx context.Context, userID, chatID int64) error {
	if m.addToChatFunc != nil {
		return m.addToChatFunc(ctx, userID, chatID)
	}
	return nil
}

func (m *mockUserRepo) GetRandomByChat(ctx context.Context, chatID int64) (*model.User, error) {
	return nil, nil
}

type mockReminderRepo struct {
	saveFunc       func(ctx context.Context, r *model.Reminder) error
	listByChatFunc func(ctx context.Context, chatID int64) ([]*model.Reminder, error)
	deleteFunc     func(ctx context.Context, reminderID int64, chatID int64) error
}

func (m *mockReminderRepo) Save(ctx context.Context, r *model.Reminder) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, r)
	}
	return nil
}
func (m *mockReminderRepo) ListPending(ctx context.Context) ([]*model.Reminder, error) {
	return nil, nil
}
func (m *mockReminderRepo) MarkSent(ctx context.Context, id int64) error { return nil }
func (m *mockReminderRepo) ListByChat(ctx context.Context, chatID int64) ([]*model.Reminder, error) {
	if m.listByChatFunc != nil {
		return m.listByChatFunc(ctx, chatID)
	}
	return nil, nil
}
func (m *mockReminderRepo) Delete(ctx context.Context, reminderID int64, chatID int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, reminderID, chatID)
	}
	return nil
}

type mockFactRepo struct {
	saveFunc            func(ctx context.Context, f *model.Fact) error
	getRandomByChatFunc func(ctx context.Context, chatID int64) (*model.Fact, error)
	listByChatFunc      func(ctx context.Context, chatID int64) ([]*model.Fact, error)
}

func (m *mockFactRepo) Save(ctx context.Context, f *model.Fact) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, f)
	}
	return nil
}
func (m *mockFactRepo) GetRandomByChat(ctx context.Context, chatID int64) (*model.Fact, error) {
	if m.getRandomByChatFunc != nil {
		return m.getRandomByChatFunc(ctx, chatID)
	}
	return nil, nil
}
func (m *mockFactRepo) ListByChat(ctx context.Context, chatID int64) ([]*model.Fact, error) {
	if m.listByChatFunc != nil {
		return m.listByChatFunc(ctx, chatID)
	}
	return nil, nil
}

type mockStickerRepo struct {
	saveFunc            func(ctx context.Context, s *model.Sticker) error
	getRandomByChatFunc func(ctx context.Context, chatID int64) (*model.Sticker, error)
	listByChatFunc      func(ctx context.Context, chatID int64) ([]*model.Sticker, error)
	deleteFunc          func(ctx context.Context, fileID string, chatID int64) error
	deleteBySetNameFunc func(ctx context.Context, setName string, chatID int64) (int, error)
}

func (m *mockStickerRepo) Save(ctx context.Context, s *model.Sticker) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, s)
	}
	return nil
}
func (m *mockStickerRepo) GetRandomByChat(ctx context.Context, chatID int64) (*model.Sticker, error) {
	if m.getRandomByChatFunc != nil {
		return m.getRandomByChatFunc(ctx, chatID)
	}
	return nil, nil
}
func (m *mockStickerRepo) ListByChat(ctx context.Context, chatID int64) ([]*model.Sticker, error) {
	if m.listByChatFunc != nil {
		return m.listByChatFunc(ctx, chatID)
	}
	return nil, nil
}
func (m *mockStickerRepo) Delete(ctx context.Context, fileID string, chatID int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, fileID, chatID)
	}
	return nil
}

func (m *mockStickerRepo) DeleteBySetName(ctx context.Context, setName string, chatID int64) (int, error) {
	if m.deleteBySetNameFunc != nil {
		return m.deleteBySetNameFunc(ctx, setName, chatID)
	}
	return 0, nil
}

type mockSubredditRepo struct {
	saveFunc            func(ctx context.Context, s *model.Subreddit) error
	getRandomByChatFunc func(ctx context.Context, chatID int64) (*model.Subreddit, error)
	listByChatFunc      func(ctx context.Context, chatID int64) ([]*model.Subreddit, error)
	deleteFunc          func(ctx context.Context, name string, chatID int64) error
}

func (m *mockSubredditRepo) Save(ctx context.Context, s *model.Subreddit) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, s)
	}
	return nil
}
func (m *mockSubredditRepo) GetRandomByChat(ctx context.Context, chatID int64) (*model.Subreddit, error) {
	if m.getRandomByChatFunc != nil {
		return m.getRandomByChatFunc(ctx, chatID)
	}
	return nil, nil
}
func (m *mockSubredditRepo) ListByChat(ctx context.Context, chatID int64) ([]*model.Subreddit, error) {
	if m.listByChatFunc != nil {
		return m.listByChatFunc(ctx, chatID)
	}
	return nil, nil
}
func (m *mockSubredditRepo) Delete(ctx context.Context, name string, chatID int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, name, chatID)
	}
	return nil
}

type mockStatRepo struct {
	saveFunc               func(ctx context.Context, s *model.Stat) error
	findByUserChatYearFunc func(ctx context.Context, userID, chatID int64, year int) (*model.Stat, error)
	findWinnerByChatFunc   func(ctx context.Context, chatID int64, year int) (*model.Stat, error)
	listByChatAndYearFunc  func(ctx context.Context, chatID int64, year int) ([]*model.Stat, error)
	listByChatFunc         func(ctx context.Context, chatID int64) ([]*model.Stat, error)
}

func (m *mockStatRepo) Save(ctx context.Context, s *model.Stat) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, s)
	}
	return nil
}
func (m *mockStatRepo) FindByUserChatYear(ctx context.Context, userID, chatID int64, year int) (*model.Stat, error) {
	if m.findByUserChatYearFunc != nil {
		return m.findByUserChatYearFunc(ctx, userID, chatID, year)
	}
	return nil, nil
}
func (m *mockStatRepo) FindWinnerByChat(ctx context.Context, chatID int64, year int) (*model.Stat, error) {
	if m.findWinnerByChatFunc != nil {
		return m.findWinnerByChatFunc(ctx, chatID, year)
	}
	return nil, nil
}
func (m *mockStatRepo) ListByChatAndYear(ctx context.Context, chatID int64, year int) ([]*model.Stat, error) {
	if m.listByChatAndYearFunc != nil {
		return m.listByChatAndYearFunc(ctx, chatID, year)
	}
	return nil, nil
}
func (m *mockStatRepo) ListByChat(ctx context.Context, chatID int64) ([]*model.Stat, error) {
	if m.listByChatFunc != nil {
		return m.listByChatFunc(ctx, chatID)
	}
	return nil, nil
}
func (m *mockStatRepo) ResetDailyWinners(ctx context.Context) error { return nil }

type mockHandler struct {
	called bool
	err    error
}

func (m *mockHandler) Handle(ctx context.Context, update *Update) error {
	m.called = true
	return m.err
}

func newTestService(chatRepo *mockChatRepo, userRepo *mockUserRepo) *app.Service {
	return app.NewService(
		chatRepo,
		userRepo,
		&mockReminderRepo{},
		&mockFactRepo{},
		&mockStickerRepo{},
		&mockSubredditRepo{},
		&mockStatRepo{},
	)
}

func TestAutoRegisterMiddlewareHandle(t *testing.T) {
	tests := []struct {
		name           string
		update         *Update
		wantChatSaved  bool
		wantUserSaved  bool
		wantNextCalled bool
	}{
		{
			name:           "NilMessage",
			update:         &Update{Message: nil},
			wantChatSaved:  false,
			wantUserSaved:  false,
			wantNextCalled: true,
		},
		{
			name: "MessageWithChat",
			update: &Update{
				Message: &Message{
					Chat: &Chat{ID: 123, Title: "Test Chat"},
				},
			},
			wantChatSaved:  true,
			wantUserSaved:  false,
			wantNextCalled: true,
		},
		{
			name: "MessageWithChatAndUser",
			update: &Update{
				Message: &Message{
					Chat: &Chat{ID: 123, Title: "Test Chat"},
					From: &User{ID: 456, UserName: "testuser", IsBot: false},
				},
			},
			wantChatSaved:  true,
			wantUserSaved:  true,
			wantNextCalled: true,
		},
		{
			name: "BotUserNotRegistered",
			update: &Update{
				Message: &Message{
					Chat: &Chat{ID: 123},
					From: &User{ID: 456, IsBot: true},
				},
			},
			wantChatSaved:  true,
			wantUserSaved:  false,
			wantNextCalled: true,
		},
		{
			name: "UserWithoutUsername",
			update: &Update{
				Message: &Message{
					Chat: &Chat{ID: 123, Type: "private"},
					From: &User{ID: 456, FirstName: "John", IsBot: false},
				},
			},
			wantChatSaved:  true,
			wantUserSaved:  true,
			wantNextCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatSaved := false
			userSaved := false

			chatRepo := &mockChatRepo{
				saveFunc: func(ctx context.Context, chat *model.Chat) error {
					chatSaved = true
					return nil
				},
			}

			userRepo := &mockUserRepo{
				saveFunc: func(ctx context.Context, user *model.User) error {
					userSaved = true
					return nil
				},
			}

			svc := newTestService(chatRepo, userRepo)
			next := &mockHandler{}
			mw := NewAutoRegisterMiddleware(svc, next)

			_ = mw.Handle(context.Background(), tt.update)

			if chatSaved != tt.wantChatSaved {
				t.Errorf("chat saved = %v, want %v", chatSaved, tt.wantChatSaved)
			}

			if userSaved != tt.wantUserSaved {
				t.Errorf("user saved = %v, want %v", userSaved, tt.wantUserSaved)
			}

			if next.called != tt.wantNextCalled {
				t.Errorf("next called = %v, want %v", next.called, tt.wantNextCalled)
			}
		})
	}
}

func TestAutoRegisterMiddlewarePropagatesNextError(t *testing.T) {
	svc := newTestService(&mockChatRepo{}, &mockUserRepo{})
	expectedErr := errors.New("next handler error")
	next := &mockHandler{err: expectedErr}
	mw := NewAutoRegisterMiddleware(svc, next)

	err := mw.Handle(context.Background(), &Update{})

	if err != expectedErr {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestWithLogging(t *testing.T) {
	called := false
	handler := WithLogging(func(ctx context.Context, update *Update) error {
		called = true
		return nil
	})

	update := &Update{
		Message: &Message{
			Text: "/test",
			From: &User{UserName: "testuser"},
		},
	}

	err := handler(context.Background(), update)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestWithRecover(t *testing.T) {
	handler := WithRecover(func(ctx context.Context, update *Update) error {
		panic("test panic")
	})

	err := handler(context.Background(), &Update{})

	if err != nil {
		t.Errorf("unexpected error after recover: %v", err)
	}
}
