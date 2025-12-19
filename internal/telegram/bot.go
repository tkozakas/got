package telegram

import (
	"context"
	"log/slog"
	"time"
)

const (
	defaultUpdateInterval = 1 * time.Second
)

type Bot struct {
	client  *Client
	handler Handler
}

type Handler interface {
	Handle(ctx context.Context, update *Update) error
}

func NewBot(client *Client, handler Handler) *Bot {
	return &Bot{
		client:  client,
		handler: handler,
	}
}

func (b *Bot) Start(ctx context.Context) {
	offset := 0
	ticker := time.NewTicker(defaultUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			offset = b.pollUpdates(ctx, offset)
		}
	}
}

func (b *Bot) pollUpdates(ctx context.Context, offset int) int {
	updates, err := b.client.GetUpdates(offset)
	if err != nil {
		slog.Error("Failed to get updates", "error", err)
		return offset
	}

	for _, update := range updates {
		if update.UpdateID >= offset {
			offset = update.UpdateID + 1
		}
		go b.processUpdate(ctx, &update)
	}

	return offset
}

func (b *Bot) processUpdate(ctx context.Context, update *Update) {
	if err := b.handler.Handle(ctx, update); err != nil {
		slog.Error("Error processing update",
			"id", update.UpdateID,
			"error", err,
		)
	}
}
