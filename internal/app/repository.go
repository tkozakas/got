package app

import (
	"context"
	"got/internal/app/model"
)

type ChatRepository interface {
	Save(ctx context.Context, chat *model.Chat) error
	Get(ctx context.Context, chatID int64) (*model.Chat, error)
}

type UserRepository interface {
	Save(ctx context.Context, user *model.User) error
	Get(ctx context.Context, userID int64) (*model.User, error)
}

type ReminderRepository interface {
	Save(ctx context.Context, reminder *model.Reminder) error
	ListPending(ctx context.Context) ([]*model.Reminder, error)
	MarkSent(ctx context.Context, reminderID int64) error
	ListByChat(ctx context.Context, chatID int64) ([]*model.Reminder, error)
}

type FactRepository interface {
	GetRandom(ctx context.Context) (*model.Fact, error)
	Save(ctx context.Context, fact *model.Fact) error
}

type StickerRepository interface {
	GetRandom(ctx context.Context) (*model.Sticker, error)
	Save(ctx context.Context, sticker *model.Sticker) error
}

type SubredditRepository interface {
	GetRandom(ctx context.Context) (*model.Subreddit, error)
	Save(ctx context.Context, subreddit *model.Subreddit) error
}
