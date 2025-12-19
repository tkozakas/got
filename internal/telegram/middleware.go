package telegram

import (
	"context"
	"log/slog"
)

type Middleware func(HandlerFunc) HandlerFunc

func WithLogging(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, update *Update) error {
		if update.Message != nil {
			slog.Info("User command received",
				"user", update.Message.From.UserName,
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
