package telegram

import (
	"context"
	"fmt"
	"got/internal/app"
	"got/internal/app/model"
	"got/internal/groq"
	"got/internal/redis"
	"got/pkg/i18n"
	"strings"
	"time"
)

const (
	defaultSubreddit = "programmerhumor"
)

type SubCommand string

const (
	SubCommandList  SubCommand = "list"
	SubCommandModel SubCommand = "model"
	SubCommandClear SubCommand = "clear"
)

type BotHandlers struct {
	client  *Client
	service *app.Service
	gpt     *groq.Client
	cache   *redis.Client
	t       *i18n.Translator
}

func NewBotHandlers(client *Client, service *app.Service, gpt *groq.Client, cache *redis.Client, t *i18n.Translator) *BotHandlers {
	return &BotHandlers{
		client:  client,
		service: service,
		gpt:     gpt,
		cache:   cache,
		t:       t,
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

	return h.client.SendMessage(update.Message.Chat.ID, h.t.Get(i18n.KeyWelcome))
}

func (h *BotHandlers) HandleFact(ctx context.Context, update *Update) error {
	fact, err := h.service.GetRandomFact(ctx)
	if err != nil {
		return h.client.SendMessage(update.Message.Chat.ID, h.t.Get(i18n.KeyFactError))
	}
	if fact == nil {
		return h.client.SendMessage(update.Message.Chat.ID, h.t.Get(i18n.KeyNoFacts))
	}
	return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf(h.t.Get(i18n.KeyFactFormat), fact.Comment))
}

func (h *BotHandlers) HandleSticker(ctx context.Context, update *Update) error {
	sticker, err := h.service.GetRandomSticker(ctx)
	if err != nil {
		return h.client.SendMessage(update.Message.Chat.ID, h.t.Get(i18n.KeyStickerError))
	}
	if sticker == nil {
		return h.client.SendMessage(update.Message.Chat.ID, h.t.Get(i18n.KeyNoStickers))
	}
	return h.client.SendSticker(update.Message.Chat.ID, sticker.FileID)
}

func (h *BotHandlers) HandleMeme(ctx context.Context, update *Update) error {
	sub, err := h.service.GetRandomSubreddit(ctx)
	if err != nil {
		return h.client.SendMessage(update.Message.Chat.ID, h.t.Get(i18n.KeySubredditError))
	}

	subName := defaultSubreddit
	if sub != nil {
		subName = sub.Name
	}

	memes, err := h.fetchMemes(ctx, subName, 1)
	if err != nil || len(memes) == 0 {
		return h.client.SendMessage(update.Message.Chat.ID, h.t.Get(i18n.KeyMemeError)+subName)
	}

	meme := memes[0]
	return h.client.SendPhoto(update.Message.Chat.ID, meme.URL, meme.Title)
}

func (h *BotHandlers) HandleGPT(ctx context.Context, update *Update) error {
	if h.gpt == nil {
		return h.client.SendMessage(update.Message.Chat.ID, h.t.Get(i18n.KeyGptNoKey))
	}

	args := update.Message.CommandArguments()
	if args == "" {
		return h.client.SendMessage(update.Message.Chat.ID, h.t.Get(i18n.KeyGptUsage))
	}

	chatID := update.Message.Chat.ID

	if args == string(SubCommandModel) {
		return h.handleGPTModels(chatID)
	}

	if args == string(SubCommandClear) {
		return h.handleGPTClear(ctx, chatID)
	}

	return h.handleGPTChat(ctx, chatID, args)
}

func (h *BotHandlers) HandleRemind(ctx context.Context, update *Update) error {
	args := update.Message.CommandArguments()
	if args == string(SubCommandList) {
		reminders, err := h.service.GetPendingReminders(ctx, update.Message.Chat.ID)
		if err != nil {
			return h.client.SendMessage(update.Message.Chat.ID, h.t.Get(i18n.KeyRemindListError))
		}
		if len(reminders) == 0 {
			return h.client.SendMessage(update.Message.Chat.ID, h.t.Get(i18n.KeyRemindNoPending))
		}
		var sb strings.Builder
		sb.WriteString(h.t.Get(i18n.KeyRemindHeader))
		for _, r := range reminders {
			sb.WriteString(fmt.Sprintf(h.t.Get(i18n.KeyRemindFormat), r.Message, r.RemindAt.Format(time.RFC822)))
		}
		return h.client.SendMessage(update.Message.Chat.ID, sb.String())
	}

	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		return h.client.SendMessage(update.Message.Chat.ID, h.t.Get(i18n.KeyRemindUsage))
	}

	duration, err := time.ParseDuration(parts[0])
	if err != nil {
		return h.client.SendMessage(update.Message.Chat.ID, h.t.Get(i18n.KeyRemindInvalid))
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

	return h.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf(h.t.Get(i18n.KeyRemindSuccess), duration))
}

func (h *BotHandlers) handleGPTModels(chatID int64) error {
	var sb strings.Builder
	sb.WriteString(h.t.Get(i18n.KeyGptModelsHeader))
	for _, m := range h.gpt.ListModels() {
		sb.WriteString("- " + m + "\n")
	}
	return h.client.SendMessage(chatID, sb.String())
}

func (h *BotHandlers) handleGPTClear(ctx context.Context, chatID int64) error {
	if h.cache != nil {
		_ = h.cache.ClearHistory(ctx, chatID)
	}
	return h.client.SendMessage(chatID, h.t.Get(i18n.KeyGptCleared))
}

func (h *BotHandlers) handleGPTChat(ctx context.Context, chatID int64, prompt string) error {
	var history []groq.Message
	if h.cache != nil {
		history, _ = h.cache.GetHistory(ctx, chatID)
	}

	response, err := h.gpt.Chat(ctx, prompt, history)
	if err != nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyGptError))
	}

	if h.cache != nil {
		history = append(history, groq.Message{Role: "user", Content: prompt})
		history = append(history, groq.Message{Role: "assistant", Content: response})
		_ = h.cache.SaveHistory(ctx, chatID, history)
	}

	return h.client.SendMessage(chatID, response)
}
