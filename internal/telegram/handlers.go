package telegram

import (
	"context"
	"fmt"
	"got/internal/app"
	"got/internal/app/model"
	"strings"
	"time"
)

type BotHandlers struct {
	client  *Client
	service *app.Service
}

func NewBotHandlers(client *Client, service *app.Service) *BotHandlers {
	return &BotHandlers{
		client:  client,
		service: service,
	}
}

func (h *BotHandlers) HandleStart(ctx context.Context, update *Update) error {
	if update.Message.From != nil {
		_ = h.service.RegisterUser(ctx, &model.User{
			UserID:   update.Message.From.ID,
			Username: update.Message.From.UserName,
		})
	}

	if update.Message.Chat != nil {
		_ = h.service.RegisterChat(ctx, &model.Chat{
			ChatID:   update.Message.Chat.ID,
			ChatName: "Telegram Chat",
		})
	}

	return h.client.SendMessage(update.Message.Chat.ID, "Welcome! I am ready.")
}

func (h *BotHandlers) HandleFact(ctx context.Context, update *Update) error {
	fact, err := h.service.GetRandomFact(ctx)
	if err != nil {
		return h.client.SendMessage(update.Message.Chat.ID, "Failed to fetch a fact.")
	}
	if fact == nil {
		return h.client.SendMessage(update.Message.Chat.ID, "No facts available.")
	}
	return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Fact: %s", fact.Comment))
}

func (h *BotHandlers) HandleSticker(ctx context.Context, update *Update) error {
	sticker, err := h.service.GetRandomSticker(ctx)
	if err != nil {
		return h.client.SendMessage(update.Message.Chat.ID, "Failed to fetch a sticker.")
	}
	if sticker == nil {
		return h.client.SendMessage(update.Message.Chat.ID, "No stickers available.")
	}
	return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Here is a sticker for you: %s (ID: %s)", sticker.StickerSetName, sticker.FileID))
}

func (h *BotHandlers) HandleMeme(ctx context.Context, update *Update) error {
	sub, err := h.service.GetRandomSubreddit(ctx)
	if err != nil {
		return h.client.SendMessage(update.Message.Chat.ID, "Failed to fetch a subreddit.")
	}
	if sub == nil {
		return h.client.SendMessage(update.Message.Chat.ID, "No subreddits available.")
	}
	return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Here is a meme from r/%s", sub.Name))
}

func (h *BotHandlers) HandleGPT(ctx context.Context, update *Update) error {
	args := update.Message.CommandArguments()
	if args == "" {
		return h.client.SendMessage(update.Message.Chat.ID, "Usage: /gpt <prompt> or /gpt model or /gpt image <prompt>")
	}

	if strings.HasPrefix(args, "model") {
		return h.client.SendMessage(update.Message.Chat.ID, "Available models: gpt-4, gpt-3.5-turbo (Mock)")
	}

	if prompt, ok := strings.CutPrefix(args, "image "); ok {
		return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Generating image for: %s", prompt))
	}

	return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf("AI Response to: %s", args))
}

func (h *BotHandlers) HandleRemind(ctx context.Context, update *Update) error {
	args := update.Message.CommandArguments()
	if args == "list" {
		reminders, err := h.service.GetPendingReminders(ctx, update.Message.Chat.ID)
		if err != nil {
			return h.client.SendMessage(update.Message.Chat.ID, "Failed to list reminders.")
		}
		if len(reminders) == 0 {
			return h.client.SendMessage(update.Message.Chat.ID, "No pending reminders.")
		}
		var sb strings.Builder
		sb.WriteString("Pending reminders:\n")
		for _, r := range reminders {
			sb.WriteString(fmt.Sprintf("- %s (at %s)\n", r.Message, r.RemindAt.Format(time.RFC822)))
		}
		return h.client.SendMessage(update.Message.Chat.ID, sb.String())
	}

	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		return h.client.SendMessage(update.Message.Chat.ID, "Usage: /remind 5s Hello World OR /remind list")
	}

	duration, err := time.ParseDuration(parts[0])
	if err != nil {
		return h.client.SendMessage(update.Message.Chat.ID, "Invalid duration format. Use 5s, 1m, 1h.")
	}

	err = h.service.AddReminder(
		ctx,
		update.Message.Chat.ID,
		update.Message.From.ID,
		parts[1],
		duration,
	)
	if err != nil {
		return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Error: %v", err))
	}

	return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Reminder set for %s", duration))
}
