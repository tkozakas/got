package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"

	"got/internal/app"
	"got/internal/app/model"
	"got/internal/groq"
	"got/internal/redis"
	"got/internal/tts"
	"got/pkg/config"
	"got/pkg/i18n"
)

const defaultSubreddit = "programmerhumor"

const (
	subCommandList   subCommand = "list"
	subCommandAdd    subCommand = "add"
	subCommandRemove subCommand = "remove"
	subCommandDelete subCommand = "delete"
	subCommandModel  subCommand = "model"
	subCommandClear  subCommand = "clear"
	subCommandForget subCommand = "forget"
	subCommandAll    subCommand = "all"
	subCommandStats  subCommand = "stats"
	subCommandMemory subCommand = "memory"
	subCommandImage  subCommand = "image"
	subCommandLogin  subCommand = "login"
	subCommandReset  subCommand = "reset"
)

const (
	actionTyping          = "typing"
	actionUploadPhoto     = "upload_photo"
	actionRecordVoice     = "record_voice"
	actionUploadVoice     = "upload_voice"
	actionUploadDocument  = "upload_document"
	actionUploadVideo     = "upload_video"
	actionRecordVideoNote = "record_video_note"
	actionUploadVideoNote = "upload_video_note"
)

type subCommand string

type BotHandlers struct {
	client      *Client
	service     *app.Service
	gpt         *groq.Client
	cache       *redis.Client
	t           *i18n.Translator
	tts         *tts.Client
	cmds        *config.CommandsConfig
	sentences   *SentenceProvider
	adminPass   string
	defaultLang string
	translators map[string]*i18n.Translator
}

var supportedLanguages = []string{"en", "ru", "lt", "ja"}

func NewBotHandlers(client *Client, service *app.Service, gpt *groq.Client, cache *redis.Client, t *i18n.Translator, tts *tts.Client, cmds *config.CommandsConfig, adminPass string) *BotHandlers {
	translators := make(map[string]*i18n.Translator)
	for _, lang := range []string{"en", "ru", "lt", "ja"} {
		translators[lang] = i18n.New(lang)
	}

	return &BotHandlers{
		client:      client,
		service:     service,
		gpt:         gpt,
		cache:       cache,
		t:           t,
		tts:         tts,
		cmds:        cmds,
		sentences:   NewSentenceProvider(),
		adminPass:   adminPass,
		defaultLang: t.Lang(),
		translators: translators,
	}
}

func (h *BotHandlers) registerChatUser(ctx context.Context, msg *Message) {
	if msg.Chat == nil || msg.From == nil {
		return
	}
	_ = h.service.RegisterChat(ctx, &model.Chat{
		ChatID:   msg.Chat.ID,
		ChatName: msg.Chat.Title,
	})
	_ = h.service.RegisterUser(ctx, &model.User{
		UserID:   msg.From.ID,
		Username: msg.From.UserName,
	}, msg.Chat.ID)
}

func (h *BotHandlers) HandleStart(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	t := h.getTranslator(ctx, chatID)
	return h.client.SendMessage(chatID, t.Get(i18n.KeyWelcome))
}

func (h *BotHandlers) HandleHelp(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	t := h.getTranslator(ctx, chatID)

	commands := []struct {
		cmd      string
		desc     i18n.Key
		subCmds  []string
		excluded bool
	}{
		{h.cmds.Start, i18n.KeyCmdStart, nil, true},
		{h.cmds.Help, i18n.KeyCmdHelp, nil, true},
		{h.cmds.Gpt, i18n.KeyCmdGpt, []string{"image", "model", "memory", "clear"}, false},
		{h.cmds.Meme, i18n.KeyCmdMeme, []string{"list", "add", "remove"}, false},
		{h.cmds.Sticker, i18n.KeyCmdSticker, []string{"list", "add", "remove"}, false},
		{h.cmds.Fact, i18n.KeyCmdFact, []string{"add"}, false},
		{h.cmds.Tts, i18n.KeyCmdTts, nil, false},
		{h.cmds.Roulette, i18n.KeyCmdRoulette, []string{"stats", "all"}, false},
		{h.cmds.Remind, i18n.KeyCmdRemind, []string{"list", "delete"}, false},
	}

	var sb strings.Builder
	sb.WriteString(t.Get(i18n.KeyHelpHeader))
	sb.WriteString("\n")
	for _, c := range commands {
		if c.excluded {
			continue
		}
		subCmdsStr := ""
		if len(c.subCmds) > 0 {
			subCmdsStr = fmt.Sprintf(" `<%s>`", strings.Join(c.subCmds, ", "))
		}
		sb.WriteString(fmt.Sprintf("- `/%s` — %s%s\n", c.cmd, t.Get(c.desc), subCmdsStr))
	}

	return h.client.SendMessage(chatID, sb.String())
}

func (h *BotHandlers) HandleFact(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	t := h.getTranslator(ctx, chatID)
	args := strings.TrimSpace(update.Message.CommandArguments())
	parts := strings.SplitN(args, " ", 2)

	if len(parts) > 0 && subCommand(parts[0]) == subCommandAdd {
		if len(parts) < 2 {
			return h.client.SendMessage(chatID, t.Get(i18n.KeyFactUsage))
		}
		if err := h.service.AddFact(ctx, parts[1], chatID); err != nil {
			return h.client.SendMessage(chatID, t.Get(i18n.KeyFactError))
		}
		return h.client.SendMessage(chatID, t.Get(i18n.KeyFactAdded))
	}

	fact, err := h.service.GetRandomFact(ctx, chatID)
	if err != nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyFactError))
	}
	if fact == nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyNoFacts))
	}
	return h.client.SendMessage(chatID, fmt.Sprintf(t.Get(i18n.KeyFactFormat), fact.Comment))
}

func (h *BotHandlers) HandleSticker(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	t := h.getTranslator(ctx, chatID)
	args := strings.TrimSpace(update.Message.CommandArguments())
	parts := strings.Fields(args)

	subCmd := subCommand("")
	if len(parts) > 0 {
		subCmd = subCommand(parts[0])
	}

	switch subCmd {
	case subCommandAdd:
		if len(parts) > 1 {
			return h.addStickerSet(ctx, chatID, parts[1])
		}
		if update.Message.ReplyToMessage == nil || update.Message.ReplyToMessage.Sticker == nil {
			return h.client.SendMessage(chatID, t.Get(i18n.KeyStickerUsage))
		}
		sticker := update.Message.ReplyToMessage.Sticker
		if err := h.service.AddSticker(ctx, sticker.FileID, sticker.SetName, chatID); err != nil {
			return h.client.SendMessage(chatID, t.Get(i18n.KeyStickerError))
		}
		return h.client.SendMessage(chatID, t.Get(i18n.KeyStickerAdded))

	case subCommandRemove:
		if len(parts) > 1 {
			return h.removeStickerSet(ctx, chatID, parts[1])
		}
		if update.Message.ReplyToMessage == nil || update.Message.ReplyToMessage.Sticker == nil {
			return h.client.SendMessage(chatID, t.Get(i18n.KeyStickerRemoveUsage))
		}
		fileID := update.Message.ReplyToMessage.Sticker.FileID
		if err := h.service.RemoveSticker(ctx, fileID, chatID); err != nil {
			return h.client.SendMessage(chatID, t.Get(i18n.KeyStickerError))
		}
		return h.client.SendMessage(chatID, t.Get(i18n.KeyStickerRemoved))

	case subCommandList:
		stickers, err := h.service.ListStickers(ctx, chatID)
		if err != nil {
			return h.client.SendMessage(chatID, t.Get(i18n.KeyStickerError))
		}
		return h.client.SendMessage(chatID, h.formatStickerList(t, stickers))

	default:
		sticker, err := h.service.GetRandomSticker(ctx, chatID)
		if err != nil {
			return h.client.SendMessage(chatID, t.Get(i18n.KeyStickerError))
		}
		if sticker == nil {
			return h.client.SendMessage(chatID, t.Get(i18n.KeyNoStickers))
		}
		return h.client.SendSticker(chatID, sticker.FileID)
	}
}

func (h *BotHandlers) HandleMeme(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	t := h.getTranslator(ctx, chatID)
	args := strings.TrimSpace(update.Message.CommandArguments())
	parts := strings.Fields(args)

	if len(parts) > 0 {
		switch subCommand(parts[0]) {
		case subCommandAdd:
			if len(parts) < 2 {
				return h.client.SendMessage(chatID, t.Get(i18n.KeyMemeUsage))
			}
			if err := h.service.AddSubreddit(ctx, parts[1], chatID); err != nil {
				return h.client.SendMessage(chatID, t.Get(i18n.KeySubredditError))
			}
			return h.client.SendMessage(chatID, fmt.Sprintf(t.Get(i18n.KeyMemeAdded), parts[1]))

		case subCommandRemove:
			if len(parts) < 2 {
				return h.client.SendMessage(chatID, t.Get(i18n.KeyMemeUsage))
			}
			if err := h.service.RemoveSubreddit(ctx, parts[1], chatID); err != nil {
				return h.client.SendMessage(chatID, t.Get(i18n.KeySubredditError))
			}
			return h.client.SendMessage(chatID, fmt.Sprintf(t.Get(i18n.KeyMemeRemoved), parts[1]))

		case subCommandList:
			subs, err := h.service.ListSubreddits(ctx, chatID)
			if err != nil {
				return h.client.SendMessage(chatID, t.Get(i18n.KeySubredditError))
			}
			return h.client.SendMessage(chatID, h.formatSubredditList(t, subs))
		}
	}

	count := 1
	var explicitSubreddit string

	if len(parts) > 0 {
		if n, err := strconv.Atoi(parts[0]); err == nil {
			if n < 1 || n > 5 {
				return h.client.SendMessage(chatID, t.Get(i18n.KeyMemeCountInvalid))
			}
			count = n
			if len(parts) > 1 {
				explicitSubreddit = parts[1]
			}
		} else {
			explicitSubreddit = parts[0]
			if len(parts) > 1 {
				if n, err := strconv.Atoi(parts[1]); err == nil {
					if n < 1 || n > 5 {
						return h.client.SendMessage(chatID, t.Get(i18n.KeyMemeCountInvalid))
					}
					count = n
				}
			}
		}
	}

	var subName string
	if explicitSubreddit != "" {
		subName = explicitSubreddit
	} else {
		sub, err := h.service.GetRandomSubreddit(ctx, chatID)
		if err != nil {
			return h.client.SendMessage(chatID, t.Get(i18n.KeySubredditError))
		}
		if sub != nil {
			subName = sub.Name
		} else {
			subName = defaultSubreddit
		}
	}

	_ = h.client.SendChatAction(chatID, actionUploadPhoto)

	memes, err := h.fetchMemes(ctx, subName, count)
	if err != nil || len(memes) == 0 {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyMemeError)+subName)
	}

	return h.sendMemes(chatID, memes)
}

func (h *BotHandlers) HandleGPT(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	t := h.getTranslator(ctx, chatID)

	if h.gpt == nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyGptNoKey))
	}

	args := update.Message.CommandArguments()
	if args == "" {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyGptUsage))
	}

	parts := strings.SplitN(args, " ", 2)
	subCmd := subCommand(parts[0])

	switch subCmd {
	case subCommandModel:
		if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
			return h.handleGPTSetModel(ctx, chatID, strings.TrimSpace(parts[1]))
		}
		return h.handleGPTModels(ctx, chatID)
	case subCommandClear, subCommandForget:
		return h.handleGPTClear(ctx, chatID)
	case subCommandMemory:
		return h.handleGPTMemory(ctx, chatID)
	case subCommandImage:
		return h.handleGPTImage(ctx, chatID, parts)
	default:
		username := ""
		if update.Message.From != nil {
			username = update.Message.From.UserName
		}
		return h.handleGPTChat(ctx, chatID, username, args)
	}
}

func (h *BotHandlers) HandleRemind(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	t := h.getTranslator(ctx, chatID)
	args := update.Message.CommandArguments()
	parts := strings.SplitN(args, " ", 2)

	if len(parts) == 0 {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRemindUsage))
	}

	switch subCommand(parts[0]) {
	case subCommandList:
		return h.handleRemindList(ctx, chatID)
	case subCommandDelete:
		return h.handleRemindDelete(ctx, chatID, parts)
	default:
		return h.handleRemindAdd(ctx, chatID, update.Message.From.ID, parts)
	}
}

func (h *BotHandlers) HandleRoulette(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	t := h.getTranslator(ctx, chatID)
	userID := update.Message.From.ID
	currentYear := time.Now().Year()

	h.registerChatUser(ctx, update.Message)
	_, _ = h.service.GetOrCreateStat(ctx, userID, chatID, currentYear)

	args := strings.TrimSpace(update.Message.CommandArguments())
	parts := strings.Fields(args)

	if len(parts) == 0 {
		return h.handleRouletteSpin(ctx, chatID, currentYear)
	}

	switch subCommand(parts[0]) {
	case subCommandAll:
		return h.handleRouletteAll(ctx, chatID)
	case subCommandStats:
		year := currentYear
		if len(parts) > 1 {
			if y, err := strconv.Atoi(parts[1]); err == nil {
				year = y
			}
		}
		return h.handleRouletteYear(ctx, chatID, year)
	default:
		if year, err := strconv.Atoi(parts[0]); err == nil {
			return h.handleRouletteYear(ctx, chatID, year)
		}
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRouletteUsage))
	}
}

func (h *BotHandlers) HandleTTS(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	t := h.getTranslator(ctx, chatID)
	text := update.Message.CommandArguments()

	if text == "" {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyTtsUsage))
	}

	if h.tts == nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyTtsError))
	}

	typing := h.startTyping(ctx, chatID, actionRecordVoice)
	defer typing.Stop()

	audioData, err := h.tts.GenerateSpeech(ctx, text)
	if err != nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyTtsError))
	}

	return h.client.SendVoice(chatID, audioData, "speech.mp3")
}

func (h *BotHandlers) HandleAdmin(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	t := h.getTranslator(ctx, chatID)
	userID := update.Message.From.ID
	isPrivate := update.Message.Chat.Type == "private"

	if h.adminPass == "" {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyAdminNoPass))
	}

	args := strings.TrimSpace(update.Message.CommandArguments())
	parts := strings.SplitN(args, " ", 2)

	if len(parts) == 0 || parts[0] == "" {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyAdminUsage))
	}

	switch subCommand(parts[0]) {
	case subCommandLogin:
		return h.handleAdminLogin(ctx, chatID, userID, parts, isPrivate)
	case subCommandReset:
		return h.handleAdminReset(ctx, chatID, userID)
	default:
		return h.client.SendMessage(chatID, t.Get(i18n.KeyAdminUsage))
	}
}

func (h *BotHandlers) HandleLang(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	args := strings.TrimSpace(update.Message.CommandArguments())

	if args == "" {
		return h.showCurrentLanguage(ctx, chatID)
	}

	return h.setLanguage(ctx, chatID, args)
}

func (h *BotHandlers) handleRemindList(ctx context.Context, chatID int64) error {
	t := h.getTranslator(ctx, chatID)
	reminders, err := h.service.GetPendingReminders(ctx, chatID)
	if err != nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRemindListError))
	}
	if len(reminders) == 0 {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRemindNoPending))
	}

	return h.client.SendMessage(chatID, h.formatReminders(t, reminders))
}

func (h *BotHandlers) handleRemindDelete(ctx context.Context, chatID int64, parts []string) error {
	t := h.getTranslator(ctx, chatID)
	if len(parts) < 2 {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRemindDeleteUsage))
	}

	reminderID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRemindDeleteUsage))
	}

	if err := h.service.DeleteReminder(ctx, reminderID, chatID); err != nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRemindDeleteError))
	}

	return h.client.SendMessage(chatID, t.Get(i18n.KeyRemindDeleted))
}

func (h *BotHandlers) handleRemindAdd(ctx context.Context, chatID int64, userID int64, parts []string) error {
	t := h.getTranslator(ctx, chatID)
	if len(parts) < 2 {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRemindUsage))
	}

	duration, err := ParseDuration(parts[0])
	if err != nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRemindInvalid))
	}

	if err := h.service.AddReminder(ctx, chatID, userID, parts[1], duration); err != nil {
		return h.client.SendMessage(chatID, fmt.Sprintf("Error: %v", err))
	}

	return h.client.SendMessage(chatID, fmt.Sprintf(t.Get(i18n.KeyRemindSuccess), duration))
}

func (h *BotHandlers) formatReminders(t *i18n.Translator, reminders []*model.Reminder) string {
	var sb strings.Builder
	sb.WriteString(t.Get(i18n.KeyRemindHeader))
	for _, r := range reminders {
		sb.WriteString(fmt.Sprintf(t.Get(i18n.KeyRemindFormat), r.ReminderID, r.Message, r.RemindAt.Format(time.RFC822)))
	}
	return sb.String()
}

func (h *BotHandlers) handleRouletteSpin(ctx context.Context, chatID int64, year int) error {
	t := h.getTranslator(ctx, chatID)
	alias := h.cmds.Roulette

	winner, err := h.service.GetTodayWinner(ctx, chatID, year)
	if err != nil {
		slog.Error("roulette: failed to get today winner", "chatID", chatID, "error", err)
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRouletteNoStats))
	}

	if winner != nil {
		msg := fmt.Sprintf(t.Get(i18n.KeyRouletteWinnerExists), alias, h.formatUser(winner.User), winner.Score)
		return h.client.SendMessage(chatID, msg)
	}

	winner, err = h.service.SelectRandomWinner(ctx, chatID, year)
	if err != nil {
		slog.Error("roulette: failed to select random winner", "chatID", chatID, "error", err)
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRouletteNoUsers))
	}
	if winner == nil {
		slog.Warn("roulette: no users found in chat", "chatID", chatID)
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRouletteNoUsers))
	}

	winnerName := h.formatUser(winner.User)
	fallbackMsg := fmt.Sprintf(t.Get(i18n.KeyRouletteWinnerNew), alias, winnerName)
	return h.sentences.SendSequence(h.client, chatID, t.Lang(), alias, winnerName, fallbackMsg)
}

func (h *BotHandlers) handleRouletteYear(ctx context.Context, chatID int64, year int) error {
	t := h.getTranslator(ctx, chatID)
	stats, err := h.service.GetStatsByYear(ctx, chatID, year)
	if err != nil || len(stats) == 0 {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRouletteNoStats))
	}

	return h.client.SendMessage(chatID, h.formatStats(t, stats, fmt.Sprintf(t.Get(i18n.KeyRouletteHeader), year)))
}

func (h *BotHandlers) handleRouletteAll(ctx context.Context, chatID int64) error {
	t := h.getTranslator(ctx, chatID)
	stats, err := h.service.GetAllStats(ctx, chatID)
	if err != nil || len(stats) == 0 {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyRouletteNoStats))
	}

	aggregated := h.aggregateStats(stats)
	return h.client.SendMessage(chatID, h.formatStats(t, aggregated, t.Get(i18n.KeyRouletteHeaderAll)))
}

func (h *BotHandlers) handleGPTModels(ctx context.Context, chatID int64) error {
	t := h.getTranslator(ctx, chatID)
	currentModel := h.getChatModel(ctx, chatID)
	models := h.fetchModelsWithFallback(ctx)

	var sb strings.Builder
	sb.WriteString(t.Get(i18n.KeyGptModelsHeader))
	for i, m := range models {
		prefix := "  "
		if m == currentModel {
			prefix = "→ "
		}
		sb.WriteString(fmt.Sprintf("%s%d. %s\n", prefix, i+1, m))
	}
	return h.client.SendMessage(chatID, sb.String())
}

func (h *BotHandlers) fetchModelsWithFallback(ctx context.Context) []string {
	models, err := h.gpt.FetchModels(ctx)
	if err != nil || len(models) == 0 {
		return h.gpt.ListModels()
	}
	return models
}

func (h *BotHandlers) handleGPTSetModel(ctx context.Context, chatID int64, modelInput string) error {
	t := h.getTranslator(ctx, chatID)
	modelName := h.resolveModelName(ctx, modelInput)

	if err := h.gpt.ValidateModel(modelName); err != nil {
		models := h.gpt.ListModels()
		var sb strings.Builder
		sb.WriteString(t.Get(i18n.KeyGptModelInvalid))
		for i, m := range models {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, m))
		}
		return h.client.SendMessage(chatID, sb.String())
	}

	if h.cache != nil {
		_ = h.cache.SetModel(ctx, chatID, modelName)
	}

	return h.client.SendMessage(chatID, fmt.Sprintf(t.Get(i18n.KeyGptModelSet), modelName))
}

func (h *BotHandlers) resolveModelName(ctx context.Context, input string) string {
	num, err := strconv.Atoi(input)
	if err != nil {
		return input
	}

	models := h.fetchModelsWithFallback(ctx)
	idx := num - 1
	if idx < 0 || idx >= len(models) {
		return input
	}

	return models[idx]
}

func (h *BotHandlers) getChatModel(ctx context.Context, chatID int64) string {
	if h.cache == nil {
		return ""
	}
	model, _ := h.cache.GetModel(ctx, chatID)
	return model
}

func (h *BotHandlers) startTyping(ctx context.Context, chatID int64, action string) *TypingIndicator {
	indicator := NewTypingIndicator(h.client, chatID)
	indicator.Start(ctx, action)
	return indicator
}

func (h *BotHandlers) handleGPTClear(ctx context.Context, chatID int64) error {
	t := h.getTranslator(ctx, chatID)
	if h.cache != nil {
		_ = h.cache.ClearHistory(ctx, chatID)
	}
	return h.client.SendMessage(chatID, t.Get(i18n.KeyGptCleared))
}

func (h *BotHandlers) handleGPTMemory(ctx context.Context, chatID int64) error {
	t := h.getTranslator(ctx, chatID)
	if h.cache == nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyGptMemoryNoRedis))
	}

	history, err := h.cache.GetHistory(ctx, chatID)
	if err != nil || len(history) == 0 {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyGptMemoryEmpty))
	}

	_ = h.client.SendChatAction(chatID, actionUploadDocument)

	content := formatHistoryAsText(history)
	filename := fmt.Sprintf("chat_history_%d.txt", chatID)
	caption := t.Get(i18n.KeyGptMemoryCaption)

	return h.client.SendDocument(chatID, []byte(content), filename, caption)
}

func (h *BotHandlers) handleGPTImage(ctx context.Context, chatID int64, parts []string) error {
	t := h.getTranslator(ctx, chatID)
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyGptImageUsage))
	}

	typing := h.startTyping(ctx, chatID, actionUploadPhoto)
	defer typing.Stop()

	prompt := strings.TrimSpace(parts[1])
	imageURL := buildImageURL(prompt)

	return h.client.SendPhoto(chatID, imageURL, prompt)
}

func (h *BotHandlers) handleGPTChat(ctx context.Context, chatID int64, username string, prompt string) error {
	t := h.getTranslator(ctx, chatID)
	typing := h.startTyping(ctx, chatID, actionTyping)
	defer typing.Stop()

	formattedPrompt := formatPromptWithUsername(username, prompt)
	model := h.getChatModel(ctx, chatID)

	var history []groq.Message
	if h.cache != nil {
		history, _ = h.cache.GetHistory(ctx, chatID)
	}

	response, err := h.gpt.ChatWithModel(ctx, formattedPrompt, history, model)
	if err != nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyGptError))
	}

	if h.cache != nil {
		history = append(history, groq.Message{Role: "user", Content: formattedPrompt})
		history = append(history, groq.Message{Role: "assistant", Content: response})
		_ = h.cache.SaveHistory(ctx, chatID, history)
	}

	return h.client.SendMessage(chatID, response)
}

func (h *BotHandlers) formatUser(user *model.User) string {
	name := user.Username
	if name == "" {
		name = fmt.Sprintf("User%d", user.UserID)
	}
	return fmt.Sprintf("[%s](tg://user?id=%d)", name, user.UserID)
}

func (h *BotHandlers) formatStats(t *i18n.Translator, stats []*model.Stat, header string) string {
	var sb strings.Builder
	sb.WriteString("*" + header + "*\n\n")

	for i, stat := range stats {
		username := stat.User.Username
		if stat.IsWinner {
			username = "👑 " + username
		}
		sb.WriteString(fmt.Sprintf(t.Get(i18n.KeyRouletteUser), i+1, username, stat.Score))
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\n*%s*", fmt.Sprintf(t.Get(i18n.KeyRouletteFooter), len(stats))))
	return sb.String()
}

func (h *BotHandlers) formatSubredditList(t *i18n.Translator, subs []*model.Subreddit) string {
	var sb strings.Builder
	sb.WriteString(t.Get(i18n.KeyMemeListHeader))
	for _, s := range subs {
		sb.WriteString("- `r/" + s.Name + "`\n")
	}
	return sb.String()
}

func (h *BotHandlers) addStickerSet(ctx context.Context, chatID int64, setName string) error {
	t := h.getTranslator(ctx, chatID)
	stickerSet, err := h.client.GetStickerSet(setName)
	if err != nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyStickerSetNotFound))
	}

	added := 0
	for _, s := range stickerSet.Stickers {
		if err := h.service.AddSticker(ctx, s.FileID, stickerSet.Name, chatID); err == nil {
			added++
		}
	}

	return h.client.SendMessage(chatID, fmt.Sprintf(t.Get(i18n.KeyStickerSetAdded), stickerSet.Title, added))
}

func (h *BotHandlers) removeStickerSet(ctx context.Context, chatID int64, setName string) error {
	t := h.getTranslator(ctx, chatID)
	removed, err := h.service.RemoveStickerSet(ctx, setName, chatID)
	if err != nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyStickerError))
	}
	if removed == 0 {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyStickerSetNotFound))
	}
	return h.client.SendMessage(chatID, fmt.Sprintf(t.Get(i18n.KeyStickerSetRemoved), setName, removed))
}

func (h *BotHandlers) formatStickerList(t *i18n.Translator, stickers []*model.Sticker) string {
	if len(stickers) == 0 {
		return t.Get(i18n.KeyNoStickers)
	}

	names := make(map[string]struct{})
	for _, s := range stickers {
		if s.StickerSetName != "" {
			names[s.StickerSetName] = struct{}{}
		}
	}

	if len(names) == 0 {
		return fmt.Sprintf(t.Get(i18n.KeyStickerCount), len(stickers))
	}

	var sb strings.Builder
	sb.WriteString(t.Get(i18n.KeyStickerListHeader))
	for name := range names {
		sb.WriteString("- `" + name + "`\n")
	}
	sb.WriteString(fmt.Sprintf("\n_%s_", fmt.Sprintf(t.Get(i18n.KeyStickerCount), len(stickers))))
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

func (h *BotHandlers) sendMemes(chatID int64, memes []model.RedditMeme) error {
	var photos []model.RedditMeme
	var gifs []model.RedditMeme

	for _, meme := range memes {
		if isAnimatedURL(meme.URL) {
			gifs = append(gifs, meme)
		} else {
			photos = append(photos, meme)
		}
	}

	if len(photos) > 1 {
		media := make([]InputMediaPhoto, len(photos))
		for i, meme := range photos {
			media[i] = InputMediaPhoto{
				Type:    "photo",
				Media:   meme.URL,
				Caption: formatMemeCaption(meme),
			}
		}
		if err := h.client.SendMediaGroup(chatID, media); err != nil {
			return err
		}
	} else if len(photos) == 1 {
		if err := h.client.SendPhoto(chatID, photos[0].URL, formatMemeCaption(photos[0])); err != nil {
			return err
		}
	}

	for _, gif := range gifs {
		if err := h.client.SendAnimation(chatID, gif.URL, formatMemeCaption(gif)); err != nil {
			return err
		}
	}

	return nil
}

func (h *BotHandlers) handleAdminLogin(ctx context.Context, chatID, userID int64, parts []string, isPrivate bool) error {
	t := h.getTranslator(ctx, chatID)
	if !isPrivate {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyAdminDMOnly))
	}

	if len(parts) < 2 {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyAdminUsage))
	}

	password := strings.TrimSpace(parts[1])
	if password != h.adminPass {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyAdminUnauthorized))
	}

	if h.cache != nil {
		_ = h.cache.SetAdminSession(ctx, userID, true)
	}

	return h.client.SendMessage(chatID, t.Get(i18n.KeyAdminLoginSuccess))
}

func (h *BotHandlers) handleAdminReset(ctx context.Context, chatID, userID int64) error {
	t := h.getTranslator(ctx, chatID)
	isAdmin, _ := h.isAdmin(ctx, userID)
	if !isAdmin {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyAdminNotLoggedIn))
	}

	year := time.Now().Year()
	if err := h.service.ResetTodayWinner(ctx, chatID, year); err != nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyAdminResetError))
	}

	return h.client.SendMessage(chatID, t.Get(i18n.KeyAdminResetSuccess))
}

func (h *BotHandlers) isAdmin(ctx context.Context, userID int64) (bool, error) {
	if h.cache == nil {
		return false, nil
	}
	return h.cache.GetAdminSession(ctx, userID)
}

func (h *BotHandlers) showCurrentLanguage(ctx context.Context, chatID int64) error {
	t := h.getTranslator(ctx, chatID)
	lang, _ := h.service.GetChatLanguage(ctx, chatID)
	if lang == "" {
		lang = h.defaultLang
	}

	msg := fmt.Sprintf(t.Get(i18n.KeyLangCurrent), lang) + "\n\n" + t.Get(i18n.KeyLangList)
	return h.client.SendMessage(chatID, msg)
}

func (h *BotHandlers) setLanguage(ctx context.Context, chatID int64, lang string) error {
	t := h.getTranslator(ctx, chatID)
	lang = strings.ToLower(strings.TrimSpace(lang))

	if !isValidLanguage(lang) {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyLangUsage))
	}

	if err := h.service.SetChatLanguage(ctx, chatID, lang); err != nil {
		return h.client.SendMessage(chatID, t.Get(i18n.KeyLangUsage))
	}

	newT := h.translators[lang]
	if newT == nil {
		newT = t
	}
	return h.client.SendMessage(chatID, fmt.Sprintf(newT.Get(i18n.KeyLangSet), lang))
}

func (h *BotHandlers) getTranslator(ctx context.Context, chatID int64) *i18n.Translator {
	lang, _ := h.service.GetChatLanguage(ctx, chatID)
	if lang == "" {
		return h.t
	}
	if t, ok := h.translators[lang]; ok {
		return t
	}
	return h.t
}

func formatHistoryAsText(history []groq.Message) string {
	var sb strings.Builder
	for _, msg := range history {
		sb.WriteString(msg.Role)
		sb.WriteString(": ")
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}

func buildImageURL(prompt string) string {
	seed := time.Now().UnixNano() % 1000000
	return fmt.Sprintf(
		"https://image.pollinations.ai/prompt/%s?width=1024&height=1024&seed=%d&nologo=true&enhance=true",
		url.QueryEscape(prompt),
		seed,
	)
}

func formatPromptWithUsername(username string, prompt string) string {
	if strings.TrimSpace(username) == "" {
		return prompt
	}
	return username + ": " + prompt
}

func isAnimatedURL(url string) bool {
	lowerURL := strings.ToLower(url)
	return strings.HasSuffix(lowerURL, ".gif") ||
		strings.Contains(lowerURL, ".gif?") ||
		strings.HasSuffix(lowerURL, ".mp4") ||
		strings.Contains(lowerURL, ".mp4?")
}

func isValidLanguage(lang string) bool {
	for _, l := range supportedLanguages {
		if l == lang {
			return true
		}
	}
	return false
}

func formatMemeCaption(meme model.RedditMeme) string {
	if meme.Subreddit == "" {
		return meme.Title
	}
	return fmt.Sprintf("%s\n\nr/%s", meme.Title, meme.Subreddit)
}
