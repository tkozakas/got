package app

import (
	"context"
	"errors"
	"got/internal/app/model"
)

type MockChatRepository struct {
	SaveFunc        func(ctx context.Context, chat *model.Chat) error
	GetFunc         func(ctx context.Context, chatID int64) (*model.Chat, error)
	ListAllFunc     func(ctx context.Context) ([]*model.Chat, error)
	SetLanguageFunc func(ctx context.Context, chatID int64, language string) error
	GetLanguageFunc func(ctx context.Context, chatID int64) (string, error)
}

func (m *MockChatRepository) Save(ctx context.Context, chat *model.Chat) error {
	return m.SaveFunc(ctx, chat)
}
func (m *MockChatRepository) Get(ctx context.Context, chatID int64) (*model.Chat, error) {
	return m.GetFunc(ctx, chatID)
}
func (m *MockChatRepository) ListAll(ctx context.Context) ([]*model.Chat, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc(ctx)
	}
	return nil, nil
}
func (m *MockChatRepository) SetLanguage(ctx context.Context, chatID int64, language string) error {
	if m.SetLanguageFunc != nil {
		return m.SetLanguageFunc(ctx, chatID, language)
	}
	return nil
}
func (m *MockChatRepository) GetLanguage(ctx context.Context, chatID int64) (string, error) {
	if m.GetLanguageFunc != nil {
		return m.GetLanguageFunc(ctx, chatID)
	}
	return "", nil
}

type MockUserRepository struct {
	SaveFunc            func(ctx context.Context, user *model.User) error
	GetFunc             func(ctx context.Context, userID int64) (*model.User, error)
	AddToChatFunc       func(ctx context.Context, userID, chatID int64) error
	GetRandomByChatFunc func(ctx context.Context, chatID int64) (*model.User, error)
}

func (m *MockUserRepository) Save(ctx context.Context, user *model.User) error {
	return m.SaveFunc(ctx, user)
}
func (m *MockUserRepository) Get(ctx context.Context, userID int64) (*model.User, error) {
	return m.GetFunc(ctx, userID)
}
func (m *MockUserRepository) AddToChat(ctx context.Context, userID, chatID int64) error {
	if m.AddToChatFunc != nil {
		return m.AddToChatFunc(ctx, userID, chatID)
	}
	return nil
}
func (m *MockUserRepository) GetRandomByChat(ctx context.Context, chatID int64) (*model.User, error) {
	if m.GetRandomByChatFunc != nil {
		return m.GetRandomByChatFunc(ctx, chatID)
	}
	return nil, nil
}

type MockReminderRepository struct {
	SaveFunc        func(ctx context.Context, reminder *model.Reminder) error
	ListPendingFunc func(ctx context.Context) ([]*model.Reminder, error)
	MarkSentFunc    func(ctx context.Context, reminderID int64) error
	ListByChatFunc  func(ctx context.Context, chatID int64) ([]*model.Reminder, error)
	DeleteFunc      func(ctx context.Context, reminderID int64, chatID int64) error
}

func (m *MockReminderRepository) Save(ctx context.Context, reminder *model.Reminder) error {
	return m.SaveFunc(ctx, reminder)
}
func (m *MockReminderRepository) ListPending(ctx context.Context) ([]*model.Reminder, error) {
	return m.ListPendingFunc(ctx)
}
func (m *MockReminderRepository) MarkSent(ctx context.Context, reminderID int64) error {
	return m.MarkSentFunc(ctx, reminderID)
}
func (m *MockReminderRepository) ListByChat(ctx context.Context, chatID int64) ([]*model.Reminder, error) {
	return m.ListByChatFunc(ctx, chatID)
}
func (m *MockReminderRepository) Delete(ctx context.Context, reminderID int64, chatID int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, reminderID, chatID)
	}
	return nil
}

type MockFactRepository struct {
	SaveFunc            func(ctx context.Context, fact *model.Fact) error
	GetRandomByChatFunc func(ctx context.Context, chatID int64) (*model.Fact, error)
	ListByChatFunc      func(ctx context.Context, chatID int64) ([]*model.Fact, error)
}

func (m *MockFactRepository) Save(ctx context.Context, fact *model.Fact) error {
	return m.SaveFunc(ctx, fact)
}
func (m *MockFactRepository) GetRandomByChat(ctx context.Context, chatID int64) (*model.Fact, error) {
	if m.GetRandomByChatFunc != nil {
		return m.GetRandomByChatFunc(ctx, chatID)
	}
	return nil, nil
}
func (m *MockFactRepository) ListByChat(ctx context.Context, chatID int64) ([]*model.Fact, error) {
	if m.ListByChatFunc != nil {
		return m.ListByChatFunc(ctx, chatID)
	}
	return nil, nil
}

type MockStickerRepository struct {
	SaveFunc            func(ctx context.Context, sticker *model.Sticker) error
	GetRandomByChatFunc func(ctx context.Context, chatID int64) (*model.Sticker, error)
	ListByChatFunc      func(ctx context.Context, chatID int64) ([]*model.Sticker, error)
	DeleteFunc          func(ctx context.Context, fileID string, chatID int64) error
	DeleteBySetNameFunc func(ctx context.Context, setName string, chatID int64) (int, error)
}

func (m *MockStickerRepository) Save(ctx context.Context, sticker *model.Sticker) error {
	return m.SaveFunc(ctx, sticker)
}
func (m *MockStickerRepository) GetRandomByChat(ctx context.Context, chatID int64) (*model.Sticker, error) {
	if m.GetRandomByChatFunc != nil {
		return m.GetRandomByChatFunc(ctx, chatID)
	}
	return nil, nil
}
func (m *MockStickerRepository) ListByChat(ctx context.Context, chatID int64) ([]*model.Sticker, error) {
	if m.ListByChatFunc != nil {
		return m.ListByChatFunc(ctx, chatID)
	}
	return nil, nil
}
func (m *MockStickerRepository) Delete(ctx context.Context, fileID string, chatID int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, fileID, chatID)
	}
	return nil
}

func (m *MockStickerRepository) DeleteBySetName(ctx context.Context, setName string, chatID int64) (int, error) {
	if m.DeleteBySetNameFunc != nil {
		return m.DeleteBySetNameFunc(ctx, setName, chatID)
	}
	return 0, nil
}

type MockSubredditRepository struct {
	SaveFunc            func(ctx context.Context, subreddit *model.Subreddit) error
	GetRandomByChatFunc func(ctx context.Context, chatID int64) (*model.Subreddit, error)
	ListByChatFunc      func(ctx context.Context, chatID int64) ([]*model.Subreddit, error)
	DeleteFunc          func(ctx context.Context, name string, chatID int64) error
}

func (m *MockSubredditRepository) Save(ctx context.Context, subreddit *model.Subreddit) error {
	return m.SaveFunc(ctx, subreddit)
}
func (m *MockSubredditRepository) GetRandomByChat(ctx context.Context, chatID int64) (*model.Subreddit, error) {
	if m.GetRandomByChatFunc != nil {
		return m.GetRandomByChatFunc(ctx, chatID)
	}
	return nil, nil
}
func (m *MockSubredditRepository) ListByChat(ctx context.Context, chatID int64) ([]*model.Subreddit, error) {
	if m.ListByChatFunc != nil {
		return m.ListByChatFunc(ctx, chatID)
	}
	return nil, nil
}
func (m *MockSubredditRepository) Delete(ctx context.Context, name string, chatID int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, name, chatID)
	}
	return nil
}

type MockStatRepository struct {
	SaveFunc               func(ctx context.Context, stat *model.Stat) error
	FindByUserChatYearFunc func(ctx context.Context, userID, chatID int64, year int) (*model.Stat, error)
	FindWinnerByChatFunc   func(ctx context.Context, chatID int64, year int) (*model.Stat, error)
	ListByChatAndYearFunc  func(ctx context.Context, chatID int64, year int) ([]*model.Stat, error)
	ListByChatFunc         func(ctx context.Context, chatID int64) ([]*model.Stat, error)
	ResetDailyWinnersFunc  func(ctx context.Context) error
	ResetWinnerByChatFunc  func(ctx context.Context, chatID int64, year int) error
	UpdateFunc             func(ctx context.Context, statID int64, score int64, isWinner bool) error
}

func (m *MockStatRepository) Save(ctx context.Context, stat *model.Stat) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, stat)
	}
	return nil
}
func (m *MockStatRepository) FindByUserChatYear(ctx context.Context, userID, chatID int64, year int) (*model.Stat, error) {
	if m.FindByUserChatYearFunc != nil {
		return m.FindByUserChatYearFunc(ctx, userID, chatID, year)
	}
	return nil, nil
}
func (m *MockStatRepository) FindWinnerByChat(ctx context.Context, chatID int64, year int) (*model.Stat, error) {
	if m.FindWinnerByChatFunc != nil {
		return m.FindWinnerByChatFunc(ctx, chatID, year)
	}
	return nil, nil
}
func (m *MockStatRepository) ListByChatAndYear(ctx context.Context, chatID int64, year int) ([]*model.Stat, error) {
	if m.ListByChatAndYearFunc != nil {
		return m.ListByChatAndYearFunc(ctx, chatID, year)
	}
	return nil, nil
}
func (m *MockStatRepository) ListByChat(ctx context.Context, chatID int64) ([]*model.Stat, error) {
	if m.ListByChatFunc != nil {
		return m.ListByChatFunc(ctx, chatID)
	}
	return nil, nil
}
func (m *MockStatRepository) ResetDailyWinners(ctx context.Context) error {
	if m.ResetDailyWinnersFunc != nil {
		return m.ResetDailyWinnersFunc(ctx)
	}
	return nil
}

func (m *MockStatRepository) ResetWinnerByChat(ctx context.Context, chatID int64, year int) error {
	if m.ResetWinnerByChatFunc != nil {
		return m.ResetWinnerByChatFunc(ctx, chatID, year)
	}
	return nil
}

func (m *MockStatRepository) Update(ctx context.Context, statID int64, score int64, isWinner bool) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, statID, score, isWinner)
	}
	return nil
}

var errMock = errors.New("mock error")
