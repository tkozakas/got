package app

import (
	"context"
	"fmt"
	"got/internal/app/model"
	"log/slog"
	"time"
)

type Service struct {
	chats      ChatRepository
	users      UserRepository
	reminders  ReminderRepository
	facts      FactRepository
	stickers   StickerRepository
	subreddits SubredditRepository
	stats      StatRepository
}

func NewService(
	chats ChatRepository,
	users UserRepository,
	reminders ReminderRepository,
	facts FactRepository,
	stickers StickerRepository,
	subreddits SubredditRepository,
	stats StatRepository,
) *Service {
	return &Service{
		chats:      chats,
		users:      users,
		reminders:  reminders,
		facts:      facts,
		stickers:   stickers,
		subreddits: subreddits,
		stats:      stats,
	}
}

func (s *Service) RegisterChat(ctx context.Context, chat *model.Chat) error {
	return s.chats.Save(ctx, chat)
}

func (s *Service) RegisterUser(ctx context.Context, user *model.User, chatID int64) error {
	if err := s.users.Save(ctx, user); err != nil {
		return err
	}
	return s.users.AddToChat(ctx, user.UserID, chatID)
}

// Fact operations
func (s *Service) AddFact(ctx context.Context, text string, chatID int64) error {
	chat, err := s.chats.Get(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}
	if chat == nil {
		return fmt.Errorf("chat not registered")
	}

	fact := &model.Fact{
		Comment: text,
		Chat:    chat,
	}
	return s.facts.Save(ctx, fact)
}

func (s *Service) GetRandomFact(ctx context.Context, chatID int64) (*model.Fact, error) {
	return s.facts.GetRandomByChat(ctx, chatID)
}

func (s *Service) ListFacts(ctx context.Context, chatID int64) ([]*model.Fact, error) {
	return s.facts.ListByChat(ctx, chatID)
}

// Sticker operations
func (s *Service) AddSticker(ctx context.Context, fileID string, chatID int64) error {
	chat, err := s.chats.Get(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}
	if chat == nil {
		return fmt.Errorf("chat not registered")
	}

	sticker := &model.Sticker{
		FileID: fileID,
		Chat:   chat,
	}
	return s.stickers.Save(ctx, sticker)
}

func (s *Service) GetRandomSticker(ctx context.Context, chatID int64) (*model.Sticker, error) {
	return s.stickers.GetRandomByChat(ctx, chatID)
}

func (s *Service) ListStickers(ctx context.Context, chatID int64) ([]*model.Sticker, error) {
	return s.stickers.ListByChat(ctx, chatID)
}

// Subreddit operations
func (s *Service) AddSubreddit(ctx context.Context, name string, chatID int64) error {
	sub := &model.Subreddit{
		Name:   name,
		ChatID: chatID,
	}
	return s.subreddits.Save(ctx, sub)
}

func (s *Service) GetRandomSubreddit(ctx context.Context, chatID int64) (*model.Subreddit, error) {
	return s.subreddits.GetRandomByChat(ctx, chatID)
}

func (s *Service) ListSubreddits(ctx context.Context, chatID int64) ([]*model.Subreddit, error) {
	return s.subreddits.ListByChat(ctx, chatID)
}

func (s *Service) RemoveSubreddit(ctx context.Context, name string, chatID int64) error {
	return s.subreddits.Delete(ctx, name, chatID)
}

// Reminder operations
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
			slog.Error("failed to mark reminder as sent", "id", r.ReminderID, "error", err)
		}
	}

	return pending, nil
}

// Stats operations
func (s *Service) GetOrCreateStat(ctx context.Context, userID, chatID int64, year int) (*model.Stat, error) {
	stat, err := s.stats.FindByUserChatYear(ctx, userID, chatID, year)
	if err != nil {
		return nil, err
	}
	if stat != nil {
		return stat, nil
	}

	user, err := s.users.Get(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	chat, err := s.chats.Get(ctx, chatID)
	if err != nil || chat == nil {
		return nil, fmt.Errorf("chat not found")
	}

	stat = &model.Stat{
		User:     user,
		Chat:     chat,
		Score:    0,
		Year:     year,
		IsWinner: false,
	}

	if err := s.stats.Save(ctx, stat); err != nil {
		return nil, err
	}

	return stat, nil
}

func (s *Service) GetTodayWinner(ctx context.Context, chatID int64, year int) (*model.Stat, error) {
	return s.stats.FindWinnerByChat(ctx, chatID, year)
}

func (s *Service) SelectRandomWinner(ctx context.Context, chatID int64, year int) (*model.Stat, error) {
	user, err := s.users.GetRandomByChat(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	stat, err := s.GetOrCreateStat(ctx, user.UserID, chatID, year)
	if err != nil {
		return nil, err
	}

	stat.Score++
	stat.IsWinner = true

	if err := s.stats.Save(ctx, stat); err != nil {
		return nil, err
	}

	return stat, nil
}

func (s *Service) GetStatsByYear(ctx context.Context, chatID int64, year int) ([]*model.Stat, error) {
	return s.stats.ListByChatAndYear(ctx, chatID, year)
}

func (s *Service) GetAllStats(ctx context.Context, chatID int64) ([]*model.Stat, error) {
	return s.stats.ListByChat(ctx, chatID)
}

func (s *Service) ResetDailyWinners(ctx context.Context) error {
	return s.stats.ResetDailyWinners(ctx)
}
