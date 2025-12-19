package app

import (
	"context"
	"got/internal/app/model"
	"testing"
	"time"
)

func TestServiceRegisterChat(t *testing.T) {
	svc, chatRepo, _, _, _, _, _ := newTestService()

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
	svc, _, userRepo, _, _, _, _ := newTestService()

	user := &model.User{UserID: 1, Username: "test"}

	userRepo.SaveFunc = func(ctx context.Context, u *model.User) error {
		if u.UserID != user.UserID {
			t.Errorf("want user ID %d, got %d", user.UserID, u.UserID)
		}
		return nil
	}

	if err := svc.RegisterUser(context.Background(), user); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestServiceAddFact(t *testing.T) {
	svc, chatRepo, _, _, factRepo, _, _ := newTestService()

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
	svc, chatRepo, userRepo, reminderRepo, _, _, _ := newTestService()

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
	svc, _, _, reminderRepo, _, _, _ := newTestService()

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
