package app

import (
	"context"
	"errors"
	"got/internal/app/model"
)

// MockChatRepository
type MockChatRepository struct {
	SaveFunc func(ctx context.Context, chat *model.Chat) error
	GetFunc  func(ctx context.Context, chatID int64) (*model.Chat, error)
}

func (m *MockChatRepository) Save(ctx context.Context, chat *model.Chat) error {
	return m.SaveFunc(ctx, chat)
}
func (m *MockChatRepository) Get(ctx context.Context, chatID int64) (*model.Chat, error) {
	return m.GetFunc(ctx, chatID)
}

// MockUserRepository
type MockUserRepository struct {
	SaveFunc func(ctx context.Context, user *model.User) error
	GetFunc  func(ctx context.Context, userID int64) (*model.User, error)
}

func (m *MockUserRepository) Save(ctx context.Context, user *model.User) error {
	return m.SaveFunc(ctx, user)
}
func (m *MockUserRepository) Get(ctx context.Context, userID int64) (*model.User, error) {
	return m.GetFunc(ctx, userID)
}

// MockReminderRepository
type MockReminderRepository struct {
	SaveFunc        func(ctx context.Context, reminder *model.Reminder) error
	ListPendingFunc func(ctx context.Context) ([]*model.Reminder, error)
	MarkSentFunc    func(ctx context.Context, reminderID int64) error
	ListByChatFunc  func(ctx context.Context, chatID int64) ([]*model.Reminder, error)
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

// MockFactRepository
type MockFactRepository struct {
	GetRandomFunc func(ctx context.Context) (*model.Fact, error)
	SaveFunc      func(ctx context.Context, fact *model.Fact) error
}

func (m *MockFactRepository) GetRandom(ctx context.Context) (*model.Fact, error) {
	return m.GetRandomFunc(ctx)
}
func (m *MockFactRepository) Save(ctx context.Context, fact *model.Fact) error {
	return m.SaveFunc(ctx, fact)
}

// MockStickerRepository
type MockStickerRepository struct {
	GetRandomFunc func(ctx context.Context) (*model.Sticker, error)
	SaveFunc      func(ctx context.Context, sticker *model.Sticker) error
}

func (m *MockStickerRepository) GetRandom(ctx context.Context) (*model.Sticker, error) {
	return m.GetRandomFunc(ctx)
}
func (m *MockStickerRepository) Save(ctx context.Context, sticker *model.Sticker) error {
	return m.SaveFunc(ctx, sticker)
}

// MockSubredditRepository
type MockSubredditRepository struct {
	GetRandomFunc func(ctx context.Context) (*model.Subreddit, error)
	SaveFunc      func(ctx context.Context, subreddit *model.Subreddit) error
}

func (m *MockSubredditRepository) GetRandom(ctx context.Context) (*model.Subreddit, error) {
	return m.GetRandomFunc(ctx)
}
func (m *MockSubredditRepository) Save(ctx context.Context, subreddit *model.Subreddit) error {
	return m.SaveFunc(ctx, subreddit)
}

// Helper to create a service with all mocks
func newTestService() (*Service, *MockChatRepository, *MockUserRepository, *MockReminderRepository, *MockFactRepository, *MockStickerRepository, *MockSubredditRepository) {
	chatRepo := &MockChatRepository{}
	userRepo := &MockUserRepository{}
	reminderRepo := &MockReminderRepository{}
	factRepo := &MockFactRepository{}
	stickerRepo := &MockStickerRepository{}
	subredditRepo := &MockSubredditRepository{}

	svc := NewService(chatRepo, userRepo, reminderRepo, factRepo, stickerRepo, subredditRepo)
	return svc, chatRepo, userRepo, reminderRepo, factRepo, stickerRepo, subredditRepo
}

var errMock = errors.New("mock error")
