package telegram

import (
	"context"
	"fmt"
	"got/internal/app"
	"got/internal/app/model"
	"strings"
	"time"
)

const (
	welcomeMessage     = "Welcome! I am ready."
	factError          = "Failed to fetch a fact."
	noFacts            = "No facts available."
	stickerError       = "Failed to fetch a sticker."
	noStickers         = "No stickers available."
	subredditError     = "Failed to fetch a subreddit."
	noSubreddits       = "No subreddits available."
	gptUsage           = "Usage: /gpt <prompt> or /gpt model or /gpt image <prompt>"
	gptModels          = "Available models: gpt-4, gpt-3.5-turbo (Mock)"
	remindListError    = "Failed to list reminders."
	remindNoPending    = "No pending reminders."
	remindUsage        = "Usage: /remind 5s Hello World OR /remind list"
	remindInvalidTime  = "Invalid duration format. Use 5s, 1m, 1h."
	remindSuccess      = "Reminder set for %s"
	remindListHeader   = "Pending reminders:\n"
	remindFormat       = "- %s (at %s)\n"
	memeFormat         = "Here is a meme from r/%s: %s"
	stickerFormat      = "Here is a sticker for you: %s (ID: %s)"
	factFormat         = "Fact: %s"
	gptImageGenerating = "Generating image for: %s"
	gptResponse        = "AI Response to: %s"
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

	return h.client.SendMessage(update.Message.Chat.ID, welcomeMessage)
}

func (h *BotHandlers) HandleFact(ctx context.Context, update *Update) error {
	fact, err := h.service.GetRandomFact(ctx)
	if err != nil {
		return h.client.SendMessage(update.Message.Chat.ID, factError)
	}
	if fact == nil {
		return h.client.SendMessage(update.Message.Chat.ID, noFacts)
	}
	return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf(factFormat, fact.Comment))
}

func (h *BotHandlers) HandleSticker(ctx context.Context, update *Update) error {
	sticker, err := h.service.GetRandomSticker(ctx)
	if err != nil {
		return h.client.SendMessage(update.Message.Chat.ID, stickerError)
	}
	if sticker == nil {
		return h.client.SendMessage(update.Message.Chat.ID, noStickers)
	}
	return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf(stickerFormat, sticker.StickerSetName, sticker.FileID))
}

func (h *BotHandlers) HandleMeme(ctx context.Context, update *Update) error {
	sub, err := h.service.GetRandomSubreddit(ctx)
	if err != nil {
		return h.client.SendMessage(update.Message.Chat.ID, subredditError)
	}

	// Fallback if no subreddits in DB
	subName := "programmerhumor"
	if sub != nil {
		subName = sub.Name
	}

	memes, err := h.fetchMemes(ctx, subName, 1)
	if err != nil || len(memes) == 0 {
		return h.client.SendMessage(update.Message.Chat.ID, "Failed to fetch meme from r/"+subName)
	}

	meme := memes[0]
	// Send text link for now as we don't handle photo upload yet
	return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf("%s\n%s", meme.Title, meme.URL))
}

func (h *BotHandlers) HandleGPT(ctx context.Context, update *Update) error {
	args := update.Message.CommandArguments()
	if args == "" {
		return h.client.SendMessage(update.Message.Chat.ID, gptUsage)
	}

	if strings.HasPrefix(args, "model") {
		return h.client.SendMessage(update.Message.Chat.ID, gptModels)
	}

	if prompt, ok := strings.CutPrefix(args, "image "); ok {
		return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf(gptImageGenerating, prompt))
	}

	return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf(gptResponse, args))
}

func (h *BotHandlers) HandleRemind(ctx context.Context, update *Update) error {
	args := update.Message.CommandArguments()
	if args == "list" {
		reminders, err := h.service.GetPendingReminders(ctx, update.Message.Chat.ID)
		if err != nil {
			return h.client.SendMessage(update.Message.Chat.ID, remindListError)
		}
		if len(reminders) == 0 {
			return h.client.SendMessage(update.Message.Chat.ID, remindNoPending)
		}
		var sb strings.Builder
		sb.WriteString(remindListHeader)
		for _, r := range reminders {
			sb.WriteString(fmt.Sprintf(remindFormat, r.Message, r.RemindAt.Format(time.RFC822)))
		}
		return h.client.SendMessage(update.Message.Chat.ID, sb.String())
	}

	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		return h.client.SendMessage(update.Message.Chat.ID, remindUsage)
	}

	duration, err := time.ParseDuration(parts[0])
	if err != nil {
		return h.client.SendMessage(update.Message.Chat.ID, remindInvalidTime)
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

	return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf(remindSuccess, duration))
}
