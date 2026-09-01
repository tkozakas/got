package app

import (
	"context"
	"got/internal/app/model"
)

type ChatRepository interface {
	Save(ctx context.Context, chat *model.Chat) error
	Get(ctx context.Context, chatID int64) (*model.Chat, error)
	ListAll(ctx context.Context) ([]*model.Chat, error)
	SetLanguage(ctx context.Context, chatID int64, language string) error
	GetLanguage(ctx context.Context, chatID int64) (string, error)
}

type UserRepository interface {
	Save(ctx context.Context, user *model.User) error
	Get(ctx context.Context, userID int64) (*model.User, error)
	AddToChat(ctx context.Context, userID, chatID int64) error
	GetRandomByChat(ctx context.Context, chatID int64) (*model.User, error)
}

type ReminderRepository interface {
	Save(ctx context.Context, reminder *model.Reminder) error
	ListPending(ctx context.Context) ([]*model.Reminder, error)
	MarkSent(ctx context.Context, reminderID int64) error
	ListByChat(ctx context.Context, chatID int64) ([]*model.Reminder, error)
	Delete(ctx context.Context, reminderID int64, chatID int64) error
}

type FactRepository interface {
	Save(ctx context.Context, fact *model.Fact) error
	GetRandomByChat(ctx context.Context, chatID int64) (*model.Fact, error)
	ListByChat(ctx context.Context, chatID int64) ([]*model.Fact, error)
}

type StickerRepository interface {
	Save(ctx context.Context, sticker *model.Sticker) error
	GetRandomByChat(ctx context.Context, chatID int64) (*model.Sticker, error)
	ListByChat(ctx context.Context, chatID int64) ([]*model.Sticker, error)
	Delete(ctx context.Context, fileID string, chatID int64) error
	DeleteBySetName(ctx context.Context, setName string, chatID int64) (int, error)
}

type SubredditRepository interface {
	Save(ctx context.Context, subreddit *model.Subreddit) error
	GetRandomByChat(ctx context.Context, chatID int64) (*model.Subreddit, error)
	ListByChat(ctx context.Context, chatID int64) ([]*model.Subreddit, error)
	Delete(ctx context.Context, name string, chatID int64) error
}

type StatRepository interface {
	Save(ctx context.Context, stat *model.Stat) error
	FindByUserChatYear(ctx context.Context, userID, chatID int64, year int) (*model.Stat, error)
	FindWinnerByChat(ctx context.Context, chatID int64, year int) (*model.Stat, error)
	ListByChatAndYear(ctx context.Context, chatID int64, year int) ([]*model.Stat, error)
	ListByChat(ctx context.Context, chatID int64) ([]*model.Stat, error)
	ResetDailyWinners(ctx context.Context) error
	ResetWinnerByChat(ctx context.Context, chatID int64, year int) error
	Update(ctx context.Context, statID int64, score int64, isWinner bool) error
}
