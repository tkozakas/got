package telegram

import (
	"context"
	"time"
)

const typingInterval = 4 * time.Second

type TypingIndicator struct {
	client *Client
	chatID int64
	done   chan struct{}
}

func NewTypingIndicator(client *Client, chatID int64) *TypingIndicator {
	return &TypingIndicator{
		client: client,
		chatID: chatID,
		done:   make(chan struct{}),
	}
}

func (t *TypingIndicator) Start(ctx context.Context, action string) {
	go t.run(ctx, action)
}

func (t *TypingIndicator) Stop() {
	close(t.done)
}

func (t *TypingIndicator) run(ctx context.Context, action string) {
	_ = t.client.SendChatAction(t.chatID, action)

	ticker := time.NewTicker(typingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.done:
			return
		case <-ticker.C:
			_ = t.client.SendChatAction(t.chatID, action)
		}
	}
}
