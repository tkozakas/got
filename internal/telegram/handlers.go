package telegram

import (
	"context"
	"fmt"
	"got/internal/app"
	"got/internal/app/model"
	"got/internal/groq"
	"got/internal/redis"
	"got/pkg/i18n"
	"strconv"
	"strings"
	"time"
)

const defaultSubreddit = "programmerhumor"

type SubCommand string

const (
	SubCommandList   SubCommand = "list"
	SubCommandAdd    SubCommand = "add"
	SubCommandRemove SubCommand = "remove"
	SubCommandModel  SubCommand = "model"
	SubCommandClear  SubCommand = "clear"
	SubCommandAll    SubCommand = "all"
	SubCommandStats  SubCommand = "stats"
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
	chatID := update.Message.Chat.ID

	if update.Message.Chat != nil {
		_ = h.service.RegisterChat(ctx, &model.Chat{
			ChatID:   chatID,
			ChatName: update.Message.Chat.Title,
		})
	}

	if update.Message.From != nil {
		_ = h.service.RegisterUser(ctx, &model.User{
			UserID:   update.Message.From.ID,
			Username: update.Message.From.UserName,
		}, chatID)
	}

	return h.client.SendMessage(chatID, h.t.Get(i18n.KeyWelcome))
}

func (h *BotHandlers) HandleFact(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	args := strings.TrimSpace(update.Message.CommandArguments())
	parts := strings.SplitN(args, " ", 2)

	if len(parts) > 0 && SubCommand(parts[0]) == SubCommandAdd {
		if len(parts) < 2 {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyFactUsage))
		}
		if err := h.service.AddFact(ctx, parts[1], chatID); err != nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyFactError))
		}
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyFactAdded))
	}

	fact, err := h.service.GetRandomFact(ctx, chatID)
	if err != nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyFactError))
	}
	if fact == nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyNoFacts))
	}
	return h.client.SendMessage(chatID, fmt.Sprintf(h.t.Get(i18n.KeyFactFormat), fact.Comment))
}

func (h *BotHandlers) HandleSticker(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	args := strings.TrimSpace(update.Message.CommandArguments())

	switch SubCommand(args) {
	case SubCommandAdd:
		if update.Message.ReplyToMessage == nil || update.Message.ReplyToMessage.Sticker == nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStickerUsage))
		}
		fileID := update.Message.ReplyToMessage.Sticker.FileID
		if err := h.service.AddSticker(ctx, fileID, chatID); err != nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStickerError))
		}
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStickerAdded))

	case SubCommandList:
		stickers, err := h.service.ListStickers(ctx, chatID)
		if err != nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStickerError))
		}
		return h.client.SendMessage(chatID, fmt.Sprintf(h.t.Get(i18n.KeyStickerListHeader), len(stickers)))

	default:
		sticker, err := h.service.GetRandomSticker(ctx, chatID)
		if err != nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStickerError))
		}
		if sticker == nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyNoStickers))
		}
		return h.client.SendSticker(chatID, sticker.FileID)
	}
}

func (h *BotHandlers) HandleMeme(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	args := strings.TrimSpace(update.Message.CommandArguments())
	parts := strings.SplitN(args, " ", 2)

	if len(parts) > 0 {
		switch SubCommand(parts[0]) {
		case SubCommandAdd:
			if len(parts) < 2 {
				return h.client.SendMessage(chatID, h.t.Get(i18n.KeyMemeUsage))
			}
			if err := h.service.AddSubreddit(ctx, parts[1], chatID); err != nil {
				return h.client.SendMessage(chatID, h.t.Get(i18n.KeySubredditError))
			}
			return h.client.SendMessage(chatID, fmt.Sprintf(h.t.Get(i18n.KeyMemeAdded), parts[1]))

		case SubCommandRemove:
			if len(parts) < 2 {
				return h.client.SendMessage(chatID, h.t.Get(i18n.KeyMemeUsage))
			}
			if err := h.service.RemoveSubreddit(ctx, parts[1], chatID); err != nil {
				return h.client.SendMessage(chatID, h.t.Get(i18n.KeySubredditError))
			}
			return h.client.SendMessage(chatID, fmt.Sprintf(h.t.Get(i18n.KeyMemeRemoved), parts[1]))

		case SubCommandList:
			subs, err := h.service.ListSubreddits(ctx, chatID)
			if err != nil {
				return h.client.SendMessage(chatID, h.t.Get(i18n.KeySubredditError))
			}
			return h.client.SendMessage(chatID, h.formatSubredditList(subs))
		}
	}

	sub, err := h.service.GetRandomSubreddit(ctx, chatID)
	if err != nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeySubredditError))
	}

	subName := defaultSubreddit
	if sub != nil {
		subName = sub.Name
	}

	memes, err := h.fetchMemes(ctx, subName, 1)
	if err != nil || len(memes) == 0 {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyMemeError)+subName)
	}

	return h.client.SendPhoto(chatID, memes[0].URL, memes[0].Title)
}

func (h *BotHandlers) HandleGPT(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID

	if h.gpt == nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyGptNoKey))
	}

	args := update.Message.CommandArguments()
	if args == "" {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyGptUsage))
	}

	switch SubCommand(args) {
	case SubCommandModel:
		return h.handleGPTModels(chatID)
	case SubCommandClear:
		return h.handleGPTClear(ctx, chatID)
	default:
		return h.handleGPTChat(ctx, chatID, args)
	}
}

func (h *BotHandlers) HandleRemind(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	args := update.Message.CommandArguments()

	if SubCommand(args) == SubCommandList {
		reminders, err := h.service.GetPendingReminders(ctx, chatID)
		if err != nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyRemindListError))
		}
		if len(reminders) == 0 {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyRemindNoPending))
		}
		var sb strings.Builder
		sb.WriteString(h.t.Get(i18n.KeyRemindHeader))
		for _, r := range reminders {
			sb.WriteString(fmt.Sprintf(h.t.Get(i18n.KeyRemindFormat), r.Message, r.RemindAt.Format(time.RFC822)))
		}
		return h.client.SendMessage(chatID, sb.String())
	}

	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyRemindUsage))
	}

	duration, err := time.ParseDuration(parts[0])
	if err != nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyRemindInvalid))
	}

	err = h.service.AddReminder(ctx, chatID, update.Message.From.ID, parts[1], duration)
	if err != nil {
		return h.client.SendMessage(chatID, fmt.Sprintf("Error: %v", err))
	}

	return h.client.SendMessage(chatID, fmt.Sprintf(h.t.Get(i18n.KeyRemindSuccess), duration))
}

func (h *BotHandlers) HandleStats(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	currentYear := time.Now().Year()

	_ = h.service.RegisterUser(ctx, &model.User{
		UserID:   userID,
		Username: update.Message.From.UserName,
	}, chatID)

	_, _ = h.service.GetOrCreateStat(ctx, userID, chatID, currentYear)

	args := strings.TrimSpace(update.Message.CommandArguments())
	parts := strings.Fields(args)

	if len(parts) == 0 {
		return h.handleStatsRoulette(ctx, chatID, currentYear)
	}

	switch SubCommand(parts[0]) {
	case SubCommandAll:
		return h.handleStatsAll(ctx, chatID)
	case SubCommandStats:
		year := currentYear
		if len(parts) > 1 {
			if y, err := strconv.Atoi(parts[1]); err == nil {
				year = y
			}
		}
		return h.handleStatsYear(ctx, chatID, year)
	default:
		if year, err := strconv.Atoi(parts[0]); err == nil {
			return h.handleStatsYear(ctx, chatID, year)
		}
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStatsUsage))
	}
}

func (h *BotHandlers) handleStatsRoulette(ctx context.Context, chatID int64, year int) error {
	alias := h.t.Get(i18n.KeyStatsAlias)

	winner, err := h.service.GetTodayWinner(ctx, chatID, year)
	if err != nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStatsNoStats))
	}

	if winner != nil {
		msg := fmt.Sprintf(h.t.Get(i18n.KeyStatsWinnerExists), alias, h.formatUser(winner.User), winner.Score)
		return h.client.SendMessage(chatID, msg)
	}

	winner, err = h.service.SelectRandomWinner(ctx, chatID, year)
	if err != nil || winner == nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStatsNoUsers))
	}

	msg := fmt.Sprintf(h.t.Get(i18n.KeyStatsWinnerNew), alias, h.formatUser(winner.User))
	return h.client.SendMessage(chatID, msg)
}

func (h *BotHandlers) handleStatsYear(ctx context.Context, chatID int64, year int) error {
	stats, err := h.service.GetStatsByYear(ctx, chatID, year)
	if err != nil || len(stats) == 0 {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStatsNoStats))
	}

	return h.client.SendMessage(chatID, h.formatStats(stats, fmt.Sprintf(h.t.Get(i18n.KeyStatsHeader), year)))
}

func (h *BotHandlers) handleStatsAll(ctx context.Context, chatID int64) error {
	stats, err := h.service.GetAllStats(ctx, chatID)
	if err != nil || len(stats) == 0 {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStatsNoStats))
	}

	aggregated := h.aggregateStats(stats)
	return h.client.SendMessage(chatID, h.formatStats(aggregated, h.t.Get(i18n.KeyStatsHeaderAll)))
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

func (h *BotHandlers) formatUser(user *model.User) string {
	return fmt.Sprintf("[%s](tg://user?id=%d)", user.Username, user.UserID)
}

func (h *BotHandlers) formatStats(stats []*model.Stat, header string) string {
	var sb strings.Builder
	sb.WriteString("*" + header + "*\n\n")

	for i, stat := range stats {
		username := stat.User.Username
		if stat.IsWinner {
			username = "👑 " + username
		}
		sb.WriteString(fmt.Sprintf(h.t.Get(i18n.KeyStatsUser), i+1, username, stat.Score))
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\n*%s*", fmt.Sprintf(h.t.Get(i18n.KeyStatsFooter), len(stats))))
	return sb.String()
}

func (h *BotHandlers) formatSubredditList(subs []*model.Subreddit) string {
	var sb strings.Builder
	sb.WriteString(h.t.Get(i18n.KeyMemeListHeader))
	for _, s := range subs {
		sb.WriteString("- r/" + s.Name + "\n")
	}
	return sb.String()
}

func (h *BotHandlers) aggregateStats(stats []*model.Stat) []*model.Stat {
	userStats := make(map[int64]*model.Stat)

	for _, s := range stats {
		if existing, ok := userStats[s.User.UserID]; ok {
			existing.Score += s.Score
			if s.IsWinner {
				existing.IsWinner = true
			}
		} else {
			userStats[s.User.UserID] = &model.Stat{
				User:     s.User,
				Chat:     s.Chat,
				Score:    s.Score,
				IsWinner: s.IsWinner,
			}
		}
	}

	result := make([]*model.Stat, 0, len(userStats))
	for _, s := range userStats {
		result = append(result, s)
	}

	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Score > result[i].Score {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}
