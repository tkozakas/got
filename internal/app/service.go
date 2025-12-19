package app

import (
	"context"
	"fmt"
	"got/internal/app/model"
	"time"
)

type Service struct {
	chats      ChatRepository
	users      UserRepository
	reminders  ReminderRepository
	facts      FactRepository
	stickers   StickerRepository
	subreddits SubredditRepository
}

func NewService(
	chats ChatRepository,
	users UserRepository,
	reminders ReminderRepository,
	facts FactRepository,
	stickers StickerRepository,
	subreddits SubredditRepository,
) *Service {
	return &Service{
		chats:      chats,
		users:      users,
		reminders:  reminders,
		facts:      facts,
		stickers:   stickers,
		subreddits: subreddits,
	}
}

func (s *Service) RegisterChat(ctx context.Context, chat *model.Chat) error {
	return s.chats.Save(ctx, chat)
}

func (s *Service) RegisterUser(ctx context.Context, user *model.User) error {
	return s.users.Save(ctx, user)
}

func (s *Service) AddFact(ctx context.Context, text string, chatID int64) error {
	chat, err := s.chats.Get(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	fact := &model.Fact{
		Comment: text,
		Chat:    chat,
	}
	return s.facts.Save(ctx, fact)
}

func (s *Service) GetRandomFact(ctx context.Context) (*model.Fact, error) {
	return s.facts.GetRandom(ctx)
}

func (s *Service) GetRandomSticker(ctx context.Context) (*model.Sticker, error) {
	return s.stickers.GetRandom(ctx)
}

func (s *Service) GetRandomSubreddit(ctx context.Context) (*model.Subreddit, error) {
	return s.subreddits.GetRandom(ctx)
}

func (s *Service) AddReminder(
	ctx context.Context,
	chatID int64,
	userID int64,
	message string,
	duration time.Duration,
) error {
	chat, err := s.chats.Get(ctx, chatID)
	if err != nil {
		return fmt.Errorf("chat not found: %w", err)
	}
	if chat == nil {
		return fmt.Errorf("chat not registered")
	}

	user, err := s.users.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not registered")
	}

	reminder := &model.Reminder{
		Chat:      chat,
		User:      user,
		Message:   message,
		RemindAt:  time.Now().Add(duration),
		CreatedAt: time.Now(),
		Sent:      false,
	}

	return s.reminders.Save(ctx, reminder)
}

func (s *Service) GetPendingReminders(ctx context.Context, chatID int64) ([]*model.Reminder, error) {
	return s.reminders.ListByChat(ctx, chatID)
}

func (s *Service) CheckReminders(ctx context.Context) ([]*model.Reminder, error) {
	pending, err := s.reminders.ListPending(ctx)
	if err != nil {
		return nil, err
	}

	for _, r := range pending {
		if err := s.reminders.MarkSent(ctx, r.ReminderID); err != nil {
			fmt.Printf("failed to mark reminder %d as sent: %v\n", r.ReminderID, err)
		}
	}

	return pending, nil
}
