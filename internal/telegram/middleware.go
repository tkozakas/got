package telegram

import (
	"context"
	"got/internal/app"
	"got/internal/app/model"
	"log/slog"
)

type Middleware func(HandlerFunc) HandlerFunc

type AutoRegisterMiddleware struct {
	service *app.Service
	next    Handler
}

func WithLogging(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, update *Update) error {
		if update.Message != nil {
			username := ""
			if update.Message.From != nil {
				username = update.Message.From.UserName
				if username == "" {
					username = update.Message.From.FirstName
				}
			}
			slog.Info("User command received",
				"user", username,
				"command", update.Message.Command(),
			)
		}
		return next(ctx, update)
	}
}

func WithRecover(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, update *Update) error {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Panic recovered", "error", r)
			}
		}()
		return next(ctx, update)
	}
}

func NewAutoRegisterMiddleware(service *app.Service, next Handler) *AutoRegisterMiddleware {
	return &AutoRegisterMiddleware{
		service: service,
		next:    next,
	}
}

func (m *AutoRegisterMiddleware) Handle(ctx context.Context, update *Update) error {
	if update.Message != nil {
		m.registerChatAndUser(ctx, update.Message)
	}
	return m.next.Handle(ctx, update)
}

func (m *AutoRegisterMiddleware) registerChatAndUser(ctx context.Context, msg *Message) {
	if msg.Chat != nil {
		chat := &model.Chat{
			ChatID:   msg.Chat.ID,
			ChatName: m.getChatName(msg.Chat),
		}
		if err := m.service.RegisterChat(ctx, chat); err != nil {
			slog.Error("Failed to register chat", "chat_id", chat.ChatID, "error", err)
		}
	}

	if msg.From != nil && !msg.From.IsBot && msg.Chat != nil {
		user := &model.User{
			UserID:   msg.From.ID,
			Username: m.getUsername(msg.From),
		}
		if err := m.service.RegisterUser(ctx, user, msg.Chat.ID); err != nil {
			slog.Error("Failed to register user", "user_id", user.UserID, "error", err)
		}
	}
}

func (m *AutoRegisterMiddleware) getChatName(chat *Chat) string {
	if chat.Title != "" {
		return chat.Title
	}
	return chat.Type
}

func (m *AutoRegisterMiddleware) getUsername(user *User) string {
	if user.UserName != "" {
		return user.UserName
	}
	return user.FirstName
}
