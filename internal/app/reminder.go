package app

import (
	"context"
	"fmt"
	"got/internal/app/model"
	"log/slog"
	"time"
)

func (s *Service) AddReminder(ctx context.Context, chatID, userID int64, message string, duration time.Duration) error {
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

func (s *Service) DeleteReminder(ctx context.Context, reminderID int64, chatID int64) error {
	return s.reminders.Delete(ctx, reminderID, chatID)
}
